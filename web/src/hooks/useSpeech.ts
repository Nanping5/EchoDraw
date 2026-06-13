// useSpeech: 封装 Web Speech API (webkitSpeechRecognition)。
//
// 用法:
//   const { listening, transcript, start, stop, supported, error } = useSpeech({
//     onResult: (text, isFinal) => { ... }
//   });
//
// 行为:
//   - 麦克风按钮触发 start/stop
//   - continuous = false: 一句识别完就停
//   - interimResults = true: 边说边返回中间结果 (用于 UI 反馈)
//   - lang = 'zh-CN' 中文优先
//   - 不支持时 supported=false, UI 提示切 Chrome

import { useCallback, useEffect, useRef, useState } from "react";

type SpeechRecognitionLike = any;

declare global {
  interface Window {
    SpeechRecognition?: any;
    webkitSpeechRecognition?: any;
  }
}

interface Options {
  lang?: string;
  continuous?: boolean;
  interimResults?: boolean;
  onResult?: (text: string, isFinal: boolean) => void;
  onError?: (err: string) => void;
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
  const [error, setError] = useState<string | null>(null);
  const recRef = useRef<SpeechRecognitionLike>(null);

  useEffect(() => {
    const SR =
      typeof window !== "undefined"
        ? window.SpeechRecognition || window.webkitSpeechRecognition
        : null;
    setSupported(!!SR);
  }, []);

  const start = useCallback(() => {
    if (recRef.current) {
      try {
        recRef.current.stop();
      } catch {}
    }
    const SR =
      window.SpeechRecognition || window.webkitSpeechRecognition;
    if (!SR) {
      setError("浏览器不支持 Web Speech API");
      onError?.("not-supported");
      return;
    }
    const rec = new SR();
    rec.lang = lang;
    rec.continuous = continuous;
    rec.interimResults = interimResults;
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
      setError(e.error || "unknown");
      onError?.(e.error || "unknown");
      setListening(false);
    };
    rec.onend = () => {
      setListening(false);
    };
    recRef.current = rec;
    try {
      rec.start();
      setListening(true);
      setError(null);
    } catch (e: any) {
      setError(e?.message || "start failed");
      onError?.(e?.message || "start-failed");
    }
  }, [lang, continuous, interimResults, onResult, onError]);

  const stop = useCallback(() => {
    try {
      recRef.current?.stop();
    } catch {}
    setListening(false);
  }, []);

  return { supported, listening, transcript, error, start, stop };
}
