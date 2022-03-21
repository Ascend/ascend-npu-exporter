#!/bin/bash
# Perform  test for  npu-exporter
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
set -e


# execute go test and echo result to report files
function execute_test() {
  if ! (go test -v -race -coverprofile cov.out "${TOP_DIR}"/... >./"$file_input")
  then
    echo '****** go test cases error! ******'
    exit 1
  else
    gocov convert cov.out | gocov-html >"$file_detail_output"
    gotestsum --junitfile unit-tests.xml "${TOP_DIR}"/collector/...
    exit 0
  fi
}


export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH
export GOFLAGS="-gcflags=all=-l"
unset GOPATH
# if didn't install the following  tools, please install firstly
#go get -insecure github.com/axw/gocov/gocov
#go get github.com/matm/gocov-html
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

file_input='testExporter.txt'
file_detail_output='api.html'

if [ -f "${TOP_DIR}"/test ]; then
  rm -rf "${TOP_DIR}"/test
fi
mkdir -p "${TOP_DIR}"/test
cd "${TOP_DIR}"/test
echo "clean old version test results"

if [ -f "$file_input" ]; then
  rm -rf "$file_input"
fi
if [ -f "$file_detail_output" ]; then
  rm -rf "$file_detail_output"
fi

echo "************************************* Start LLT Test *************************************"
execute_test
echo "************************************* End   LLT Test *************************************"
