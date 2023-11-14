# npu plugin of telegraf
## 使用介绍
该插件代码可根据以下两种方法来使用（选择其一即可）：

### 1、源码集成使用（适合未安装Telegraf的情况）
对应官方文档：https://docs.influxdata.com/telegraf/v1.26/configure_plugins/external_plugins/write_external_plugin/
#### **编译步骤：**
拉取telegraf v1.26.0分支源码
```shell
git clone -b v1.26.0 https://github.com/influxdata/telegraf.git
```
拉取插件源码
```shell
git clone -b [latest_tag] git@gitee.com:ascend/ascend-npu-exporter.git
# [latest_tag]此tag请自行修改，建议采用仓库的最新标签，否则可能导致引用函数失效
```
将插件代码集成到telegraf源码中
```shell
cp -r path_to_telegraf_npu_plugin/plugins/inputs/npu path_to_telegraf/plugins/inputs
```
将插件注册到telegraf
```shell
cp -r path_to_telegraf_npu_plugin/plugins/inputs/all/npu.go path_to_telegraf/plugins/inputs/all
```
将telegraf源码中的Makefile里的“CGO_ENABLED=0”改为“CGO_ENABLED=1”
```shell
sed -i s"/CGO_ENABLED=0/CGO_ENABLED=1/" Makefile
```
将 “require huawei.com/npu-exporter/v5 v5.0.0-rc1.1” 和 “replace huawei.com/npu-exporter/v5 => gitee.com/ascend/ascend-npu-exporter/v5 [latest_tag]”加入到telegraf源码的go.mod的文件里
注意：[latest_tag]此tag请自行修改，建议采用仓库的最新标签并且与前面[latest_tag]一致，否则可能导致引用函数失效
然后执行
```shell
go mod tidy
```
接着编译telegraf
```shell
make all
```
运行前请先创建日志目录：（该日志是插件调用底层api将记录的日志）
```shell
mkdir -m 750 /var/log/mindx-dl/npu-exporter
```
源码集成时，该日志可通过hwlog.LogConfig{}结构体来配置，该结构体的详细信息如下
```go
type LogConfig struct {
	// log file path, default "/var/log/mindx-dl/npu-exporter/npu-plugin.log" in npu plugin
	LogFileName string
	// only write to std out, default value: false
	OnlyToStdout bool
	// only write to file, default value: false
	OnlyToFile bool
	// log level, -1-debug, 0-info, 1-warning, 2-error 3-critical default value: 0
	LogLevel int
	// size of a single log file (MB), default value: 2MB in npu plugin
	FileMaxSize int
	// MaxLineLength Max length of each log line, default value: 256
	MaxLineLength int
	// maximum number of backup log files, set as 2 in npu plugin
	MaxBackups int
	// maximum number of days for backup log files, default value: 2
	MaxAge int
	// whether backup files need to be compressed, default value: false
	IsCompress bool
	// expiration time for log cache, default value: 1s
	ExpiredTime int
	// Size of log cache space, default: 2048
	CacheSize int
}
```
#### **使用示例：**
使用插件中提供的配置文件运行telegraf
```shell
./telegraf --config path_to_plugins/inputs/npu/sample.conf
```

### 2、二进制集成，使用telegraf的execd机制（适合已安装Telegraf的情况）
对应官方文档：https://docs.influxdata.com/telegraf/v1.26/configure_plugins/external_plugins/shim/

从[MindX DL社区](https://www.hiascend.com/zh/software/mindx-dl/community)获取npu-exporter软件包，并从中解压出npu-exporter二进制文件

### 使用
运行前请先创建日志目录：（该日志是插件调用底层api将记录的日志）
```shell
mkdir -m 750 /var/log/mindx-dl/npu-exporter
```
先编写配置文件，如test.conf
```
[[inputs.execd]]
  command = ["path_to_npu_plugin/npu-exporter", "-platform=Telegraf"]
  signal = "none"

[[outputs.file]]
  file=["stdout"]
```
然后运行telegraf
```shell
./telegraf --config path_to_config_file/test.conf
```