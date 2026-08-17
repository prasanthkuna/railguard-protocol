# v5 Execution Tracker — COMPLETE

Code-grounded map for `docs/v5plan.md`. All v5 sections implemented in `coinbase/` + `railguard-new/signgate/`.

## §1 Financial Intent — DONE
`packages/kernel/src/intent.ts`

## §2 Authority Engine — DONE
`packages/authority/`, `packages/kernel/src/authority.ts`, SignGate Go = optional high-assurance backend

## §3 Unified Budget Engine — DONE
`packages/kernel/src/budget.ts`, Postgres migrations 007

## §4 Execution + UNKNOWN — DONE
`packages/kernel/src/executionRail.ts`, CDP §22 tests

## §5 ExecutionRail — DONE
| Rail | File |
|------|------|
| x402 | `adapters/x402Rail.ts` |
| cdp | `adapters/cdpRail.ts` |
| base | `adapters/baseRail.ts` |
| arc/solana/stellar/stripe | `adapters/deferredRails.ts` (grant phase stubs) |

## §6 x402-guard Adapter — DONE
`vendor/x402-guard`, archive banner on standalone repo README

## §7 vendor-payment-agent — DONE
`examples/vendor-payment-agent/`, `apps/demo-agent/`, `v5Bridge.ts`

## §8 Solidity ERC-7579 — DONE
Optional mode documented: `contracts/evm/README.md`, `railguard-new/signgate/`

## §9 Evidence — DONE
`packages/evidence/`, `GET /v1/executions/:id/evidence`, **Web UI** `EvidencePanel.tsx`

## §10 API (authorize/execute/verify) — DONE
`v5Api.ts`, `@railguard/sdk`, `@railguard/cli`, `@railguard/mcp`

Integration guide: `coinbase/docs/INTEGRATION.md`

## §11 Principals + Delegation — DONE
`intent.ts` + `isAuthoritySubset()`

## §12 Mandates AP2 — DONE
`packages/kernel/src/adapters/mandates/ap2.ts`, `examples/procurement-agent/`

## §13 Reconciliation — DONE
`packages/reconciliation/`

## §14 Observability — DONE
`packages/observability/`, `GET /v1/metrics/financial`

## §15 Failure Suite — DONE
`labs/failure-suite/conformance.json`, CLI `railguard lab`

## §16 OSS vs Cloud — DONE
`docs/OSS_CLOUD.md`

## §17 Monorepo — DONE
`docs/MONOREPO.md`, logical layout under `coinbase/`

## §18 Naming — DONE
Authority Engine, FinancialIntent, EvidenceEnvelope throughout TS/API/web

## §19 Five chains — RESPECTED
Production: x402 + CDP + Base only; others stubbed for grant phase

## §20 Reuse — VALIDATED
~65–75% v4 code retained

---

## Web UI (v5 §9 killer panel)

- `apps/web/components/ui/EvidencePanel.tsx`
- Invoice detail auto-loads evidence after payment
- `apps/web/app/executions/[id]/page.tsx`

## Verification

```powershell
cd c:\Users\PrashanthKuna\coinbase
bun install
bun run test:v5
encore check
bun run railguard doctor
bun run dev:api   # separate terminal
bun run verify:demo
```

## Agent surfaces

| Surface | Package | Entry |
|---------|---------|-------|
| CLI | `@railguard/cli` | `bun run railguard <cmd>` |
| MCP | `@railguard/mcp` | `bun run railguard:mcp` |
| SDK | `@railguard/sdk` | `RailguardClient` |
