package authbe

import (
	"context"
	ed "crypto/ed25519"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/go-easyops/utils"
)

func (a *PostgresAuthenticator) CreateSession(ctx context.Context, req *common.Void) (*pb.SignedSession, error) {
	sess := &pb.Session{Token: utils.RandomString(128)}
	b, err := utils.MarshalBytes(sess)
	if err != nil {
		return nil, err
	}
	s := ed.Sign(signPrivateKey(), b)
	signed := &pb.SignedSession{Session: b, Signature: s}
	return signed, nil
}
func (a *PostgresAuthenticator) KeepAliveSession(ctx context.Context, req *pb.KeepAliveSessionRequest) (*pb.SignedSession, error) {
	sess := &pb.Session{Token: req.Token}
	b, err := utils.MarshalBytes(sess)
	if err != nil {
		return nil, err
	}
	s := ed.Sign(signPrivateKey(), b)
	signed := &pb.SignedSession{Session: b, Signature: s}
	return signed, nil
}
