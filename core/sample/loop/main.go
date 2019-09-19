package main

import (
	"fmt"
	"os"

	ui "github.com/jroimartin/gocui"
)

func printHelp() {
	fmt.Println("loop gconnd_ip cli_port srv_port")
}

func main() {
	if len(os.Args) < 4 {
		printHelp()
		return
	}

	g, err := ui.NewGui(ui.OutputNormal)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)
	g.SetKeybinding("", ui.KeyCtrlC, ui.ModNone, quit)

	// run lobby routine
	srvAddr := fmt.Sprintf("%s:%s", "127.0.0.1", os.Args[3])
	go lobbyLoop(srvAddr, g)

	// run client routine
	cliAddr := fmt.Sprintf("%s:%s", os.Args[1], os.Args[2])
	go clientLoop(cliAddr, g)

	// run main loop
	if err := g.MainLoop(); err != nil && err != ui.ErrQuit {
		fmt.Println(err)
		return
	}
}

func layout(g *ui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(getClientView(), 0, 0, maxX/2-1, maxY-4); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Title = getClientViewTitle()
		v.Wrap = true
		v.Autoscroll = true
	}
	if v, err := g.SetView(getLobbyView(), maxX/2, 0, maxX-1, maxY-4); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Title = getLobbyTitle()
		v.Wrap = true
		v.Autoscroll = true
	}
	name := getClientEdit()
	if v, err := g.SetView(name, 0, maxY-3, maxX-1, maxY-1); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Title = getClientEditTitle()
		v.Editable = true
		v.Wrap = true
		g.SetKeybinding(name, ui.KeyEnter, ui.ModNone, input)
	}
	return nil
}

func quit(g *ui.Gui, v *ui.View) error {
	return ui.ErrQuit
}

func input(g *ui.Gui, v *ui.View) error {
	text := v.Buffer()

	// clear the input
	v.SetCursor(0, 0)
	v.Clear()

	// send input text to client routine
	if text != "" {
		sendClientKeyboard(text)
	}
	return nil
}
