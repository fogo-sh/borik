import express from "express";
import { renderAnimation, shutdownBrowser } from "./renderer.js";
import { listAnimations, getAnimation } from "./animations/index.js";

const app = express();
const PORT = parseInt(process.env.PORT ?? "3001", 10);

app.use(express.json({ limit: "50mb" }));

app.get("/health", (_req, res) => {
  res.json({ status: "ok" });
});

app.get("/animations", (_req, res) => {
  const animations = listAnimations().map((a) => ({
    name: a.name,
    description: a.description,
    params: a.params ?? {},
  }));
  res.json({ animations });
});

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

// Graceful shutdown
process.on("SIGTERM", async () => {
  console.log("Shutting down...");
  await shutdownBrowser();
  process.exit(0);
});

process.on("SIGINT", async () => {
  console.log("Shutting down...");
  await shutdownBrowser();
  process.exit(0);
});

app.listen(PORT, () => {
  console.log(`node-renderer listening on port ${PORT}`);
  console.log(`Animations available: ${listAnimations().map((a) => a.name).join(", ")}`);
});
