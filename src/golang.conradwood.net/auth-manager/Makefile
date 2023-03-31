.PHONY:	server
all:	server
server:
	cd server ; go install ${LDFLAGS} auth-manager-server.go `ls -1 *.go|grep -v auth-manager-server.go`
