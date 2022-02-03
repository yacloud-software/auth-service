package authbe

import (
	"context"
	gosql "database/sql"
	"flag"
	"fmt"
	pb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/email"
	"golang.conradwood.net/authdb/db"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/prometheus"
	"golang.conradwood.net/go-easyops/sql"
	"golang.conradwood.net/go-easyops/tokens"
	"golang.conradwood.net/go-easyops/utils"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"strings"
	"time"
)

const (
	INTERNAL_USER_TOKEN_DOMAIN = "token.yacloud.eu"
	DEFAULTLIFETIMESECS        = 60 * 60 * 10
	SERVICELIFETIMESECS        = 60 * 60 * 24 * 365 * 10
	USER_TABLE_COLUMNS         = "users.id,passwd,email,firstname,lastname,abbrev,active,serviceaccount,emailverified,organisationid"
)

var (
	sudoers              *db.DBSudoStatus
	groups               *db.DBGroupDB
	es                   email.EmailServiceClient
	psql                 *sql.DB
	tokendb              *db.DBUserTokens
	orgdb                *db.DBOrganisation
	lorggroupdb          *db.DBLinkGroupOrganisation
	create_user_services = flag.String("root_services", "", "list services that are allowed to create users")
	accessCounter        = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_grpc_access",
			Help: "V=1 UNIT=hz DESC=incremented each time we process a login, by token or password",
		},
		[]string{"method", "callingserviceid", "callingservicename"},
	)
)

type PostgresAuthenticator struct {
}

func (a *PostgresAuthenticator) Start() error {
	prometheus.MustRegister(accessCounter)
	dbx, err := sql.Open()
	if err != nil {
		return err
	}
	psql = dbx
	sudoers = db.NewDBSudoStatus(psql)
	groups = db.NewDBGroupDB(psql)
	remoteuserdb = db.NewDBRemoteUserDetail(psql)
	tokendb = db.NewDBUserTokens(psql)
	orgdb = db.NewDBOrganisation(psql)
	lorggroupdb = db.NewDBLinkGroupOrganisation(psql)
	ctx := context.Background()
	fmt.Printf("Creating tables...\n")
	utils.Bail("failed to create table", sudoers.CreateTable(ctx))
	utils.Bail("failed to create table", groups.CreateTable(ctx))
	utils.Bail("failed to create table", remoteuserdb.CreateTable(ctx))
	utils.Bail("failed to create table", tokendb.CreateTable(ctx))
	utils.Bail("failed to create table", orgdb.CreateTable(ctx))
	utils.Bail("failed to create table", lorggroupdb.CreateTable(ctx))
	if !table_exists("users") {
		_, err = psql.ExecContext(ctx, "auth_create_users", `
create sequence if not exists users_serial;
CREATE TABLE users (
    id integer DEFAULT nextval('users_serial'::regclass) NOT NULL,
    passwd character varying(256) NOT NULL,
    email character varying(256) NOT NULL,
    firstname character varying(256) NOT NULL,
    lastname character varying(256) NOT NULL,
    abbrev character varying(256) NOT NULL,
    active boolean NOT NULL,
    serviceaccount boolean NOT NULL,
    emailverified boolean DEFAULT false NOT NULL,
    organisationid bigint DEFAULT 1 NOT NULL
);

ALTER TABLE ONLY users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX userid_idx ON users USING btree (id);
ALTER TABLE ONLY users ADD CONSTRAINT orgiduser FOREIGN KEY (organisationid) REFERENCES organisation(id) ON DELETE CASCADE;
`)
		if err != nil {
			fmt.Printf("Failed to create users table: %s\n", err)
		}
	}
	if !table_exists("user_groups") {
		_, err = psql.ExecContext(ctx, "auth_create_user_groups", `CREATE TABLE user_groups (
    userid integer NOT NULL,
    groupid integer NOT NULL
);
ALTER TABLE ONLY user_groups ADD CONSTRAINT foouq UNIQUE (userid, groupid);
ALTER TABLE ONLY user_groups ADD CONSTRAINT user_groups_group_id_fkey FOREIGN KEY (groupid) REFERENCES groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY user_groups ADD CONSTRAINT user_groups_user_id_fkey FOREIGN KEY (userid) REFERENCES users(id) ON DELETE CASCADE;
`)
		if err != nil {
			fmt.Printf("Failed to create user_groups table: %s\n", err)
		}
	}

	_, err = orgdb.ByID(ctx, 1)
	if err != nil {
		//create default organisation
		orgdb.SaveWithID(ctx, &pb.Organisation{ID: 1, Name: "Individual"})
	}
	go pg_expirer()
	return nil
}

func GetRootServices() []string {
	return strings.Split(*create_user_services, ",")
}
func (a *PostgresAuthenticator) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	u, err := a.CreateUserWithErr(ctx, req)
	if err != nil {
		send_notification("failed to create user \"%s\": %s", req.Email, utils.ErrorString(err))
	}
	return u, err
}
func (a *PostgresAuthenticator) CreateUserWithErr(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	e := errors.NeedServiceOrRoot(ctx, GetRootServices())
	if e != nil {
		return nil, e
	}
	if !validEmail(req.Email) {
		return nil, errors.InvalidArgs(ctx, "invalid emailaddress", "email address \"%s\" is not a valid emailaddress", req.Email)
	}
	if (req.Password != "") && (len(req.Password) < 8) {
		return nil, errors.InvalidArgs(ctx, "password too short", "password too short")
	}
	if len(req.FirstName) < 3 {
		return nil, errors.InvalidArgs(ctx, "firstname too short", "firstname too short")
	}
	if len(req.LastName) < 3 {
		return nil, errors.InvalidArgs(ctx, "lastname too short", "lastname too short")
	}
	if len(req.Abbrev) < 3 {
		return nil, errors.InvalidArgs(ctx, "abbrev too short", "abbrev too short")
	}
	if len(req.Abbrev) > 7 {
		return nil, errors.InvalidArgs(ctx, "abbrev too long", "abbrev too long")
	}
	res := &pb.User{Password: req.Password,
		Email:          req.Email,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Abbrev:         req.Abbrev,
		Active:         true,
		ServiceAccount: false,
		EmailVerified:  req.EmailVerified,
		OrganisationID: "1",
	}
	err := a.createUser(ctx, res)
	if err != nil {
		return nil, err
	}
	sign(res)
	return res, nil
}

func (a *PostgresAuthenticator) CreateService(ctx context.Context, req *pb.CreateServiceRequest) (*pb.NewService, error) {
	allow := false
	if auth.IsRoot(ctx) {
		allow = true
	} else {
		us := auth.GetService(ctx)
		if us != nil && us.ID == "3539" { // repobuilder
			allow = true
		}
	}

	if !allow {
		return nil, errors.AccessDenied(ctx, "create service requires special privileges")
	}
	if len(req.ServiceName) < 3 {
		return nil, fmt.Errorf("Servicename too short or missing")
	}
	res := &pb.NewService{
		User: &pb.User{Password: utils.RandomString(64),
			Email:          fmt.Sprintf("%s%s", req.ServiceName, *default_domain),
			FirstName:      req.ServiceName,
			LastName:       req.ServiceName,
			Abbrev:         req.ServiceName,
			Active:         true,
			ServiceAccount: true,
		},
	}
	err := a.createUser(ctx, res.User)
	if err != nil {
		return nil, err
	}

	// set parameters for token
	lifetime := SERVICELIFETIMESECS
	newtok := req.Token
	userid, err := strconv.ParseUint(res.User.ID, 10, 64)
	if err != nil {
		return nil, err
	}

	// create a token
	if newtok == "" {
		newtok = utils.RandomString(64)
	}
	res.Token = newtok
	now := time.Now().Unix()
	pdb := &pb.UserTokens{UserID: userid, Token: newtok, Created: uint32(now), Expiry: uint32(now + int64(lifetime)), TokenType: pb.TokenType_PERMANENT}
	_, err = tokendb.Save(ctx, pdb)
	if err != nil {
		return nil, err
	}
	sign(res.User)
	return res, nil
}

func (d *PostgresAuthenticator) GetTokenForMe(ctx context.Context, req *pb.GetTokenRequest) (*pb.TokenResponse, error) {
	user := auth.GetUser(ctx)
	if user == nil {
		return nil, errors.Unauthenticated(ctx, "gettokenforme() requires a user in context")
	}
	now := uint32(time.Now().Unix())
	userid, err := strconv.ParseUint(user.ID, 10, 64)
	if err != nil {
		return nil, err
	}
	var exp time.Time
	tktype := pb.TokenType_SESSION
	if req.DurationSecs == 0 {
		tktype = pb.TokenType_PERMANENT
		exp = time.Now().AddDate(10, 0, 0) // 10 years...
	} else {
		exp = time.Now().Add(time.Duration(req.DurationSecs) * time.Second)
	}
	tok := utils.RandomString(64)
	pdb := &pb.UserTokens{UserID: userid, Token: tok, Created: uint32(now), Expiry: uint32(exp.Unix()), TokenType: tktype}

	_, err = tokendb.Save(ctx, pdb)
	if err != nil {
		return nil, err
	}
	res := &pb.TokenResponse{Token: tok, Expiry: uint32(exp.Unix())}
	return res, nil
}

func (a *PostgresAuthenticator) GetUserByID(ctx context.Context, req *pb.ByIDRequest) (*pb.User, error) {
	if !*nonsignedusers {
		return nil, errors.NotImplemented(ctx, "nonsigned users no longer supported")
	}
	if len(req.UserID) < 1 {
		return nil, fmt.Errorf("Missing id")
	}
	userid, err := strconv.Atoi(req.UserID)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, "get_by_id", "select "+USER_TABLE_COLUMNS+" from users where id = $1", userid)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		rows.Close()
		return nil, errors.InvalidArgs(ctx, "user not found", "user \"%s\" not found in database", req.UserID)
	}
	u, err := userFromRow(rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	u.Password = ""
	err = a.SetGroups(ctx, u)
	if err != nil {
		return nil, err
	}
	sign(u)
	return u, nil

}
func (a *PostgresAuthenticator) GetUserByEmail(ctx context.Context, req *pb.ByEmailRequest) (*pb.User, error) {
	if !*nonsignedusers {
		return nil, errors.NotImplemented(ctx, "nonsigned users no longer supported")
	}

	if !validEmail(req.Email) {
		return nil, errors.InvalidArgs(ctx, "invalid emailaddress", "email address \"%s\" is not a valid emailaddress", req.Email)
	}

	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, "get_by_email", "select "+USER_TABLE_COLUMNS+" from users where email = $1", req.Email)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		rows.Close()
		return nil, errors.InvalidArgs(ctx, "user not found", "user \"%s\" not found in database", req.Email)
	}
	u, err := userFromRow(rows)
	if err != nil {
		return nil, err
	}
	rows.Close()
	u.Password = ""
	err = a.SetGroups(ctx, u)
	if err != nil {
		return nil, err
	}
	sign(u)
	return u, nil

}
func getlabels(ctx context.Context, method string) prometheus.Labels {
	svc := auth.GetService(ctx)
	if svc == nil {
		return prometheus.Labels{"method": method, "callingserviceid": "0", "callingservicename": "[undef]"}
	}
	return prometheus.Labels{"method": method,
		"callingserviceid":   svc.ID,
		"callingservicename": svc.Abbrev,
	}

}

// authenticate a user by username/password, return token
func (a *PostgresAuthenticator) GetByPassword(ctx context.Context, req *pb.AuthenticatePasswordRequest) (*pb.AuthResponse, error) {
	accessCounter.With(getlabels(ctx, "getbypassword")).Inc()
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	// it might be a username in the form of [userid]@token.yacloud.eu with the password being the token
	if strings.HasSuffix(req.Email, "@"+INTERNAL_USER_TOKEN_DOMAIN) {
		st := strings.TrimSuffix(req.Email, "@"+INTERNAL_USER_TOKEN_DOMAIN)
		ar, err := a.GetByToken(ctx, &pb.AuthenticateTokenRequest{Token: req.Password})
		if err != nil {
			return nil, err
		}
		if ar.User != nil && ar.User.ID == st {
			return ar, nil
		}
		return &pb.AuthResponse{Valid: false,
			PublicMessage: "access denied",
			LogMessage:    fmt.Sprintf("tokenrequest for email \"%s\" denied", req.Email),
		}, nil
	}
	rows, err := db.QueryContext(ctx, "get_by_password", "select "+USER_TABLE_COLUMNS+" from users where email = $1", req.Email)
	if err != nil {
		return nil, err
	}
	res := &pb.AuthResponse{Valid: false, PublicMessage: "access denied"}
	if !rows.Next() {
		res.LogMessage = "no token found for request"
		rows.Close()
		return res, nil
	}
	u, err := userFromRow(rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password))
	if err != nil {
		res.LogMessage = fmt.Sprintf("password mismatch (%s) (len=%d)", err, len(req.Password))
		return res, nil
	}
	u.Password = ""
	tok, err := needToken(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	res.User = u
	res.Token = tok
	res.Valid = true
	res.PublicMessage = "OK"
	res.LogMessage = "authenticated " + res.User.Email
	err = a.SetGroups(ctx, res.User)
	if err != nil {
		return nil, err
	}
	sign(res.User)
	return res, nil
}

// authenticate a user by token, return same token
func (a *PostgresAuthenticator) GetByToken(ctx context.Context, req *pb.AuthenticateTokenRequest) (*pb.AuthResponse, error) {
	accessCounter.With(getlabels(ctx, "getbytoken")).Inc()
	if len(req.Token) < 12 {
		res := &pb.AuthResponse{Valid: false, PublicMessage: "access denied", LogMessage: fmt.Sprintf("Token too short (%d) characters", len(req.Token))}
		return res, nil
	}
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	if *debug {
		fmt.Printf("Attempting to get user by token \"%s\"\n", req.Token)
	}
	now := time.Now().Unix()
	rows, err := db.QueryContext(ctx, "get_by_token", "select "+USER_TABLE_COLUMNS+" from users,tokens where tokens.userid = users.id and tokens.token = $1 and tokens.expiry > $2", req.Token, now)
	if err != nil {
		return nil, err
	}
	res := &pb.AuthResponse{Valid: false, PublicMessage: "access denied"}
	if !rows.Next() {
		if *debug || *print_failures {
			fmt.Printf("Token \"%s\" not found in db\n", req.Token)
		}
		res.LogMessage = "no token found for request"
		rows.Close()
		return res, nil
	}
	if *debug {
		fmt.Printf("Token \"%s\" was found in db\n", req.Token)
	}
	u, err := userFromRow(rows)
	rows.Close()
	if err != nil {
		if *debug || *print_failures {
			fmt.Printf("Token \"%s\" was found in db, but cannot get user (%s)\n", req.Token, err)
		}
		return nil, err
	}

	u.Password = ""
	res.User = u
	res.Token = req.Token
	res.Valid = true
	res.PublicMessage = "OK"
	res.LogMessage = "authenticated " + res.User.Email

	err = a.SetGroups(ctx, res.User)
	if err != nil {
		if *debug || *print_failures {
			fmt.Printf("Token \"%s\" was found in db, but no groups (%s)\n", req.Token, err)
		}
		return nil, err
	}
	if *debug {
		fmt.Printf("Token \"%s\" was resolved to user (%s [%s])\n", req.Token, res.User.Email, res.User.ID)
	}
	sign(res.User)
	return res, nil

}

func userFromRow(rows *gosql.Rows) (*pb.User, error) {
	res := &pb.User{}
	orgid := gosql.NullInt64{}
	err := rows.Scan(&res.ID, &res.Password, &res.Email, &res.FirstName, &res.LastName, &res.Abbrev, &res.Active, &res.ServiceAccount, &res.EmailVerified, &orgid)
	if err != nil {
		return nil, err
	}
	res.OrganisationID = fmt.Sprintf("%d", orgid.Int64)
	return res, nil
}

// get or create SESSION token for user
func needToken(ctx context.Context, userid string) (string, error) {
	db, err := sql.Open()
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()
	minexp := now + 60*10 // min remaining lifetime: 10 minutes
	rows, err := db.QueryContext(ctx, "get_tokens_for_user", "select token from tokens where userid = $1 and expiry > $2 and tokentype = 2 order by expiry desc limit 1", userid, minexp)
	if err != nil {
		return "", err
	}
	token := ""
	for rows.Next() {
		err = rows.Scan(&token)
		if err != nil {
			rows.Close()
			return "", err
		}
	}
	rows.Close()
	if len(token) > 8 {
		return token, nil
	}
	// no token found
	tok := utils.RandomString(64)
	uid, err := strconv.ParseUint(userid, 10, 64)
	if err != nil {
		return "", err
	}
	pdb := &pb.UserTokens{UserID: uid, Token: tok, Created: uint32(now), Expiry: uint32(now + DEFAULTLIFETIMESECS), TokenType: pb.TokenType_SESSION}
	_, err = tokendb.Save(ctx, pdb)
	//	_, err = db.ExecContext(ctx, "create_token_for_user", "insert into tokens (userid,token,created,expiry) values ($1,$2,$3,$4)", userid, tok, now, now+DEFAULTLIFETIMESECS)
	if err != nil {
		return "", err
	}
	return tok, nil
}
func (a *PostgresAuthenticator) SetGroups(ctx context.Context, user *pb.User) error {
	if user == nil {
		return nil
	}
	db, err := sql.Open()
	if err != nil {
		return err
	}

	rows, err := db.QueryContext(ctx, "get_groups_for_user", "select id,name,description from groups,user_groups where user_groups.userid = $1 and user_groups.groupid = groups.id", user.ID)
	if err != nil {
		return err
	}
	var gr []*pb.Group
	for rows.Next() {
		g := &pb.Group{}
		err = rows.Scan(&g.ID, &g.Name, &g.Description)
		if err != nil {
			rows.Close()
			return err
		}
		gr = append(gr, g)
	}
	rows.Close()
	for _, g := range gr {
		ex := false
		for _, ug := range user.Groups {
			if ug.ID == g.ID {
				ex = true
				break
			}
		}
		if !ex {
			user.Groups = append(user.Groups, g)
		}
	}
	AddStandardGroups(user)

	uid, err := strconv.ParseUint(user.ID, 10, 64)
	if err != nil {
		return err
	}
	// add the sudoers groups:
	sts, err := sudoers.ByUserID(ctx, uid)
	if err != nil {
		return err
	}
	now := uint32(time.Now().Unix())
	for _, s := range sts {
		if s.Expiry <= now {
			sudoers.DeleteByID(ctx, s.ID)
			continue
		}
		exists := false
		for _, g := range user.Groups {
			if g.ID == s.GroupID {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		xgid, err := strconv.ParseUint(s.GroupID, 10, 64)
		if err != nil {
			return err
		}
		gs, err := groups.ByID(ctx, xgid)
		if err != nil {
			return err
		}
		user.Groups = append(user.Groups, &pb.Group{ID: fmt.Sprintf("%d", gs.ID), Name: gs.Name, Description: gs.Description})
	}

	return nil
}
func validEmail(email string) bool {
	if len(email) < 3 {
		return false
	}
	if !strings.Contains(email, "@") {
		return false
	}
	return true
}

// return default of '1' (individuum) unless otherwise specificed
func getorgidforcreate(ctx context.Context, user *pb.User) (int, error) {
	if user.OrganisationID == "" {
		// no organisation specified? create in same org as calling user
		cuser := auth.GetUser(ctx)
		if cuser == nil {
			fmt.Printf("No organisationid specified, and no user in context either")
			return 1, nil
		}
		if cuser.OrganisationID == "" {
			fmt.Printf("No organisationid specified, and calling user in context has no organisationid either")
			return 1, nil
		}
		user.OrganisationID = cuser.OrganisationID
	}
	orgid, err := strconv.Atoi(user.OrganisationID)
	if err != nil {
		return 0, fmt.Errorf("(createuser) invalid organisation id (%s)", err)
	}
	return orgid, nil

}
func (a *PostgresAuthenticator) createUser(ctx context.Context, user *pb.User) error {
	send_notification("Creating user \"%s\"", user.Email)
	db, err := sql.Open()
	if err != nil {
		return err
	}
	if user.Password == "" {
		user.Password = utils.RandomString(12)
	}
	bc, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		fmt.Printf("Failed to encode password: %s\n", err)
		return err
	}
	orgid, err := getorgidforcreate(ctx, user)
	if err != nil {
		return fmt.Errorf("(createuser) invalid organisation id (%s)", err)
	}
	pw := string(bc)

	// new codepath:
	if user.Email == "prober_user_foo_bar@conradwood.net" {
	}

	// old codepath:
	rid, err := db.QueryContext(ctx, "create_user", "insert into users (passwd,email,firstname,lastname,abbrev,active,serviceaccount,emailverified,organisationid) values ($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id", pw, user.Email, user.FirstName, user.LastName, user.Abbrev, user.Active, user.ServiceAccount, user.EmailVerified, orgid)
	if err != nil {
		return err
	}
	if !rid.Next() {
		rid.Close()
		return fmt.Errorf("user created, but no ID!")
	}
	err = rid.Scan(&user.ID)
	rid.Close()
	if err != nil {
		return fmt.Errorf("error scanning newly created user id: %s\n", err)
	}
	err = a.SetGroups(ctx, user)
	if err != nil {
		return err
	}
	sign(user)
	return nil
}
func (a *PostgresAuthenticator) ResetPasswordEmail(ctx context.Context, req *pb.ResetRequest) (*common.Void, error) {
	u, err := a.GetUserByEmail(ctx, &pb.ByEmailRequest{Email: req.Email})
	if err != nil {
		return nil, err
	}
	if u == nil {
		fmt.Printf("Silently supressing request to reset password for %s (user does not exist)", req.Email)
		return &common.Void{}, nil
	}
	if es == nil {
		es = email.GetEmailServiceClient()
	}
	tok, err := needToken(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	ter := &email.TemplateEmailRequest{Sender: "donotreply@yacloud.eu",
		Recipient:    req.Email,
		TemplateName: "forgotemail",
		Values:       make(map[string]string),
	}
	ter.Values["token"] = tok
	ter.Values["firstname"] = u.FirstName
	ter.Values["lastname"] = u.LastName
	ter.Values["email"] = u.Email
	// send asynchronously
	go func(vals *email.TemplateEmailRequest) {
		ctx := tokens.ContextWithToken()
		tr, err := es.SendTemplate(ctx, vals)
		if err != nil {
			fmt.Printf("Send Email to %s failed: %s\n", vals.Values["email"], err)
			return
		}
		if !tr.Success {
			fmt.Printf("Send Email unsuccessfull to %s: %s\n", vals.Values["email"], err)
			return
		}
	}(ter)
	return &common.Void{}, nil
}
func (a *PostgresAuthenticator) ListAllGroups(ctx context.Context) (*pb.GroupList, error) {
	res := &pb.GroupList{}
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, "get_all_groups", "select id,name,description from groups")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		g := &pb.Group{}
		err = rows.Scan(&g.ID, &g.Name, &g.Description)
		if err != nil {
			rows.Close()
			return nil, err
		}
		res.Groups = append(res.Groups, g)
	}
	rows.Close()
	return res, nil
}
func (a *PostgresAuthenticator) ExpireToken(ctx context.Context, req *pb.ExpireTokenRequest) (*common.Void, error) {
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	ar, err := a.GetByToken(ctx, &pb.AuthenticateTokenRequest{Token: req.Token})
	if err != nil {
		return nil, err
	}
	if !ar.Valid {
		return nil, errors.AccessDenied(ctx, ar.LogMessage)
	}
	if ar.User == nil {
		return nil, errors.Unauthenticated(ctx, "token did not map to a user")
	}
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "no user provided")
	}

	if u.ID != ar.User.ID {
		return nil, errors.AccessDenied(ctx, "expire token: token for user %s called by user %s", ar.User.ID, u.ID)
	}

	_, err = db.ExecContext(ctx, "delete_token", "delete from tokens where userid = $1 and token = $2", u.ID, req.Token)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (d *PostgresAuthenticator) GetPublicSigningKey(ctx context.Context, req *common.Void) (*pb.KeyResponse, error) {
	return &pb.KeyResponse{CloudName: *cloudname, Key: signPublicKey()}, nil
}
func (d *PostgresAuthenticator) Sudo(ctx context.Context, req *pb.SudoRequest) error {
	u := auth.GetUser(ctx)
	if u == nil {
		return fmt.Errorf("no current user")
	}
	uids := u.ID
	if req.UserID != "" {
		ok := false
		if u.ID == "1" && req.UserID == "7" {
			ok = true
		}
		if u.ID == "7" && req.UserID == "1" {
			ok = true
		}
		if u.ID == req.UserID {
			ok = true
		}
		if !ok {
			return errors.AccessDenied(ctx, "invalid userid combination")
		}
		uids = req.UserID

	}
	uid, err := strconv.ParseUint(uids, 10, 64)
	if err != nil {
		return err
	}
	if uid != 1 && uid != 7 {
		return fmt.Errorf("access denied")
	}
	e := uint32(time.Now().Add(time.Duration(30) * time.Minute).Unix())
	sts := &pb.SudoStatus{UserID: uid, GroupID: "1", Expiry: e}
	_, err = sudoers.Save(ctx, sts)
	return err

}

func (a *PostgresAuthenticator) GetGroupByID(ctx context.Context, req *pb.GetGroupRequest) (*pb.Group, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "missing useraccount")
	}
	a.SetGroups(ctx, u)
	for _, g := range u.Groups {
		if g.ID == req.ID {
			return g, nil
		}
	}
	return nil, errors.NotFound(ctx, "group %s not found for user %s(%s)", req.ID, u.ID, u.Email)
}
func (a *PostgresAuthenticator) LogSomeoneOut(ctx context.Context, req *pb.ByIDRequest) (*pb.User, error) {
	return a.logOut(ctx, req.UserID)
}
func (a *PostgresAuthenticator) LogMeOut(ctx context.Context, req *common.Void) (*pb.User, error) {
	u := auth.GetUser(ctx)
	if u == nil {
		return nil, errors.Unauthenticated(ctx, "need user to log out")
	}
	return a.logOut(ctx, u.ID)
}
func (a *PostgresAuthenticator) logOut(ctx context.Context, userid string) (*pb.User, error) {
	uid, err := strconv.ParseUint(userid, 10, 64)
	if err != nil {
		return nil, err
	}
	tks, err := tokendb.ByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	for _, tk := range tks {
		if tk.TokenType != pb.TokenType_SESSION {
			continue
		}
		err = tokendb.DeleteByID(ctx, tk.ID)
		if err != nil {
			return nil, err
		}
	}
	f, err := a.GetUserByID(ctx, &pb.ByIDRequest{UserID: userid})
	return f, err
}

/* the "signed stuff */
func (a *PostgresAuthenticator) SignedGetUserByEmail(ctx context.Context, req *pb.ByEmailRequest) (*pb.SignedUser, error) {
	u, err := a.GetUserByEmail(ctx, req)
	if err != nil {
		return nil, err
	}
	return userToSignedUser(u)
}
func (a *PostgresAuthenticator) SignedGetUserByID(ctx context.Context, req *pb.ByIDRequest) (*pb.SignedUser, error) {
	u, err := a.GetUserByID(ctx, req)
	if err != nil {
		return nil, err
	}
	return userToSignedUser(u)
}
func (a *PostgresAuthenticator) SignedGetByPassword(ctx context.Context, req *pb.AuthenticatePasswordRequest) (*pb.SignedAuthResponse, error) {
	r, err := a.GetByPassword(ctx, req)
	if err != nil {
		return nil, err
	}
	return ResponseToSignedResponse(r)
}
func (a *PostgresAuthenticator) SignedGetByToken(ctx context.Context, req *pb.AuthenticateTokenRequest) (*pb.SignedAuthResponse, error) {
	r, err := a.GetByToken(ctx, req)
	if err != nil {
		return nil, err
	}
	return ResponseToSignedResponse(r)

}
func (a *PostgresAuthenticator) GetByAbbreviation(ctx context.Context, req *pb.ByAbbrevRequest) (*pb.User, error) {
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, "get_by_abbrev", "select "+USER_TABLE_COLUMNS+" from users where abbrev = $1 order by id asc limit 1", req.Abbrev)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		rows.Close()
		return nil, errors.InvalidArgs(ctx, "user not found", "user \"%s\" not found in database", req.Abbrev)
	}
	u, err := userFromRow(rows)
	if err != nil {
		return nil, err
	}
	rows.Close()
	u.Password = ""
	err = a.SetGroups(ctx, u)
	if err != nil {
		return nil, err
	}
	sign(u)
	return u, nil

}
func (a *PostgresAuthenticator) GetAllUsers(ctx context.Context, req *common.Void) (*pb.UserList, error) {
	err := errors.NeedsRoot(ctx)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, "get_all", "select "+USER_TABLE_COLUMNS+" from users where serviceaccount = false order by id")
	if err != nil {
		return nil, err
	}
	var users []*pb.User
	for rows.Next() {
		u, err := userFromRow(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		u.Password = ""
		users = append(users, u)
		err = a.SetGroups(ctx, u)
		if err != nil {
			rows.Close()
			return nil, err
		}
	}
	rows.Close()
	ul := &pb.UserList{}
	for _, u := range users {
		su, err := userToSignedUser(u)
		if err != nil {
			return nil, err
		}
		ul.Users = append(ul.Users, su)
	}
	return ul, nil
}

func table_exists(tname string) bool {
	_, err := psql.ExecContext(context.Background(), "check_exists", "select * from "+tname)
	if err == nil {
		fmt.Printf("Table %s exists already\n", tname)
		return true
	}
	fmt.Printf("Table %s does not exist yet\n", tname)
	return false
}
