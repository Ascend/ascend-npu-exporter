# NPU-Exporter
-   [组件介绍](#组件介绍.md)
-   [编译NPU-Exporter](#编译NPU-Exporter.md)
-   [组件安装](#组件安装.md)
-   [更新日志](#更新日志.md)
-   [附录](#附录.md)

<h2 id="组件介绍.md">组件介绍</h2>


Prometheus（普罗米修斯）是一个开源的系统监控和警报工具包，Exporter就是专门为Prometheus提供数据源的组件。由于Prometheus社区的活跃和大量的使用，已经有很多厂商或者服务提供了Exporter，如Prometheus官方的Node Exporter，MySQL官方出的MySQL Server Exporter和NVIDA的NVIDIA GPU Exporter。这些Exporter负责将特定监控对象的指标，转成Prometheus能够识别的数据格式，供Prometheus集成。NPU-Expoter是华为自研的专门收集华为NPU各种监控信息和指标，并封装成Prometheus专用数据格式的一个服务组件。

<h2 id="编译NPU-Exporter.md">编译NPU-Exporter</h2>

1.  下载源码包，获得ascend-npu-exporter。

    示例：源码放在/home/test/ascend-npu-exporter目录下

2.  执行以下命令，进入构建目录，执行构建脚本，在“output“目录下生成二进制npu-exporter、yaml文件和Dockerfile等文件。

    **cd **_/home/test/_**ascend-npu-exporter/build/**

    **chmod +x build.sh**

    **./build.sh**

3.  执行以下命令，查看**output**生成的软件列表。

    **ll **_/home/test/_**ascend-npu-exporter/output**

    ```
    drwxr-xr-x  2 root root     4096 Jan 29 15:01 ./
    drwxr-xr-x 11 root root     4096 Jan 29 15:01 ../
    -r--------  1 root root      623 Jan 29 15:01 Dockerfile
    -r--------  1 root root      964 Jan 29 15:01 Dockerfile-310P-1usoc
    -r-x------  1 root root 15808008 Jan 29 15:01 npu-exporter
    -r--------  1 root root     4087 Jan 29 15:01 npu-exporter-310P-1usoc-v3.0.0.yaml
    -r--------  1 root root     3436 Jan 29 15:01 npu-exporter-v3.0.0.yaml
    -r-x------  1 root root     2554 Jan 29 15:01 run_for_310P_1usoc.sh
    ```

<h2 id="组件安装.md">组件安装</h2>

1.  请参考《MindX DL用户指南》(https://www.hiascend.com/software/mindx-dl)
    中的“集群调度用户指南 > 安装部署指导 \> 安装集群调度组件 \> 典型安装场景 \> 集群调度场景”进行。

<h2 id="更新日志.md">更新日志</h2>

| 版本       | 发布日期   | 修改说明       |
| ---------- | ---------- | -------------- |
| v3.0.0 | 2022-1230 | 首次发布 |

<h2 id="附录.md">附录</h2>
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
