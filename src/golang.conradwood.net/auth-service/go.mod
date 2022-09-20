module golang.conradwood.net/auth-service

go 1.17

replace golang.conradwood.net/authbe => ../../golang.conradwood.net/authbe

replace golang.conradwood.net/apis/auth => ../../golang.conradwood.net/apis/auth

replace golang.conradwood.net/authdb => ../../golang.conradwood.net/authdb

require (
	golang.conradwood.net/apis/auth v1.1.1787
	golang.conradwood.net/authbe v0.0.0-00010101000000-000000000000
	golang.conradwood.net/go-easyops v0.1.13564
	google.golang.org/grpc v1.49.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_golang v1.13.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	golang.conradwood.net/apis/autodeployer v1.1.1842 // indirect
	golang.conradwood.net/apis/common v1.1.1842 // indirect
	golang.conradwood.net/apis/deploymonkey v1.1.1842 // indirect
	golang.conradwood.net/apis/echoservice v1.1.1842 // indirect
	golang.conradwood.net/apis/email v1.1.1842 // indirect
	golang.conradwood.net/apis/errorlogger v1.1.1842 // indirect
	golang.conradwood.net/apis/framework v1.1.1842 // indirect
	golang.conradwood.net/apis/h2gproxy v1.1.1842 // indirect
	golang.conradwood.net/apis/objectstore v1.1.1842 // indirect
	golang.conradwood.net/apis/registry v1.1.1842 // indirect
	golang.conradwood.net/apis/rpcinterceptor v1.1.1842 // indirect
	golang.conradwood.net/apis/slackgateway v1.1.1842 // indirect
	golang.conradwood.net/authdb v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/crypto v0.0.0-20220919173607-35f4265a4bc0 // indirect
	golang.org/x/net v0.0.0-20220919232410-f2f64ebce3c1 // indirect
	golang.org/x/sys v0.0.0-20220919091848-fb04ddd9f9c8 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220919141832-68c03719ef51 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)
