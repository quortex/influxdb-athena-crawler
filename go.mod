module github.com/quortex/influxdb-athena-crawler

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.20.1
	github.com/aws/aws-sdk-go-v2/config v1.15.9
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.14
	github.com/aws/aws-sdk-go-v2/service/s3 v1.38.2
	github.com/influxdata/influxdb-client-go/v2 v2.9.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/rs/zerolog v1.26.1
)
