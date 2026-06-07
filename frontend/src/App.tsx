import { useMemo, useState } from "react";
import { AppHeader } from "./components/AppHeader";
import { InputPanel } from "./components/InputPanel";
import { ResultPanel } from "./components/ResultPanel";
import { ScreenplayPreviewPage } from "./components/ScreenplayPreviewPage";
import {
  BACKEND_CONNECTION_ERROR,
  checkBackendHealth,
  convertText,
  errorMessage,
} from "./lib/api";
import { countChapters } from "./lib/chapters";
import { copyText } from "./lib/clipboard";
import { downloadTextFile, slugify } from "./lib/download";
import { generationModeLabel } from "./lib/mode";
import { SAMPLE_TEXT } from "./lib/sample";
import type { ConvertResponse, CopyState, HealthStatus, RequestStatus } from "./types";

export function App() {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [importedContent, setImportedContent] = useState("");
  const [isReadingFile, setIsReadingFile] = useState(false);
  const [status, setStatus] = useState<RequestStatus>("idle");
  const [error, setError] = useState("");
  const [result, setResult] = useState<ConvertResponse | null>(null);
  const [copyState, setCopyState] = useState<CopyState>("idle");
  const [healthStatus, setHealthStatus] = useState<HealthStatus>("unchecked");
  const [isPreviewOpen, setIsPreviewOpen] = useState(false);

  const estimatedChapterCount = useMemo(() => countChapters(content), [content]);
  const statusLabel = result
    ? `已识别 ${result.chapter_count} 章 / ${generationModeLabel(result.mode)}`
    : estimatedChapterCount > 0
      ? `已识别 ${estimatedChapterCount} 章`
      : "等待输入";

  async function handleGenerate() {
    if (status === "loading") return;

    setError("");
    setCopyState("idle");

    const trimmedContent = content.trim();
    if (trimmedContent.length === 0) {
      setError("请上传 .txt / .md 文件，或粘贴小说正文。");
      return;
    }

    if (countChapters(trimmedContent) < 3) {
      setError("至少需要 3 个章节才能生成剧本初稿。");
      return;
    }

    setStatus("loading");
    try {
      const response = await convertText(title, trimmedContent);
      setResult(response);
      setIsPreviewOpen(false);
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
      setError(ok ? "" : BACKEND_CONNECTION_ERROR);
    } catch {
      setHealthStatus("failed");
      setError(BACKEND_CONNECTION_ERROR);
    }
  }

  async function handleCopy() {
    if (!result?.screenplay_yaml) return;

    try {
      await copyText(result.screenplay_yaml);
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
    setResult(null);
    setIsPreviewOpen(false);
  }

  async function handleFileChange(nextFile: File | null, nextError = "") {
    setError(nextError);
    setResult(null);
    setIsPreviewOpen(false);
    if (!nextFile) {
      setFile(null);
      setImportedContent("");
      return;
    }

    setIsReadingFile(true);
    try {
      const text = await nextFile.text();
      if (text.trim().length === 0) {
        setError("上传文件内容不能为空。");
        return;
      }

      setFile(nextFile);
      setImportedContent(text);
      setContent(text);
    } catch {
      setError("读取上传文件失败，请重新选择文件。");
    } finally {
      setIsReadingFile(false);
    }
  }

  function loadSampleText() {
    setTitle("雨夜来信");
    setContent(SAMPLE_TEXT);
    setFile(null);
    setImportedContent("");
    setError("");
    setResult(null);
    setIsPreviewOpen(false);
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
            isFileEdited={Boolean(file) && content !== importedContent}
            isReadingFile={isReadingFile}
            status={status}
            title={title}
            onContentChange={handleContentChange}
            onFileChange={handleFileChange}
            onGenerate={handleGenerate}
            onHealthCheck={handleHealthCheck}
            onLoadSample={loadSampleText}
            onTitleChange={setTitle}
          />
          <ResultPanel
            copyState={copyState}
            result={result}
            onCopy={handleCopy}
            onDownload={handleDownload}
            onOpenPreview={() => setIsPreviewOpen(true)}
          />
        </section>
      </div>

      {isPreviewOpen && result?.screenplay_yaml ? (
        <ScreenplayPreviewPage
          chapterCount={result.chapter_count}
          mode={result.mode}
          yamlText={result.screenplay_yaml}
          onClose={() => setIsPreviewOpen(false)}
        />
      ) : null}
    </main>
  );
}
