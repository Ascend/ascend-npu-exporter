#!/bin/bash
# Perform  build npu-exporter
# Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.

set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v2.0.3"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '5p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*:}
fi

arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

OUTPUT_NAME="cert-importer"
function clear_env() {
  if [ -d "${TOP_DIR}/output" ]; then
    rm -rf "${TOP_DIR}"/output/${OUTPUT_NAME}
  else
    mkdir -p "${TOP_DIR}"/output
  fi

}

function build() {
  cd "${TOP_DIR}/cmd/importer"
  CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie -ldflags "-s -extldflags=-Wl,-z,now  -X huawei.com/npu-exporter/hwlog.BuildName=${OUTPUT_NAME} \
            -X huawei.com/npu-exporter/hwlog.BuildVersion=${build_version}_linux-${arch}" \
    -o ${OUTPUT_NAME}
  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to build ${OUTPUT_NAME}"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}"/cmd/importer/${OUTPUT_NAME} "${TOP_DIR}"/output
  chmod 500 "${TOP_DIR}"/output/${OUTPUT_NAME}
}

function main() {
  clear_env
  build
  mv_file
}
if [ "$1" = clean ]; then
  clear_env
  exit 0
fi
main
