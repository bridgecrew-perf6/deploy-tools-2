package main

import (
	"bufio"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"strings"
)

// FileExists 目录是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// ConverseStd 转换管道中的内容
func ConverseStd(std io.ReadCloser) (stdStr string) {
	defer std.Close()
	// 读取管道中的内容
	result, err := io.ReadAll(std)
	if err != nil {
		zap.S().Fatal(err)
	}
	return string(result)
}

// ReadPipeOutFile 持续读区取管道中的数据并输出到日志文件中
func ReadPipeOutFile(std io.ReadCloser) {
	for {
		if std == nil {
			continue
		}
		in := bufio.NewScanner(std)
		for in.Scan() {
			ZapFileLogger.Infow(string(in.Bytes()))
		}
	}
}

// KillPidFile 结束 .pid 记录的进程
func KillPidFile() {
	_, err := os.Stat(PID_FILE)
	// .pid 文件存在
	if err == nil {
		open, err := os.Open(PID_FILE)
		if err != nil {
			zap.S().Warnw("打开 %s 文件失败 Error: %s \n", PID_FILE, err)
		}
		pidByte, err := io.ReadAll(open)
		if err != nil {
			zap.S().Warnw("读取 %s 文件失败 Error: %s \n", PID_FILE, err)
		}
		pid := strings.Trim(string(pidByte), "\n")

		// 通过 进程名获取到 pid 列表
		pids := GetPidByProcessName(Args.ProjectName)
		for _, p := range pids {
			// 如果正在运行中的 pid = pid 文件内容则结束进程
			if p == pid {
				killCmd := exec.Command("kill", pid)
				zap.S().Infof("结束 pid % 进程 \n", pid)
				killCmd.Run()
			}
		}

		// 删除进程文件
		zap.S().Infof("删除 %s 进程文件 \n", PID_FILE)
		delCmd := exec.Command("rm", "-f", PID_FILE)
		delCmd.Run()
	}
}

// GetPidByProcessName 通过进程名获取到 pid 列表
func GetPidByProcessName(name string) []string {
	pidCmd := exec.Command("pidof", name)
	stdout, err := pidCmd.StdoutPipe()
	if err != nil {
		return []string{}
	}
	defer stdout.Close()
	if err := pidCmd.Start(); err != nil {
		return []string{}
	}
	text := strings.Trim(ConverseStd(stdout), "\n")
	pidCmd.Wait()
	list := strings.Split(text, " ")
	return list
}
