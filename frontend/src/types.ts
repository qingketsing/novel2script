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

export type ScreenplayPreview = {
  title: string;
  sourceChapterCount: number;
  generatedMode: string;
  characters: PreviewCharacter[];
  sourceChapters: PreviewSourceChapter[];
  acts: PreviewAct[];
};

export type PreviewCharacter = {
  id: string;
  name: string;
  role: string;
  description: string;
  motivation: string;
};

export type PreviewSourceChapter = {
  id: string;
  title: string;
  order: number;
  summary: string;
};

export type PreviewAct = {
  id: string;
  title: string;
  order: number;
  scenes: PreviewScene[];
};

export type PreviewScene = {
  id: string;
  sourceChapterIds: string[];
  heading: {
    location: string;
    time: string;
    interior: boolean | null;
  };
  summary: string;
  characters: string[];
  beats: PreviewBeat[];
};

export type PreviewBeat = {
  type: string;
  text: string;
  characterId: string;
  characterName: string;
};
