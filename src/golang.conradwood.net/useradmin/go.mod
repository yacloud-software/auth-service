module golang.conradwood.net/useradmin

go 1.21.1

replace golang.conradwood.net/authbe => ../../golang.conradwood.net/authbe

replace golang.conradwood.net/apis/auth => ../../golang.conradwood.net/apis/auth

replace golang.conradwood.net/authdb => ../../golang.conradwood.net/authdb

require (
	golang.conradwood.net/apis/auth v1.1.2901
	golang.conradwood.net/apis/common v1.1.2916
	golang.conradwood.net/go-easyops v0.1.27487
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.50.0 // indirect
	github.com/prometheus/procfs v0.13.0 // indirect
	golang.conradwood.net/apis/autodeployer v1.1.2878 // indirect
	golang.conradwood.net/apis/deploymonkey v1.1.2878 // indirect
	golang.conradwood.net/apis/framework v1.1.2878 // indirect
	golang.conradwood.net/apis/goeasyops v1.1.2916 // indirect
	golang.conradwood.net/apis/grafanadata v1.1.2878 // indirect
	golang.conradwood.net/apis/objectstore v1.1.2878 // indirect
	golang.conradwood.net/apis/registry v1.1.2878 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.yacloud.eu/apis/fscache v1.1.2897 // indirect
	golang.yacloud.eu/apis/session v1.1.2916 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240227224415-6ceb2ff114de // indirect
	google.golang.org/grpc v1.63.2 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
