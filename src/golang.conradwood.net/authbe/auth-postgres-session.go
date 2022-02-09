package authbe

import (
	"context"
	ed "crypto/ed25519"
	"fmt"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/authdb/db"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/utils"
	"time"
)

func sess_userid(suser *pb.SignedUser) string {
	if suser == nil {
		return ""
	}
	user := &pb.User{}
	err := utils.UnmarshalBytes(suser.User, user)
	if err != nil {
		fmt.Printf("Invalid user: %s\n", err)
		return ""
	}
	return user.ID

}
func GetSession(ctx context.Context, token string, suser *pb.SignedUser) (*pb.PersistSession, error) {
	psa, err := db.DefaultDBPersistSession().ByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if len(psa) != 0 {
		p := psa[0]
		if p.UserID == "" && suser != nil {
			u := sess_userid(suser)
			if u != "" {
				p.UserID = u
				err = db.DefaultDBPersistSession().Update(ctx, p)
				fmt.Printf("Linked session %s to user %s\n", token, p.UserID)
				if err != nil {
					fmt.Printf("Failed to update session: %s\n", err)
				}
			}
		}
		return psa[0], nil
	}
	u := sess_userid(suser)
	ps := &pb.PersistSession{Token: token, UserID: u, Created: uint32(time.Now().Unix())}
	_, err = db.DefaultDBPersistSession().Save(ctx, ps)
	if err != nil {
		fmt.Printf("Failed to persist session: %s\n", err)
	}
	return ps, nil
}
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
	tok := req.Token
	GetSession(ctx, tok, req.User)
	psa, err := db.DefaultDBPersistSession().ByToken(ctx, tok)
	if err != nil {
		return nil, err
	}
	if len(psa) == 0 {
		return nil, errors.NotFound(ctx, "session token not found")
	}
	sess := &pb.Session{Token: tok}
	b, err := utils.MarshalBytes(sess)
	if err != nil {
		return nil, err
	}
	s := ed.Sign(signPrivateKey(), b)
	signed := &pb.SignedSession{Session: b, Signature: s}
	return signed, nil
}
