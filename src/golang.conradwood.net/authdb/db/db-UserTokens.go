package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBUserTokens
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence tokens_seq;

Main Table:

 CREATE TABLE tokens (id integer primary key default nextval('tokens_seq'),userid bigint not null  ,token text not null  ,created integer not null  ,expiry integer not null  ,tokentype integer not null  );

Alter statements:
ALTER TABLE tokens ADD COLUMN userid bigint not null default 0;
ALTER TABLE tokens ADD COLUMN token text not null default '';
ALTER TABLE tokens ADD COLUMN created integer not null default 0;
ALTER TABLE tokens ADD COLUMN expiry integer not null default 0;
ALTER TABLE tokens ADD COLUMN tokentype integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE tokens_archive (id integer unique not null,userid bigint not null,token text not null,created integer not null,expiry integer not null,tokentype integer not null);
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
	default_def_DBUserTokens *DBUserTokens
)

type DBUserTokens struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBUserTokens() *DBUserTokens {
	if default_def_DBUserTokens != nil {
		return default_def_DBUserTokens
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBUserTokens(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBUserTokens = res
	return res
}
func NewDBUserTokens(db *sql.DB) *DBUserTokens {
	foo := DBUserTokens{DB: db}
	foo.SQLTablename = "tokens"
	foo.SQLArchivetablename = "tokens_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBUserTokens) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBUserTokens", "insert into "+a.SQLArchivetablename+"+ (id,userid, token, created, expiry, tokentype) values ($1,$2, $3, $4, $5, $6) ", p.ID, p.UserID, p.Token, p.Created, p.Expiry, p.TokenType)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBUserTokens) Save(ctx context.Context, p *savepb.UserTokens) (uint64, error) {
	qn := "DBUserTokens_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (userid, token, created, expiry, tokentype) values ($1, $2, $3, $4, $5) returning id", p.UserID, p.Token, p.Created, p.Expiry, p.TokenType)
	if e != nil {
		return 0, a.Error(ctx, qn, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, qn, fmt.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, qn, fmt.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// Save using the ID specified
func (a *DBUserTokens) SaveWithID(ctx context.Context, p *savepb.UserTokens) error {
	qn := "insert_DBUserTokens"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,userid, token, created, expiry, tokentype) values ($1,$2, $3, $4, $5, $6) ", p.ID, p.UserID, p.Token, p.Created, p.Expiry, p.TokenType)
	return a.Error(ctx, qn, e)
}

func (a *DBUserTokens) Update(ctx context.Context, p *savepb.UserTokens) error {
	qn := "DBUserTokens_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set userid=$1, token=$2, created=$3, expiry=$4, tokentype=$5 where id = $6", p.UserID, p.Token, p.Created, p.Expiry, p.TokenType, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBUserTokens) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBUserTokens_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBUserTokens) ByID(ctx context.Context, p uint64) (*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No UserTokens with id %d", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) UserTokens with id %d", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBUserTokens) All(ctx context.Context) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" order by id")
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

// get all "DBUserTokens" rows with matching UserID
func (a *DBUserTokens) ByUserID(ctx context.Context, p uint64) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserTokens) ByLikeUserID(ctx context.Context, p uint64) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserTokens" rows with matching Token
func (a *DBUserTokens) ByToken(ctx context.Context, p string) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where token = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserTokens) ByLikeToken(ctx context.Context, p string) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByLikeToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where token ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserTokens" rows with matching Created
func (a *DBUserTokens) ByCreated(ctx context.Context, p uint32) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where created = $1", p)
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
func (a *DBUserTokens) ByLikeCreated(ctx context.Context, p uint32) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByLikeCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where created ilike $1", p)
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

// get all "DBUserTokens" rows with matching Expiry
func (a *DBUserTokens) ByExpiry(ctx context.Context, p uint32) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByExpiry"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where expiry = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByExpiry: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserTokens) ByLikeExpiry(ctx context.Context, p uint32) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByLikeExpiry"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where expiry ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByExpiry: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserTokens" rows with matching TokenType
func (a *DBUserTokens) ByTokenType(ctx context.Context, p uint32) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByTokenType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where tokentype = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTokenType: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTokenType: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserTokens) ByLikeTokenType(ctx context.Context, p uint32) ([]*savepb.UserTokens, error) {
	qn := "DBUserTokens_ByLikeTokenType"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, token, created, expiry, tokentype from "+a.SQLTablename+" where tokentype ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTokenType: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTokenType: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBUserTokens) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.UserTokens, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBUserTokens) Tablename() string {
	return a.SQLTablename
}

func (a *DBUserTokens) SelectCols() string {
	return "id,userid, token, created, expiry, tokentype"
}
func (a *DBUserTokens) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userid, " + a.SQLTablename + ".token, " + a.SQLTablename + ".created, " + a.SQLTablename + ".expiry, " + a.SQLTablename + ".tokentype"
}

func (a *DBUserTokens) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.UserTokens, error) {
	var res []*savepb.UserTokens
	for rows.Next() {
		foo := savepb.UserTokens{}
		err := rows.Scan(&foo.ID, &foo.UserID, &foo.Token, &foo.Created, &foo.Expiry, &foo.TokenType)
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
func (a *DBUserTokens) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid bigint not null  ,token text not null  ,created integer not null  ,expiry integer not null  ,tokentype integer not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid bigint not null  ,token text not null  ,created integer not null  ,expiry integer not null  ,tokentype integer not null  );`,
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
func (a *DBUserTokens) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
