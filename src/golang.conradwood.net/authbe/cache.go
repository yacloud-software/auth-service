package authbe

import (
	"flag"
	"fmt"
	apb "golang.conradwood.net/apis/auth"
	"sync"
	"time"
)

const (
	USER_ENTRY_MAX_AGE = time.Duration(2) * time.Minute
	USER_TOKEN_MAX_AGE = time.Duration(5) * time.Minute
)

var (
	debug_user_cache = flag.Bool("debug_user_cache", false, "debug the user cache")
	uclock           sync.Mutex
	user_cache       = &UserCache{} // hashmap might be better. key? perhaps userid?
)

type UserCache struct {
	cache_entries []*UserCacheEntry
}

type UserCacheEntry struct {
	sync.Mutex
	created time.Time
	updated time.Time
	user    *apb.User
	tokens  []*cached_token
}
type cached_token struct {
	created time.Time
	updated time.Time
	token   string
}

func (uce *UserCacheEntry) IsTooOld() bool {
	return time.Since(uce.updated) > USER_ENTRY_MAX_AGE
}
func (ct *cached_token) IsTooOld() bool {
	return time.Since(ct.updated) > USER_TOKEN_MAX_AGE
}

func (uc *UserCache) getuser(user *apb.User) *UserCacheEntry {
	uclock.Lock()
	defer uclock.Unlock()
	for _, ue := range uc.cache_entries {
		if ue.user.ID == user.ID {
			ue.user = user
			ue.updated = time.Now()
			return ue
		}
	}
	uce := &UserCacheEntry{
		created: time.Now(),
		updated: time.Now(),
		user:    user,
	}
	uc.cache_entries = append(uc.cache_entries, uce)
	return uce

}

func (uc *UserCache) SaveUserWithToken(user *apb.User, token string) {
	uce := uc.getuser(user)
	uce.addToken(token)
	uc.debugf("added token %s for user %s\n", token, user.ID)
}

func (uce *UserCacheEntry) addToken(token string) {
	uce.Lock()
	defer uce.Unlock()
	for _, t := range uce.tokens {
		if t.token != token {
			continue
		}
		t.updated = time.Now()
		return
	}
	ct := &cached_token{
		created: time.Now(),
		updated: time.Now(),
		token:   token,
	}
	uce.tokens = append(uce.tokens, ct)
}
func (uc *UserCache) GetUserByToken(token string) *apb.User {
	uclock.Lock()
	defer uclock.Unlock()
	for _, uce := range uc.cache_entries {
		if uce.IsTooOld() {
			continue
		}
		for _, t := range uce.tokens {
			if t.token != token {
				continue
			}
			if t.IsTooOld() {
				continue
			}
			uc.debugf("returned user from cache\n")
			return uce.user
		}
	}
	uc.debugf("user not in cache\n")
	return nil
}

func (uc *UserCache) debugf(format string, args ...interface{}) {
	if !*debug_user_cache {
		return
	}
	s := fmt.Sprintf(format, args...)
	fmt.Printf("[usercache] %s", s)
}
