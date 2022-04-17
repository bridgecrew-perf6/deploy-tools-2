package main

import (
	"go.uber.org/zap"
	"os"

	"github.com/fsnotify/fsnotify"
)

// FileChangeWatcher 监听文件目录变动
func FileChangeWatcher() {
	// 获取当前 分支的 head 信息文件
	getwd, _ := os.Getwd()
	path := getwd + "/" + Args.Dir

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Fatalf(err.Error())
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				zap.S().Infof("event:%v \n", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					zap.S().Infof("%s 文件发生变动 \n", Args.Dir)
					Deploy <- true
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				zap.S().Infow("error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		Fatalf(err.Error())
	}
	<-done
}
