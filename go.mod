module huawei.com/npu-exporter

go 1.14

require (
	github.com/containerd/containerd v1.5.5
	github.com/docker/docker v20.10.0+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/protobuf v1.5.1
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prashantv/gostub v1.0.1-0.20191007164320-bbe3712b9c4a
	github.com/prometheus/client_golang v1.7.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	google.golang.org/grpc v1.33.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	huawei.com/kmc v1.0.6
	k8s.io/cri-api v0.20.6
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.5.0
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	google.golang.org/grpc => github.com/grpc/grpc-go v1.28.0
	huawei.com/kmc => codehub-dg-y.huawei.com/it-edge-native/edge-native-core/coastguard.git v1.0.6
	k8s.io/cri-api => k8s.io/cri-api v0.19.4
)
