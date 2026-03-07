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
