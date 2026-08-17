# Railguard launch video system

Manifest-driven Remotion production for the Railguard proof-led release.

## Outputs

- 30-second launch film
- 3-minute flagship proof film
- 5-minute reproducible walkthrough
- 16:9, 9:16, and 1:1 compositions for each
- Neural narration and generated ambient sound bed
- Claim-to-evidence manifest
- Seedance abstract bridge prompts
- video-use founder-footage intake memory
- Platform-ready launch copy and descriptions
- Automated render QA

## Build

```powershell
cd assets/video-launch
npm install
pip install edge-tts
npm run audio
npm run check
npm run render:all
powershell -ExecutionPolicy Bypass -File scripts/qa-renders.ps1
```

## Preview

```powershell
npm run studio
```

## Content changes

Edit `src/content/manifest.json`. It is the canonical source for scene order, duration, visible copy, narration, proof assets, and format variants.

## Evidence policy

`evidence/claims.json` maps public claims to repository proof. Generated footage may be used only as illustration and never as terminal, product, transaction, or security evidence.
