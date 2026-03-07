import { AnimationDefinition } from "./types.js";

export const lowPolySphere: AnimationDefinition = {
  name: "low-poly-sphere",
  description: "Faceted low-poly sphere with configurable detail level.",
  render({ p5, mainImage, size, currentFrame, totalFrames, bgColor, params }) {
    const detailY = params.detail ?? 6;

    p5.background(bgColor);
    p5.rotateY(Math.PI + currentFrame * (2 * Math.PI / totalFrames));
    p5.texture(mainImage);
    (p5 as any).sphere(size / 3, 6, detailY);
  },
  params: {
    detail: { type: "number", min: 2, max: 20, default: 6, description: "Polygon subdivision level (lower = more faceted)" },
  },
};
