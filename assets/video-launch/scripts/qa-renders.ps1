$ErrorActionPreference = "Stop"

$expected = @(
    @{ File = "railguard-launch-30s-16x9.mp4"; Duration = 30; Width = 1920; Height = 1080 },
    @{ File = "railguard-launch-30s-9x16.mp4"; Duration = 30; Width = 1080; Height = 1920 },
    @{ File = "railguard-launch-30s-1x1.mp4"; Duration = 30; Width = 1080; Height = 1080 },
    @{ File = "railguard-flagship-3m-16x9.mp4"; Duration = 180; Width = 1920; Height = 1080 },
    @{ File = "railguard-flagship-3m-9x16.mp4"; Duration = 180; Width = 1080; Height = 1920 },
    @{ File = "railguard-flagship-3m-1x1.mp4"; Duration = 180; Width = 1080; Height = 1080 },
    @{ File = "railguard-walkthrough-5m-16x9.mp4"; Duration = 300; Width = 1920; Height = 1080 },
    @{ File = "railguard-walkthrough-5m-9x16.mp4"; Duration = 300; Width = 1080; Height = 1920 },
    @{ File = "railguard-walkthrough-5m-1x1.mp4"; Duration = 300; Width = 1080; Height = 1080 }
)

$results = foreach ($item in $expected) {
    $path = Join-Path "renders" $item.File
    if (-not (Test-Path -LiteralPath $path)) {
        throw "Missing render: $path"
    }

    $probe = ffprobe -v error -show_entries format=duration,size -show_entries stream=codec_type,codec_name,width,height -of json $path | ConvertFrom-Json
    $video = $probe.streams | Where-Object codec_type -eq "video"
    $audio = $probe.streams | Where-Object codec_type -eq "audio"
    $duration = [double]$probe.format.duration

    if (-not $video -or -not $audio) {
        throw "Missing video or audio stream: $path"
    }
    if ($video.width -ne $item.Width -or $video.height -ne $item.Height) {
        throw "Unexpected dimensions: $path"
    }
    if ([Math]::Abs($duration - $item.Duration) -gt 0.2) {
        throw "Unexpected duration: $path ($duration seconds)"
    }

    [pscustomobject]@{
        File = $item.File
        Duration = [Math]::Round($duration, 2)
        Frame = "$($video.width)x$($video.height)"
        Video = $video.codec_name
        Audio = $audio.codec_name
        MB = [Math]::Round(([double]$probe.format.size / 1MB), 1)
        Status = "PASS"
    }
}

$results | Format-Table -AutoSize
