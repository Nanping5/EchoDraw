// applyIntent: 把后端返回的 Intent 翻译为 store 操作。
//
// 关键: 增量画/重画/修改/清空 四种 Action 的处理路径
//   - delta + create: addShape (push 到末尾)
//   - delta + scene: addShapes (批量)
//   - modify + update: updateShape / updateShapes / moveTo / setScale / setRotation
//   - clear + clear: clearAll
//   - select: select(ids)
//   - undo/redo: undo() / redo()
//   - export: 触发 PNG 导出 (用 canvas.toDataURL)
//   - ask_back: 上层显示 question
//   - delete: deleteShapes

import { useCanvasStore } from "../store/canvasStore";
import type { Intent } from "../types";

export function applyIntent(intent: Intent): void {
  const store = useCanvasStore.getState();
  switch (intent.cmd) {
    case "create":
      if (intent.shape) store.addShape(intent.shape);
      return;
    case "scene":
      if (intent.scenes && intent.scenes.length > 0) {
        store.addShapes(intent.scenes);
      }
      return;
    case "update": {
      const ids = intent.target?.ids ?? store.selectedIds;
      if (ids.length === 0) return;
      if (intent.patch) {
        const patch: any = { style: { ...store.shapes.find(s => s.id === ids[0])?.style, ...intent.patch } };
        // 如果 patch 里有 fill/stroke 等, 写到 style
        if (intent.patch.fill) patch.style.fill = intent.patch.fill;
        if (intent.patch.stroke) patch.style.stroke = intent.patch.stroke;
        if (intent.patch.strokeWidth) patch.style.strokeWidth = intent.patch.strokeWidth;
        if (intent.patch.opacity != null) patch.style.opacity = intent.patch.opacity;
        store.updateShapes(ids, patch);
      }
      if (intent.moveTo) {
        store.moveTo(ids, intent.moveTo.x, intent.moveTo.y);
      }
      if (intent.scale != null) {
        store.setScale(ids, intent.scale);
      }
      if (intent.rotation != null) {
        store.setRotation(ids, intent.rotation);
      }
      return;
    }
    case "delete": {
      const ids = intent.target?.ids ?? [];
      if (ids.length > 0) store.deleteShapes(ids);
      return;
    }
    case "select": {
      const ids = intent.target?.ids ?? [];
      store.select(ids);
      return;
    }
    case "undo":
      store.undo();
      return;
    case "redo":
      store.redo();
      return;
    case "clear":
      store.clearAll();
      return;
    case "export":
      // 由 UI 层监听 export intent 触发, 此处 no-op
      return;
    case "ask_back":
    case "unknown":
    default:
      return;
  }
}
