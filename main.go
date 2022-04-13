package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	branch       string // 分支名称
	dir          string // 监听目录
	timeInterval int    // 监听时间间隔
	deploy       bool   // 启动后是否立即部署一次
)

func main() {
	ParseCommandVar()
	// 切换分支
	CheckBranch(branch)
	// 显示当前所在分支信息
	if branch == "" {
		branch = BranchName()
	}
	log.Println("当前分支为：" + branch)
	go GitCommitListener(BranchName())
	// 自动拉取代码
	// 监听文件是否发生变化
	// 阻塞进程
	//go GitPull(timeInterval)
	select {}
}

// GitCommitListener 监听 Git 提交
func GitCommitListener(branch string) {
	//headPath := ".git/logs/refs/heads/" + branch
	//getwd, _ := os.Getwd()

	headPath := ".git/logs/refs/heads/" + branch
	fmt.Println(headPath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
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
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(headPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

// ParseCommandVar 解析命令行参数
func ParseCommandVar() {
	flag.StringVar(&branch, "branch", "", "指定 Git 仓库分支")
	flag.StringVar(&dir, "dir", ".git", "监听目录变动，文件发生变动时执行部署脚本")
	flag.IntVar(&timeInterval, "", 5, "自动监听 Git 仓库时间间隔(秒) 默认为30秒")
	flag.BoolVar(&deploy, "d", false, "(true|false) true：启动程序时执行部署脚本")
	// 解析命令行参数写入注册的flag里
	flag.Parse()
	log.Printf("自动化部署开始，启动参数：分支:%s, dir:%s, t(timeInterval):%d秒", branch, dir, timeInterval)
}

// BranchName 获取当前分支名称
func BranchName() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("ERROR --- 获取当前分支失败：", err)
	}
	cmd.Start()
	name := ConverseStd(stdout)
	cmd.Wait()
	return name
}

// GitPull 拉取代码
func GitPull(timeInterVal int) {
	for {
		log.Println("Git pulls the latest code")
		cmd := exec.Command("git", "pull")
		stdout, stdoutErr := cmd.StdoutPipe()
		stderr, stderrErr := cmd.StderrPipe()
		if stdoutErr != nil {
			log.Fatal("stdoutErr", stdoutErr)
		}
		if stderrErr != nil {
			log.Fatal("stderrErr", stderrErr)
		}

		if err := cmd.Start(); err != nil { // 执行命令
			log.Fatal(err)
		}
		// 读取管道中的内容
		if str := ConverseStd(stdout); str != "" {
			log.Println("INFO ---", str)
		}
		if str := ConverseStd(stderr); str != "" {
			log.Println("ERROR ---", str)
		}
		cmd.Wait() // 等待命令运行结束
		time.Sleep(time.Duration(timeInterVal) * time.Second)
	}

}

// ConverseStd 转换管道中的内容
func ConverseStd(std io.ReadCloser) (stdStr string) {
	defer std.Close()
	// 读取管道中的内容
	result, err := io.ReadAll(std)
	if err != nil {
		log.Fatal(err)
	}
	return string(result)
}

// CheckBranch 切换分支
func CheckBranch(branch string) {
	if branch == "" {
		return
	}
	branchCmd := exec.Command("git", "checkout", branch)
	stdout, stdoutErr := branchCmd.StdoutPipe()
	stderr, stderrErr := branchCmd.StderrPipe()
	if stdoutErr != nil {
		log.Fatal("stdoutErr", stdoutErr)
	}
	if stderrErr != nil {
		log.Fatal("stderrErr", stderrErr)
	}
	if err := branchCmd.Start(); err != nil { // 执行命令并等待命令执行完毕
		log.Fatal(err)
	}
	log.Println(ConverseStd(stdout))
	if er := ConverseStd(stderr); strings.Contains(er, "error") {
		log.Fatal("ERROR --- 切换分支错误：", er)
	}
	branchCmd.Wait()
	log.Printf("切换 %s 分支成功！\n", branch)
}
