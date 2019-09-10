package main

import (
	"fmt"
	"os"
)

func printHelp() {
	fmt.Printf("getcd start getcd.xml")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "start":
		if len(os.Args) < 3 {
			printHelp()
			return
		}
		r := start(os.Args[2])
		if !r {
			fmt.Printf("start failed, find the error in log\n")
			return
		}
	default:
		printHelp()
	}
}
