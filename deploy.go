package main

import (
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DeployProject 项目部署
func DeployProject() {
	for {
		// 如果不需要部署
		if !Deploy {
			time.Sleep(1 * time.Second)
			continue
		}
		// 如果已脚本的方式启动
		switch Args.Language {
		case "script":
			ScriptDeploy()
		case "go":
			GoDeploy()
		case "maven":
			MavenDeploy()
		case "npm":
			NodeDeploy("npm")
		case "cnpm":
			NodeDeploy("cnpm")
		case "yarn":
			NodeDeploy("yarn")
		default:
			if Args.Language == "" {
				Fatalf("未指定 Shell 脚本或部署工具！")
			} else {
				Fatalf("暂时不支持 %s 工具部署，请使用 Shell 脚本启动项目！", Args.Language)
			}
		}
		// 本次部署结束
		Deploy = false
		zap.S().Infow("等待下一次代码变动..")
	}
}

// ScriptDeploy 脚本部署
func ScriptDeploy() {
	// 判断脚本文件事否存在
	if exists := FileExists(Args.ScriptPath); !exists {
		Fatalf("%s 脚本文件不存在", Args.ScriptPath)
	}
	zap.S().Infof("使用 %s 脚本部署项目 \n", Args.ScriptPath)
	cmd := exec.Command("bash", Args.ScriptPath)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		Fatalf("ERROR --- Bash: 执行命令失败：%s", err)
	}
	cmd.Start()
	if str := ConverseStd(stdout); str != "" {
		zap.S().Infow(str)
	}
	if str := ConverseStd(stderr); str != "" {
		zap.S().Infow(str)
	}
	cmd.Wait()
	zap.S().Infof("Bash: %s 脚本全部执行完成 \n", Args.ProjectName)

}

func MavenDeploy() {
	zap.S().Infow("Maven 部署项目")
	var (
		jarName string // jar 包名称
		jarPath string // jar 包路径
		pomPath string // pom.xml 路径
		pack    string // 打包命令
		clean   string // 清理命令
		target  string // target目录
	)
	// 如果没有制指定项目目录则默认当前目录
	if Args.ProjectDir == "" {
		pomPath = "pom.xml"
		clean = "mvn clean"
		pack = "mvn package"
		target = "target"
	} else {
		pomPath = Args.ProjectDir + "/pom.xml"
		clean = fmt.Sprintf("mvn -f %s clean", Args.ProjectDir)
		pack = fmt.Sprintf("mvn -f %s package", Args.ProjectDir)
		target = fmt.Sprintf("%s/target", Args.ProjectDir)
	}

	// 判断当前是否是 maven 项目
	if exists := FileExists(pomPath); !exists {
		Fatalf("%s 不存在 请检查项目路径是否正确！", pomPath)
	}

	// 清理
	cleanCmd := exec.Command("/bin/bash", "-c", clean)
	cleanStdout, _ := cleanCmd.StdoutPipe()
	cleanStderr, _ := cleanCmd.StderrPipe()
	if err := cleanCmd.Start(); err != nil {
		Fatalf("%s 清理命令执行失败", clean)
	}
	if msg := ConverseStd(cleanStdout); msg != "" {
		zap.S().Infow(msg)
		if !strings.Contains(msg, "BUILD SUCCESS") {
			Fatalf("\n Error %s 清理程序发生错误", clean)
		}
	}
	if msg := ConverseStd(cleanStderr); msg != "" {
		zap.S().Infow(msg)
	}
	cleanCmd.Wait()

	// 打包
	zap.S().Infof("正在打包 %s \n", pack)
	packCmd := exec.Command("/bin/bash", "-c", pack)
	packStdout, _ := packCmd.StdoutPipe()
	packStderr, _ := packCmd.StderrPipe()
	if err := packCmd.Start(); err != nil {
		Fatalf("%s 打包命令执行失败", pack)

	}
	if msg := ConverseStd(packStdout); msg != "" {
		zap.S().Infow(msg)
		if !strings.Contains(msg, "BUILD SUCCESS") {
			Fatalf("\n Error %s 打包程序发生错误", pack)
		}
	}
	if msg := ConverseStd(packStderr); msg != "" {
		zap.S().Infow(msg)
	}
	packCmd.Wait()

	zap.S().Infof("Maven: %s 打包成功 \n", Args.ProjectDir)

	// 获取打包成功的 jar 包名称
	files, err := ioutil.ReadDir(target)
	if err != nil {
		Fatalf("读取 %s 目录失败！", target)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".jar") {
			jarName = f.Name()
			jarPath = target + "/" + f.Name()
			break
		}
	}
	if jarName == "" {
		Fatalf("获取 %s 目录下 jar 包失败！", target)
	}

	// 结束之前启动的进程
	KillPidFile(jarName)

	// 创建日志目录
	os.MkdirAll(Args.LogDir, os.ModePerm)

	zap.S().Infow("项目启动中...")
	// 后台启动程序并返回 pid 至 .pid 文件
	//nohup java -jar sprint-boot.jar > logs/sprint-boot.log & 2>&1 & echo $!
	nphub := fmt.Sprintf("nohup java -jar %s > logs/%s-%d.log 2>&1 & echo $! > %s", jarPath, jarName, time.Now().Unix(), PID_FILE)
	startCmd := exec.Command("/bin/bash", "-c", nphub)
	if err := startCmd.Start(); err != nil {
		Fatalf("ERROR --- 启动 %s 失败，Error：%s \n", jarPath, err)
	}
	startCmd.Wait()

	// 判断 pid 进程是否存活
	running := CheckPidFileIsRunning()
	if !running {
		Fatalf("ERROR --- 启动 %s 失败， 程序已停止运行，请查看日志检查错误。 \n", jarPath)
	}
	zap.S().Infof("Maven: %s 项目启动成功! \n", jarPath)

	// 监听 pid 进程是否正常 如果被关闭则结束当前程序
	go PidFileListener(jarPath)

}

// GoDeploy 部署 go 项目
func GoDeploy() {
	zap.S().Infow("Go 部署项目...")
	var cmd *exec.Cmd
	// 如果没有制指定项目目录则默认当前目录
	if Args.ProjectDir == "" {
		cmd = exec.Command("go", "build", "-o", Args.ProjectName)
	} else {
		cmd = exec.Command("go", "build", "-o", Args.ProjectName, Args.ProjectDir)
	}
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		Fatalf("ERROR --- Go Build:执行命令失败：%s", err)
	}
	cmd.Start()
	if str := ConverseStd(stdout); str != "" {
		zap.S().Infow(str)
	}
	if str := ConverseStd(stderr); str != "" {
		zap.S().Infow(str)
	}
	cmd.Wait()
	zap.S().Infof("Go: %s 编译成功 \n", Args.ProjectName)

	// 结束之前启动的进程
	KillPidFile(Args.ProjectName)

	// 创建日志目录
	os.MkdirAll(Args.LogDir, os.ModePerm)

	zap.S().Infow("项目启动中...")
	// 后台启动程序并返回 pid 至 .pid 文件
	// nohup ./consult-im-api > logs/consult-im.log & 2>&1 & echo $!
	nphub := fmt.Sprintf("nohup ./%s > logs/%s-%d.log 2>&1 & echo $! > %s", Args.ProjectName, Args.ProjectName, time.Now().Unix(), PID_FILE)
	startCmd := exec.Command("/bin/bash", "-c", nphub)
	if err := startCmd.Start(); err != nil {
		Fatalf("ERROR --- 启动 %s 失败，Error：%s \n", Args.ProjectName, err)
	}
	startCmd.Wait()
	// 判断 pid 进程是否存活
	running := CheckPidFileIsRunning()
	if !running {
		Fatalf("ERROR --- 启动 %s 失败，程序已停止运行，请查看日志检查错误。 \n", Args.ProjectName)
	}
	zap.S().Infof("Go: %s 项目启动成功! \n", Args.ProjectName)
	// 监听 pid 进程是否正常 如果被关闭则结束当前程序
	go PidFileListener(Args.ProjectName)
}

// NodeDeploy Node 项目部署
func NodeDeploy(tool string) {
	zap.S().Infof("%s 部署项目...\n", tool)
	// 如果没有制指定项目目录则默认当前目录
	pack := "package.json"
	install := tool + " install"

	if Args.ProjectDir != "" {
		pack = Args.ProjectDir + "/" + "package.json"
		install = fmt.Sprintf("cd %s && %s", Args.ProjectDir, install)
	}
	fmt.Println("package", pack)
	if exists := FileExists(pack); !exists {
		Fatalf("%s 不存在，请检查项目目录是否正确！", pack)
	}

	zap.S().Infof("正在安装 package.json  %s \n", Args.ProjectDir)
	// 安装 Node 依赖
	installCmd := exec.Command("/bin/bash", "-c", install)
	stdout, _ := installCmd.StdoutPipe()
	stderr, _ := installCmd.StderrPipe()
	defer stdout.Close()
	if err := installCmd.Start(); err != nil {
		Fatalf("ERROR --- %s install 执行命令失败：%s", tool, err)
	}
	go func() {
		for {
			tmp := make([]byte, 1024)
			_, err := stdout.Read(tmp)
			zap.S().Infof(string(tmp))
			if err != nil {
				break
			}
		}
	}()
	go func() {
		for {
			tmp := make([]byte, 1024)
			_, err := stderr.Read(tmp)
			zap.S().Infof(string(tmp))
			if err != nil {
				break
			}
		}
	}()
	installCmd.Wait()
	zap.S().Infof("install 已完成 %s \n", Args.ProjectDir)
	// 创建日志目录
	os.MkdirAll(Args.LogDir, os.ModePerm)

	zap.S().Infow("正在打包...")
	nphub := ""
	if Args.ProjectDir != "" {
		nphub = fmt.Sprintf("cd %s && %s run build ", Args.ProjectDir, tool)
	} else {
		nphub = fmt.Sprintf("%s run build", tool)
	}
	fmt.Println(nphub)
	startCmd := exec.Command("/bin/bash", "-c", nphub)
	startStd, _ := startCmd.StdoutPipe()
	startStdErr, _ := startCmd.StderrPipe()
	if err := startCmd.Start(); err != nil {
		Fatalf("ERROR --- %s build 打包失败，Error：%s \n", tool, err)
	}
	go func() {
		for {
			tmp := make([]byte, 1024)
			_, err := startStd.Read(tmp)
			zap.S().Infof(string(tmp))
			if err != nil {
				break
			}
		}
	}()
	go func() {
		for {
			tmp := make([]byte, 1024)
			_, err := startStdErr.Read(tmp)
			zap.S().Infof(string(tmp))
			if err != nil {
				break
			}
		}
	}()
	startCmd.Wait()
	zap.S().Infof("%s run build 打包成功！", tool)
}

// PidFileListener 监听 pid 文件进程是否已被停止 如果停止则结束监听
func PidFileListener(name string) {
	for {
		running := CheckPidFileIsRunning()
		if !running {
			Fatalf("%s 项目运行异常，进程已停止。%s 结束监听", name, os.Args[0])
		}
		time.Sleep(1 * time.Second)
	}
}
