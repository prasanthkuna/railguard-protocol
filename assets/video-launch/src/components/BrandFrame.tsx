import type {ReactNode} from "react";
import {AbsoluteFill, interpolate, useCurrentFrame, useVideoConfig} from "remotion";
import {colors, mono, sans} from "../theme";

export const BrandFrame = ({
  children,
  label,
  sceneIndex,
  sceneCount,
}: {
  children: ReactNode;
  label: string;
  sceneIndex: number;
  sceneCount: number;
}) => {
  const frame = useCurrentFrame();
  const {durationInFrames, width, height} = useVideoConfig();
  const vertical = height > width;
  const progress = interpolate(frame, [0, durationInFrames], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  return (
    <AbsoluteFill
      style={{
        background: colors.bg,
        color: colors.text,
        fontFamily: sans,
        overflow: "hidden",
      }}
    >
      <AbsoluteFill
        style={{
          opacity: 0.28,
          backgroundImage:
            "linear-gradient(rgba(255,255,255,.035) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,.035) 1px, transparent 1px)",
          backgroundSize: `${vertical ? 54 : 72}px ${vertical ? 54 : 72}px`,
          maskImage: "linear-gradient(to bottom, black, transparent 82%)",
        }}
      />
      <AbsoluteFill
        style={{
          background:
            "radial-gradient(circle at 16% 10%, rgba(59,130,246,.15), transparent 32%), radial-gradient(circle at 86% 72%, rgba(167,139,250,.08), transparent 30%)",
        }}
      />
      <div
        style={{
          position: "absolute",
          top: vertical ? 52 : 44,
          left: vertical ? 48 : 64,
          right: vertical ? 48 : 64,
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          fontSize: vertical ? 20 : 17,
          zIndex: 5,
        }}
      >
        <div style={{display: "flex", gap: 14, alignItems: "center"}}>
          <span style={{fontWeight: 800, color: colors.blue, fontSize: vertical ? 25 : 22}}>
            Railguard
          </span>
          <span style={{color: colors.muted, fontFamily: mono}}>v0.1-reference</span>
        </div>
        <span style={{fontFamily: mono, color: colors.muted}}>
          {String(sceneIndex + 1).padStart(2, "0")} / {String(sceneCount).padStart(2, "0")}
        </span>
      </div>
      <div
        style={{
          position: "absolute",
          top: vertical ? 116 : 100,
          left: vertical ? 48 : 64,
          right: vertical ? 48 : 64,
          height: 2,
          background: colors.border,
          zIndex: 5,
        }}
      >
        <div
          style={{
            width: `${progress * 100}%`,
            height: "100%",
            background: colors.blue,
            boxShadow: `0 0 20px ${colors.blue}`,
          }}
        />
      </div>
      {children}
      <div
        style={{
          position: "absolute",
          bottom: vertical ? 44 : 34,
          left: vertical ? 48 : 64,
          right: vertical ? 48 : 64,
          display: "flex",
          justifyContent: "space-between",
          color: colors.muted,
          fontFamily: mono,
          fontSize: vertical ? 16 : 13,
          letterSpacing: ".02em",
          zIndex: 5,
        }}
      >
        <span>{label}</span>
        <span>reference implementation · not mainnet production</span>
      </div>
    </AbsoluteFill>
  );
};
