package authbe

import (
	"context"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/go-easyops/utils"
	"os"
)

var (
	nonsignedusers = flag.Bool("non_signed_users", true, "if true support unsigned (old-style) signed user accounts, otherwise only signed ones")
	backend        = flag.String("backend", "none", "backend to use: any|none|postgres")
	service        = flag.Bool("service", false, "When using the 'any' backend, produce Service Accounts")
	default_domain = flag.String("service_email_suffix", ".services@conradwood.net", "default suffix for new service email addresses")
	debug          = flag.Bool("debug_backend", false, "debug authentication backend")
	print_failures = flag.Bool("debug_fails", false, "print authentication failures")
	cloudname      = flag.String("cloudname", "yacloud.eu", "set the cloudname so that config and credentials can be seperated by cloud on clients' machines")
)

func New() Authenticator {
	var authBE Authenticator
	if *backend == "postgres" {
		ps := &PostgresAuthenticator{}
		err := ps.Start()
		utils.Bail("failed to start psql", err)
		authBE = ps
	} else {
		fmt.Printf("Invalid backend \"%s\"\n", *backend)
		os.Exit(10)
	}
	if authBE == nil {
		fmt.Printf("Unable to initialise backend\n")
		os.Exit(10)
	}
	return authBE
}

type Authenticator interface {
	pb.AuthenticationServiceServer
	GetUserByEmail(ctx context.Context, req *pb.ByEmailRequest) (*pb.User, error)
	SignedGetUserByID(ctx context.Context, req *pb.ByIDRequest) (*pb.SignedUser, error)
	SignedGetUserByEmail(ctx context.Context, req *pb.ByEmailRequest) (*pb.SignedUser, error)
	GetUserByID(ctx context.Context, req *pb.ByIDRequest) (*pb.User, error)
	CreateService(ctx context.Context, req *pb.CreateServiceRequest) (*pb.NewService, error)
	CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error)
	ResetPasswordEmail(ctx context.Context, req *pb.ResetRequest) (*common.Void, error)
	ExpireToken(ctx context.Context, req *pb.ExpireTokenRequest) (*common.Void, error)
	// get a new token for current (ONLY!) user
	GetTokenForMe(ctx context.Context, req *pb.GetTokenRequest) (*pb.TokenResponse, error)
	ListAllGroups(ctx context.Context) (*pb.GroupList, error)
	Sudo(ctx context.Context, req *pb.SudoRequest) error
	GetGroupByID(ctx context.Context, req *pb.GetGroupRequest) (*pb.Group, error)
	StoreRemote(ctx context.Context, req *pb.RemoteStoreRequest) (*common.Void, error)
	UserByRemoteToken(ctx context.Context, req *pb.RemoteUserRequest) (*pb.RemoteUser, error)
	GetMyRemoteDetails(ctx context.Context, req *common.Void) (*pb.RemoteUser, error)
	LogMeOut(ctx context.Context, req *common.Void) (*pb.User, error)
	LogSomeoneOut(ctx context.Context, req *pb.ByIDRequest) (*pb.User, error)
	GetByAbbreviation(ctx context.Context, req *pb.ByAbbrevRequest) (*pb.User, error)
	GetAllUsers(ctx context.Context, req *common.Void) (*pb.UserList, error)
	CreateSession(ctx context.Context, req *common.Void) (*pb.Session, error)
	KeepAliveSession(ctx context.Context, req *pb.SessionToken) (*pb.Session, error)
}

func ResponseToSignedResponse(a *pb.AuthResponse) (*pb.SignedAuthResponse, error) {
	res := &pb.SignedAuthResponse{
		Valid:         a.Valid,
		PublicMessage: a.PublicMessage,
		LogMessage:    a.LogMessage,
		Token:         a.Token,
	}
	su, err := userToSignedUser(a.User)
	if err != nil {
		return nil, err
	}
	res.User = su
	return res, nil
}

func AddStandardGroups(user *pb.User) {
	g := &pb.Group{ID: "all", Name: "AllUsers", Description: "All users are automatically members of this group"}
	user.Groups = append(user.Groups, g)
}
