package store_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/railguard/signgate/internal/store"
)

func TestStoreDecisionAndReceiptRoundTrip(t *testing.T) {
	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		url = "postgres://railguard:railguard@localhost:5432/railguard?sslmode=disable"
	}
	ctx := context.Background()
	st, err := store.New(ctx, url)
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	defer st.Close()

	decisionID := "dec_integration_test_" + time.Now().UTC().Format("20060102150405.000000000")
	intentHash := "0x96734b72ae38ed4166ef08446996462802dd2c7577fe608fdc0c6371a571d150"
	if err := st.SaveDecision(ctx, decisionID, intentHash, "ALLOW", []string{"WITHIN_LIMITS"}, "0xabc"); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]string{"decisionId": decisionID, "decision": "ALLOW"})
	if err := st.SaveReceipt(ctx, decisionID, intentHash, "", "0xreceipt", payload, "0xsig", "test-key"); err != nil {
		t.Fatal(err)
	}

	decision, receiptHash, signature, raw, err := st.GetReceipt(ctx, decisionID)
	if err != nil {
		t.Fatal(err)
	}
	if decision != "ALLOW" || receiptHash != "0xreceipt" || signature != "0xsig" {
		t.Fatalf("unexpected receipt fields: %s %s %s", decision, receiptHash, signature)
	}
	if len(raw) == 0 {
		t.Fatal("expected receipt payload")
	}
}

func TestChainExecutionBySessionID(t *testing.T) {
	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		url = "postgres://railguard:railguard@localhost:5432/railguard?sslmode=disable"
	}
	ctx := context.Background()
	st, err := store.New(ctx, url)
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	defer st.Close()
	if err := st.EnsureExecutionDigestSchema(ctx); err != nil {
		t.Fatalf("schema migration: %v", err)
	}

	sessionID := "0x1111111111111111111111111111111111111111111111111111111111111111"
	exec := store.ChainExecution{
		Account:         "0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266",
		SessionID:       sessionID,
		NonceKey:        "424242",
		FrameSpend:      "50000000",
		TotalSpendAfter: "50000000",
		ExecutionDigest: "0xdeadbeef",
		BlockNumber:     12,
		TxHash:          "0xabc123",
		LogIndex:        1,
	}
	if _, _, err := st.RecordChainExecution(ctx, exec); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetChainExecutionBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if got.TxHash != exec.TxHash || got.FrameSpend != exec.FrameSpend {
		t.Fatalf("unexpected chain execution: %+v", got)
	}
}
