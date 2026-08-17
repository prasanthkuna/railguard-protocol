package reservation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestService(t *testing.T) (*Service, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	return NewWithClient(redis.NewClient(&redis.Options{Addr: mr.Addr()})), mr
}

func TestReserveIdempotencyScopedPerSession(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	a, err := svc.Reserve(ctx, "sess_a", "idem_1", "100", "500", time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.Reserve(ctx, "sess_b", "idem_1", "100", "500", time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("expected different reservation ids across sessions")
	}
	same, err := svc.Reserve(ctx, "sess_a", "idem_1", "100", "500", time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if a != same {
		t.Fatalf("expected idempotent replay, got %s vs %s", a, same)
	}
}

func TestReserveRejectsInvalidAmount(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	if _, err := svc.Reserve(ctx, "sess_a", "idem_1", "0", "500", time.Minute, time.Hour); err == nil {
		t.Fatal("expected invalid amount error")
	}
	if _, err := svc.Reserve(ctx, "sess_a", "idem_2", "", "500", time.Minute, time.Hour); err == nil {
		t.Fatal("expected invalid amount error")
	}
}

func TestReserveRejectsEmptyIdempotencyKey(t *testing.T) {
	svc, _ := newTestService(t)
	_, err := svc.Reserve(context.Background(), "sess_a", "  ", "100", "500", time.Minute, time.Hour)
	if err == nil || !strings.Contains(err.Error(), "idempotencyKey") {
		t.Fatalf("expected idempotency error, got %v", err)
	}
}

func TestReserveLargeIntegerAmounts(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()
	large := "9007199254740993" // > 2^53
	max := "9007199254740999"
	resID, err := svc.Reserve(ctx, "sess_large", "idem_large", large, max, time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(resID, "res_") {
		t.Fatalf("unexpected reservation id %s", resID)
	}
	_, err = svc.Reserve(ctx, "sess_large", "idem_overflow", "7", max, time.Minute, time.Hour)
	if err == nil || !strings.Contains(err.Error(), "BUDGET_DENIED") {
		t.Fatalf("expected budget denied, got %v", err)
	}
}

func TestFreezeReservationPreventsExpirySweep(t *testing.T) {
	svc, mr := newTestService(t)
	ctx := context.Background()
	resID, err := svc.Reserve(ctx, "sess_freeze", "idem_freeze", "100", "500", time.Second, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.FreezeReservation(ctx, resID); err != nil {
		t.Fatal(err)
	}
	mr.FastForward(time.Second * 2)
	if err := svc.SweepExpired(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	if got, _ := mr.Get("reserve:sess_freeze"); got != "100" {
		t.Fatalf("expected frozen reservation to remain reserved, got %s", got)
	}
}

func TestReleaseAndCommitReservation(t *testing.T) {
	svc, mr := newTestService(t)
	ctx := context.Background()
	resID, err := svc.Reserve(ctx, "sess_rel", "idem_rel", "200", "500", time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := mr.Get("reserve:sess_rel"); got != "200" {
		t.Fatalf("expected reserved 200, got %s", got)
	}
	if err := svc.ReleaseReservation(ctx, resID); err != nil {
		t.Fatal(err)
	}
	if got, _ := mr.Get("reserve:sess_rel"); got != "0" {
		t.Fatalf("expected released reserve 0, got %s", got)
	}

	resID2, err := svc.Reserve(ctx, "sess_rel", "idem_rel2", "150", "500", time.Minute, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.CommitReservation(ctx, resID2); err != nil {
		t.Fatal(err)
	}
	if _, err := mr.Get("resmeta:" + resID2); err == nil {
		t.Fatal("expected reservation metadata removed on commit")
	}
	if got, _ := mr.Get("reserve:sess_rel"); got != "150" {
		t.Fatalf("expected committed reserve 150, got %s", got)
	}
}
