# CDP SDK Phase 1 Archaeology — Sep 4, 2026

**Repo:** [coinbase/cdp-sdk](https://github.com/coinbase/cdp-sdk)  
**Map issue:** [#790](https://github.com/coinbase/cdp-sdk/issues/790) (permit2_data — not duplicated; adjacent paths investigated)  
**Leverage:** Coinbase Staff CAE application packet (Sep 26 target)

## State machine mapped

```text
createEvmSwapQuote (HTTP)
  → createSwapQuote (transform + attach execute())
    → sendSwapTransaction | sendSwapOperation
      → [optional] signEvmTypedData (Permit2)
      → sendTransaction | sendUserOperation
```

**External boundaries:** CDP Swaps API · CDP signEvmTypedData · chain RPC (via sendTransaction)

## Hypothesis matrix

| ID | Invariant | Result | Level |
|----|-----------|--------|-------|
| H1 | `sendSwapTransaction` rejects when `issues.balance` present | **VIOLATED** | 2 |
| H2 | `sendSwapTransaction` rejects when `simulationIncomplete` | **VIOLATED** | 2 |
| H3 | Permit2-required calldata never submitted unsigned | **VIOLATED** (adjacent to #790) | 2 |
| H4 | Python `QuoteSwapResult` preserves API `issues` | **VIOLATED** | 2 |
| H5 | #790 root cause in TS `createSwapQuote` parsing | Duplicate of #790 | kill |

## Findings

| Finding | File | Severity |
|---------|------|----------|
| OSS-001 | [OSS-001-swap-execute-missing-guards.md](./OSS-001-swap-execute-missing-guards.md) | P1 |
| OSS-003 | [OSS-003-python-issues-dropped.md](./OSS-003-python-issues-dropped.md) | P1 |

## Repro tests

```powershell
# TypeScript (4/4 pass)
cd c:\Users\PrashanthKuna\personal\web3\cdp-sdk\typescript\packages\cdp-sdk
pnpm test -- src/actions/evm/swap/swapIssueGuards.archaeology.test.ts

# Python (2/2 pass)
cd c:\Users\PrashanthKuna\personal\web3\cdp-sdk\python
pytest ..\..\findings\tests\test_swap_issue_guards.py -v
```

## Next

- Promote OSS-001 to GitHub issue (private if SECURITY.md applies — swap path is not custody-critical; public issue OK)
- Cross-link in Coinbase application packet
- Do **not** fix #790 — use as map only
