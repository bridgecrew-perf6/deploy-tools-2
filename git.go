package main

import (
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GitCommitWatcher 监听 Git 提交变更
func GitCommitWatcher() {
	// 获取当前 分支的 head 信息文件
	getwd, _ := os.Getwd()
	headPath := getwd + "/test"
	//headPath := getwd + "/.git/logs/refs/heads/" + Args.Branch

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.S().Fatal(err)
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
				zap.S().Infow("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					zap.S().Infof("git commit %s modified\n", Args.Branch)
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

	err = watcher.Add(headPath)
	if err != nil {
		zap.S().Fatal(err)
	}
	<-done
}

// GitPullTimer 定时拉取代码
func GitPullTimer() {
	for {
		GitPull()
		time.Sleep(time.Duration(Args.TimeInterval) * time.Second)
	}

}

// GitPull 拉取代码
func GitPull() {
	zap.S().Infow("Git pulls the latest code")
	cmd := exec.Command("git", "pull")
	stdout, stdoutErr := cmd.StdoutPipe()
	stderr, stderrErr := cmd.StderrPipe()
	if stdoutErr != nil {
		zap.S().Fatal("stdoutErr", stdoutErr)
	}
	if stderrErr != nil {
		zap.S().Fatal("stderrErr", stderrErr)
	}

	if err := cmd.Start(); err != nil { // 执行命令
		zap.S().Fatal(err)
	}
	// 读取管道中的内容
	if str := ConverseStd(stdout); str != "" {
		zap.S().Infof("INFO --- %s", str)
	}
	if str := ConverseStd(stderr); str != "" {
		zap.S().Infof("ERROR --- %s", str)
	}
	cmd.Wait() // 等待命令运行结束

}

// CheckBranch 切换分支
func CheckBranch(branch string) {
	if branch == "" {
		// 获取并设置当前分支名称
		Args.Branch = BranchName()
		return
	}
	branchCmd := exec.Command("git", "checkout", branch)
	stdout, stdoutErr := branchCmd.StdoutPipe()
	stderr, stderrErr := branchCmd.StderrPipe()
	if stdoutErr != nil {
		zap.S().Fatal("stdoutErr", stdoutErr)
	}
	if stderrErr != nil {
		zap.S().Fatal("stderrErr", stderrErr)
	}
	if err := branchCmd.Start(); err != nil { // 执行命令并等待命令执行完毕
		zap.S().Fatal()
		zap.S().Fatal(err)
	}
	zap.S().Infow(ConverseStd(stdout))
	if er := ConverseStd(stderr); strings.Contains(er, "error") {
		zap.S().Fatal("ERROR --- 切换分支错误：", er)
	}
	branchCmd.Wait()
	zap.S().Infof("切换 %s 分支成功！\n", branch)
	zap.S().Infow("当前分支为：" + Args.Branch)
}

// BranchName 获取当前分支名称
func BranchName() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		zap.S().Fatal("ERROR --- 获取当前分支失败：", err)
	}
	cmd.Start()
	name := ConverseStd(stdout)
	cmd.Wait()
	return strings.Trim(name, "\n")
}
