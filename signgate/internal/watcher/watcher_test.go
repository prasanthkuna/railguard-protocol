package watcher

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/railguard/signgate/internal/config"
	"github.com/railguard/signgate/internal/store"
	"github.com/rs/zerolog"
)

type captureChainExecStore struct {
	store.Noop
	got *store.ChainExecution
}

func (c *captureChainExecStore) RecordChainExecution(_ context.Context, exec store.ChainExecution) (string, bool, error) {
	copy := exec
	c.got = &copy
	return "", false, nil
}

func TestIngestLogParsesExecutionAllowed(t *testing.T) {
	account := common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	sessionID := crypto.Keccak256Hash([]byte("session"))
	nonceKey := big.NewInt(424242)
	executionDigest := crypto.Keccak256Hash([]byte("digest"))
	frameSpend := big.NewInt(50_000_000)
	totalSpend := big.NewInt(50_000_000)
	txHash := common.HexToHash("0xabc123def456")

	data := make([]byte, 96)
	copy(data[0:32], executionDigest.Bytes())
	copy(data[32:64], common.LeftPadBytes(frameSpend.Bytes(), 32))
	copy(data[64:96], common.LeftPadBytes(totalSpend.Bytes(), 32))

	lg := types.Log{
		Topics: []common.Hash{
			executionAllowedSig,
			common.BytesToHash(common.LeftPadBytes(account.Bytes(), 32)),
			sessionID,
			common.BytesToHash(common.LeftPadBytes(nonceKey.Bytes(), 32)),
		},
		Data:        data,
		BlockNumber: 42,
		TxHash:      txHash,
		Index:       7,
	}

	st := &captureChainExecStore{}
	r := &Reconciler{log: zerolog.Nop(), store: st, cfg: config.Config{}}
	if err := r.ingestLog(context.Background(), lg); err != nil {
		t.Fatal(err)
	}
	if st.got == nil {
		t.Fatal("expected RecordChainExecution to be called")
	}
	if st.got.ExecutionDigest != strings.ToLower(executionDigest.Hex()) {
		t.Fatalf("executionDigest got %s", st.got.ExecutionDigest)
	}
	if st.got.FrameSpend != "50000000" {
		t.Fatalf("frameSpend got %s", st.got.FrameSpend)
	}
}

func TestIngestLogSkipsMalformedTopics(t *testing.T) {
	r := &Reconciler{log: zerolog.Nop(), store: store.NewNoop(), cfg: config.Config{}}
	if err := r.ingestLog(context.Background(), types.Log{Topics: []common.Hash{{}}}); err != nil {
		t.Fatal(err)
	}
}

func TestExecutionAllowedTopicHash(t *testing.T) {
	want := crypto.Keccak256Hash([]byte("ExecutionAllowed(address,bytes32,uint192,bytes32,uint256,uint256)"))
	if executionAllowedSig != want {
		t.Fatalf("topic mismatch got %s want %s", executionAllowedSig.Hex(), want.Hex())
	}
}

func TestComputeSafeHeadWaitsForConfirmationDepth(t *testing.T) {
	safe, ready := computeSafeHead(2, 3)
	if ready || safe != 0 {
		t.Fatalf("head below confirm depth should not be ready: safe=%d ready=%v", safe, ready)
	}
	safe, ready = computeSafeHead(5, 3)
	if !ready || safe != 2 {
		t.Fatalf("safe head = head - confirm: got safe=%d ready=%v", safe, ready)
	}
	safe, ready = computeSafeHead(10, 0)
	if !ready || safe != 10 {
		t.Fatalf("confirm=0 uses head: got safe=%d ready=%v", safe, ready)
	}
}
