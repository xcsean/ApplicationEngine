package main

import (
	"fmt"
	"os"
	"strconv"
)

func printHelp() {
	fmt.Printf("gconnd start gconnd.xml id\n")
}

func main() {
	if len(os.Args) < 4 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "start":
		c, err := newConfig(os.Args[2])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		id, err := strconv.ParseInt(os.Args[3], 10, 64)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		// validate the start argument
		id2, err := c.GetID()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if id2 != id {
			fmt.Printf("start id=%d not equal the id in division=%s", id, c.Division)
			return
		}
		start(c, id)
	default:
		printHelp()
		return
	}
}
