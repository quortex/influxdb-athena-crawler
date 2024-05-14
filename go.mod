module github.com/quortex/influxdb-athena-crawler

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.26.1
	github.com/aws/aws-sdk-go-v2/config v1.27.13
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.16.17
	github.com/aws/aws-sdk-go-v2/service/s3 v1.53.2
	github.com/influxdata/influxdb-client-go/v2 v2.9.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/rs/zerolog v1.32.0
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)
