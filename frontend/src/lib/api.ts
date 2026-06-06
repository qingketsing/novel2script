import type { ApiErrorResponse, ConvertResponse } from "../types";

export const BACKEND_CONNECTION_ERROR = "无法连接后端服务，请确认服务已启动";

export async function checkBackendHealth(): Promise<boolean> {
  const response = await fetch("/health");
  const body = (await response.json().catch(() => null)) as { status?: string } | null;

  return response.ok && body?.status === "ok";
}

export async function convertText(title: string, content: string): Promise<ConvertResponse> {
  const response = await fetch("/api/convert", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      title,
      content,
      input_type: "text",
    }),
  });

  return readApiResponse(response);
}

export async function convertUploadedFile(title: string, file: File): Promise<ConvertResponse> {
  const formData = new FormData();
  formData.append("title", title);
  formData.append("file", file);

  const response = await fetch("/api/convert/upload", {
    method: "POST",
    body: formData,
  });

  return readApiResponse(response);
}

export function errorMessage(error: unknown): string {
  if (error instanceof TypeError) {
    return BACKEND_CONNECTION_ERROR;
  }

  return error instanceof Error && error.message ? error.message : "生成失败，请稍后重试。";
}

async function readApiResponse(response: Response): Promise<ConvertResponse> {
  const body = (await response.json().catch(() => null)) as ApiErrorResponse | ConvertResponse | null;

  if (!response.ok) {
    throw new Error(apiErrorMessage(body) || fallbackErrorMessage(response, body));
  }

  return body as ConvertResponse;
}

function fallbackErrorMessage(response: Response, body: ApiErrorResponse | ConvertResponse | null): string {
  if (!body && response.status >= 500) {
    return BACKEND_CONNECTION_ERROR;
  }

  return "请求失败，请检查输入后重试。";
}

function apiErrorMessage(body: ApiErrorResponse | ConvertResponse | null): string {
  if (body && "error" in body) {
    return body.error?.message ?? "";
  }

  return "";
}
