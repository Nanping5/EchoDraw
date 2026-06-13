// 单个图元的 Konva 节点。
// 根据 shape.type 渲染不同的 Konva 组件。
// 选中时显示 Transformer (可拖动/缩放/旋转)。
import { Circle, Ellipse, Rect, Line, Star, Text, Arrow, Group } from "react-konva";
import type { Shape } from "../types";

interface Props {
  shape: Shape;
  selected: boolean;
  onSelect: () => void;
  onChange: (patch: Partial<Shape>) => void;
}

export function ShapeNode({ shape, selected, onSelect, onChange }: Props) {
  const common = {
    id: shape.id,
    x: shape.x,
    y: shape.y,
    rotation: shape.style?.rotation ?? 0,
    draggable: true,
    onClick: onSelect,
    onTap: onSelect,
    onDragEnd: (e: any) => onChange({ x: e.target.x(), y: e.target.y() }),
    fill: shape.style?.fill,
    stroke: selected ? "#1976d2" : shape.style?.stroke,
    strokeWidth: selected ? 3 : shape.style?.strokeWidth ?? 1,
    opacity: shape.style?.opacity ?? 1,
  };

  switch (shape.type) {
    case "circle":
      return (
        <Circle
          {...common}
          radius={shape.radius ?? 30}
        />
      );
    case "ellipse":
      return (
        <Ellipse
          {...common}
          radiusX={(shape.width ?? 60) / 2}
          radiusY={(shape.height ?? 40) / 2}
        />
      );
    case "rect":
      return (
        <Rect
          {...common}
          width={shape.width ?? 100}
          height={shape.height ?? 100}
          offsetX={(shape.width ?? 100) / 2}
          offsetY={(shape.height ?? 100) / 2}
        />
      );
    case "triangle":
      return (
        <Line
          {...common}
          points={trianglePoints(shape.width ?? 80, shape.height ?? 80)}
          closed
        />
      );
    case "star":
      return (
        <Star
          {...common}
          numPoints={5}
          innerRadius={(shape.width ?? 60) / 4}
          outerRadius={(shape.width ?? 60) / 2}
        />
      );
    case "line":
      return (
        <Line
          {...common}
          points={shape.points ?? [0, 0, 100, 0]}
          stroke={shape.style?.stroke ?? "#37474f"}
          fill={undefined}
        />
      );
    case "arrow":
      return (
        <Arrow
          {...common}
          points={shape.points ?? [0, 0, 100, 0]}
          stroke={shape.style?.stroke ?? "#37474f"}
          fill={shape.style?.stroke ?? "#37474f"}
        />
      );
    case "text":
      return (
        <Text
          {...common}
          text={shape.text ?? "文本"}
          fontSize={shape.style?.fontSize ?? 24}
          fill={shape.style?.fill ?? "#212121"}
          stroke={undefined}
          width={shape.width ?? 200}
          align="center"
          verticalAlign="middle"
          offsetX={(shape.width ?? 200) / 2}
          offsetY={(shape.style?.fontSize ?? 24) / 2}
        />
      );
    default:
      return null;
  }
}

function trianglePoints(w: number, h: number): number[] {
  // 等边三角形, 顶点在上方
  return [0, -h / 2, w / 2, h / 2, -w / 2, h / 2];
}
