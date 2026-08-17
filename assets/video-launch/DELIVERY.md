# Railguard launch-video delivery

## Status

Production complete. All nine MP4 outputs passed codec, audio-stream, duration, and dimension validation.

## Watch first

1. `renders/railguard-launch-30s-16x9.mp4`
2. `renders/railguard-flagship-3m-16x9.mp4`
3. `renders/railguard-walkthrough-5m-16x9.mp4`

## Upload-ready masters

| Asset | Widescreen | Vertical | Square |
|---|---|---|---|
| 30-second launch | `railguard-launch-30s-16x9.mp4` | `railguard-launch-30s-9x16.mp4` | `railguard-launch-30s-1x1.mp4` |
| 3-minute flagship | `railguard-flagship-3m-16x9.mp4` | `railguard-flagship-3m-9x16.mp4` | `railguard-flagship-3m-1x1.mp4` |
| 5-minute walkthrough | `railguard-walkthrough-5m-16x9.mp4` | `railguard-walkthrough-5m-9x16.mp4` | `railguard-walkthrough-5m-1x1.mp4` |

All files are in `renders/`.

## QA result

```text
30-second family:   30.06s · H.264 · AAC · PASS
3-minute family:   180.05s · H.264 · AAC · PASS
5-minute family:   300.05s · H.264 · AAC · PASS
```

Representative wide compositions were additionally reviewed as six-, twelve-, and fifteen-scene contact sheets. Vertical lifecycle layout and widescreen proof imagery were manually inspected.

## Publishing order

1. Publish the 30-second widescreen or vertical launch film.
2. Link to the three-minute flagship as the primary proof.
3. Publish the five-minute walkthrough for technical reviewers.
4. Use `distribution/launch-copy.md` for platform-specific copy.
5. Use the matching image from `renders/thumbnails/`.
6. Keep `docs/PORTFOLIO.md` as the canonical destination.

## Editable production source

- `src/content/manifest.json` — canonical scene copy, timing, and narration
- `src/components/` — responsive motion system
- `evidence/claims.json` — claim-to-proof mapping
- `seedance/abstract-bridges.md` — optional cinematic illustration prompts
- `footage/edit/project.md` — video-use memory for future founder footage
- `scripts/render-all.ps1` — deterministic nine-format render
- `scripts/qa-renders.ps1` — automated media validation

## Release boundary

Every output states:

> reference implementation · not mainnet production

Do not remove this line until the production-readiness gaps documented in the repository have been closed.
