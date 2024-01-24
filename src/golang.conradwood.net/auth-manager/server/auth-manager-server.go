package main

import (
	"context"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/email"
	"golang.conradwood.net/authbe"
	"golang.conradwood.net/authdb/db"
	"golang.conradwood.net/go-easyops/auth"
	"strings"
	//	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/server"
	"golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/utils"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"strconv"
	"time"
)

var (
	list_group_services = flag.String("list_group_services", "", "services which are allowed to list all groups")
	//	sender   = flag.String("email_sender", "donotreply@singingcat.net", "the email sender for forgot email emails")
	port     = flag.Int("port", 12312, "grpc port")
	authBE   authbe.Authenticator
	evpstore *db.DBEmailVerifyPins
	dbs      *sql.DB
)

const (
	PIN_LIFETIME_MINUTES = 30
)

func main() {
	var err error
	flag.Parse()
	authBE = authbe.New()
	dbs, err = sql.Open()
	utils.Bail("auth db error", err)
	evpstore = db.NewDBEmailVerifyPins(dbs)
	sd := server.NewServerDef()
	sd.SetPort(*port)
	sd.SetRegister(func(server *grpc.Server) error {
		pb.RegisterAuthManagerServiceServer(server, &am{})
		return nil
	})
	utils.Bail("failed to start server", server.ServerStartup(sd))
}

type am struct {
}

func (a *am) CreateService(ctx context.Context, req *pb.CreateServiceRequest) (*pb.NewService, error) {
	return authBE.CreateService(ctx, req)
}

func (a *am) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	return authBE.CreateUser(ctx, req)
}
func (a *am) CreateFakeUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	return nil, errors.NotImplemented(ctx, "createfakeuser")
}

func (a *am) GetUserByEmail(ctx context.Context, req *pb.ByEmailRequest) (*pb.User, error) {
	return authBE.GetUserByEmail(ctx, req)
}
func (a *am) SignedGetUserByEmail(ctx context.Context, req *pb.ByEmailRequest) (*pb.SignedUser, error) {
	return authBE.SignedGetUserByEmail(ctx, req)
}
func (a *am) SignedGetUserByID(ctx context.Context, req *pb.ByIDRequest) (*pb.SignedUser, error) {
	return usercache_SignedGetUserByID(ctx, req)
}

func (a *am) AddUserToGroup(ctx context.Context, req *pb.AddToGroupRequest) (*common.Void, error) {
	return authBE.AddUserToGroup(ctx, req)
}

func (a *am) ListGroups(ctx context.Context, req *common.Void) (*pb.GroupList, error) {
	e := errors.NeedServiceOrRoot(ctx, GetListGroupServices())
	if e != nil {
		return nil, e
	}

	return authBE.ListAllGroups(ctx)

}
func (a *am) GetUserByID(ctx context.Context, req *pb.ByIDRequest) (*pb.User, error) {
	return usercache_GetUserByID(ctx, req)

}
func (a *am) ForceUpdatePassword(ctx context.Context, req *pb.ForceUpdatePasswordRequest) (*common.Void, error) {
	e := errors.NeedServiceOrRoot(ctx, authbe.GetRootServices())
	if e != nil {
		return nil, e
	}
	if req.UserID == "" {
		return nil, errors.InvalidArgs(ctx, "missing userid", "missing userid")
	}
	e = checkPassword(ctx, req.NewPassword)
	if e != nil {
		return nil, e
	}
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(req.UserID)
	if err != nil {
		return nil, err
	}
	bc, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		fmt.Printf("Failed to encode password: %s\n", err)
		return nil, err
	}

	pw := string(bc)
	_, err = db.ExecContext(ctx, "update_password", "update users set passwd = $1 where id = $2", pw, id)
	if err != nil {
		return nil, err
	}
	return &common.Void{}, nil
}
func (a *am) UpdateMyPassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*common.Void, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, fmt.Errorf("missing user")
	}
	e := checkPassword(ctx, req.NewPassword)
	if e != nil {
		return nil, e
	}
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(u.ID)
	if err != nil {
		return nil, err
	}
	bc, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 10)
	if err != nil {
		fmt.Printf("Failed to encode password: %s\n", err)
		return nil, err
	}

	pw := string(bc)
	_, err = db.ExecContext(ctx, "update_password", "update users set passwd = $1 where id = $2", pw, id)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Updated password for user %d\n", id)
	return &common.Void{}, nil
}
func (a *am) ResetPasswordEmail(ctx context.Context, req *pb.ResetRequest) (*common.Void, error) {
	return authBE.ResetPasswordEmail(ctx, req)
}
func (a *am) ExpireToken(ctx context.Context, req *pb.ExpireTokenRequest) (*common.Void, error) {
	return authBE.ExpireToken(ctx, req)
}
func (a *am) GetTokenForMe(ctx context.Context, req *pb.GetTokenRequest) (*pb.TokenResponse, error) {
	return authBE.GetTokenForMe(ctx, req)
}
func (a *am) GetTokenForService(ctx context.Context, req *pb.GetTokenRequest) (*pb.TokenResponse, error) {
	return authBE.GetTokenForService(ctx, req)
}

func (a *am) SendEmailVerify(ctx context.Context, req *common.Void) (*common.Void, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "no user to send email to")
	}
	uid, err := strconv.Atoi(u.ID)
	if err != nil {
		return nil, err
	}
	emailadr := u.Email
	pinno := utils.RandomInt(89999) + 10000
	pin := fmt.Sprintf("%d", pinno)
	_, err = evpstore.Save(ctx, &pb.EmailVerifyPins{
		UserID:  uint64(uid),
		Pin:     pin,
		Created: uint64(time.Now().Unix()),
	})
	if err != nil {
		return nil, err
	}
	es := email.GetEmailServiceClient()
	ter := &email.TemplateEmailRequest{Sender: authbe.GetEmailSender(),
		Recipient:    emailadr,
		TemplateName: "verifyemail",
		Values:       make(map[string]string),
	}
	ter.Values["pin"] = pin
	ter.Values["validuntil"] = utils.TimeString(time.Now().Add(time.Duration(PIN_LIFETIME_MINUTES) * time.Minute))
	tr, err := es.SendTemplate(ctx, ter)
	if err != nil {
		fmt.Printf("Send Email to %s failed: %s\n", emailadr, err)
		return nil, err
	}
	if !tr.Success {
		fmt.Printf("Send Email unsuccessfull to %s: %s\n", emailadr, err)
		return nil, fmt.Errorf("Failed to send email to %s", emailadr)
	}
	fmt.Printf("Send pin verification email to user email %s\n", emailadr)
	return req, nil
}
func (a *am) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "no user to send email to")
	}

	uid, err := strconv.Atoi(u.ID)
	if err != nil {
		return nil, err
	}
	ps, err := evpstore.ByUserID(ctx, uint64(uid))
	if err != nil {
		return nil, err
	}
	dout := fmt.Sprintf("[pinverification, user=%s, pin=%s]", u.Email, req.Pin)
	for _, evp := range ps {
		if evp.Pin != req.Pin {
			continue
		}
		if evp.Accepted != 0 {
			fmt.Printf("%s - used previously\n", dout)
			return &pb.VerifyEmailResponse{Verified: false}, nil
		}
		t := time.Unix(int64(evp.Created), 0)
		if time.Since(t) > (time.Duration(PIN_LIFETIME_MINUTES) * time.Minute) {
			fmt.Printf("%s - expired\n", dout)
			return &pb.VerifyEmailResponse{Verified: false}, nil
		}

		// pin is acceptable!
		evp.Accepted = uint64(time.Now().Unix())
		evpstore.Update(ctx, evp)
		fmt.Printf("%s - verified\n", dout)
		_, err = db.ExecContext(ctx, "update_user_email_verified", "update users set emailverified = true where id = $1", uid)
		if err != nil {
			return nil, err
		}

		return &pb.VerifyEmailResponse{Verified: true}, nil
	}
	fmt.Printf("%s - no such pin\n", dout)
	return &pb.VerifyEmailResponse{Verified: false}, nil

}

// return error if password is too short, too simple or stupid
func checkPassword(ctx context.Context, password string) error {
	if len(password) < 7 {
		return errors.InvalidArgs(ctx, "password too short. min 8 chars.", "password too short. min 8 chars. received %d characters", len(password))
	}
	return nil
}

func (a *am) WhoAmI(ctx context.Context, req *common.Void) (*pb.User, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		user = auth.GetService(ctx)
		if user == nil {
			fmt.Printf("No user and no service. WhoAmI() access denied\n")
			return nil, errors.Unauthenticated(ctx, "access denied")
		}
	}
	f := &pb.ByIDRequest{UserID: user.ID}
	return a.GetUserByID(ctx, f)
}
func (a *am) Sudo(ctx context.Context, req *pb.SudoRequest) (*common.Void, error) {
	err := authBE.Sudo(ctx, req)
	if err != nil {
		return nil, err
	}
	return &common.Void{}, nil
}
func (a *am) TokenCompromised(ctx context.Context, req *pb.TokenCompromisedRequest) (*pb.NewToken, error) {
	if !auth.IsRoot(ctx) {
		return nil, errors.AccessDenied(ctx, "access root only")
	}
	// somewhat inconsistently this uses the database directly
	newtoken := utils.RandomString(64)
	_, err := dbs.ExecContext(ctx, "updatetoken", "update tokens set token = $1 where token = $2", newtoken, req.Token)
	if err != nil {
		return nil, err
	}
	return &pb.NewToken{Token: newtoken}, nil
}

func (a *am) GetGroupByID(ctx context.Context, req *pb.GetGroupRequest) (*pb.Group, error) {
	return authBE.GetGroupByID(ctx, req)
}
func (a *am) StoreRemote(ctx context.Context, req *pb.RemoteStoreRequest) (*common.Void, error) {
	err := errors.NeedServiceOrRoot(ctx, []string{"43", "4907"})
	if err != nil {
		return nil, err
	}
	return authBE.StoreRemote(ctx, req)
}
func (a *am) UserByRemoteToken(ctx context.Context, req *pb.RemoteUserRequest) (*pb.RemoteUser, error) {
	return authBE.UserByRemoteToken(ctx, req)
}
func (a *am) GetMyRemoteDetails(ctx context.Context, req *common.Void) (*pb.RemoteUser, error) {
	return authBE.GetMyRemoteDetails(ctx, req)
}
func (a *am) LogMeOut(ctx context.Context, req *common.Void) (*pb.User, error) {
	return authBE.LogMeOut(ctx, req)
}
func (a *am) LogSomeoneOut(ctx context.Context, req *pb.ByIDRequest) (*pb.User, error) {
	if !auth.IsRoot(ctx) {
		return nil, errors.AccessDenied(ctx, "access root only")
	}
	return authBE.LogSomeoneOut(ctx, req)
}
func (a *am) GetByAbbreviation(ctx context.Context, req *pb.ByAbbrevRequest) (*pb.User, error) {
	return authBE.GetByAbbreviation(ctx, req)
}
func (a *am) GetAllUsers(ctx context.Context, req *common.Void) (*pb.UserList, error) {
	return authBE.GetAllUsers(ctx, req)
}

func (a *am) CreateSession(ctx context.Context, req *common.Void) (*pb.SignedSession, error) {
	return authBE.CreateSession(ctx, req)
}
func (a *am) KeepAliveSession(ctx context.Context, req *pb.KeepAliveSessionRequest) (*pb.SignedSession, error) {
	return authBE.KeepAliveSession(ctx, req)
}
func (a *am) GetUserIDsForGroup(ctx context.Context, req *pb.GetUsersInGroupRequest) (*pb.UserIDList, error) {
	return authBE.GetUserIDsForGroup(ctx, req)
}
func GetListGroupServices() []string {
	return strings.Split(*list_group_services, ",")
}
