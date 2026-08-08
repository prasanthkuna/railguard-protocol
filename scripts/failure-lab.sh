#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> Agent Payment Failure Lab"
make -C "$ROOT/../agent-payment-failure-lab" ci

echo "==> x402-guard"
make -C "$ROOT/../x402-guard" failure-lab

echo "==> railguard-cdp"
make -C "$ROOT/../coinbase" failure-lab

echo "==> GNU Taler lab"
make -C "$ROOT/../gnu-taler-merchant-reliability-lab" test

echo "==> Stellar kit"
make -C "$ROOT/../stellar-payment-assurance-kit" test

if command -v forge >/dev/null 2>&1; then
  echo "==> railguard-protocol contracts"
  make -C "$ROOT" contracts-deps
  make -C "$ROOT" contracts-test
elif [ -x "$HOME/.foundry/bin/forge" ]; then
  echo "==> railguard-protocol contracts"
  make -C "$ROOT" contracts-deps FORGE="$HOME/.foundry/bin/forge"
  make -C "$ROOT" contracts-test FORGE="$HOME/.foundry/bin/forge"
fi

if [ "${TESTNET_INTEGRATION:-}" = "1" ]; then
  echo "==> Live testnet evidence"
  powershell -NoProfile -File "$ROOT/scripts/testnet-evidence.ps1" 2>/dev/null || \
    pwsh -NoProfile -File "$ROOT/scripts/testnet-evidence.ps1"
fi

echo "Full product failure-lab checks passed."
