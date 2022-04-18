package main

import (
	"fmt"
	"testing"
)

func TestNewNotifyFile(t *testing.T) {
	file := NewNotifyFile()
	file.WatchPath("deploy.sh")
	fmt.Println("succ")
	select {}
}
