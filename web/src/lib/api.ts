// 后端 API 客户端
// 全部走相对路径 /api/*, 由 Vite dev server 代理到 :8080

import type { UnderstandRequest, UnderstandResponse } from "../types";

export async function understand(req: UnderstandRequest): Promise<UnderstandResponse> {
  const res = await fetch("/api/understand", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

export async function checkHealth(): Promise<{ ok: boolean; llm_enabled: boolean }> {
  const res = await fetch("/health");
  return res.json();
}
