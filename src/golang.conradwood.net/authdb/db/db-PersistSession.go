package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBPersistSession
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence users_seq;

Main Table:

 CREATE TABLE users (id integer primary key default nextval('users_seq'),token text not null  unique  ,userid text not null  ,created integer not null  );

Alter statements:
ALTER TABLE users ADD COLUMN token text not null unique  default '';
ALTER TABLE users ADD COLUMN userid text not null default '';
ALTER TABLE users ADD COLUMN created integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE users_archive (id integer unique not null,token text not null,userid text not null,created integer not null);
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
	default_def_DBPersistSession *DBPersistSession
)

type DBPersistSession struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBPersistSession() *DBPersistSession {
	if default_def_DBPersistSession != nil {
		return default_def_DBPersistSession
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBPersistSession(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBPersistSession = res
	return res
}
func NewDBPersistSession(db *sql.DB) *DBPersistSession {
	foo := DBPersistSession{DB: db}
	foo.SQLTablename = "users"
	foo.SQLArchivetablename = "users_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBPersistSession) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBPersistSession", "insert into "+a.SQLArchivetablename+"+ (id,token, userid, created) values ($1,$2, $3, $4) ", p.ID, p.Token, p.UserID, p.Created)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBPersistSession) Save(ctx context.Context, p *savepb.PersistSession) (uint64, error) {
	qn := "DBPersistSession_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (token, userid, created) values ($1, $2, $3) returning id", p.Token, p.UserID, p.Created)
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
func (a *DBPersistSession) SaveWithID(ctx context.Context, p *savepb.PersistSession) error {
	qn := "insert_DBPersistSession"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,token, userid, created) values ($1,$2, $3, $4) ", p.ID, p.Token, p.UserID, p.Created)
	return a.Error(ctx, qn, e)
}

func (a *DBPersistSession) Update(ctx context.Context, p *savepb.PersistSession) error {
	qn := "DBPersistSession_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set token=$1, userid=$2, created=$3 where id = $4", p.Token, p.UserID, p.Created, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBPersistSession) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBPersistSession_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBPersistSession) ByID(ctx context.Context, p uint64) (*savepb.PersistSession, error) {
	qn := "DBPersistSession_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,token, userid, created from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No PersistSession with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) PersistSession with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBPersistSession) All(ctx context.Context) ([]*savepb.PersistSession, error) {
	qn := "DBPersistSession_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,token, userid, created from "+a.SQLTablename+" order by id")
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

// get all "DBPersistSession" rows with matching Token
func (a *DBPersistSession) ByToken(ctx context.Context, p string) ([]*savepb.PersistSession, error) {
	qn := "DBPersistSession_ByToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,token, userid, created from "+a.SQLTablename+" where token = $1", p)
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
func (a *DBPersistSession) ByLikeToken(ctx context.Context, p string) ([]*savepb.PersistSession, error) {
	qn := "DBPersistSession_ByLikeToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,token, userid, created from "+a.SQLTablename+" where token ilike $1", p)
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

// get all "DBPersistSession" rows with matching UserID
func (a *DBPersistSession) ByUserID(ctx context.Context, p string) ([]*savepb.PersistSession, error) {
	qn := "DBPersistSession_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,token, userid, created from "+a.SQLTablename+" where userid = $1", p)
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
func (a *DBPersistSession) ByLikeUserID(ctx context.Context, p string) ([]*savepb.PersistSession, error) {
	qn := "DBPersistSession_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,token, userid, created from "+a.SQLTablename+" where userid ilike $1", p)
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

// get all "DBPersistSession" rows with matching Created
func (a *DBPersistSession) ByCreated(ctx context.Context, p uint32) ([]*savepb.PersistSession, error) {
	qn := "DBPersistSession_ByCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,token, userid, created from "+a.SQLTablename+" where created = $1", p)
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
func (a *DBPersistSession) ByLikeCreated(ctx context.Context, p uint32) ([]*savepb.PersistSession, error) {
	qn := "DBPersistSession_ByLikeCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,token, userid, created from "+a.SQLTablename+" where created ilike $1", p)
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
func (a *DBPersistSession) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.PersistSession, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBPersistSession) Tablename() string {
	return a.SQLTablename
}

func (a *DBPersistSession) SelectCols() string {
	return "id,token, userid, created"
}
func (a *DBPersistSession) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".token, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".created"
}

func (a *DBPersistSession) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.PersistSession, error) {
	var res []*savepb.PersistSession
	for rows.Next() {
		foo := savepb.PersistSession{}
		err := rows.Scan(&foo.ID, &foo.Token, &foo.UserID, &foo.Created)
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
func (a *DBPersistSession) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),token text not null  unique  ,userid text not null  ,created integer not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),token text not null  unique  ,userid text not null  ,created integer not null  );`,
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
func (a *DBPersistSession) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
