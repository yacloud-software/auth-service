package authbe

import (
	"context"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/go-easyops/utils"
)

func (a *PostgresAuthenticator) CreateSession(ctx context.Context, req *common.Void) (*pb.SessionToken, error) {
	sess := utils.RandomString(128)
	return &pb.SessionToken{Token: sess}, nil
}
func (a *PostgresAuthenticator) KeepAliveSession(ctx context.Context, req *pb.SessionToken) (*pb.SessionToken, error) {
	sess := utils.RandomString(128)
	return &pb.SessionToken{Token: sess}, nil
}
