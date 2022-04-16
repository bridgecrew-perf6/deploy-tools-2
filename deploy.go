package main

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"time"
)

// DeployProject 项目部署
func DeployProject() {
	for {
		select {
		case <-Deploy:
			// 如果已脚本的方式启动
			switch Args.Language {
			case "script":
				ScriptDeploy()
			case "go":
				GoDeploy()
			case "maven":
			case "npm":
			case "yarn":
			default:
				zap.S().Fatalf("暂时不支持 %s 项目，请使用 Shell 启动项目！", Args.Language)
			}
		}
	}

}

func ScriptDeploy() {
	zap.S().Infof("使用 %s 脚本部署项目 \n", Args.ScriptPath)
}

func MavenDeploy() {
	zap.S().Infow("Maven 部署项目")
}

func YarnDeploy() {
	zap.S().Infow("Yarn 部署项目")
}

func NPMDeploy() {
	zap.S().Infow("Npm 部署项目")
}

// GoDeploy 部署 go 项目
func GoDeploy() {
	zap.S().Infow("Go 部署项目...")
	cmd := exec.Command("go", "build", "-o", Args.ProjectName)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		zap.S().Fatal("ERROR --- Go Build:执行命令失败：", err)
	}
	cmd.Start()
	fmt.Println(ConverseStd(stdout))
	fmt.Println(ConverseStd(stderr))
	cmd.Wait()
	zap.S().Infof("Go: %s 编译成功 \n", Args.ProjectName)

	// 结束之前启动的进程
	KillPidFile()

	// 创建日志目录
	os.MkdirAll(Args.LogDir, os.ModePerm)

	zap.S().Infow("项目启动中...")
	// 后台启动程序并返回 pid 至 .pid 文件
	// nohup ./consult-im-api > logs/consult-im.log & 2>&1 & echo $!
	nphub := fmt.Sprintf("nohup ./%s > logs/%s-%d.log 2>&1 & echo $! > %s", Args.ProjectName, Args.ProjectName, time.Now().Unix(), PID_FILE)
	startCmd := exec.Command("/bin/bash", "-c", nphub)
	if err := startCmd.Start(); err != nil {
		zap.S().Fatalf("ERROR --- 启动 %s 失败，Error：%s \n", Args.ProjectName, err)
	}
	zap.S().Infof("Go: %s 项目启动成功! \n", Args.ProjectName)
	startCmd.Wait()
}