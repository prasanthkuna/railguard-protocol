# Railguard Portfolio — Start Here

> **Send recruiters and reviewers only this link:**  
> `https://github.com/prasanthkuna/railguard-protocol/blob/master/docs/PORTFOLIO.md`

[![v0.1-reference](https://img.shields.io/badge/release-v0.1--reference-blue)](./RELEASE_v0.1-reference.md)
[![ci](https://github.com/prasanthkuna/railguard-protocol/actions/workflows/ci.yml/badge.svg)](https://github.com/prasanthkuna/railguard-protocol/actions/workflows/ci.yml)
[![E2E proof](https://img.shields.io/badge/E2E-docker%20happy--path-green)](../../scripts/e2e-happy-path.ps1)
[![status](https://img.shields.io/badge/status-v0.1%20reference%20impl-lightgrey)](./RELEASE_v0.1-reference.md)

**One-line pitch:** Policy-enforced payment safety for AI-agent stablecoin payments — adversarial failure lab, pre-sign x402 policy, on-chain enforcement, CDP reconciliation.

**Status:** v0.1 **reference implementation** — live testnet evidence, CI green, **known production gaps documented**. Not production-ready for mainnet funds.

**Public evidence:** [evidence/](../evidence/) — Base Sepolia, Stellar testnet, APF-003/004, end-to-end CDP flow.

---

## Four-layer architecture

```text
Agent Payment Failure Lab   → tests the system (APF-001..006)
x402-guard                  → pre-payment authorization and budget reservation
railguard-protocol          → on-chain session enforcement (SignGate + hook)
railguard-cdp               → enterprise execution and reconciliation
```

| Layer | Repo | Role |
|-------|------|------|
| **Test** | [agent-payment-failure-lab](https://github.com/prasanthkuna/agent-payment-failure-lab) | Adversarial failure profiles |
| **Policy** | [x402-guard](https://github.com/prasanthkuna/x402-guard) | Pre-sign `authorizePayment` |
| **On-chain** | [railguard-protocol](https://github.com/prasanthkuna/railguard-protocol) | SignGate + ERC-7579 hook |
| **Execution** | [railguard-cdp](https://github.com/prasanthkuna/railguard-cdp) | CDP wallet + reconciliation |

---

## Reviewer path

| Time | Do this |
|------|---------|
| **2 min** | Read [What I built](#what-i-built) + [Failure modes](#what-failed-in-audit--what-i-fixed) below |
| **5 min** | `cd x402-guard && bun test packages/policy/src/authorize.test.ts` |
| **10 min** | `cd railguard-protocol/contracts && forge test --match-contract PrdDemo -vv` |
| **15 min** | Docker E2E: `docker compose up -d --build` → `apply-db-migrations.ps1` → `e2e-happy-path.ps1` |

---

## What I built

Frozen v0.1 scope (no feature creep):

```text
Pre-sign x402 policy
Atomic replay + budget authorization
Railguard SignGate / session path
On-chain hook enforcement proof
CDP invoice execution proof
Audit + reconciliation
Failure-mode tests
Portfolio docs
```

**Invariant:**

```text
Intent → Policy → Session → Signature → Hook → Receipt → Reconcile
```

> We separated policy intelligence from asset safety and hardened the glue: every ALLOW binds to immutable facts, every budget reserves atomically, and every terminal payment state converges to chain evidence.

---

## What failed in audit — what I fixed

| Bug | Fix | Proof |
|-----|-----|-------|
| Mutable ALLOW limits | Limits in intent hash | `intent_test.go` |
| Budget TOCTOU | `authorizePayment` reserve/commit | `authorize.test.ts` |
| Post-broadcast `failed` | `unknown` + reconciler | [APF-003 evidence](../evidence/apf-003/) |
| FIFO reconciliation | `executionDigest` | `e2e-happy-path.ps1` |

Full table: [FAILURE_MODES_FIXED.md](./FAILURE_MODES_FIXED.md)

---

## What is still open (honest)

- Full Postgres integration fault-injection at API boundaries
- Deep reorg rewind (confirmation depth is configurable)
- HSM/MPC for SignGate cosigners ([THREAT_MODEL.md](./THREAT_MODEL.md))
- CDP monorepo dependency audit cleanup ([railguard-cdp DEPENDENCY_AUDIT](https://github.com/prasanthkuna/railguard-cdp/blob/main/docs/DEPENDENCY_AUDIT.md))

**Not in v0.1:** Paymaster, Solana, multi-chain, dashboard, arbitrary routers, mainnet funds.

---

## 5-minute demo

```powershell
# x402 atomic budget (~2 min)
cd x402-guard
bun test packages/policy/src/authorize.test.ts

# On-chain attacks blocked (~3 min)
cd ..\railguard-protocol\contracts
forge test --match-contract PrdDemo -vv
```

Full loop (~15 min): see [RELEASE_v0.1-reference.md](./RELEASE_v0.1-reference.md)

---

## Three enforcement boundaries

| # | Boundary | Repo |
|---|----------|------|
| 1 | Pre-sign x402 policy | [x402-guard](https://github.com/prasanthkuna/x402-guard) |
| 2 | Session + on-chain ceiling | [railguard-protocol](https://github.com/prasanthkuna/railguard-protocol) |
| 3 | Invoice / CDP product | [railguard-cdp](https://github.com/prasanthkuna/railguard-cdp) |

**CDP vs hook:** CDP proves invoice workflow + broadcast reconciliation. Hook proves smart-account caps. v0.1 shares policy/audit primitives; full CDP→smart-account routing is v0.2+.

---

## Source of truth

| Question | Authority |
|----------|-----------|
| x402 payment allowed? | `x402-guard` `authorizePayment` store |
| On-chain spend allowed? | Hook + session config |
| CDP broadcast happened? | `broadcastedTxHash` |
| Transfer succeeded? | Receipt `status === success` |
| Audit trail? | Hash-chained audit + receipts |
| Ambiguous UI state? | `submitted` / `unknown` until reconciler |
| Reservation ↔ execution? | `executionDigest` |

---

## Architecture

[THREE_PROJECT_SYSTEM_DIAGRAM.md](./THREE_PROJECT_SYSTEM_DIAGRAM.md)

---

## OSS contribution

[UPSTREAM_CONTRIBUTION.md](./UPSTREAM_CONTRIBUTION.md) — one upstream PR/comment plan for x402-go.

---

## Release

Tag **`v0.1-reference`** on all three repos. Notes: [RELEASE_v0.1-reference.md](./RELEASE_v0.1-reference.md)

| Repo | CI |
|------|-----|
| railguard-protocol | [![ci](https://github.com/prasanthkuna/railguard-protocol/actions/workflows/ci.yml/badge.svg)](https://github.com/prasanthkuna/railguard-protocol/actions/workflows/ci.yml) |
| x402-guard | [![CI](https://github.com/prasanthkuna/x402-guard/actions/workflows/ci.yml/badge.svg)](https://github.com/prasanthkuna/x402-guard/actions/workflows/ci.yml) |
| railguard-cdp | PR checks on `main` |
