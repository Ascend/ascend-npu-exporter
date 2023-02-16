#!/bin/bash
# Perform  build npu-exporter
# Copyright @ Huawei Technologies CO., Ltd. 2022-2022. All rights reserved
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

set -e

# create soft link for ubuntu image
os="$(cat /etc/*release* | grep -i "ubuntu")"
if [[ "$os" != "" ]]
then
    echo -e "[INFO]\t $(date +"%F %T:%N")\t use ubuntu image, so create soft link \"/lib64\" for \"/lib\""
    ln -s /lib /lib64 2>&1 >> /dev/null
fi

umask 027

echo -e "[INFO]\t $(date +"%F %T:%N")\t create driver's related directory"
mkdir -m 750 /var/driver -m 750 /var/dmp -m 750 /usr/slog -p -m 750 /home/drv/hdc_ppc

echo -e "[INFO]\t $(date +"%F %T:%N")\t modify owner and permission"
chown HwDmUser:HwDmUser /var/dmp
chown HwHiAiUser:HwHiAiUser /var/driver
chown HwHiAiUser:HwHiAiUser /home/drv/hdc_ppc
chown HwHiAiUser:HwHiAiUser /usr/slog

# log process run in background
echo -e "[INFO]\t $(date +"%F %T:%N")\t start slogd server in background"
su - HwHiAiUser -c "export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/ && /var/slogd &"
echo -e "[INFO]\t $(date +"%F %T:%N")\t start dmp_daemon server in background"
# dcmi interface process run in background
su - HwDmUser -c "export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/ && /var/dmp_daemon -I -M -U 8087 &"

export LD_LIBRARY_PATH=/usr/local/lib:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/add-ons:/usr/local/Ascend/driver/lib64:/usr/local/dcmi
# the host is openEuler, so the parameters "endpoint" and "containerd" are set to adapt to "-containerMode=docker" in default
# in openEuler os, the path of parameters "endpoint" and "containerd" are not in the default place
echo -e "[INFO]\t $(date +"%F %T:%N")\t start npu-exporter server"
/usr/local/bin/npu-exporter -port=8082 -ip=0.0.0.0 -updateTime=5 -logFile=/var/log/mindx-dl/npu-exporter/npu-exporter.log -logLevel=0 -containerMode=docker -endpoint=/run/dockershim.sock -containerd=/run/docker/containerd/containerd.sock

