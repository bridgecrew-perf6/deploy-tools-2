package main

import (
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type NotifyFile struct {
	watch   *fsnotify.Watcher
	Exclude map[string]bool
}

func NewNotifyFile() *NotifyFile {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Fatalf(err.Error())
	}
	//排除的目录 true 为排除
	excludeDir := map[string]bool{
		".git":         true,
		".idea":        true,
		"logs":         true,
		"tmp":          true,
		"node_modules": true,
		"dist":         true,
		"target":       true,
	}
	return &NotifyFile{watch: watcher, Exclude: excludeDir}
}

// WatchPath 递归监听目录或文件
func (n *NotifyFile) WatchPath(path string) {

	// 监控当前传递目录或文件
	n.watch.Add(path)
	stat, err := os.Stat(path)
	if err != nil {
		Fatalf("打开路径失败，path:%s  err:%s", path, err)
		return
	}
	// 递归监控目录下的所有子目录
	if stat.IsDir() {
		zap.S().Warnf("%#v", n.Exclude)
		n.AddListDir(path)
	}
	// 遍历所有子目录
	go n.WatchEvent()
}

// WatchEvent 监听文件目录变动
func (n *NotifyFile) WatchEvent() {
	for {
		select {
		case event, ok := <-n.watch.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				zap.S().Infof("%s 文件发生变动 \n", event.Name)
				Deploy = true
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				stat, err := os.Stat(event.Name)
				if err == nil && stat.IsDir() {
					// 排除目录
					if ex, ok := n.Exclude[stat.Name()]; !ok || (ok && !ex) {
						n.watch.Add(event.Name)
					}
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				stat, err := os.Stat(event.Name)
				if err == nil && stat.IsDir() {
					n.watch.Remove(event.Name)
				}
			}
		case err, ok := <-n.watch.Errors:
			if !ok {
				return
			}
			zap.S().Infow("error:", err)
		}
	}
}

// AddListDir 添加监听
func (n *NotifyFile) AddListDir(path string) {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		Fatalf("读取目录[%s]失败 err:%s", path, err)
	}
	for _, f := range dir {
		if f.IsDir() {
			filename := ""
			if path == "./" {
				filename = f.Name()
			} else {
				filename = path + "/" + f.Name()
			}
			if ex, ok := n.Exclude[f.Name()]; ok && ex {
				continue
			}
			abs, err := filepath.Abs(filename)
			if err != nil {
				return
			}
			n.watch.Add(abs)
			n.AddListDir(filename)
		}
	}
}
