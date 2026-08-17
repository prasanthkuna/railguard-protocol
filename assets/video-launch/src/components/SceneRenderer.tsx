import {
  AbsoluteFill,
  Img,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";
import type {SceneSpec} from "../types";
import {accentColor, colors, mono, sans} from "../theme";
import {media} from "../media";

const Flow = ({accent}: {accent: string}) => {
  const frame = useCurrentFrame();
  const items = ["Intent", "Policy", "Session", "Signature", "Hook", "Receipt", "Reconcile"];
  return (
    <div style={{display: "flex", flexWrap: "wrap", gap: 12, alignItems: "center", marginTop: 42}}>
      {items.map((item, index) => {
        const reveal = spring({frame: frame - index * 4, fps: 24, config: {damping: 18}});
        return (
          <div key={item} style={{display: "flex", alignItems: "center", gap: 12, opacity: reveal}}>
            <div
              style={{
                border: `1px solid ${index === items.length - 1 ? accent : colors.border}`,
                background: index === items.length - 1 ? `${accent}18` : colors.panel,
                borderRadius: 12,
                padding: "14px 18px",
                fontFamily: mono,
                fontSize: 17,
              }}
            >
              {item}
            </div>
            {index < items.length - 1 ? <span style={{color: colors.muted}}>→</span> : null}
          </div>
        );
      })}
    </div>
  );
};

const StatusNodes = ({accent}: {accent: string}) => {
  const frame = useCurrentFrame();
  const pulse = interpolate(Math.sin(frame / 6), [-1, 1], [0.35, 1]);
  return (
    <div style={{display: "grid", gridTemplateColumns: "1fr auto 1fr", gap: 24, alignItems: "center", marginTop: 48}}>
      <div style={{padding: 26, borderRadius: 16, background: colors.panel, border: `1px solid ${colors.green}`}}>
        <div style={{fontFamily: mono, color: colors.green, fontSize: 15}}>CHAIN</div>
        <div style={{fontSize: 25, fontWeight: 750, marginTop: 8}}>Broadcast</div>
      </div>
      <div style={{fontFamily: mono, color: colors.muted, fontSize: 24}}>≠</div>
      <div style={{padding: 26, borderRadius: 16, background: colors.panel, border: `1px solid ${accent}`, boxShadow: `0 0 ${24 * pulse}px ${accent}44`}}>
        <div style={{fontFamily: mono, color: accent, fontSize: 15}}>APPLICATION</div>
        <div style={{fontSize: 25, fontWeight: 750, marginTop: 8}}>Unknown</div>
      </div>
    </div>
  );
};

export const SceneRenderer = ({scene}: {scene: SceneSpec}) => {
  const frame = useCurrentFrame();
  const {fps, width, height, durationInFrames} = useVideoConfig();
  const vertical = height > width;
  const accent = accentColor(scene.accent);
  const enter = spring({frame, fps, config: {damping: 18, stiffness: 120}});
  const exit = interpolate(frame, [durationInFrames - 9, durationInFrames], [1, 0], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const y = interpolate(enter, [0, 1], [34, 0]);
  const image = scene.image ? media[scene.image] : undefined;
  const maxWidth = vertical ? width - 96 : Math.min(width - 128, 1460);

  return (
    <AbsoluteFill
      style={{
        padding: vertical ? "170px 48px 130px" : "146px 64px 100px",
        display: "flex",
        justifyContent: "center",
        opacity: exit,
      }}
    >
      <div
        style={{
          width: "100%",
          maxWidth,
          display: "grid",
          gridTemplateColumns: !vertical && image ? "0.95fr 1.05fr" : "1fr",
          gap: vertical ? 34 : 54,
          alignItems: "center",
          transform: `translateY(${y}px)`,
          opacity: enter,
        }}
      >
        <div>
          <div
            style={{
              color: accent,
              fontFamily: mono,
              fontSize: vertical ? 19 : 17,
              fontWeight: 700,
              letterSpacing: ".13em",
              marginBottom: 24,
            }}
          >
            {scene.kicker}
          </div>
          <h1
            style={{
              fontFamily: sans,
              fontSize: vertical ? 61 : image ? 63 : 76,
              lineHeight: 1.04,
              letterSpacing: "-.045em",
              margin: 0,
              maxWidth: vertical ? "100%" : 1320,
            }}
          >
            {scene.title}
          </h1>
          {scene.body ? (
            <p
              style={{
                color: colors.muted,
                fontSize: vertical ? 28 : 27,
                lineHeight: 1.42,
                maxWidth: 1120,
                margin: "26px 0 0",
              }}
            >
              {scene.body}
            </p>
          ) : null}
          {scene.kind === "pipeline" ? <Flow accent={accent} /> : null}
          {scene.kind === "failure" ? <StatusNodes accent={accent} /> : null}
          {scene.bullets ? (
            <div
              style={{
                display: "grid",
                gridTemplateColumns: vertical ? "1fr" : "1fr 1fr",
                gap: 14,
                marginTop: 34,
              }}
            >
              {scene.bullets.map((bullet, index) => (
                <div
                  key={bullet}
                  style={{
                    padding: vertical ? "19px 21px" : "17px 20px",
                    borderRadius: 14,
                    border: `1px solid ${index === 0 ? `${accent}99` : colors.border}`,
                    background: colors.panel,
                    fontSize: vertical ? 23 : 19,
                  }}
                >
                  <span style={{color: accent, fontFamily: mono, marginRight: 12}}>
                    {String(index + 1).padStart(2, "0")}
                  </span>
                  {bullet}
                </div>
              ))}
            </div>
          ) : null}
          {scene.code ? (
            <div
              style={{
                marginTop: 34,
                padding: vertical ? 24 : 28,
                border: `1px solid ${colors.border}`,
                borderRadius: 16,
                background: "#080a0e",
                fontFamily: mono,
                fontSize: vertical ? 21 : 20,
                lineHeight: 1.75,
                boxShadow: "0 24px 70px rgba(0,0,0,.35)",
              }}
            >
              {scene.code.map((line, index) => (
                <div key={line} style={{color: index === 0 ? accent : colors.text}}>
                  <span style={{color: colors.muted, marginRight: 16}}>{index === 0 ? "$" : "›"}</span>
                  {line}
                </div>
              ))}
            </div>
          ) : null}
          {scene.kind === "comparison" ? (
            <div style={{display: "grid", gridTemplateColumns: "1fr 1fr", gap: 18, marginTop: 42}}>
              <div style={{border: `1px solid ${colors.red}`, background: `${colors.red}0c`, borderRadius: 16, padding: 25}}>
                <div style={{fontFamily: mono, color: colors.red, fontSize: 15}}>UNSAFE</div>
                <div style={{fontWeight: 760, fontSize: vertical ? 25 : 24, marginTop: 12}}>Read → pay → record</div>
              </div>
              <div style={{border: `1px solid ${colors.green}`, background: `${colors.green}0c`, borderRadius: 16, padding: 25}}>
                <div style={{fontFamily: mono, color: colors.green, fontSize: 15}}>RAILGUARD</div>
                <div style={{fontWeight: 760, fontSize: vertical ? 25 : 24, marginTop: 12}}>Reserve → execute → reconcile</div>
              </div>
            </div>
          ) : null}
          {scene.kind === "cta" ? (
            <div
              style={{
                marginTop: 42,
                height: 3,
                width: interpolate(enter, [0, 1], [0, vertical ? 460 : 640]),
                background: accent,
                boxShadow: `0 0 24px ${accent}`,
              }}
            />
          ) : null}
        </div>
        {image ? (
          <div
            style={{
              position: "relative",
              padding: 8,
              borderRadius: 20,
              background: `linear-gradient(135deg, ${accent}aa, ${colors.border}, transparent)`,
              boxShadow: `0 30px 100px ${accent}1a`,
            }}
          >
            <Img
              src={image}
              style={{
                display: "block",
                width: "100%",
                maxHeight: vertical ? 720 : 650,
                objectFit: "cover",
                objectPosition: "center",
                borderRadius: 14,
              }}
            />
            <div
              style={{
                position: "absolute",
                top: 24,
                right: 24,
                borderRadius: 999,
                padding: "9px 14px",
                background: "#07090ddd",
                border: `1px solid ${accent}`,
                color: accent,
                fontFamily: mono,
                fontSize: 14,
              }}
            >
              VERIFIED ARTIFACT
            </div>
          </div>
        ) : null}
      </div>
    </AbsoluteFill>
  );
};
