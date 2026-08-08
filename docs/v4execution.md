# v4 execution plan (code-grounded)

This document is the **execution order** for `v4plan.md`. It reflects what the code contains.

`v4plan.md` remains the architecture constitution.

---

## Phase status

| Phase | Scope | Status |
|-------|--------|--------|
| **P0** | CDP provider idempotency + request persist | **Done** |
| **P0** | SignGate Redis freeze on submit + watcher Redis commit | **Done** |
| **P1** | Extract `@railguard/kernel` + execution-driver protocol | **Done** |
| **P1** | Promote `PostgresGuardStateStore` to `@x402-guard/policy` | **Done** |
| **P1** | `@railguard/x402` + `@railguard/x402-policy` alias packages | **Done** |
| **P2** | Purchases / quotes / fulfilments / outbox (migration 010) | **Done** |
| **P2** | Wire domain into CDP API + outbox cron | **Done** (Step 2) |
| **P2** | Converge `DbGuardStateStore` → shared `PostgresGuardStateStore` | **Done** (Step 3) |
| **P2** | v4 §22 CDP drop-response recovery lab test | **Done** (Step 1) |
| **P2** | SignGate Postgres budget authority (`BUDGET_AUTHORITY=postgres`) | **Done** |
| **P2** | `RailguardVaultExecutor` + stub vault driver | **Done** |
| **P2** | KMS SignGate config stub (`KMS_SIGNER_*`) | **Done** |
| **P2** | Hierarchical budget scope columns on guard auths | **Done** (schema) |
| **P3** | Live vault CDP integration, full KMS, npm publish, repo rename | **Backlog** |

---

## P0 — CDP spine (`coinbase/`)

| Component | Location |
|-----------|----------|
| `execution_attempts` table | `apps/api/migrations/009_execution_attempts.up.sql` |
| Canonical CDP request + hash | `packages/cdp/src/cdpRequest.ts` |
| Persist + submit + recovery | `cdpExecutionSubmit.ts`, `executionAttempts.ts`, `providers.ts` |
| Tests | `cdpExecutionDriver.test.ts`, `cdpExecutionSubmit.test.ts` |

Live CDP calls pass `idempotencyKey` to `account.transfer()`. Recovery reuses the stored `provider_idempotency_key`.

---

## P0 — SignGate post-submit freeze (`railguard-new/`)

| Component | Location |
|-----------|----------|
| `FreezeReservation` | `signgate/internal/reservation/reservation.go` |
| Submit hook | `signgate/internal/api/server.go` |
| Watcher Redis commit | `signgate/internal/watcher/watcher.go` |

---

## P1 — Shared kernel (`coinbase/packages/kernel`)

Extracted pure payment logic (re-exported from `apps/api/*`):

- `lifecycle.ts` — state invariants
- `reconciliation.ts` — settlement transitions
- `correlation.ts` — execution correlation + guard input
- `cdpDriver.ts` — prepare/recovery helpers
- `executionDriver.ts` — `prepare` / `submit` / `observe` / `findByCorrelation` interfaces
- `domain.ts` — Purchase, Quote, Fulfilment, OutboxEvent types
- `vaultDriver.ts` — `StubVaultExecutionDriver` for CDP_VAULT_CALL

---

## P1 — Durable x402 store (`x402-guard/`)

| Component | Location |
|-----------|----------|
| `PostgresGuardStateStore` | `packages/policy/src/postgresStore.ts` |
| Encore adapter | `coinbase/apps/api/encoreGuardSqlAdapter.ts` → `x402GuardDbStore.ts` |
| Shared store tests | `packages/policy/src/postgresStore.test.ts` |
| Alias package | `packages/railguard-x402-policy` → `@railguard/x402-policy` |
| Middleware alias | `packages/railguard-x402` → `@railguard/x402` |

---

## P2 — Domain + outbox (`coinbase/`)

Migration `010_v4_domain_outbox.up.sql`:

- `purchases` — `UNIQUE (organization_id, business_idempotency_key)`
- `quotes` — versioned per purchase
- `fulfilments` — `UNIQUE (merchant_id, fulfilment_id)`
- `outbox_events` — transactional outbox table
- `settlement_observations` — finality facts (PROVISIONAL/SAFE/FINALIZED)
- `x402_guard_budget_authorizations.scope_type/scope_id` — hierarchical budget hook

Services (wired into `api.ts` + `reconcile.ts`):

- `purchaseService.ts` — `createPurchase`, `addQuote` (called from `createPaymentIntent`)
- `fulfilmentService.ts` — idempotent `recordFulfilment` (via `purchaseFulfilment.ts`)
- `outboxService.ts` — enqueue on purchase/fulfilment; `outboxPoller.ts` cron every 5m

`createPaymentIntent` now sets `payment_intents.purchase_id`. Confirmed settlement calls `completePurchaseFulfilmentForPaymentIntent`.

---

## P2 — SignGate Postgres budget authority

Set `BUDGET_AUTHORITY=postgres` to use PostgreSQL as live spend authority:

- `signgate/internal/reservation/postgres.go`
- `signgate/internal/store/store.go` → `ReserveSessionBudget`
- Default remains `redis` for backward compatibility

---

## P2 — Vault + KMS stubs

| Component | Location |
|-----------|----------|
| On-chain vault reference | `railguard-new/contracts/src/RailguardVaultExecutor.sol` |
| Stub driver | `packages/kernel/src/vaultDriver.ts` |
| KMS config | `signgate/internal/config/config.go` — `KMS_SIGNER_ENABLED`, `KMS_SIGNER_KEY_ID` |

Production KMS signing and live vault CDP calls are P3.

---

## Three systems (intentionally separate enforcement modes)

| Path | Money authority | Status |
|------|-----------------|--------|
| **CDP invoice** | Postgres + kernel driver | Production path |
| **7579 hook** | Redis or Postgres (`BUDGET_AUTHORITY`) | Configurable |
| **x402 library** | In-memory or Postgres store | Embedded or via coinbase |

---

## Steps 1–3 (completed)

| Step | What | Status |
|------|------|--------|
| **1 — Prove the spine** | §22 lab: first CDP call drops → retry reuses `provider_idempotency_key` → one attempt row | **Done** — `cdpSection22.test.ts`, `cdpRecoveryScenario.ts` |
| **2 — Wire domain** | Purchase + quote before intent; fulfilment after confirmed; outbox cron | **Done** — `api.ts`, `reconcile.ts`, `outboxPoller.ts` |
| **3 — Converge duplicates** | `DbGuardStateStore` wraps shared `PostgresGuardStateStore` | **Done** — `encoreGuardSqlAdapter.ts` |

**You still need to run migrations locally:**

```powershell
cd c:\Users\PrashanthKuna\coinbase\apps\api
encore run
```

Encore applies `009_execution_attempts` and `010_v4_domain_outbox` on startup. Confirm with `\dt purchases` / `\dt execution_attempts` in the Encore DB shell.

---

## Completion test (v4 §22)

Unit proofs:

```powershell
cd c:\Users\PrashanthKuna\coinbase
bun test apps/api/cdpSection22.test.ts apps/api/cdpExecutionDriver.test.ts

cd c:\Users\PrashanthKuna\x402-guard\packages\policy
npm test

cd c:\Users\PrashanthKuna\railguard-protocol\signgate
go test ./...
```

Full Encore E2E with lost CDP response in CI remains optional (requires `ENCORE_RUNTIME_LIB`).

---

## Step 4 — open source (Phase A + B complete)

| Item | Status |
|------|--------|
| `@railguard/x402-core`, `@railguard/x402-policy`, `@railguard/x402-receipts`, `@railguard/x402` packaged | Done — [x402-guard](https://github.com/prasanthkuna/x402-guard) |
| `@railguard/sdk` manifest + release CI | Done — tag `sdk-v0.1.0` publishes via `.github/workflows/release-sdk.yml` |
| npm scope `@railguard/*` (not `@x402-guard/*`) | Locked — create org + `NPM_TOKEN` before first publish |
| Repo rename `railguard-new` → `railguard-protocol` | Docs/CI updated; run `gh repo rename railguard-protocol` on GitHub when ready |

**Publish (after `@railguard` org + `NPM_TOKEN`):**

```powershell
# x402-guard
git tag v0.1.0 && git push origin v0.1.0

# railguard-protocol
git tag sdk-v0.1.0 && git push origin sdk-v0.1.0
```

See [x402-guard/docs/PUBLISH.md](https://github.com/prasanthkuna/x402-guard/blob/main/docs/PUBLISH.md).

---

## P3 backlog (do not start until needed)

- Live `CDP_VAULT_CALL` through CDP smart account
- KMS-backed SignGate signer implementation
- npm publish `@railguard/evidence` (future package)
- Full hierarchical budget enforcement across org/team/agent/merchant
- Observability trace propagation (v4 §18)

Keep v4 §29 do-not-build list.

---

## Key commands

```powershell
# Start Docker Desktop first, then:
cd c:\Users\PrashanthKuna\coinbase
$env:PAYMENT_MODE="demo"
$env:ALLOW_DEV_HEADER_AUTH="true"
encore run

# Separate terminal:
bun run dev:web

# Full API verification:
$env:RAILGUARD_BASE_URL="http://127.0.0.1:4000"
bun run verify:demo

# Unit tests:
bun test apps/api/cdpSection22.test.ts apps/api/cdpExecutionDriver.test.ts
```

Architecture target: [`v4plan.md`](./v4plan.md)
