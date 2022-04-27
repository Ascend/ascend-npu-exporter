#!/bin/bash
# Perform  build npu-exporter
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved

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

OUTPUT_NAME="npu-exporter"
DOCKER_FILE_NAME="Dockerfile"


function clean() {
  rm -rf "${TOP_DIR}"/output
  mkdir -p "${TOP_DIR}"/output
}

function build() {
  cd "${TOP_DIR}/cmd/npu-exporter"
  CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go build -mod=mod -buildmode=pie -ldflags "-s -extldflags=-Wl,-z,now  -X huawei.com/npu-exporter/hwlog.BuildName=${OUTPUT_NAME} \
            -X huawei.com/npu-exporter/hwlog.BuildVersion=${build_version}_linux-${arch}" \
    -o ${OUTPUT_NAME}
  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find npu-exporter"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}"/cmd/npu-exporter/${OUTPUT_NAME} "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/npu-exporter.yaml "${TOP_DIR}"/output/npu-exporter-"${build_version}".yaml
  sed -i "s/npu-exporter:.*/npu-exporter:${build_version}/" "${TOP_DIR}"/output/npu-exporter-"${build_version}".yaml
  # need CI prepare so lib before excute build.sh
  cp -r "${TOP_DIR}"/lib "${TOP_DIR}"/output/ || true
  cp "${TOP_DIR}"/build/${DOCKER_FILE_NAME} "${TOP_DIR}"/output
  chmod 400 "${TOP_DIR}"/output/*
  chmod 550 "${TOP_DIR}"/output/lib
  chmod 500 "${TOP_DIR}"/output/lib/*
  chmod 500 "${TOP_DIR}"/output/${OUTPUT_NAME}

}

function main() {
  clean
  build
  mv_file
  bash "${TOP_DIR}"/build/buildtool.sh
}

if [ "$1" = clean ]; then
  clean
  exit 0
fi
main
