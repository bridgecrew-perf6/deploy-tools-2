package main

import (
	"bufio"
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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
		Fatalf(err.Error())
	}
	return string(result)
}

// KillPidFile 结束 .pid 记录的进程
func KillPidFile(processName string) {
	pids := GetPidByProcessName(processName)
	_, err := os.Stat(PID_FILE)
	// .pid 文件不存在
	if err != nil {
		if len(pids) > 0 {
			zap.S().Infof(".pid 文件不存在，但有疑似进程正在运行，请自行检查，并终止 pid：%v", pids)
		}
		return
	}
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

// GetPidByProcessName 通过进程名获取到 pid 列表
func GetPidByProcessName(name string) []string {
	pidCmd := exec.Command("/bin/bash", "-c", fmt.Sprintf(`ps -ef | grep "%s" | grep -v grep | awk '{print $2}'`, name))
	stdout, err := pidCmd.StdoutPipe()
	if err != nil {
		return []string{}
	}
	defer stdout.Close()
	if err := pidCmd.Start(); err != nil {
		return []string{}
	}
	text := ConverseStd(stdout)
	list := strings.Split(strings.Trim(text, "\n"), "\n")
	pidCmd.Wait()
	return list
}

// ArgsHasSuffixAndDelete 获取 os.Args 中的指定值并删除
func ArgsHasSuffixAndDelete(value string) string {
	for index, param := range os.Args {
		// 如果传入了 .sh 文件则使用脚本部署 忽略设置的语言
		if strings.HasSuffix(param, value) {
			os.Args = append(os.Args[0:index], os.Args[index+1:]...)
			return param
		}
	}
	return ""
}

// ProcessIsRunningByPid 检查 pid 是否在运行
func ProcessIsRunningByPid(pid int) bool {
	if pid == 0 {
		return false
	}
	// 返回一个 pid 信息
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// 发送 signal 信息判断进程是否运行
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false
	}
	return true
}

// CheckPidFileIsRunning 验证 pid 文件中的 进程是否在运行
func CheckPidFileIsRunning() bool {
	open, err := os.Open(PID_FILE)
	if err != nil {
		zap.S().Warnw("打开 %s 文件失败 Error: %s \n", PID_FILE, err)
	}
	pidByte, err := io.ReadAll(open)
	if err != nil {
		zap.S().Warnw("读取 %s 文件失败 Error: %s \n", PID_FILE, err)
	}
	pidStr := strings.Trim(string(pidByte), "\n")
	pid, _ := strconv.Atoi(pidStr)
	return ProcessIsRunningByPid(pid)
}

// Fatalf 封装错误退出
func Fatalf(msg string, args ...interface{}) {
	zap.S().Fatalf(msg+"\n --- 程序异常结束 ---", args...)
}

// ReadPipe 时实读取管道中内容
func ReadPipe(stdout io.ReadCloser) {
	defer stdout.Close()
	tmp := make([]byte, 1024)
	for {
		reader := bufio.NewReader(stdout)
		n, err := reader.Read(tmp)
		if err != nil {
			if err == io.EOF {
				break
			}
			if n == 0 {
				break
			}
			zap.S().Warnf("读取管道内容失败 n=%d err:%s", n, err)
			continue
		}
		if str := string(tmp[:n]); str != "" {
			zap.S().Infof(str)
		}
	}
}
