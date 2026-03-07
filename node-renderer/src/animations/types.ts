import type p5 from "p5";

export interface AnimationContext {
  p5: p5;
  mainImage: p5.Image;
  size: number;
  currentFrame: number;
  totalFrames: number;
  bgColor: string;
  params: Record<string, number>;
}

export interface ParamSchema {
  type: "number";
  min: number;
  max: number;
  default: number;
  description?: string;
}

export interface AnimationDefinition {
  name: string;
  description: string;
  render: (ctx: AnimationContext) => void;
  params?: Record<string, ParamSchema>;
}
