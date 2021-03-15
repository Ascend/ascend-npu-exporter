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
    |npu-exporter-{verson}-{arch}.tar.gz| Docker镜像包（需使用docker load导入） |
    |npu-exporter-{version}.yaml |K8s启动yaml|

    >! **说明：** 
    >
    >- _\{version\}_：表示版本号，请根据实际写入。
    >- _\{arch\}_：表示系统架构，请根据实际写入。

2. 执行以下命令，启动NPU-Exporter。

    **表 2**  操作命令

| 启动类型            | 启动命令                                                     |
| :-------------------- | -------------------------------------------------------- |
| 二进制启动          | ./npu-exporter                                             |
| Docker启动          | docker run -it --privileged   （） -- rm --volume=/var/log/npu-exporter:/var/log/npu-exporter --volume=/etc/localtime:/etc/localtime:ro --volume=/usr/local/Ascend/driver:/usr/local/Ascend/driver:ro --publish=8082:8082 --name=npu-exporter npu-exporter:{version} |
| K8s启动（推荐方式） | kubectl apply -f npu-exporter-{version}.yaml                 |

   >! **说明：** 
   >
   >- _\{version\}_：表示版本号，请根据实际写入。
   >- 二进制启动时，添加-h参数可查看可用参数及说明。

3.  执行以下命令，访问http接口查看服务。

    **http://**_\{ip\}_**:8082/metrics**

    >! **说明：** 
    >_\{ip\}_：表示物理机IP地址（二进制或者Docker启动时）或容器IP地址（K8s启动时，只显示端口到K8s集群内部。并配置了网络策略，默认只能Prometheus访问，网络策略详情参考[Kubernetes文档](https://kubernetes.io/zh/docs/concepts/services-networking/network-policies/)），请根据实际写入。


#### Prometheus集成方法

-   如果Prometheus和NPU-Exporter部署在K8s中，考虑到NPU-Exporter配置了网络策略，Prometheus的启动yaml中需要配置app: prometheus的标签（labels）。

    ```
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      labels:
        name: prometheus-deployment
      name: prometheus
     spec:
      replicas: 1
      selector:
        matchLabels:
          app: prometheus
      template:
        metadata:
          labels:
            app: prometheus
        spec:
          nodeSelector:
            masterselector: dls-master-node
          containers:
          - image: prom/prometheus:v2.10.0
            name: prometheus
            command:
            - "/bin/prometheus"
           ....
    ```

-   Prometheus的config.yaml中增加如下scrape\_configs配置抓取NPU-Exporter。

    ```
      - job_name: 'kubernetes-npu-exporter'
          kubernetes_sd_configs:
          - role: pod
          scheme: http
          relabel_configs:
          - action: keep
            source_labels: [__meta_kubernetes_namespace]
            regex: npu-exporter
          - source_labels: [__meta_kubernetes_pod_node_name]
            target_label: job
            replacement: ${1}
    ```

## 环境依赖

华为NPU驱动20.1.0及以后。

## 目录结构

```
npu-exporter                                                              
├── build                                      #编译和配置文件目录
│   ├── build.sh
│   ├── cov.out
│   ├── Dockerfile
│   ├── npu-exporter.yaml
│   └── test.sh
├── collector                                  #源码主要目录                           
│   ├── cov.out
│   ├── npu_collector.go
│   ├── npu_collector_test.go
│   ├── testdata                              #测试数据
│   │   ├── prometheus_metrics
│   │   └── prometheus_metrics2
│   └── types.go
├── dsmi                                       #驱动相关接口封装
│   ├── constants.go
│   ├── devicetype.go
│   ├── dsmi_common_interface.h
│   ├── dcmi_interface_api.h
│   ├── dsmi.go
│   ├── dsmi_mock_err.go
│   └── dsmi_mock.go
├── go.mod                                     #go语言依赖文件
├── go.sum
├── LICENSE
├── main.go                                    #程序入口
└── README.md
```

## 版本更新信息

| 版本       | 发布日期   | 修改说明       |
| ---------- | ---------- | -------------- |
| v2.0.1 | 2020-3-30 | 适配710 |
| v20.2.0 | 2020-12-30 | 第一次正式发布 |

## 附录

### 日志切割配置

1. 设置日志目录权限

   ```
   chmod -R 750 /var/log/npu-exporter/
   ```

2. 配置日志转储

   ```
   cat <<EOF >/etc/logrotate.d/npu-exporter
   /var/log/npu-exporter/*.log{
   daily
   rotate 8
   size 10M
   compress
   dateext
   missingok
   notifempty
   copytruncate
   create 0640 root root
   sharedscripts
   postrotate
   chmod 640 /var/log/npu-exporter/*.log
   chmod 440 /var/log/npu-exporter/*.log-*
   endscript
   }
   EOF
   
   chmod 640 /etc/logrotate.d/npu-exporter
   ```

### metrics标签

| 标签名称                       | 标签说明                                              | 数值单位         |
| ------------------------------ | ----------------------------------------------------- | ---------------- |
| machine_npu_nums               | 昇腾系列AI处理器数目                                  | 个               |
| machine_npu_name               | 昇腾系列AI处理器名称                                  | N/A              |
| npu_chip_info_error_code       | 昇腾系列AI处理器错误码                                | N/A              |
| npu_chip_info_health_status    | 昇腾系列AI处理器健康状态                              | 1：健康0：不健康 |
| npu_chip_info_power            | 昇腾系列AI处理器功耗(710板载功耗，910和310为芯片功耗)                                 | 瓦特（W）        |
| npu_chip_info_temperature      | 昇腾系列AI处理器温度                                  | 摄氏度（℃）      |
| npu_chip_info_used_memory      | 昇腾系列AI处理器已使用内存                            | MB               |
| npu_chip_info_total_memory     | 昇腾系列AI处理器总内存                                | MB               |
| npu_chip_info_hbm_used_memory  | 昇腾系列AI处理器HBM已使用内存（昇腾910 AI处理器专属） | MB               |
| npu_chip_info_hbm_total_memory | 昇腾系列AI处理器HBM总内存（昇腾910 AI处理器专属）     | MB               |
| npu_chip_info_utilization      | 昇腾系列AI处理器AI Core利用率                         | %                |
| npu_chip_info_voltage          | 昇腾系列AI处理器电压                                  | 伏特（V）        |

