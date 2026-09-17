## Finding: OSS-001

- **Repo / path:** `coinbase/cdp-sdk` → `typescript/packages/cdp-sdk/src/actions/evm/swap/sendSwapTransaction.ts`, `sendSwapOperation.ts`
- **Invariant violated:** If CDP Swaps API returns `issues.balance` or `simulationIncomplete: true`, the SDK must not broadcast the swap transaction.
- **Failure condition:** User calls `createSwapQuote()` → receives quote with balance issue → calls `swapQuote.execute()` or `sendSwapTransaction({ swapQuote })` → transaction is broadcast → on-chain revert / wasted gas.
- **Repro steps (local):**
  1. Mock `createEvmSwapQuote` response with `liquidityAvailable: true`, `issues.balance.currentBalance < requiredBalance`, valid `transaction`.
  2. Call `sendSwapTransaction(client, { address, swapQuote })`.
  3. Observe `sendTransaction` is invoked (should throw before broadcast).
- **Duplicate search:** GitHub issues search `balance swap sendSwapTransaction simulationIncomplete` — no duplicate (Sep 4, 2026).
- **Severity:** P1 — guaranteed bad UX; examples/docs tell users to check manually but execute path does not enforce.
- **Company leverage:** Coinbase / CDP SDK
- **Evidence:** `findings/tests/swap-issue-guards.test.ts` (OSS-001, OSS-001b, control test for allowance)
- **Next:** Open GitHub issue with suggested fix mirroring allowance guard:

```typescript
if (swap.issues?.balance) {
  const { currentBalance, requiredBalance, token } = swap.issues.balance;
  throw new Error(
    `Insufficient token balance for swap. Current: ${currentBalance}, required: ${requiredBalance}, token: ${token}`,
  );
}
if (swap.issues?.simulationIncomplete) {
  throw new Error("Swap simulation incomplete; refusing to broadcast transaction");
}
```

**Railguard parallel:** Same class of bug as executing payment intents without checking reservation invariants — off-chain state says "don't proceed" but executor broadcasts anyway.
