# Final Grant + Hiring Operating Plan

## 1. The single strategic thesis

Do not build a collection of unrelated grant projects.

Build a small **open-source payment reliability studio** around one expertise:

> **Prasanth Kuna builds systems that prove whether programmable payments were authorized correctly, executed safely and reconciled accurately.**

The reusable lifecycle is:

```text
Intent
→ Policy
→ Authorization
→ Reservation
→ Execution
→ Settlement verification
→ Reconciliation
→ Audit evidence
```

Your existing three repositories already cover different parts of this lifecycle:

* `x402-guard`: pre-payment policy and budget authorization
* `railguard-new`: delegated smart-account restrictions and on-chain enforcement
* `railguard-cdp`: enterprise invoice, approval, CDP execution and reconciliation

The grant strategy should produce **three active products**, not ten shallow chain ports.

---

# 2. Non-negotiable engineering foundation

No major security or payment grant should be submitted while the central lifecycle has a known accounting inconsistency.

## Fix the post-broadcast reservation lifecycle

The current CDP path can receive a transaction hash, classify the payment as `unknown`, and still release the guard authorization.

The reconciler may later confirm the transaction, but it does not restore or commit that released authorization.

Replace the lifecycle with:

```text
AUTHORIZED
    ↓
RESERVED
    ↓
EXECUTING
    ↓
SUBMITTED
    ├── CONFIRMED → COMMITTED
    ├── REVERTED → RELEASED
    ├── UNKNOWN → keep frozen
    └── RECONCILIATION_REQUIRED → keep frozen
```

Rules:

* Release only when broadcasting definitely did not occur.
* A transaction hash means the authorization must remain reserved.
* Unknown state must block another attempt.
* Reconciliation—not an HTTP response—decides commit versus release.
* Every transition must be idempotent.

## Persist complete payment correlation

Add durable fields such as:

```text
payment_identifier
guard_fingerprint
guard_authorization_id
guard_receipt_id
guard_status
tx_hash
expected_chain_id
expected_token
expected_sender
expected_recipient
expected_amount
settlement_status
reconciliation_attempts
```

A process restart must not lose the relationship between:

```text
business payment intent
↔ policy authorization
↔ budget reservation
↔ blockchain transaction
↔ settlement evidence
```

## Verify financial facts, not only receipt success

The current provider checks transaction success and confirmations but does not fully prove that the expected token transfer occurred.

Confirmation must verify:

```text
correct network
correct token contract
correct sender
correct recipient
correct amount
correct Transfer event
successful receipt
required confirmation depth
```

A successful transaction with the wrong recipient must become:

```text
RECONCILIATION_REQUIRED
```

not `CONFIRMED`.

## Add adversarial proofs

Required tests:

```text
Crash before broadcast
Crash immediately after broadcast
Crash after storing tx hash
Concurrent duplicate execution
Late successful confirmation
Definitive revert
Successful wrong-recipient transaction
Successful wrong-amount transaction
Repeated reconciler execution
Process restart with no in-memory guard state
```

## Correct public positioning

The `x402-guard` README currently says x402 provides no replay/idempotency surface, which is no longer a safe claim.

Replace it with:

> x402 provides protocol hooks and payment identifiers. This project provides durable multi-agent budget state, atomic reservations, crash recovery and settlement reconciliation around those primitives.

Do not publish packages under `@x402-guard/*`.

Use a personal scope until a permanent brand is legally and commercially checked:

```text
@prasanthkuna/payment-policy
@prasanthkuna/payment-state
@prasanthkuna/x402-adapter
```

## Make every repository reviewer-friendly

Provide one portable interface:

```bash
make setup
make test
make e2e
make failure-lab
```

The complete proof path must run on Linux, macOS and Windows without requiring reviewers to interpret PowerShell-specific instructions.

---

# 3. Product One: Agent Payment Failure Lab

This is the flagship hiring, open-source and ecosystem asset.

Use a descriptive identity initially:

```text
Agent Payment Failure Lab
```

Do not invest further in the Railguard public brand until naming checks are complete.

## Purpose

A reusable test runner that detects failures in autonomous and stablecoin payment integrations.

It should test buyers, wallets, resource servers, facilitators and enterprise payment backends.

## Initial failure profiles

### APF-001 — Payment replay

```text
One valid payment proof
N concurrent resource requests
Expected: one financial authorization and one eligible grant
```

### APF-002 — Shared-budget race

```text
Remaining budget: 100
Concurrent requests: 70 and 70
Expected: only one authorization succeeds
```

### APF-003 — Crash after broadcast

```text
Transaction broadcasts
Application crashes before local commit
Expected:
- payment becomes UNKNOWN
- reservation remains frozen
- retry is blocked
- reconciler commits after chain verification
```

### APF-004 — Successful wrong transfer

```text
Transaction receipt succeeds
Transfer event contains the wrong recipient or amount
Expected:
- payment is not confirmed
- authorization is not committed
- manual/automated reconciliation is required
```

### APF-005 — Stale approval

```text
Approval is granted for Wallet A and Amount X
Payment facts are changed
Expected: old approval is rejected
```

### APF-006 — Off-chain policy bypass

```text
Agent bypasses policy middleware
Attempts direct smart-account execution
Expected: on-chain session restrictions block it
```

## Repository structure

```text
agent-payment-failure-lab/
├── profiles/
│   ├── replay-grant/
│   ├── budget-race/
│   ├── crash-after-broadcast/
│   ├── wrong-transfer/
│   ├── stale-approval/
│   └── middleware-bypass/
├── fixtures/
│   ├── vulnerable-x402/
│   ├── fixed-x402/
│   ├── vulnerable-cdp/
│   ├── fixed-cdp/
│   └── smart-account/
├── adapters/
│   ├── base/
│   ├── cdp/
│   ├── x402/
│   └── circle-arc/
├── reporters/
│   ├── json/
│   ├── junit/
│   └── sarif/
├── evidence/
├── docs/
└── github-action/
```

## Output contract

Every test should produce machine-verifiable evidence:

```json
{
  "profile": "APF-003",
  "result": "FAIL",
  "invariant": "budget must remain reserved after broadcast",
  "observed_state": {
    "payment_status": "unknown",
    "authorization_status": "released",
    "tx_hash_present": true
  },
  "severity": "critical",
  "evidence_hash": "..."
}
```

## How the three existing repos support it

### `x402-guard`

Becomes the reference implementation for:

* atomic replay claim;
* rolling budgets;
* reserve/commit/release;
* official x402 hook integration;
* payment-identifier mapping;
* authorization receipts.

### `railguard-new`

Becomes the hard-enforcement fixture for:

* delegated sessions;
* recipient restrictions;
* token restrictions;
* per-transfer caps;
* cumulative caps;
* expiry;
* batch validation;
* on-chain escape prevention.

### `railguard-cdp`

Becomes the enterprise fixture for:

* invoice extraction;
* duplicate detection;
* vendor-wallet validation;
* approval-snapshot binding;
* separation of duties;
* exactly-once execution claim;
* post-broadcast recovery;
* settlement reconciliation.

## Funding targets

### Base

Base currently offers weekly Builder Rewards and retroactive Builder Grants of approximately 1–5 ETH. Base explicitly values deployed projects, documentation and measurable impact; mainnet deployment is advantageous even where prototypes are accepted. ([Base documentation][1])

The Base module should include:

```text
Base mainnet micro-fixture
USDC Transfer verification
Base confirmation/finality profile
Builder Code attribution
public failure dashboard
```

Base Builder Codes can attribute transactions back to the application, helping demonstrate onchain usage. ([Base documentation][2])

Use Builder Rewards as visibility and possible small income. Use the larger Builder Grant only after external usage exists.

### CDP Founders Fuel

Founders Fuel provides up to $15,000 in CDP/Paymaster credits, $5,000 in AWS credits, product-lead access and launch support. It expects a functioning product, a live product URL, CDP integration and a credible public-launch plan. ([Coinbase][3])

The application should feature:

```text
CDP Server Wallet adapter
post-broadcast recovery profile
hosted payment verifier
public launch URL
clear product roadmap
```

The main value is not only credits. Direct contact with CDP product leads can materially help hiring, ecosystem credibility and future funding.

### Circle

Circle’s main grant page presents an “Apply now” path and prioritizes Arc, Circle products and agentic economic activity. However, a separate official application page currently says the window is closed. Treat the program as **prepared but status-conflicted** until the actual grants portal accepts a submission. ([Circle][4])

The Circle proposal should not be “generic test suite plus Arc.”

It should be:

## Arc Agent Settlement Assurance

```text
Agent Wallet payment intent
→ Arc/Circle execution
→ submitted/unknown/confirmed lifecycle
→ USDC settlement-fact verification
→ reconciliation API
→ signed evidence receipt
```

Circle-specific integrations:

```text
Arc
Circle Agent Wallets
USDC
Gateway
CCTP where relevant
Nanopayments where relevant
```

Circle prioritizes products where Arc is central, Circle products are meaningful architecture components, and the applicant can demonstrate usage, pilots or a credible path to traction. ([Circle][4])

### x402 Foundation

Do not pursue a nonexistent grant program.

Use the Foundation for standards credibility:

* attend open TSC and working-group meetings;
* propose neutral failure fixtures;
* contribute Payment-Identifier tests;
* contribute crash-recovery examples;
* submit specification clarifications;
* avoid branded promotional documentation.

The public community currently supports Slack, GitHub participation, open meetings and working groups around domain discovery, identity, tax and card acceptance. ([x402][5])

Your goal is:

```text
one accepted neutral contribution
one maintainer discussion
one external repository running a test profile
```

---

# 4. Product Two: GNU Taler Merchant Reliability Lab

This is a genuinely separate public-good project, not a renamed x402 adapter.

NLnet’s active NGI TALER call accepts proposals for privacy-preserving digital-payment work. It supports open-source engineering, validation, testing, continuous integration, security work, documentation and standardization. Individuals and organizations may apply, and proposed projects are generally in the €5,000–€50,000 range. ([nlnet.nl][6])

## Proposal thesis

> GNU Taler protects payment privacy, but merchant implementations still need robust order binding, failure recovery, refund correctness and reconciliation tooling.

## Scope

```text
Taler order/payment binding
duplicate fulfilment prevention
merchant crash recovery
exchange/merchant state reconciliation
refund lifecycle verification
idempotent fulfilment
tamper-evident merchant evidence
fault-injection test fixtures
CI integration for merchant deployments
```

## Deliverables

### Merchant lifecycle model

```text
ORDER_CREATED
→ PAYMENT_REQUIRED
→ PAID
→ FULFILLED
→ REFUND_PENDING
→ REFUNDED
→ RECONCILIATION_REQUIRED
```

### Failure profiles

```text
same payment fulfils multiple orders
same order fulfilled multiple times
merchant crash after payment
merchant crash before fulfilment
refund acknowledged but not completed
exchange and merchant state disagree
database restart loses fulfilment state
```

### Reference implementation

* vulnerable merchant fixture;
* corrected merchant fixture;
* test runner;
* CI pipeline;
* documentation;
* reproducible environment;
* upstream-compatible contribution path.

## Recommended application size

Request an amount proportional to a clearly bounded open-source scope, approximately:

```text
€20,000–€30,000
```

Do not request the maximum merely because it exists.

The application should emphasize:

* open-source results;
* direct GNU Taler relevance;
* security and reliability;
* value for merchants;
* upstream engagement;
* measurable milestones;
* transparent disclosure of reused testing infrastructure.

NLnet requires fully open outcomes and evaluates technical merit, strategic relevance and value for money. ([nlnet.nl][7])

---

# 5. Product Three: Stellar Payment Assurance Kit

The Stellar Community Fund currently has a live Build Award round with Open, Integration and RFP tracks, requiring an interest form before a full invited submission. Awards can reach up to $150,000 in XLM. ([Stellar Community Fund][8])

Do not ask for the maximum as a first-time Stellar applicant.

Recommended scope:

```text
$25,000–$45,000
```

## Proposal thesis

> Business payment systems on Stellar need a deterministic relationship between internal payment intent, submitted Stellar transaction, token movement and accounting outcome.

## Scope

```text
payment-intent identity
duplicate payout protection
transaction-envelope binding
submission lifecycle
Soroban/token event verification
USDC settlement verification
crash-after-submit recovery
invoice/payment reconciliation
evidence receipts
GitHub Action
hosted verification service
```

## Stellar-specific implementation

The product must be genuinely native:

```text
Stellar accounts and transactions
Soroban where appropriate
Stellar token/USDC event semantics
ledger sequence and finality
memo/payment-reference handling
Stellar testnet and mainnet fixtures
```

Do not present it as:

> “Railguard ported to Stellar.”

Present it as:

> “Payment reliability and reconciliation infrastructure designed around Stellar’s transaction and asset model.”

## Required external proof

Target:

* feedback from Stellar developers;
* one integration partner;
* one public reference application;
* reproducible testnet evidence;
* a clear path to mainnet.

---

# 6. Secondary grant modules

Do not build all of these simultaneously.

Activate one only after an active product has external validation.

## NEAR Agent Payment Mandates

NEAR’s Protocol Rewards program currently accepts applications for future cohorts and advertises launch-stage rewards up to $10,000 per month based on traction and progress. ([nearprotocolrewards.com][9])

Potential module:

```text
NEAR Intents outcome verification
agent spending mandates
cross-chain intent reconciliation
agent/payee policies
proof of final execution
```

Use only when the implementation genuinely depends on NEAR Intents or NEAR’s agent infrastructure.

## Aptos Move Payment Invariant Tester

Aptos Ecosystem Grants are live, milestone-based and generally range from $5,000 to $50,000, but require a live product and evidence of adoption or community momentum. ([Aptos Network][10])

Potential module:

```text
Move coin-flow invariants
resource-account permissions
transaction simulation
wrong-event verification
payment concurrency tests
```

Do not apply until the generic failure lab has actual external usage.

## Tezos Payment Reliability Toolkit

The Tezos Foundation accepts grant proposals throughout each quarter and reviews them after quarter-end. ([Tezos Foundation][11])

Pursue only after identifying a specific Tezos ecosystem need through developer conversations.

## Solana and Superteam

Use Superteam as a repeatable bounty and network channel, not as the primary company strategy. Current India listings include development grants, Solana Foundation India grants and small technical bounties. ([Superteam][12])

Select work that compounds your positioning:

```text
Solana payment subscriptions
agent spending allowances
transaction reconciliation
RPC correctness
wallet/payment safety
```

Reject unrelated content and meme bounties unless the payout-to-effort ratio is exceptional.

---

# 7. Infrastructure-credit strategy

Track credits separately from cash.

## Cloudflare for Startups

Cloudflare currently offers a $10,000 bootstrapped tier and larger tiers for qualifying startups. It requires an incorporated for-profit technology company, a public website, public social presence and a domain-matching business email. ([Cloudflare][13])

Use Cloudflare for:

```text
Workers — test execution APIs
Durable Objects — per-test coordination
Queues — asynchronous test runs
Workflows — multi-step failure simulations
R2 — evidence bundle storage
D1/Postgres — test metadata
WAF/rate limits — public runner protection
```

Apply only through a legitimate qualifying entity, such as Stackular, if the eligibility and ownership structure are valid.

## CDP credits

Use only for:

```text
CDP wallet execution
Paymaster usage
x402 fixtures
testnet/mainnet integration
hosted reference demonstrations
```

## Credit accounting

Maintain two values:

```text
Face value
Actual expected cost avoided
```

Example:

```text
Cloudflare face value: $10,000
Expected real usage: $1,800
Economic value recorded: $1,800
```

Never present $35,000 of credits as $35,000 of revenue.

---

# 8. Grant operating system

Create a private repository:

```text
grant-ops/
├── opportunities/
├── applications/
├── work-packages/
├── budgets/
├── disclosures/
├── evidence/
├── metrics/
├── legal-tax/
├── reusable-copy/
└── decisions/
```

## Opportunity record

```yaml
program:
status:
cash_or_credits:
maximum_award:
recommended_request:
eligibility:
entity_required:
open_source_required:
ecosystem:
required_integration:
traction_required:
submission_mechanism:
decision_mechanism:
strategic_value:
estimated_probability:
application_effort:
delivery_effort:
hiring_value:
```

## Work-package funding ledger

Every engineering unit gets a funding owner:

| Work package             | Status               | Funding source |
| ------------------------ | -------------------- | -------------- |
| Failure schema v1        | Self-funded          | Prasanth       |
| Generic runner           | Self-funded          | Prasanth       |
| Taler merchant fixture   | Proposed             | NLnet          |
| Arc adapter              | Proposed             | Circle         |
| Base mainnet profile     | Proposed             | Base           |
| Stellar adapter          | Proposed             | SCF            |
| Cloudflare hosted runner | Credits              | Cloudflare     |
| CDP adapter              | Existing/self-funded | Prasanth       |

Rule:

> One funder may fund one defined work package. Shared code may be reused, but completed or previously funded work must be disclosed.

---

# 9. Standard grant application package

Prepare one reusable evidence bundle.

## Founder section

* 11+ years backend/platform engineering;
* payments, fintech and crypto integrations;
* distributed-system experience;
* Go, TypeScript, .NET, Python;
* existing payment-security reference implementations;
* ability to ship across API, database and on-chain layers.

## Problem evidence

Show a concrete failure:

```text
Invariant:
A broadcast payment must retain its reservation until settlement is known.

Failure:
Transaction succeeds while application releases reservation.

Impact:
Available budget is overstated and another payment may be authorized.
```

## Product evidence

Provide:

```text
working repository
green CI
portable demo
hosted URL
architecture diagram
three-minute demonstration
machine-readable test result
known limitations
security disclosure process
```

## Milestone structure

Each milestone must have:

```text
deliverable
acceptance test
public artifact
measurable result
budget
dependency
risk
```

Example:

```text
Milestone:
Implement Arc settlement-fact verifier

Acceptance:
Given expected USDC contract, sender, recipient and amount,
the verifier returns CONFIRMED only when all facts match.

Evidence:
Open-source code, integration test, hosted result and signed receipt.
```

## Budget structure

Request based on work, not the maximum.

Use:

```text
engineering
security/testing
documentation
infrastructure
community/integration support
legal/accounting where eligible
```

Avoid inflated founder salary, vague “marketing” budgets and unexplained contingencies.

## Sustainability section

Explain how the project survives:

```text
open-source core
ecosystem adapters
hosted private testing
sponsored maintenance
enterprise support
future grants for distinct modules
```

---

# 10. Application qualification rules

Score every opportunity using:

```text
Expected strategic value
=
award value
× realistic probability
× ecosystem/hiring multiplier
÷ application and delivery effort
```

## Apply when

```text
The integration is ecosystem-native
The application is currently open
You satisfy entity requirements
A working proof exists
The work package is not already funded
The grant creates reusable evidence
The ecosystem relationship helps hiring or adoption
```

## Reject when

```text
The only fit is changing the chain name
The application requires fake traction
The integration is ornamental
The token reward has weak liquidity or severe lockups
The project distracts from the payment-reliability thesis
The funder wants ownership incompatible with the open core
The effort cannot produce future users or credibility
```

---

# 11. Public identity and hiring conversion

Your public positioning should be:

> **Senior Backend and Platform Engineer building lifecycle-correct payment infrastructure for AI agents and stablecoin systems.**

GitHub memory hook:

> **The engineer who built executable tests for replay, budget races, post-broadcast failures and settlement mismatches.**

## Profile structure

Pin:

1. Agent Payment Failure Lab
2. `railguard-cdp`
3. `railguard-new`
4. `x402-guard`
5. `dispute-defense-engine`
6. one relevant Solana/Go system

## Portfolio front door

Include:

```text
one-line thesis
architecture
failure demonstrations
three project roles
grant-supported modules
external integrations
standards contributions
resume
contact
```

## Every grant should create a hiring asset

| Grant output               | Hiring evidence                               |
| -------------------------- | --------------------------------------------- |
| NLnet proposal/win         | Open-source payments and security credibility |
| Stellar SCF integration    | Cross-chain payment architecture              |
| Circle integration         | Stablecoin and agent-economy experience       |
| Base grant/reward          | Mainnet shipping and ecosystem impact         |
| x402 contribution          | Standards participation                       |
| Cloudflare credits         | Production deployment architecture            |
| External test integrations | Evidence others use your work                 |

Do not pause job applications while pursuing grants.

The grant projects should improve your hiring story, not replace the hiring process.

---

# 12. Final priority order

## Priority 1 — make the existing payment lifecycle correct

```text
Fix reservation-after-broadcast behavior
Persist guard/payment correlation
Verify transfer facts
Add restart and reconciliation tests
Correct README claims
Provide cross-platform execution
```

## Priority 2 — submit the deadline-sensitive public-good applications

```text
GNU Taler Merchant Reliability Lab
Stellar Payment Assurance Kit
```

Both have live, verified funding routes; NLnet supports individuals and open-source payment-security work, while Stellar’s current Build round provides a direct ecosystem-grant pathway. ([nlnet.nl][7])

## Priority 3 — launch the flagship failure lab

```text
Four to six deeply proven profiles
Vulnerable and corrected fixtures
GitHub Action
SARIF/JUnit/JSON output
Hosted runner
Public postmortem
Short exploit demonstration
```

## Priority 4 — obtain infrastructure support

```text
Cloudflare startup credits, subject to entity eligibility
CDP Founders Fuel after public launch
Circle-related credits where genuinely useful
```

## Priority 5 — establish ecosystem authority

```text
x402 meetings
neutral specification/test contributions
Base Builder Rewards
external maintainers running profiles
academic-author feedback
```

## Priority 6 — pursue cash after proof

```text
Base Builder Grant
Circle Developer Grant when portal submission is truly open
additional Stellar milestones
future CDP cohorts
```

## Priority 7 — activate secondary chains selectively

```text
NEAR only for NEAR Intents/agent mandates
Aptos only after traction
Tezos only after ecosystem validation
Solana through high-fit Superteam opportunities
```

# Final operating model

You are not becoming a person who repeatedly asks ecosystems to fund copied apps.

You are becoming:

> **An open-source payment-reliability specialist who maintains a common failure-testing core and delivers genuinely native reliability modules for different payment ecosystems.**

The winning portfolio is:

```text
Agent Payment Failure Lab
    → Base, CDP, Circle, x402, hiring

GNU Taler Merchant Reliability Lab
    → NLnet and privacy-preserving payments

Stellar Payment Assurance Kit
    → Stellar SCF and business-payment infrastructure
```

Success should be measured separately:

```text
Cash grants
Small rewards and bounties
Infrastructure credits
External integrations
Upstream contributions
Maintainer acknowledgements
Qualified hiring loops
```

The strongest realistic outcome is not the maximum number of grant logos.

It is:

> **Multiple legitimate funding wins, one recognized open-source testing asset, external users, standards credibility and a sharply differentiated senior payments-engineering profile.**

[1]: https://docs.base.org/get-started/get-funded?utm_source=chatgpt.com "Get Funded - Base Documentation"
[2]: https://docs.base.org/apps/builder-codes/builder-codes?utm_source=chatgpt.com "Base Builder Codes - Base Documentation"
[3]: https://www.coinbase.com/developer-platform/discover/launches/founders-fuel?utm_source=chatgpt.com "Introducing: Founders Fuel (Benefits for Builders) | Coinbase"
[4]: https://www.circle.com/grant?utm_source=chatgpt.com "Circle Developer Grants | Circle"
[5]: https://x402.org/get-involved/?utm_source=chatgpt.com "Get Involved | x402"
[6]: https://nlnet.nl/taler/eligibility/?utm_source=chatgpt.com "NLnet; NGI TALER Eligibility information"
[7]: https://nlnet.nl/taler/guideforapplicants/index.html?utm_source=chatgpt.com "NLnet; NGI TALER Guide for Applicants"
[8]: https://communityfund.stellar.org/awards?utm_source=chatgpt.com "Build Awards $150K+ | Stellar Ecosystem Grants & Funding"
[9]: https://www.nearprotocolrewards.com/?utm_source=chatgpt.com "NEAR Protocol Rewards | Merit-Based Developer Incentives"
[10]: https://aptosnetwork.com/grants/ecosystem?utm_source=chatgpt.com "Ecosystem Grants | Aptos Foundation | Aptos Network"
[11]: https://tezos.foundation/ecosystem-grants-program/?utm_source=chatgpt.com "Ecosystem Grants Program - Tezos Foundation"
[12]: https://superteam.fun/earn/regions/india/?utm_source=chatgpt.com "Welcome to Superteam Earn India | Discover Bounties and Grants"
[13]: https://www.cloudflare.com/startups/?utm_source=chatgpt.com "Startup Program | Cloudflare"
