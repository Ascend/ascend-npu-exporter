#!/bin/bash
# Perform  build npu-exporter
# Copyright @ Huawei Technologies CO., Ltd. 2020-2021. All rights reserved

set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
unset GOPATH
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="v2.0.1"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '5p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*:}
fi
arch=$(arch 2>&1)
echo "Build Architecture  is" "${arch}"
if [ "${arch:0:5}" = 'aarch' ]; then
  arch=arm64
else
  arch=amd64
fi

sed -i "s/npu-exporter:.*/npu-exporter:${build_version}/" "${TOP_DIR}"/build/npu-exporter.yaml

OUTPUT_NAME="npu-exporter"
DOCKER_FILE_NAME="Dockerfile"
docker_zip_name="npu-exporter-${build_version}-${arch}.tar.gz"
docker_images_name="npu-exporter:${build_version}"
function clear_env() {
  rm -rf "${TOP_DIR}"/output
  mkdir -p "${TOP_DIR}"/output
}

function build() {
  cd "${TOP_DIR}"
  go build -ldflags "-X huawei.com/npu-exporter/collector.BuildName=${OUTPUT_NAME} \
            -X huawei.com/npu-exporter/collector.BuildVersion=${build_version}" \
    -o ${OUTPUT_NAME}
  ls ${OUTPUT_NAME}
  if [ $? -ne 0 ]; then
    echo "fail to find npu-exporter"
    exit 1
  fi
}

function mv_file() {
  mv "${TOP_DIR}"/${OUTPUT_NAME} "${TOP_DIR}"/output
  cp "${TOP_DIR}"/build/npu-exporter.yaml "${TOP_DIR}"/output/npu-exporter-"${build_version}".yaml
}

function build_docker_image() {
  cp "${TOP_DIR}"/build/${DOCKER_FILE_NAME} "${TOP_DIR}"/output
  cd "${TOP_DIR}"/output
  docker rmi "${docker_images_name}" || true
  docker build -t "${docker_images_name}" --no-cache .
  docker save "${docker_images_name}" | gzip >"${docker_zip_name}"
  rm -f ${DOCKER_FILE_NAME}
}

function main() {
  clear_env
  build
  mv_file
  build_docker_image
}

main
