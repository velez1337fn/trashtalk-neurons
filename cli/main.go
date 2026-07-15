package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	apiKey     = "sk-L2HKJL0VysIiDI9MiibyXyppApPb6Z7FQFYXo7qKs1STf18L"
	apiBaseURL = "https://apihub.agnes-ai.com/v1"
	modelName  = "agnes-2.0-flash"
	dataDir    = ".trashneurons"
	chatsDir   = "chats"
)

type Mode int

const (
	ModeCode Mode = iota
	ModeChat
)

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolParameter struct {
	Type       string            `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string          `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolFunction struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Parameters  ToolParameter `json:"parameters"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function ToolFnCall   `json:"function"`
}

type ToolFnCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Choice struct {
	Index        int      `json:"index"`
	Message      Message  `json:"message"`
	FinishReason string   `json:"finish_reason"`
}

type ChatCompletionResponse struct {
	Choices []Choice `json:"choices"`
}

type ChatSession struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Mode      Mode      `json:"mode"`
	CreatedAt time.Time `json:"created_at"`
	Messages  []Message `json:"messages"`
}

type App struct {
	pages         *tview.Pages
	sidePanel     *tview.Frame
	modeText      *tview.TextView
	arrowUp       *tview.TextView
	arrowDown     *tview.TextView
	separator     *tview.TextView
	rightPanel    *tview.Flex
	mainInput     *tview.InputField
	mainInputBox  *tview.Box
	shipLogo      *tview.TextView
	animationLbl  *tview.TextView
	chatTopBox    *tview.Box
	aiAnswerLbl   *tview.TextView
	thinkingLbl   *tview.TextView
	chatBottomBox *tview.Box
	userQuestion  *tview.TextView
	currentMode   Mode
	currentChat   *ChatSession
	chats         []*ChatSession
	isStreaming   bool
	tools         []Tool
	app           *tview.Application
}

func NewApp() *App {
	return &App{
		currentMode: ModeCode,
	}
}

func (a *App) Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dataPath := filepath.Join(home, dataDir, chatsDir)
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return err
	}

	a.chats = loadChats(dataPath)
	a.currentChat = a.newChatSession()

	a.initTools()
	a.buildUI()
	return nil
}

func (a *App) initTools() {
	a.tools = []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "execute_command",
				Description: "Execute a shell command",
				Parameters: ToolParameter{
					Type: "object",
					Properties: map[string]Property{
						"command": {Type: "string", Description: "The shell command to execute"},
					},
					Required: []string{"command"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "git_commit",
				Description: "Commit changes to git repository",
				Parameters: ToolParameter{
					Type: "object",
					Properties: map[string]Property{
						"message": {Type: "string", Description: "Commit message"},
					},
					Required: []string{"message"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "compile_files",
				Description: "Compile source code files",
				Parameters: ToolParameter{
					Type: "object",
					Properties: map[string]Property{
						"files":   {Type: "string", Description: "Files to compile (comma separated)"},
						"command": {Type: "string", Description: "Compilation command"},
					},
					Required: []string{"files", "command"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "read_file",
				Description: "Read contents of a file",
				Parameters: ToolParameter{
					Type: "object",
					Properties: map[string]Property{
						"path": {Type: "string", Description: "File path to read"},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "list_directory",
				Description: "List contents of a directory",
				Parameters: ToolParameter{
					Type: "object",
					Properties: map[string]Property{
						"path": {Type: "string", Description: "Directory path to list"},
					},
					Required: []string{"path"},
				},
			},
		},
	}
}

func (a *App) newChatSession() *ChatSession {
	id := fmt.Sprintf("chat_%d", time.Now().UnixMilli())
	name := fmt.Sprintf("Chat %s", time.Now().Format("01/02 15:04"))

	session := &ChatSession{
		ID:        id,
		Name:      name,
		Mode:      a.currentMode,
		CreatedAt: time.Now(),
		Messages: []Message{
			{Role: "system", Content: a.getSystemPrompt()},
		},
	}

	home, _ := os.UserHomeDir()
	saveChats(home, a.chats)
	return session
}

func (a *App) getSystemPrompt() string {
	prompt := "You are TrashNeurons, a powerful AI assistant for Linux desktop and development."
	if a.currentMode == ModeCode {
		prompt += " You specialize in coding, debugging, and software development. Use tools to execute commands, read files, and manage code."
	} else {
		prompt += " You are in chat mode. Have conversations, answer questions, and be helpful."
	}
	return prompt
}

func (a *App) buildUI() {
	a.pages = tview.NewPages()

	a.buildSidePanel()
	a.buildRightPanel()

	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	mainFlex.AddItem(a.sidePanel, 20, 1, false)
	mainFlex.AddItem(a.rightPanel, 0, 10, false)

	mainFrame := tview.NewFrame(mainFlex)
	mainFrame.SetBorder(true)

	a.pages.AddPage("main", mainFrame, true, true)

	a.app = tview.NewApplication()
	a.app.SetRoot(a.pages, true)

	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			a.ToggleMode()
			return nil
		}
		return event
	})
}

func (a *App) buildSidePanel() {
	panelFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	modeLabel := tview.NewTextView().SetText("current mode:").SetTextColor(tcell.ColorGray)
	panelFlex.AddItem(modeLabel, 1, 1, false)

	a.modeText = tview.NewTextView().SetText("code").SetTextColor(tcell.ColorYellow)
	panelFlex.AddItem(a.modeText, 1, 1, false)

	a.arrowUp = tview.NewTextView().SetText("   ^").SetTextColor(tcell.ColorWhite)
	panelFlex.AddItem(a.arrowUp, 1, 1, false)

	a.arrowDown = tview.NewTextView().SetText("   v").SetTextColor(tcell.ColorWhite)
	panelFlex.AddItem(a.arrowDown, 1, 1, false)

	a.separator = tview.NewTextView().SetText("──┴───────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────").SetTextColor(tcell.ColorGray)
	panelFlex.AddItem(a.separator, 1, 1, false)

	a.sidePanel = tview.NewFrame(panelFlex)
	a.sidePanel.SetBorder(true)
	a.sidePanel.SetTitle(" trashneurons ")
}

func (a *App) getModeName() string {
	if a.currentMode == ModeCode {
		return "code"
	}
	return "chat"
}

func (a *App) buildRightPanel() {
	a.rightPanel = tview.NewFlex().SetDirection(tview.FlexRow)

	a.shipLogo = tview.NewTextView().SetText(`            _                 _                                                          _
           | |_ _ __ __ _ ___| |__  _ __   ___ _   _ _ __ ___  _ __  ___    ___ ___   __| | ___
           | __| '__/ _` + "`" + ` / __| '_ \| '_ \ / _ \ | | | '__/ _ \| '_ \/ __|  / __/ _ \ / _` + "`" + ` |/ _ \
           | |_| | | (_| \__ \ | | | | | |  __/ |_| | | | (_) | | | \__ \ | (_| (_) | (_| |  __/
            \__|_|  \__,_|___/_| |_|_| |_|\___|\__,_|_|  \___/|_| |_|___/  \___\___/ \__,_|\___|`).SetTextColor(tcell.ColorYellow)

	a.rightPanel.AddItem(a.shipLogo, 8, 1, false)

	a.animationLbl = tview.NewTextView().SetText("")
	a.rightPanel.AddItem(a.animationLbl, 15, 1, false)

	a.mainInputBox = tview.NewBox().SetBorder(true).SetTitle("")
	a.rightPanel.AddItem(a.mainInputBox, 1, 1, false)

	a.mainInput = tview.NewInputField().SetFieldWidth(50)
	a.mainInput.SetPlaceholder("ask anything... (use tab to switch modes)")
	a.mainInput.SetChangedFunc(func(text string) {
		if len(text) > 0 {
			display := text[:minInt(len(text), 40)]
			a.mainInputBox.SetTitle(fmt.Sprintf(" %s ", display))
		} else {
			a.mainInputBox.SetTitle(" ask anything... (use tab to switch modes) ")
		}
	})
	a.mainInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := a.mainInput.GetText()
			if text != "" {
				a.currentChat.Messages = append(a.currentChat.Messages, Message{Role: "user", Content: text})
				a.mainInput.SetText("")
				a.mainInputBox.SetTitle("")
				go a.processMessage(text)
			}
		}
	})

	a.rightPanel.AddItem(a.mainInput, 1, 1, true)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *App) processMessage(content string) {
	a.isStreaming = true

	a.thinkingLbl = tview.NewTextView().SetText("thinking...").SetTextColor(tcell.ColorGreen)

	resp, err := callAgnesAPI(a.currentChat.Messages, a.tools, true)
	if err != nil {
		a.aiAnswerLbl = tview.NewTextView().SetText(fmt.Sprintf("Error: %v", err)).SetTextColor(tcell.ColorRed)
		a.isStreaming = false
		return
	}

	var fullResponse strings.Builder
	var toolCalls []ToolCall

	for _, choice := range resp.Choices {
		if choice.Message.Content != "" {
			fullResponse.WriteString(choice.Message.Content)
		}
		if len(choice.Message.ToolCalls) > 0 {
			toolCalls = append(toolCalls, choice.Message.ToolCalls...)
		}
	}

	minutes := 0
	seconds := 1
	a.thinkingLbl = tview.NewTextView().SetText(fmt.Sprintf("thinked for %dm%ds", minutes, seconds)).SetTextColor(tcell.ColorGray)

	if len(toolCalls) > 0 {
		a.handleToolCalls(toolCalls)
	} else {
		a.aiAnswerLbl = tview.NewTextView().SetText("trashneurons: " + fullResponse.String()).SetTextColor(tcell.ColorLightCyan)
		a.currentChat.Messages = append(a.currentChat.Messages, Message{
			Role:    "assistant",
			Content: fullResponse.String(),
		})
	}

	home, _ := os.UserHomeDir()
	saveChats(home, a.chats)

	a.isStreaming = false
}

func (a *App) handleToolCalls(toolCalls []ToolCall) {
	for _, tc := range toolCalls {
		result := executeTool(tc.Function.Name, tc.Function.Arguments)

		a.aiAnswerLbl = tview.NewTextView().SetText(fmt.Sprintf("tool: %s\n%s", tc.Function.Name, result)).SetTextColor(tcell.ColorYellow)

		a.currentChat.Messages = append(a.currentChat.Messages, Message{
			Role:    "tool",
			Content: result,
		})

		time.Sleep(500 * time.Millisecond)

		resp, err := callAgnesAPI(a.currentChat.Messages, nil, false)
		if err != nil {
			continue
		}

		var finalResp strings.Builder
		for _, choice := range resp.Choices {
			finalResp.WriteString(choice.Message.Content)
		}

		a.aiAnswerLbl = tview.NewTextView().SetText("trashneurons: " + finalResp.String()).SetTextColor(tcell.ColorLightCyan)
		a.currentChat.Messages = append(a.currentChat.Messages, Message{
			Role:    "assistant",
			Content: finalResp.String(),
		})
	}
}

func (a *App) showChatScreen() {
	topFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	a.chatTopBox = tview.NewBox().SetBorder(true).SetTitle("")
	topFlex.AddItem(a.chatTopBox, 1, 1, false)

	if a.thinkingLbl != nil {
		topFlex.AddItem(a.thinkingLbl, 1, 1, false)
	}

	if a.aiAnswerLbl != nil {
		topFlex.AddItem(a.aiAnswerLbl, 0, 10, true)
	}

	bottomFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	a.chatBottomBox = tview.NewBox().SetBorder(true).SetTitle("")
	bottomFlex.AddItem(a.chatBottomBox, 1, 1, false)

	if a.userQuestion != nil {
		bottomFlex.AddItem(a.userQuestion, 0, 10, false)
	}

	centerFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	centerFlex.AddItem(topFlex, 0, 7, false)
	centerFlex.AddItem(bottomFlex, 0, 3, false)

	chatFrame := tview.NewFrame(centerFlex)
	chatFrame.SetBorder(true)

	a.pages.AddAndSwitchToPage("chat", chatFrame, true)
}

func (a *App) ToggleMode() {
	if a.currentMode == ModeCode {
		a.currentMode = ModeChat
	} else {
		a.currentMode = ModeCode
	}

	a.currentChat.Mode = a.currentMode
	a.modeText.SetText(a.getModeName())
	a.currentChat.Messages[0].Content = a.getSystemPrompt()
}

func (a *App) Run() error {
	go a.runAnimation()
	return a.app.Run()
}

func (a *App) runAnimation() {
	particles := []string{"·", "·", "·", "·", "·", "⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"}
	idx := 0

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		a.app.QueueUpdateDraw(func() {
			if a.animationLbl != nil {
				text := ""
				for i := 0; i < 20; i++ {
					text += particles[idx%len(particles)]
					idx++
				}
				a.animationLbl.SetText(text)
			}
		})
	}
}

func callAgnesAPI(messages []Message, tools []Tool, thinking bool) (*ChatCompletionResponse, error) {
	payload := map[string]interface{}{
		"model":    modelName,
		"messages": messages,
		"stream":   false,
	}

	if thinking {
		payload["thinking"] = map[string]interface{}{
			"type":          "enabled",
			"budget_tokens": 2048,
		}
	}

	if len(tools) > 0 {
		payload["tools"] = tools
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiBaseURL+"/chat/completions", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func executeTool(name string, argsJSON string) string {
	switch name {
	case "execute_command":
		var params struct {
			Command string `json:"command"`
		}
		json.Unmarshal([]byte(argsJSON), &params)
		cmd := exec.Command("sh", "-c", params.Command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
		}
		return string(output)

	case "git_commit":
		var params struct {
			Message string `json:"message"`
		}
		json.Unmarshal([]byte(argsJSON), &params)
		cmd1 := exec.Command("git", "add", ".")
		cmd1.Run()
		cmd2 := exec.Command("git", "commit", "-m", params.Message)
		output, err := cmd2.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
		}
		return string(output)

	case "compile_files":
		var params struct {
			Files   string `json:"files"`
			Command string `json:"command"`
		}
		json.Unmarshal([]byte(argsJSON), &params)
		cmd := exec.Command("sh", "-c", params.Command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Error: %v\nOutput: %s", err, string(output))
		}
		return string(output)

	case "read_file":
		var params struct {
			Path string `json:"path"`
		}
		json.Unmarshal([]byte(argsJSON), &params)
		data, err := os.ReadFile(params.Path)
		if err != nil {
			return fmt.Sprintf("Error reading file: %v", err)
		}
		return string(data)

	case "list_directory":
		var params struct {
			Path string `json:"path"`
		}
		json.Unmarshal([]byte(argsJSON), &params)
		entries, err := os.ReadDir(params.Path)
		if err != nil {
			return fmt.Sprintf("Error listing directory: %v", err)
		}
		var sb strings.Builder
		for _, e := range entries {
			info, _ := e.Info()
			sb.WriteString(fmt.Sprintf("%s %s\n", info.Mode().String(), e.Name()))
		}
		return sb.String()
	}

	return "Unknown tool: " + name
}

func loadChats(path string) []*ChatSession {
	var chats []*ChatSession
	files, err := os.ReadDir(path)
	if err != nil {
		return chats
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			data, err := os.ReadFile(filepath.Join(path, f.Name()))
			if err == nil {
				var chat ChatSession
				if err := json.Unmarshal(data, &chat); err == nil {
					chats = append(chats, &chat)
				}
			}
		}
	}
	return chats
}

func saveChats(home string, chats []*ChatSession) {
	path := filepath.Join(home, dataDir, chatsDir)
	for _, chat := range chats {
		filename := filepath.Join(path, chat.ID+".json")
		data, _ := json.MarshalIndent(chat, "", "  ")
		os.WriteFile(filename, data, 0644)
	}
}

func showBootScreen() {
	bootApp := tview.NewApplication()

	outer := tview.NewFlex().SetDirection(tview.FlexRow)

	ship1 := tview.NewTextView().SetText(`            ___________                    .__      _______
            \__    ___/___________    _____|  |__   \      \   ____  __ _________  ____   ____   ______
              |    |  \_  __ \__  \  /  ___/  |  \  /   |   \_/ __ \|  |  \_  __ \/  _ \ /    \ /  ___/
              |    |   |  | \// __ \_\___ \|   Y  \/    |    \  ___/|  |  /|  | \(  <_> )   |  \\___ \
              |____|   |__|  (____  /____  >___|  /\____|__  /\___  >____/ |__|   \____/|___|  /____  >
                                  \/     \/     \/         \/     \/                         \/     \/`).SetTextColor(tcell.ColorYellow)

	outer.AddItem(ship1, 15, 1, false)

	sepLine := tview.NewTextView().SetText(strings.Repeat("─", 80)).SetTextColor(tcell.ColorGray)
	outer.AddItem(sepLine, 1, 1, false)

	ship2 := tview.NewTextView().SetText(`                                                                         __  .__    .__
           _____________   ____   ______ ______   _____    ____ ___.__._/  |_|  |__ |__| ____    ____
           \____ \_  __ \_/ __ \ /  ___//  ___/   \__  \  /    <   |  |\   __\  |  \|  |/    \  / ___\
           |  |_> >  | \/\  ___/ \___ \ \___ \     / __ \|   |  \___  | |  | |   Y  \  |   |  \/ /_/  >
           |   __/|__|    \___  >____  >____  >   (____  /___|  / ____| |__| |___|  /__|___|  /\___  /
           |__|               \/     \/     \/         \/     \/\/                \/        \//_____/`).SetTextColor(tcell.ColorYellow)

	outer.AddItem(ship2, 12, 1, false)

	statusText := tview.NewTextView().SetText("waiting for user input:").SetTextColor(tcell.ColorGreen)
	outer.AddItem(statusText, 1, 1, false)

	inputField := tview.NewInputField()
	inputField.SetFieldWidth(40)
	inputField.SetPlaceholder("press any key to continue...")
	inputField.SetChangedFunc(func(text string) {})

	done := make(chan struct{})
	inputField.SetDoneFunc(func(key tcell.Key) {
		bootApp.Stop()
		go func() {
			for i := 3; i > 0; i-- {
				bootApp.QueueUpdateDraw(func() {
					statusText.SetText(fmt.Sprintf("successfully! opening in %d...", i))
				})
				time.Sleep(1 * time.Second)
			}
			bootApp.Stop()
			done <- struct{}{}
		}()
	})

	inputField.SetChangedFunc(func(text string) {
		if text != "" {
			bootApp.Stop()
			go func() {
				for i := 3; i > 0; i-- {
					bootApp.QueueUpdateDraw(func() {
						statusText.SetText(fmt.Sprintf("successfully! opening in %d...", i))
					})
					time.Sleep(1 * time.Second)
				}
				bootApp.Stop()
				done <- struct{}{}
			}()
		}
	})

	outer.AddItem(inputField, 1, 1, true)

	bootOuter := tview.NewFlex().SetDirection(tview.FlexRow)
	bootOuter.AddItem(nil, 1, 1, false)
	bootOuter.AddItem(outer, 0, 10, false)
	bootOuter.AddItem(nil, 1, 1, false)

	bootApp.SetRoot(bootOuter, true)
	bootApp.Run()
	<-done
}

func main() {
	app := NewApp()

	if err := app.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}

	showBootScreen()
	app.Run()
}
