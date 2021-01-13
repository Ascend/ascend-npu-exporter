#!/bin/bash
# Perform  test for  npu-exporter
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
set -e


# execute go test and echo result to report files
function execute_test() {
  if ! (go test -v -race -coverprofile cov.out "${TOP_DIR}"/collector >./"$file_input")
  then
    echo '****** go test cases error! ******'
    echo 'Failed' >"$file_input"
  else
    gocov convert cov.out | gocov-html >"$file_detail_output"
  fi

  {
    echo "<html<body><h1>==================================================</h1><table border='2'>"
    echo "<html<body><h1>npu exporter testCase</h1><table border='1'>"
    echo "<html<body><h1>==================================================</h1><table border='2'>"
  } >>./"$file_detail_output"

  while read -r line
  do
    echo -e "<tr>
    $(echo "$line" | awk 'BEGIN{FS="|"}''{i=1;while(i<=NF) {print "<td>"$i"</td>";i++}}')
    </tr>" >>"$file_detail_output"
  done <"$file_input"
  echo "</table></body></html>" >>./"$file_detail_output"
}


export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH
unset GOPATH
# if didn't install the following  tools, please install firstly
#go get -insecure github.com/axw/gocov/gocov
#go get github.com/matm/gocov-html
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

file_input='testExporter.txt'
file_detail_output='exporterCoverageReport.html'

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
