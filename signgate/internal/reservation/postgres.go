package reservation

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/railguard/signgate/internal/store"
)

// PostgresService uses PostgreSQL as the live budget authority (v4 §5–6).
type PostgresService struct {
	store store.Repository
}

func NewPostgres(st store.Repository) *PostgresService {
	return &PostgresService{store: st}
}

func (p *PostgresService) Ping(ctx context.Context) error {
	_, err := p.store.GetWatcherBlockCursor(ctx)
	if err != nil {
		return fmt.Errorf("postgres budget authority unavailable")
	}
	return nil
}

func (p *PostgresService) Reserve(
	ctx context.Context,
	sessionID, idempotencyKey, amountAtomic, maxTotalSpend string,
	_, _ time.Duration,
) (string, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return "", fmt.Errorf("idempotencyKey required")
	}
	if existing, err := p.store.GetReservationIDByIdempotency(ctx, idempotencyKey); err == nil && existing != "" {
		return existing, nil
	}
	amount, ok := new(big.Int).SetString(amountAtomic, 10)
	if !ok || amount.Sign() <= 0 {
		return "", fmt.Errorf("invalid amountAtomic")
	}
	maxTotal, ok := new(big.Int).SetString(maxTotalSpend, 10)
	if !ok || maxTotal.Sign() <= 0 {
		return "", fmt.Errorf("invalid maxTotalSpend")
	}
	reservationID, err := p.store.ReserveSessionBudget(ctx, sessionID, idempotencyKey, amount, maxTotal)
	if err != nil {
		return "", err
	}
	return reservationID, nil
}

func (p *PostgresService) CommitReservation(ctx context.Context, reservationID string) error {
	return p.store.UpdateReservationStatus(ctx, reservationID, "BUDGET_COMMITTED")
}

func (p *PostgresService) ReleaseReservation(ctx context.Context, reservationID string) error {
	return p.store.UpdateReservationStatus(ctx, reservationID, "BUDGET_RELEASED")
}

func (p *PostgresService) FreezeReservation(_ context.Context, _ string) error {
	return nil
}
