module huawei.com/npu-exporter

go 1.16

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/protobuf v1.5.2
	github.com/patrickmn/go-cache v2.1.1-0.20191004192108-46f407853014+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prashantv/gostub v1.0.1-0.20191007164320-bbe3712b9c4a
	github.com/prometheus/client_golang v1.5.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	google.golang.org/grpc v1.38.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	huawei.com/kmc v1.0.6
	k8s.io/cri-api v0.19.4
)

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.5.1
	google.golang.org/grpc => github.com/grpc/grpc-go v1.28.0
	huawei.com/kmc => codehub-dg-y.huawei.com/it-edge-native/edge-native-core/coastguard.git v1.0.6
	k8s.io/cri-api => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/cri-api v1.19.4-h2
)
