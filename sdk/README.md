# `@kuna5678/railguard-sdk`

TypeScript SDK for Railguard: EIP-712 payment intents, session helpers, and x402 guard adapters.

```bash
npm i @kuna5678/railguard-sdk
```

```ts
import { buildPaymentIntent, createX402Guard } from "@kuna5678/railguard-sdk"

const guard = createX402Guard("agent_demo")
```

Depends on `@kuna5678/x402` and `@kuna5678/x402-core`.

**Local dev (sibling checkout):**

```powershell
cd C:\Users\PrashanthKuna\x402-guard
npm run build
$core = npm pack -w @kuna5678/x402-core --silent
$policy = npm pack -w @kuna5678/x402-policy --silent
$receipts = npm pack -w @kuna5678/x402-receipts --silent
$mw = npm pack -w @kuna5678/x402 --silent
cd ..\railguard-protocol\sdk
npm install "..\x402-guard\$core" "..\x402-guard\$policy" "..\x402-guard\$receipts" "..\x402-guard\$mw"
npm test && npm run build
```

**Product vs protocol:** PreBroadcast (hosted AP product) lives in the CDP app repo. This SDK is the open protocol toolkit.

License: Apache-2.0
