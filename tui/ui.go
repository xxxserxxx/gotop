package tui

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/shibukawa/configdir"
	"github.com/xxxserxxx/gotop/v4"
	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/gotop/v4/layout"
	"github.com/xxxserxxx/gotop/v4/widgets"
)

const (
	graphHorizontalScaleDelta = 3
	defaultUI                 = "2:cpu\ndisk/1 2:mem/2\ntemp\n2:net 2:procs"
	minimalUI                 = "cpu\nmem procs"
	batteryUI                 = "cpu/2 batt/1\ndisk/1 2:mem/2\ntemp\nnet procs"
	procsUI                   = "cpu 4:procs\ndisk\nmem\nnet"
	kitchensink               = "3:cpu/2 3:mem/1\n4:temp/1 3:disk/2\npower\n3:net 3:procs"
)

var (
	help *widgets.HelpMenu
	bar  *widgets.StatusBar
)

type TUI struct {
	conf gotop.Config
}

func New(conf gotop.Config) (TUI, error) {
	return TUI{conf}, termui.Init()
}

func (t TUI) ShutdownUI() {
	termui.Close()
}

func (t TUI) LoopUI() error {
	lstream, err := getLayout(t.conf)
	if err != nil {
		return err
	}
	ly := layout.ParseLayout(lstream)

	devNames := make([]string, 0)
	// First, check the layout; each widget has a default associated device
	if !t.conf.NoLocal {
		for _, row := range ly.Rows {
			for _, col := range row {
				devNames = append(devNames, col.Widget)
			}
		}
	}
	devInsts, errs := devices.Startup(devNames, t.conf)
	if len(errs) > 0 {
		stderrLogger := log.New(os.Stderr, "", 0)
		for _, err := range errs {
			stderrLogger.Print(err)
		}
		return fmt.Errorf("device initialization error(s)")
	}
	devices.Spawn(devInsts, t.conf)
	setDefaultTermuiColors(t.conf) // done before initializing widgets to allow inheriting colors
	help = widgets.NewHelpMenu(t.conf.Tr)
	if t.conf.Statusbar {
		bar = widgets.NewStatusBar()
	}

	grid, err := layout.NewLayout(ly, t.conf, devInsts)
	if err != nil {
		return err
	}

	termWidth, termHeight := termui.TerminalDimensions()
	if t.conf.Statusbar {
		grid.SetRect(0, 0, termWidth, termHeight-1)
	} else {
		grid.SetRect(0, 0, termWidth, termHeight)
	}
	help.Resize(termWidth, termHeight)

	termui.Render(grid)
	if t.conf.Statusbar {
		bar.SetRect(0, termHeight-1, termWidth, termHeight)
		termui.Render(bar)
	}
	eventLoop(t.conf, grid)
	return nil
}

func setDefaultTermuiColors(c gotop.Config) {
	termui.Theme.Default = termui.NewStyle(termui.Color(c.Colorscheme.Fg), termui.Color(c.Colorscheme.Bg))
	termui.Theme.Block.Title = termui.NewStyle(termui.Color(c.Colorscheme.BorderLabel), termui.Color(c.Colorscheme.Bg))
	termui.Theme.Block.Border = termui.NewStyle(termui.Color(c.Colorscheme.BorderLine), termui.Color(c.Colorscheme.Bg))
}

func eventLoop(c gotop.Config, grid *layout.Screen) {
	drawTicker := time.NewTicker(c.UpdateInterval).C

	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	uiEvents := termui.PollEvents()

	previousKey := ""

	for {
		select {
		case <-sigTerm:
			return
		case <-drawTicker:
			if !c.HelpVisible {
				grid.Widgets.Update()
				termui.Render(grid)
				if c.Statusbar {
					termui.Render(bar)
				}
			}
		case e := <-uiEvents:
			if grid.Proc != nil && grid.Proc.HandleEvent(e) {
				termui.Render(grid.Proc)
				break
			}
			switch e.ID {
			case "q", "<C-c>":
				return
			case "?":
				c.HelpVisible = !c.HelpVisible
			case "<Resize>":
				payload := e.Payload.(termui.Resize)
				termWidth, termHeight := payload.Width, payload.Height
				if c.Statusbar {
					grid.SetRect(0, 0, termWidth, termHeight-1)
					bar.SetRect(0, termHeight-1, termWidth, termHeight)
				} else {
					grid.SetRect(0, 0, payload.Width, payload.Height)
				}
				help.Resize(payload.Width, payload.Height)
				termui.Clear()
			}

			if c.HelpVisible {
				switch e.ID {
				case "?":
					termui.Clear()
					termui.Render(help)
				case "<Escape>":
					c.HelpVisible = false
					termui.Render(grid)
				case "<Resize>":
					termui.Render(help)
				}
			} else {
				switch e.ID {
				case "?":
					termui.Render(grid)
				case "h":
					c.GraphHorizontalScale += graphHorizontalScaleDelta
					for _, item := range grid.Lines {
						item.Scale(c.GraphHorizontalScale)
					}
					termui.Render(grid)
				case "l":
					if c.GraphHorizontalScale > graphHorizontalScaleDelta {
						c.GraphHorizontalScale -= graphHorizontalScaleDelta
						for _, item := range grid.Lines {
							item.Scale(c.GraphHorizontalScale)
							termui.Render(item)
						}
					}
				case "b":
					if grid.Net != nil {
						grid.Net.Mbps = !grid.Net.Mbps
					}
				case "<Resize>":
					termui.Render(grid)
					if c.Statusbar {
						termui.Render(bar)
					}
				case "<MouseLeft>":
					if grid.Proc != nil {
						payload := e.Payload.(termui.Mouse)
						grid.Proc.HandleClick(payload.X, payload.Y)
						termui.Render(grid.Proc)
					}
				case "k", "<Up>", "<MouseWheelUp>":
					if grid.Proc != nil {
						grid.Proc.ScrollUp()
						termui.Render(grid.Proc)
					}
				case "j", "<Down>", "<MouseWheelDown>":
					if grid.Proc != nil {
						grid.Proc.ScrollDown()
						termui.Render(grid.Proc)
					}
				case "<Home>":
					if grid.Proc != nil {
						grid.Proc.ScrollTop()
						termui.Render(grid.Proc)
					}
				case "g":
					if grid.Proc != nil {
						if previousKey == "g" {
							grid.Proc.ScrollTop()
							termui.Render(grid.Proc)
						}
					}
				case "G", "<End>":
					if grid.Proc != nil {
						grid.Proc.ScrollBottom()
						termui.Render(grid.Proc)
					}
				case "<C-d>":
					if grid.Proc != nil {
						grid.Proc.ScrollHalfPageDown()
						termui.Render(grid.Proc)
					}
				case "<C-u>":
					if grid.Proc != nil {
						grid.Proc.ScrollHalfPageUp()
						termui.Render(grid.Proc)
					}
				case "<C-f>":
					if grid.Proc != nil {
						grid.Proc.ScrollPageDown()
						termui.Render(grid.Proc)
					}
				case "<C-b>":
					if grid.Proc != nil {
						grid.Proc.ScrollPageUp()
						termui.Render(grid.Proc)
					}
				case "d":
					if grid.Proc != nil {
						if previousKey == "d" {
							grid.Proc.KillProc("SIGTERM")
						}
					}
				case "3":
					if grid.Proc != nil {
						if previousKey == "d" {
							grid.Proc.KillProc("SIGQUIT")
						}
					}
				case "9":
					if grid.Proc != nil {
						if previousKey == "d" {
							grid.Proc.KillProc("SIGKILL")
						}
					}
				case "<Tab>":
					if grid.Proc != nil {
						grid.Proc.ToggleShowingGroupedProcs()
						termui.Render(grid.Proc)
					}
				case "m", "c", "p":
					if grid.Proc != nil {
						grid.Proc.ChangeProcSortMethod(widgets.ProcSortMethod(e.ID))
						termui.Render(grid.Proc)
					}
				case "/":
					if grid.Proc != nil {
						grid.Proc.SetEditingFilter(true)
						termui.Render(grid.Proc)
					}
				}

				if previousKey == e.ID {
					previousKey = ""
				} else {
					previousKey = e.ID
				}
			}

		}
	}
}

func getLayout(conf gotop.Config) (io.Reader, error) {
	switch conf.Layout {
	case "-":
		return os.Stdin, nil
	case "default":
		return strings.NewReader(defaultUI), nil
	case "minimal":
		return strings.NewReader(minimalUI), nil
	case "battery":
		return strings.NewReader(batteryUI), nil
	case "procs":
		return strings.NewReader(procsUI), nil
	case "kitchensink":
		return strings.NewReader(kitchensink), nil
	default:
		folder := conf.ConfigDir.QueryFolderContainsFile(conf.Layout)
		if folder == nil {
			paths := make([]string, 0)
			for _, d := range conf.ConfigDir.QueryFolders(configdir.Existing) {
				paths = append(paths, d.Path)
			}
			return nil, fmt.Errorf(conf.Tr.Value("error.findlayout", conf.Layout, strings.Join(paths, ", ")))
		}
		lo, err := folder.ReadFile(conf.Layout)
		if err != nil {
			return nil, err
		}
		return strings.NewReader(string(lo)), nil
	}
}
