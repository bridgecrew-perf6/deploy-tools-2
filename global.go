package main

import "go.uber.org/zap"

type Parameter struct {
	Branch       string // 分支名称
	Dir          string // 监听目录
	TimeInterval int    // 监听时间间隔
	StartProject bool   // 启动后是否立即部署一次
	Language     string // 语言
	ScriptPath   string // 脚本路径
	ProjectName  string // 项目名称
	LogDir       string // 日志文件目录
}

var (
	Args          *Parameter
	ZapFileLogger *zap.SugaredLogger
	Deploy        = make(chan bool)
)

const (
	PID_FILE = ".pid" // pid 文件
)
