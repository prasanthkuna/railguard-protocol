export type Accent = "blue" | "red" | "green" | "amber" | "cyan" | "purple";

export type SceneKind =
  | "hook"
  | "failure"
  | "pipeline"
  | "comparison"
  | "proof"
  | "architecture"
  | "honest"
  | "steps"
  | "cta";

export type SceneSpec = {
  id: string;
  kind: SceneKind;
  seconds: number;
  kicker: string;
  title: string;
  body?: string;
  bullets?: string[];
  code?: string[];
  accent?: Accent;
  image?: "terminal-cdp" | "terminal-intent" | "architecture";
  voice: string;
};

export type VideoSpec = {
  slug: string;
  label: string;
  title: string;
  scenes: SceneSpec[];
};
