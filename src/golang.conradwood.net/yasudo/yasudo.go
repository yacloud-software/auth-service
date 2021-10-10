package main

import (
	"flag"
	"fmt"
	apb "golang.conradwood.net/apis/auth"
	ar "golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
)

var (
	userid = flag.String("userid", "", "userid to sudo. Default: use current users' userid")
)

func main() {
	flag.Parse()
	fmt.Printf("Yasudo\n")
	ctx := ar.Context()
	sr := &apb.SudoRequest{
		UserID: *userid,
	}
	for i := 0; i < 10; i++ {
		_, err := ar.GetAuthManagerClient().Sudo(ctx, sr)
		utils.Bail("Failed to sudo", err)
	}
	fmt.Printf("Sudo successful\n")

}
