# NPU-Exporter.zh
-   [NPU-Exporter介绍](#NPU-Exporter介绍.md)
-   [快速上手](#快速上手.md)
-   [环境依赖](#环境依赖.md)
-   [目录结构](#目录结构.md)
-   [附录](#附录.md)
-   [版本更新信息](#版本更新信息.md)
<h2 id="NPU-Exporter介绍.md">NPU-Exporter介绍</h2>

Prometheus（普罗米修斯）是一个开源的系统监控和警报工具包，Exporter就是专门为Prometheus提供数据源的组件。由于Prometheus社区的活跃和大量的使用，已经有很多厂商或者服务提供了Exporter，如Prometheus官方的Node Exporter，MySQL官方出的MySQL Server Exporter和NVIDA的NVIDIA GPU Exporter。这些Exporter负责将特定监控对象的指标，转成Prometheus能够识别的数据格式，供Prometheus集成。NPU-Expoter是华为研发的专门收集华为NPU各种监控信息和指标，并封装成Prometheus专用数据格式的一个服务组件。

<h2 id="快速上手.md">快速上手</h2>

## 编译前准备<a name="section2078393613277"></a>

-   确保PC机与互联网网络互通，并已完成Git和Docker的安装。请参见：[Git安装](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)，[Docker-ce安装](https://docs.docker.com/engine/install/ubuntu/)进行安装。

-   已完成Go语言环境的安装（版本\>1.13，建议使用最新的bugfix版本）。请参见（[https://golang.org/](https://golang.org/)）。
-   根据所在网络环境配置Go代理地址，国内可使用**Goproxy China**，例如：

    ```
    go env -w GOPROXY=https://goproxy.cn,direct
    ```


## 编译NPU-Exporter<a name="section124015514383"></a>

1.  下载好源代码，执行以下命令，编译NPU-Exporter。

    **cd build**

    **bash build.sh**

    编译生成的文件在源码根目录下的output目录，文件如[表1](#table1860618363516)所示。

    **表 1**  编译生成的文件列表

    <a name="table1860618363516"></a>
    <table><thead align="left"><tr id="row1760620363510"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p860763675120"><a name="p860763675120"></a><a name="p860763675120"></a>文件名</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p1860718366515"><a name="p1860718366515"></a><a name="p1860718366515"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row14578104981510"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p33610228529"><a name="p33610228529"></a><a name="p33610228529"></a>npu-exporter</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p6361112211524"><a name="p6361112211524"></a><a name="p6361112211524"></a>NPU-Exporter二进制文件</p>
    </td>
    </tr>
    <tr id="row1860733675117"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p19634137205214"><a name="p19634137205214"></a><a name="p19634137205214"></a>Dockerfile</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p16348373528"><a name="p16348373528"></a><a name="p16348373528"></a>NPU-Exporter镜像构建文本文件</p>
    </td>
    </tr>
    <tr id="row11607103616516"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p12914103485214"><a name="p12914103485214"></a><a name="p12914103485214"></a>npu-exporter-<em id="i1306546185717"><a name="i1306546185717"></a><a name="i1306546185717"></a>{version}</em>.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p1591413485217"><a name="p1591413485217"></a><a name="p1591413485217"></a>NPU-Exporter的启动配置文件</p>
    </td>
    </tr>
    </tbody>
    </table>

    > **说明：** 
    >-   _\{__version__\}_：表示版本号，请根据实际写入。
    >-   arm和x86的二进制依赖不同，需要在对应架构上进行编译。


## 安装前准备<a name="section2739745153910"></a>

需要先完成《[MindX DL用户指南](https://www.hiascend.com/software/mindx-dl)》“安装前准备”章节中除“准备软件包”章节之外的其他章节内容。

请参考《[MindX DL用户指南](https://www.hiascend.com/software/mindx-dl)》中的“安装部署 \> 安装前准备”。

## 安装NPU-Exporter<a name="section3436132203218"></a>

请参考《[MindX DL用户指南](https://www.hiascend.com/software/mindx-dl)》中的“安装部署 \> 安装MindX DL \> 安装NPU-Exporter”。

**说明：** 社区版不支持https,如需要该功能请自行配置反向代理或使用npu-exporter商用版

## Prometheus集成方法<a name="section9854718262"></a>

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


<h2 id="环境依赖.md">环境依赖</h2>

华为NPU驱动20.1.0及以后。

<h2 id="目录结构.md">目录结构</h2>

```
npu-exporter                                                              
├── build                                      # 编译和配置文件目录
│   ├── build.sh
│   ├── Dockerfile
│   ├── npu-exporter.yaml
│   └── test.sh
├── collector                                  # 源码主要目录
│   ├── npu_collector.go
│   ├── npu_collector_test.go
│   ├── testdata                              # 测试数据
│   │   ├── prometheus_metrics
│   │   └── prometheus_metrics2
│   └── types.go
├── dsmi                                       # 驱动相关接口封装
│   ├── constants.go
│   ├── dcmi_interface_api.h
│   ├── devicetype.go
│   ├── dsmi_common_interface.h
│   ├── dsmi.go
│   ├── dsmi_mock_err.go
│   └── dsmi_mock.go
├── go.mod                                     # go语言依赖文件
├── go.sum
├── hwlog                                      # 日志记录，转储工具
│   ├── adaptor.go
│   └── logger.go
├── LICENSE
├── main.go                                    # 程序入口
├── output
├── README.EN.md
└── README.md
```

<h2 id="附录.md">附录</h2>

## Metrics标签<a name="section748613916322"></a>

**表 1**  Metrics标签

<a name="table29132853317"></a>
<table><thead align="left"><tr id="row17913285334"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p79131580337"><a name="p79131580337"></a><a name="p79131580337"></a>标签名称</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p591388183312"><a name="p591388183312"></a><a name="p591388183312"></a>标签说明</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p2913881334"><a name="p2913881334"></a><a name="p2913881334"></a>数值单位</p>
</th>
</tr>
</thead>
<tbody><tr id="row45172411220"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p95171941162219"><a name="p95171941162219"></a><a name="p95171941162219"></a>npu_exporter_version_info</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1451754162215"><a name="p1451754162215"></a><a name="p1451754162215"></a>NPU-Exporter版本</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p4517194114221"><a name="p4517194114221"></a><a name="p4517194114221"></a>N/A</p>
</td>
</tr>
<tr id="row17913686338"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p11913118173316"><a name="p11913118173316"></a><a name="p11913118173316"></a>machine_npu_nums</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1891310863312"><a name="p1891310863312"></a><a name="p1891310863312"></a><span id="ph18448782017"><a name="ph18448782017"></a><a name="ph18448782017"></a>昇腾系列AI处理器</span>数目</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1691318163310"><a name="p1691318163310"></a><a name="p1691318163310"></a>个</p>
</td>
</tr>
<tr id="row186411500243"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1564218019248"><a name="p1564218019248"></a><a name="p1564218019248"></a>npu_chip_info_name</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p186426062416"><a name="p186426062416"></a><a name="p186426062416"></a><span id="ph9831155162416"><a name="ph9831155162416"></a><a name="ph9831155162416"></a>昇腾系列AI处理器</span>名称和id</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p86422010245"><a name="p86422010245"></a><a name="p86422010245"></a>N/A</p>
</td>
</tr>
<tr id="row891428103320"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1391410833320"><a name="p1391410833320"></a><a name="p1391410833320"></a>npu_chip_info_error_code</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p5914188338"><a name="p5914188338"></a><a name="p5914188338"></a><span id="ph121151628810"><a name="ph121151628810"></a><a name="ph121151628810"></a>昇腾系列AI处理器</span>错误码</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p139141988330"><a name="p139141988330"></a><a name="p139141988330"></a>N/A</p>
</td>
</tr>
<tr id="row191420815333"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p991419853315"><a name="p991419853315"></a><a name="p991419853315"></a>npu_chip_info_health_status</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p149147813338"><a name="p149147813338"></a><a name="p149147813338"></a><span id="ph1117519411810"><a name="ph1117519411810"></a><a name="ph1117519411810"></a>昇腾系列AI处理器</span>健康状态</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><a name="ul123461183911"></a><a name="ul123461183911"></a><ul id="ul123461183911"><li>1：健康</li><li>0：不健康</li></ul>
</td>
</tr>
<tr id="row19914158163310"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1491410812333"><a name="p1491410812333"></a><a name="p1491410812333"></a>npu_chip_info_power</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1491419811332"><a name="p1491419811332"></a><a name="p1491419811332"></a><span id="ph1192711517812"><a name="ph1192711517812"></a><a name="ph1192711517812"></a>昇腾系列AI处理器</span>功耗</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p9914138103316"><a name="p9914138103316"></a><a name="p9914138103316"></a>瓦特（W）</p>
</td>
</tr>
<tr id="row19914138123313"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p189148818334"><a name="p189148818334"></a><a name="p189148818334"></a>npu_chip_info_temperature</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1391417843316"><a name="p1391417843316"></a><a name="p1391417843316"></a><span id="ph7171380817"><a name="ph7171380817"></a><a name="ph7171380817"></a>昇腾系列AI处理器</span>温度</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p891408153317"><a name="p891408153317"></a><a name="p891408153317"></a>摄氏度（℃）</p>
</td>
</tr>
<tr id="row18728152433415"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p11729524173417"><a name="p11729524173417"></a><a name="p11729524173417"></a>npu_chip_info_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p127291924153412"><a name="p127291924153412"></a><a name="p127291924153412"></a><span id="ph36036221588"><a name="ph36036221588"></a><a name="ph36036221588"></a>昇腾系列AI处理器</span>已使用内存</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p127293244346"><a name="p127293244346"></a><a name="p127293244346"></a>MB</p>
</td>
</tr>
<tr id="row05841735123419"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p11584193510347"><a name="p11584193510347"></a><a name="p11584193510347"></a>npu_chip_info_total_memory</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p135846358346"><a name="p135846358346"></a><a name="p135846358346"></a><span id="ph1667342513819"><a name="ph1667342513819"></a><a name="ph1667342513819"></a>昇腾系列AI处理器</span>总内存</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p18584143593420"><a name="p18584143593420"></a><a name="p18584143593420"></a>MB</p>
</td>
</tr>
<tr id="row73302032193413"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p15330432173420"><a name="p15330432173420"></a><a name="p15330432173420"></a>npu_chip_info_hbm_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7330193216343"><a name="p7330193216343"></a><a name="p7330193216343"></a><span id="ph11721102816810"><a name="ph11721102816810"></a><a name="ph11721102816810"></a>昇腾系列AI处理器</span>HBM已使用内存（<span id="ph10107162619264"><a name="ph10107162619264"></a><a name="ph10107162619264"></a>昇腾910 AI处理器</span>专属）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p63301132123411"><a name="p63301132123411"></a><a name="p63301132123411"></a>MB</p>
</td>
</tr>
<tr id="row97263274340"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p37267278344"><a name="p37267278344"></a><a name="p37267278344"></a>npu_chip_info_hbm_total_memory</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p6726202719349"><a name="p6726202719349"></a><a name="p6726202719349"></a><span id="ph412945811820"><a name="ph412945811820"></a><a name="ph412945811820"></a>昇腾系列AI处理器</span>HBM总内存（<span id="ph2035018501982"><a name="ph2035018501982"></a><a name="ph2035018501982"></a>昇腾910 AI处理器</span>专属）</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2726112763411"><a name="p2726112763411"></a><a name="p2726112763411"></a>MB</p>
</td>
</tr>
<tr id="row1047230183412"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2047630113416"><a name="p2047630113416"></a><a name="p2047630113416"></a>npu_chip_info_utilization</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p16471630163416"><a name="p16471630163416"></a><a name="p16471630163416"></a><span id="ph1803900912"><a name="ph1803900912"></a><a name="ph1803900912"></a>昇腾系列AI处理器</span>AI Core利用率</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p16474300347"><a name="p16474300347"></a><a name="p16474300347"></a>%</p>
</td>
</tr>
<tr id="row1078813555344"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1788195511341"><a name="p1788195511341"></a><a name="p1788195511341"></a>npu_chip_info_voltage</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1278895513413"><a name="p1278895513413"></a><a name="p1278895513413"></a><span id="ph825893794"><a name="ph825893794"></a><a name="ph825893794"></a>昇腾系列AI处理器</span>电压</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p978815513344"><a name="p978815513344"></a><a name="p978815513344"></a>伏特（V）</p>
</td>
</tr>
</tbody>
</table>

<h2 id="版本更新信息.md">版本更新信息</h2>

<a name="table7854542104414"></a>
<table><thead align="left"><tr id="zh-cn_topic_0280467800_row785512423445"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="zh-cn_topic_0280467800_p19856144274419"><a name="zh-cn_topic_0280467800_p19856144274419"></a><a name="zh-cn_topic_0280467800_p19856144274419"></a>版本</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="zh-cn_topic_0280467800_p3856134219446"><a name="zh-cn_topic_0280467800_p3856134219446"></a><a name="zh-cn_topic_0280467800_p3856134219446"></a>发布日期</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="zh-cn_topic_0280467800_p585634218445"><a name="zh-cn_topic_0280467800_p585634218445"></a><a name="zh-cn_topic_0280467800_p585634218445"></a>修改说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row539119585391"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p13391105873914"><a name="p13391105873914"></a><a name="p13391105873914"></a>v2.0.3</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p18391658133920"><a name="p18391658133920"></a><a name="p18391658133920"></a>2021-10-15</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p1839175810397"><a name="p1839175810397"></a><a name="p1839175810397"></a>修复了一些已知bug</p>
</td>
</tr>

<tr id="row539119585390"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p13391105873914"><a name="p13391105873914"></a><a name="p13391105873914"></a>v2.0.2</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p18391658133920"><a name="p18391658133920"></a><a name="p18391658133920"></a>2021-07-15</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p1839175810397"><a name="p1839175810397"></a><a name="p1839175810397"></a>支持默认以HTTPS启动。增加证书、私钥等文件导入功能。</p>
</td>
</tr>

<tr id="row4908113219334"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p10908832143316"><a name="p10908832143316"></a><a name="p10908832143316"></a>v2.0.1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p590810328337"><a name="p590810328337"></a><a name="p590810328337"></a>2021-03-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p1690843203317"><a name="p1690843203317"></a><a name="p1690843203317"></a>增加标签内容，提供Prometheus集成指导。</p>
</td>
</tr>
<tr id="zh-cn_topic_0280467800_row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0280467800_p08571442174415"><a name="zh-cn_topic_0280467800_p08571442174415"></a><a name="zh-cn_topic_0280467800_p08571442174415"></a>v20.2.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0280467800_p38571542154414"><a name="zh-cn_topic_0280467800_p38571542154414"></a><a name="zh-cn_topic_0280467800_p38571542154414"></a>2020-12-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0280467800_p5857142154415"><a name="zh-cn_topic_0280467800_p5857142154415"></a><a name="zh-cn_topic_0280467800_p5857142154415"></a>第一次发布。</p>
</td>
</tr>
</tbody>
</table>

