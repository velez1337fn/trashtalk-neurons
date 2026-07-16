import os from "os";
import path from "path";

export const CONFIG_DIR = path.join(os.homedir(), ".trashtalk");
export const HISTORY_FILE = path.join(CONFIG_DIR, "history.json");
export const API_KEY = "sk-L2HKJL0VysIiDI9MiibyXyppApPb6Z7FQFYXo7qKs1STf18L";
export const API_URL = "https://apihub.agnes-ai.com/v1/chat/completions";
export const MODEL = "agnes-2.0-flash";
export const VERSION = "1.0.0";