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
	fmt.Println("inspector inspector.xml [test]")
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

	// check whether it's in test mode or not
	if len(os.Args) == 3 && os.Args[2] == "test" {
		cont, err := getContent()
		if err == nil {
			fmt.Println(matchContent(cont))
		} else {
			fmt.Println(err.Error())
		}
		return
	}

	// init the startup content
	cont, err := getContent()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cont = matchContent(cont)
	saveContent(cont, true)

	// loop for inspect the target
	tick := time.NewTicker(time.Duration(config.getInterval()) * time.Second)
	for {
		select {
		case <-tick.C:
			cont, err = getContent()
			if err == nil {
				cont = matchContent(cont)
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

func matchContent(cont string) string {
	content := ""
	keywords := config.getKeywords()
	ss := strings.Split(cont, "\n\n")
	for i := 0; i < len(ss); i++ {
		s := ss[i]
		for j := 0; j < len(keywords); j++ {
			keyword := keywords[j]
			if strings.Contains(s, keyword) {
				content += s
				content += "\n\n"
			}
		}
	}
	return content
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
