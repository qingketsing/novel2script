import { useMemo, useState } from "react";
import { AppHeader } from "./components/AppHeader";
import { InputPanel } from "./components/InputPanel";
import { ResultPanel } from "./components/ResultPanel";
import {
  checkBackendHealth,
  convertText,
  convertUploadedFile,
  errorMessage,
} from "./lib/api";
import { countChapters } from "./lib/chapters";
import { downloadTextFile, slugify } from "./lib/download";
import { SAMPLE_TEXT } from "./lib/sample";
import type { ConvertResponse, CopyState, HealthStatus, RequestStatus } from "./types";

export function App() {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [status, setStatus] = useState<RequestStatus>("idle");
  const [error, setError] = useState("");
  const [result, setResult] = useState<ConvertResponse | null>(null);
  const [copyState, setCopyState] = useState<CopyState>("idle");
  const [healthStatus, setHealthStatus] = useState<HealthStatus>("unchecked");

  const estimatedChapterCount = useMemo(() => countChapters(content), [content]);
  const statusLabel = result
    ? `${result.chapter_count} 章 / ${result.mode} mode`
    : estimatedChapterCount > 0
      ? `已识别 ${estimatedChapterCount} 章`
      : "等待输入";

  async function handleGenerate() {
    setError("");
    setCopyState("idle");

    const trimmedContent = content.trim();
    if (!file && trimmedContent.length === 0) {
      setError("请上传 .txt / .md 文件，或粘贴小说正文。");
      return;
    }

    if (!file && countChapters(trimmedContent) < 3) {
      setError("至少需要 3 个章节才能生成剧本初稿。");
      return;
    }

    setStatus("loading");
    try {
      const response = file ? await convertUploadedFile(title, file) : await convertText(title, trimmedContent);
      setResult(response);
    } catch (err: unknown) {
      setError(errorMessage(err));
    } finally {
      setStatus("idle");
    }
  }

  async function handleHealthCheck() {
    setHealthStatus("checking");
    setError("");
    try {
      const ok = await checkBackendHealth();
      setHealthStatus(ok ? "ok" : "failed");
      setError(ok ? "" : "后端不可用，请确认 Go 后端正在 8080 端口运行。");
    } catch {
      setHealthStatus("failed");
      setError("后端不可用，请确认 Go 后端正在 8080 端口运行。");
    }
  }

  async function handleCopy() {
    if (!result?.screenplay_yaml) return;

    try {
      await navigator.clipboard.writeText(result.screenplay_yaml);
      setCopyState("done");
      window.setTimeout(() => setCopyState("idle"), 1800);
    } catch {
      setCopyState("failed");
    }
  }

  function handleDownload() {
    if (!result?.screenplay_yaml) return;

    downloadTextFile(result.screenplay_yaml, `${slugify(title || "screenplay")}.yaml`);
  }

  function handleContentChange(value: string) {
    setContent(value);
    setFile(null);
    setResult(null);
  }

  function handleFileChange(nextFile: File | null, nextError = "") {
    setError(nextError);
    setResult(null);
    setFile(nextFile);
  }

  function loadSampleText() {
    setTitle("雨夜来信");
    setContent(SAMPLE_TEXT);
    setFile(null);
    setError("");
    setResult(null);
  }

  return (
    <main className="min-h-[100dvh] bg-zinc-50 text-ink">
      <div className="mx-auto flex w-full max-w-[1440px] flex-col gap-5 px-4 py-5 sm:px-6 lg:px-8">
        <AppHeader statusLabel={statusLabel} />

        <section className="grid gap-5 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
          <InputPanel
            content={content}
            error={error}
            file={file}
            healthStatus={healthStatus}
            status={status}
            title={title}
            onContentChange={handleContentChange}
            onFileChange={handleFileChange}
            onGenerate={handleGenerate}
            onHealthCheck={handleHealthCheck}
            onLoadSample={loadSampleText}
            onTitleChange={setTitle}
          />
          <ResultPanel copyState={copyState} result={result} onCopy={handleCopy} onDownload={handleDownload} />
        </section>
      </div>
    </main>
  );
}
