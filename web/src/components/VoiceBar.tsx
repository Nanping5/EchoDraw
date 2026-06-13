// 麦克风按钮 + 状态显示 + 提示气泡
import { useState } from "react";
import { Mic, MicOff } from "lucide-react";
import { useSpeech } from "../hooks/useSpeech";
import { understand } from "../lib/api";
import { applyIntent } from "../lib/applyIntent";
import { useCanvasStore } from "../store/canvasStore";

export function VoiceBar() {
  const [busy, setBusy] = useState(false);
  const [hint, setHint] = useState<string>("点击麦克风开始说话");
  const [lastReply, setLastReply] = useState<string>("");

  const { supported, listening, transcript, start, stop, error } = useSpeech({
    onResult: (text, isFinal) => {
      setHint(isFinal ? `你说: ${text}` : `识别中: ${text}`);
      if (isFinal && text.trim()) {
        handleCommand(text.trim());
      }
    },
    onError: (e) => setHint(`识别错误: ${e}`),
  });

  async function handleCommand(text: string) {
    setBusy(true);
    try {
      const ctx = useCanvasStore.getState().shapes;
      const resp = await understand({ text, context: ctx });
      for (const it of resp.intents) {
        applyIntent(it);
        if (it.question) setLastReply(it.question);
        else if (it.reply) setLastReply(it.reply);
      }
      if (resp.reply) setLastReply(resp.reply);
    } catch (e: any) {
      setHint(`请求失败: ${e.message}`);
    } finally {
      setBusy(false);
    }
  }

  if (!supported) {
    return (
      <div className="rounded-md border border-amber-300 bg-amber-50 p-3 text-sm text-amber-800">
        当前浏览器不支持语音识别，请使用 Chrome / Edge 桌面版
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2 rounded-md border bg-white p-3 shadow-sm">
      <div className="flex items-center gap-3">
        <button
          onClick={listening ? stop : start}
          disabled={busy}
          className={`flex h-12 w-12 items-center justify-center rounded-full transition-colors ${
            listening
              ? "bg-red-500 text-white animate-pulse"
              : "bg-slate-900 text-white hover:bg-slate-700"
          } disabled:opacity-50`}
          aria-label={listening ? "停止录音" : "开始录音"}
        >
          {listening ? <MicOff size={20} /> : <Mic size={20} />}
        </button>
        <div className="flex-1 text-sm">
          <div className="font-medium text-slate-700">{hint}</div>
          {transcript && (
            <div className="text-xs text-slate-400">"{transcript}"</div>
          )}
          {error && <div className="text-xs text-red-500">错误: {error}</div>}
        </div>
      </div>
      {lastReply && (
        <div className="rounded bg-slate-50 px-3 py-2 text-sm text-slate-600">
          💬 {lastReply}
        </div>
      )}
    </div>
  );
}
