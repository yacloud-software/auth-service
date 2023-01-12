package authbe

import (
    "golang.conradwood.net/go-easyops/authremote"
	"fmt"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/go-easyops/utils"
	"time"
)

const (
	EXPIRE_INTERVAL = 60
)

func pg_expirer() {
	last_expire_tokens := time.Now()
	for {
		if time.Since(last_expire_tokens) > time.Duration(EXPIRE_INTERVAL)*time.Minute {
			last_expire_tokens = time.Now()
			err := pg_expire_tokens()
			if err != nil {
				fmt.Printf("failed to expire tokens: %s\n", utils.ErrorString(err))
			}
		}
		time.Sleep(time.Duration(5) * time.Minute)
		ctx := authremote.Context()
		sts, err := sudoers.All(ctx)
		if err != nil {
			fmt.Printf("failed to get all sudos: %s", err)
			continue
		}
		now := uint32(time.Now().Unix())
		for _, s := range sts {
			if s.Expiry >= now {
				continue
			}
			err := sudoers.DeleteByID(ctx, s.ID)
			fmt.Printf("Failed to delete sudoers #%d: %s\n", s.ID, err)
		}

	}
}
func pg_expire_tokens() error {
	ctx := authremote.Context()
	cutoff := time.Now().Add(0 - time.Duration(24)*time.Hour)
	cf := cutoff.Unix()
	_, err := psql.ExecContext(ctx, "expire_tokens", "delete from tokens where expiry < $1 and tokentype = $2", cf, pb.TokenType_SESSION)
	return err
}
