package authbe

import (
	"context"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/authdb/db"
	au "golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/errors"
	"time"
)

var (
	remoteuserdb *db.DBRemoteUserDetail
)

func (a *PostgresAuthenticator) StoreRemote(ctx context.Context, req *pb.RemoteStoreRequest) (*common.Void, error) {
	ru := &pb.RemoteUserDetail{
		UserID:       req.UserID,
		Provider:     req.Provider,
		OurToken:     req.OurToken,
		RemoteUserID: req.RemoteUserID,
		Created:      uint32(time.Now().Unix()),
	}
	_, err := remoteuserdb.Save(ctx, ru)
	if err != nil {
		return nil, err
	}
	return &common.Void{}, nil
}
func (a *PostgresAuthenticator) UserByRemoteToken(ctx context.Context, req *pb.RemoteUserRequest) (*pb.RemoteUser, error) {
	ruds, err := remoteuserdb.ByOurToken(ctx, req.OurToken)
	if err != nil {
		return nil, err
	}
	if len(ruds) == 0 {
		return nil, errors.InvalidArgs(ctx, "no such user", "no such user by token")
	}
	res := &pb.RemoteUser{}
	uid := ruds[0].UserID
	user, err := a.GetUserByID(ctx, &pb.ByIDRequest{UserID: uid})
	if err != nil {
		return nil, err
	}
	res.User = user
	for _, r := range ruds {
		if r.UserID != uid {
			// this is weird. same token, but different singingcat users?
			return nil, errors.FailedPrecondition(ctx, "multiple users for token at remoteuserid %d", ruds[0].ID)
		}
		res.Details = append(res.Details, r)

	}
	return res, nil
}
func (a *PostgresAuthenticator) GetMyRemoteDetails(ctx context.Context, req *common.Void) (*pb.RemoteUser, error) {
	user := au.GetUser(ctx)
	if user == nil {
		return nil, errors.Unauthenticated(ctx, "please log in")
	}
	ruds, err := remoteuserdb.ByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	res := &pb.RemoteUser{User: user}
	for _, r := range ruds {
		res.Details = append(res.Details, r)

	}
	return res, nil

}
