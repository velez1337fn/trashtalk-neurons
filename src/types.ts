// src/types.ts
export interface Message {
  role: "system" | "user" | "assistant" | "tool";
  content: string;
  tool_calls?: any[];
  tool_call_id?: string;
}

export interface Todo {
  id: string;
  title: string;
  status: "not started" | "in progress" | "done";
  created: number;
  updated: number;
}

export interface GitHubConfig {
  repository?: string;
  github_token?: string;
  configured: boolean;
}

export interface Session {
  id: string;
  mode: "code" | "chat";
  messages: Message[];
  created: number;
  updated: number;
  totalTokens: number;
  todos: Todo[];
}

export interface HistoryStore {
  sessions: Session[];
  currentSessionId: string | null;
}