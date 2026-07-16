// src/history.ts
import fs from "fs";
import path from "path";
import { CONFIG_DIR, HISTORY_FILE } from "./config";
import { Session, Message, HistoryStore, Todo, GitHubConfig } from "./types";

function ensureConfigDir() {
  if (!fs.existsSync(CONFIG_DIR)) {
    fs.mkdirSync(CONFIG_DIR, { recursive: true });
  }
}

function loadStore(): HistoryStore {
  ensureConfigDir();
  if (fs.existsSync(HISTORY_FILE)) {
    try {
      const data = fs.readFileSync(HISTORY_FILE, "utf-8");
      const parsed = JSON.parse(data);
      if (parsed.sessions) {
        parsed.sessions = parsed.sessions.map((s: any) => ({
          ...s,
          totalTokens: typeof s.totalTokens === 'number' ? s.totalTokens : 0,
          todos: s.todos || [],
        }));
      }
      return parsed;
    } catch (e) {
      return { sessions: [], currentSessionId: null };
    }
  }
  return { sessions: [], currentSessionId: null };
}

function saveStore(store: HistoryStore) {
  ensureConfigDir();
  fs.writeFileSync(HISTORY_FILE, JSON.stringify(store, null, 2), "utf-8");
}

function generateId(): string {
  return Date.now().toString(36) + "-" + Math.random().toString(36).slice(2, 7);
}

function getCurrentSessionOrCreate(): Session {
  const store = loadStore();
  let session = store.sessions.find(s => s.id === store.currentSessionId);
  if (!session) {
    session = {
      id: generateId(),
      mode: "code",
      messages: [],
      created: Date.now(),
      updated: Date.now(),
      totalTokens: 0,
      todos: [],
    };
    store.sessions.push(session);
    store.currentSessionId = session.id;
    saveStore(store);
  }
  return session;
}

export function getCurrentSession(): Session | null {
  const store = loadStore();
  if (!store.currentSessionId) return null;
  const session = store.sessions.find(s => s.id === store.currentSessionId);
  return session || null;
}

export function createSession(mode: "code" | "chat"): Session {
  const store = loadStore();
  const session: Session = {
    id: generateId(),
    mode,
    messages: [],
    created: Date.now(),
    updated: Date.now(),
    totalTokens: 0,
    todos: [],
  };
  store.sessions.push(session);
  store.currentSessionId = session.id;
  saveStore(store);
  return session;
}

export function switchSession(sessionId: string): Session | null {
  const store = loadStore();
  const session = store.sessions.find(s => s.id === sessionId);
  if (session) {
    store.currentSessionId = sessionId;
    saveStore(store);
    return session;
  }
  return null;
}

export function deleteSession(sessionId: string): boolean {
  const store = loadStore();
  const idx = store.sessions.findIndex(s => s.id === sessionId);
  if (idx === -1) return false;
  store.sessions.splice(idx, 1);
  if (store.currentSessionId === sessionId) {
    store.currentSessionId = store.sessions.length > 0 ? store.sessions[0].id : null;
  }
  saveStore(store);
  return true;
}

export function listSessions(): Session[] {
  const store = loadStore();
  return store.sessions;
}

export function addMessage(role: "user" | "assistant" | "tool", content: string, toolCallId?: string): void {
  const session = getCurrentSessionOrCreate();
  const msg: Message = { role, content };
  if (toolCallId) msg.tool_call_id = toolCallId;
  session.messages.push(msg);
  session.updated = Date.now();
  const store = loadStore();
  const idx = store.sessions.findIndex(s => s.id === session.id);
  if (idx !== -1) {
    store.sessions[idx] = session;
    saveStore(store);
  }
}

export function getContextMessages(): Message[] {
  const session = getCurrentSession();
  if (!session) return [];
  return session.messages.filter(m => m.role !== "system");
}

export function clearSession(): void {
  const store = loadStore();
  if (!store.currentSessionId) return;
  const session = store.sessions.find(s => s.id === store.currentSessionId);
  if (session) {
    session.messages = [];
    session.totalTokens = 0;
    session.todos = [];
    session.updated = Date.now();
    saveStore(store);
  }
}

export function setSessionMode(mode: "code" | "chat"): void {
  const session = getCurrentSessionOrCreate();
  session.mode = mode;
  session.updated = Date.now();
  const store = loadStore();
  const idx = store.sessions.findIndex(s => s.id === session.id);
  if (idx !== -1) {
    store.sessions[idx] = session;
    saveStore(store);
  }
}

export function getSessionMode(): "code" | "chat" {
  const session = getCurrentSession();
  return session ? session.mode : "code";
}

export function addUsage(tokens: number): void {
  const session = getCurrentSessionOrCreate();
  session.totalTokens += tokens;
  session.updated = Date.now();
  const store = loadStore();
  const idx = store.sessions.findIndex(s => s.id === session.id);
  if (idx !== -1) {
    store.sessions[idx] = session;
    saveStore(store);
  }
}

export function getTotalTokens(): number {
  const session = getCurrentSession();
  return session ? session.totalTokens : 0;
}

// ----- Todo functions -----
export function addTodo(title: string): Todo {
  const session = getCurrentSessionOrCreate();
  const todo: Todo = {
    id: generateId(),
    title,
    status: "not started",
    created: Date.now(),
    updated: Date.now(),
  };
  session.todos.push(todo);
  session.updated = Date.now();
  const store = loadStore();
  const idx = store.sessions.findIndex(s => s.id === session.id);
  if (idx !== -1) {
    store.sessions[idx] = session;
    saveStore(store);
  }
  return todo;
}

export function updateTodoStatus(todoId: string, status: "not started" | "in progress" | "done"): Todo | null {
  const session = getCurrentSessionOrCreate();
  const todo = session.todos.find(t => t.id === todoId);
  if (!todo) return null;
  todo.status = status;
  todo.updated = Date.now();
  session.updated = Date.now();
  const store = loadStore();
  const idx = store.sessions.findIndex(s => s.id === session.id);
  if (idx !== -1) {
    store.sessions[idx] = session;
    saveStore(store);
  }
  return todo;
}

export function getTodos(): Todo[] {
  const session = getCurrentSession();
  return session ? session.todos : [];
}

export function deleteTodo(todoId: string): boolean {
  const session = getCurrentSessionOrCreate();
  const initialLength = session.todos.length;
  session.todos = session.todos.filter(t => t.id !== todoId);
  if (session.todos.length === initialLength) return false;
  session.updated = Date.now();
  const store = loadStore();
  const idx = store.sessions.findIndex(s => s.id === session.id);
  if (idx !== -1) {
    store.sessions[idx] = session;
    saveStore(store);
  }
  return true;
}

// ----- GitHub config -----
const GITHUB_CONFIG_FILE = path.join(CONFIG_DIR, "github_config.json");

export function loadGitHubConfig(): GitHubConfig {
  ensureConfigDir();
  if (fs.existsSync(GITHUB_CONFIG_FILE)) {
    try {
      const data = fs.readFileSync(GITHUB_CONFIG_FILE, "utf-8");
      return JSON.parse(data);
    } catch (e) {
      // если файл повреждён, возвращаем пустой
    }
  }
  return { configured: false };
}

export function saveGitHubConfig(config: GitHubConfig) {
  ensureConfigDir();
  fs.writeFileSync(GITHUB_CONFIG_FILE, JSON.stringify(config, null, 2), "utf-8");
}

export function isGitHubConfigured(): boolean {
  const config = loadGitHubConfig();
  return config.configured && !!config.repository && !!config.github_token;
}

export function setGitHubConfig(repository?: string, github_token?: string): GitHubConfig {
  const config = loadGitHubConfig();
  if (repository !== undefined) config.repository = repository;
  if (github_token !== undefined) config.github_token = github_token;
  config.configured = true;
  saveGitHubConfig(config);
  return config;
}