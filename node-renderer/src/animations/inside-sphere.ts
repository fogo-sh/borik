import { AnimationDefinition } from "./types.js";

export const insideSphere: AnimationDefinition = {
  name: "inside-sphere",
  description: "Camera inside a large textured sphere, looking around.",
  render({ p5, mainImage, size, currentFrame, totalFrames, bgColor }) {
    p5.background(bgColor);
    p5.texture(mainImage);
    p5.rotateY(Math.sin(currentFrame / totalFrames * Math.PI * 4));
    p5.sphere(2 * size + 100);
  },
};
