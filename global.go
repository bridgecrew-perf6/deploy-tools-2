package main

type Parameter struct {
	Branch       string // 分支名称
	ListenerPath string // 监听目录
	TimeInterval int    // 监听时间间隔
	Start        bool   // 启动后是否立即部署一次
	Language     string // 语言
	ScriptPath   string // 脚本路径
	LogDir       string // 日志文件目录
	FileLog      bool   // 将本程序运行日志保存在日志文件中
	Help         bool   // 查看帮助信息
	Background   bool   // 后台启动
	ProjectDir   string // 项目目录 - 有些 git 仓库存在多个项目则需要指定

	ProjectName string // 项目名称

}

var (
	Args   *Parameter
	Deploy bool // 是否部署,只有部署成功后才 = false 防止多次部署
)

var (
	PID_FILE = ".pid" // pid文件路径
)
