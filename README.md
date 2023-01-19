# NPU-Exporter
-   [NPU-Exporter介绍](#NPU-Exporter介绍.md)
-   [快速上手](#快速上手.md)
-   [环境依赖](#环境依赖.md)
-   [目录结构](#目录结构.md)
-   [版本更新信息](#版本更新信息.md)
-   [附录](#附录.md)
##NPU-Exporter介绍

Prometheus（普罗米修斯）是一个开源的系统监控和警报工具包，Exporter就是专门为Prometheus提供数据源的组件。由于Prometheus社区的活跃和大量的使用，已经有很多厂商或者服务提供了Exporter，如Prometheus官方的Node Exporter，MySQL官方出的MySQL Server Exporter和NVIDA的NVIDIA GPU Exporter。这些Exporter负责将特定监控对象的指标，转成Prometheus能够识别的数据格式，供Prometheus集成。NPU-Expoter是华为自研的专门收集华为NPU各种监控信息和指标，并封装成Prometheus专用数据格式的一个服务组件。

##快速上手

#### 编译前准备

-   确保PC机与互联网网络互通，并已完成Git和Docker的安装。请参见：[Git安装](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)，[Docker-ce安装](https://docs.docker.com/engine/install/ubuntu/)进行安装。

-   已完成Go语言环境的安装（版本\>1.13，建议使用最新的bugfix版本）。请参见（[https://golang.org/](https://golang.org/)）。
-   根据所在网络环境配置Go代理地址，国内可使用**Goproxy China**，例如：

    ```
    go env -w GOPROXY=https://goproxy.cn,direct
    ```
#### 编译和启动NPU-Exporter

1. 下载好源代码，执行以下命令，编译NPU-Exporter。

    **cd build**

    **bash build.sh**

    编译生成的文件在源码根目录下的output目录，文件[表1](#table1860618363516)所示。

    **表 1**  编译生成的文件列表

    |文件名 |说明 |
    | ---- | ---- |
    | npu-exporter |二进制文件  |
    | npu-exporter-{version}.yaml |K8s启动yaml|
    | npu-exporter-310P-1usoc-{version}.yaml |1usoc场景K8s启动yaml|
    | Dockerfile |常规镜像制作配置文件  |
    | Dockerfile-310P-1usoc | 1usoc场景镜像制作专用配置文件  |
    | run_for_310P_1usoc.sh | 1usoc场景启动脚本  |

    >! **说明：** 
    >
    >- _\{version\}_：表示版本号，请根据实际写入。

2. 镜像制作
    确保执行步骤一，生成相应的二进制文件和Dockerfile文件
    cd output
    构建镜像（镜像名为）
    docker build --no-cache -t npu-exporter:v3.0.0 ./
    分发镜像（
    
    2.1 先保存为压缩包
    
        docker save hccl-controller:v3.0.0 > hccl-controller-v3.0.0-linux-arrch64.tar
    
    2.2 然后分发到各节点
    
        scp hccl-controller-v3.0.0-linux-arrch64.tar root@{目标节点IP地址}:保存路径
    
    2.3 节点上加载镜像
    
        docker load < hccl-controller-v3.0.0-linux-arrch64.tar 
    
3. 执行以下命令，启动NPU-Exporter。

    **表 2**  操作命令
    
    | 启动类型            | 启动命令                                                     |
    | :-------------------- | -------------------------------------------------------- |
    | 二进制启动          | ./npu-exporter   -ip=0.0.0.0                                          |
    | K8s启动（推荐方式） | kubectl apply -f npu-exporter-{version}.yaml                 |

   >! **说明：** 
   >
   >- _\{version\}_：表示版本号，请根据实际写入。
   >- 二进制启动时，添加-h参数可查看可用参数及说明。
   >- k8s启动时，注意yaml文件使用镜像名为自己构建的镜像。                                                                                                                                                                                                                                                                                                                                                                                                                 >

4.  执行以下命令，访问http接口查看服务。

    **http://**_\{ip\}_**:8082/metrics**

    >! **说明：** 
    >_\{ip\}_：表示物理机IP地址（二进制或者Docker启动时）或容器IP地址（K8s启动时，只显示端口到K8s集群内部。并配置了网络策略，默认只能Prometheus访问，网络策略详情参考[Kubernetes文档](https://kubernetes.io/zh/docs/concepts/services-networking/network-policies/)），请根据实际写入。

## 环境依赖

华为NPU驱动"Ascend HDK 22.0.RC2"及以后。

## 目录结构

```
npu-exporter                                                              
├── build                                      #编译和配置文件目录
│   ├── build.sh
│   ├── Dockerfile
│   ├── Dockerfile-310P-1usoc
│   ├── npu-exporter.yaml
│   ├── npu-exporter-310P-1usoc.yaml
│   ├── run-for-310P-1usoc.sh
│   └── test.sh
├── cmd                                      #主函数入口 
│   └──npu-exporter 
│       └──── main.go
├── collector                                  #指标收集                           
│   ├── container
│   │   ├── v1
│   │   │   ├── containerd.pb.go
│   │   │   ├── containerd.proto
│   │   │   └── spec.go
│   │   ├── utils.go
│   │   ├── parser.go
│   │   ├── runtime_ops.go
│   ├── testdata                              #测试数据
│   │   ├── prometheus_metrics
│   │   └── prometheus_metrics2
│   ├── npu_collector.go
│   ├── npu_collector_test.go
│   └── types.go
├── common-utils                                  #公共函数，其他组件会调用                           
│   ├── cache
│   │   ├── lrucache.go
│   │   └── lrucache_test.go
│   ├── hwlog                            
│   │   ├── api.go
│   │   ├── api_test.go
│   │   ├── hwlog_adaptor.go
│   │   ├── hwlog_adaptor_test.go
│   │   ├── log_limiter.go
│   │   ├── logger.go
│   │   ├── logger_test.go
│   │   ├── rolog.go
│   │   ├── rolog_test.go
│   │   ├── type.go
│   │   ├── utils.go
│   │   └── utils_test.go
│   ├── limiter                            
│   │   ├── limit_handler.go
│   │   ├── limit_handler_test.go
│   │   ├── limit_listener.go
│   │   └── limit_listener_test.go
│   ├── rand                            
│   │   ├── rand_linux.go
│   │   ├── rand_linux_test.go
│   │   ├── random.go
│   │   └── random_test.go
│   └── utils
│   │   ├── file.go
│   │   ├── file_check.go
│   │   ├── file_check_test.go
│   │   ├── file_test.go
│   │   ├── interface.go
│   │   ├── interface_test.go
│   │   ├── up_utils.go
│   │   ├── up_utils_test.go
│   │   ├── path.go
│   │   ├── path_test.go
│   │   ├── pwd_util.go
│   │   ├── pwd_util_test.go
│   │   ├── string.go
│   │   └── string_test.go
├── devmanager      
│   ├── common
│   │   ├── constants.go
│   │   ├── utils.go
│   │   └── type.go  
│   ├── dcmi
│   │   ├── constants.go
│   │   ├── dcmi.go                      #驱动相关接口封装
│   │   └── dcmi_interface_api.go
│   ├── a310mgr.go
│   ├── a310pmgr.go
│   ├── a910mgr.h
│   ├── devmanager.go
│   ├── devmanager_mock.go
│   └── devmanager_mock_err.go
├── vensions      
│   └── vension.go
├── go.mod                                     #go语言依赖文件
│   └── go.sum
├── LICENSE
└── README.md
```

## 版本更新信息

| 版本       | 发布日期   | 修改说明       |
| ---------- | ---------- | -------------- |
| v3.0.0 | 2022-1230 | 第一次发布 |

### metrics标签

| 标签名称                       | 标签说明                                              | 数值单位         |
| ------------------------------ | ----------------------------------------------------- | ---------------- |
| machine_npu_nums               | 昇腾系列AI处理器数目                                  | 个               |
| machine_npu_name               | 昇腾系列AI处理器名称                                  | N/A              |
| npu_chip_info_error_code       | 昇腾系列AI处理器错误码                                | N/A              |
| npu_chip_info_health_status    | 昇腾系列AI处理器健康状态                              | 1：健康0：不健康 |
| npu_chip_info_power            | 昇腾系列AI处理器功耗(310P板载功耗，910和310为芯片功耗)                                 | 瓦特（W）        |
| npu_chip_info_temperature      | 昇腾系列AI处理器温度                                  | 摄氏度（℃）      |
| npu_chip_info_used_memory      | 昇腾系列AI处理器已使用内存                            | MB               |
| npu_chip_info_total_memory     | 昇腾系列AI处理器总内存                                | MB               |
| npu_chip_info_hbm_used_memory  | 昇腾系列AI处理器HBM已使用内存（昇腾910 AI处理器专属） | MB               |
| npu_chip_info_hbm_total_memory | 昇腾系列AI处理器HBM总内存（昇腾910 AI处理器专属）     | MB               |
| npu_chip_info_utilization      | 昇腾系列AI处理器AI Core利用率                         | %                |
| npu_chip_info_voltage          | 昇腾系列AI处理器电压                                  | 伏特（V）        |
| npu_container_info             | 昇腾系列AI处理器在容器中的分配状态                     | N/A              |
