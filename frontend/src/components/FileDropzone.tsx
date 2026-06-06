import * as Label from "@radix-ui/react-label";
import type { ChangeEvent, DragEvent } from "react";
import { useId, useState } from "react";

const ACCEPTED_FILE_TYPES = [".txt", ".md"] as const;
const MAX_FILE_SIZE = 2 * 1024 * 1024;

type FileDropzoneProps = {
  disabled?: boolean;
  file: File | null;
  isEdited: boolean;
  isReading: boolean;
  onFileChange: (file: File | null, error?: string) => void;
};

export function FileDropzone({ disabled = false, file, isEdited, isReading, onFileChange }: FileDropzoneProps) {
  const inputId = useId();
  const [isDragging, setIsDragging] = useState(false);
  const isDisabled = disabled || isReading;

  function handleFileInputChange(event: ChangeEvent<HTMLInputElement>) {
    if (isDisabled) return;

    const selected = event.target.files?.[0] ?? null;
    applyFile(selected, event.currentTarget);
  }

  function handleDragOver(event: DragEvent<HTMLLabelElement>) {
    event.preventDefault();
    event.dataTransfer.dropEffect = isDisabled ? "none" : "copy";
    if (isDisabled) return;

    setIsDragging(true);
  }

  function handleDragLeave(event: DragEvent<HTMLLabelElement>) {
    if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
      setIsDragging(false);
    }
  }

  function handleDrop(event: DragEvent<HTMLLabelElement>) {
    event.preventDefault();
    setIsDragging(false);
    if (isDisabled) return;

    applyFile(event.dataTransfer.files[0] ?? null);
  }

  function applyFile(selected: File | null, input?: HTMLInputElement) {
    if (isDisabled) return;

    if (!selected) {
      onFileChange(null);
      return;
    }

    if (!isAcceptedFile(selected.name)) {
      if (input) input.value = "";
      onFileChange(null, "当前仅支持 .txt 或 .md 文件。");
      return;
    }

    if (selected.size > MAX_FILE_SIZE) {
      if (input) input.value = "";
      onFileChange(null, "上传文件不能超过 2MB。");
      return;
    }

    onFileChange(selected);
  }

  function removeFile() {
    if (isDisabled) return;

    onFileChange(null);
  }

  return (
    <div className="space-y-2">
      <div>
        <span className="field-label">上传小说文件</span>
      </div>
      <Label.Root
        className={dropzoneClassName(isDragging, isDisabled)}
        htmlFor={inputId}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
      >
        <input
          id={inputId}
          className="sr-only"
          type="file"
          accept={ACCEPTED_FILE_TYPES.join(",")}
          disabled={isDisabled}
          onChange={handleFileInputChange}
        />
        <span className="flex h-10 w-10 items-center justify-center rounded-full border border-zinc-200 bg-white text-lg font-semibold text-accent">
          +
        </span>
        <span className="mt-3 text-sm font-semibold text-zinc-900">
          {isReading ? "正在读取文件..." : disabled ? "生成中，暂不可更换文件" : "拖拽 .txt / .md 到这里，或点击选择文件"}
        </span>
        <span className="mt-1 max-w-sm text-sm leading-6 text-zinc-500">
          文件内容会导入上方正文，可编辑后再生成
        </span>
      </Label.Root>

      {file ? (
        <div className="flex flex-col gap-3 rounded-lg border border-zinc-200 bg-white px-3 py-3 text-sm shadow-sm shadow-zinc-200/60 sm:flex-row sm:items-center sm:justify-between">
          <div className="min-w-0">
            <p className="truncate font-medium text-zinc-900">{file.name}</p>
            <p className="mt-1 text-zinc-500">
              {formatFileSize(file.size)} / {file.type || file.name.split(".").pop()?.toUpperCase() || "文本文件"}
            </p>
            <p className={`mt-1 ${isEdited ? "text-amber-700" : "text-emerald-700"}`}>
              {isEdited ? "内容已编辑，将按当前正文生成" : "内容已导入，可在上方编辑"}
            </p>
          </div>
          <button className="secondary-button shrink-0" type="button" onClick={removeFile} disabled={isDisabled}>
            移除文件
          </button>
        </div>
      ) : null}
    </div>
  );
}

function dropzoneClassName(isDragging: boolean, isDisabled: boolean): string {
  const base =
    "flex min-h-44 cursor-pointer flex-col items-center justify-center rounded-lg border border-dashed px-5 py-8 text-center outline-none transition";

  if (isDisabled) {
    return `${base} cursor-not-allowed border-zinc-200 bg-zinc-100 text-zinc-400`;
  }

  return isDragging
    ? `${base} border-accent bg-blue-50 ring-4 ring-blue-100`
    : `${base} border-zinc-300 bg-zinc-50 hover:border-zinc-400 hover:bg-white focus-visible:border-accent focus-visible:ring-4 focus-visible:ring-blue-100`;
}

function isAcceptedFile(filename: string): boolean {
  return ACCEPTED_FILE_TYPES.some((extension) => filename.toLowerCase().endsWith(extension));
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}
