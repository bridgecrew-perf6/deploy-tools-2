package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// 解析命令行参数
	ParseCommandVar()
	// 初始化
	Init()
	// 优雅退出程序
	quit := make(chan os.Signal) // 程序退出
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	<-quit
	zap.S().Infow("程序已退出...")
}

func Init() {
	// 切换分支
	CheckBranch(Args.Branch)
	// 监听项目部署
	go DeployProject()
	// 如果设置启动部署程序则 拉取最新代码后直接部署
	if Args.Start {
		// 自动拉取代码
		GitPull()
		zap.S().Infow("开始部署项目")
		Deploy <- true
	}

	// 监听目录中文件变动
	go GitCommitWatcher()
	if Args.Dir != "" {
		// 监听文件是否发生变化
		go FileChangeWatcher()
	}
	//定时拉取最新的 Git 提交
	go GitPullTimer()

}

// ParseCommandVar 解析命令行参数
func ParseCommandVar() {
	// 在 flag 解析之前提取 .sh 参数 防止 flag 解析失败
	sh := ArgsHasSuffixAndDelete(".sh")
	// 从 os.Args 中获取出 -d 参数并将其删除
	for index, param := range os.Args {
		// 存在 -d 参数
		if param == "-d" {
			os.Args = append(os.Args[0:index], os.Args[index+1:]...) // 删除该参数
			param := append(os.Args, "-file-log")                    // 设置日志以文件保存
			if sh != "" {
				param = append(param, sh)
			}
			pid := BackgroundRun(param)
			fmt.Printf("%s 部署工具静默启动 \n [PID] %d running...\n", os.Args[0], pid)
			fmt.Printf("启动日志请查看日志文件！\n")
			os.Exit(0)
		}
	}
	Args = &Parameter{}
	log.Println(os.Args)
	// 解析命令行参数写入注册的flag里
	flag.StringVar(&Args.Branch, "branch", "", "[可选] 指定 Git 仓库分支 默认当前分支")
	flag.IntVar(&Args.TimeInterval, "interval", 30, "自动监听 Git 仓库时间间隔(秒) 默认为30秒")
	flag.BoolVar(&Args.Start, "start", true, "默认启动程序时执行部署")
	flag.StringVar(&Args.Language, "language", "", "项目部署工具 目前支持 [go|maven|yarn|npm]")
	flag.StringVar(&Args.LogDir, "log-dir", "logs", "日志存放目录 默认在项目根目录下的 logs")
	flag.BoolVar(&Args.FileLog, "file-log", false, "将本程序运行日志保存在日志文件中")
	flag.StringVar(&Args.Dir, "dir", "", "监听目录变动，文件发生变动时执行部署脚本")
	flag.BoolVar(&Args.Help, "help", false, "查看帮助")
	flag.BoolVar(&Args.Help, "h", false, "查看帮助")
	flag.BoolVar(&Args.Background, "d", false, "后台启动")
	flag.StringVar(&Args.ProjectDir, "project-dir", "", "如果 git 仓库中存在多个项目则需指定项目目录,默认当前目录")

	flag.StringVar(&Args.ProjectName, "name", "", "项目名称 默认为当前程序运行目录名称")

	// 如果传递了脚本文件
	if sh != "" {
		Args.ScriptPath = sh
		Args.Language = "script"
	}

	flag.Parse() // 解析参数

	// 查看帮助信息
	if Args.Help {
		_, _ = fmt.Fprintf(os.Stderr, "Usage :\n %s  <shell script path> [...] \n %s -language <then project language> [...]\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	// 获取当前程序执行的目录名称
	if Args.ProjectName == "" {
		pwd, _ := os.Getwd()
		// 获取当前文件夹名称
		_, Args.ProjectName = filepath.Split(pwd)
	}

	// 初始化 Logger
	if Args.FileLog {
		InitFileLogger(Args.LogDir)
	} else {
		InitLogger()
	}

	// 设置 .pid 文件路径
	if Args.LogDir != "" {
		PID_FILE = Args.LogDir + "/" + ".pid"
	}

	if len(flag.Args()) > 0 {
		zap.S().Warnw("未解析参数", flag.Args())
	}
	zap.S().Infof("自动化部署开始，启动参数：%+v \n", Args)
	if Args.Language == "" {
		fmt.Println(123123)
		Fatalf("未指定 Shell 脚本或部署工具！")
	}
}

// InitLogger 初始化日志组件
func InitLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	// 替换全局 Logger
	zap.ReplaceGlobals(logger)
}

// InitFileLogger 生成输出到文件的 Logger
func InitFileLogger(logDir string) {

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	// 修改时间编码器
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 在日志文件中使用大写字母记录日志级别
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 设置日期格式
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.0000")
	// NewConsoleEncoder 打印更符合人们观察的方式
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 日志轮转
	write := &lumberjack.Logger{
		Filename:   logDir + "/" + "deploy-tools.log", // 日志路径名称
		MaxSize:    100,                               // 日志大小限制，单位MB
		MaxBackups: 10,                                // 最大保留历史日志数量
		MaxAge:     28,                                // 历史日志文件保留天数
		Compress:   true,                              // 历史日志文件压缩标识
	}

	writeSyncer := zapcore.AddSync(write)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)
}

// BackgroundRun 后台启动
func BackgroundRun(param []string) int {
	// 执行命令并将整理好的参数重新传递过去
	cmd := exec.Command(os.Args[0], param[1:]...)
	if err := cmd.Start(); err != nil {
		fmt.Printf("start %s failed, error: %v\n", os.Args[0], err)
		os.Exit(1)
	}
	return cmd.Process.Pid
}
