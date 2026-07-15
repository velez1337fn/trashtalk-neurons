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
	separator     *tview.TextView
	mainFlex      *tview.Flex
	inputBox      *tview.Box
	inputField    *tview.InputField
	chatArea      *tview.TextView
	animationLbl  *tview.TextView
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
	a.buildMainArea()

	mainFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	mainFlex.AddItem(a.sidePanel, 20, 1, false)
	mainFlex.AddItem(a.mainFlex, 0, 10, false)

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

func (a *App) buildMainArea() {
	a.mainFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	shipLogo := tview.NewTextView().SetText(`            _                 _                                                          _
           | |_ _ __ __ _ ___| |__  _ __   ___ _   _ _ __ ___  _ __  ___    ___ ___   __| | ___
           | __| '__/ _` + "`" + ` / __| '_ \| '_ \ / _ \ | | | '__/ _ \| '_ \/ __|  / __/ _ \ / _` + "`" + ` |/ _ \
           | |_| | | (_| \__ \ | | | | | |  __/ |_| | | | (_) | | | \__ \ | (_| (_) | (_| |  __/
            \__|_|  \__,_|___/_| |_|_| |_|\___|\__,_|_|  \___/|_| |_|___/  \___\___/ \__,_|\___|`).SetTextColor(tcell.ColorYellow)
	a.mainFlex.AddItem(shipLogo, 8, 1, false)

	a.animationLbl = tview.NewTextView().SetText("")
	a.mainFlex.AddItem(a.animationLbl, 15, 1, false)

	a.chatArea = tview.NewTextView().SetScrollable(true)
	a.mainFlex.AddItem(a.chatArea, 0, 10, false)

	a.inputBox = tview.NewBox().SetBorder(true).SetTitle("")
	a.mainFlex.AddItem(a.inputBox, 1, 1, false)

	a.inputField = tview.NewInputField().SetFieldWidth(50)
	a.inputField.SetPlaceholder("ask anything... (use tab to switch modes)")
	a.inputField.SetChangedFunc(func(text string) {
		if len(text) > 0 {
			display := text[:minInt(len(text), 40)]
			a.inputBox.SetTitle(fmt.Sprintf(" %s ", display))
		} else {
			a.inputBox.SetTitle(" ask anything... (use tab to switch modes) ")
		}
	})
	a.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := a.inputField.GetText()
			if text != "" {
				a.currentChat.Messages = append(a.currentChat.Messages, Message{Role: "user", Content: text})
				a.renderChatHistory()
				a.inputField.SetText("")
				a.inputBox.SetTitle("")
				go a.processMessage(text)
			}
		}
	})

	a.mainFlex.AddItem(a.inputField, 1, 1, true)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (a *App) appendChat(text string, color tcell.Color) {
	a.chatArea.SetTextColor(color)
	a.chatArea.Write([]byte(text))
	a.chatArea.ScrollToEnd()
}

func (a *App) processMessage(content string) {
	a.isStreaming = true

	a.appendChat("\n"+strings.Repeat("─", 60)+"\n", tcell.ColorGreen)
	a.appendChat("thinking...\n", tcell.ColorGreen)

	resp, err := callAgnesAPI(a.currentChat.Messages, a.tools, true)
	if err != nil {
		a.appendChat(fmt.Sprintf("\nError: %v\n", err), tcell.ColorRed)
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

	a.appendChat("thinked for 0m1s\n", tcell.ColorGray)
	a.appendChat(strings.Repeat("─", 60)+"\n\n", tcell.ColorGray)
	a.appendChat("trashneurons: ", tcell.ColorLightCyan)

	if len(toolCalls) > 0 {
		a.handleToolCalls(toolCalls)
	} else {
		a.appendChat(fullResponse.String()+"\n", tcell.ColorLightCyan)
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
		a.appendChat(fmt.Sprintf("tool call: %s\n", tc.Function.Name), tcell.ColorYellow)
		result := executeTool(tc.Function.Name, tc.Function.Arguments)
		a.appendChat(result+"\n", tcell.ColorDefault)

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

		a.appendChat("trashneurons: ", tcell.ColorLightCyan)
		a.appendChat(finalResp.String()+"\n", tcell.ColorLightCyan)
		a.currentChat.Messages = append(a.currentChat.Messages, Message{
			Role:    "assistant",
			Content: finalResp.String(),
		})
	}
}

func (a *App) renderChatHistory() {
	a.chatArea.Clear()
	a.chatArea.SetTextColor(tcell.ColorDefault)

	for _, msg := range a.currentChat.Messages[1:] {
		switch msg.Role {
		case "user":
			a.appendChat("\n"+strings.Repeat("─", 60)+"\n", tcell.ColorGreen)
			a.appendChat("you: ", tcell.ColorGreen)
			a.appendChat(msg.Content+"\n", tcell.ColorDefault)
		case "assistant":
			a.appendChat("\n", tcell.ColorDefault)
			a.appendChat("trashneurons: ", tcell.ColorLightCyan)
			a.appendChat(msg.Content+"\n", tcell.ColorDefault)
		case "tool":
			a.appendChat("\n", tcell.ColorDefault)
			a.appendChat("[tool output]: ", tcell.ColorYellow)
			a.appendChat(msg.Content+"\n", tcell.ColorDefault)
		}
	}
	a.chatArea.ScrollToEnd()
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
	if fi, err := os.Stdin.Stat(); err == nil && (fi.Mode()&os.ModeNamedPipe) != 0 {
		return
	}
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

	done := make(chan struct{})

	enterDone := func() {
		bootApp.Stop()
		go func() {
			for i := 3; i > 0; i-- {
				bootApp.QueueUpdateDraw(func() {
					statusText.SetText(fmt.Sprintf("successfully! opening in %d...", i))
				})
				time.Sleep(1 * time.Second)
			}
			done <- struct{}{}
		}()
	}

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			enterDone()
		}
	})

	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() != tcell.KeyEnter {
			enterDone()
			return nil
		}
		return event
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
