package main

import (
	"fmt"
	"os"

	ui "github.com/jroimartin/gocui"
)

func printHelp() {
	fmt.Println("vmcli gconnd_ip gconnd_port ghost_ip ghost_port")
}

func main() {
	if len(os.Args) < 5 {
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

	// run vm routine
	hostAddr := fmt.Sprintf("%s:%s", os.Args[3], os.Args[4])
	go vmLoop(hostAddr, g)

	// run client routine
	connAddr := fmt.Sprintf("%s:%s", os.Args[1], os.Args[2])
	go clientLoop(connAddr, g)

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
	if v, err := g.SetView(getVMView(), maxX/2, 0, maxX-1, maxY-4); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Title = getVMTitle()
		v.Wrap = true
		v.Autoscroll = true
	}
	name := getClientEdit()
	if v, err := g.SetView(name, 0, maxY-3, maxX/2-1, maxY-1); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Title = getClientEditTitle()
		v.Editable = true
		v.Wrap = true
		g.SetKeybinding(name, ui.KeyEnter, ui.ModNone, input)
	}
	name = getVMEdit()
	if v, err := g.SetView(name, maxX/2, maxY-3, maxX-1, maxY-1); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Title = getVMEditTitle()
		v.Editable = true
		v.Wrap = true
		g.SetKeybinding(name, ui.KeyEnter, ui.ModNone, inputForVM)
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

func inputForVM(g *ui.Gui, v *ui.View) error {
	text := v.Buffer()

	// clear the input
	v.SetCursor(0, 0)
	v.Clear()

	// send input text to client routine
	if text != "" {
		sendVMKeyboard(text)
	}
	return nil
}
