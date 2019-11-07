package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func printHelp() {
	fmt.Println("inspector inspector.xml")
}

var (
	content string
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	// save config
	c, err := newConfig(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	config = c

	// init the directory
	os.MkdirAll(config.getDir(), os.ModePerm)

	// init the startup content
	cont, err := getContent()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cont = pickContent(cont)
	saveContent(cont, true)

	// loop for inspect the target
	tick := time.NewTicker(time.Duration(config.getInterval()) * time.Second)
	for {
		select {
		case <-tick.C:
			cont, err = getContent()
			if err == nil {
				cont = pickContent(cont)
				saveContent(cont, false)
			}
		}
	}
}

func getContent() (string, error) {
	rsp, err := http.Get(config.getCmd())
	if err != nil {
		return "", err
	}

	cont, err := ioutil.ReadAll(rsp.Body)
	rsp.Body.Close()
	if err != nil {
		return "", err
	}

	return string(cont[:]), nil
}

func pickContent(cont string) string {
	cont2 := ""
	arr1 := strings.Split(cont, config.getPrefix())
	if len(arr1) == 1 {
		cont2 = arr1[0]
	} else if len(arr1) == 0 {
		cont2 = cont
	} else {
		cont2 = arr1[1]
	}

	arr2 := strings.Split(cont2, config.getPostfix())
	if len(arr2) == 1 {
		return arr2[0]
	}

	if len(arr2) == 0 {
		return cont2
	}

	return arr2[0]
}

func saveContent(cont string, startup bool) error {
	if cont != content {
		content = cont
		fileName := config.getDir() + "/" + time.Now().Format("2006_01_02_15_04_05")
		if startup {
			fileName += "_startup"
		}

		// write the content to file
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer f.Close()
		f.WriteString(cont)
	}
	return nil
}
