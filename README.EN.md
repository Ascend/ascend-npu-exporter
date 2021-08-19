# NPU-Exporter.en
-   [Introduction to NPU-Exporter](#introduction-to-npu-exporter.md)
-   [Quick Start](#quick-start.md)
-   [Environment Dependencies](#environment-dependencies.md)
-   [Directory Structure](#directory-structure.md)
-   [Appendix](#appendix.md)
-   [Version Updates](#version-updates.md)
<h2 id="introduction-to-npu-exporter.md">Introduction to NPU-Exporter</h2>

Prometheus is an open-source system monitoring and alarm toolkit. Exporter is a component that provides data sources for Prometheus. Due to the active use of Prometheus, many service vendors have provided Exporters, such as the official Node Exporter from Prometheus, MySQL Server Exporter from MySQL, and NVIDIA GPU Exporter from NVIDIA. These Exporters convert the indicators of specific monitored objects into the data format that can be identified by Prometheus for integration. NPU-Expoter is a Huawei-developed service component that collects monitoring information and indicators of Huawei NPUs and encapsulates the information and indicators into Prometheus-specific data formats.

<h2 id="quick-start.md">Quick Start</h2>

## Preparations<a name="section2078393613277"></a>

-   Ensure that the PC can access the Internet and Git and Docker have been installed. For details, see  [Installing Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)  and  [Install Docker Engine on Ubuntu](https://docs.docker.com/engine/install/ubuntu/).

-   The Go language environment \(version 1.13 or later\) has been installed. You are advised to use the latest bugfix version. For details, visit  [https://golang.org/](https://golang.org/).
-   Configure the Go proxy address based on the network environment. In China, you can use  **Goproxy China**. For example:

    ```
    go env -w GOPROXY=https://goproxy.cn,direct
    ```


## Compiling and Starting the NPU-Exporter<a name="section124015514383"></a>

1.  Download the source code and build NPU-Exporter:

    **cd build**

    **bash build.sh**

    The files generated after compilation are stored in the  **output**  folder in the root directory of the source code, as shown in  [Table 1](#table1860618363516).

    **Table  1**  Files generated after compilation

    <a name="table1860618363516"></a>
    <table><thead align="left"><tr id="row1760620363510"><th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.1"><p id="p860763675120"><a name="p860763675120"></a><a name="p860763675120"></a>File Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="50%" id="mcps1.2.3.1.2"><p id="p1860718366515"><a name="p1860718366515"></a><a name="p1860718366515"></a>Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1860733675117"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p10251103995117"><a name="p10251103995117"></a><a name="p10251103995117"></a>npu-exporter</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p460753612511"><a name="p460753612511"></a><a name="p460753612511"></a>Binary file</p>
    </td>
    </tr>
    <tr id="row10607103612512"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p15607183645112"><a name="p15607183645112"></a><a name="p15607183645112"></a>npu-exporter-<em id="i102891035111512"><a name="i102891035111512"></a><a name="i102891035111512"></a>{verson}</em>-<em id="i88561384151"><a name="i88561384151"></a><a name="i88561384151"></a>{arch}</em>.tar.gz</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p860763613515"><a name="p860763613515"></a><a name="p860763613515"></a>Docker image package (exported by running the <strong id="b14873165812123"><a name="b14873165812123"></a><a name="b14873165812123"></a>docker save</strong> command)</p>
    </td>
    </tr>
    <tr id="row11607103616516"><td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.1 "><p id="p1760783615519"><a name="p1760783615519"></a><a name="p1760783615519"></a>npu-exporter-<em id="i11917433158"><a name="i11917433158"></a><a name="i11917433158"></a>{version}</em>.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="50%" headers="mcps1.2.3.1.2 "><p id="p2691654185113"><a name="p2691654185113"></a><a name="p2691654185113"></a>YAML file for startup by K8s</p>
    </td>
    </tr>
    </tbody>
    </table>

    >![](public_sys-resources/icon-note.gif) **NOTE:** 
    >-   _\{version\}_: indicates the version number. Set it to the actual version number.
    >-   _\{arch\}_: indicates the system architecture. Set it to the actual architecture.

2.  Start NPU-Exporter.

    **Table  2**  Commands

    <a name="table753174755410"></a>
    <table><thead align="left"><tr id="row453847165419"><th class="cellrowborder" valign="top" width="19.03%" id="mcps1.2.3.1.1"><p id="p853184719541"><a name="p853184719541"></a><a name="p853184719541"></a>Startup Mode</p>
    </th>
    <th class="cellrowborder" valign="top" width="80.97%" id="mcps1.2.3.1.2"><p id="p45374715541"><a name="p45374715541"></a><a name="p45374715541"></a>Command</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row135314477545"><td class="cellrowborder" valign="top" width="19.03%" headers="mcps1.2.3.1.1 "><p id="p1153184719542"><a name="p1153184719542"></a><a name="p1153184719542"></a>Binary startup</p>
    </td>
    <td class="cellrowborder" valign="top" width="80.97%" headers="mcps1.2.3.1.2 "><p id="p35384765410"><a name="p35384765410"></a><a name="p35384765410"></a><strong id="b1199652954414"><a name="b1199652954414"></a><a name="b1199652954414"></a>./npu-exporter</strong></p>
    </td>
    </tr>
    <tr id="row85374719546"><td class="cellrowborder" valign="top" width="19.03%" headers="mcps1.2.3.1.1 "><p id="p1612783925510"><a name="p1612783925510"></a><a name="p1612783925510"></a>Startup by Docker</p>
    </td>
    <td class="cellrowborder" valign="top" width="80.97%" headers="mcps1.2.3.1.2 "><p id="p153204718545"><a name="p153204718545"></a><a name="p153204718545"></a><strong id="b14812142011815"><a name="b14812142011815"></a><a name="b14812142011815"></a>docker run -it --privileged --rm</strong> <strong id="b198124207189"><a name="b198124207189"></a><a name="b198124207189"></a>--volume=/var/log/npu-exporter:/var/log/npu-exporter --volume=/etc/localtime:/etc/localtime:ro --volume=/usr/local/Ascend/driver:/usr/local/Ascend/driver:ro --publish=8082:8082 --name=npu-exporter npu-exporter:</strong><em id="i1623015711169"><a name="i1623015711169"></a><a name="i1623015711169"></a>{version}</em></p>
    </td>
    </tr>
    <tr id="row1753104717546"><td class="cellrowborder" valign="top" width="19.03%" headers="mcps1.2.3.1.1 "><p id="p95311478543"><a name="p95311478543"></a><a name="p95311478543"></a>(Recommended) Startup by K8s</p>
    </td>
    <td class="cellrowborder" valign="top" width="80.97%" headers="mcps1.2.3.1.2 "><p id="p35344718548"><a name="p35344718548"></a><a name="p35344718548"></a><strong id="b16387131111610"><a name="b16387131111610"></a><a name="b16387131111610"></a>kubectl apply -f npu-exporter-</strong><em id="i12483312111610"><a name="i12483312111610"></a><a name="i12483312111610"></a>{version}</em><strong id="b1438741115162"><a name="b1438741115162"></a><a name="b1438741115162"></a>.yaml</strong></p>
    </td>
    </tr>
    </tbody>
    </table>

    >![](public_sys-resources/icon-note.gif) **NOTE:** 
    >-   _\{version\}_: indicates the version number. Set it to the actual version number.
    >-   During binary startup, the  **-h**  parameter can be added to view available parameters and description.

3.  Access the HTTP interface to view services.

    **http://**_\{ip\}_**:8082/metrics**

    >![](public_sys-resources/icon-note.gif) **NOTE:** 
    >_\{ip\}_: indicates the IP address of the physical machine for the binary startup or startup by Docker, or the IP address of the container for startup by K8s, only the port to the K8s cluster is displayed. Network policies are configured. By default, only Prometheus can access the IP address. For details about the network policies, see the  [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/network-policies/). Set the IP address based on the site requirements.


## Integrating Prometheus<a name="section9854718262"></a>

-   If Prometheus and NPU-Exporter are deployed in K8s, the label  **app: prometheus**  must be configured in the YAML file for starting Prometheus because network policies are configured for NPU-Exporter.

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
│   ├── cov.out
│   ├── Dockerfile
│   ├── npu-exporter.yaml
│   └── test.sh
├── collector                                  # Main directory of the source code
│   ├── cov.out
│   ├── npu_collector.go
│   ├── npu_collector_test.go
│   ├── container
│   │   ├── cri_name.go
│   │   ├── docker_name.go
│   │   ├── name_fetcher.go
│   │   ├── parser.go
│   │   ├── runtime_ops.go
│   ├── testdata                              # Test data
│   │   ├── prometheus_metrics
│   │   └── prometheus_metrics2
│   └── types.go
├── dsmi                                       # Driver-related API encapsulation
│   ├── constants.go
│   ├── devicetype.go
│   ├── dsmi_common_interface.h
│   ├── dcmi_interface_api.h
│   ├── dsmi.go
│   ├── dsmi_mock_err.go
│   └── dsmi_mock.go
├── go.mod                                     # Go language dependencies
├── go.sum
├── LICENSE
├── main.go                                    # Program entry
└── README.md
```

<h2 id="appendix.md">Appendix</h2>

## Log Splitting Configuration<a name="section11964111913012"></a>

1.  Set the permission on the log directory.

    ```
    chmod -R 750 /var/log/npu-exporter/
    ```

2.  Configure log dumping.

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
<tbody><tr id="row17913686338"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p11913118173316"><a name="p11913118173316"></a><a name="p11913118173316"></a>machine_npu_nums</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1891310863312"><a name="p1891310863312"></a><a name="p1891310863312"></a>Number of <span id="ph6510175913514"><a name="ph6510175913514"></a><a name="ph6510175913514"></a>Ascend AI Processor</span>s</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1691318163310"><a name="p1691318163310"></a><a name="p1691318163310"></a>Number</p>
</td>
</tr>
<tr id="row189132863317"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p891368173319"><a name="p891368173319"></a><a name="p891368173319"></a>machine_npu_name</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p391398103310"><a name="p391398103310"></a><a name="p391398103310"></a>Name of the <span id="ph4454353173613"><a name="ph4454353173613"></a><a name="ph4454353173613"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p891418812333"><a name="p891418812333"></a><a name="p891418812333"></a>N/A</p>
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
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p149147813338"><a name="p149147813338"></a><a name="p149147813338"></a>Health status of an <span id="ph14295161123718"><a name="ph14295161123718"></a><a name="ph14295161123718"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><a name="ul123461183911"></a><a name="ul123461183911"></a><ul id="ul123461183911"><li>1: healthy</li><li>0: unhealthy</li></ul>
</td>
</tr>
<tr id="row19914158163310"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1491410812333"><a name="p1491410812333"></a><a name="p1491410812333"></a>npu_chip_info_power</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1491419811332"><a name="p1491419811332"></a><a name="p1491419811332"></a>Power consumption of an <span id="ph16273634103712"><a name="ph16273634103712"></a><a name="ph16273634103712"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p9914138103316"><a name="p9914138103316"></a><a name="p9914138103316"></a>W</p>
</td>
</tr>
<tr id="row19914138123313"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p189148818334"><a name="p189148818334"></a><a name="p189148818334"></a>npu_chip_info_temperature</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1391417843316"><a name="p1391417843316"></a><a name="p1391417843316"></a>Temperature of an <span id="ph1733523918371"><a name="ph1733523918371"></a><a name="ph1733523918371"></a>Ascend AI Processor</span></p>
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
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p16471630163416"><a name="p16471630163416"></a><a name="p16471630163416"></a>AI Core usage of an <span id="ph8733175811376"><a name="ph8733175811376"></a><a name="ph8733175811376"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p16474300347"><a name="p16474300347"></a><a name="p16474300347"></a>%</p>
</td>
</tr>
<tr id="row1078813555344"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1788195511341"><a name="p1788195511341"></a><a name="p1788195511341"></a>npu_chip_info_voltage</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1278895513413"><a name="p1278895513413"></a><a name="p1278895513413"></a>Voltage of an <span id="ph422882103813"><a name="ph422882103813"></a><a name="ph422882103813"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p978815513344"><a name="p978815513344"></a><a name="p978815513344"></a>V</p>
</td>
</tr>
<tr id="row1078813555344"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p1788195511341"><a name="p1788195511341"></a><a name="p1788195511341"></a>npu_container_info</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p1278895513413"><a name="p1278895513413"></a><a name="p1278895513413"></a>Allocation status among containers of <span id="ph422882103813"><a name="ph422882103813"></a><a name="ph422882103813"></a>Ascend AI Processor</span></p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p978815513344"><a name="p978815513344"></a><a name="p978815513344"></a>N/A</p>
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
<tbody><tr id="en-us_topic_0280467800_row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="en-us_topic_0280467800_p08571442174415"><a name="en-us_topic_0280467800_p08571442174415"></a><a name="en-us_topic_0280467800_p08571442174415"></a>v2.0.1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="en-us_topic_0280467800_p38571542154414"><a name="en-us_topic_0280467800_p38571542154414"></a><a name="en-us_topic_0280467800_p38571542154414"></a>2021-03-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="en-us_topic_0280467800_p5857142154415"><a name="en-us_topic_0280467800_p5857142154415"></a><a name="en-us_topic_0280467800_p5857142154415"></a>Add support to 710 chip</p>
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

