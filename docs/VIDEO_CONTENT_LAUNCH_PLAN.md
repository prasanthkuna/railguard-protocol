# Railguard proof-led launch and education plan

## Executive decision

Do not launch Railguard v0.1 as a production-ready payments product.

Launch it as a **proof-led open-source reference release for safer AI-agent stablecoin payments**:

> Railguard demonstrates how to stop agent-payment failures across intent, policy, signing, execution, receipt, and reconciliation.

This positioning matches the repository:

- v0.1 is explicitly a reference implementation.
- Live testnet and failure-mode evidence already exists.
- Known production gaps are documented.
- The largest current gap is external adoption, not asset generation.

The launch objective is therefore:

> Make one technical failure understandable in 30 seconds, verifiable in 3 minutes, and adoptable in 10 minutes.

## What exists already

Railguard has enough raw material for a professional campaign:

- 62 generated campaign PNGs in `assets/x-campaign/`
- Eight educational carousel sets
- Brand mark, wordmark, banner, hero, and Open Graph assets
- Terminal proof images for intent hashing and CDP payment state
- Base Sepolia, Stellar testnet, APF-003/APF-004, and CDP flow evidence
- A five-minute demo path
- Architecture, threat model, test matrix, failure-mode, and honest-gap documents
- A reusable HTML/Chrome asset generator with design tokens

The current assets are technically credible and visually consistent. Their limitation is that they are mostly static and document-first. They need to become a narrative system built around:

```text
Failure → consequence → unsafe behavior → Railguard invariant → live proof → next action
```

## The launch audience

### Primary: protocol and payments engineers

They need:

- a real failure they recognize;
- a reproducible command;
- a before/after result;
- supported environments and limitations;
- an integration path.

Desired action:

> Run the failure profile or add the Action to one repository.

### Secondary: ecosystem maintainers and grant reviewers

They need:

- evidence of a meaningful ecosystem problem;
- live testnet proof;
- explicit milestones and gaps;
- evidence that the work can become shared infrastructure.

Desired action:

> Review, recommend, fund, or accept an integration contribution.

### Tertiary: recruiters and technical leaders

They need:

- system-design depth;
- a clear explanation of one hard correctness bug;
- evidence of personal engineering judgment;
- an honest account of trade-offs and incomplete work.

Desired action:

> Open the portfolio, watch the three-minute proof, and start a technical conversation.

## Message architecture

### Public one-liner

> Policy-enforced safety for AI-agent stablecoin payments, with adversarial failure tests and testnet evidence.

### Plain-language hook

> A payment can be broadcast successfully while your app thinks it failed. If the budget is released too early, the agent can pay again.

### Technical thesis

> Payment safety is an end-to-end state-machine problem, not a validator problem.

### Proof line

> Railguard binds intent, reserves budget atomically, preserves unknown outcomes, and reconciles execution by digest.

### Mandatory status line

> v0.1 reference implementation; testnet evidence; not production-ready for mainnet funds.

Every launch asset should contain the status line directly or link to a page where it is immediately visible.

## Content portfolio

Produce a small set of durable master assets, then derive short variants from them.

### P0 — flagship proof film

**Title:** `The payment succeeded. The app said it failed.`

- Length: 2:30–3:15
- Format: 16:9 master, with 9:16 and 1:1 excerpts
- Purpose: launch anchor and portfolio front door
- Audience: all three audience groups
- CTA: run the demo, inspect the evidence, or review the portfolio

Beat sheet:

1. **0:00–0:12 — Hook**
   - Show a payment broadcast.
   - Simulate the process losing certainty after broadcast.
   - Display: “failed” is not a safe conclusion.
2. **0:12–0:35 — Consequence**
   - Budget is released.
   - A second authorization becomes possible.
   - Display the double-payment risk.
3. **0:35–1:10 — Root cause**
   - Animate the lifecycle:
     `Intent → Policy → Session → Signature → Hook → Receipt → Reconcile`
   - Highlight where truth diverges.
4. **1:10–2:05 — Railguard correction**
   - Reserve budget atomically.
   - Use `unknown` after an ambiguous broadcast.
   - Reconcile by `executionDigest`.
   - Commit only after settlement fact is established.
5. **2:05–2:40 — Real proof**
   - Use genuine test output and testnet evidence.
   - Show command, result, and evidence location.
6. **2:40–3:05 — Honest boundary**
   - Show what v0.1 proves and what remains open.
   - CTA to portfolio and reproducible demo.

### P0 — 30-second launch film

**Title:** `Agent payments fail on glue, not validators`

- One failure
- One consequence
- One invariant
- One proof image
- One CTA

This is the primary social launch asset. It must work without sound and use captions.

### P0 — five-minute reproducible walkthrough

**Title:** `Reproduce four AI-agent payment failures`

Show:

1. repository and prerequisites;
2. one command;
3. vulnerable behavior;
4. fixed behavior;
5. evidence files;
6. current limitations.

Avoid cinematic material here. It should feel like a precise engineering walkthrough.

### P1 — eight technical micro-lessons

Convert the existing C01–C08 carousels into 25–45 second explainers:

1. Four bugs in agent-payment stacks
2. Three enforcement boundaries
3. The invariant pipeline
4. Atomic payment authorization
5. Why post-broadcast failure is dangerous
6. Execution digest versus FIFO reconciliation
7. Source of truth across app, provider, and chain
8. Honest production gaps

Each micro-lesson uses:

```text
Question → diagram → code/proof → takeaway → portfolio link
```

### P1 — founder/engineer series

Record five short, direct-to-camera explanations:

1. The bug I initially modeled incorrectly
2. Why confirmation is not enough
3. Where atomicity actually lives
4. Why replay identity must be explicit
5. What I would require before mainnet

These are more valuable for trust and hiring than an artificial digital-human presenter. Use the real engineer and edit tightly.

### P2 — ecosystem variants

Create focused 45–60 second versions for:

- x402 resource servers
- wallet and agent-payment projects
- stablecoin backends
- Base/CDP ecosystem reviewers
- Stellar ecosystem reviewers

Each version changes the opening failure, proof, and CTA—not merely the title.

## Role of the four video tools

### Remotion: canonical production engine

Use Remotion as the source of truth for:

- typography;
- lifecycle and state-machine animation;
- code and terminal scenes;
- charts and boundary diagrams;
- captions;
- aspect-ratio variants;
- deterministic batch rendering.

The existing HTML/Chrome generator and design tokens should feed the same visual system. Do not rebuild the Railguard design language independently inside every tool.

### HyperFrames: specialist motion layer

Use HyperFrames only for shots that are expensive or awkward in the canonical template:

- a polished opening sequence;
- a spatial architecture reveal;
- a transaction-state transition;
- a closing visual motif.

Render those shots with transparent or clean backgrounds and composite them into the Remotion master. Avoid maintaining two full video timelines.

### video-use: human-footage editor

Use video-use when real footage exists:

- founder/engineer camera takes;
- screen-recorded demos;
- interviews;
- launch commentary.

Its best role is:

```text
raw takes → transcript-led edit → precise cuts → captions → B-roll/overlay slots
```

Then place deterministic Railguard overlays into those slots. Its documented workflow includes transcript-based cutting, overlays, rendering, and cut-boundary self-evaluation.

### Seedance: illustrative bridge, not evidence

Use Seedance for no more than 5–10% of a technical video:

- an abstract opening metaphor;
- a payment moving through uncertain state;
- a short atmospheric transition;
- a visual representation of “truth divergence.”

Never use generated footage to depict:

- a real terminal result;
- a live transaction;
- a product UI;
- an integration;
- a security guarantee.

AI-generated visuals must be treated as illustration. Railguard’s differentiator is verifiable evidence.

## Closed-loop production system

The production loop should be:

```text
Content brief
  ↓
Claim and evidence map
  ↓
Script with proof IDs
  ↓
Fresh command/test capture
  ↓
Remotion scene generation
  ↓
Optional HyperFrames/Seedance inserts
  ↓
Optional video-use talking-head edit
  ↓
Master render
  ↓
Technical QA + visual QA + claim QA
  ↓
16:9, 9:16, 1:1 variants
  ↓
Publish
  ↓
Measure qualified actions
  ↓
Revise hook, proof, or CTA
```

The loop is not closed when a video renders. It is closed when a viewer can take a measurable technical action and the result improves the next version.

## Evidence and claim rules

Every technical video must pass these checks:

- Commands shown in the video run from a clean documented environment.
- Terminal footage is captured from actual output, not reconstructed for appearance.
- Testnet network, transaction hash, and evidence path are visible when relevant.
- The visual clearly distinguishes test, simulation, and live testnet activity.
- “Blocked,” “prevented,” and “safe” claims state their scope.
- Known gaps are not hidden in small end credits.
- No mainnet-production implication is allowed for v0.1.
- The portfolio link is the canonical destination.
- Captions, code, and hashes remain readable on a phone.
- A technical reviewer signs off on the final claim sheet before publication.

## Asset reuse map

| Existing asset | New use |
|---|---|
| `carousels/C01-four-bugs` | 30-second overview and four short hooks |
| `carousels/C02-three-boundaries` | animated architecture explainer |
| `carousels/C03-invariant-pipeline` | flagship lifecycle sequence |
| `carousels/C04-authorize-payment` | atomicity micro-lesson |
| `carousels/C05-post-broadcast` | flagship narrative core |
| `carousels/C06-execution-digest` | reconciliation micro-lesson |
| `carousels/C07-source-of-truth` | technical education episode |
| `carousels/C08-honest-gaps` | trust-building release-status film |
| `screenshots/terminal-*` | proof scenes, after fresh verification |
| `evidence/` | live/testnet claim anchors |
| `railguard-tokens.json` | shared motion and static design system |

## Four-week release sequence

### Week 0 — production preparation

- Freeze the launch message and status line.
- Define the flagship claim/evidence sheet.
- Verify every command and evidence link.
- Build one Remotion master template.
- Record the flagship demo and optional founder take.
- Render all aspect ratios.

Exit gate:

> One three-minute master, one 30-second cut, and one reproducible walkthrough pass technical and visual QA.

### Week 1 — problem launch

- Publish the 30-second hook.
- Publish the flagship proof film.
- Publish the postmortem: “My payment guard released budget after broadcast.”
- Pin the portfolio link.

CTA:

> Reproduce the failure.

### Week 2 — education

- Release post-broadcast, atomicity, and execution-digest micro-lessons.
- Publish the five-minute walkthrough.
- Answer technical questions with timestamped clips.

CTA:

> Run one profile and inspect one report.

### Week 3 — trust

- Release the honest-gaps video.
- Publish live testnet evidence explainers.
- Release a founder clip explaining one rejected design.

CTA:

> Review the threat model or open a technical issue.

### Week 4 — adoption

- Publish three ecosystem-specific integration films.
- Offer three small external integration PRs:
  - one x402 resource server;
  - one wallet/agent-payment project;
  - one stablecoin backend.
- Report real outcomes, including failed outreach and discovered incompatibilities.

CTA:

> Add one profile to CI.

## Metrics that matter

Do not optimize primarily for views.

Measure:

1. portfolio visits from video;
2. demo command runs or repository clones;
3. evidence-page visits;
4. serious issues or technical questions;
5. external repositories running the Action/profile;
6. accepted integration PRs;
7. maintainer acknowledgements;
8. repeat viewers of technical episodes.

The strongest launch result is:

> One credible external repository keeps Railguard’s failure check in CI.

## Production backlog

### Build now

- One shared motion design package using the existing design tokens
- One proof-scene component for commands, results, hashes, and evidence links
- One lifecycle/state-machine component
- One comparison scene for vulnerable versus fixed behavior
- One limitation/status scene
- One CTA scene
- Automated 16:9, 9:16, and 1:1 renders
- Caption and safe-area validation
- A per-video claim/evidence manifest

### Do not build yet

- a digital-human presenter;
- a large cinematic brand film;
- dozens of generic social clips;
- a separate visual identity for video;
- AI-generated fake dashboards;
- long tutorials before the five-minute path is proven;
- a content automation system without adoption metrics.

## First production sprint

The first sprint should produce exactly three videos:

1. **30 seconds:** Agent payments fail on glue, not validators.
2. **3 minutes:** The payment succeeded. The app said it failed.
3. **5 minutes:** Reproduce four AI-agent payment failures.

After publication, use viewer behavior and technical responses to decide whether the next investment should be:

- more education;
- integration walkthroughs;
- founder-led engineering stories;
- grant-specific evidence packages;
- product onboarding.

This sequence prevents the team from industrializing content before discovering which proof actually converts.
