// Mirrors the Go DTOs in internal/api/dto.go.

export type JobStatus =
  | "PENDING"
  | "RUNNING"
  | "COMPLETED"
  | "FAILED"
  | "CANCELED";

export interface TranscribeJob {
  id: string;
  path: string;
  language: string;
  format: string;
  output_path: string;
  progress: number;
  status: JobStatus;
  result: string;
  callback: string;
  model: string;
  duration: string;
  priority: number;
  error?: string;
}

export interface TranscribeInput {
  path: string;
  output_path: string;
  language?: string;
  format?: string;
  callback?: string;
  priority?: number;
  model?: string;
}

export interface ModelInfo {
  id: string;
  name: string;
  default: boolean;
}
