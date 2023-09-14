//go:build !custom || inputs || inputs.npu

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/npu" // register plugin
