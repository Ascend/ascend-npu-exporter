module huawei.com/npu-exporter

go 1.14

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prashantv/gostub v1.0.1-0.20191007164320-bbe3712b9c4a
	github.com/prometheus/client_golang v1.5.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	huawei.com/kmc v1.0.6
)

replace huawei.com/kmc => codehub-dg-y.huawei.com/it-edge-native/edge-native-core/coastguard.git v1.0.6
