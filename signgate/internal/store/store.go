package store

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is the persistence surface used by SignGate API handlers.
type Repository interface {
	SaveIntent(ctx context.Context, intentHash, agentID, account string, chainID int64, token, recipient, amount, maxPerTransfer, maxTotalSpend string, allowBatch bool, domain, path, idem string) error
	SaveDecision(ctx context.Context, decisionID, intentHash, decision string, reasonCodes []string, policyHash string) error
	SaveSessionDraft(ctx context.Context, sessionID string, cfg map[string]any, status string) error
	SaveReservation(ctx context.Context, reservationID, sessionID, intentHash, agentID, amount, status, idem string) error
	UpdateReservationStatus(ctx context.Context, reservationID, status string) error
	SaveUserOp(ctx context.Context, userOpHash, reservationID, sessionID, status, bundler string) error
	FinalizeUserOp(ctx context.Context, userOpHash, txHash string, blockNumber int64, status string) error
	SaveReceipt(ctx context.Context, decisionID, intentHash, sessionID, receiptHash string, payload json.RawMessage, signature, signerKeyID string) error
	GetReceipt(ctx context.Context, decisionID string) (decision, receiptHash, signature string, payload json.RawMessage, err error)
	GetAllowDecision(ctx context.Context, decisionID string) (intentHash, policyHash string, intent IntentSnapshot, err error)
	AuthorizeSession(ctx context.Context, in SessionAuthInput, sessionID string, cfg map[string]any) error
	GetReservationIDByUserOp(ctx context.Context, userOpHash string) (string, error)
	GetReservationIDByIdempotency(ctx context.Context, idempotencyKey string) (string, error)
	BindReservationExecutionDigest(ctx context.Context, reservationID, executionDigest string) error
	GetWatcherBlockCursor(ctx context.Context) (uint64, error)
	SetWatcherBlockCursor(ctx context.Context, block uint64, blockHash string) error
	RecordChainExecution(ctx context.Context, exec ChainExecution) (reservationID string, committed bool, err error)
	GetChainExecutionBySessionID(ctx context.Context, sessionID string) (ChainExecution, error)
	GetSessionReserveSnapshot(ctx context.Context, sessionID string) (SessionReserveSnapshot, error)
	MarkStaleUserOpsReconciliationRequired(ctx context.Context, olderThan time.Time) (int, error)
	CommitBudgetOnChainExecution(ctx context.Context, sessionID, executionDigest string) (reservationID string, committed bool, err error)
	ReserveSessionBudget(ctx context.Context, sessionID, idempotencyKey string, amount, maxTotal *big.Int) (reservationID string, err error)
}

type SessionReserveSnapshot struct {
	AgentID        string
	IntentHash     string
	MaxPerTransfer string
	MaxTotalSpend  string
	ValidAfter     int64
	ValidUntil     int64
	Status         string
}

type ChainExecution struct {
	Account         string
	SessionID       string
	NonceKey        string
	FrameSpend      string
	TotalSpendAfter string
	ExecutionDigest string
	BlockNumber     int64
	TxHash          string
	LogIndex        int
}

type IntentSnapshot struct {
	AgentID         string
	Account         string
	Token           string
	Recipient       string
	AmountAtomic    string
	MaxPerTransfer  string
	MaxTotalSpend   string
	AllowBatch      bool
}

type SessionAuthInput struct {
	DecisionID string
	SessionKey string
	NonceKey   string
	ValidAfter int64
	ValidUntil int64
}

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, url string) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) SaveIntent(ctx context.Context, intentHash, agentID, account string, chainID int64, token, recipient, amount, maxPerTransfer, maxTotalSpend string, allowBatch bool, domain, path, idem string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO payment_intents (
			id, intent_hash, agent_id, account, chain_id, token, recipient, amount_atomic,
			max_per_transfer, max_total_spend, allow_batch, resource_domain, resource_path, idempotency_key
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (intent_hash) DO NOTHING
	`, uuid.New(), intentHash, agentID, account, chainID, token, recipient, amount, maxPerTransfer, maxTotalSpend, allowBatch, domain, path, idem)
	return err
}

func (s *Store) SaveDecision(ctx context.Context, decisionID, intentHash, decision string, reasonCodes []string, policyHash string) error {
	rc, _ := json.Marshal(reasonCodes)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO policy_decisions (id, decision_id, intent_hash, decision, reason_codes, policy_hash)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, uuid.New(), decisionID, intentHash, decision, rc, policyHash)
	return err
}

func (s *Store) SaveSessionDraft(ctx context.Context, sessionID string, cfg map[string]any, status string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (id, session_id, account, agent_id, session_key, token, allowed_target, allowed_recipient, allowed_selector, nonce_key, max_per_transfer, max_total_spend, valid_after, valid_until, allow_batch, policy_hash, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (session_id) DO UPDATE SET status = EXCLUDED.status
	`,
		uuid.New(),
		sessionID,
		cfg["account"],
		cfg["agentId"],
		cfg["sessionKey"],
		cfg["token"],
		cfg["allowedTarget"],
		cfg["allowedRecipient"],
		cfg["allowedSelector"],
		cfg["nonceKey"],
		cfg["maxPerTransfer"],
		cfg["maxTotalSpend"],
		cfg["validAfter"],
		cfg["validUntil"],
		cfg["allowBatch"],
		cfg["policyHash"],
		status,
	)
	return err
}

func (s *Store) SaveReservation(ctx context.Context, reservationID, sessionID, intentHash, agentID, amount, status, idem string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO budget_reservations (id, reservation_id, session_id, intent_hash, agent_id, amount_atomic, status, idempotency_key)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, uuid.New(), reservationID, sessionID, intentHash, agentID, amount, status, idem)
	return err
}

func (s *Store) SaveUserOp(ctx context.Context, userOpHash, reservationID, sessionID, status, bundler string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO userop_lifecycle (id, userop_hash, reservation_id, session_id, status, bundler, submitted_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (userop_hash) DO UPDATE SET status = EXCLUDED.status, bundler = EXCLUDED.bundler, submitted_at = EXCLUDED.submitted_at
	`, uuid.New(), userOpHash, reservationID, sessionID, status, bundler, time.Now().UTC())
	return err
}

func (s *Store) UpdateReservationStatus(ctx context.Context, reservationID, status string) error {
	_, err := s.pool.Exec(ctx, `UPDATE budget_reservations SET status = $2 WHERE reservation_id = $1`, reservationID, status)
	return err
}

func (s *Store) FinalizeUserOp(ctx context.Context, userOpHash, txHash string, blockNumber int64, status string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE userop_lifecycle
		SET status = $2, tx_hash = $3, block_number = $4, finalized_at = $5
		WHERE userop_hash = $1
	`, userOpHash, status, txHash, blockNumber, time.Now().UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("userop not found")
	}
	return nil
}

func (s *Store) SaveReceipt(ctx context.Context, decisionID, intentHash, sessionID, receiptHash string, payload json.RawMessage, signature, signerKeyID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_receipts (id, decision_id, intent_hash, session_id, receipt_hash, receipt_json, signature, signer_key_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, uuid.New(), decisionID, intentHash, sessionID, receiptHash, payload, signature, signerKeyID)
	return err
}

func (s *Store) GetReceipt(ctx context.Context, decisionID string) (string, string, string, json.RawMessage, error) {
	var receiptHash, signature string
	var payload json.RawMessage
	err := s.pool.QueryRow(ctx, `
		SELECT ar.receipt_hash, ar.signature, ar.receipt_json
		FROM audit_receipts ar
		WHERE ar.decision_id = $1
	`, decisionID).Scan(&receiptHash, &signature, &payload)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("receipt not found")
	}
	var decision string
	err = s.pool.QueryRow(ctx, `SELECT decision FROM policy_decisions WHERE decision_id = $1`, decisionID).Scan(&decision)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("decision lookup failed: %w", err)
	}
	return decision, receiptHash, signature, payload, nil
}

func (s *Store) GetAllowDecision(ctx context.Context, decisionID string) (string, string, IntentSnapshot, error) {
	var intentHash, policyHash, decision string
	err := s.pool.QueryRow(ctx, `
		SELECT intent_hash, policy_hash, decision
		FROM policy_decisions
		WHERE decision_id = $1
	`, decisionID).Scan(&intentHash, &policyHash, &decision)
	if err != nil {
		return "", "", IntentSnapshot{}, fmt.Errorf("decision not found")
	}
	if decision != "ALLOW" {
		return "", "", IntentSnapshot{}, fmt.Errorf("decision is not ALLOW")
	}

	var snap IntentSnapshot
	var allowBatch bool
	err = s.pool.QueryRow(ctx, `
		SELECT agent_id, account, token, recipient, amount_atomic::text,
		       COALESCE(max_per_transfer, amount_atomic)::text,
		       COALESCE(max_total_spend, amount_atomic)::text,
		       COALESCE(allow_batch, false)
		FROM payment_intents
		WHERE intent_hash = $1
	`, intentHash).Scan(
		&snap.AgentID, &snap.Account, &snap.Token, &snap.Recipient, &snap.AmountAtomic,
		&snap.MaxPerTransfer, &snap.MaxTotalSpend, &allowBatch,
	)
	if err != nil {
		return "", "", IntentSnapshot{}, fmt.Errorf("intent not found for decision")
	}
	snap.AllowBatch = allowBatch
	return intentHash, policyHash, snap, nil
}

func (s *Store) AuthorizeSession(ctx context.Context, in SessionAuthInput, sessionID string, cfg map[string]any) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var consumedAt *time.Time
	var consumedSessionID *string
	var decision string
	err = tx.QueryRow(ctx, `
		SELECT decision, consumed_at, consumed_session_id
		FROM policy_decisions
		WHERE decision_id = $1
		FOR UPDATE
	`, in.DecisionID).Scan(&decision, &consumedAt, &consumedSessionID)
	if err != nil {
		return fmt.Errorf("decision not found")
	}
	if decision != "ALLOW" {
		return fmt.Errorf("decision is not ALLOW")
	}
	if consumedAt != nil {
		if consumedSessionID != nil && *consumedSessionID == sessionID {
			return tx.Commit(ctx)
		}
		return fmt.Errorf("decision already consumed")
	}

	tag, err := tx.Exec(ctx, `
		UPDATE policy_decisions
		SET consumed_at = now(), consumed_session_id = $2
		WHERE decision_id = $1 AND consumed_at IS NULL
	`, in.DecisionID, sessionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("decision already consumed")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO sessions (
			id, session_id, account, agent_id, session_key, token, allowed_target, allowed_recipient,
			allowed_selector, nonce_key, max_per_transfer, max_total_spend, valid_after, valid_until,
			allow_batch, policy_hash, status
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (session_id) DO NOTHING
	`,
		uuid.New(),
		sessionID,
		cfg["account"],
		cfg["agentId"],
		cfg["sessionKey"],
		cfg["token"],
		cfg["allowedTarget"],
		cfg["allowedRecipient"],
		cfg["allowedSelector"],
		cfg["nonceKey"],
		cfg["maxPerTransfer"],
		cfg["maxTotalSpend"],
		cfg["validAfter"],
		cfg["validUntil"],
		cfg["allowBatch"],
		cfg["policyHash"],
		"SESSION_AUTHORIZED",
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) GetReservationIDByUserOp(ctx context.Context, userOpHash string) (string, error) {
	var reservationID string
	err := s.pool.QueryRow(ctx, `SELECT reservation_id FROM userop_lifecycle WHERE userop_hash = $1`, userOpHash).Scan(&reservationID)
	if err != nil {
		return "", fmt.Errorf("userop not found")
	}
	return reservationID, nil
}

func (s *Store) GetReservationIDByIdempotency(ctx context.Context, idempotencyKey string) (string, error) {
	var reservationID string
	err := s.pool.QueryRow(ctx, `SELECT reservation_id FROM budget_reservations WHERE idempotency_key = $1`, idempotencyKey).Scan(&reservationID)
	if err != nil {
		return "", fmt.Errorf("reservation not found")
	}
	return reservationID, nil
}

func (s *Store) GetWatcherBlockCursor(ctx context.Context) (uint64, error) {
	var value string
	err := s.pool.QueryRow(ctx, `SELECT value FROM watcher_state WHERE key = 'last_block'`).Scan(&value)
	if err != nil {
		return 0, err
	}
	var block uint64
	_, err = fmt.Sscan(value, &block)
	return block, err
}

func (s *Store) SetWatcherBlockCursor(ctx context.Context, block uint64, blockHash string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO watcher_state (key, value, updated_at)
		VALUES ('last_block', $1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`, fmt.Sprintf("%d", block), time.Now().UTC())
	if err != nil {
		return err
	}
	if blockHash != "" {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO watcher_state (key, value, updated_at)
			VALUES ('last_block_hash', $1, $2)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
		`, blockHash, time.Now().UTC())
	}
	return err
}

func (s *Store) RecordChainExecution(ctx context.Context, exec ChainExecution) (string, bool, error) {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO chain_executions (
			id, account, session_id, nonce_key, frame_spend, total_spend_after,
			execution_digest, block_number, tx_hash, log_index
		)
		VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7, ''),$8,$9,$10)
		ON CONFLICT (tx_hash, log_index) DO NOTHING
	`, uuid.New(), exec.Account, exec.SessionID, exec.NonceKey, exec.FrameSpend, exec.TotalSpendAfter,
		exec.ExecutionDigest, exec.BlockNumber, exec.TxHash, exec.LogIndex)
	if err != nil {
		return "", false, err
	}
	if exec.ExecutionDigest == "" {
		return "", false, nil
	}
	return s.CommitBudgetOnChainExecution(ctx, exec.SessionID, exec.ExecutionDigest)
}

func (s *Store) GetChainExecutionBySessionID(ctx context.Context, sessionID string) (ChainExecution, error) {
	var exec ChainExecution
	err := s.pool.QueryRow(ctx, `
		SELECT account, session_id, nonce_key, frame_spend::text, total_spend_after::text, block_number, tx_hash, log_index
		FROM chain_executions
		WHERE lower(session_id) = lower($1)
		ORDER BY created_at DESC
		LIMIT 1
	`, sessionID).Scan(
		&exec.Account, &exec.SessionID, &exec.NonceKey, &exec.FrameSpend, &exec.TotalSpendAfter,
		&exec.BlockNumber, &exec.TxHash, &exec.LogIndex,
	)
	return exec, err
}

func (s *Store) MarkStaleUserOpsReconciliationRequired(ctx context.Context, olderThan time.Time) (int, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE userop_lifecycle
		SET status = 'RECONCILIATION_REQUIRED', last_checked_at = $1
		WHERE status = 'USEROP_SUBMITTED' AND submitted_at IS NOT NULL AND submitted_at < $2
	`, time.Now().UTC(), olderThan)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (s *Store) GetSessionReserveSnapshot(ctx context.Context, sessionID string) (SessionReserveSnapshot, error) {
	var snap SessionReserveSnapshot
	err := s.pool.QueryRow(ctx, `
		SELECT
			s.agent_id,
			COALESCE((
				SELECT intent_hash FROM policy_decisions
				WHERE consumed_session_id = s.session_id AND decision = 'ALLOW'
				ORDER BY created_at DESC
				LIMIT 1
			), ''),
			s.max_per_transfer::text,
			s.max_total_spend::text,
			s.valid_after,
			s.valid_until,
			s.status
		FROM sessions s
		WHERE lower(s.session_id) = lower($1)
	`, sessionID).Scan(
		&snap.AgentID, &snap.IntentHash, &snap.MaxPerTransfer, &snap.MaxTotalSpend,
		&snap.ValidAfter, &snap.ValidUntil, &snap.Status,
	)
	if err != nil {
		return SessionReserveSnapshot{}, fmt.Errorf("session not found")
	}
	return snap, nil
}

func (s *Store) BindReservationExecutionDigest(ctx context.Context, reservationID, executionDigest string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE budget_reservations
		SET execution_digest = $2
		WHERE reservation_id = $1 AND status IN ('BUDGET_RESERVED', 'USEROP_SUBMITTED')
	`, reservationID, executionDigest)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("reservation not found")
	}
	return nil
}

func (s *Store) ReserveSessionBudget(
	ctx context.Context,
	sessionID, idempotencyKey string,
	amount, maxTotal *big.Int,
) (string, error) {
	if existing, err := s.GetReservationIDByIdempotency(ctx, idempotencyKey); err == nil && existing != "" {
		return existing, nil
	}

	var reservedTotal string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_atomic), 0)::text
		FROM budget_reservations
		WHERE lower(session_id) = lower($1)
		  AND status IN ('BUDGET_RESERVED', 'USEROP_SUBMITTED')
	`, sessionID).Scan(&reservedTotal)
	if err != nil {
		return "", err
	}
	current, ok := new(big.Int).SetString(reservedTotal, 10)
	if !ok {
		current = big.NewInt(0)
	}
	next := new(big.Int).Add(current, amount)
	if next.Cmp(maxTotal) > 0 {
		return "", fmt.Errorf("BUDGET_DENIED")
	}
	return "res_" + uuid.NewString(), nil
}

func (Noop) ReserveSessionBudget(context.Context, string, string, *big.Int, *big.Int) (string, error) {
	return "res_noop", nil
}

func (s *Store) CommitBudgetOnChainExecution(ctx context.Context, sessionID, executionDigest string) (string, bool, error) {
	if executionDigest == "" {
		return "", false, nil
	}
	var reservationID string
	err := s.pool.QueryRow(ctx, `
		SELECT reservation_id
		FROM budget_reservations
		WHERE lower(session_id) = lower($1)
		  AND lower(execution_digest) = lower($2)
		  AND status IN ('BUDGET_RESERVED', 'USEROP_SUBMITTED')
		LIMIT 1
	`, sessionID, executionDigest).Scan(&reservationID)
	if err != nil {
		return "", false, nil
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE budget_reservations
		SET status = 'BUDGET_COMMITTED', finalized_at = $2
		WHERE reservation_id = $1
		  AND status IN ('BUDGET_RESERVED', 'USEROP_SUBMITTED')
	`, reservationID, time.Now().UTC())
	if err != nil {
		return "", false, err
	}
	return reservationID, tag.RowsAffected() > 0, nil
}

func (s *Store) EnsureExecutionDigestSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		ALTER TABLE chain_executions ADD COLUMN IF NOT EXISTS execution_digest TEXT;
		ALTER TABLE budget_reservations ADD COLUMN IF NOT EXISTS execution_digest TEXT;
	`)
	return err
}

// Noop satisfies Repository for local runs without Postgres.
type Noop struct{}

func NewNoop() Noop { return Noop{} }

func (Noop) SaveIntent(context.Context, string, string, string, int64, string, string, string, string, string, bool, string, string, string) error {
	return nil
}
func (Noop) SaveDecision(context.Context, string, string, string, []string, string) error { return nil }
func (Noop) SaveSessionDraft(context.Context, string, map[string]any, string) error       { return nil }
func (Noop) SaveReservation(context.Context, string, string, string, string, string, string, string) error {
	return nil
}
func (Noop) UpdateReservationStatus(context.Context, string, string) error { return nil }
func (Noop) SaveUserOp(context.Context, string, string, string, string, string) error {
	return nil
}
func (Noop) FinalizeUserOp(context.Context, string, string, int64, string) error {
	return fmt.Errorf("userop not found")
}
func (Noop) SaveReceipt(context.Context, string, string, string, string, json.RawMessage, string, string) error {
	return nil
}
func (Noop) GetReceipt(context.Context, string) (string, string, string, json.RawMessage, error) {
	return "", "", "", nil, fmt.Errorf("receipt not found")
}
func (Noop) GetAllowDecision(context.Context, string) (string, string, IntentSnapshot, error) {
	return "", "", IntentSnapshot{}, fmt.Errorf("decision not found")
}
func (Noop) AuthorizeSession(context.Context, SessionAuthInput, string, map[string]any) error {
	return fmt.Errorf("decision not found")
}
func (Noop) GetReservationIDByUserOp(context.Context, string) (string, error) {
	return "", fmt.Errorf("userop not found")
}
func (Noop) GetReservationIDByIdempotency(context.Context, string) (string, error) {
	return "", fmt.Errorf("reservation not found")
}
func (Noop) BindReservationExecutionDigest(context.Context, string, string) error { return nil }
func (Noop) GetWatcherBlockCursor(context.Context) (uint64, error) { return 0, fmt.Errorf("noop") }
func (Noop) SetWatcherBlockCursor(context.Context, uint64, string) error { return nil }
func (Noop) RecordChainExecution(context.Context, ChainExecution) (string, bool, error) {
	return "", false, nil
}
func (Noop) GetChainExecutionBySessionID(context.Context, string) (ChainExecution, error) {
	return ChainExecution{}, fmt.Errorf("noop")
}
func (Noop) GetSessionReserveSnapshot(context.Context, string) (SessionReserveSnapshot, error) {
	return SessionReserveSnapshot{
		AgentID: "agent_support_bot_1", MaxPerTransfer: "100000000", MaxTotalSpend: "500000000",
		ValidAfter: 1, ValidUntil: 9999999999, Status: "SESSION_AUTHORIZED",
	}, nil
}
func (Noop) MarkStaleUserOpsReconciliationRequired(context.Context, time.Time) (int, error) {
	return 0, nil
}
func (Noop) CommitBudgetOnChainExecution(context.Context, string, string) (string, bool, error) {
	return "", false, nil
}
