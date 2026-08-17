from __future__ import annotations

import asyncio
import json
import subprocess
from pathlib import Path

import edge_tts

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "src" / "content" / "manifest.json"
AUDIO = ROOT / "public" / "audio"
VOICE = "en-IN-PrabhatNeural"


async def generate_voiceovers() -> None:
    AUDIO.mkdir(parents=True, exist_ok=True)
    specs = json.loads(MANIFEST.read_text(encoding="utf-8"))
    jobs = []

    for spec in specs:
        target = AUDIO / spec["slug"]
        target.mkdir(parents=True, exist_ok=True)
        for index, scene in enumerate(spec["scenes"], start=1):
            output = target / f"{index:02d}-{scene['id']}.mp3"
            jobs.append(edge_tts.Communicate(scene["voice"], VOICE, rate="-3%").save(str(output)))

    await asyncio.gather(*jobs)


def generate_bed() -> None:
    output = AUDIO / "bed.mp3"
    command = [
        "ffmpeg",
        "-y",
        "-f",
        "lavfi",
        "-i",
        "sine=frequency=55:sample_rate=48000:duration=305",
        "-f",
        "lavfi",
        "-i",
        "sine=frequency=82.41:sample_rate=48000:duration=305",
        "-filter_complex",
        "[0:a]volume=0.16,tremolo=f=0.12:d=0.5[a0];"
        "[1:a]volume=0.08,tremolo=f=0.17:d=0.35[a1];"
        "[a0][a1]amix=inputs=2,lowpass=f=420,afade=t=in:d=2,"
        "afade=t=out:st=302:d=3,volume=0.7",
        "-c:a",
        "libmp3lame",
        "-b:a",
        "192k",
        str(output),
    ]
    subprocess.run(command, check=True)


if __name__ == "__main__":
    asyncio.run(generate_voiceovers())
    generate_bed()
    print(f"Generated narration and sound bed in {AUDIO}")
