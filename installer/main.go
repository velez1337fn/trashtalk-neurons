package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	releaseURL = "https://api.github.com/repos/velez1337fn/trashtalk-neurons/releases/latest"
	token      = ""
)

func main() {
	app := tview.NewApplication()

	outer := tview.NewFlex().SetDirection(tview.FlexRow)

	logo := tview.NewTextView().SetText(`   .___                 __         .__  .__                              ____    _______      _______
   |   | ____   _______/  |______  |  | |  |   ___________      ___  __ /_   |   \   _  \     \   _  \
   |   |/    \ /  ___/\   __\__  \ |  | |  | _/ __ \_  __ \     \  \/ /  |   |   /  /_\  \    /  /_\  \
   |   |   |  \\___ \  |  |  / __ \|  |_|  |_\  ___/|  | \/      \   /   |   |   \  \_/   \   \  \_/   \
   |___|___|  /____  > |__| (____  /____/____/\___  >__|          \_/    |___| /\ \_____  / /\ \_____  /
            \/     \/            \/               \/                           \/       \/  \/       \/`).SetTextColor(tcell.ColorLightCyan)

	outer.AddItem(logo, 8, 1, false)

	user, _ := os.UserHomeDir()
	user = strings.TrimSuffix(user, "/"+os.Getenv("USER"))
	if user == "" {
		user = "user"
	}

	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		arch = "amd64"
	}

	sysInfo := fmt.Sprintf("linux-%s", arch)

	sidePanel := tview.NewFlex().SetDirection(tview.FlexRow)

	welcome := tview.NewTextView().SetText(fmt.Sprintf("welcome, %s!", user)).SetTextColor(tcell.ColorLightCyan)
	sidePanel.AddItem(welcome, 1, 1, false)

	matrix := tview.NewTextView().SetText("matrix in this box").SetTextColor(tcell.ColorBlue)
	sidePanel.AddItem(matrix, 1, 1, false)

	detected := tview.NewTextView().SetText(fmt.Sprintf("detected sys: %s", sysInfo)).SetTextColor(tcell.ColorYellow)
	sidePanel.AddItem(detected, 1, 1, false)

	status := tview.NewTextView().SetText("installer status: ready").SetTextColor(tcell.ColorYellow)
	sidePanel.AddItem(status, 1, 1, false)

	waiting := tview.NewTextView().SetText("waiting for user input...").SetTextColor(tcell.ColorYellow)
	sidePanel.AddItem(waiting, 1, 1, false)

	winning := tview.NewTextView().SetText(`are ya winning son?     ***   ***
    ─┬────┐             ***** *****
     └─┬──┘              *********
   ┌───┼──┐               *******
   │   │  │                *****
     ┌─┼─┐                  ***
     │   │                   *
     └─┬─┬┘                    ──
       └─┘                         ***`).SetTextColor(tcell.ColorRed)
	sidePanel.AddItem(winning, 8, 1, false)

	opt1 := tview.NewTextView().SetText("  1. Install TrashNeurons CLI").SetTextColor(tcell.ColorGreen)
	opt2 := tview.NewTextView().SetText("  2. Exit").SetTextColor(tcell.ColorGreen)
	sidePanel.AddItem(opt1, 1, 1, false)
	sidePanel.AddItem(opt2, 1, 1, false)

	sideBox := tview.NewFrame(sidePanel).SetBorder(true)

	inputField := tview.NewInputField().SetLabel("Choose option [1-2]: ")
	inputField.SetFieldWidth(2)

	done := make(chan int, 1)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			val := strings.TrimSpace(inputField.GetText())
			if val == "1" {
				done <- 1
			} else if val == "2" {
				done <- 2
			} else {
				done <- 0
			}
		}
	})

	menuFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	menuFlex.AddItem(sideBox, 20, 1, false)
	menuFlex.AddItem(tview.NewBox(), 0, 1, false)
	menuFlex.AddItem(inputField, 0, 1, true)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	mainFlex.AddItem(menuFlex, 0, 10, false)

	frame := tview.NewFrame(mainFlex)
	frame.SetBorder(true)

	app.SetRoot(frame, true)

	go func() {
		app.Run()
	}()

	choice := <-done
	close(done)

	if choice == 2 {
		os.Exit(0)
	}
	if choice != 1 {
		fmt.Println("Invalid option.")
		os.Exit(1)
	}

	downloadAndInstall()
}

func downloadAndInstall() {
	app := tview.NewApplication()

	view := tview.NewTextView()
	view.SetBorder(true).SetTitle(" installer ")

	installFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	installFlex.AddItem(view, 0, 10, false)

	app.SetRoot(installFlex, true)

	go func() {
		writeLine(view, "> allocating path where app/cli been installed...")

		home, _ := os.UserHomeDir()
		installDir := filepath.Join(home, ".trashneurons")
		binDir := filepath.Join(installDir, "bin")
		os.MkdirAll(binDir, 0755)

		writeLine(view, fmt.Sprintf("found! path is: %s", installDir))
		writeLine(view, "")
		writeLine(view, "> checking for latest release...")

		arch := runtime.GOARCH
		switch arch {
		case "amd64", "arm64":
		default:
			arch = "amd64"
		}

		binName := fmt.Sprintf("trashneurons-cli-%s", arch)
		tmpBin := filepath.Join("/tmp", binName)

		resp, err := http.Get(releaseURL)
		if err != nil {
			writeLine(view, fmt.Sprintf("ERROR: %v", err))
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var release struct {
			Assets []struct {
				Name              string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
		}
		jsonUnmarshal(body, &release)

		var assetURL string
		for _, a := range release.Assets {
			if strings.Contains(a.Name, arch) {
				assetURL = a.BrowserDownloadURL
				break
			}
		}

		if assetURL == "" {
			writeLine(view, fmt.Sprintf("ERROR: No release found for arch %s", arch))
			writeLine(view, "Please run the GitHub Actions workflow first to build the binary.")
			return
		}

		writeLine(view, fmt.Sprintf("found release! downloading..."))

		dlResp, err := http.Get(assetURL)
		if err != nil {
			writeLine(view, fmt.Sprintf("ERROR: Download failed: %v", err))
			return
		}
		defer dlResp.Body.Close()

		outFile, _ := os.Create(tmpBin)
		io.Copy(outFile, dlResp.Body)
		outFile.Close()
		os.Chmod(tmpBin, 0755)

		writeLine(view, "downloaded!")
		writeLine(view, "installing your app/cli...")

		destBin := filepath.Join(binDir, "trashneurons-cli")
		os.Rename(tmpBin, destBin)

		symPath := "/usr/local/bin/trashneurons-cli"
		exec.Command("sudo", "ln", "-sf", destBin, symPath).Run()

		writeLine(view, "")
		writeLine(view, fmt.Sprintf("finished install! all files located in %s", installDir))
		writeLine(view, fmt.Sprintf("for use: type \"trashneurons-cli\""))
		writeLine(view, "")
		writeLine(view, "PRESS ENTER TO CLOSE INSTALLER")
	}()

	inputField := tview.NewInputField()
	inputField.SetLabel(" PRESS CTRL+C TO CLOSE INSTALLER ")
	inputField.SetChangedFunc(func(text string) {})
	inputField.SetDoneFunc(func(key tcell.Key) {
		app.Stop()
		os.Exit(0)
	})

	installFlex.AddItem(inputField, 1, 1, true)
	app.Run()
}

func writeLine(view *tview.TextView, line string) {
	view.SetText(view.GetText(false) + line + "\n")
	view.ScrollToEnd()
}

func jsonUnmarshal(data []byte, v interface{}) {
	_ = data
	_ = v
}
