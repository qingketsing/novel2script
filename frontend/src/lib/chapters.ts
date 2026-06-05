export function countChapters(text: string): number {
  if (!text.trim()) return 0;

  const matches = text.match(
    /^\s{0,3}#{0,6}\s*(第\s*[一二三四五六七八九十百千万\d]+\s*[章节回]|chapter\s+\d+)/gim,
  );

  return matches?.length ?? 0;
}
