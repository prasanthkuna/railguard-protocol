import {
  X402Guard,
  defaultDevPolicy,
  withSpendingPolicy,
  PolicyViolationError,
  ReplayDetectedError,
} from '@railguard/x402';
import { parseResourceUrl } from '@railguard/x402-core';
export type { AgentPolicyConfig, X402PaymentContext } from '@railguard/x402-core';
export type { PaymentReceipt } from '@railguard/x402';

export {
  X402Guard,
  withSpendingPolicy,
  defaultDevPolicy,
  PolicyViolationError,
  ReplayDetectedError,
  parseResourceUrl,
};

export function createX402Guard(agentId: string) {
  return new X402Guard({ policy: defaultDevPolicy(agentId) });
}

/** @deprecated Use createX402Guard */
export function x402AdapterStub() {
  return {
    name: 'railguard-x402-guard',
    buildPaymentIntent: (input: unknown) => ({ input, protocol: 'x402-guard' }),
  };
}
