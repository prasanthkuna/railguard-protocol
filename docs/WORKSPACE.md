# Workspace map

Canonical index for all folders in the Cursor multi-root workspace.

| Folder | Git | Role | Docs |
|--------|-----|------|------|
| **coinbase/** | yes | Railguard v5 monorepo — API, web, kernel, CLI, MCP | `coinbase/README.md`, `railguard-new/docs/v5execution.md` |
| **railguard-new/** | yes | Protocol, Authority Engine Go, Solidity, v5 plans | `docs/v5plan.md`, `docs/PORTFOLIO.md` |
| **x402-guard/** | yes | x402 adapter (OSS) | `README.md` |
| **agent-payment-failure-lab/** | yes | APF conformance | `README.md` |
| **grant-ops/** | yes | Grant applications (private) | `decisions/product-complete.md` |
| **gnu-taler-merchant-reliability-lab/** | yes | Taler harness | `README.md` |
| **stellar-payment-assurance-kit/** | yes | Stellar assurance | `README.md` |
| **resume/** | no | Job search (local — `.gitignore`) | — |

## Naming (v5)

| Legacy | Current |
|--------|---------|
| railguard-cdp | **coinbase/** Railguard monorepo |
| PreBroadcast | Operator UI product name (Vercel) |
| SignGate (product) | **Authority Engine** |
| Payment intent (external) | **Financial intent** |
| Receipt | **Evidence envelope** |

## Live URLs (2026-08)

| Surface | URL |
|---------|-----|
| API (Encore staging) | https://staging-railguard-s4ii.encr.app |
| Web console | https://prebroadcast.vercel.app |

## Do not commit

Across repos, `.gitignore` now blocks: `*.pdf`, `Prasanth_*`, `*.tgz`, `out/pw-*`, CapCut/`~/` artifacts, local seed JSON, personal resume copies in repo roots.

## Verify all products

```powershell
powershell -NoProfile -File railguard-new/scripts/failure-lab.ps1
cd coinbase; bun run test:v5; encore check
```
