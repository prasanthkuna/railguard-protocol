## Finding: OSS-003

- **Repo / path:** `coinbase/cdp-sdk` → `python/cdp/actions/evm/swap/create_swap_quote.py`, `types.py` (`QuoteSwapResult`)
- **Invariant violated:** Swap quote objects must surface API `issues` (balance, allowance, simulationIncomplete) so callers can fail closed before broadcast.
- **Failure condition:** API returns balance/allowance issues → `create_swap_quote()` builds `QuoteSwapResult` without any `issues` field → Python users cannot inspect problems; `send_swap_transaction()` proceeds if quote object is passed directly.
- **Repro steps (local):**
  1. Mock API response with `issues.balance` populated.
  2. Call `create_swap_quote(...)`.
  3. Assert returned model has no `issues` attribute.
  4. Call `send_swap_transaction` with that quote → transaction send invoked.
- **Duplicate search:** No existing issue found for Python issues stripping (Sep 4, 2026).
- **Severity:** P1 — strictly worse than TypeScript (TS at least exposes `issues` on quote; Python drops them entirely).
- **Company leverage:** Coinbase / CDP SDK
- **Evidence:** `findings/tests/test_swap_issue_guards.py`
- **Next:** File alongside OSS-001 or combine into single cross-language issue.

**Note:** TypeScript official examples (`examples/typescript/evm/swaps/*.ts`) manually validate `issues` before execute — confirming maintainers expect client-side checks that library does not enforce.
