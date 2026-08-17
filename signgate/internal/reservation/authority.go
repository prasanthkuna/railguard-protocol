package reservation

import (
	"context"
	"time"
)

// Authority is the budget reservation backend (Redis or Postgres).
type Authority interface {
	Ping(ctx context.Context) error
	Reserve(
		ctx context.Context,
		sessionID, idempotencyKey, amountAtomic, maxTotalSpend string,
		preSubmitTTL, sessionAggregateTTL time.Duration,
	) (string, error)
	CommitReservation(ctx context.Context, reservationID string) error
	ReleaseReservation(ctx context.Context, reservationID string) error
	FreezeReservation(ctx context.Context, reservationID string) error
}
