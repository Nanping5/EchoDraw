import { describe, it, expect, beforeEach } from "vitest";
import { useCanvasStore } from "./canvasStore";
import type { Shape } from "../types";

const circle = (id: string, x = 100, y = 100): Shape => ({
  id,
  type: "circle",
  x,
  y,
  radius: 50,
  style: { fill: "#e53935" },
});

beforeEach(() => {
  useCanvasStore.getState()._setShapes([]);
});

describe("canvasStore basic", () => {
  it("starts empty", () => {
    const s = useCanvasStore.getState();
    expect(s.shapes).toEqual([]);
    expect(s.selectedIds).toEqual([]);
    expect(s.canUndo()).toBe(false);
    expect(s.canRedo()).toBe(false);
  });

  it("addShape pushes to past", () => {
    const store = useCanvasStore.getState();
    store.addShape(circle("a"));
    expect(useCanvasStore.getState().shapes).toHaveLength(1);
    expect(useCanvasStore.getState().canUndo()).toBe(true);
  });

  it("addShapes batch", () => {
    useCanvasStore.getState().addShapes([circle("a"), circle("b"), circle("c")]);
    expect(useCanvasStore.getState().shapes).toHaveLength(3);
  });
});

describe("canvasStore mutations", () => {
  it("updateShape changes props and pushes history", () => {
    useCanvasStore.getState().addShape(circle("a"));
    useCanvasStore.getState().updateShape("a", { x: 999 });
    expect(useCanvasStore.getState().shapes[0].x).toBe(999);
  });

  it("updateShapes batches", () => {
    useCanvasStore.getState().addShapes([circle("a", 0, 0), circle("b", 0, 0)]);
    useCanvasStore.getState().updateShapes(["a", "b"], { y: 500 });
    const ss = useCanvasStore.getState().shapes;
    expect(ss.every((s) => s.y === 500)).toBe(true);
  });

  it("deleteShapes removes and clears selection", () => {
    useCanvasStore.getState().addShapes([circle("a"), circle("b")]);
    useCanvasStore.getState().select(["a", "b"]);
    useCanvasStore.getState().deleteShapes(["a"]);
    expect(useCanvasStore.getState().shapes).toHaveLength(1);
    expect(useCanvasStore.getState().selectedIds).toEqual(["b"]);
  });

  it("clearAll empties everything", () => {
    useCanvasStore.getState().addShapes([circle("a"), circle("b")]);
    useCanvasStore.getState().clearAll();
    expect(useCanvasStore.getState().shapes).toEqual([]);
  });
});

describe("canvasStore undo/redo", () => {
  it("undo restores previous state", () => {
    useCanvasStore.getState().addShape(circle("a", 100, 100));
    useCanvasStore.getState().addShape(circle("b", 200, 200));
    expect(useCanvasStore.getState().shapes).toHaveLength(2);
    useCanvasStore.getState().undo();
    expect(useCanvasStore.getState().shapes).toHaveLength(1);
    expect(useCanvasStore.getState().canRedo()).toBe(true);
  });

  it("redo restores later state", () => {
    useCanvasStore.getState().addShape(circle("a"));
    useCanvasStore.getState().addShape(circle("b"));
    useCanvasStore.getState().undo();
    useCanvasStore.getState().redo();
    expect(useCanvasStore.getState().shapes).toHaveLength(2);
  });

  it("new action clears future", () => {
    useCanvasStore.getState().addShape(circle("a"));
    useCanvasStore.getState().addShape(circle("b"));
    useCanvasStore.getState().undo(); // 回到 [a], future=[b]
    expect(useCanvasStore.getState().canRedo()).toBe(true);
    useCanvasStore.getState().addShape(circle("c")); // 新动作, future 清空
    expect(useCanvasStore.getState().canRedo()).toBe(false);
  });

  it("history limit 50", () => {
    for (let i = 0; i < 60; i++) {
      useCanvasStore.getState().addShape(circle(`s_${i}`));
    }
    expect(useCanvasStore.getState()._past.length).toBe(50);
  });
});

describe("canvasStore transforms", () => {
  it("moveTo updates x/y of selected", () => {
    useCanvasStore.getState().addShape(circle("a", 100, 100));
    useCanvasStore.getState().moveTo(["a"], 500, 600);
    expect(useCanvasStore.getState().shapes[0]).toMatchObject({ x: 500, y: 600 });
  });

  it("setScale multiplies radius for circle", () => {
    useCanvasStore.getState().addShape(circle("a"));
    useCanvasStore.getState().setScale(["a"], 2);
    expect(useCanvasStore.getState().shapes[0].radius).toBe(100);
  });

  it("setScale multiplies width/height for rect", () => {
    useCanvasStore.getState().addShapes([
      { id: "r", type: "rect", x: 0, y: 0, width: 100, height: 50, style: {} },
    ]);
    useCanvasStore.getState().setScale(["r"], 0.5);
    expect(useCanvasStore.getState().shapes[0]).toMatchObject({ width: 50, height: 25 });
  });

  it("setRotation accumulates", () => {
    useCanvasStore.getState().addShape(circle("a"));
    useCanvasStore.getState().setRotation(["a"], 45);
    useCanvasStore.getState().setRotation(["a"], 45);
    expect(useCanvasStore.getState().shapes[0].style.rotation).toBe(90);
  });
});
