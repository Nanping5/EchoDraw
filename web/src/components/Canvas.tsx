// 主画布组件。
// - 监听 store 变化, 自动重新渲染所有 ShapeNode
// - 点击空白处取消选中
// - 键盘 ESC 取消选中
import { useEffect, useRef } from "react";
import { Stage, Layer, Transformer } from "react-konva";
import { useCanvasStore } from "../store/canvasStore";
import { ShapeNode } from "./ShapeNode";
import type { Shape } from "../types";

interface Props {
  width: number;
  height: number;
}

export function Canvas({ width, height }: Props) {
  const shapes = useCanvasStore((s) => s.shapes);
  const selectedIds = useCanvasStore((s) => s.selectedIds);
  const select = useCanvasStore((s) => s.select);
  const updateShape = useCanvasStore((s) => s.updateShape);

  const stageRef = useRef<any>(null);
  const trRef = useRef<any>(null);

  // 选中节点时, 给 Transformer 挂上节点
  useEffect(() => {
    if (!trRef.current) return;
    const stage = stageRef.current;
    if (!stage) return;
    const nodes = selectedIds
      .map((id) => stage.findOne(`#${id}`))
      .filter(Boolean);
    trRef.current.nodes(nodes);
    trRef.current.getLayer()?.batchDraw();
  }, [selectedIds, shapes]);

  // ESC 取消选中
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") select([]);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [select]);

  return (
    <Stage
      ref={stageRef}
      width={width}
      height={height}
      onMouseDown={(e) => {
        // 点击空白处取消选中
        if (e.target === e.target.getStage()) {
          select([]);
        }
      }}
    >
      <Layer>
        {shapes.map((s) => (
          <ShapeNode
            key={s.id}
            shape={s}
            selected={selectedIds.includes(s.id)}
            onSelect={() => select([s.id])}
            onChange={(patch) => updateShape(s.id, patch)}
          />
        ))}
        <Transformer
          ref={trRef}
          rotateEnabled={true}
          borderStroke="#1976d2"
          anchorStroke="#1976d2"
          anchorFill="#ffffff"
        />
      </Layer>
    </Stage>
  );
}
