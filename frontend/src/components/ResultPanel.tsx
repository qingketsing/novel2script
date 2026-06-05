import type { CopyState, ConvertResponse } from "../types";

type ResultPanelProps = {
  copyState: CopyState;
  result: ConvertResponse | null;
  onCopy: () => void;
  onDownload: () => void;
};

export function ResultPanel({ copyState, result, onCopy, onDownload }: ResultPanelProps) {
  const hasYaml = Boolean(result?.screenplay_yaml);

  return (
    <section className="grid min-h-[620px] gap-4 rounded-lg border border-zinc-200 bg-white p-4 shadow-sm shadow-zinc-200/60">
      <div className="flex flex-col justify-between gap-3 border-b border-zinc-200 pb-3 sm:flex-row sm:items-center">
        <div>
          <h2 className="text-lg font-semibold text-zinc-950">YAML 输出</h2>
          <p className="text-sm text-zinc-500">
            {result ? `章节数 ${result.chapter_count}，生成模式 ${result.mode}` : "生成后显示结构化结果"}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <button className="secondary-button" type="button" onClick={onCopy} disabled={!hasYaml}>
            {copyState === "done" ? "已复制" : "复制 YAML"}
          </button>
          <button className="secondary-button" type="button" onClick={onDownload} disabled={!hasYaml}>
            下载 YAML
          </button>
        </div>
      </div>

      <pre className="yaml-panel">{result?.screenplay_yaml || 'schema_version: "1.0"\n# 等待生成...'}</pre>

      {copyState === "failed" ? (
        <p className="text-sm text-red-700">复制失败，请手动选择 YAML 内容复制。</p>
      ) : null}
    </section>
  );
}
