// 麦克风按钮 + 状态显示 + 错误提示 + 文本输入降级
//
// 行为:
//   - 默认: 麦克风按钮 → Web Speech API
//   - 失败/不支持: 自动显示文本输入框, 用户可以打字发指令
//   - 错误提示用中文, 区分 network / not-allowed / no-speech 等
import { useState, useRef, useEffect } from "react";
import { Mic, MicOff, Send, X, AlertTriangle } from "lucide-react";
import { useSpeech, SpeechErrorMessages, type SpeechError } from "../hooks/useSpeech";
import { understand } from "../lib/api";
import { applyIntent } from "../lib/applyIntent";
import { useCanvasStore } from "../store/canvasStore";

export function VoiceBar() {
  const [busy, setBusy] = useState(false);
  const [hint, setHint] = useState<string>("点击麦克风开始说话");
  const [lastReply, setLastReply] = useState<string>("");
  const [showTextInput, setShowTextInput] = useState(false);
  const [textValue, setTextValue] = useState("");
  const [dismissedError, setDismissedError] = useState<SpeechError | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const { supported, listening, transcript, error, permission, start, stop } = useSpeech({
    onResult: (text, isFinal) => {
      setHint(isFinal ? `你说: ${text}` : `识别中: ${text}`);
      if (isFinal && text.trim()) {
        handleCommand(text.trim());
      }
    },
    onError: (e) => {
      setHint("识别失败");
    },
  });

  // 错误出现时, 3s 后清掉 dismissed, 让气泡重新可见
  useEffect(() => {
    if (error && error !== dismissedError) {
      // 默认显示错误气泡, 5s 后用户没关就自动展开文本输入
      const t = setTimeout(() => {
        if (error === "network" || error === "not-supported" || error === "not-allowed") {
          setShowTextInput(true);
        }
      }, 1500);
      return () => clearTimeout(t);
    }
  }, [error, dismissedError]);

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
      setHint("点击麦克风继续说话");
    } catch (e: any) {
      setHint(`请求失败: ${e.message}`);
    } finally {
      setBusy(false);
    }
  }

  // 不支持浏览器
  if (!supported) {
    return (
      <div className="flex flex-col gap-2 rounded-md border bg-white p-3 shadow-sm">
        <div className="flex items-start gap-2 rounded bg-amber-50 p-2 text-sm text-amber-800">
          <AlertTriangle size={16} className="mt-0.5 shrink-0" />
          <div>
            浏览器不支持语音识别, 请用 <b>Chrome / Edge 桌面版</b>。
            可用下方文本输入框代替。
          </div>
        </div>
        <TextInput
          value={textValue}
          onChange={setTextValue}
          onSubmit={() => {
            if (textValue.trim()) {
              handleCommand(textValue.trim());
              setTextValue("");
            }
          }}
          busy={busy}
        />
      </div>
    );
  }

  const errMsg = error ? SpeechErrorMessages[error] : null;
  const showErrorBanner = errMsg && error !== dismissedError;

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
          {transcript && listening && (
            <div className="text-xs text-slate-400">"{transcript}"</div>
          )}
          {permission === "denied" && (
            <div className="text-xs text-amber-600">
              麦克风被禁用, 请在地址栏允许权限
            </div>
          )}
        </div>
        <button
          onClick={() => setShowTextInput((v) => !v)}
          className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-600 hover:bg-slate-50"
          title="切换文本输入"
        >
          ⌨
        </button>
      </div>

      {/* 错误气泡 (用中文, 不用 raw 'network') */}
      {showErrorBanner && errMsg && (
        <div className="flex items-start gap-2 rounded bg-red-50 p-2 text-xs text-red-700">
          <AlertTriangle size={14} className="mt-0.5 shrink-0" />
          <div className="flex-1">{errMsg}</div>
          <button
            onClick={() => setDismissedError(error)}
            className="shrink-0 text-red-400 hover:text-red-700"
            aria-label="关闭"
          >
            <X size={12} />
          </button>
        </div>
      )}

      {/* 文本输入 fallback (network 错误或用户主动展开) */}
      {showTextInput && (
        <TextInput
          value={textValue}
          onChange={setTextValue}
          onSubmit={() => {
            if (textValue.trim()) {
              handleCommand(textValue.trim());
              setTextValue("");
            }
          }}
          busy={busy}
          autoFocus
        />
      )}

      {lastReply && (
        <div className="rounded bg-slate-50 px-3 py-2 text-sm text-slate-600">
          💬 {lastReply}
        </div>
      )}
    </div>
  );
}

interface TextInputProps {
  value: string;
  onChange: (v: string) => void;
  onSubmit: () => void;
  busy: boolean;
  autoFocus?: boolean;
}

function TextInput({ value, onChange, onSubmit, busy, autoFocus }: TextInputProps) {
  const ref = useRef<HTMLInputElement>(null);
  useEffect(() => {
    if (autoFocus) ref.current?.focus();
  }, [autoFocus]);
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!busy) onSubmit();
      }}
      className="flex gap-2"
    >
      <input
        ref={ref}
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="输入指令, 例如: 画一个红色的大圆"
        disabled={busy}
        className="flex-1 rounded border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none disabled:opacity-50"
      />
      <button
        type="submit"
        disabled={busy || !value.trim()}
        className="flex items-center gap-1 rounded bg-slate-900 px-3 py-2 text-sm text-white hover:bg-slate-700 disabled:opacity-50"
        aria-label="发送"
      >
        <Send size={14} />
      </button>
    </form>
  );
}
