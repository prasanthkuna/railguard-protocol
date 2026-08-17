import type {Accent} from "./types";

export const colors = {
  bg: "#07090d",
  panel: "#11151c",
  panelSoft: "#0d1117",
  border: "#2a303b",
  text: "#f4f4f5",
  muted: "#a1a1aa",
  blue: "#3b82f6",
  red: "#ef4444",
  green: "#22c55e",
  amber: "#f59e0b",
  cyan: "#22d3ee",
  purple: "#a78bfa",
};

export const accentColor = (accent: Accent = "blue") => colors[accent];

export const mono = '"JetBrains Mono", "Cascadia Code", monospace';
export const sans = 'Inter, "Segoe UI", sans-serif';
