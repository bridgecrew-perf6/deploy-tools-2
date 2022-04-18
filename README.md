# 自动化部署工具

# 编译

```shell
go build -o deploy-tools
```

# 参数说明

```shell
Usage :
  deploy-tools  <shell script path> [...]
  deploy-tools <then project language> [...]
```

| 参数           | 值类型    | 默认     | 说明                                 |
|--------------|--------|--------|------------------------------------|
| -d           | bool   | false  | 后台静默启动，日志输出到 log 文件中               |
| -name        | string | 当前目录名称 | 项目名称 默认为当前程序运行目录名称                 |
| -branch      | string | 当前分支   | 指定 Git 仓库分支                        |
| -language    | string |        | 项目部署工具 目前支持 [go, maven, yarn, npm] |
| -interval    | int    | 30     | 自动监听 Git 仓库时间间隔(秒) 默认为30秒          |
| -start       | bool   | true   | 程序启动时部署项目，如果不需要设置为false            |                               |                               |                                |
| -log-dir     | string | logs   | 日志存放目录 默认在项目根目录下的 logs             |
| -dir         | string |        | 监听目录变动，文件发生变动时执行部署脚本               |
| -project-dir | string |        | 如果 git 仓库中存在多个项目则需指定项目目录,默认当前目录    |
| -h && -help  | bool   | false  | 查看帮助                               |

# 示例

## Go

1. 监听当前目录下的 git 分支提交并部署

```shell
./depoly-tool -language go
```

2. 切换 dev 分支并后台启动

```shell
./depoly-tool -language go -branch dev -d
```

3. 脚本部署并设置监听目录变化

```shell
./depoly-tool depoly.sh -dir src -d
```

## Maven

1. 简单部署

```shell
 ./deploy-tools -language maven -d 
```

2. 单仓库多项目指定部署项目

```shell
 ./deploy-tools -language maven -project-dir job-api -d 
```

## Yarn & Npm & Cnpm
1. 单仓库多项目指定部署项目
```shell
./deploy-tools -language cnpm -project-dir job-admin-ui -dir job-admin-ui -d
```