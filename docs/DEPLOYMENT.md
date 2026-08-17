# Deployment & operations

Last updated: 2026-08-17

> **Workspace index:** [WORKSPACE.md](./WORKSPACE.md) · **v5 status:** [v5execution.md](./v5execution.md)

## Live URLs

| Surface | URL |
|---------|-----|
| API (staging) | https://staging-railguard-s4ii.encr.app |
| Web console | https://prebroadcast.vercel.app |

Browser calls Encore directly or via Vercel proxy depending on env config.

## Deploy API (Encore)

Migrations apply automatically on Encore deploy (latest: `011_v5_financial_intents` in coinbase).

```powershell
cd coinbase
git push encore main
```

Verify after deploy:

```powershell
$env:RAILGUARD_BASE_URL = "https://staging-railguard-s4ii.encr.app"
cd coinbase
bun run verify:demo
```

## Deploy web (Vercel)

Vercel auto-deploys from `prasanthkuna/railguard-cdp` on push to `main`.

Production URL: **https://prebroadcast.vercel.app**

Set `NEXT_PUBLIC_API_URL=https://staging-railguard-s4ii.encr.app` on Vercel.

## Repo map

| Legacy name | Folder | GitHub |
|-------------|--------|--------|
| railguard-cdp | `coinbase/` | prasanthkuna/railguard-cdp |
| railguard-protocol | `railguard-new/` | prasanthkuna/railguard-new |
