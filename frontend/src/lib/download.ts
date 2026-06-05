export function downloadTextFile(content: string, filename: string): void {
  const blob = new Blob([content], {
    type: "application/x-yaml;charset=utf-8",
  });
  const href = URL.createObjectURL(blob);
  const link = document.createElement("a");

  link.href = href;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(href);
}

export function slugify(value: string): string {
  return (
    value
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9\u4e00-\u9fa5]+/gi, "-")
      .replace(/^-+|-+$/g, "") || "screenplay"
  );
}
