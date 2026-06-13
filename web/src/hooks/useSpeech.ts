// useSpeech: 封装 Web Speech API (webkitSpeechRecognition)。
//
// 关键设计:
//   1. 先 getUserMedia 拿麦克风权限, 再 start recognition
//      解决 Chrome 'network' 错误时连麦克风权限都不弹的问题
//   2. 错误码 → 中文提示的映射, 用户能看到具体原因
//   3. 暴露 supported/listening/error, 上层 UI 根据 error 决定是否降级到文本输入
//
// 已知限制:
//   - Web Speech API 依赖 Google STT 服务, 国内/部分网络环境会 'network' 失败
//   - 这是 Chrome 设计, 没法绕开; 后续 v0.2 可加 Whisper 自托管 API 兜底
//   - 此版本只修麦克风权限时序和错误提示, 不接入 Whisper

import { useCallback, useEffect, useRef, useState } from "react";

type SpeechRecognitionLike = any;

declare global {
  interface Window {
    SpeechRecognition?: any;
    webkitSpeechRecognition?: any;
  }
}

export type SpeechError =
  | "not-supported"
  | "not-allowed" // 用户拒绝麦克风
  | "no-speech" // 静默超时
  | "audio-capture" // 没有麦克风设备
  | "network" // Google STT 不可达 (国内常见)
  | "aborted"
  | "language-not-supported"
  | "service-not-allowed"
  | "unknown";

export const SpeechErrorMessages: Record<SpeechError, string> = {
  "not-supported": "当前浏览器不支持语音识别, 请用 Chrome / Edge 桌面版",
  "not-allowed": "麦克风权限被拒绝, 请在浏览器地址栏允许后重试",
  "no-speech": "没有检测到语音, 请重试",
  "audio-capture": "没有可用的麦克风设备",
  "network": "无法连接 Google 语音服务, 请检查网络 (国内可能需代理) 或改用文本输入",
  "aborted": "识别已取消",
  "language-not-supported": "当前语言不支持, 试试中文(zh-CN)或英文(en-US)",
  "service-not-allowed": "浏览器禁止了语音服务, 检查 Chrome 设置 → 隐私 → 语音服务",
  "unknown": "未知错误, 请刷新页面重试",
};

interface Options {
  lang?: string;
  continuous?: boolean;
  interimResults?: boolean;
  onResult?: (text: string, isFinal: boolean) => void;
  onError?: (err: SpeechError) => void;
}

export function useSpeech(opts: Options = {}) {
  const {
    lang = "zh-CN",
    continuous = false,
    interimResults = true,
    onResult,
    onError,
  } = opts;

  const [supported, setSupported] = useState(false);
  const [listening, setListening] = useState(false);
  const [transcript, setTranscript] = useState("");
  const [error, setError] = useState<SpeechError | null>(null);
  const [permission, setPermission] = useState<"unknown" | "granted" | "denied">("unknown");
  const recRef = useRef<SpeechRecognitionLike>(null);
  const streamRef = useRef<MediaStream | null>(null);

  // 检测 API 支持
  useEffect(() => {
    const SR =
      typeof window !== "undefined"
        ? window.SpeechRecognition || window.webkitSpeechRecognition
        : null;
    setSupported(!!SR);
  }, []);

  // 卸载时清理
  useEffect(() => {
    return () => {
      try {
        recRef.current?.stop();
      } catch {}
      streamRef.current?.getTracks().forEach((t) => t.stop());
    };
  }, []);

  const start = useCallback(async () => {
    // 0) 检测 API 支持
    const SR = window.SpeechRecognition || window.webkitSpeechRecognition;
    if (!SR) {
      setError("not-supported");
      onError?.("not-supported");
      return;
    }

    setError(null);
    setTranscript("");

    // 1) 先拿麦克风权限 (关键: 解决 'network' 错误时连权限都不弹)
    if (navigator.mediaDevices?.getUserMedia) {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        streamRef.current = stream;
        setPermission("granted");
        // 拿到权限后立即 stop track (Web Speech 用自己的音频流)
        stream.getTracks().forEach((t) => t.stop());
      } catch (e: any) {
        const msg: string = e?.message || e?.name || "";
        if (msg.includes("Permission denied") || msg.includes("NotAllowed")) {
          setError("not-allowed");
          setPermission("denied");
          onError?.("not-allowed");
        } else if (msg.includes("NotFound") || msg.includes("audio-capture")) {
          setError("audio-capture");
          onError?.("audio-capture");
        } else {
          setError("unknown");
          onError?.("unknown");
        }
        return;
      }
    }

    // 2) 启动 Web Speech
    if (recRef.current) {
      try {
        recRef.current.stop();
      } catch {}
    }
    const rec = new SR();
    rec.lang = lang;
    rec.continuous = continuous;
    rec.interimResults = interimResults;
    rec.maxAlternatives = 1;

    rec.onresult = (e: any) => {
      let interim = "";
      let final = "";
      for (let i = e.resultIndex; i < e.results.length; i++) {
        const r = e.results[i];
        if (r.isFinal) final += r[0].transcript;
        else interim += r[0].transcript;
      }
      const text = (final + interim).trim();
      setTranscript(text);
      onResult?.(text, !!final);
    };

    rec.onerror = (e: any) => {
      const code = (e.error || "unknown") as SpeechError;
      setError(code);
      onError?.(code);
      setListening(false);
    };

    rec.onend = () => {
      setListening(false);
    };

    rec.onstart = () => {
      setListening(true);
    };

    recRef.current = rec;
    try {
      rec.start();
    } catch (e: any) {
      setError("unknown");
      onError?.("unknown");
    }
  }, [lang, continuous, interimResults, onResult, onError]);

  const stop = useCallback(() => {
    try {
      recRef.current?.stop();
    } catch {}
    setListening(false);
  }, []);

  return { supported, listening, transcript, error, permission, start, stop };
}
