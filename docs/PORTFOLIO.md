# Railguard — portfolio (start here)

> **Recruiter link:** `https://github.com/prasanthkuna/railguard-protocol/blob/master/docs/PORTFOLIO.md`

[![status](https://img.shields.io/badge/maturity-v0.1.0--alpha-lightgrey)](./RELEASE_v0.1-reference.md)
[![ci](https://github.com/prasanthkuna/railguard-protocol/actions/workflows/ci.yml/badge.svg)](https://github.com/prasanthkuna/railguard-protocol/actions/workflows/ci.yml)

**One line:** Open-source **financial execution firewall** for autonomous software.  
**Tagline:** Agent money. Guarded.

**Maturity:** v0.1.0-alpha · testnet reference · documented production gaps · not for mainnet funds.

## Lifecycle (one product)

```text
Financial Intent → Authorize → Reserve → Execute → Observe → Reconcile → Evidence
```

## Components (not four companies)

| Component | Repository | Role |
|-----------|------------|------|
| **Railguard Core** | [railguard-protocol](https://github.com/prasanthkuna/railguard-protocol) (this repo) | Architecture, SignGate, contracts, security model |
| **Railguard Gateway** | [railguard-cdp](https://github.com/prasanthkuna/railguard-cdp) → `railguard-gateway` | Reference runtime API + operator console |
| **Failure Lab + Atlas** | [agent-payment-failure-lab](https://github.com/prasanthkuna/agent-payment-failure-lab) | Adversarial tests; **Atlas taxonomy owner** |
| **x402 adapter** | [x402-guard](https://github.com/prasanthkuna/x402-guard) | Integration only |

Gateway component map: [COMPONENTS.md](https://github.com/prasanthkuna/railguard-cdp/blob/main/docs/COMPONENTS.md).

## What runs today

```text
Agent / CLI / MCP
      ↓
Railguard SDK (check · authorize · execute · verify)
      ↓
Gateway (Encore / TypeScript) — staging-railguard-s4ii.encr.app
      ↓
CDP + settlement verify (Base Sepolia testnet, …)
      ↓
Evidence / receipts

Optional ENFORCE: SignGate + on-chain hook (this repo)
```

## Reviewer path (≈15 min)

| Step | Action |
|------|--------|
| 1 | `cd agent-payment-failure-lab && npm run lab` |
| 2 | `cd railguard-gateway && bun run railguard scan --base-url https://staging-railguard-s4ii.encr.app` |
| 3 | `cd railguard-protocol/contracts && forge test` |
| 4 | Evidence: [evidence/](../evidence/) · Gateway [arbitrum-sepolia](https://github.com/prasanthkuna/railguard-cdp/tree/main/evidence/arbitrum-sepolia) |

## Demo loop (incubator / grants)

```bash
railguard attack    # vulnerabilities
railguard protect
railguard attack    # blocked
```

## Evidence

[evidence/](../evidence/) — Base Sepolia, Stellar testnet, APF profiles.  
Failure Atlas: [agent-payment-failure-lab/atlas](https://github.com/prasanthkuna/agent-payment-failure-lab/tree/main/atlas).

## Internal history

v5 planning docs remain in-repo for maintainers — not required for product understanding. See [v5plan.md](./v5plan.md) (internal).

## License

MIT (protocol). Gateway and Failure Lab: Apache-2.0. Relicense alignment to Apache-2.0 for core stack planned where copyright permits.
