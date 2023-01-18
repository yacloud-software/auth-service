module golang.conradwood.net/authbe

go 1.18

replace golang.conradwood.net/authbe => ../../golang.conradwood.net/authbe

replace golang.conradwood.net/apis/auth => ../../golang.conradwood.net/apis/auth

replace golang.conradwood.net/authdb => ../../golang.conradwood.net/authdb

require (
	github.com/golang/protobuf v1.5.2
	golang.conradwood.net/apis/auth v1.1.2073
	golang.conradwood.net/apis/common v1.1.2106
	golang.conradwood.net/apis/email v1.1.2106
	golang.conradwood.net/apis/slackgateway v1.1.2106
	golang.conradwood.net/authdb v0.0.0-00010101000000-000000000000
	golang.conradwood.net/go-easyops v0.1.15450
	golang.org/x/crypto v0.5.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	golang.conradwood.net/apis/autodeployer v1.1.2106 // indirect
	golang.conradwood.net/apis/deploymonkey v1.1.2106 // indirect
	golang.conradwood.net/apis/framework v1.1.2106 // indirect
	golang.conradwood.net/apis/goeasyops v1.1.2106 // indirect
	golang.conradwood.net/apis/h2gproxy v1.1.2106 // indirect
	golang.conradwood.net/apis/objectstore v1.1.2106 // indirect
	golang.conradwood.net/apis/registry v1.1.2106 // indirect
	golang.conradwood.net/apis/rpcinterceptor v1.1.2106 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6 // indirect
	google.golang.org/grpc v1.52.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)
