# Railguard distribution copy

All public copy must retain the phrase “v0.1 reference implementation” and must not imply mainnet production readiness.

## X launch post

The payment succeeded.

The application lost the response.

Calling it “failed” releases the budget—and can let an AI agent pay twice.

Railguard v0.1 is a reference implementation for keeping truth intact across intent, policy, signing, execution, receipts, and reconciliation.

Watch the 30-second failure. Then reproduce it:

https://github.com/prasanthkuna/railguard-new/blob/master/docs/PORTFOLIO.md

Testnet evidence. Known production gaps documented. No mainnet funds.

## X technical thread

### Post 1

Agent payments fail on glue, not validators.

I audited the lifecycle around an agent-payment stack and found four correctness failures:

1. mutable ALLOW limits;
2. budget TOCTOU;
3. a post-broadcast “failed” lie;
4. FIFO reconciliation.

### Post 2

The dangerous case:

```text
payment broadcast
→ process loses certainty
→ application marks failed
→ budget is released
→ a second payment is authorized
```

After broadcast, “failed” is not a fact. The safe state is `unknown`.

### Post 3

The invariant:

```text
Intent → Policy → Session → Signature → Hook → Receipt → Reconcile
```

Identity and limits must remain bound across every boundary.

### Post 4

The correction:

- reserve budget atomically;
- freeze ambiguous outcomes as unknown;
- reconcile by execution digest;
- commit only after settlement fact.

### Post 5

Railguard v0.1 is deliberately scoped as a reference implementation.

It includes adversarial tests and public testnet evidence. It does not claim mainnet production readiness.

Run it, inspect it, and challenge the state machine:

https://github.com/prasanthkuna/railguard-new/blob/master/docs/PORTFOLIO.md

## LinkedIn launch

I built Railguard after finding a payment failure that ordinary validation does not solve:

A transaction can be broadcast successfully while the application believes it failed. If the application releases the budget at that moment, an AI agent may authorize the payment again.

The correction is an end-to-end lifecycle invariant—not another isolated validator:

`Intent → Policy → Session → Signature → Hook → Receipt → Reconcile`

Railguard v0.1 demonstrates:

- atomic budget reservation;
- honest unknown states after ambiguous broadcast;
- execution-digest reconciliation;
- pre-sign, session, on-chain, and receipt boundaries;
- adversarial tests and public testnet evidence.

It is a reference implementation, not a production-ready mainnet system. The open gaps are documented alongside the proof.

The three-minute proof and reproducible reviewer path:

https://github.com/prasanthkuna/railguard-new/blob/master/docs/PORTFOLIO.md

## YouTube — flagship

**Title:** The Payment Succeeded. The App Said It Failed. | Railguard

**Description:**

A stablecoin payment can be broadcast while an application loses certainty. Marking that operation failed can release the budget and make a second payment possible.

This proof film explains four agent-payment correctness failures and the Railguard lifecycle invariant used to address them.

Railguard v0.1 is a reference implementation with adversarial tests and testnet evidence. It is not production-ready for mainnet funds.

Reviewer path and evidence:
https://github.com/prasanthkuna/railguard-new/blob/master/docs/PORTFOLIO.md

Chapters:

- 00:00 The dangerous moment
- 00:28 Why payment glue fails
- 00:56 Four failure modes
- 01:28 Atomic reserve and unknown state
- 02:00 Execution-digest reconciliation
- 02:24 Public proof
- 02:46 Honest release boundary

## YouTube — walkthrough

**Title:** Reproduce Four AI-Agent Payment Failures in Five Minutes

**Description:**

Follow the Railguard reviewer path:

1. run the tests;
2. inspect the four failure modes;
3. reproduce the on-chain attack demo;
4. inspect Base Sepolia, Stellar testnet, APF, and CDP-flow evidence;
5. review the known production gaps.

Railguard v0.1 is a reference implementation. No mainnet funds.

https://github.com/prasanthkuna/railguard-new/blob/master/docs/PORTFOLIO.md

## GitHub release summary

Railguard v0.1-reference demonstrates policy-enforced safety for AI-agent stablecoin payments across three enforcement boundaries:

- pre-sign x402 policy and atomic authorization;
- SignGate sessions and on-chain ceilings;
- CDP execution and reconciliation.

The release includes adversarial tests for four audited failure modes, public testnet evidence, a five-minute reviewer path, and explicit production limitations.

Start here:

https://github.com/prasanthkuna/railguard-new/blob/master/docs/PORTFOLIO.md
