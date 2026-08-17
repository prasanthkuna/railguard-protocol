import {Composition} from "remotion";
import manifest from "./content/manifest.json";
import {durationFor, FPS, NarrativeVideo} from "./NarrativeVideo";
import type {VideoSpec} from "./types";

const specs = manifest as VideoSpec[];

const variants = [
  {suffix: "Wide", width: 1920, height: 1080},
  {suffix: "Vertical", width: 1080, height: 1920},
  {suffix: "Square", width: 1080, height: 1080},
] as const;

const names: Record<string, string> = {
  "launch-30": "Launch30",
  "flagship-180": "Flagship",
  "walkthrough-300": "Walkthrough",
};

export const RailguardVideoRoot = () => (
  <>
    {specs.flatMap((spec) =>
      variants.map((variant) => (
        <Composition
          key={`${spec.slug}-${variant.suffix}`}
          id={`${names[spec.slug]}-${variant.suffix}`}
          component={NarrativeVideo}
          durationInFrames={durationFor(spec)}
          fps={FPS}
          width={variant.width}
          height={variant.height}
          defaultProps={{spec}}
        />
      )),
    )}
  </>
);
