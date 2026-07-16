#!/usr/bin/env node

import blessed from "blessed";
import axios from "axios";
import clipboardy from "clipboardy";
import { exec } from "child_process";
import fs from "fs/promises";
import path from "path";
import {
  getCurrentSession,
  createSession,
  switchSession,
  deleteSession,
  listSessions,
  addMessage,
  getContextMessages,
  clearSession,
  setSessionMode,
  getSessionMode,
  addUsage,
  getTotalTokens,
  addTodo,
  updateTodoStatus,
  getTodos,
  deleteTodo,
  loadGitHubConfig,
  saveGitHubConfig,
  isGitHubConfigured,
  setGitHubConfig,
} from "./history";
import { API_KEY, API_URL, MODEL, VERSION } from "./config";

const colors = {
  cyan: "\x1b[36m",
  green: "\x1b[32m",
  yellow: "\x1b[33m",
  magenta: "\x1b[35m",
  red: "\x1b[31m",
  grey: "\x1b[90m",
  bright: "\x1b[1m",
  reset: "\x1b[0m",
};

const ASCII_ART = `
  ████████╗██████╗  █████╗ ███████╗██╗  ██╗███╗   ██╗███████╗██╗   ██╗██████╗  ██████╗ ███╗   ██╗███████╗
  ╚══██╔══╝██╔══██╗██╔══██╗██╔════╝██║  ██║████╗  ██║██╔════╝██║   ██║██╔══██╗██╔═══██╗████╗  ██║██╔════╝
     ██║   ██████╔╝███████║███████╗███████║██╔██╗ ██║█████╗  ██║   ██║██████╔╝██║   ██║██╔██╗ ██║█████╗  
     ██║   ██╔══██╗██╔══██║╚════██║██╔══██║██║╚██╗██║██╔══╝  ██║   ██║██╔══██╗██║   ██║██║╚██╗██║██╔══╝  
     ██║   ██║  ██║██║  ██║███████║██║  ██║██║ ╚████║███████╗╚██████╔╝██║  ██║╚██████╔╝██║ ╚████║███████╗
     ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚══════╝
`;

const screen = blessed.screen({
  smartCSR: true,
  title: "TrashNeurons CLI",
  fullUnicode: true,
});

let currentCancelToken: any = null;
let exitAttemptTimeout: NodeJS.Timeout | null = null;
let exitAttemptCount = 0;

const artBox = blessed.box({
  top: 0,
  left: "center",
  width: "100%",
  height: 7,
  content: ASCII_ART,
  style: { fg: "green", bold: true },
  align: "center",
});

let mode = getSessionMode();

const infoBox = blessed.box({
  top: 7,
  left: 0,
  width: "100%",
  height: 3,
  content: "",
  style: { fg: "white" },
  align: "center",
  tags: false,
  ansi: true,
});

const outputBox = blessed.box({
  top: 10,
  left: 0,
  width: "100%",
  height: "100%-16",
  scrollable: true,
  alwaysScroll: true,
  scrollbar: { ch: " ", track: { bg: "grey" }, style: { bg: "white" } },
  style: { fg: "white" },
  content: "",
  tags: false,
  ansi: true,
});

const statusLine = blessed.box({
  bottom: 3,
  left: 0,
  width: "100%",
  height: 1,
  content: "",
  style: { fg: "cyan" },
  tags: false,
  ansi: true,
});

let inputValue = "";
let cursorPos = 0;
let displayPlaceholder = false;
let inputBox: any;
let inputHeight = 3;
let outputLines: string[] = [];

let isConfigPrompt = false;
let configPromptState: "initial" | "repo" | "token" = "initial";
let configAnswers: { repository?: string; github_token?: string } = {};

function renderOutput() {
  outputBox.setContent(outputLines.join("\n"));
  outputBox.scrollTo(outputBox.getScrollHeight());
  screen.render();
}

function addLine(text: string) {
  outputLines.push(text);
  renderOutput();
}

function updateLastLine(text: string) {
  if (outputLines.length > 0) {
    outputLines[outputLines.length - 1] = text;
  } else {
    outputLines.push(text);
  }
  renderOutput();
}

function updateInfoLine() {
  const session = getCurrentSession();
  const msgCount = session ? session.messages.length : 0;
  const sessionId = session ? session.id : "none";
  const modeLabel = mode === "code" ? "Code" : "Chat";
  const modelLabel = mode === "code" ? "trashneurons-1.1-code" : "trashneurons-1.3-chat-reasoner";
  const totalTokens = getTotalTokens() || 0;
  const maxContext = 512 * 1024;
  const percent = totalTokens > 0 ? Math.min(100, Math.round((totalTokens / maxContext) * 100)) : 0;
  const contextStr = `${totalTokens} / 512K tokens (${percent}%)`;
  const contextColor = percent > 80 ? colors.red : percent > 50 ? colors.yellow : colors.green;
  const todos = getTodos();
  const activeTodos = todos.filter(t => t.status !== "done").length;
  const todoStr = activeTodos > 0 ? `${colors.yellow}●${colors.reset} ${activeTodos}` : `${colors.green}✓${colors.reset} 0`;
  const config = loadGitHubConfig();
  const gitStatus = config.repository ? "✓" : "✗";

  const line1 = `${colors.cyan}Session: ${colors.bright}${sessionId}${colors.reset}` +
    ` ${colors.grey}│${colors.reset} ${colors.yellow}Mode: ${colors.bright}${modeLabel}${colors.reset}` +
    ` ${colors.grey}│${colors.reset} ${colors.magenta}Model: ${colors.bright}${modelLabel}${colors.reset}` +
    ` ${colors.grey}│${colors.reset} ${colors.green}Msg: ${colors.bright}${msgCount}${colors.reset}` +
    ` ${colors.grey}│${colors.reset} ${colors.cyan}Todo: ${colors.bright}${todoStr}${colors.reset}` +
    ` ${colors.grey}│${colors.reset} ${colors.magenta}Git: ${colors.bright}${gitStatus}${colors.reset}`;

  const line2 = `${colors.grey}Context: ${contextColor}${contextStr}${colors.reset}`;

  infoBox.setContent(line1 + "\n" + line2);
  screen.render();
}

function updateStatus(text: string) {
  statusLine.setContent(text);
  screen.render();
}

function refreshUI() {
  updateInfoLine();
  renderOutput();
}

function getCursorPosition(pos: number, text: string): { x: number; y: number } {
  const lines = text.split("\n");
  let remaining = pos;
  for (let i = 0; i < lines.length; i++) {
    const lineLen = lines[i].length;
    if (remaining <= lineLen) {
      return { x: remaining, y: i };
    }
    remaining -= lineLen + 1;
    if (remaining < 0) return { x: 0, y: i };
  }
  const lastLine = lines.length - 1;
  return { x: lines[lastLine]?.length || 0, y: lastLine };
}

function updateInput() {
  const lines = inputValue.split("\n");
  let newHeight = lines.length + 1;
  if (newHeight < 3) newHeight = 3;
  if (newHeight > 10) newHeight = 10;
  if (inputBox) {
    inputBox.height = newHeight;
    inputBox.setContent(displayPlaceholder ? "[pasted 10+ lines]" : inputValue);
    screen.render();

    if (!displayPlaceholder) {
      const pos = getCursorPosition(cursorPos, inputValue);
      const absX = inputBox.aleft + 1 + pos.x;
      const absY = inputBox.atop + 1 + pos.y;
      screen.program.cup(absY, absX);
    } else {
      const absX = inputBox.aleft + 1;
      const absY = inputBox.atop + 1;
      screen.program.cup(absY, absX);
    }
  }
}

function createInputBox() {
  inputBox = blessed.box({
    bottom: 0,
    left: 0,
    width: "100%",
    height: inputHeight,
    content: "",
    style: { fg: "white", bg: "black", border: { fg: "white" } },
    border: { type: "line" },
    tags: false,
    input: true,
    keys: true,
    mouse: true,
    focusable: true,
  });
  screen.append(inputBox);
  inputBox.focus();
  updateInput();
}

function getLineCount(str: string): number {
  return str.split("\n").length;
}

function handleCommand(cmd: string): boolean {
  const parts = cmd.trim().split(/\s+/);
  const command = parts[0].toLowerCase();
  const knownCommands = ["/clear", "/history", "/sessions", "/new", "/load", "/delete", "/help", "/todos", "/config"];
  if (!knownCommands.includes(command)) return false;

  if (command === "/clear") {
    clearSession();
    outputLines = [];
    renderOutput();
    updateStatus(colors.green + "Session cleared." + colors.reset);
    refreshUI();
    return true;
  }
  if (command === "/history") {
    const session = getCurrentSession();
    if (!session) {
      updateStatus(colors.red + "No active session." + colors.reset);
      return true;
    }
    const history = session.messages.map(m => `[${m.role}] ${m.content}`).join("\n");
    outputLines = history.split("\n");
    renderOutput();
    return true;
  }
  if (command === "/sessions") {
    const sessions = listSessions();
    const current = getCurrentSession();
    const list = sessions.map(s => {
      const marker = s.id === current?.id ? " * " : "   ";
      return `${marker}${s.id} (${s.mode}) - ${new Date(s.updated).toLocaleString()}`;
    }).join("\n");
    outputLines = list.split("\n");
    renderOutput();
    return true;
  }
  if (command === "/new") {
    const newMode = parts[1] === "chat" ? "chat" : "code";
    const session = createSession(newMode);
    mode = newMode;
    setSessionMode(newMode);
    outputLines = [];
    addLine(colors.green + "New " + mode + " session created: " + session.id + colors.reset);
    refreshUI();
    updateStatus(colors.green + "Session " + session.id + " active." + colors.reset);
    return true;
  }
  if (command === "/load" && parts.length > 1) {
    const sessionId = parts[1];
    const session = switchSession(sessionId);
    if (session) {
      mode = session.mode;
      setSessionMode(mode);
      outputLines = [];
      addLine(colors.green + "Loaded session " + sessionId + colors.reset);
      refreshUI();
      updateStatus(colors.green + "Session " + sessionId + " loaded." + colors.reset);
    } else {
      updateStatus(colors.red + "Session " + sessionId + " not found." + colors.reset);
    }
    return true;
  }
  if (command === "/delete" && parts.length > 1) {
    const sessionId = parts[1];
    if (deleteSession(sessionId)) {
      outputLines = [];
      addLine(colors.green + "Session " + sessionId + " deleted." + colors.reset);
      const current = getCurrentSession();
      if (current) {
        mode = current.mode;
        setSessionMode(mode);
      }
      refreshUI();
      updateStatus(colors.green + "Session " + sessionId + " deleted." + colors.reset);
    } else {
      updateStatus(colors.red + "Session " + sessionId + " not found." + colors.reset);
    }
    return true;
  }
  if (command === "/todos") {
    const todos = getTodos();
    if (todos.length === 0) {
      outputLines = ["No todos."];
      renderOutput();
      return true;
    }
    const lines = todos.map(t => {
      const icon = t.status === "done" ? "✅" : t.status === "in progress" ? "🔄" : "⏳";
      return `${icon} ${t.title} (${t.status})`;
    });
    outputLines = lines;
    renderOutput();
    return true;
  }
  if (command === "/config") {
    const config = loadGitHubConfig();
    if (parts.length === 1) {
      const lines = [
        "GitHub Configuration:",
        `  Repository: ${config.repository || "(not set)"}`,
        `  Token: ${config.github_token ? "***" : "(not set)"}`,
        `  Configured: ${config.configured ? "Yes" : "No"}`,
      ];
      outputLines = lines;
      renderOutput();
      return true;
    }
    if (parts.length >= 3 && parts[1] === "set") {
      const key = parts[2];
      const value = parts.slice(3).join(" ");
      if (key === "repository") {
        setGitHubConfig(value, undefined);
        updateStatus(colors.green + "Repository updated." + colors.reset);
        refreshUI();
        return true;
      } else if (key === "github_token") {
        setGitHubConfig(undefined, value);
        updateStatus(colors.green + "Token updated." + colors.reset);
        refreshUI();
        return true;
      } else {
        updateStatus(colors.red + "Invalid key. Use repository or github_token." + colors.reset);
        return true;
      }
    }
    updateStatus(colors.red + "Usage: /config or /config set repository <url> or /config set github_token <token>" + colors.reset);
    return true;
  }
  if (command === "/help") {
    const help = `
${colors.bright}Commands:${colors.reset}
  /clear          Clear current session history
  /history        Show current session history
  /sessions       List all sessions
  /new [code|chat] Create new session
  /load <id>      Load session by ID
  /delete <id>    Delete session by ID
  /todos          Show current todo list
  /config         Show GitHub configuration
  /config set repository <url>  Set GitHub repository
  /config set github_token <token>  Set GitHub token
  /help           Show this help
  ${colors.yellow}Tab${colors.reset}             Switch mode (code/chat)
  ${colors.yellow}Esc${colors.reset}             Cancel ongoing request (or press twice to exit)
  ${colors.yellow}Ctrl+C${colors.reset}          Same as Esc (press twice to exit)
`;
    outputLines = help.split("\n");
    renderOutput();
    return true;
  }
  return false;
}

function handleKey(ch: any, key: any) {
  if (isConfigPrompt) {
    if (key.name === "enter") {
      const answer = inputValue.trim();
      inputValue = "";
      cursorPos = 0;
      updateInput();

      if (configPromptState === "initial") {
        if (answer.toLowerCase() === "y" || answer.toLowerCase() === "yes") {
          configPromptState = "repo";
          updateStatus(colors.cyan + "Enter repository URL (e.g., https://github.com/user/repo): " + colors.reset);
          return;
        } else {
          setGitHubConfig("", "");
          isConfigPrompt = false;
          updateStatus(colors.green + "GitHub configuration skipped." + colors.reset);
          refreshUI();
          updateStatus(colors.grey + "Press " + colors.yellow + "Tab" + colors.reset + colors.grey + " to switch mode, " + colors.yellow + "/help" + colors.reset + colors.grey + " for commands" + colors.reset);
          return;
        }
      } else if (configPromptState === "repo") {
        if (answer) {
          configAnswers.repository = answer;
          configPromptState = "token";
          updateStatus(colors.cyan + "Enter GitHub access token: " + colors.reset);
        } else {
          updateStatus(colors.red + "Repository URL cannot be empty. Try again: " + colors.reset);
        }
        return;
      } else if (configPromptState === "token") {
        if (answer) {
          configAnswers.github_token = answer;
          setGitHubConfig(configAnswers.repository, configAnswers.github_token);
          isConfigPrompt = false;
          updateStatus(colors.green + "GitHub configuration saved!" + colors.reset);
          refreshUI();
          updateStatus(colors.grey + "Press " + colors.yellow + "Tab" + colors.reset + colors.grey + " to switch mode, " + colors.yellow + "/help" + colors.reset + colors.grey + " for commands" + colors.reset);
        } else {
          setGitHubConfig(configAnswers.repository, "");
          isConfigPrompt = false;
          updateStatus(colors.yellow + "GitHub token not set. You can set it later with /config set github_token <token>" + colors.reset);
          refreshUI();
          updateStatus(colors.grey + "Press " + colors.yellow + "Tab" + colors.reset + colors.grey + " to switch mode, " + colors.yellow + "/help" + colors.reset + colors.grey + " for commands" + colors.reset);
        }
        return;
      }
      return;
    }
    if (ch && typeof ch === "string" && ch.length > 0 && ch.charCodeAt(0) >= 32) {
      const before = inputValue.slice(0, cursorPos);
      const after = inputValue.slice(cursorPos);
      inputValue = before + ch + after;
      cursorPos++;
      updateInput();
      return;
    }
    if (key.name === "backspace") {
      if (cursorPos > 0) {
        const before = inputValue.slice(0, cursorPos - 1);
        const after = inputValue.slice(cursorPos);
        inputValue = before + after;
        cursorPos--;
        updateInput();
      }
      return;
    }
    if (key.name === "delete") {
      if (cursorPos < inputValue.length) {
        const before = inputValue.slice(0, cursorPos);
        const after = inputValue.slice(cursorPos + 1);
        inputValue = before + after;
        updateInput();
      }
      return;
    }
    if (key.name === "space") {
      const before = inputValue.slice(0, cursorPos);
      const after = inputValue.slice(cursorPos);
      inputValue = before + " " + after;
      cursorPos++;
      updateInput();
      return;
    }
    return;
  }

  if (key.name === "escape" || (key.ctrl && key.name === "c")) {
    if (currentCancelToken) {
      currentCancelToken.cancel("Request cancelled by user");
      currentCancelToken = null;
      updateStatus(colors.yellow + "✖ Request cancelled." + colors.reset);
      addLine(colors.yellow + "─── Request cancelled by user ───" + colors.reset);
      screen.render();
      return;
    } else {
      handleExitAttempt();
      return;
    }
  }

  if (key.name === "tab") {
    mode = mode === "code" ? "chat" : "code";
    setSessionMode(mode);
    refreshUI();
    return;
  }
  if (key.name === "enter") {
    if (key.shift) {
      if (displayPlaceholder) displayPlaceholder = false;
      const before = inputValue.slice(0, cursorPos);
      const after = inputValue.slice(cursorPos);
      inputValue = before + "\n" + after;
      cursorPos++;
      updateInput();
      return;
    } else {
      const prompt = inputValue.trim();
      if (prompt.length === 0) return;
      inputValue = "";
      cursorPos = 0;
      displayPlaceholder = false;
      updateInput();

      if (prompt.startsWith("/")) {
        const isCommand = handleCommand(prompt);
        if (isCommand) return;
      }
      submitPrompt(prompt);
      return;
    }
  }
  if (key.name === "backspace") {
    if (displayPlaceholder) displayPlaceholder = false;
    if (cursorPos > 0) {
      const before = inputValue.slice(0, cursorPos - 1);
      const after = inputValue.slice(cursorPos);
      inputValue = before + after;
      cursorPos--;
    }
    updateInput();
    return;
  }
  if (key.name === "delete") {
    if (displayPlaceholder) displayPlaceholder = false;
    if (cursorPos < inputValue.length) {
      const before = inputValue.slice(0, cursorPos);
      const after = inputValue.slice(cursorPos + 1);
      inputValue = before + after;
    }
    updateInput();
    return;
  }
  if (key.name === "left") {
    if (cursorPos > 0) cursorPos--;
    updateInput();
    return;
  }
  if (key.name === "right") {
    if (cursorPos < inputValue.length) cursorPos++;
    updateInput();
    return;
  }
  if (key.name === "home") {
    cursorPos = 0;
    updateInput();
    return;
  }
  if (key.name === "end") {
    cursorPos = inputValue.length;
    updateInput();
    return;
  }
  if (key.ctrl && key.name === "v") {
    const pasted = clipboardy.readSync();
    if (pasted) {
      const before = inputValue.slice(0, cursorPos);
      const after = inputValue.slice(cursorPos);
      inputValue = before + pasted + after;
      cursorPos += pasted.length;
      displayPlaceholder = getLineCount(inputValue) > 10;
      updateInput();
    }
    return;
  }
  if (ch && typeof ch === "string" && ch.length > 0 && ch.charCodeAt(0) >= 32) {
    if (displayPlaceholder) displayPlaceholder = false;
    const before = inputValue.slice(0, cursorPos);
    const after = inputValue.slice(cursorPos);
    inputValue = before + ch + after;
    cursorPos++;
    displayPlaceholder = getLineCount(inputValue) > 10;
    updateInput();
    return;
  }
  if (key.name === "space") {
    if (displayPlaceholder) displayPlaceholder = false;
    const before = inputValue.slice(0, cursorPos);
    const after = inputValue.slice(cursorPos);
    inputValue = before + " " + after;
    cursorPos++;
    displayPlaceholder = getLineCount(inputValue) > 10;
    updateInput();
    return;
  }
  if (key.name && key.name.length === 1 && !key.ctrl && !key.meta) {
    if (displayPlaceholder) displayPlaceholder = false;
    const before = inputValue.slice(0, cursorPos);
    const after = inputValue.slice(cursorPos);
    inputValue = before + key.name + after;
    cursorPos++;
    displayPlaceholder = getLineCount(inputValue) > 10;
    updateInput();
    return;
  }
}

function handleExitAttempt() {
  if (exitAttemptTimeout) {
    clearTimeout(exitAttemptTimeout);
    exitAttemptTimeout = null;
  }
  exitAttemptCount++;
  if (exitAttemptCount >= 2) {
    process.exit(0);
  } else {
    const msg = colors.yellow + "Press again to exit..." + colors.reset;
    updateStatus(msg);
    exitAttemptTimeout = setTimeout(() => {
      exitAttemptCount = 0;
      exitAttemptTimeout = null;
      updateStatus(colors.grey + "Press " + colors.yellow + "Esc" + colors.reset + colors.grey + " or " + colors.yellow + "Ctrl+C" + colors.reset + colors.grey + " to exit" + colors.reset);
    }, 2000);
  }
}

// ---------- Tools ----------
const tools = [
  {
    type: "function",
    function: {
      name: "execute_command",
      description: "Execute a shell command and return its output.",
      parameters: {
        type: "object",
        properties: {
          command: { type: "string", description: "The shell command to execute." },
          cwd: { type: "string", description: "Working directory (optional)." },
        },
        required: ["command"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "read_file",
      description: "Read the contents of a file.",
      parameters: {
        type: "object",
        properties: {
          path: { type: "string", description: "Path to the file." },
        },
        required: ["path"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "write_file",
      description: "Write content to a file (overwrites existing).",
      parameters: {
        type: "object",
        properties: {
          path: { type: "string", description: "Path to the file." },
          content: { type: "string", description: "Content to write." },
        },
        required: ["path", "content"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "list_directory",
      description: "List files and directories in a given path.",
      parameters: {
        type: "object",
        properties: {
          path: { type: "string", description: "Directory path." },
        },
        required: ["path"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "manage_todo",
      description: "Manage todo tasks in the current session.",
      parameters: {
        type: "object",
        properties: {
          action: {
            type: "string",
            enum: ["create", "update", "list", "delete"],
            description: "Action to perform: create a new todo, update status, list all, or delete.",
          },
          title: {
            type: "string",
            description: "Todo title (required for create).",
          },
          todo_id: {
            type: "string",
            description: "Todo ID (required for update and delete).",
          },
          status: {
            type: "string",
            enum: ["not started", "in progress", "done"],
            description: "New status (required for update).",
          },
        },
        required: ["action"],
      },
    },
  },
  {
    type: "function",
    function: {
      name: "git_push",
      description: "Push the current repository to GitHub using stored credentials.",
      parameters: {
        type: "object",
        properties: {
          message: {
            type: "string",
            description: "Commit message (optional).",
          },
        },
        required: [],
      },
    },
  },
];

// Реализация git_push через exec
function execPromise(command: string, cwd?: string): Promise<string> {
  return new Promise((resolve, reject) => {
    exec(command, { cwd: cwd || process.cwd(), timeout: 60000 }, (error, stdout, stderr) => {
      if (error) {
        reject(new Error(stderr || error.message));
      } else {
        resolve(stdout || stderr);
      }
    });
  });
}

async function gitPush(message?: string): Promise<string> {
  const config = loadGitHubConfig();
  if (!config.repository || !config.github_token) {
    return "GitHub repository or token not configured. Please set with /config set repository <url> and /config set github_token <token>";
  }

  try {
    // Проверяем, находимся ли в репозитории
    await execPromise("git rev-parse --is-inside-work-tree", process.cwd());

    // Проверяем, есть ли изменения
    const status = await execPromise("git status --porcelain", process.cwd());
    if (!status.trim()) {
      return "No changes to commit.";
    }

    // Добавляем все изменения
    await execPromise("git add .", process.cwd());

    // Коммитим
    const commitMsg = message || `Auto-commit by TrashNeurons at ${new Date().toISOString()}`;
    await execPromise(`git commit -m "${commitMsg.replace(/"/g, '\\"')}"`, process.cwd());

    // Получаем текущую ветку
    const branch = (await execPromise("git rev-parse --abbrev-ref HEAD", process.cwd())).trim();

    // Устанавливаем remote URL с токеном
    const repoUrl = config.repository.replace(/^https:\/\//, `https://${config.github_token}@`);
    await execPromise(`git remote set-url origin ${repoUrl}`, process.cwd());

    // Пушим
    await execPromise(`git push origin ${branch}`, process.cwd());

    return `Successfully pushed to ${config.repository} (branch: ${branch}).`;
  } catch (err: any) {
    return `Git push failed: ${err.message}`;
  }
}

async function executeToolCall(toolCall: any): Promise<string> {
  const name = toolCall.function.name;
  const args = JSON.parse(toolCall.function.arguments);
  let result = "";

  switch (name) {
    case "execute_command": {
      const { command, cwd } = args;
      try {
        const out = await execPromise(command, cwd || process.cwd());
        return out;
      } catch (err: any) {
        return `Error: ${err.message}`;
      }
    }
    case "read_file": {
      try {
        const content = await fs.readFile(args.path, "utf-8");
        return content;
      } catch (err: any) {
        return `Error reading file: ${err.message}`;
      }
    }
    case "write_file": {
      try {
        await fs.writeFile(args.path, args.content, "utf-8");
        return `File written successfully: ${args.path}`;
      } catch (err: any) {
        return `Error writing file: ${err.message}`;
      }
    }
    case "list_directory": {
      try {
        const items = await fs.readdir(args.path, { withFileTypes: true });
        const list = items.map(item => {
          const isDir = item.isDirectory();
          return (isDir ? "[DIR] " : "[FILE]") + item.name;
        }).join("\n");
        return list || "Directory is empty.";
      } catch (err: any) {
        return `Error listing directory: ${err.message}`;
      }
    }
    case "manage_todo": {
      const { action, title, todo_id, status } = args;
      switch (action) {
        case "create": {
          if (!title) return "Error: title is required for create action.";
          const todo = addTodo(title);
          return `Todo created: ${todo.id} - "${todo.title}" (status: ${todo.status})`;
        }
        case "update": {
          if (!todo_id) return "Error: todo_id is required for update action.";
          if (!status) return "Error: status is required for update action.";
          const updated = updateTodoStatus(todo_id, status);
          if (!updated) return `Todo with ID ${todo_id} not found.`;
          return `Todo updated: ${updated.id} - "${updated.title}" (status: ${updated.status})`;
        }
        case "list": {
          const todos = getTodos();
          if (todos.length === 0) return "No todos.";
          return todos.map(t => `${t.id}: "${t.title}" (${t.status})`).join("\n");
        }
        case "delete": {
          if (!todo_id) return "Error: todo_id is required for delete action.";
          const success = deleteTodo(todo_id);
          return success ? `Todo ${todo_id} deleted.` : `Todo ${todo_id} not found.`;
        }
        default:
          return `Unknown action: ${action}`;
      }
    }
    case "git_push": {
      const msg = args.message || undefined;
      return await gitPush(msg);
    }
    default:
      return `Unknown tool: ${name}`;
  }
}

function typewriterEffect(text: string, callback: () => void) {
  let index = 0;
  const prefix = colors.green + "[Assistant] " + colors.reset;
  addLine(prefix);
  let accumulated = "";
  const interval = setInterval(() => {
    if (index < text.length) {
      accumulated += text[index];
      const line = prefix + accumulated;
      updateLastLine(line);
      index++;
    } else {
      clearInterval(interval);
      const finalLine = prefix + text;
      updateLastLine(finalLine);
      callback();
    }
  }, 10);
}

async function processWithTools(messages: any[], cancelToken: any, temperature: number): Promise<{ content: string; totalTokens: number }> {
  let finished = false;
  let finalContent = "";
  let totalTokens = 0;

  while (!finished) {
    if (cancelToken && cancelToken.token.reason) {
      throw new axios.Cancel("Request cancelled");
    }

    const requestBody: any = {
      model: MODEL,
      messages,
      stream: false,
      temperature: temperature,
      tools: tools,
      tool_choice: "auto",
    };

    if (mode === "chat") {
      requestBody.thinking = { type: "enabled", budget_tokens: 2048 };
    }

    try {
      const response = await axios.post(API_URL, requestBody, {
        headers: {
          Authorization: `Bearer ${API_KEY}`,
          "Content-Type": "application/json",
        },
        cancelToken: cancelToken ? cancelToken.token : undefined,
      });

      const choice = response.data.choices[0];
      const message = choice.message;
      if (response.data.usage) {
        totalTokens += response.data.usage.total_tokens || 0;
      }

      if (message.tool_calls && message.tool_calls.length > 0) {
        messages.push({
          role: "assistant",
          content: message.content || null,
          tool_calls: message.tool_calls,
        });

        if (message.content) {
          addLine(colors.yellow + "[Assistant] " + colors.reset + message.content);
        }

        for (const toolCall of message.tool_calls) {
          if (cancelToken && cancelToken.token.reason) {
            throw new axios.Cancel("Request cancelled");
          }
          const result = await executeToolCall(toolCall);
          messages.push({
            role: "tool",
            tool_call_id: toolCall.id,
            content: result,
          });
          const toolName = toolCall.function.name;
          addLine(colors.magenta + "[Tool] " + colors.reset + toolName + " called.");
          if (result.length < 200) {
            addLine(colors.cyan + "[Result] " + colors.reset + result);
          } else {
            addLine(colors.cyan + "[Result] " + colors.reset + result.slice(0, 200) + "... (truncated)");
          }
        }
      } else {
        finalContent = message.content || "";
        messages.push({
          role: "assistant",
          content: finalContent,
        });
        finished = true;
      }
    } catch (err: any) {
      if (axios.isCancel(err)) {
        throw err;
      }
      finalContent = "Error: try again!";
      finished = true;
    }
  }
  return { content: finalContent, totalTokens };
}

function submitPrompt(prompt: string) {
  const isChatMode = mode === "chat";
  const temperature = isChatMode ? 0.7 : 0.3;

  const config = loadGitHubConfig();
  const gitInfo = config.repository ? `\nGitHub repository: ${config.repository}\nGitHub token: ${config.github_token ? "set" : "not set"}` : "\nGitHub not configured.";

  const systemPrompt = `You are TrashNeurons, an advanced AI assistant with full system access. Your capabilities include:
- Executing shell commands (execute_command)
- Reading files (read_file)
- Writing files (write_file)
- Listing directories (list_directory)
- Managing todo tasks (manage_todo)
- Git push (git_push) – use it to push changes to GitHub. This tool works automatically: it adds, commits, and pushes all changes using the configured repository and token.

When a user gives you a complex task, automatically break it down into actionable steps and create a todo list. Update statuses as you progress.

Always be concise and helpful. If the request is ambiguous, ask clarifying questions.

Current working directory: ${process.cwd()}
${gitInfo}

User's query: "${prompt}"`;

  const messages = [
    { role: "system", content: systemPrompt },
    ...getContextMessages(),
    { role: "user", content: prompt },
  ];

  addLine(colors.cyan + "[User] " + colors.reset + prompt);
  addLine(colors.grey + "──────────────────────────────────────────────────" + colors.reset);

  const startTime = Date.now();
  let timerInterval: NodeJS.Timeout | null = null;
  let thinkingDone = false;
  let spinChars = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
  let spinIndex = 0;

  function updateThinkingTimer() {
    if (!thinkingDone) {
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
      const spinner = spinChars[spinIndex % spinChars.length];
      spinIndex++;
      updateStatus(colors.yellow + spinner + colors.reset + " thinking... " + elapsed + "s  " + colors.grey + "(press Esc to interrupt)" + colors.reset);
    }
  }

  timerInterval = setInterval(updateThinkingTimer, 100);
  updateStatus("thinking... 0.0s  (press Esc to interrupt)");

  addMessage("user", prompt);

  const cancelTokenSource = axios.CancelToken.source();
  currentCancelToken = cancelTokenSource;

  processWithTools(messages, cancelTokenSource, temperature)
    .then((result) => {
      if (timerInterval) {
        clearInterval(timerInterval);
        timerInterval = null;
      }
      thinkingDone = true;
      const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
      updateStatus(colors.green + "✓" + colors.reset + " thinked for " + elapsed + " seconds!");

      if (result.totalTokens > 0) {
        addUsage(result.totalTokens);
        refreshUI();
      }

      if (result.content) {
        typewriterEffect(result.content, () => {
          addLine(colors.grey + "──────────────────────────────────────────────────" + colors.reset);
          refreshUI();
        });
      } else {
        addLine(colors.grey + "──────────────────────────────────────────────────" + colors.reset);
        refreshUI();
      }
      currentCancelToken = null;
    })
    .catch((err: any) => {
      if (timerInterval) {
        clearInterval(timerInterval);
        timerInterval = null;
      }
      if (axios.isCancel(err)) {
        currentCancelToken = null;
        return;
      }
      addLine(colors.red + "Error: try again!" + colors.reset);
      addLine(colors.grey + "──────────────────────────────────────────────────" + colors.reset);
      updateStatus(colors.red + "Error occurred." + colors.reset);
      screen.render();
      currentCancelToken = null;
    });
}

function checkGitHubConfigAndPrompt() {
  const config = loadGitHubConfig();
  if (!config.configured) {
    isConfigPrompt = true;
    configPromptState = "initial";
    updateStatus(colors.yellow + "GitHub not configured. Configure now? (y/n): " + colors.reset);
  } else {
    updateStatus(colors.grey + "Press " + colors.yellow + "Tab" + colors.reset + colors.grey + " to switch mode, " + colors.yellow + "/help" + colors.reset + colors.grey + " for commands" + colors.reset);
  }
}

screen.append(artBox);
screen.append(infoBox);
screen.append(outputBox);
screen.append(statusLine);
createInputBox();

inputBox.on("keypress", handleKey);

const existing = getCurrentSession();
if (!existing) {
  createSession("code");
} else {
  mode = existing.mode;
  setSessionMode(mode);
}

refreshUI();
checkGitHubConfigAndPrompt();

screen.render();
inputBox.focus();