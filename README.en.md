# NPU-Exporter.en
-   [Introduction to NPU-Exporter](#introduction-to-npu-exporter.md)
-   [Quick Start](#quick-start.md)
-   [Environment Dependencies](#environment-dependencies.md)
-   [Directory Structure](#directory-structure.md)
-   [Appendix](#appendix.md)
-   [Version Updates](#version-updates.md)
<h2 id="introduction-to-npu-exporter.md">Introduction to NPU-Exporter</h2>

Prometheus is an open source system monitoring and alarm toolkit. Exporter is a component that provides data sources for Prometheus. Due to frequent and wide use of Prometheus, many service vendors have provided Exporters, such as Prometheus Node Exporter, MySQL Server Exporter, and NVIDIA GPU Exporter. These Exporters convert the indicators of specific monitored objects into the data format that can be identified by Prometheus for integration. NPU-Exporter is a Huawei-developed service component that collects monitoring information and indicators of Huawei NPUs and encapsulates the information and indicators into Prometheus-specific data formats.

<h2 id="quick-start.md">Quick Start</h2>

## Preparations<a name="section2078393613277"></a>

-   Ensure that the PC can access the Internet and Git and Docker have been installed. For details, see the  [_Installing Git_](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)  and  [_Install Docker Engine on Ubuntu_](https://docs.docker.com/engine/install/ubuntu/).

-   The Go language environment \(version 1.13 or later\) has been installed. You are advised to use the latest bugfix version. For details, visit  [https://golang.org/](https://golang.org/).
-   Configure the Go proxy address based on the network environment. In China, you can use  **Goproxy China**. For example:

    ```
    go env -w GOPROXY=https://goproxy.cn,direct
    ```


## Building NPU-Exporter<a name="section124015514383"></a>

1. Download the source code and run the following command to build NPU-Exporter:

    **cd build**

    **bash build.sh**

    The files generated after build are stored in the  **output**  folder in the root directory of the source code, as shown in  [Table 1](#table1860618363516).

    **Table  1**  Files generated after build

    <a name="table1860618363516"></a>
    <table><thead align="left"><tr id="row1760620363510"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p860763675120"><a name="p860763675120"></a><a name="p860763675120"></a>File Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p1860718366515"><a name="p1860718366515"></a><a name="p1860718366515"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row14578104981510"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p33610228529"><a name="p33610228529"></a><a name="p33610228529"></a>npu-exporter</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p6361112211524"><a name="p6361112211524"></a><a name="p6361112211524"></a>NPU-Exporter binary file</p>
    </td>
    </tr>
    <tr id="row1860733675117"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p19634137205214"><a name="p19634137205214"></a><a name="p19634137205214"></a>Dockerfile</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p16348373528"><a name="p16348373528"></a><a name="p16348373528"></a>NPU-Exporter image building text file</p>
    </td>
    </tr>
    <tr id="row11607103616516"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p12914103485214"><a name="p12914103485214"></a><a name="p12914103485214"></a>npu-exporter-<em id="i1306546185717"><a name="i1306546185717"></a><a name="i1306546185717"></a>{version}</em>.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p1591413485217"><a name="p1591413485217"></a><a name="p1591413485217"></a>NPU-Exporter startup configuration file</p>
    </td>
    </tr>
    </tbody>
    </table>

    > **NOTE:** 
    >-   _\{version\}_: indicates the version number. Set it based on the actual situation.
    >-   The binary dependency of ARM is different from that of x86. Therefore, build needs to be performed on the corresponding architecture.


## Prerequisites<a name="section2739745153910"></a>

Perform operations described in all sections except "Preparing Software Packages" in section "Preparing for Installation" in the  [_MindX DL User Guide_](https://www.hiascend.com/software/mindx-dl).

For details, see "Installation and Deployment \> Preparations Before Installation" in the [_MindX DL User Guide_](https://www.hiascend.com/software/mindx-dl).

## Installing NPU-Exporter<a name="section3436132203218"></a>

For details, see "Installation and Deployment \> Installing MindX DL \> Installing NPU-Exporter" in the  [_MindX DL User Guide_](https://www.hiascend.com/software/mindx-dl).

**NOTE:** The community edition does not support HTTPS. If this function is required, please configure reverse proxy by yourself or use the npu-exporter commercial edition. 

## Integrating Prometheus<a name="section9854718262"></a>

-   If Prometheus and NPU-Exporter are deployed in Kubernetes, the label  **app: prometheus**  must be configured in the YAML file for starting Prometheus because network policies are configured for NPU-Exporter.

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

-   Add the following  **scrape\_configs**  configuration in the  **config.yaml**  file of Prometheus to capture NPU-Exporter.

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


<h2 id="environment-dependencies.md">Environment Dependencies</h2>

Huawei NPU driver 20.1.0 or later.

<h2 id="directory-structure.md">Directory Structure</h2>

```
npu-exporter                                                              
├── build                                      # Directory for storing build and configuration files
│   ├── build.sh
│   ├── Dockerfile
│   ├── npu-exporter.yaml
│   └── test.sh
├── collector                                  # Main directory of the source code
│   ├── npu_collector.go
│   ├── npu_collector_test.go
│   ├── testdata                              # Test data
│   │   ├── prometheus_metrics
│   │   └── prometheus_metrics2
│   └── types.go
├── dsmi                                       # Driver-related API encapsulation
│   ├── constants.go
│   ├── dcmi_interface_api.h
│   ├── devicetype.go
│   ├── dsmi_common_interface.h
│   ├── dsmi.go
│   ├── dsmi_mock_err.go
│   └── dsmi_mock.go
├── go.mod                                     # Go language dependencies
├── go.sum
├── hwlog                                      # Log recording and dumping tool
│   ├── adaptor.go
│   └── logger.go
├── LICENSE
├── main.go                                    # Program entry
├── output
├── README.EN.md
└──  README.md   
```

<h2 id="appendix.md">Appendix</h2>

## Metrics Labels<a name="section748613916322"></a>

**Table  1**  Metrics labels

<a name="table29132853317"></a>
<table><thead align="left"><tr id="row17913285334"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p79131580337"><a name="p79131580337"></a><a name="p79131580337"></a>Label</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p591388183312"><a name="p591388183312"></a><a name="p591388183312"></a>Description</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p2913881334"><a name="p2913881334"></a><a name="p2913881334"></a>Unit</p>
</th>
</tr>
</thead>
<tbody><tr id="row45172411220"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p95171941162219"><a name="p95171941162219"></a><a name="p95171941162219"></a>npu_exporter_version_info</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1451754162215"><a name="p1451754162215"></a><a name="p1451754162215"></a>NPU-Exporter version</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p4517194114221"><a name="p4517194114221"></a><a name="p4517194114221"></a>N/A</p>
</td>
</tr>
<tr id="row17913686338"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p11913118173316"><a name="p11913118173316"></a><a name="p11913118173316"></a>machine_npu_nums</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1891310863312"><a name="p1891310863312"></a><a name="p1891310863312"></a>Number of <span id="ph6510175913514"><a name="ph6510175913514"></a><a name="ph6510175913514"></a>Ascend AI Processor</span>s</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1691318163310"><a name="p1691318163310"></a><a name="p1691318163310"></a>Number</p>
</td>
</tr>
<tr id="row186411500243"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1564218019248"><a name="p1564218019248"></a><a name="p1564218019248"></a>npu_chip_info_name</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p186426062416"><a name="p186426062416"></a><a name="p186426062416"></a>Name and ID of the <span id="ph9831155162416"><a name="ph9831155162416"></a><a name="ph9831155162416"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p86422010245"><a name="p86422010245"></a><a name="p86422010245"></a>N/A</p>
</td>
</tr>
<tr id="row891428103320"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1391410833320"><a name="p1391410833320"></a><a name="p1391410833320"></a>npu_chip_info_error_code</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p5914188338"><a name="p5914188338"></a><a name="p5914188338"></a>Error code of the <span id="ph0988751378"><a name="ph0988751378"></a><a name="ph0988751378"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p139141988330"><a name="p139141988330"></a><a name="p139141988330"></a>N/A</p>
</td>
</tr>
<tr id="row191420815333"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p991419853315"><a name="p991419853315"></a><a name="p991419853315"></a>npu_chip_info_health_status</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p149147813338"><a name="p149147813338"></a><a name="p149147813338"></a>Health status of the <span id="ph14295161123718"><a name="ph14295161123718"></a><a name="ph14295161123718"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><a name="ul123461183911"></a><a name="ul123461183911"></a><ul id="ul123461183911"><li>1: healthy</li><li>0: unhealthy</li></ul>
</td>
</tr>
<tr id="row19914158163310"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1491410812333"><a name="p1491410812333"></a><a name="p1491410812333"></a>npu_chip_info_power</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1491419811332"><a name="p1491419811332"></a><a name="p1491419811332"></a>Power consumption of the <span id="ph16273634103712"><a name="ph16273634103712"></a><a name="ph16273634103712"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p9914138103316"><a name="p9914138103316"></a><a name="p9914138103316"></a>W</p>
</td>
</tr>
<tr id="row19914138123313"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p189148818334"><a name="p189148818334"></a><a name="p189148818334"></a>npu_chip_info_temperature</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1391417843316"><a name="p1391417843316"></a><a name="p1391417843316"></a>Temperature of the <span id="ph1733523918371"><a name="ph1733523918371"></a><a name="ph1733523918371"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p891408153317"><a name="p891408153317"></a><a name="p891408153317"></a>&deg;C</p>
</td>
</tr>
<tr id="row18728152433415"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p11729524173417"><a name="p11729524173417"></a><a name="p11729524173417"></a>npu_chip_info_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p127291924153412"><a name="p127291924153412"></a><a name="p127291924153412"></a>Used memory of the <span id="ph79544414376"><a name="ph79544414376"></a><a name="ph79544414376"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p127293244346"><a name="p127293244346"></a><a name="p127293244346"></a>MB</p>
</td>
</tr>
<tr id="row05841735123419"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p11584193510347"><a name="p11584193510347"></a><a name="p11584193510347"></a>npu_chip_info_total_memory</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p135846358346"><a name="p135846358346"></a><a name="p135846358346"></a>Total memory of the <span id="ph4644114817375"><a name="ph4644114817375"></a><a name="ph4644114817375"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p18584143593420"><a name="p18584143593420"></a><a name="p18584143593420"></a>MB</p>
</td>
</tr>
<tr id="row73302032193413"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p15330432173420"><a name="p15330432173420"></a><a name="p15330432173420"></a>npu_chip_info_hbm_used_memory</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p7330193216343"><a name="p7330193216343"></a><a name="p7330193216343"></a>Used HBM memory dedicated for the <span id="ph116117517379"><a name="ph116117517379"></a><a name="ph116117517379"></a>Ascend 910 AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p63301132123411"><a name="p63301132123411"></a><a name="p63301132123411"></a>MB</p>
</td>
</tr>
<tr id="row97263274340"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p37267278344"><a name="p37267278344"></a><a name="p37267278344"></a>npu_chip_info_hbm_total_memory</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p6726202719349"><a name="p6726202719349"></a><a name="p6726202719349"></a>Total HBM memory dedicated for the <span id="ph397614552374"><a name="ph397614552374"></a><a name="ph397614552374"></a>Ascend 910 AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p2726112763411"><a name="p2726112763411"></a><a name="p2726112763411"></a>MB</p>
</td>
</tr>
<tr id="row1047230183412"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2047630113416"><a name="p2047630113416"></a><a name="p2047630113416"></a>npu_chip_info_utilization</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p16471630163416"><a name="p16471630163416"></a><a name="p16471630163416"></a>AI Core usage of the <span id="ph8733175811376"><a name="ph8733175811376"></a><a name="ph8733175811376"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p16474300347"><a name="p16474300347"></a><a name="p16474300347"></a>%</p>
</td>
</tr>
<tr id="row1078813555344"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1788195511341"><a name="p1788195511341"></a><a name="p1788195511341"></a>npu_chip_info_voltage</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1278895513413"><a name="p1278895513413"></a><a name="p1278895513413"></a>Voltage of the <span id="ph422882103813"><a name="ph422882103813"></a><a name="ph422882103813"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p978815513344"><a name="p978815513344"></a><a name="p978815513344"></a>V</p>
</td>
</tr>
</tbody>
</table>

<h2 id="version-updates.md">Version Updates</h2>

<a name="table7854542104414"></a>
<table><thead align="left"><tr id="en-us_topic_0280467800_row785512423445"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="en-us_topic_0280467800_p19856144274419"><a name="en-us_topic_0280467800_p19856144274419"></a><a name="en-us_topic_0280467800_p19856144274419"></a>Version</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="en-us_topic_0280467800_p3856134219446"><a name="en-us_topic_0280467800_p3856134219446"></a><a name="en-us_topic_0280467800_p3856134219446"></a>Date</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="en-us_topic_0280467800_p585634218445"><a name="en-us_topic_0280467800_p585634218445"></a><a name="en-us_topic_0280467800_p585634218445"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row539119585390"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p13391105873914"><a name="p13391105873914"></a><a name="p13391105873914"></a>v2.0.3</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p18391658133920"><a name="p18391658133920"></a><a name="p18391658133920"></a>2021-10-15</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p1839175810397"><a name="p1839175810397"></a><a name="p1839175810397"></a>Fixed some known bugs</p>
</td>
</tr>

<tr id="row539119585390"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p13391105873914"><a name="p13391105873914"></a><a name="p13391105873914"></a>v2.0.2</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p18391658133920"><a name="p18391658133920"></a><a name="p18391658133920"></a>2021-07-15</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p1839175810397"><a name="p1839175810397"></a><a name="p1839175810397"></a>Supported the startup in HTTPS mode by default. Added the function of importing files, such as certificate files and private key files.</p>
</td>
</tr><tr id="row4908113219334"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p10908832143316"><a name="p10908832143316"></a><a name="p10908832143316"></a>v2.0.1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p590810328337"><a name="p590810328337"></a><a name="p590810328337"></a>2021-03-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p1690843203317"><a name="p1690843203317"></a><a name="p1690843203317"></a>Added the label content, and provided the Prometheus integration guide.</p>
</td>
</tr>
<tr id="en-us_topic_0280467800_row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="en-us_topic_0280467800_p08571442174415"><a name="en-us_topic_0280467800_p08571442174415"></a><a name="en-us_topic_0280467800_p08571442174415"></a>v20.2.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="en-us_topic_0280467800_p38571542154414"><a name="en-us_topic_0280467800_p38571542154414"></a><a name="en-us_topic_0280467800_p38571542154414"></a>2020-12-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="en-us_topic_0280467800_p5857142154415"><a name="en-us_topic_0280467800_p5857142154415"></a><a name="en-us_topic_0280467800_p5857142154415"></a>This is the first official release.</p>
</td>
</tr>
</tbody>
</table>

