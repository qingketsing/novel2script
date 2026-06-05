export type RequestStatus = "idle" | "loading";
export type CopyState = "idle" | "done" | "failed";
export type HealthStatus = "unchecked" | "checking" | "ok" | "failed";

export type ConvertResponse = {
  screenplay_yaml: string;
  chapter_count: number;
  mode: string;
};

export type ApiErrorResponse = {
  error?: {
    message?: string;
  };
};
