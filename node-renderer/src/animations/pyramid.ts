import { AnimationDefinition } from "./types.js";

export const pyramid: AnimationDefinition = {
  name: "pyramid",
  description: "Image mapped onto a rotating 4-sided pyramid.",
  render({ p5, mainImage, size, currentFrame, totalFrames, bgColor }) {
    p5.background(bgColor);
    (p5 as any).textureMode((p5 as any).NORMAL);

    p5.rotateY(currentFrame * (2 * Math.PI / totalFrames));
    p5.scale(0.55);
    p5.translate(-size / 2, -size / 1.75, -size / 2);

    p5.texture(mainImage);

    // Front face
    p5.beginShape();
    (p5 as any).vertex(0, size, 0, 0, 1);
    (p5 as any).vertex(size / 2, 0, size / 2, 0.5, 0);
    (p5 as any).vertex(size, size, 0, 1, 1);
    p5.endShape();

    // Right face
    p5.beginShape();
    (p5 as any).vertex(size, size, 0, 0, 1);
    (p5 as any).vertex(size / 2, 0, size / 2, 0.5, 0);
    (p5 as any).vertex(size, size, size, 1, 1);
    p5.endShape();

    // Left face
    p5.beginShape();
    (p5 as any).vertex(0, size, 0, 0, 1);
    (p5 as any).vertex(size / 2, 0, size / 2, 0.5, 0);
    (p5 as any).vertex(0, size, size, 1, 1);
    p5.endShape();

    // Back face
    p5.beginShape();
    (p5 as any).vertex(0, size, size, 0, 1);
    (p5 as any).vertex(size / 2, 0, size / 2, 0.5, 0);
    (p5 as any).vertex(size, size, size, 1, 1);
    p5.endShape();
  },
};
