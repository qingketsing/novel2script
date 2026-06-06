import type { HealthStatus, RequestStatus } from "../types";
import { FileDropzone } from "./FileDropzone";
import { HealthBadge } from "./HealthBadge";

type InputPanelProps = {
  content: string;
  error: string;
  file: File | null;
  healthStatus: HealthStatus;
  isFileEdited: boolean;
  isReadingFile: boolean;
  status: RequestStatus;
  title: string;
  onContentChange: (value: string) => void;
  onFileChange: (file: File | null, error?: string) => void;
  onGenerate: () => void;
  onHealthCheck: () => void;
  onLoadSample: () => void;
  onTitleChange: (value: string) => void;
};

export function InputPanel({
  content,
  error,
  file,
  healthStatus,
  isFileEdited,
  isReadingFile,
  status,
  title,
  onContentChange,
  onFileChange,
  onGenerate,
  onHealthCheck,
  onLoadSample,
  onTitleChange,
}: InputPanelProps) {
  const isGenerating = status === "loading";

  return (
    <section className="space-y-4 rounded-lg border border-zinc-200 bg-white p-4 shadow-sm shadow-zinc-200/60">
      <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-end">
        <label className="block">
          <span className="field-label">标题</span>
          <input
            className="field"
            value={title}
            onChange={(event) => onTitleChange(event.target.value)}
            placeholder="可为空"
          />
        </label>
        <button className="secondary-button" type="button" onClick={onHealthCheck}>
          {healthStatus === "checking" ? "检查中" : "健康检查"}
        </button>
      </div>

      <div className="flex flex-wrap items-center gap-2 text-sm">
        <span className="text-zinc-500">后端状态</span>
        <HealthBadge status={healthStatus} />
      </div>

      <label className="block">
        <div className="mb-1.5 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <span className="field-label mb-0">粘贴小说正文</span>
          <button className="secondary-button" type="button" onClick={onLoadSample}>
            加载示例文本
          </button>
        </div>
        <textarea
          className="field min-h-[380px] resize-y font-mono text-sm leading-6"
          value={content}
          onChange={(event) => onContentChange(event.target.value)}
          placeholder="粘贴至少 3 个章节，例如：第一章 / 第二章 / 第三章"
        />
      </label>

      <FileDropzone
        disabled={isGenerating}
        file={file}
        isEdited={isFileEdited}
        isReading={isReadingFile}
        onFileChange={onFileChange}
      />

      {error ? <div className="error-box">{error}</div> : null}

      <button
        className="primary-button"
        type="button"
        onClick={onGenerate}
        disabled={isGenerating || isReadingFile}
      >
        {isReadingFile ? "正在读取文件..." : isGenerating ? "正在解析章节并生成剧本..." : "生成剧本 YAML"}
      </button>
    </section>
  );
}
