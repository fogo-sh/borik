# Node Renderer Sidecar Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a Node.js sidecar service that renders p5.js 3D animations headlessly, exposed over HTTP, and integrate it into the Go bot with 5 initial animation commands.

**Architecture:** A `node-renderer/` directory at repo root containing a TypeScript HTTP API. It uses `gl` (headless WebGL) and p5.js in instance mode to render animation frames. The Go bot calls it via HTTP, receives PNG frames, assembles GIFs via ImageMagick, and uploads to Discord.

**Tech Stack:** TypeScript, Node.js, p5.js, `gl` (headless WebGL), Express/Hono, Go, ImageMagick (imagick.v3)

**Design doc:** `docs/plans/2026-03-07-node-renderer-sidecar-design.md`

---

### Task 1: Scaffold the node-renderer project

**Files:**
- Create: `node-renderer/package.json`
- Create: `node-renderer/tsconfig.json`
- Create: `node-renderer/.gitignore`

**Step 1: Create package.json**

```json
{
  "name": "borik-node-renderer",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "tsx watch src/server.ts",
    "build": "tsc",
    "start": "node dist/server.js"
  },
  "dependencies": {
    "express": "^5.0.0",
    "gl": "^6.0.2",
    "p5": "^1.11.0",
    "sharp": "^0.33.0"
  },
  "devDependencies": {
    "@types/express": "^5.0.0",
    "@types/node": "^22.0.0",
    "tsx": "^4.0.0",
    "typescript": "^5.7.0"
  }
}
```

Note: We use `sharp` for fast PNG encoding of canvas pixels. The `gl` package provides headless WebGL. Exact version numbers should be resolved at install time -- these are minimums.

**Step 2: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "outDir": "dist",
    "rootDir": "src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "dist"]
}
```

**Step 3: Create .gitignore**

```
node_modules/
dist/
```

**Step 4: Install dependencies**

Run from `node-renderer/`:
```bash
npm install
```

Expected: dependencies resolve, `node_modules/` created, `package-lock.json` generated.

Note: `gl` requires native build tools. On macOS this should work out of the box. On Linux (Docker), mesa/GL dev headers are needed -- handled in Task 7's Dockerfile.

**Step 5: Commit**

```bash
git add node-renderer/package.json node-renderer/tsconfig.json node-renderer/.gitignore node-renderer/package-lock.json
git commit -m "feat: scaffold node-renderer sidecar project"
```

---

### Task 2: Implement the animation type system and registry

**Files:**
- Create: `node-renderer/src/animations/types.ts`
- Create: `node-renderer/src/animations/index.ts`

**Step 1: Create the animation types**

```ts
// node-renderer/src/animations/types.ts
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
```

**Step 2: Create the animation registry**

```ts
// node-renderer/src/animations/index.ts
import { AnimationDefinition } from "./types.js";

const registry = new Map<string, AnimationDefinition>();

export function registerAnimation(animation: AnimationDefinition): void {
  registry.set(animation.name, animation);
}

export function getAnimation(name: string): AnimationDefinition | undefined {
  return registry.get(name);
}

export function listAnimations(): AnimationDefinition[] {
  return Array.from(registry.values());
}

// Import and register all animations (added in Task 3)
```

**Step 3: Commit**

```bash
git add node-renderer/src/animations/
git commit -m "feat: add animation type system and registry"
```

---

### Task 3: Port the 5 animations from 3dgifmaker

**Files:**
- Create: `node-renderer/src/animations/rotating-sphere.ts`
- Create: `node-renderer/src/animations/inside-sphere.ts`
- Create: `node-renderer/src/animations/low-poly-sphere.ts`
- Create: `node-renderer/src/animations/pyramid.ts`
- Create: `node-renderer/src/animations/spin-360.ts`
- Create: `node-renderer/src/animations/utils.ts`
- Modify: `node-renderer/src/animations/index.ts`

**Step 1: Create the shared utility**

The `getScaleFactor` helper is used by 360 Spin and potentially other animations:

```ts
// node-renderer/src/animations/utils.ts
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
```

**Step 2: Create rotating-sphere.ts**

```ts
// node-renderer/src/animations/rotating-sphere.ts
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
```

**Step 3: Create inside-sphere.ts**

```ts
// node-renderer/src/animations/inside-sphere.ts
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
```

**Step 4: Create low-poly-sphere.ts**

```ts
// node-renderer/src/animations/low-poly-sphere.ts
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
```

Note: p5.js TypeScript types may not expose the `detailX`/`detailY` overload of `sphere()`, hence the `as any` cast.

**Step 5: Create pyramid.ts**

```ts
// node-renderer/src/animations/pyramid.ts
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

    // All 4 faces use the same image (extraSideImages not supported yet)
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
```

Note: Extensive `as any` casts because p5.js WEBGL vertex UV overloads are poorly typed. This is standard for headless p5 TypeScript usage.

**Step 6: Create spin-360.ts**

```ts
// node-renderer/src/animations/spin-360.ts
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
```

**Step 7: Update the registry to import all animations**

```ts
// node-renderer/src/animations/index.ts
import { AnimationDefinition } from "./types.js";
import { rotatingSphere } from "./rotating-sphere.js";
import { insideSphere } from "./inside-sphere.js";
import { lowPolySphere } from "./low-poly-sphere.js";
import { pyramid } from "./pyramid.js";
import { spin360 } from "./spin-360.js";

const registry = new Map<string, AnimationDefinition>();

export function registerAnimation(animation: AnimationDefinition): void {
  registry.set(animation.name, animation);
}

export function getAnimation(name: string): AnimationDefinition | undefined {
  return registry.get(name);
}

export function listAnimations(): AnimationDefinition[] {
  return Array.from(registry.values());
}

// Register all built-in animations
[rotatingSphere, insideSphere, lowPolySphere, pyramid, spin360].forEach(registerAnimation);
```

**Step 8: Commit**

```bash
git add node-renderer/src/animations/
git commit -m "feat: port 5 animations from 3dgifmaker (sphere, pyramid, 360-spin)"
```

---

### Task 4: Implement the headless p5.js renderer

This is the core engine that runs p5.js in headless WEBGL mode and extracts frames.

**Files:**
- Create: `node-renderer/src/renderer.ts`

**Step 1: Research headless p5.js + gl integration**

Before writing code, verify that p5.js works with the `gl` npm package in headless mode. Key things to test:
- `createCanvas(w, h, WEBGL)` with headless gl context
- `texture()` and `sphere()` work in the headless context
- Pixel extraction via `loadPixels()` / canvas buffer

p5.js uses `HTMLCanvasElement.getContext('webgl')` internally. The `gl` npm package provides a WebGL context without a display. The integration path is:
1. Create a `gl` context (headless)
2. Create a fake canvas object that returns this gl context
3. Pass it to p5.js

Alternatively, use `jsdom` + `gl` to provide a DOM-like environment where p5.js can create its own canvas.

This task may require experimentation. The implementation should:
- Create a p5 instance in instance mode with WEBGL
- Load the source image as a p5.Image
- Call the animation render function for each frame
- Extract RGBA pixels from the WebGL framebuffer
- Encode each frame as PNG using sharp

**Step 2: Implement renderer.ts**

```ts
// node-renderer/src/renderer.ts
import createGL from "gl";
import p5 from "p5";
import sharp from "sharp";
import { getAnimation } from "./animations/index.js";

// Monkey-patch global to provide headless GL
// p5.js expects a browser-like environment; we need to shim it.
// This approach may need adjustment based on actual p5.js headless behavior.

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

export async function renderAnimation(req: RenderRequest): Promise<RenderResult> {
  const animation = getAnimation(req.animation);
  if (!animation) {
    throw new Error(`Unknown animation: ${req.animation}`);
  }

  const totalFrames = req.params.totalFrames ?? 24;
  const size = req.params.size ?? 400;
  const bgColor = req.params.bgColor ?? "#000000";
  const delay = Math.round(1000 / 15); // ~15fps, ~67ms per frame

  const frames: Buffer[] = [];

  // Create headless GL context
  const glContext = createGL(size, size, { preserveDrawingBuffer: true });

  // Create a p5 instance in instance mode
  // The exact integration with headless GL will require shimming.
  // p5.js instance mode: new p5((p) => { p.setup = ...; p.draw = ...; })
  //
  // For headless rendering, we may need to:
  // 1. Use jsdom to provide document/window
  // 2. Override canvas.getContext to return our headless GL context
  // 3. Run p5 in instance mode
  //
  // This is the most experimental part. The implementation should be
  // validated by actually running it and checking output frames.

  // Load the source image into a p5.Image
  // p5.loadImage() expects a URL or path; for a buffer we may need to
  // convert to a data URL or use createImage() + set pixels manually.

  // For each frame:
  for (let frame = 0; frame < totalFrames; frame++) {
    // Call animation.render() with the frame context
    // Read pixels from GL framebuffer
    const pixels = new Uint8Array(size * size * 4);
    glContext.readPixels(0, 0, size, size, glContext.RGBA, glContext.UNSIGNED_BYTE, pixels);

    // WebGL reads bottom-to-top; flip vertically
    const flipped = flipVertically(pixels, size, size);

    // Encode to PNG
    const png = await sharp(Buffer.from(flipped), {
      raw: { width: size, height: size, channels: 4 },
    }).png().toBuffer();

    frames.push(png);
  }

  glContext.getExtension("STACKGL_destroy_context")?.destroy();

  return { frames, delay };
}

function flipVertically(pixels: Uint8Array, width: number, height: number): Uint8Array {
  const rowSize = width * 4;
  const result = new Uint8Array(pixels.length);
  for (let y = 0; y < height; y++) {
    const srcOffset = y * rowSize;
    const dstOffset = (height - 1 - y) * rowSize;
    result.set(pixels.subarray(srcOffset, srcOffset + rowSize), dstOffset);
  }
  return result;
}
```

**Important note:** The exact p5.js headless integration is the riskiest part of this project. The code above is a scaffold -- the implementer should:
1. Try running it and see what p5.js needs (jsdom? canvas shim? other patches?)
2. Look at existing headless p5.js solutions (e.g., `p5-node`, `p5.js-node` on npm, or GitHub issues on the p5.js repo about headless/server-side rendering)
3. Adjust the approach based on what works

If p5.js proves too difficult to run headlessly, fallback options:
- Use `node-canvas` + `headless-gl` with raw WebGL calls (no p5.js wrapper)
- Use Three.js which has better documented headless rendering support

**Step 3: Commit**

```bash
git add node-renderer/src/renderer.ts
git commit -m "feat: implement headless p5.js renderer engine (initial scaffold)"
```

---

### Task 5: Implement the HTTP API server

**Files:**
- Create: `node-renderer/src/server.ts`

**Step 1: Create the Express server**

```ts
// node-renderer/src/server.ts
import express from "express";
import { renderAnimation } from "./renderer.js";
import { listAnimations, getAnimation } from "./animations/index.js";

const app = express();
const PORT = parseInt(process.env.PORT ?? "3001", 10);

// Accept large payloads (images can be big)
app.use(express.json({ limit: "50mb" }));

// Health check
app.get("/health", (_req, res) => {
  res.json({ status: "ok" });
});

// List available animations and their param schemas
app.get("/animations", (_req, res) => {
  const animations = listAnimations().map((a) => ({
    name: a.name,
    description: a.description,
    params: a.params ?? {},
  }));
  res.json({ animations });
});

// Render an animation
app.post("/render", async (req, res) => {
  try {
    const { animation, image, params } = req.body;

    if (!animation || typeof animation !== "string") {
      res.status(400).json({ error: "Missing or invalid 'animation' field" });
      return;
    }
    if (!image || typeof image !== "string") {
      res.status(400).json({ error: "Missing or invalid 'image' field (expected base64 string)" });
      return;
    }

    const animDef = getAnimation(animation);
    if (!animDef) {
      res.status(404).json({
        error: `Unknown animation: ${animation}`,
        available: listAnimations().map((a) => a.name),
      });
      return;
    }

    const imageBuffer = Buffer.from(image, "base64");
    const result = await renderAnimation({
      animation,
      image: imageBuffer,
      params: params ?? {},
    });

    res.json({
      frames: result.frames.map((f) => f.toString("base64")),
      delay: result.delay,
    });
  } catch (err: any) {
    console.error("Render error:", err);
    res.status(500).json({ error: err.message ?? "Internal server error" });
  }
});

app.listen(PORT, () => {
  console.log(`node-renderer listening on port ${PORT}`);
  console.log(`Animations available: ${listAnimations().map((a) => a.name).join(", ")}`);
});
```

**Step 2: Verify the server starts**

Run from `node-renderer/`:
```bash
npx tsx src/server.ts
```

Expected: Server starts, prints listening message and animation list. The `/health` and `/animations` endpoints should respond. `/render` may fail if the headless GL integration isn't wired up yet -- that's expected at this stage.

**Step 3: Commit**

```bash
git add node-renderer/src/server.ts
git commit -m "feat: add HTTP API server for node-renderer"
```

---

### Task 6: Go integration -- config, renderer client, and bot commands

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `.env.dist`
- Create: `pkg/bot/renderer_client.go`
- Create: `pkg/bot/node_renderer.go`
- Modify: `pkg/bot/bot.go`

**Step 1: Add NodeRendererUrl to config**

In `pkg/config/config.go`, add to the `Config` struct:

```go
NodeRendererUrl string `default:"" split_words:"true"`
```

This follows the existing pattern (e.g., `OpenaiBaseUrl`). The env var will be `BORIK_NODE_RENDERER_URL`.

**Step 2: Add to .env.dist**

Add:
```
BORIK_NODE_RENDERER_URL=http://localhost:3001
```

**Step 3: Create the renderer HTTP client**

```go
// pkg/bot/renderer_client.go
package bot

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/fogo-sh/borik/pkg/config"
)

type renderRequest struct {
	Animation string         `json:"animation"`
	Image     string         `json:"image"`
	Params    map[string]any `json:"params"`
}

type renderResponse struct {
	Frames []string `json:"frames"`
	Delay  int      `json:"delay"`
	Error  string   `json:"error,omitempty"`
}

// RenderAnimation calls the node-renderer sidecar to render an animation.
// Returns the rendered frames as PNG byte slices and the frame delay in ms.
func RenderAnimation(animation string, imageBytes []byte, params map[string]any) ([][]byte, int, error) {
	rendererURL := config.Instance.NodeRendererUrl
	if rendererURL == "" {
		return nil, 0, fmt.Errorf("node renderer URL not configured (set BORIK_NODE_RENDERER_URL)")
	}

	reqBody := renderRequest{
		Animation: animation,
		Image:     base64.StdEncoding.EncodeToString(imageBytes),
		Params:    params,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("error marshaling render request: %w", err)
	}

	resp, err := http.Post(rendererURL+"/render", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("error calling node renderer: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading renderer response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp renderResponse
		_ = json.Unmarshal(body, &errResp)
		return nil, 0, fmt.Errorf("renderer error (status %d): %s", resp.StatusCode, errResp.Error)
	}

	var result renderResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, fmt.Errorf("error parsing renderer response: %w", err)
	}

	frames := make([][]byte, len(result.Frames))
	for i, f := range result.Frames {
		decoded, err := base64.StdEncoding.DecodeString(f)
		if err != nil {
			return nil, 0, fmt.Errorf("error decoding frame %d: %w", i, err)
		}
		frames[i] = decoded
	}

	return frames, result.Delay, nil
}
```

**Step 4: Create the bot commands for renderer-based operations**

```go
// pkg/bot/node_renderer.go
package bot

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/fogo-sh/borik/pkg/config"
)

// RendererOperation defines an animation that is rendered by the node-renderer sidecar.
// Unlike ImageOperation which processes frames one-by-one, this sends a single image
// to the renderer and gets back all frames at once.
type RendererOperation struct {
	Animation string
	Params    func(args RendererArgs) map[string]any
}

type RendererArgs struct {
	ImageURL string `default:"" description:"URL to the image to process. Leave blank to automatically attempt to find an image."`
	Frames   uint   `default:"24" description:"Number of frames in the output GIF."`
	Size     uint   `default:"400" description:"Canvas size in pixels."`
}

func (args RendererArgs) GetImageURL() string {
	return args.ImageURL
}

func invokeRendererOperation(ctx *OperationContext, args RendererArgs, op RendererOperation) {
	defer TypingIndicatorForContext(ctx)()

	if err := ctx.DeferResponse(); err != nil {
		log.Error().Err(err).Msg("Failed to defer response")
		return
	}

	imageUrl := args.GetImageURL()
	if imageUrl == "" {
		var err error
		imageUrl, err = ctx.FindImageURL()
		if err != nil {
			log.Error().Err(err).Msg("Error while attempting to find image to process")
			return
		}
	}

	srcBytes, err := DownloadImage(imageUrl)
	if err != nil {
		log.Error().Err(err).Msg("Failed to download image to process")
		return
	}

	// Build params from args and the operation-specific param builder
	params := map[string]any{
		"totalFrames": args.Frames,
		"size":        args.Size,
	}
	if op.Params != nil {
		for k, v := range op.Params(args) {
			params[k] = v
		}
	}

	// Call the node renderer
	frames, delay, err := RenderAnimation(op.Animation, srcBytes, params)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render animation")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to render animation: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error response")
		}
		return
	}

	// Assemble frames into a GIF via ImageMagick
	resultImage := imagick.NewMagickWand()
	for i, frameBytes := range frames {
		frameWand := imagick.NewMagickWand()
		if err := frameWand.ReadImageBlob(frameBytes); err != nil {
			log.Error().Err(err).Int("frame", i).Msg("Failed to read rendered frame")
			return
		}
		if err := resultImage.AddImage(frameWand); err != nil {
			log.Error().Err(err).Int("frame", i).Msg("Failed to add frame to result")
			return
		}
	}

	resultImage.ResetIterator()

	if err := resultImage.SetImageFormat("GIF"); err != nil {
		log.Error().Err(err).Msg("Failed to set result format")
		return
	}

	if err := resultImage.SetImageDelay(uint(delay / 10)); err != nil {
		// ImageMagick delay is in centiseconds (1/100th of a second)
		log.Error().Err(err).Msg("Failed to set frame delay")
		return
	}

	if err := resultImage.ResetImagePage("0x0+0+0"); err != nil {
		log.Error().Err(err).Msg("Failed to repage image")
	}

	resultImage = resultImage.DeconstructImages()

	imageBlob, err := resultImage.GetImagesBlob()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get image blob")
		return
	}

	destBuffer := new(bytes.Buffer)
	if _, err := destBuffer.Write(imageBlob); err != nil {
		log.Error().Err(err).Msg("Failed to write image blob")
		return
	}

	resultFileName := fmt.Sprintf("%s.gif", strings.ReplaceAll(op.Animation, "-", "_"))
	if err := ctx.SendFiles([]*discordgo.File{{
		Name:   resultFileName,
		Reader: destBuffer,
	}}); err != nil {
		log.Error().Err(err).Msg("Failed to send rendered GIF")
		if sendErr := ctx.SendText(fmt.Sprintf("Failed to send resulting image: `%s`", err.Error())); sendErr != nil {
			log.Error().Err(sendErr).Msg("Failed to send error message")
		}
	}
}

// MakeRendererTextCommand creates a text command handler for a renderer operation
func MakeRendererTextCommand(op RendererOperation) func(*discordgo.MessageCreate, RendererArgs) {
	return func(message *discordgo.MessageCreate, args RendererArgs) {
		invokeRendererOperation(NewOperationContextFromMessage(Instance.session, message), args, op)
	}
}

// MakeRendererSlashCommand creates a slash command handler for a renderer operation
func MakeRendererSlashCommand(op RendererOperation) func(*discordgo.Session, *discordgo.InteractionCreate, RendererArgs) {
	return func(session *discordgo.Session, interaction *discordgo.InteractionCreate, args RendererArgs) {
		invokeRendererOperation(NewOperationContextFromInteraction(session, interaction), args, op)
	}
}

// Renderer operation definitions
var (
	SphereOp = RendererOperation{Animation: "rotating-sphere"}
	InsideSphereOp = RendererOperation{Animation: "inside-sphere"}
	LowPolySphereOp = RendererOperation{Animation: "low-poly-sphere"}
	PyramidOp = RendererOperation{Animation: "pyramid"}
	Spin360Op = RendererOperation{Animation: "360-spin"}
)
```

**Step 5: Register the commands in bot.go**

Add to the `commands` slice in `pkg/bot/bot.go`, guarded by `NodeRendererUrl` being set:

```go
{
    name:         "sphere",
    description:  "Map an image onto a rotating sphere.",
    textHandler:  MakeRendererTextCommand(SphereOp),
    slashHandler: MakeRendererSlashCommand(SphereOp),
    enabled:      func(c *configPkg.Config) bool { return c.NodeRendererUrl != "" },
},
{
    name:         "insidesphere",
    description:  "View an image from inside a sphere.",
    textHandler:  MakeRendererTextCommand(InsideSphereOp),
    slashHandler: MakeRendererSlashCommand(InsideSphereOp),
    enabled:      func(c *configPkg.Config) bool { return c.NodeRendererUrl != "" },
},
{
    name:         "lowpolysphere",
    description:  "Map an image onto a low-poly rotating sphere.",
    textHandler:  MakeRendererTextCommand(LowPolySphereOp),
    slashHandler: MakeRendererSlashCommand(LowPolySphereOp),
    enabled:      func(c *configPkg.Config) bool { return c.NodeRendererUrl != "" },
},
{
    name:         "pyramid",
    description:  "Map an image onto a rotating pyramid.",
    textHandler:  MakeRendererTextCommand(PyramidOp),
    slashHandler: MakeRendererSlashCommand(PyramidOp),
    enabled:      func(c *configPkg.Config) bool { return c.NodeRendererUrl != "" },
},
{
    name:         "spin",
    description:  "Spin an image 360 degrees in 3D space.",
    textHandler:  MakeRendererTextCommand(Spin360Op),
    slashHandler: MakeRendererSlashCommand(Spin360Op),
    enabled:      func(c *configPkg.Config) bool { return c.NodeRendererUrl != "" },
},
```

**Step 6: Verify Go code compiles**

Run from repo root:
```bash
go build ./...
```

Expected: compiles without errors.

**Step 7: Commit**

```bash
git add pkg/config/config.go pkg/bot/renderer_client.go pkg/bot/node_renderer.go pkg/bot/bot.go .env.dist
git commit -m "feat: add Go integration for node-renderer (client, commands, config)"
```

---

### Task 7: Docker and compose setup

**Files:**
- Create: `node-renderer/Dockerfile`
- Modify: `compose.yml`

**Step 1: Create the renderer Dockerfile**

```dockerfile
# node-renderer/Dockerfile
FROM node:22-bookworm

# Install native dependencies for headless GL
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      libgl1-mesa-dev \
      libxi-dev \
      xvfb \
      python3 \
      make \
      g++ && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY tsconfig.json ./
COPY src/ ./src/

RUN npm run build

# Use xvfb-run to provide a virtual display for headless GL
ENTRYPOINT ["xvfb-run", "-a", "node", "dist/server.js"]
```

Note: `xvfb-run` provides a virtual X display, which some GL implementations need even in "headless" mode. If the `gl` npm package works with OSMesa (pure software, no display), this can be simplified.

**Step 2: Update compose.yml**

```yaml
services:
  borik:
    image: ghcr.io/fogo-sh/borik
    restart: always
    build: .
    env_file:
      - .env
    environment:
      - BORIK_NODE_RENDERER_URL=http://renderer:3001
    depends_on:
      - renderer
  renderer:
    build: ./node-renderer
    restart: always
```

**Step 3: Verify Docker builds**

```bash
docker compose build renderer
```

Expected: image builds successfully.

**Step 4: Commit**

```bash
git add node-renderer/Dockerfile compose.yml
git commit -m "feat: add Docker setup for node-renderer sidecar"
```

---

### Task 8: End-to-end validation

**Step 1: Start the renderer locally**

```bash
cd node-renderer && npx tsx src/server.ts
```

**Step 2: Test the health endpoint**

```bash
curl http://localhost:3001/health
```

Expected: `{"status":"ok"}`

**Step 3: Test the animations list**

```bash
curl http://localhost:3001/animations
```

Expected: JSON with all 5 animations and their param schemas.

**Step 4: Test rendering with a sample image**

```bash
# Encode a test image to base64
BASE64_IMG=$(base64 < /path/to/test-image.png)

# Call the render endpoint
curl -X POST http://localhost:3001/render \
  -H "Content-Type: application/json" \
  -d "{\"animation\": \"rotating-sphere\", \"image\": \"$BASE64_IMG\", \"params\": {\"totalFrames\": 8, \"size\": 200}}" \
  | python3 -c "import sys,json; data=json.load(sys.stdin); print(f'Got {len(data[\"frames\"])} frames, delay={data[\"delay\"]}ms')"
```

Expected: `Got 8 frames, delay=67ms` (or similar). If the headless GL integration isn't working, debug and fix in Task 4.

**Step 5: Test the full pipeline with Docker Compose**

```bash
docker compose up --build
```

Verify both services start. Test by sending a Discord command to the bot (if configured) or by calling the renderer directly.

**Step 6: Commit any fixes from validation**

```bash
git add -A
git commit -m "fix: end-to-end validation fixes for node-renderer"
```

---

### Known risks and mitigations

1. **Headless p5.js may not work out of the box.** p5.js was designed for browsers. The `gl` npm package provides a WebGL context but p5 also expects DOM elements. Mitigation: Use `jsdom` to provide a minimal DOM, or look at `p5-node` / `p5.js-server` packages. Worst case, rewrite the render loop using raw WebGL calls instead of p5.js (the animation math is simple enough).

2. **Image loading into p5.Image.** p5's `loadImage()` is async and expects URLs. For buffer input, we may need to use `p5.createImage()` + manually set pixels, or convert to a data URL. This is solvable but needs experimentation.

3. **Performance.** Software GL rendering is slow. For 24 frames at 400x400, expect 2-10 seconds per render. This is acceptable for a Discord bot (users expect some delay). If it's too slow, consider reducing default frame count or size.
