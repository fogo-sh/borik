import { AnimationDefinition } from "./types.js";

export const rotatingSphere: AnimationDefinition = {
  name: "rotating-sphere",
  description: "Image textured on a rotating sphere.",
  render({ p5, mainImage, size, currentFrame, totalFrames, bgColor, params }) {
    const wobbliness = params.wobbliness ?? 0;
    const angle = currentFrame * (2 * Math.PI / totalFrames);

    p5.background(bgColor);
    p5.rotateY(Math.PI + angle);
    p5.rotateZ(Math.sin(angle) * wobbliness / 20);
    p5.rotateX(Math.sin(angle) * wobbliness / 20);
    p5.texture(mainImage);
    p5.sphere(size / 3);
  },
  params: {
    wobbliness: { type: "number", min: 0, max: 100, default: 0, description: "Amount of wobble during rotation" },
  },
};
