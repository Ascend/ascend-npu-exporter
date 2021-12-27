#!/bin/bash
# Perform  build kmc-ext
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved

set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)



cd "${TOP_DIR}"/third_party/build/build

cmake CMakeLists.txt -DUSE_LOCAL_OPENSSL=ON  -DOPENSSL_TARBALL_PATH="${TOP_DIR}"/third_party/external/opensource/openssl \
    -DHW_SECUREC_DIR="${TOP_DIR}"/third_party/external/huawei/securec ..

make clean
make
rm -rf "${TOP_DIR}"/lib
mv  -f "${TOP_DIR}"/third_party/build/release/lib "${TOP_DIR}"
