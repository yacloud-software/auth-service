package main

import (
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/authbe"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/utils"
	"google.golang.org/grpc"
)

// static variables for flag parser
var (
	port   = flag.Int("port", 4998, "The server port")
	authBE authbe.Authenticator
)

/**************************************************
* helpers
***************************************************/
// main

func main() {
	flag.Parse() // parse stuff. see "var" section above
	authBE = authbe.New()
	err := start()
	utils.Bail("failed", err)
}

func st(server *grpc.Server) error {
	// Register the handler object
	pb.RegisterAuthenticationServiceServer(server, authBE)
	return nil
}
func start() error {
	var err error

	sd := server.NewServerDef()
	sd.SetPort(*port)
	// we ARE the authentication service so don't insist on authenticated calls
	sd.SetNoAuth()

	sd.SetRegister(st)
	err = server.ServerStartup(sd)
	if err != nil {
		fmt.Printf("failed to start server: %s\n", err)
	}
	fmt.Printf("Done\n")
	return nil
}
