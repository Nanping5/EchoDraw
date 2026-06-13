import { Canvas } from "./components/Canvas";
import { VoiceBar } from "./components/VoiceBar";
import { useCanvasStore } from "./store/canvasStore";
import { Undo2, Redo2, Trash2, Download } from "lucide-react";

const CANVAS_W = 1200;
const CANVAS_H = 800;

export default function App() {
  const undo = useCanvasStore((s) => s.undo);
  const redo = useCanvasStore((s) => s.redo);
  const clearAll = useCanvasStore((s) => s.clearAll);
  const canUndo = useCanvasStore((s) => s.canUndo());
  const canRedo = useCanvasStore((s) => s.canRedo());
  const shapes = useCanvasStore((s) => s.shapes);

  function exportPNG() {
    // 找 canvas 元素
    const canvas = document.querySelector("canvas") as HTMLCanvasElement | null;
    if (!canvas) return;
    const url = canvas.toDataURL("image/png");
    const a = document.createElement("a");
    a.href = url;
    a.download = `echodraw-${Date.now()}.png`;
    a.click();
  }

  return (
    <div className="flex h-full flex-col bg-slate-50">
      <header className="flex items-center justify-between border-b bg-white px-6 py-3 shadow-sm">
        <h1 className="text-xl font-bold text-slate-900">EchoDraw · 绘声</h1>
        <div className="flex items-center gap-2 text-sm text-slate-500">
          <span>画布 {CANVAS_W} × {CANVAS_H}</span>
          <span className="mx-2">·</span>
          <span>{shapes.length} 个图元</span>
        </div>
      </header>
      <main className="flex flex-1 gap-4 overflow-hidden p-4">
        <div className="flex flex-1 items-center justify-center overflow-auto">
          <div className="rounded-lg bg-white shadow-lg" style={{ width: CANVAS_W, height: CANVAS_H }}>
            <Canvas width={CANVAS_W} height={CANVAS_H} />
          </div>
        </div>
        <aside className="flex w-80 flex-col gap-3">
          <VoiceBar />
          <div className="rounded-md border bg-white p-3 shadow-sm">
            <div className="mb-2 text-sm font-medium text-slate-700">操作</div>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={undo}
                disabled={!canUndo}
                className="flex items-center justify-center gap-1 rounded border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50 disabled:opacity-40"
              >
                <Undo2 size={14} /> 撤销
              </button>
              <button
                onClick={redo}
                disabled={!canRedo}
                className="flex items-center justify-center gap-1 rounded border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50 disabled:opacity-40"
              >
                <Redo2 size={14} /> 重做
              </button>
              <button
                onClick={clearAll}
                className="flex items-center justify-center gap-1 rounded border border-slate-300 px-3 py-2 text-sm hover:bg-red-50 hover:text-red-600"
              >
                <Trash2 size={14} /> 清空
              </button>
              <button
                onClick={exportPNG}
                className="flex items-center justify-center gap-1 rounded border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50"
              >
                <Download size={14} /> 导出
              </button>
            </div>
          </div>
          <div className="rounded-md border bg-white p-3 text-xs text-slate-500 shadow-sm">
            <div className="mb-1 font-medium text-slate-700">试试这些指令</div>
            <ul className="space-y-1">
              <li>· 画一个红色的大圆</li>
              <li>· 画一个蓝色矩形在左边</li>
              <li>· 画一个夜空有月亮和星星</li>
              <li>· 把它改成黄色</li>
              <li>· 撤销 / 重做 / 清空</li>
            </ul>
          </div>
        </aside>
      </main>
    </div>
  );
}
