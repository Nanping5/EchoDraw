// 前后端共享的 Shape 类型定义 (与 server/internal/model/types.go 字段一致)
// 后续可考虑用 OpenAPI 自动生成, MVP 阶段手动同步。

export type ShapeType =
  | "circle"
  | "rect"
  | "line"
  | "ellipse"
  | "triangle"
  | "star"
  | "text"
  | "arrow";

export interface ShapeStyle {
  fill?: string;
  stroke?: string;
  strokeWidth?: number;
  fontSize?: number;
  opacity?: number;
  rotation?: number;
}

export interface Shape {
  id: string;
  type: ShapeType;
  x: number;
  y: number;
  width?: number;
  height?: number;
  radius?: number;
  points?: number[]; // line/arrow 相对中心偏移
  text?: string;
  style: ShapeStyle;
}

export type CommandType =
  | "create"
  | "update"
  | "delete"
  | "select"
  | "undo"
  | "redo"
  | "clear"
  | "export"
  | "ask_back"
  | "scene"
  | "unknown";

export type Action = "delta" | "redraw" | "modify" | "clear";

export interface Selection {
  ref?: string;
  filter?: string;
  ids?: string[];
}

export interface Intent {
  cmd: CommandType;
  action?: Action;
  shape?: Shape;
  scenes?: Shape[];
  target?: Selection;
  patch?: ShapeStyle;
  moveTo?: { x: number; y: number };
  scale?: number;
  rotation?: number;
  question?: string;
  reply?: string;
  raw?: string;
}

export interface UnderstandRequest {
  text: string;
  context?: Shape[];
}

export interface UnderstandResponse {
  intents: Intent[];
  reply?: string;
}

// 画布快照 (用于撤销/重做)
export interface CanvasSnapshot {
  shapes: Shape[];
  selectedIds: string[];
}
