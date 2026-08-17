package api

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/railguard/signgate/internal/config"
	"github.com/railguard/signgate/internal/eip712"
	"github.com/railguard/signgate/internal/intent"
	"github.com/railguard/signgate/internal/policy"
	"github.com/railguard/signgate/internal/receipt"
	"github.com/railguard/signgate/internal/reservation"
	"github.com/railguard/signgate/internal/session"
	"github.com/railguard/signgate/internal/store"
	"github.com/rs/zerolog"
)

type Server struct {
	log         zerolog.Logger
	cfg         config.Config
	policy      *policy.Engine
	reservation reservation.Authority
	store       store.Repository
	eip712      eip712.Signer
	receipt     receipt.Signer
}

func New(log zerolog.Logger, cfg config.Config, pe *policy.Engine, rs reservation.Authority, st store.Repository) *Server {
	return &Server{
		log:         log,
		cfg:         cfg,
		policy:      pe,
		reservation: rs,
		store:       st,
		eip712: eip712.Signer{
			ChainID:            cfg.ChainID,
			AdapterAddress:     cfg.AdapterAddress,
			RailguardSignerKey: cfg.RailguardSignerKey,
		},
		receipt: receipt.Signer{KeyID: cfg.SignerKeyID, PrivateKey: cfg.ReceiptSignerKey},
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", s.health)
	r.Post("/v1/intents/evaluate", s.evaluateIntent)
	r.Group(func(r chi.Router) {
		r.Use(requireAPIKey(s.cfg.APIKey))
		r.Post("/v1/sessions/register", s.registerSession)
		r.Post("/v1/reservations/reserve", s.reserveBudget)
		r.Post("/v1/userops/submitted", s.userOpSubmitted)
		r.Post("/v1/userops/finalized", s.userOpFinalized)
		r.Get("/v1/receipts/{decisionId}", s.getReceipt)
		r.Get("/v1/reconciliation/executions/{sessionId}", s.getChainExecution)
	})
	return r
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) evaluateIntent(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req intent.EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	intentHash, ci, err := intent.HashFromRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.SaveIntent(
		r.Context(),
		intentHash,
		req.AgentID,
		req.Account,
		req.ChainID,
		req.Token,
		req.Recipient,
		req.AmountAtomic,
		req.Limits.MaxPerTransfer,
		req.Limits.MaxTotalSpend,
		req.Limits.AllowBatch,
		ci.Domain,
		ci.Path,
		req.IdempotencyKey,
	); err != nil {
		http.Error(w, "failed to persist intent", http.StatusInternalServerError)
		return
	}

	pin := policy.Input{
		AgentID:      req.AgentID,
		Account:      req.Account,
		ChainID:      req.ChainID,
		Token:        req.Token,
		Recipient:    req.Recipient,
		AmountAtomic: req.AmountAtomic,
	}
	pin.Resource.Method = req.Resource.Method
	pin.Resource.Domain = req.Resource.Domain
	pin.Resource.Path = req.Resource.Path
	pin.Limits.MaxPerTransfer = req.Limits.MaxPerTransfer
	pin.Limits.MaxTotalSpend = req.Limits.MaxTotalSpend

	pout, err := s.policy.Evaluate(r.Context(), pin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	decisionID := "dec_" + uuid.NewString()
	if err := s.store.SaveDecision(r.Context(), decisionID, intentHash, pout.Decision, pout.ReasonCodes, pout.PolicyHash); err != nil {
		http.Error(w, "failed to persist decision", http.StatusInternalServerError)
		return
	}

	if pout.Decision == "ALLOW" && s.cfg.ReceiptSignerKey != "" {
		signed, err := s.receipt.Sign(receipt.Payload{
			DecisionID:   decisionID,
			Decision:     pout.Decision,
			ReasonCodes:  pout.ReasonCodes,
			AgentID:      req.AgentID,
			IntentHash:   intentHash,
			PolicyHash:   pout.PolicyHash,
			ChainID:      req.ChainID,
			Token:        req.Token,
			Recipient:    req.Recipient,
			AmountAtomic: req.AmountAtomic,
		})
		if err == nil {
			payload, _ := json.Marshal(signed.Payload)
			_ = s.store.SaveReceipt(r.Context(), decisionID, intentHash, "", signed.ReceiptHash, payload, signed.Signature, s.cfg.SignerKeyID)
		}
	}

	s.log.Info().
		Str("decisionId", decisionID).
		Str("intentHash", intentHash).
		Str("decision", pout.Decision).
		Int64("latencyMs", time.Since(start).Milliseconds()).
		Msg("intent evaluated")

	writeJSON(w, http.StatusOK, map[string]any{
		"decision":    pout.Decision,
		"decisionId":  decisionID,
		"intentHash":  intentHash,
		"policyHash":  pout.PolicyHash,
		"reasonCodes": pout.ReasonCodes,
	})
}

type registerSessionReq struct {
	DecisionID string `json:"decisionId"`
	SessionKey string `json:"sessionKey"`
	NonceKey   string `json:"nonceKey"`
	ValidAfter int64  `json:"validAfter"`
	ValidUntil int64  `json:"validUntil"`
}

func (s *Server) registerSession(w http.ResponseWriter, r *http.Request) {
	if s.cfg.AdapterAddress == "" || s.cfg.RailguardSignerKey == "" {
		http.Error(w, "session cosigning not configured", http.StatusServiceUnavailable)
		return
	}
	var req registerSessionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.DecisionID == "" || req.SessionKey == "" || req.NonceKey == "" {
		http.Error(w, "decisionId, sessionKey, and nonceKey are required", http.StatusBadRequest)
		return
	}
	if req.ValidAfter <= 0 || req.ValidUntil <= req.ValidAfter {
		http.Error(w, "validAfter and validUntil are required", http.StatusBadRequest)
		return
	}

	_, policyHash, intent, err := s.store.GetAllowDecision(r.Context(), req.DecisionID)
	if err != nil {
		http.Error(w, "decision not authorizable", http.StatusForbidden)
		return
	}
	if !strings.EqualFold(policyHash, s.policy.PolicyHash()) {
		http.Error(w, "policyHash mismatch", http.StatusBadRequest)
		return
	}

	cfg := session.Config{
		Account:          intent.Account,
		NonceKey:         req.NonceKey,
		SessionKey:       req.SessionKey,
		Token:            intent.Token,
		AllowedTarget:    intent.Token,
		AllowedRecipient: intent.Recipient,
		AllowedSelector:  "0xa9059cbb",
		MaxPerTransfer:   intent.MaxPerTransfer,
		MaxTotalSpend:    intent.MaxTotalSpend,
		ValidAfter:       req.ValidAfter,
		ValidUntil:       req.ValidUntil,
		AllowBatch:       intent.AllowBatch,
		PolicyHash:       policyHash,
		ChainID:          s.cfg.ChainID,
		AdapterAddress:   s.cfg.AdapterAddress,
	}
	if err := session.ValidateV1(cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessionID, err := session.DeriveSessionID(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	physical, _ := session.SessionConfigPhysicalHash(cfg)
	digest, sig, err := s.eip712.SignRailguard(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessionCfg := map[string]any{
		"account": cfg.Account, "agentId": intent.AgentID, "sessionKey": cfg.SessionKey,
		"token": cfg.Token, "allowedTarget": cfg.AllowedTarget, "allowedRecipient": cfg.AllowedRecipient,
		"allowedSelector": cfg.AllowedSelector, "nonceKey": cfg.NonceKey, "maxPerTransfer": cfg.MaxPerTransfer,
		"maxTotalSpend": cfg.MaxTotalSpend, "validAfter": cfg.ValidAfter, "validUntil": cfg.ValidUntil,
		"allowBatch": cfg.AllowBatch, "policyHash": cfg.PolicyHash,
	}
	if err := s.store.AuthorizeSession(r.Context(), store.SessionAuthInput{
		DecisionID: req.DecisionID,
		SessionKey: req.SessionKey,
		NonceKey:   req.NonceKey,
		ValidAfter: req.ValidAfter,
		ValidUntil: req.ValidUntil,
	}, sessionID.Hex(), sessionCfg); err != nil {
		http.Error(w, "session authorization failed", http.StatusConflict)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"sessionId":                 sessionID.Hex(),
		"sessionConfigPhysicalHash": physical.Hex(),
		"authorizationDigest":       digest.Hex(),
		"railguardSignature":        "0x" + hexEncode(sig),
	})
}

func hexEncode(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexdigits[v>>4]
		out[i*2+1] = hexdigits[v&0x0f]
	}
	return string(out)
}

type reserveReq struct {
	SessionID      string `json:"sessionId"`
	AgentID        string `json:"agentId"`
	IntentHash     string `json:"intentHash"`
	AmountAtomic   string `json:"amountAtomic"`
	IdempotencyKey string `json:"idempotencyKey"`
	MaxTotalSpend  string `json:"maxTotalSpend"`
}

func (s *Server) reserveBudget(w http.ResponseWriter, r *http.Request) {
	var req reserveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	snap, err := s.store.GetSessionReserveSnapshot(r.Context(), req.SessionID)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if snap.Status != "" && strings.ToUpper(snap.Status) != "SESSION_AUTHORIZED" {
		http.Error(w, "session not active", http.StatusConflict)
		return
	}
	now := time.Now().Unix()
	if now < snap.ValidAfter || now > snap.ValidUntil {
		http.Error(w, "session outside validity window", http.StatusConflict)
		return
	}
	if !strings.EqualFold(req.AgentID, snap.AgentID) {
		http.Error(w, "agent mismatch", http.StatusConflict)
		return
	}
	if snap.IntentHash != "" && !strings.EqualFold(req.IntentHash, snap.IntentHash) {
		http.Error(w, "intent mismatch", http.StatusConflict)
		return
	}
	amount, ok := new(big.Int).SetString(req.AmountAtomic, 10)
	if !ok || amount.Sign() <= 0 {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}
	maxPer, ok := new(big.Int).SetString(snap.MaxPerTransfer, 10)
	if !ok || amount.Cmp(maxPer) > 0 {
		writeJSON(w, http.StatusConflict, map[string]string{"status": "BUDGET_DENIED"})
		return
	}
	sessionTTL := time.Until(time.Unix(snap.ValidUntil, 0))
	if sessionTTL < time.Minute {
		sessionTTL = time.Minute
	}
	resID, err := s.reservation.Reserve(
		r.Context(), req.SessionID, req.IdempotencyKey, req.AmountAtomic, snap.MaxTotalSpend,
		5*time.Minute, sessionTTL,
	)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"status": "BUDGET_DENIED"})
		return
	}
	if err := s.store.SaveReservation(r.Context(), resID, req.SessionID, req.IntentHash, req.AgentID, req.AmountAtomic, "BUDGET_RESERVED", req.IdempotencyKey); err != nil {
		if existing, getErr := s.store.GetReservationIDByIdempotency(r.Context(), req.IdempotencyKey); getErr == nil && existing == resID {
			writeJSON(w, http.StatusOK, map[string]string{"reservationId": resID, "status": "RESERVED"})
			return
		}
		_ = s.reservation.ReleaseReservation(r.Context(), resID)
		http.Error(w, "failed to persist reservation", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"reservationId": resID, "status": "RESERVED"})
}

type submittedReq struct {
	ReservationID   string `json:"reservationId"`
	UserOpHash      string `json:"userOpHash"`
	Bundler         string `json:"bundler"`
	ExecutionDigest string `json:"executionDigest"`
}

func (s *Server) userOpSubmitted(w http.ResponseWriter, r *http.Request) {
	var req submittedReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := s.store.UpdateReservationStatus(r.Context(), req.ReservationID, "USEROP_SUBMITTED"); err != nil {
		http.Error(w, "reservation not found", http.StatusNotFound)
		return
	}
	if req.ExecutionDigest != "" {
		if err := s.store.BindReservationExecutionDigest(r.Context(), req.ReservationID, req.ExecutionDigest); err != nil {
			http.Error(w, "failed to bind execution digest", http.StatusConflict)
			return
		}
	}
	if err := s.store.SaveUserOp(r.Context(), req.UserOpHash, req.ReservationID, "", "USEROP_SUBMITTED", req.Bundler); err != nil {
		http.Error(w, "failed to persist userop", http.StatusInternalServerError)
		return
	}
	if err := s.reservation.FreezeReservation(r.Context(), req.ReservationID); err != nil {
		http.Error(w, "failed to freeze reservation", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "USEROP_SUBMITTED"})
}

type finalizedReq struct {
	UserOpHash  string `json:"userOpHash"`
	TxHash      string `json:"txHash"`
	BlockNumber int64  `json:"blockNumber"`
	Status      string `json:"status"`
}

func (s *Server) userOpFinalized(w http.ResponseWriter, r *http.Request) {
	var req finalizedReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	switch status {
	case "USEROP_FINALIZED", "USEROP_REVERTED":
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	if err := s.store.FinalizeUserOp(r.Context(), req.UserOpHash, req.TxHash, req.BlockNumber, status); err != nil {
		http.Error(w, "userop not found", http.StatusNotFound)
		return
	}
	reservationID, err := s.store.GetReservationIDByUserOp(r.Context(), req.UserOpHash)
	if err == nil {
		switch status {
		case "USEROP_FINALIZED":
			_ = s.reservation.CommitReservation(r.Context(), reservationID)
			_ = s.store.UpdateReservationStatus(r.Context(), reservationID, "BUDGET_COMMITTED")
		case "USEROP_REVERTED":
			_ = s.reservation.ReleaseReservation(r.Context(), reservationID)
			_ = s.store.UpdateReservationStatus(r.Context(), reservationID, "BUDGET_RELEASED")
		}
	}
	outStatus := "BUDGET_COMMITTED"
	if status == "USEROP_REVERTED" {
		outStatus = "BUDGET_RELEASED"
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": outStatus})
}

func (s *Server) getReceipt(w http.ResponseWriter, r *http.Request) {
	decisionID := chi.URLParam(r, "decisionId")
	decision, receiptHash, signature, payload, err := s.store.GetReceipt(r.Context(), decisionID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var receiptPayload any
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &receiptPayload)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"decisionId":  decisionID,
		"decision":    decision,
		"receiptHash": receiptHash,
		"signature":   signature,
		"payload":     receiptPayload,
	})
}

func (s *Server) getChainExecution(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	exec, err := s.store.GetChainExecutionBySessionID(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"account":         exec.Account,
		"sessionId":       exec.SessionID,
		"nonceKey":        exec.NonceKey,
		"frameSpend":      exec.FrameSpend,
		"totalSpendAfter": exec.TotalSpendAfter,
		"blockNumber":     exec.BlockNumber,
		"txHash":          exec.TxHash,
		"logIndex":        exec.LogIndex,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func ListenAndServe(ctx context.Context, log zerolog.Logger, cfg config.Config, srv *Server) error {
	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: srv.Router(),
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()
	log.Info().Str("port", cfg.HTTPPort).Msg("signgate listening")
	return httpServer.ListenAndServe()
}
