package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBUser
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence users_seq;

Main Table:

 CREATE TABLE users (id integer primary key default nextval('users_seq'),email text not null  ,firstname text not null  ,lastname text not null  ,passwd text not null  ,abbrev text not null  ,active boolean not null  ,serviceaccount boolean not null  ,emailverified boolean not null  ,organisationid text not null  ,created integer not null  );

Alter statements:
ALTER TABLE users ADD COLUMN email text not null default '';
ALTER TABLE users ADD COLUMN firstname text not null default '';
ALTER TABLE users ADD COLUMN lastname text not null default '';
ALTER TABLE users ADD COLUMN passwd text not null default '';
ALTER TABLE users ADD COLUMN abbrev text not null default '';
ALTER TABLE users ADD COLUMN active boolean not null default false;
ALTER TABLE users ADD COLUMN serviceaccount boolean not null default false;
ALTER TABLE users ADD COLUMN emailverified boolean not null default false;
ALTER TABLE users ADD COLUMN organisationid text not null default '';
ALTER TABLE users ADD COLUMN created integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE users_archive (id integer unique not null,email text not null,firstname text not null,lastname text not null,passwd text not null,abbrev text not null,active boolean not null,serviceaccount boolean not null,emailverified boolean not null,organisationid text not null,created integer not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBUser *DBUser
)

type DBUser struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBUser() *DBUser {
	if default_def_DBUser != nil {
		return default_def_DBUser
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBUser(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBUser = res
	return res
}
func NewDBUser(db *sql.DB) *DBUser {
	foo := DBUser{DB: db}
	foo.SQLTablename = "users"
	foo.SQLArchivetablename = "users_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBUser) Archive(ctx context.Context, id string) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBUser", "insert into "+a.SQLArchivetablename+"+ (id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ", p.ID, p.Email, p.FirstName, p.LastName, p.Password, p.Abbrev, p.Active, p.ServiceAccount, p.EmailVerified, p.OrganisationID, p.Created)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBUser) Save(ctx context.Context, p *savepb.User) (string, error) {
	qn := "DBUser_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id", p.Email, p.FirstName, p.LastName, p.Password, p.Abbrev, p.Active, p.ServiceAccount, p.EmailVerified, p.OrganisationID, p.Created)
	if e != nil {
		return "", a.Error(ctx, qn, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return "", a.Error(ctx, qn, fmt.Errorf("No rows after insert"))
	}
	var id string
	e = rows.Scan(&id)
	if e != nil {
		return "", a.Error(ctx, qn, fmt.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// Save using the ID specified
func (a *DBUser) SaveWithID(ctx context.Context, p *savepb.User) error {
	qn := "insert_DBUser"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11) ", p.ID, p.Email, p.FirstName, p.LastName, p.Password, p.Abbrev, p.Active, p.ServiceAccount, p.EmailVerified, p.OrganisationID, p.Created)
	return a.Error(ctx, qn, e)
}

func (a *DBUser) Update(ctx context.Context, p *savepb.User) error {
	qn := "DBUser_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set email=$1, firstname=$2, lastname=$3, passwd=$4, abbrev=$5, active=$6, serviceaccount=$7, emailverified=$8, organisationid=$9, created=$10 where id = $11", p.Email, p.FirstName, p.LastName, p.Password, p.Abbrev, p.Active, p.ServiceAccount, p.EmailVerified, p.OrganisationID, p.Created, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBUser) DeleteByID(ctx context.Context, p string) error {
	qn := "deleteDBUser_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBUser) ByID(ctx context.Context, p string) (*savepb.User, error) {
	qn := "DBUser_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No User with id %d", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) User with id %d", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBUser) All(ctx context.Context) ([]*savepb.User, error) {
	qn := "DBUser_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" order by id")
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("All: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, fmt.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBUser" rows with matching Email
func (a *DBUser) ByEmail(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByEmail"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where email = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEmail: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEmail: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeEmail(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeEmail"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where email ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEmail: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEmail: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching FirstName
func (a *DBUser) ByFirstName(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByFirstName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where firstname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFirstName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFirstName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeFirstName(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeFirstName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where firstname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFirstName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFirstName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching LastName
func (a *DBUser) ByLastName(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByLastName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where lastname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeLastName(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeLastName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where lastname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching Password
func (a *DBUser) ByPassword(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByPassword"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where passwd = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPassword: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPassword: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikePassword(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByLikePassword"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where passwd ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPassword: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPassword: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching Abbrev
func (a *DBUser) ByAbbrev(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByAbbrev"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where abbrev = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAbbrev: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAbbrev: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeAbbrev(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeAbbrev"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where abbrev ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAbbrev: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAbbrev: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching Active
func (a *DBUser) ByActive(ctx context.Context, p bool) ([]*savepb.User, error) {
	qn := "DBUser_ByActive"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where active = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByActive: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByActive: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeActive(ctx context.Context, p bool) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeActive"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where active ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByActive: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByActive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching ServiceAccount
func (a *DBUser) ByServiceAccount(ctx context.Context, p bool) ([]*savepb.User, error) {
	qn := "DBUser_ByServiceAccount"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where serviceaccount = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceAccount: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceAccount: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeServiceAccount(ctx context.Context, p bool) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeServiceAccount"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where serviceaccount ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceAccount: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByServiceAccount: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching EmailVerified
func (a *DBUser) ByEmailVerified(ctx context.Context, p bool) ([]*savepb.User, error) {
	qn := "DBUser_ByEmailVerified"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where emailverified = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEmailVerified: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEmailVerified: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeEmailVerified(ctx context.Context, p bool) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeEmailVerified"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where emailverified ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEmailVerified: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByEmailVerified: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching OrganisationID
func (a *DBUser) ByOrganisationID(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByOrganisationID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where organisationid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOrganisationID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOrganisationID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeOrganisationID(ctx context.Context, p string) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeOrganisationID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where organisationid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOrganisationID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOrganisationID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUser" rows with matching Created
func (a *DBUser) ByCreated(ctx context.Context, p uint32) ([]*savepb.User, error) {
	qn := "DBUser_ByCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where created = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUser) ByLikeCreated(ctx context.Context, p uint32) ([]*savepb.User, error) {
	qn := "DBUser_ByLikeCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created from "+a.SQLTablename+" where created ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBUser) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.User, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBUser) Tablename() string {
	return a.SQLTablename
}

func (a *DBUser) SelectCols() string {
	return "id,email, firstname, lastname, passwd, abbrev, active, serviceaccount, emailverified, organisationid, created"
}
func (a *DBUser) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".email, " + a.SQLTablename + ".firstname, " + a.SQLTablename + ".lastname, " + a.SQLTablename + ".passwd, " + a.SQLTablename + ".abbrev, " + a.SQLTablename + ".active, " + a.SQLTablename + ".serviceaccount, " + a.SQLTablename + ".emailverified, " + a.SQLTablename + ".organisationid, " + a.SQLTablename + ".created"
}

func (a *DBUser) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.User, error) {
	var res []*savepb.User
	for rows.Next() {
		foo := savepb.User{}
		err := rows.Scan(&foo.ID, &foo.Email, &foo.FirstName, &foo.LastName, &foo.Password, &foo.Abbrev, &foo.Active, &foo.ServiceAccount, &foo.EmailVerified, &foo.OrganisationID, &foo.Created)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBUser) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),email text not null  ,firstname text not null  ,lastname text not null  ,passwd text not null  ,abbrev text not null  ,active boolean not null  ,serviceaccount boolean not null  ,emailverified boolean not null  ,organisationid text not null  ,created integer not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),email text not null  ,firstname text not null  ,lastname text not null  ,passwd text not null  ,abbrev text not null  ,active boolean not null  ,serviceaccount boolean not null  ,emailverified boolean not null  ,organisationid text not null  ,created integer not null  );`,
	}
	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBUser) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
