/**
 * Test script: renders a few frames of the rotating-sphere animation
 * with a solid-color test image, saves them as PNGs to disk.
 *
 * Usage: npx tsx src/test-render.ts
 */
import { writeFile, mkdir } from "node:fs/promises";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import sharp from "sharp";
import { renderAnimation, shutdownBrowser } from "./renderer.js";

const __dirname = dirname(fileURLToPath(import.meta.url));
const OUT_DIR = resolve(__dirname, "../test-output");

async function createTestImage(width: number, height: number): Promise<Buffer> {
  // Create a simple gradient image using sharp
  // Red-to-blue gradient so it's easy to see rotation
  const channels = 4;
  const pixels = Buffer.alloc(width * height * channels);

  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const idx = (y * width + x) * channels;
      pixels[idx + 0] = Math.floor((x / width) * 255);     // R: left-to-right
      pixels[idx + 1] = Math.floor((y / height) * 255);     // G: top-to-bottom
      pixels[idx + 2] = Math.floor(((width - x) / width) * 255); // B: right-to-left
      pixels[idx + 3] = 255;                                 // A: fully opaque
    }
  }

  return sharp(pixels, { raw: { width, height, channels } })
    .png()
    .toBuffer();
}

async function main() {
  console.log("Creating test image...");
  const testImage = await createTestImage(256, 256);
  console.log(`Test image: ${testImage.length} bytes PNG`);

  await mkdir(OUT_DIR, { recursive: true });

  // Test rotating-sphere with 5 frames
  console.log("\nRendering rotating-sphere (5 frames)...");
  const t0 = Date.now();
  const result = await renderAnimation({
    animation: "rotating-sphere",
    image: testImage,
    params: {
      totalFrames: 5,
      size: 480,
      bgColor: "#1a1a2e",
    },
  });
  const elapsed = Date.now() - t0;

  console.log(`Rendered ${result.frames.length} frames in ${elapsed}ms (delay=${result.delay}ms)`);

  // Save frames
  for (let i = 0; i < result.frames.length; i++) {
    const path = resolve(OUT_DIR, `rotating-sphere-${String(i).padStart(3, "0")}.png`);
    await writeFile(path, result.frames[i]);
    console.log(`  Saved: ${path} (${result.frames[i].length} bytes)`);
  }

  // Test all other animations
  const otherAnimations = [
    { name: "pyramid", bgColor: "#2e1a1a" },
    { name: "inside-sphere", bgColor: "#1a2e1a" },
    { name: "low-poly-sphere", bgColor: "#2e2e1a" },
    { name: "360-spin", bgColor: "#1a1a2e" },
  ];

  for (const { name, bgColor } of otherAnimations) {
    console.log(`\nRendering ${name} (3 frames)...`);
    const t = Date.now();
    const res = await renderAnimation({
      animation: name,
      image: testImage,
      params: { totalFrames: 3, size: 480, bgColor },
    });
    const ms = Date.now() - t;
    console.log(`  ${res.frames.length} frames in ${ms}ms`);
    for (let i = 0; i < res.frames.length; i++) {
      const path = resolve(OUT_DIR, `${name}-${String(i).padStart(3, "0")}.png`);
      await writeFile(path, res.frames[i]);
      console.log(`  Saved: ${path} (${res.frames[i].length} bytes)`);
    }
  }

  // Shut down the browser
  await shutdownBrowser();
  console.log("\nDone! Check test-output/ directory for rendered frames.");
}

main().catch((err) => {
  console.error("Test failed:", err);
  shutdownBrowser().finally(() => process.exit(1));
});
