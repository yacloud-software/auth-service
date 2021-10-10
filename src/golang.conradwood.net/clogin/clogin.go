package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	ap "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/tokens"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"strings"
)

var (
	perm = flag.Bool("permanent", false, "if true, creates a (semi) permanent token")
)

func main() {
	flag.Parse()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your username (email): ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Please enter your password: ")
	password, _ := reader.ReadString('\n')
	username = strings.Trim(username, "\n")
	password = strings.Trim(password, "\n")
	fmt.Printf("Username: \"%s\", Password: \"%s\"\n", username, password)
	tr := &ap.AuthenticatePasswordRequest{
		Email:    username,
		Password: password,
	}
	resp, err := authremote.GetAuthClient().SignedGetByPassword(context.Background(), tr)
	utils.Bail("authentication failed", err)
	if !resp.Valid {
		fmt.Printf("Authentication failure: %s\n", resp.PublicMessage)
		return
	}
	new_token := resp.Token
	fmt.Printf("Saving your temporary token: %s\n", new_token)
	err = tokens.SaveUserToken(new_token)
	utils.Bail("failed to save your token", err)

	if *perm {
		p := &ap.GetTokenRequest{DurationSecs: 60 * 24 * 365 * 3}
		resp, err := authremote.GetAuthManagerClient().GetTokenForMe(authremote.Context(), p)
		utils.Bail("failed to create longterm token", err)
		new_token = resp.Token
		fmt.Printf("Saving your permanent token: %s\n", new_token)
		err = tokens.SaveUserToken(new_token)
		utils.Bail("failed to save your token", err)
	}

}
