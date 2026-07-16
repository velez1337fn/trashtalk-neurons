import fs from "fs";
import path from "path";
import { CONFIG_DIR } from "./config";

export interface GitHubConfig {
  repository: string;
  github_token: string;
}

const CONFIG_FILE = path.join(CONFIG_DIR, "config.json");

function ensureConfigDir() {
  if (!fs.existsSync(CONFIG_DIR)) {
    fs.mkdirSync(CONFIG_DIR, { recursive: true });
  }
}

function loadConfig(): GitHubConfig | null {
  ensureConfigDir();
  if (fs.existsSync(CONFIG_FILE)) {
    try {
      const data = fs.readFileSync(CONFIG_FILE, "utf-8");
      return JSON.parse(data);
    } catch (e) {
      return null;
    }
  }
  return null;
}

function saveConfig(config: GitHubConfig): void {
  ensureConfigDir();
  fs.writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2), "utf-8");
}

export function getGitHubConfig(): GitHubConfig | null {
  return loadConfig();
}

export function setGitHubConfig(repository: string, github_token: string): void {
  saveConfig({ repository, github_token });
}

export function updateGitHubConfig(repository?: string, github_token?: string): void {
  const existing = loadConfig();
  const updated = {
    repository: repository || (existing ? existing.repository : ""),
    github_token: github_token || (existing ? existing.github_token : ""),
  };
  if (!updated.repository || !updated.github_token) {
    throw new Error("Both repository and github_token must be set.");
  }
  saveConfig(updated);
}   