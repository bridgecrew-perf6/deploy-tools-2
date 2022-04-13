package main

import (
	"log"
	"os/exec"
)

func main() {
	for {
		// 自动拉取代码
		cmd := exec.Command("git", "pull")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		defer stdout.Close()
	}
}
