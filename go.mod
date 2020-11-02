module github.com/jukylin/nx

go 1.14

require (
	github.com/apache/rocketmq-client-go/v2 v2.0.0
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/jaegertracing/jaeger v1.20.0
	github.com/jhump/protoreflect v1.7.0 // indirect
	github.com/jukylin/esim v0.1.9
	github.com/magiconair/properties v1.8.4
	github.com/mercari/grpc-http-proxy v0.1.2
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ory/dockertest/v3 v3.6.2
	github.com/prometheus/client_golang v1.8.0
	github.com/sony/sonyflake v1.0.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	go.etcd.io/etcd v3.3.25+incompatible
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.33.1
	google.golang.org/grpc/examples v0.0.0-20201020200225-9519efffeb5d
	gorm.io/gorm v1.20.3
)

replace github.com/jukylin/nx/sagas/domain/entity => /data/go/src/github.com/jukylin/nx/sagas/domain/entity
