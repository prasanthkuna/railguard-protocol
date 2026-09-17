# Failure modes (recruiter path)

**Status:** v0.1 reference — extends [FAILURE_MODES_FIXED.md](./FAILURE_MODES_FIXED.md) with external-system parallels.

## Railguard invariants

| Invariant | Violation symptom | Mitigation in this workspace |
|-----------|-------------------|------------------------------|
| Budget reserve before spend | Double spend on retry | x402 `authorizePayment` atomic reserve |
| Post-broadcast truth | UI shows `failed` while tx mined | `unknown` + reconciler (see [FAILURE_MODES_FIXED](./FAILURE_MODES_FIXED.md)) |
| Off-chain vs on-chain | Ledger diverges from chain | `executionDigest` reconciliation |
| Policy drift | Approve under policy A, execute under B | `policy_snapshot_hash` on approval |

## OSS parallel (cdp-sdk) — Level 2+ reproduced

**OSS-001:** CDP swap `sendSwapTransaction` checks **allowance** but not **balance** or **simulationIncomplete** — can broadcast doomed txs.

**OSS-003:** Python `QuoteSwapResult` drops API `issues` — callers cannot fail closed.

Evidence: [evidence/oss-cdp-sdk/](../evidence/oss-cdp-sdk/) · Repro tests in upstream archaeology suite.

**Recruiter link:** [PORTFOLIO.md](./PORTFOLIO.md)
