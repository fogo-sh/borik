import type p5 from "p5";

export function getScaleFactor(opts: {
  mainImage: p5.Image;
  size: number;
  percentage?: number;
}): number {
  const { mainImage, size, percentage = 0.63 } = opts;
  const maxDim = mainImage.width > mainImage.height ? mainImage.width : mainImage.height;
  return (size / maxDim) * percentage;
}
