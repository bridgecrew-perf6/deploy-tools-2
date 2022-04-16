package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 解析命令行参数
	ParseCommandVar()
	// 初始化 Logger
	InitLogger()
	// 初始化
	Init()
	// 阻塞进程
	select {}
}

func Init() {
	// 切换分支
	CheckBranch(Args.Branch)
	// 监听项目部署
	go DeployProject()

	// 如果设置启动部署程序则 拉取最新代码后直接部署
	if Args.StartProject {
		// 自动拉取代码
		GitPull()
		zap.S().Infow("开始部署项目")
		Deploy <- true
	}
	// 监听文件是否发生变化
	go GitCommitWatcher()
	//定时拉取最新的 Git 提交
	go GitPullTimer()

}

// ParseCommandVar 解析命令行参数
func ParseCommandVar() {
	_, _ = fmt.Fprintf(os.Stderr, "Usage :\n %s <shell script path> [...] \n %s -language <then project language> [...]\n", os.Args[0], os.Args[0])

	Args = &Parameter{}
	flag.StringVar(&Args.Branch, "branch", "", "[可选]指定 Git 仓库分支")
	flag.IntVar(&Args.TimeInterval, "interval", 5, "[可选]自动监听 Git 仓库时间间隔(秒) 默认为30秒")
	flag.BoolVar(&Args.StartProject, "start", false, "启动程序时执行部署脚本")
	flag.StringVar(&Args.Language, "language", "", "项目使用的语言")
	flag.StringVar(&Args.LogDir, "log-dir", "logs", "[可选] 日志存放目录 默认在项目根目录下的 logs")
	flag.StringVar(&Args.ProjectName, "name", "", "[可选] 项目名称 默认为当前程序运行目录名称")
	flag.StringVar(&Args.Dir, "dir", "", "[可选]监听目录变动，文件发生变动时执行部署脚本")
	// 解析命令行参数写入注册的flag里
	flag.Parse()
	fmt.Printf("自动化部署开始，启动参数：%#v \n", Args)
	pwd, _ := os.Getwd()
	// 获取当前程序执行的目录名称
	if Args.ProjectName == "" {
		// 获取当前文件夹名称
		_, Args.ProjectName = filepath.Split(pwd)
	}
	for _, param := range flag.Args() {
		// 如果传入了 .sh 文件则使用脚本部署 忽略设置的语言
		if strings.HasSuffix(param, ".sh") {
			Args.ScriptPath = param
			Args.Language = "script"
			Deploy <- true
			break
		}

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
	// 将项目部署的日志写入日志文件中
	ZapSugarFile(fmt.Sprintf("%s/%s.log", Args.LogDir, Args.ProjectName))
}

// ZapSugarFile 生成输出到文件的 Looger
func ZapSugarFile(filePath string) {

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
		Filename:   filePath, // 日志路径名称
		MaxSize:    100,      // 日志大小限制，单位MB
		MaxBackups: 10,       // 最大保留历史日志数量
		MaxAge:     28,       // 历史日志文件保留天数
		Compress:   true,     // 历史日志文件压缩标识
	}

	writeSyncer := zapcore.AddSync(write)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	// zap.AddCaller()  添加将调用函数信息记录到日志中的功能。
	ZapFileLogger = zap.New(core, zap.AddCaller()).Sugar()
}
