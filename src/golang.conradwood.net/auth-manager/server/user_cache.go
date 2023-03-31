package main

import (
	"context"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/go-easyops/common"
	"sync"
	"time"
)

var (
	cached_user_lock sync.Mutex
	cached_users     []*user_cache_entry
)

type user_cache_entry struct {
	refreshed   time.Time
	userid      string
	signed_user *pb.SignedUser
	user        *pb.User
}

func (uc *user_cache_entry) isValid() bool {
	if time.Since(uc.refreshed) > time.Duration(30)*time.Second {
		return false
	}
	return true
}

func usercache_SignedGetUserByID(ctx context.Context, req *pb.ByIDRequest) (*pb.SignedUser, error) {
	for _, uc := range cached_users {
		if uc.isValid() && uc.userid == req.UserID {
			return uc.signed_user, nil
		}
	}
	res, err := authBE.SignedGetUserByID(ctx, req)
	if err != nil {
		return nil, err
	}
	usercache_add(res)
	return res, nil
}
func usercache_GetUserByID(ctx context.Context, req *pb.ByIDRequest) (*pb.User, error) {
	for _, uc := range cached_users {
		if uc.isValid() && uc.userid == req.UserID {
			return uc.user, nil
		}
	}
	res, err := authBE.SignedGetUserByID(ctx, req)
	if err != nil {
		return nil, err
	}
	u := usercache_add(res)
	return u, nil
}

func usercache_add(u *pb.SignedUser) *pb.User {
	user := common.VerifySignedUser(u)
	if user == nil {
		return nil
	}
	userid := user.ID
	cached_user_lock.Lock()
	defer cached_user_lock.Unlock()
	for _, uc := range cached_users {
		if uc.userid == userid {
			// do not allocate new one, reuse entry
			uc.user = user
			uc.signed_user = u
			uc.refreshed = time.Now()
			return uc.user
		}
	}
	uc := &user_cache_entry{userid: userid, signed_user: u, user: user, refreshed: time.Now()}
	cached_users = append(cached_users, uc)
	return uc.user
}
