import puppeteer, { type Browser } from "puppeteer";
import { readFile } from "node:fs/promises";
import { resolve, dirname } from "node:path";
import { fileURLToPath } from "node:url";
import { getAnimation, listAnimations } from "./animations/index.js";
import { getScaleFactor } from "./animations/utils.js";

const __dirname = dirname(fileURLToPath(import.meta.url));

export interface RenderRequest {
  animation: string;
  image: Buffer;
  params: {
    totalFrames?: number;
    size?: number;
    bgColor?: string;
    [key: string]: any;
  };
}

export interface RenderResult {
  frames: Buffer[];
  delay: number;
}

let browserInstance: Browser | null = null;
let p5Source: string | null = null;

async function getP5Source(): Promise<string> {
  if (!p5Source) {
    const p5Path = resolve(__dirname, "../node_modules/p5/lib/p5.min.js");
    p5Source = await readFile(p5Path, "utf-8");
  }
  return p5Source;
}

async function getBrowser(): Promise<Browser> {
  if (!browserInstance || !browserInstance.connected) {
    browserInstance = await puppeteer.launch({
      headless: true,
      args: [
        "--no-sandbox",
        "--disable-setuid-sandbox",
        "--disable-dev-shm-usage",
        "--use-gl=angle",
        "--enable-webgl",
        "--ignore-gpu-blocklist",
      ],
    });
  }
  return browserInstance;
}

export async function shutdownBrowser(): Promise<void> {
  if (browserInstance) {
    await browserInstance.close();
    browserInstance = null;
  }
}

/**
 * Serialize a render function to valid standalone JS.
 *
 * When TypeScript compiles object method shorthand like:
 *   { render({ p5, ... }) { ... } }
 * calling .toString() yields: "render({p5,...}){...}"
 * which is a method definition, not a standalone function expression.
 * We prepend "function " to make it valid.
 *
 * Also patches transpiler-generated module references:
 * - (0, import_utils.getScaleFactor)  →  getScaleFactor
 */
function serializeRenderFn(animName: string): string {
  const anim = getAnimation(animName);
  if (!anim) throw new Error(`Unknown animation: ${animName}`);

  let src = anim.render.toString();

  // Fix method shorthand → function expression
  if (src.startsWith("render(") || src.startsWith("render (")) {
    src = "function " + src;
  }

  // Patch transpiled import references.
  // tsx/esbuild rewrites `import { getScaleFactor } from "./utils.js"`
  // to something like `(0, import_utils.getScaleFactor)`.
  // We replace these with direct calls to the global helper we inject.
  src = src.replace(/\(0,\s*import_utils\.(\w+)\)/g, "$1");

  return src;
}

/**
 * Build a <script> block that declares all helper functions used by
 * animation render functions as globals in the browser page.
 */
function buildHelperScript(): string {
  return `var getScaleFactor = ${getScaleFactor.toString()};`;
}

export async function renderAnimation(req: RenderRequest): Promise<RenderResult> {
  const { animation, image, params } = req;
  const totalFrames = params.totalFrames ?? 20;
  const size = params.size ?? 480;
  const bgColor = params.bgColor ?? "#1a1a2e";

  // Validate animation exists
  const animDef = getAnimation(animation);
  if (!animDef) {
    throw new Error(
      `Unknown animation: "${animation}". Available: ${listAnimations().map((a) => a.name).join(", ")}`,
    );
  }

  // Merge default params from the animation definition
  const mergedParams: Record<string, number> = {};
  if (animDef.params) {
    for (const [key, schema] of Object.entries(animDef.params)) {
      mergedParams[key] = (params[key] as number) ?? schema.default;
    }
  }

  const renderFnSource = serializeRenderFn(animation);
  const helperScript = buildHelperScript();
  const imageBase64 = image.toString("base64");
  const p5Src = await getP5Source();

  const browser = await getBrowser();
  const page = await browser.newPage();

  try {
    // Forward browser console/errors for debugging
    page.on("console", (msg) => {
      if (msg.type() === "error") {
        console.error("[browser]", msg.text());
      }
    });
    page.on("pageerror", (err: unknown) => {
      console.error("[browser error]", err instanceof Error ? err.message : String(err));
    });

    await page.setViewport({ width: size, height: size });

    const html = buildHtmlPage(p5Src, helperScript, size);
    await page.setContent(html, { waitUntil: "domcontentloaded" });

    // Wait for p5 constructor to be available
    await page.waitForFunction("typeof p5 !== 'undefined'", { timeout: 10000 });

    // Render all frames inside the browser
    const frameDataUrls = await page.evaluate(
      (renderFnStr, imgBase64, totalFrames, size, bgColor, mergedParams) => {
        return new Promise<string[]>((resolvePromise, reject) => {
          try {
            // eslint-disable-next-line no-new-func
            const renderFn = new Function("return (" + renderFnStr + ")")();

            let img: any = null;
            let sketch: any = null;

            const sketchFn = (p: any) => {
              p.preload = () => {
                img = p.loadImage("data:image/png;base64," + imgBase64);
              };

              p.setup = () => {
                p.createCanvas(size, size, p.WEBGL);
                p.pixelDensity(1);
                p.noLoop();
                sketch = p;
              };

              p.draw = () => {
                // We drive the render loop manually, not via p5's loop.
              };
            };

            const container = document.getElementById("canvas-container")!;
            new (window as any).p5(sketchFn, container);

            // Poll until p5 setup completes and image is loaded
            const waitForReady = (): Promise<void> =>
              new Promise((res) => {
                const tick = () => {
                  if (sketch && img && img.width > 0) res();
                  else setTimeout(tick, 50);
                };
                tick();
              });

            waitForReady()
              .then(() => {
                const canvas = document.querySelector("canvas");
                if (!canvas) throw new Error("No canvas element found");

                const frames: string[] = [];
                for (let frame = 0; frame < totalFrames; frame++) {
                  // Reset the transform matrix before each frame
                  sketch.push();

                  renderFn({
                    p5: sketch,
                    mainImage: img,
                    size,
                    currentFrame: frame,
                    totalFrames,
                    bgColor,
                    params: mergedParams,
                  });

                  sketch.pop();

                  frames.push(canvas.toDataURL("image/png"));
                }
                resolvePromise(frames);
              })
              .catch((err: any) => reject(err?.message ?? String(err)));
          } catch (err: any) {
            reject(err?.message ?? String(err));
          }
        });
      },
      renderFnSource,
      imageBase64,
      totalFrames,
      size,
      bgColor,
      mergedParams,
    );

    // Convert data URLs → Buffers
    const frames = frameDataUrls.map((dataUrl) => {
      const base64Data = dataUrl.replace(/^data:image\/png;base64,/, "");
      return Buffer.from(base64Data, "base64");
    });

    const delay = 50; // 50ms per frame → 20 fps

    return { frames, delay };
  } finally {
    await page.close();
  }
}

function buildHtmlPage(p5Source: string, helperScript: string, size: number): string {
  // esbuild/tsx's --keep-names transform wraps named const assignments with
  // __name(fn, "name"). This helper exists in the Node runtime but not in the
  // browser, so we polyfill it here.
  const esbuildPolyfill = `var __name = function(fn, name) {
    Object.defineProperty(fn, "name", { value: name, configurable: true });
    return fn;
  };`;

  return `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    * { margin: 0; padding: 0; }
    body { width: ${size}px; height: ${size}px; overflow: hidden; }
    #canvas-container { width: ${size}px; height: ${size}px; }
    canvas { display: block; }
  </style>
</head>
<body>
  <div id="canvas-container"></div>
  <script>${esbuildPolyfill}</script>
  <script>${p5Source}</script>
  <script>${helperScript}</script>
</body>
</html>`;
}
