import { Canvas } from "./components/Canvas";

const CANVAS_W = 1200;
const CANVAS_H = 800;

export default function App() {
  return (
    <div className="flex h-full flex-col bg-slate-50">
      <header className="flex items-center justify-between border-b bg-white px-6 py-3 shadow-sm">
        <h1 className="text-xl font-bold text-slate-900">EchoDraw · 绘声</h1>
        <div className="text-sm text-slate-500">画布 {CANVAS_W} × {CANVAS_H}</div>
      </header>
      <main className="flex flex-1 items-center justify-center overflow-auto p-6">
        <div className="rounded-lg bg-white shadow-lg" style={{ width: CANVAS_W, height: CANVAS_H }}>
          <Canvas width={CANVAS_W} height={CANVAS_H} />
        </div>
      </main>
    </div>
  );
}
