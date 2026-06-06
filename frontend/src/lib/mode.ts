export function generationModeLabel(mode: string): string {
  switch (mode) {
    case "mock":
      return "Mock 演示模式";
    case "api":
      return "DeepSeek API 模式";
    default:
      return mode;
  }
}
