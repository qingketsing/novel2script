import { useMemo } from "react";
import type { ReactNode } from "react";
import { generationModeLabel } from "../lib/mode";
import { parseScreenplayPreview } from "../lib/screenplayYaml";
import type { PreviewBeat, PreviewScene, ScreenplayPreview } from "../types";

type ScreenplayPreviewPageProps = {
  yamlText: string;
  chapterCount: number;
  mode: string;
  onClose: () => void;
};

export function ScreenplayPreviewPage({ yamlText, chapterCount, mode, onClose }: ScreenplayPreviewPageProps) {
  const parseResult = useMemo(() => parseScreenplayPreview(yamlText), [yamlText]);

  return (
    <section className="fixed inset-0 z-50 overflow-auto bg-zinc-50 text-ink">
      <div className="mx-auto flex w-full max-w-[1440px] flex-col gap-5 px-4 py-5 sm:px-6 lg:px-8">
        <header className="flex flex-col justify-between gap-4 rounded-lg border border-zinc-200 bg-white p-4 shadow-sm shadow-zinc-200/60 lg:flex-row lg:items-center">
          <div>
            <p className="text-sm font-medium text-zinc-500">只读剧本预览</p>
            <h1 className="mt-1 text-2xl font-semibold text-zinc-950">
              {parseResult.ok ? parseResult.preview.title : "剧本预览"}
            </h1>
            <p className="mt-2 text-sm text-zinc-500">
              已识别 {chapterCount} 个章节，生成模式：{generationModeLabel(mode)}
            </p>
          </div>
          <button className="secondary-button self-start lg:self-auto" type="button" onClick={onClose}>
            返回生成页
          </button>
        </header>

        {parseResult.ok ? <ReadablePreview preview={parseResult.preview} /> : <PreviewParseError message={parseResult.message} />}
      </div>
    </section>
  );
}

function ReadablePreview({ preview }: { preview: ScreenplayPreview }) {
  const scenes = preview.acts.flatMap((act) => act.scenes);

  return (
    <>
      <section className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="来源章节" value={preview.sourceChapterCount || preview.sourceChapters.length} />
        <StatCard label="人物" value={preview.characters.length} />
        <StatCard label="幕" value={preview.acts.length} />
        <StatCard label="场景" value={scenes.length} />
      </section>

      <section className="grid gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
        <aside className="flex flex-col gap-5">
          <PreviewSection title="人物卡片" emptyText="YAML 中没有可展示的人物。">
            {preview.characters.map((character) => (
              <article key={character.id} className="rounded-lg border border-zinc-200 bg-white p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h2 className="text-base font-semibold text-zinc-950">{character.name}</h2>
                    <p className="mt-1 text-sm text-zinc-500">{character.role}</p>
                  </div>
                  <span className="rounded-full bg-zinc-100 px-2 py-1 text-xs font-medium text-zinc-600">{character.id}</span>
                </div>
                {character.description ? <p className="mt-3 text-sm leading-6 text-zinc-700">{character.description}</p> : null}
                {character.motivation ? (
                  <p className="mt-3 rounded-lg bg-blue-50 px-3 py-2 text-sm leading-6 text-blue-900">动机：{character.motivation}</p>
                ) : null}
              </article>
            ))}
          </PreviewSection>

          <PreviewSection title="来源章节" emptyText="YAML 中没有来源章节。">
            {preview.sourceChapters.map((chapter) => (
              <article key={chapter.id} className="rounded-lg border border-zinc-200 bg-white p-3">
                <div className="flex items-center justify-between gap-3">
                  <h2 className="text-sm font-semibold text-zinc-950">{chapter.title}</h2>
                  <span className="text-xs text-zinc-500">第 {chapter.order || "-"} 章</span>
                </div>
                {chapter.summary ? <p className="mt-2 text-sm leading-6 text-zinc-600">{chapter.summary}</p> : null}
              </article>
            ))}
          </PreviewSection>
        </aside>

        <section className="flex flex-col gap-5">
          {preview.acts.map((act) => (
            <article key={act.id} className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm shadow-zinc-200/60">
              <div className="border-b border-zinc-200 pb-3">
                <p className="text-sm font-medium text-zinc-500">第 {act.order || "-"} 幕</p>
                <h2 className="mt-1 text-xl font-semibold text-zinc-950">{act.title}</h2>
              </div>
              <div className="mt-4 flex flex-col gap-4">
                {act.scenes.length > 0 ? (
                  act.scenes.map((scene) => <SceneCard key={scene.id} scene={scene} preview={preview} />)
                ) : (
                  <p className="rounded-lg bg-zinc-50 px-3 py-2 text-sm text-zinc-500">这一幕暂无场景。</p>
                )}
              </div>
            </article>
          ))}
        </section>
      </section>
    </>
  );
}

function SceneCard({ scene, preview }: { scene: PreviewScene; preview: ScreenplayPreview }) {
  const characterNames = scene.characters
    .map((id) => preview.characters.find((character) => character.id === id)?.name || id)
    .filter(Boolean);

  return (
    <article className="rounded-lg border border-zinc-200 bg-zinc-50 p-4">
      <div className="flex flex-col justify-between gap-3 md:flex-row md:items-start">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <span className="rounded-full bg-zinc-950 px-2 py-1 text-xs font-medium text-white">{scene.id}</span>
            <span className="rounded-full bg-white px-2 py-1 text-xs font-medium text-zinc-600">
              {scene.heading.interior === null ? "内外景未标注" : scene.heading.interior ? "内景" : "外景"}
            </span>
          </div>
          <h3 className="mt-3 text-base font-semibold text-zinc-950">
            {scene.heading.location} / {scene.heading.time}
          </h3>
        </div>
        <div className="flex flex-wrap gap-2">
          {scene.sourceChapterIds.map((id) => (
            <span key={id} className="rounded-full bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700">
              {id}
            </span>
          ))}
        </div>
      </div>

      {scene.summary ? <p className="mt-3 text-sm leading-6 text-zinc-700">{scene.summary}</p> : null}

      {characterNames.length > 0 ? (
        <p className="mt-3 text-sm text-zinc-500">出场人物：{characterNames.join("、")}</p>
      ) : null}

      <div className="mt-4 flex flex-col gap-2">
        {scene.beats.length > 0 ? (
          scene.beats.map((beat, index) => <BeatRow key={`${scene.id}-${index}`} beat={beat} />)
        ) : (
          <p className="rounded-lg bg-white px-3 py-2 text-sm text-zinc-500">暂无节拍。</p>
        )}
      </div>
    </article>
  );
}

function BeatRow({ beat }: { beat: PreviewBeat }) {
  const isDialogue = beat.type === "dialogue";
  const label = isDialogue ? beat.characterName || beat.characterId || "角色" : beatLabel(beat.type);

  return (
    <div className="rounded-lg bg-white px-3 py-2">
      <p className="text-xs font-semibold uppercase tracking-wide text-zinc-400">{label}</p>
      <p className="mt-1 whitespace-pre-wrap text-sm leading-6 text-zinc-800">{beat.text || "未填写内容"}</p>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm shadow-zinc-200/60">
      <p className="text-sm text-zinc-500">{label}</p>
      <p className="mt-2 text-2xl font-semibold text-zinc-950">{value}</p>
    </div>
  );
}

function PreviewSection({ title, emptyText, children }: { title: string; emptyText: string; children: ReactNode }) {
  const items = Array.isArray(children) ? children.filter(Boolean) : children;
  const isEmpty = Array.isArray(items) ? items.length === 0 : !items;

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-base font-semibold text-zinc-950">{title}</h2>
      {isEmpty ? <p className="rounded-lg border border-zinc-200 bg-white p-3 text-sm text-zinc-500">{emptyText}</p> : items}
    </section>
  );
}

function PreviewParseError({ message }: { message: string }) {
  return (
    <section className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-800">
      <h2 className="text-base font-semibold">无法打开结构化预览</h2>
      <p className="mt-2 text-sm leading-6">{message}</p>
      <p className="mt-2 text-sm leading-6">可以返回生成页查看 YAML 源码，或重新生成后再打开预览。</p>
    </section>
  );
}

function beatLabel(type: string): string {
  switch (type) {
    case "action":
      return "动作";
    case "note":
      return "备注";
    default:
      return type || "节拍";
  }
}
