package main

import (
	"flag"
	"fmt"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/authremote"
	//	"golang.conradwood.net/go-easyops/client"
	//	pb "golang.conradwood.net/apis/auth"
	gc "golang.conradwood.net/go-easyops/common"
	"golang.conradwood.net/go-easyops/utils"
	"os"
)

var (
	debug       = flag.Bool("debug", false, "debug mode")
	print_keys  = flag.Bool("print_keys", false, "also print ssh keys")
	unix        = flag.Bool("unix", false, "also get unix user details")
	email_only  = flag.Bool("email_only", false, "if true only print my email address")
	bash_prompt = flag.Bool("bash_prompt", false, "change behaviour to be suitable for inclusion as a bash prompt")
)

func isQuiet() bool {
	return *email_only
}

func main() {
	flag.Parse()
	//	client.GetSignatureFromAuth()
	if !isQuiet() {
		fmt.Printf("You are a human being\n")
	}
	/*
		s := tokens.GetUserTokenParameter()
		if s == "" && os.Getenv("GE_CTX") == "" {
			fmt.Printf("You have no authentication token. (missing file ~/.go-easyops/tokens/user_token ?)\n")
			os.Exit(10)
		}
	*/
	if *debug {
		fmt.Printf("Creating context...\n")
	}
	ctx := authremote.Context()
	if ctx == nil {
		fmt.Printf("No user information available\n")
		os.Exit(10)
	}
	u := auth.GetUser(ctx)
	if *debug {
		fmt.Printf("Got context for \"%s\"\n", auth.Description(u))
	}
	if !isQuiet() {
		auth.PrintUser(auth.GetUser(ctx))
	}
	if *debug {
		fmt.Printf("Calling authmanager...\n")
	}
	u, err := authremote.GetAuthManagerClient().WhoAmI(ctx, &common.Void{})
	utils.Bail("failed to get user account", err)
	if !isQuiet() {
		fmt.Printf("Cloud  : %s\n", gc.GetCloudName())
		fmt.Printf("UserID : %s\n", u.ID)
		fmt.Printf("Name   : %s %s\n", u.FirstName, u.LastName)
		fmt.Printf("Email  : %s\n", u.Email)
		fmt.Printf("Org-ID : \"%s\"\n", u.OrganisationID)
		fmt.Printf("Created: %s\n", utils.TimestampString(u.Created))
		/*
			var keys [][]byte
				if *unix || *print_keys {
					unx, err := authremote.GetAuthManagerClient().GetUnixData(ctx, &pb.ByIDRequest{UserID: u.ID})
					utils.Bail("failed to get unix data", err)
					fmt.Printf("Unixid : %d\n", unx.Uid)
					fmt.Printf("SSHkeys: %d\n", len(unx.SSHKeys))
					keys = unx.SSHKeys
			}
		*/
		fmt.Printf("Groups:\n")
		for _, g := range u.Groups {
			fmt.Printf("  %4s (%s)\n", g.ID, g.Name)
		}
		/*
			if *print_keys && len(keys) > 0 {
				for i, k := range keys {
					fmt.Printf("%d. Key:\n%s\n", i+1, k)
				}
			}
		*/
		sig := gc.VerifySignature(u)
		if !sig {
			fmt.Printf("WARNING: Signature is not valid, information is not trustworthy\n")
			fmt.Printf("Signature: %v\n", u.SignatureFull)
		}
	}
	if *email_only {
		fmt.Printf("%s\n", u.Email)
	}
}
