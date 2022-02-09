package authbe

import (
	"context"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/go-easyops/utils"
)

func (a *PostgresAuthenticator) CreateSession(ctx context.Context, req *common.Void) (*pb.Session, error) {
	sess := utils.RandomString(128)
	return &pb.Session{Token: sess}, nil
}
func (a *PostgresAuthenticator) KeepAliveSession(ctx context.Context, req *pb.SessionToken) (*pb.Session, error) {
	sess := utils.RandomString(128)
	return &pb.Session{Token: sess}, nil
}
