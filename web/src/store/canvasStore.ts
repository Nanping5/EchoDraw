// 画布状态管理 (Zustand)
//
// 状态结构 (PRD 附录 A.2):
//   shapes: Shape[]                // 顺序即 z-order, 末尾 = 最上层
//   selectedIds: string[]          // 当前选中
//   history: { past: Snapshot[], future: Snapshot[] }  // 撤销/重做
//
// 规则:
//   - shapes 变更前 push 当前快照到 past (除 undo/redo/select 自身外)
//   - undo(): pop past → current, push current → future
//   - redo(): pop future → current, push current → past
//   - 上限 50 个快照 (内存保护)
import { create } from "zustand";
import type { CanvasSnapshot, Shape } from "../types";

const HISTORY_LIMIT = 50;

interface CanvasState {
  shapes: Shape[];
  selectedIds: string[];

  // history 是私有的, 暴露 actions 而非内部结构
  _past: CanvasSnapshot[];
  _future: CanvasSnapshot[];

  // 读
  canUndo: () => boolean;
  canRedo: () => boolean;

  // 写 (变更前自动 push 到 past, 除 _noHistory 外)
  addShape: (s: Shape) => void;
  addShapes: (ss: Shape[]) => void; // 场景生成批量加
  updateShape: (id: string, patch: Partial<Shape>) => void;
  updateShapes: (ids: string[], patch: Partial<Shape>) => void; // 批量改 (如全选变色)
  deleteShapes: (ids: string[]) => void;
  clearAll: () => void;
  select: (ids: string[]) => void;
  moveTo: (ids: string[], x: number, y: number) => void;
  setScale: (ids: string[], factor: number) => void;
  setRotation: (ids: string[], deg: number) => void;

  undo: () => void;
  redo: () => void;
  // 强制覆盖 (不压栈, 用于 import 恢复)
  _setShapes: (shapes: Shape[]) => void;
}

function snapshot(s: CanvasState): CanvasSnapshot {
  return { shapes: s.shapes, selectedIds: s.selectedIds };
}

function pushPast(s: CanvasState): CanvasSnapshot[] {
  const next = [...s._past, snapshot(s)];
  if (next.length > HISTORY_LIMIT) next.shift();
  return next;
}

export const useCanvasStore = create<CanvasState>((set, get) => ({
  shapes: [],
  selectedIds: [],
  _past: [],
  _future: [],

  canUndo: () => get()._past.length > 0,
  canRedo: () => get()._future.length > 0,

  addShape: (s) =>
    set((st) => ({
      _past: pushPast(st),
      _future: [],
      shapes: [...st.shapes, s],
    })),

  addShapes: (ss) =>
    set((st) => ({
      _past: pushPast(st),
      _future: [],
      shapes: [...st.shapes, ...ss],
    })),

  updateShape: (id, patch) =>
    set((st) => ({
      _past: pushPast(st),
      _future: [],
      shapes: st.shapes.map((s) => (s.id === id ? { ...s, ...patch } : s)),
    })),

  updateShapes: (ids, patch) =>
    set((st) => {
      const idSet = new Set(ids);
      return {
        _past: pushPast(st),
        _future: [],
        shapes: st.shapes.map((s) =>
          idSet.has(s.id) ? { ...s, ...patch } : s
        ),
      };
    }),

  deleteShapes: (ids) =>
    set((st) => {
      const idSet = new Set(ids);
      return {
        _past: pushPast(st),
        _future: [],
        shapes: st.shapes.filter((s) => !idSet.has(s.id)),
        selectedIds: st.selectedIds.filter((id) => !idSet.has(id)),
      };
    }),

  clearAll: () =>
    set((st) => ({
      _past: pushPast(st),
      _future: [],
      shapes: [],
      selectedIds: [],
    })),

  select: (ids) => set({ selectedIds: ids }),

  moveTo: (ids, x, y) =>
    set((st) => {
      const idSet = new Set(ids);
      return {
        _past: pushPast(st),
        _future: [],
        shapes: st.shapes.map((s) => (idSet.has(s.id) ? { ...s, x, y } : s)),
      };
    }),

  setScale: (ids, factor) =>
    set((st) => {
      const idSet = new Set(ids);
      return {
        _past: pushPast(st),
        _future: [],
        shapes: st.shapes.map((s) => {
          if (!idSet.has(s.id)) return s;
          if (s.type === "circle" && s.radius) {
            return { ...s, radius: s.radius * factor };
          }
          if ((s.type === "rect" || s.type === "ellipse" || s.type === "triangle" || s.type === "star") && s.width && s.height) {
            return { ...s, width: s.width * factor, height: s.height * factor };
          }
          return s;
        }),
      };
    }),

  setRotation: (ids, deg) =>
    set((st) => {
      const idSet = new Set(ids);
      return {
        _past: pushPast(st),
        _future: [],
        shapes: st.shapes.map((s) => {
          if (!idSet.has(s.id)) return s;
          const cur = s.style?.rotation ?? 0;
          return {
            ...s,
            style: { ...s.style, rotation: (cur + deg) % 360 },
          };
        }),
      };
    }),

  undo: () =>
    set((st) => {
      if (st._past.length === 0) return st;
      const prev = st._past[st._past.length - 1];
      return {
        _past: st._past.slice(0, -1),
        _future: [snapshot(st), ...st._future],
        shapes: prev.shapes,
        selectedIds: prev.selectedIds,
      };
    }),

  redo: () =>
    set((st) => {
      if (st._future.length === 0) return st;
      const next = st._future[0];
      return {
        _past: [...st._past, snapshot(st)],
        _future: st._future.slice(1),
        shapes: next.shapes,
        selectedIds: next.selectedIds,
      };
    }),

  _setShapes: (shapes) => set({ shapes, selectedIds: [], _past: [], _future: [] }),
}));
