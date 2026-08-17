I would redesign Railguard around one rule:

**the codebase should model a financial action, not a blockchain transaction.**

Today your invariant is roughly:

`Intent → Policy → Session → Signature → Hook → Receipt → Reconcile`

That was excellent for proving payment correctness. But for the new company, `Session`, `Signature`, `Hook`, CDP and x402 are implementation details. The public product needs a simpler invariant:

```text
Financial Intent
      ↓
Authorize
      ↓
Reserve Authority
      ↓
Execute
      ↓
Observe
      ↓
Reconcile
      ↓
Evidence
```

Your current repos already contain most of those ingredients. `railguard-protocol` has SignGate, smart-account enforcement, reservations and receipts; `x402-guard` has budgets, replay protection and policy; `railguard-cdp` has execution, settlement verification and reconciliation.

## I would move to one core monorepo

```text
railguard/
│
├── apps/
│   ├── api/
│   ├── console/
│   ├── worker/
│   └── demo-agent/
│
├── packages/
│   ├── intent/
│   ├── policy/
│   ├── authority/
│   ├── execution/
│   ├── reconciliation/
│   ├── evidence/
│   ├── sdk/
│   └── observability/
│
├── adapters/
│   ├── x402/
│   ├── base/
│   ├── arc/
│   ├── solana/
│   ├── stellar/
│   ├── stripe/
│   ├── cdp/
│   └── wallet/
│
├── contracts/
│   └── evm/
│
├── labs/
│   └── failure-suite/
│
├── examples/
│   ├── x402-agent/
│   ├── procurement-agent/
│   └── api-buying-agent/
│
└── docs/
```

That architectural change is more important than changing programming language or framework.

---

# 1. `packages/intent` becomes the heart of Railguard

This is the biggest missing abstraction.

Right now the code has payment-specific intent builders, x402 request data and invoice/payment objects spread across systems. I would create one canonical `FinancialIntent`.

Something like:

```ts
type FinancialIntent = {
  id: string;

  principal: {
    organizationId: string;
    actorId: string;
    actorType: "agent" | "human" | "service";
  };

  action: {
    type: "pay" | "purchase" | "transfer";
    purpose?: string;
  };

  counterparty: {
    id?: string;
    address?: string;
    domain?: string;
  };

  value: {
    amount: string;
    asset: string;
    maxAmount?: string;
  };

  constraints: {
    expiresAt: string;
    network?: string;
  };

  context?: Record<string, unknown>;

  idempotencyKey: string;
};
```

Notice what's deliberately **not there**:

```text
CDP
Base
x402
EIP-712
ERC-4337
Solana
Arc
```

Those belong later.

This allows the exact same Railguard policy engine to authorize:

```text
$0.03 x402 API request
$500 USDC vendor payment
$29 SaaS subscription
$3,000 procurement transaction
```

That abstraction is what makes Railguard a company rather than an x402 utility.

---

# 2. Kill `SignGate` as a product concept

Keep the code.

Kill the name from the external architecture.

`SignGate` currently combines OPA policy, signing, Redis reservations, Postgres audit and session handling.

I would split it conceptually into:

```text
Authority Service
├ Policy Decision
├ Budget Reservation
├ Approval
├ Credential Authority
└ Authorization Grant
```

The new primitive is:

### `AuthorizationGrant`

```ts
type AuthorizationGrant = {
  grantId: string;
  intentId: string;

  decision: "allow" | "deny" | "approval_required";

  limits: {
    reservedAmount: string;
    asset: string;
  };

  policyVersion: string;

  validUntil: string;

  executionConstraints: {
    recipients?: string[];
    networks?: string[];
    rails?: string[];
  };

  evidenceHash: string;
};
```

Then EIP-712 signing is just one way of materializing a grant.

That's much cleaner.

Today:

```text
SignGate signs transaction
```

Tomorrow:

```text
Railguard grants financial authority
```

Huge difference.

---

# 3. Unify the budget engine from x402-guard

Your x402 repo already has one of the most valuable pieces:

* durable budget state
* atomic reservation
* rolling limits
* replay protection
* settlement reconciliation

But it is wrongly scoped under x402.

Move that into:

```text
packages/authority/budget
```

and make x402 consume it.

Then the hierarchy becomes:

```text
Organization
   ↓
Department
   ↓
Agent
   ↓
Task
   ↓
Intent
```

Example:

```yaml
organization:
  monthly: 100000

team:
  marketing:
    monthly: 10000

agent:
  creative-agent:
    daily: 500

task:
  campaign-239:
    total: 80

transaction:
  max: 5
```

This is one of the areas I would invest heavily in because **hierarchical machine spending authority** is much more defensible than another wallet integration.

---

# 4. Replace payment state with an Execution state machine

Your CDP repo has excellent thinking here.

The invariant:

```text
No broadcast proof → release authorization
Broadcast occurred → freeze authorization
Confirmed → commit authorization
Reverted → release authorization
Mismatch/uncertain → reconciliation required
```

is one of Railguard's strongest pieces.

I would generalize it.

Canonical lifecycle:

```text
CREATED
   ↓
EVALUATING
   ↓
AUTHORIZED
   ↓
RESERVED
   ↓
EXECUTING
   ↓
SUBMITTED
   ↓
SETTLED
```

But failure states are first-class:

```text
DENIED
APPROVAL_REQUIRED
FAILED_SAFE
UNKNOWN
DISPUTED
REVERSED
EXPIRED
```

Most payment libraries treat uncertainty like an exception.

Railguard should treat:

# `UNKNOWN`

as a first-class financial state.

That's a significant differentiator.

For example:

```ts
switch (execution.status) {
  case "UNKNOWN":
    freeze(grant);
    enqueueReconciliation();
    alertFundsAtRisk();
    break;
}
```

Never:

```ts
catch {
  retry();
}
```

That tiny difference is part of what makes this credible financial infrastructure.

---

# 5. Create a proper execution adapter interface

Every ecosystem grant can fund an adapter without contaminating core.

Something like:

```ts
interface ExecutionRail {
  prepare(intent: FinancialIntent, grant: AuthorizationGrant):
    Promise<PreparedExecution>;

  execute(prepared: PreparedExecution):
    Promise<ExecutionSubmission>;

  observe(submission: ExecutionSubmission):
    Promise<ExecutionObservation>;

  reconcile(executionId: string):
    Promise<SettlementResult>;
}
```

Then:

```text
X402Rail
CDPRail
BaseSmartAccountRail
ArcRail
SolanaRail
StellarRail
StripeRail
```

All implement the same interface.

Now Circle can fund:

```text
adapters/arc
```

Solana can fund:

```text
adapters/solana
```

Stellar can fund:

```text
adapters/stellar
```

Base can fund:

```text
adapters/base
```

while **none of them own Railguard's architecture**.

That's exactly how I would structure it for the grant strategy we discussed.

---

# 6. `x402-guard` becomes an adapter, not a separate product

Today you expose:

```ts
createX402Guard()
withSpendingPolicy()
```

and x402 has its own packages for core, policy and receipts.

I'd collapse most of that duplication.

Future developer code:

```ts
import { Railguard } from "@railguard/sdk";
import { x402 } from "@railguard/x402";

const guard = new Railguard({
  agent: "research-agent"
});

await guard.execute({
  intent,
  rail: x402()
});
```

Or even:

```ts
await railguard.pay({
  amount: "0.05",
  asset: "USDC",
  recipient,
  purpose: "purchase API response"
});
```

Railguard chooses the execution rail.

Your current x402 functionality remains valuable.

It simply moves down one architectural layer.

---

# 7. `railguard-cdp` should disappear as a product

I would preserve virtually all useful code but eliminate the standalone product identity.

Its strongest modules become:

```text
packages/execution
packages/reconciliation
adapters/cdp
packages/evidence
```

The current invoice interface is useful as an **example application**, not Railguard itself.

Move it to:

```text
examples/vendor-payment-agent
```

or:

```text
apps/demo-agent
```

This cleans up one major confusion.

Today a reviewer has to understand:

```text
Railguard
x402-guard
PreBroadcast
CDP
three boundaries
four-layer stack
```

That's intellectually interesting.

Terrible startup onboarding.

Future reviewer should understand in 20 seconds:

```text
Agent
   ↓
Railguard
   ↓
payment rails
```

Done.

---

# 8. Keep the Solidity contracts, but demote them

Your `RailguardExecutionHook`, `RailguardAccountAdapter` and `RailguardSessionValidator` are technically valuable. The current implementation enforces token, recipient, spend caps, batch limits, expiry and replay constraints on-chain.

But don't make smart accounts mandatory.

Architecture:

```text
                         Railguard
                            │
                 Authorization Grant
                            │
           ┌────────────────┼────────────────┐
           │                │                │
      software gate     wallet policy    on-chain hook
                                           │
                                    hard enforcement
```

On-chain enforcement becomes:

### optional high-assurance mode

That's attractive for:

enterprise treasury
large agent budgets
institutional transactions
high-risk agents.

Meanwhile a developer spending $2/day through x402 doesn't need to deploy contracts.

This dramatically reduces adoption friction.

---

# 9. Evidence becomes a first-class product

Right now receipts/audit evidence exist in several places:

* hash-chained events
* signed receipts
* x402 receipts
* watcher observations
* settlement evidence

I would unify them into:

```text
EvidenceEnvelope
```

Example:

```ts
type EvidenceEnvelope = {
  intent: Hash;
  policyDecision: Hash;
  authorizationGrant: Hash;

  execution: {
    provider: string;
    submissionId?: string;
    txHash?: string;
  };

  settlement: {
    status: SettlementStatus;
    observedAt?: string;
  };

  policyVersion: string;
  sequence: number;
  previousHash?: string;

  signature: string;
};
```

Then give it a killer API:

```text
GET /v1/executions/{id}/evidence
```

And UI:

```text
WHY WAS THIS PAYMENT ALLOWED?
```

Click.

```text
Agent        research-bot-3
Task         competitor research
Requested    $3.20
Budget       $50/day
Merchant     Allowed
Policy       v17
Decision     ALLOW
Rail         x402/Base
Settlement   VERIFIED
Evidence     VALID
```

This is where Visa/Mastercard/enterprise conversations become much stronger.

---

# 10. Build the API around three verbs

I would resist exposing 15 backend concepts.

Public API:

```text
authorize()
execute()
verify()
```

Possibly convenience:

```text
pay()
```

### Authorize

```http
POST /v1/intents
POST /v1/intents/{id}/authorize
```

### Execute

```http
POST /v1/intents/{id}/execute
```

### Verify

```http
GET /v1/executions/{id}
GET /v1/executions/{id}/evidence
```

Internal services can remain complex.

External mental model should be tiny.

---

# 11. Introduce principals properly

This becomes critical for agent finance.

Don't just have:

```text
agentId
wallet
```

Create:

```text
Principal
```

Possible principal types:

```text
organization
human
agent
service
workflow
```

And delegation:

```text
Kuna / Company
       ↓
procurement-agent
       ↓
research-subagent
```

Authority propagates downward but can only become narrower.

Mathematically:

```text
child authority ⊆ parent authority
```

Never greater.

So:

```text
Company
$100K/month

Procurement agent
$10K/month

Laptop-buying task
$2K

Research subagent
$20
```

This is the beginning of Railguard's real moat.

---

# 12. Add mandates, but don't invent another standard

A future request could carry:

```ts
{
  principal,
  intent,
  authority,
  constraints,
  evidence
}
```

Railguard should later interoperate with AP2/agent commerce standards rather than defining an incompatible universe.

Your x402 repo already has AP2 mandate gating on the roadmap.

I'd convert that from an x402 feature into:

```text
adapters/mandates/ap2
```

Railguard's job is:

```text
external mandate
      ↓
normalize
      ↓
FinancialIntent
      ↓
Railguard policy
```

---

# 13. Build reconciliation as its own engine

This deserves far more architectural importance than it currently appears to have.

```text
packages/reconciliation
```

should support providers independently.

Interface:

```ts
interface Reconciler {
  observe(execution: Execution): Promise<Observation[]>;
  determine(observations: Observation[]): SettlementDecision;
}
```

And critically:

```text
one execution ≠ one RPC answer
```

Eventually:

```text
RPC A
RPC B
indexer
provider receipt
merchant acknowledgement
```

feed into a settlement decision.

This is how you solve one of the current known limitations, where the CDP path relies on a single RPC and lacks robust reorg/quorum handling.

For high value:

```text
SETTLED only after required evidence threshold
```

That is institutional-grade thinking.

---

# 14. Observability should use financial metrics

Not:

```text
HTTP 500 count
CPU
request latency
```

only.

Railguard's SRE model should expose:

```text
funds_at_risk
authorization_hold_value
unknown_execution_count
unknown_execution_value
duplicate_prevented_value
policy_denied_value
reconciliation_age
settlement_latency
budget_utilization
manual_approval_value
```

Those are metrics fintech operators care about.

This also strengthens both the product and your backend/platform story.

---

# 15. Turn your failure lab into a weapon

Don't throw away the adversarial work.

Create:

```text
labs/railguard-failure-suite
```

Scenarios:

```text
provider timeout after broadcast
duplicate agent request
API retry storm
stale policy
wallet response lost
chain reorg
wrong recipient
wrong amount
replay
budget race
concurrent subagents
merchant mismatch
partial batch execution
```

Run every adapter through the same conformance suite.

Example:

```text
Railguard Certified Executor

✓ replay safe
✓ idempotent
✓ timeout safe
✓ settlement aware
✓ budget atomic
✓ evidence complete
```

Now the failure lab becomes an ecosystem asset.

Potential grant story:

> **Open conformance suite for safe agent payments.**

While the commercial service remains Railguard Cloud.

That's a very good open-source/commercial boundary.

---

# 16. What stays open source vs commercial

I would make this explicit now.

### Open source

```text
FinancialIntent schema
SDK
adapter interfaces
x402 adapter
chain adapters
execution state machine
evidence schema
failure/conformance suite
Solidity contracts
local policy engine
```

### Railguard Cloud

```text
hosted authority service
durable hierarchical budgets
enterprise policy management
managed reconciliation
multi-provider settlement quorum
approval workflows
key custody integrations
dashboard
alerts
analytics
compliance exports
organization/RBAC
long-term evidence storage
SLA
```

This is much more commercially sensible than putting every operational advantage into MIT packages.

---

# 17. Repository strategy

I would ultimately have:

```text
github.com/railguard/railguard
```

as the obvious front door.

Then maybe:

```text
railguard/spec
railguard/examples
```

only if they become sufficiently large.

I would **archive but not delete**:

```text
x402-guard
railguard-cdp
```

with a banner:

> This project has moved into Railguard.

Preserve the stars/history/evidence.

Don't keep developing three independent architectures.

---

# 18. Naming cleanup

I'd also remove some legacy terms from the main developer path.

| Existing         | New                         |
| ---------------- | --------------------------- |
| SignGate         | Authority Engine            |
| reservation      | Authority Reservation       |
| x402 guard       | x402 Adapter                |
| receipt          | Evidence Envelope           |
| watcher          | Settlement Observer         |
| CDP payment      | CDP Executor                |
| hook             | Onchain Enforcement Adapter |
| userop submitted | Execution Submitted         |
| invoice          | application-level object    |
| decisionId       | authorizationId             |
| payment state    | execution state             |

Not because the old names are wrong.

Because the new terms describe **Railguard's domain**, rather than specific implementation machinery.

---

# The architecture I'd actually ship

```text
                         AI AGENT
                            │
                            │ FinancialIntent
                            ▼
                 ┌─────────────────────┐
                 │    RAILGUARD API    │
                 └──────────┬──────────┘
                            │
                  ┌─────────▼─────────┐
                  │ Authority Engine  │
                  │                  │
                  │ Identity         │
                  │ Policy           │
                  │ Budget           │
                  │ Delegation       │
                  │ Approval         │
                  └─────────┬─────────┘
                            │
                    AuthorizationGrant
                            │
                  ┌─────────▼─────────┐
                  │ Execution Router  │
                  └─────────┬─────────┘
                            │
       ┌────────┬───────────┼────────────┬────────┐
       ▼        ▼           ▼            ▼        ▼
     x402      Base        Arc         Solana   Stripe
       │        │           │            │        │
       └────────┴───────────┼────────────┴────────┘
                            ▼
                  Settlement Observer
                            │
                            ▼
                    Reconciliation
                            │
               ┌────────────┴───────────┐
               ▼                        ▼
            SETTLED                  UNKNOWN
               │                        │
               ▼                        ▼
            Evidence              freeze + retry
                                  reconciliation
```

That's Railguard.

---

# What I would NOT rebuild

This matters because rebuilding everything would be a mistake.

I estimate we can preserve most of the difficult engineering.

**Keep largely intact:**

```text
OPA policy logic
Redis atomic budget/reservation concepts
Postgres guard state
EIP-712 signing
execution digest/replay protection
Solidity enforcement
x402 policy implementation
receipt hashing
outbox semantics
settlement verification
CDP executor
reconciliation logic
failure tests
```

The code isn't the problem.

**The domain boundaries are.**

I'd estimate roughly:

```text
65-75% reuse
15-20% extraction/refactoring
10-20% genuinely new product code
```

The major new code is:

```text
FinancialIntent
Principal/delegation model
AuthorizationGrant
ExecutionRail interface
normalized execution state machine
adapter framework
unified evidence envelope
product API
console
```

---

# And I would deliberately NOT support five chains now

Even though grants tempt us.

Initial production architecture:

```text
Core
+
x402
+
Base
+
one wallet/provider
```

Then grant-funded adapters:

```text
Arc
Solana
Stellar
```

should come later.

Otherwise we'll spend three months doing ecosystem integrations while having **zero users giving agents actual budgets**.

That would be grant success and startup failure.

---

## My first 6 implementation PRs

I would restructure the current code in exactly this sequence:

1. **`core: introduce FinancialIntent + Principal`**

   No behavior changes.

2. **`core: extract AuthorityGrant and hierarchical budget engine`**

   Move generic logic out of x402.

3. **`execution: introduce ExecutionRail + canonical lifecycle`**

   Adapt CDP/Base paths behind interfaces.

4. **`reconciliation: make UNKNOWN first-class`**

   Move existing CDP lifecycle and watcher logic here.

5. **`adapters: migrate x402 + CDP + Base`**

   Old implementations now consume core abstractions.

6. **`api/sdk: expose authorize(), execute(), verify()`**

   Then build the killer agent demo.

Only after this would I build Arc/Solana/Stellar.

The key architectural insight is therefore:

> **Don't redesign Railguard around agents, x402 or stablecoins. Redesign it around delegated financial authority.**

Agents are today's wedge.

Stablecoins are today's best rail.

x402 is today's machine-payment protocol.

Base/Arc/Solana/Stellar are execution environments.

But **delegated financial authority + safe execution + evidence** is the durable abstraction that can survive all of them.

That is the version I'd be comfortable taking to YC, Circle, Base, Visa or an enterprise fintech architect without changing the core story.

---

## Implementation status

See **[v5execution.md](./v5execution.md)** for code-grounded completion map (kernel, API, SDK, adapters).
