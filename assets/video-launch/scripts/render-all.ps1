$ErrorActionPreference = "Stop"

$renders = @(
    @{ Id = "Launch30-Wide"; Out = "railguard-launch-30s-16x9.mp4" },
    @{ Id = "Launch30-Vertical"; Out = "railguard-launch-30s-9x16.mp4" },
    @{ Id = "Launch30-Square"; Out = "railguard-launch-30s-1x1.mp4" },
    @{ Id = "Flagship-Wide"; Out = "railguard-flagship-3m-16x9.mp4" },
    @{ Id = "Flagship-Vertical"; Out = "railguard-flagship-3m-9x16.mp4" },
    @{ Id = "Flagship-Square"; Out = "railguard-flagship-3m-1x1.mp4" },
    @{ Id = "Walkthrough-Wide"; Out = "railguard-walkthrough-5m-16x9.mp4" },
    @{ Id = "Walkthrough-Vertical"; Out = "railguard-walkthrough-5m-9x16.mp4" },
    @{ Id = "Walkthrough-Square"; Out = "railguard-walkthrough-5m-1x1.mp4" }
)

New-Item -ItemType Directory -Force -Path "renders" | Out-Null

foreach ($render in $renders) {
    $outputPath = "renders/$($render.Out)"
    if (Test-Path -LiteralPath $outputPath) {
        Write-Host "Skipping existing $($render.Out)"
        continue
    }
    Write-Host "Rendering $($render.Id)"
    & npx remotion render src/index.ts $render.Id $outputPath --codec=h264 --crf=19 --log=error
    if ($LASTEXITCODE -ne 0) {
        throw "Render failed: $($render.Id)"
    }
}
