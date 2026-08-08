# Final Railguard execution order

## 1. Freeze the architecture

Adopt these three enforcement modes and stop redesigning them:

```text
CDP_DIRECT
CDP_VAULT_CALL
RAILGUARD_7579_ACCOUNT
```

Use:

* `CDP_DIRECT` for provider-controlled execution with pre-sign policy and post-settlement verification.
* `CDP_VAULT_CALL` for CDP smart-account calls through a Railguard-controlled vault.
* `RAILGUARD_7579_ACCOUNT` for full account-level validator and hook enforcement.

Do not claim they provide equal protection.

---

## 2. Finalize repository responsibilities

### `railguard-new` → `railguard-protocol`

GitHub repo **`prasanthkuna/railguard`** stays the PreBroadcast product (CDP). Protocol repo renames to **`railguard-protocol`**; npm packages publish under **`@kuna5678/*`**.

Make this the canonical platform containing:

* Domain contracts
* PostgreSQL financial authority
* Policy API
* SignGate
* Execution-driver protocol
* Watcher
* Reconciler
* Evidence receipts
* Solidity contracts
* Operator APIs

Rename the old archived `railguard` repository to `railguard-legacy` first.

### Keep `x402-guard` separate

It becomes the standalone developer-distribution layer:

```text
@kuna5678/x402
@kuna5678/x402-policy
@kuna5678/x402-receipts
```

Support:

* Embedded mode
* Remote Railguard control-plane mode

### Keep `railguard-cdp` temporarily

Use it as the reference invoice application and live CDP integration. Do not archive it until the extracted CDP driver works through stable Railguard interfaces.

---

## 3. Define the canonical domain objects

Create versioned schemas for:

```text
Purchase
Quote
PaymentIntent
PolicyDecision
Approval
Authorization
BudgetWindow
ExecutionAttempt
ProviderRequest
SettlementObservation
ReconciliationCase
Fulfilment
EvidenceReceipt
```

Do not use one enormous payment object.

---

## 4. Establish the identifier chain

Every transaction must preserve:

```text
business_idempotency_key
→ purchase_id
→ quote_id
→ intent_id
→ policy_decision_id
→ approval_id
→ authorization_id
→ execution_id
→ provider_idempotency_key
→ provider_operation_id / userOpHash
→ transaction_hash
→ fulfilment_id
→ receipt_id
```

Never correlate transactions using only timestamps, wallet addresses or “oldest pending reservation.”

---

## 5. Implement PostgreSQL as financial truth

Move authoritative monetary state into PostgreSQL:

* Purchases
* Budgets
* Reservations
* Execution attempts
* Provider requests
* Settlement observations
* Reconciliation cases
* Audit records

Use Redis only for:

* Non-financial throttling
* Cached policies
* Temporary UI counters
* Optional performance caching

Redis loss must not change spend limits or payment outcomes.

---

## 6. Add financial database constraints

At minimum:

```sql
UNIQUE (principal_id, business_idempotency_key)
UNIQUE (intent_id)
UNIQUE (authorization_id)
UNIQUE (execution_id)
UNIQUE (provider, provider_idempotency_key)
UNIQUE (provider, provider_operation_id)
UNIQUE (chain_id, transaction_hash)
UNIQUE (merchant_id, fulfilment_id)
```

Use atomic conditional updates or deterministic row locking for hierarchical budgets.

---

## 7. Build hierarchical budget reservations

Support:

```text
Organisation budget
→ Team budget
→ Agent budget
→ Merchant budget
→ Resource-specific budget
```

Reserve all applicable budgets in one PostgreSQL transaction.

The reservation operation must:

* Fail closed
* Prevent concurrent overspend
* Produce an `authorization_id`
* Bind the amount, asset, recipient and policy version
* Remain recoverable after service restart

---

## 8. Adopt the orthogonal lifecycle model

### Intent status

```text
CREATED
APPROVAL_REQUIRED
APPROVED
REJECTED
EXPIRED
```

### Authorization status

```text
UNRESERVED
RESERVED
FROZEN
COMMITTED
RELEASED
```

### Execution status

```text
NOT_STARTED
PREPARED
SUBMITTING
SUBMITTED
UNKNOWN
REVERTED
```

### Settlement status

```text
UNOBSERVED
INCLUDED
SAFE
FINALIZED
MISMATCH
```

### Reconciliation status

```text
CLEAN
REQUIRED
IN_PROGRESS
RESOLVED
```

Do not replace financially significant ambiguity with a generic `FAILED` state.

---

## 9. Create structured diagnoses

Diagnoses must be machine-readable:

```json
{
  "code": "CDP_REQUEST_TIMED_OUT_AFTER_DISPATCH",
  "severity": "HIGH",
  "retry_safety": "REQUIRES_OBSERVATION",
  "reservation_action": "FREEZE",
  "recovery_action": "RETRY_PROVIDER_SAME_KEY"
}
```

Implement an action engine that derives:

```text
allowed_actions
forbidden_actions
```

Clients must not independently decide whether resubmission is safe.

---

## 10. Define the execution-driver protocol

Do not use one `execute()` method returning a supposedly final receipt.

Use:

```text
prepare
submit
observe
findByCorrelation
```

The submission result must distinguish:

```text
BROADCAST_CONFIRMED
BROADCAST_UNKNOWN
REJECTED_BEFORE_BROADCAST
```

Policy, reservations, settlement verification and reconciliation remain outside the driver.

---

## 11. Build the `CDP_DIRECT` driver first

Before calling CDP:

1. Create the execution attempt.
2. Generate a UUID v4 provider idempotency key.
3. Store the exact canonical CDP request.
4. Store its hash.
5. Set execution to `SUBMITTING`.
6. Freeze the authorization.
7. Commit the PostgreSQL transaction.

Only then call CDP.

Persist:

* Provider response
* `userOpHash` or operation reference
* Transaction hash when available
* Response hash
* Provider status

---

## 12. Implement CDP ambiguous-outcome recovery

Deliberately test:

```text
CDP accepts the request
→ Railguard loses the response
→ Railguard process restarts
```

Recovery must:

1. Load the stored CDP idempotency key.
2. Recreate the exact stored request.
3. Verify its request hash.
4. Retry CDP with the same idempotency key.
5. Recover the original provider response.
6. Persist the original operation hash.
7. Observe settlement.
8. Prevent another execution from being created.
9. Commit or quarantine the reservation based on chain evidence.

Never retry with a new CDP idempotency key while the original execution is unresolved.

---

## 13. Implement durable purchase identity

Create `purchase_id` before merchant network I/O.

Connected mode:

```text
Application
→ Railguard createPurchase
→ PostgreSQL commit
→ merchant request
```

Embedded mode must use a pluggable durable store:

```text
SQLite
PostgreSQL
File store for demos
In-memory store marked non-durable
```

Document that idempotency guarantees weaken when the agent, merchant or provider fails to preserve the identifier.

---

## 14. Add quote versioning

A merchant retry may return different x402 payment requirements.

Store:

```text
purchase
├── quote_1
├── quote_2
└── quote_3
```

Compare material facts:

* Merchant
* Resource
* HTTP method
* Request-body hash
* Network
* Token
* Recipient
* Amount
* Expiry

Material changes require policy re-evaluation and possibly human reapproval.

---

## 15. Publish the Railguard control-plane API contract

Create a stable versioned protocol:

```text
GET  /v1/capabilities
POST /v1/purchases
POST /v1/policy/evaluate
POST /v1/authorizations/reserve
POST /v1/authorizations/{id}/commit
POST /v1/authorizations/{id}/release
POST /v1/executions
GET  /v1/executions/{id}
GET  /v1/purchases/{id}
```

Every client sends:

```text
Railguard-Protocol-Version
Railguard-SDK-Version
```

Define compatibility rules before publishing packages.

---

## 16. Convert `x402-guard` into `@kuna5678/x402`

Wrap the official x402 packages rather than reproducing the protocol.

The package must:

* Create or receive a durable purchase identity
* Normalise payment requirements
* Propagate the x402 Payment Identifier where supported
* Call embedded or remote policy evaluation
* Reserve budgets before payment
* Prevent unresolved execution retries
* Record settlement results
* Expose evidence to the caller

It must remain usable without deploying the complete Railguard platform.

---

## 17. Add resource-fulfilment idempotency

Financial idempotency and fulfilment idempotency are separate.

Implement a merchant-side contract such as:

```text
purchase_id
payment_identifier
settlement_receipt
fulfilment_id
```

Enforce:

```sql
UNIQUE (merchant_id, fulfilment_id)
```

Prove:

* One payment cannot fulfil twice.
* One fulfilment cannot trigger two payments.
* A retried HTTP request returns the existing fulfilment result where appropriate.

---

## 18. Add complete observability

Propagate one trace context across:

```text
@kuna5678/x402
→ Railguard API
→ policy evaluation
→ authorization reservation
→ CDP driver
→ provider response
→ watcher
→ settlement verifier
→ reconciler
→ merchant fulfilment
```

Metrics should include:

* Policy latency
* Reservation latency
* Provider submission latency
* Unknown-execution count
* Time to settlement
* Time to reconciliation
* Frozen authorization value
* Funds potentially at risk
* Duplicate executions prevented
* Fulfilment duplicates prevented

---

## 19. Implement settlement verification

Verify actual onchain facts against the authorised intent:

* Chain
* Token contract
* Sender
* Recipient
* Amount
* Transaction success
* Expected contract or vault
* Execution identifier where available

A successful receipt with incorrect transfer facts becomes:

```text
settlement_status = MISMATCH
reconciliation_status = REQUIRED
authorization_status = FROZEN
```

---

## 20. Add minimal Base finality handling

Store:

```text
transaction_hash
block_number
block_hash
receipt_status
observed_transfer_facts
confidence
```

Use:

```text
PROVISIONAL
SAFE
FINALIZED
```

Detect:

* Receipt disappearance
* Inclusion block-hash change
* Transaction replacement
* Reverted transaction

Do not build a full common-ancestor rewind engine yet.

---

## 21. Build the transactional outbox

Any transaction that changes financial state should also insert an outbox record.

Use the outbox for:

* Watcher jobs
* Webhooks
* Audit events
* Notifications
* Cache updates
* Evidence generation

Never rely on “database commit followed by best-effort message publishing.”

---

## 22. Build the canonical failure demo

Demonstrate:

```text
Durable purchase created
→ x402 requirements received
→ policy approved
→ PostgreSQL budget reserved
→ CDP request persisted with idempotency key
→ CDP accepts request
→ provider response deliberately lost
→ Railguard process terminated
→ same purchase retried
→ second execution blocked
→ same CDP request replayed with same key
→ original operation recovered
→ Base settlement verified
→ authorization committed
→ merchant fulfils once
→ signed evidence receipt issued
```

Required result:

```text
HTTP attempts:          2
Purchases:              1
Authorizations:         1
Execution attempts:     1
Onchain settlements:    1
Fulfilments:            1
Final reconciliation:   CLEAN
```

---

## 23. Build the degraded-guarantee demo

Demonstrate:

```text
Agent durable store deleted
→ new purchase identity generated
→ merchant does not preserve previous key
```

Railguard should:

* Detect missing prior correlation where possible
* Mark the guarantee as degraded
* Require approval or manual review before another material payment
* Avoid falsely claiming global exactly-once behaviour

---

## 24. Add the `CDP_VAULT_CALL` driver

Create a small Railguard vault/executor contract that enforces:

* Authorised execution ID
* Intent hash
* Token
* Recipient
* Amount
* Expiry
* Replay protection

Emit:

```text
RailguardExecution(
  executionId,
  intentHash,
  token,
  recipient,
  amount
)
```

Call it through a CDP smart account with provider paymaster support.

Compare it with `CDP_DIRECT` on:

* Gas
* Latency
* Recovery
* Bypass resistance
* Operational complexity
* Evidence strength

---

## 25. Preserve the ERC-7579 path as the highest-assurance mode

Keep your existing:

* Railguard account adapter
* Session validator
* Execution hook
* SignGate cosigning
* Execution-digest watcher

Do not make this path block the first unified release.

Present it as:

```text
RAILGUARD_7579_ACCOUNT
```

and clearly distinguish it from CDP vault calls.

---

## 26. Extract the CDP integration

After the execution protocol stabilises, extract from `railguard-cdp`:

```text
CDPDirectDriver
CDPVaultCallDriver
CDPExecutionObserver
CDPProviderRecovery
```

Make `railguard-cdp` consume those stable interfaces as an external reference application.

Only then consider archiving or renaming the repository.

---

## 27. Publish external packages and a clean sample

Publish:

```text
@kuna5678/railguard-sdk
@kuna5678/x402
@railguard/evidence
```

Create a separate sample repository that uses only published versions.

It must not rely on:

* Sibling repository paths
* Local package overrides
* Copied source
* Private environment assumptions

---

## 28. Add security infrastructure after correctness works

Then add:

* KMS-backed SignGate signer
* Key rotation
* Separation of approval and signing authority
* Production secret management
* Operator role controls
* Emergency pause
* Tamper-evident evidence signing

Do not let KMS work delay ambiguity recovery and idempotency.

---

## 29. Add customer-driven capabilities only

After the complete proof works, add features only when demanded:

* Additional EVM chain
* AP2 ingress adapter
* More wallet providers
* Policy editor
* Advanced finality handling
* Additional compliance connectors

Do not add:

* A token
* A custom facilitator
* A custom paymaster
* Solana without a concrete requirement
* Trading features
* ZK proofs
* Another Railguard repository

---

# Final completion test

Railguard is ready to present seriously when it can truthfully prove:

> For a durable purchase identity, Railguard records what was authorised, prevents an unresolved provider execution from being repeated, verifies the resulting onchain settlement, commits financial authority only after matching evidence, and ensures the merchant fulfils the purchase no more than once.
