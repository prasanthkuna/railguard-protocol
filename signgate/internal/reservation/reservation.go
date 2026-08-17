package reservation

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	rdb *redis.Client
}

func New(addr string) *Service {
	return &Service{rdb: redis.NewClient(&redis.Options{Addr: addr})}
}

func NewWithClient(rdb *redis.Client) *Service {
	return &Service{rdb: rdb}
}

func (s *Service) Ping(ctx context.Context) error {
	return s.rdb.Ping(ctx).Err()
}

func idemKey(sessionID, idempotencyKey string) (string, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	sessionID = strings.TrimSpace(sessionID)
	if idempotencyKey == "" {
		return "", fmt.Errorf("idempotencyKey required")
	}
	if sessionID == "" {
		return "", fmt.Errorf("sessionId required")
	}
	return fmt.Sprintf("idem:%s:%s", sessionID, idempotencyKey), nil
}

func sessionKey(sessionID string) string {
	return "reserve:" + sessionID
}

func lockKey(sessionID string) string {
	return "reserve-lock:" + sessionID
}

func reservationMetaKey(reservationID string) string {
	return "resmeta:" + reservationID
}

func expiryIndexKey() string { return "reserve-expiry" }

// expiryMember encodes durable release data independent of metadata TTL.
func expiryMember(reservationID, sessionID, amountAtomic string) string {
	return reservationID + "|" + sessionID + "|" + amountAtomic
}

func parseExpiryMember(member string) (reservationID, sessionID, amountAtomic string, ok bool) {
	parts := strings.SplitN(member, "|", 3)
	if len(parts) != 3 {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

// SweepExpired releases reservations whose pre-submit TTL has elapsed (H-01).
func (s *Service) SweepExpired(ctx context.Context, now time.Time) error {
	members, err := s.rdb.ZRangeByScore(ctx, expiryIndexKey(), &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", now.Unix()),
	}).Result()
	if err != nil {
		return err
	}
	for _, member := range members {
		if reservationID, sessionID, amountAtomic, ok := parseExpiryMember(member); ok {
			if err := s.releaseAmount(ctx, sessionID, amountAtomic); err != nil {
				return err
			}
			_ = s.rdb.Del(ctx, reservationMetaKey(reservationID)).Err()
		} else if err := s.ReleaseReservation(ctx, member); err != nil {
			return err
		}
		if err := s.rdb.ZRem(ctx, expiryIndexKey(), member).Err(); err != nil {
			return err
		}
	}
	return nil
}

// Reserve atomically checks session budget using string big-int math in Go (safe above 2^53).
func (s *Service) Reserve(
	ctx context.Context,
	sessionID, idempotencyKey, amountAtomic, maxTotalSpend string,
	preSubmitTTL, sessionAggregateTTL time.Duration,
) (string, error) {
	if err := s.SweepExpired(ctx, time.Now()); err != nil {
		return "", err
	}
	idem, err := idemKey(sessionID, idempotencyKey)
	if err != nil {
		return "", err
	}
	amount, ok := new(big.Int).SetString(amountAtomic, 10)
	if !ok || amount.Sign() <= 0 {
		return "", fmt.Errorf("invalid amountAtomic")
	}
	maxTotal, ok := new(big.Int).SetString(maxTotalSpend, 10)
	if !ok || maxTotal.Sign() <= 0 {
		return "", fmt.Errorf("invalid maxTotalSpend")
	}

	if existing, err := s.rdb.Get(ctx, idem).Result(); err == nil && existing != "" {
		return existing, nil
	}

	lock := lockKey(sessionID)
	acquired, err := s.rdb.SetNX(ctx, lock, "1", 10*time.Second).Result()
	if err != nil {
		return "", err
	}
	if !acquired {
		return "", fmt.Errorf("BUDGET_BUSY")
	}
	defer s.rdb.Del(ctx, lock)

	currentStr, err := s.rdb.Get(ctx, sessionKey(sessionID)).Result()
	if err == redis.Nil {
		currentStr = "0"
	} else if err != nil {
		return "", err
	}
	current, ok := new(big.Int).SetString(currentStr, 10)
	if !ok {
		current = big.NewInt(0)
	}
	next := new(big.Int).Add(current, amount)
	if next.Cmp(maxTotal) > 0 {
		return "", fmt.Errorf("BUDGET_DENIED")
	}

	reservationID := "res_" + uuid.NewString()
	ttlSec := int(preSubmitTTL.Seconds())
	if ttlSec <= 0 {
		ttlSec = 300
	}
	if sessionAggregateTTL <= 0 {
		sessionAggregateTTL = 24 * time.Hour
	}

	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, sessionKey(sessionID), next.String(), sessionAggregateTTL)
	pipe.Set(ctx, idem, reservationID, time.Duration(ttlSec)*time.Second)
	metaTTL := time.Duration(ttlSec*2) * time.Second
	pipe.Set(ctx, reservationMetaKey(reservationID), sessionID+"|"+amountAtomic, metaTTL)
	pipe.ZAdd(ctx, expiryIndexKey(), redis.Z{
		Score:  float64(time.Now().Add(time.Duration(ttlSec) * time.Second).Unix()),
		Member: expiryMember(reservationID, sessionID, amountAtomic),
	})
	if _, err := pipe.Exec(ctx); err != nil {
		return "", err
	}
	return reservationID, nil
}

// FreezeReservation removes a reservation from the pre-submit expiry sweep
// after a UserOp has been broadcast while keeping metadata for commit/release.
func (s *Service) FreezeReservation(ctx context.Context, reservationID string) error {
	meta, err := s.rdb.Get(ctx, reservationMetaKey(reservationID)).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}

	pipe := s.rdb.TxPipeline()
	members, err := s.rdb.ZRange(ctx, expiryIndexKey(), 0, -1).Result()
	if err == nil {
		for _, member := range members {
			if id, _, _, ok := parseExpiryMember(member); ok && id == reservationID {
				pipe.ZRem(ctx, expiryIndexKey(), member)
				break
			}
		}
	}
	pipe.Set(ctx, reservationMetaKey(reservationID), meta, 7*24*time.Hour)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Service) CommitReservation(ctx context.Context, reservationID string) error {
	pipe := s.rdb.TxPipeline()
	pipe.Del(ctx, reservationMetaKey(reservationID))
	members, err := s.rdb.ZRange(ctx, expiryIndexKey(), 0, -1).Result()
	if err == nil {
		for _, member := range members {
			if id, _, _, ok := parseExpiryMember(member); ok && id == reservationID {
				pipe.ZRem(ctx, expiryIndexKey(), member)
				break
			}
		}
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Service) ReleaseReservation(ctx context.Context, reservationID string) error {
	meta, err := s.rdb.Get(ctx, reservationMetaKey(reservationID)).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	parts := strings.SplitN(meta, "|", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid reservation metadata")
	}
	if err := s.releaseAmount(ctx, parts[0], parts[1]); err != nil {
		return err
	}
	pipe := s.rdb.TxPipeline()
	pipe.Del(ctx, reservationMetaKey(reservationID))
	members, err := s.rdb.ZRange(ctx, expiryIndexKey(), 0, -1).Result()
	if err == nil {
		for _, member := range members {
			if id, _, _, ok := parseExpiryMember(member); ok && id == reservationID {
				pipe.ZRem(ctx, expiryIndexKey(), member)
				break
			}
		}
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *Service) releaseAmount(ctx context.Context, sessionID, amountAtomic string) error {
	amount, ok := new(big.Int).SetString(amountAtomic, 10)
	if !ok {
		return fmt.Errorf("invalid reserved amount")
	}

	lock := lockKey(sessionID)
	acquired, err := s.rdb.SetNX(ctx, lock, "1", 10*time.Second).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return fmt.Errorf("BUDGET_BUSY")
	}
	defer s.rdb.Del(ctx, lock)

	currentStr, err := s.rdb.Get(ctx, sessionKey(sessionID)).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	current, ok := new(big.Int).SetString(currentStr, 10)
	if !ok {
		return fmt.Errorf("invalid session reserve state")
	}
	next := new(big.Int).Sub(current, amount)
	if next.Sign() < 0 {
		next = big.NewInt(0)
	}
	return s.rdb.SetArgs(ctx, sessionKey(sessionID), next.String(), redis.SetArgs{KeepTTL: true}).Err()
}
