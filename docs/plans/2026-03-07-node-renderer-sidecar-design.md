# Node Renderer Sidecar Design

## Problem

Borik currently uses ImageMagick for all image processing, which limits it to 2D operations. Effects like rotating spheres, pyramids, and other 3D animations (as seen on 3dgifmaker.com) require WebGL/3D rendering that ImageMagick cannot do.

## Solution

A Node.js sidecar service (`node-renderer/`) that runs p5.js animations headlessly and exposes them over HTTP. The Go bot calls this service the same way it calls other external APIs.

## Why a sidecar

3dgifmaker.com implements its effects as pure p5.js WEBGL animation functions. A Node sidecar lets us 1:1 port these functions with minimal translation. The architecture is renderer-agnostic -- p5.js is the first backend, but Three.js, raw Canvas 2D, or other renderers can be added behind the same API contract.

## Architecture

```
borik (Go)                     node-renderer (Node.js)
   |                                |
   |  POST /render                  |
   |  {animation, image, params} -->|
   |                                | headless p5.js WEBGL
   |                                | renders N frames
   |  <-- {frames[], delay}         |
   |                                |
   | assembles GIF via ImageMagick  |
   | uploads to Discord             |
```

### Directory structure

```
borik/
  node-renderer/
    package.json
    tsconfig.json
    Dockerfile
    src/
      server.ts              # HTTP API
      renderer.ts            # Headless p5.js frame rendering engine
      animations/
        index.ts             # Animation registry
        types.ts             # Shared types
        rotating-sphere.ts
        inside-sphere.ts
        low-poly-sphere.ts
        pyramid.ts
        360-spin.ts
  pkg/
    bot/
      renderer_client.go     # HTTP client for node-renderer
      sphere.go              # sphere command (and similar commands)
    config/
      config.go              # + NodeRendererUrl field
  compose.yml                # + renderer service
```

### API contract

**`POST /render`**

Request body (JSON):

```json
{
  "animation": "rotating-sphere",
  "image": "<base64-encoded image>",
  "params": {
    "totalFrames": 24,
    "size": 400,
    "bgColor": "#000000",
    "extraSliderNum": 0
  }
}
```

Response body (JSON):

```json
{
  "frames": ["<base64 PNG>", "<base64 PNG>", "..."],
  "delay": 80
}
```

**`GET /animations`**

Returns the list of available animations and their parameter schemas, for discoverability.

### Headless rendering

Key dependencies:
- `gl` (npm) -- provides headless WebGL via native OpenGL/OSMesa
- `p5` (npm) -- p5.js in instance mode
- A lightweight HTTP framework (Hono, Fastify, or Express)

Each render request:
1. Creates a headless p5.js instance in WEBGL mode
2. Loads the input image as a p5.Image texture
3. For each frame 0..totalFrames-1, calls the animation function and extracts canvas pixels
4. Encodes each frame as PNG
5. Returns the frame array

### Animation functions

Direct ports from 3dgifmaker.com's minified source. Example:

```ts
// rotating-sphere.ts
export const rotatingSphere: AnimationDefinition = {
  name: "rotating-sphere",
  render({ p5, mainImage, size, currentFrame, totalFrames, bgColor, params }) {
    p5.background(bgColor);
    p5.rotateY(Math.PI + currentFrame * (2 * Math.PI / totalFrames));
    const wobble = (params.wobbliness ?? 0) / 20;
    p5.rotateZ(Math.sin(currentFrame * (2 * Math.PI / totalFrames)) * wobble);
    p5.rotateX(Math.sin(currentFrame * (2 * Math.PI / totalFrames)) * wobble);
    p5.texture(mainImage);
    p5.sphere(size / 3);
  },
  params: {
    wobbliness: { type: "number", min: 0, max: 100, default: 0 }
  }
};
```

### Go integration

**Config** -- new env var `BORIK_NODE_RENDERER_URL` (default: `http://localhost:3001`).

**Renderer client** -- `pkg/bot/renderer_client.go` provides a `RenderAnimation(animation string, imageBytes []byte, params map[string]any) ([][]byte, int, error)` function that POSTs to the sidecar and returns frame bytes + delay.

**Bot commands** -- renderer-based commands don't fit the existing `ImageOperation` signature (which processes input frames one-by-one). Instead, they follow the pattern from `image_gen.go`: a custom handler that takes a single input image, calls the renderer, gets all output frames back, assembles them into a GIF via ImageMagick, and uploads to Discord.

### Deployment

`compose.yml` adds a `renderer` service:

```yaml
services:
  borik:
    # ... existing
    environment:
      - BORIK_NODE_RENDERER_URL=http://renderer:3001
  renderer:
    build: ./node-renderer
    restart: always
```

The renderer Dockerfile installs Node.js + native GL dependencies (mesa-utils, libgl1-mesa-dev, xvfb or OSMesa for headless).

## Initial scope

Five animations to validate the architecture:

| Animation | Description | Source |
|-----------|-------------|--------|
| rotating-sphere | Image textured on a rotating sphere | 3dgifmaker "Rotating Sphere" |
| inside-sphere | Camera inside a large textured sphere | 3dgifmaker "Inside Sphere" |
| low-poly-sphere | Faceted sphere with configurable detail | 3dgifmaker "Low Poly Sphere" |
| pyramid | Image on a rotating pyramid | 3dgifmaker "Pyramid" |
| 360-spin | Flat image rotating in 3D space | 3dgifmaker "360 Spin" |

## Future extensions

- Additional 3dgifmaker animation ports (there are 70+)
- Three.js-based renderers for effects that benefit from it
- Raw Canvas 2D renderers for simpler 2D animations
- User-defined animations (like 3dgifmaker's "User Animation" mode)
