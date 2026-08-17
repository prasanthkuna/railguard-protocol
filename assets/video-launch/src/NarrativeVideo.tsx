import "@fontsource/inter/400.css";
import "@fontsource/inter/700.css";
import "@fontsource/inter/800.css";
import "@fontsource/jetbrains-mono/400.css";
import "@fontsource/jetbrains-mono/700.css";
import {Audio} from "@remotion/media";
import {Sequence, staticFile} from "remotion";
import type {VideoSpec} from "./types";
import {BrandFrame} from "./components/BrandFrame";
import {SceneRenderer} from "./components/SceneRenderer";

export const FPS = 24;

export const durationFor = (spec: VideoSpec) =>
  spec.scenes.reduce((sum, scene) => sum + Math.round(scene.seconds * FPS), 0);

export const NarrativeVideo = ({spec}: {spec: VideoSpec}) => {
  let cursor = 0;
  return (
    <>
      <Audio src={staticFile("audio/bed.mp3")} volume={0.07} />
      {spec.scenes.map((scene, index) => {
        const duration = Math.round(scene.seconds * FPS);
        const from = cursor;
        cursor += duration;
        return (
          <Sequence key={scene.id} from={from} durationInFrames={duration} premountFor={FPS}>
            <BrandFrame
              label={spec.label}
              sceneIndex={index}
              sceneCount={spec.scenes.length}
            >
              <SceneRenderer scene={scene} />
            </BrandFrame>
            <Audio
              src={staticFile(`audio/${spec.slug}/${String(index + 1).padStart(2, "0")}-${scene.id}.mp3`)}
              volume={1}
            />
          </Sequence>
        );
      })}
    </>
  );
};
