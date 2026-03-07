import { AnimationDefinition } from "./types.js";
import { getScaleFactor } from "./utils.js";

export const spin360: AnimationDefinition = {
  name: "360-spin",
  description: "Flat image rotating 360 degrees in 3D space.",
  render({ p5, mainImage, size, currentFrame, totalFrames, bgColor, params }) {
    const extraScale = (params.scale ?? 0) / 100;

    p5.background(bgColor);
    p5.scale(getScaleFactor({ mainImage, size }) + extraScale);
    p5.rotateY(currentFrame * (2 * Math.PI / totalFrames));
    p5.texture(mainImage);
    p5.plane(mainImage.width, mainImage.height);
  },
  params: {
    scale: { type: "number", min: 0, max: 100, default: 0, description: "Additional scale factor" },
  },
};
