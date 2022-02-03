package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBRemoteUserDetail
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence remoteuserdetail_seq;

Main Table:

 CREATE TABLE remoteuserdetail (id integer primary key default nextval('remoteuserdetail_seq'),userid text not null  ,provider text not null  ,ourtoken text not null  ,created integer not null  ,remoteuserid text not null  );

Alter statements:
ALTER TABLE remoteuserdetail ADD COLUMN userid text not null default '';
ALTER TABLE remoteuserdetail ADD COLUMN provider text not null default '';
ALTER TABLE remoteuserdetail ADD COLUMN ourtoken text not null default '';
ALTER TABLE remoteuserdetail ADD COLUMN created integer not null default 0;
ALTER TABLE remoteuserdetail ADD COLUMN remoteuserid text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE remoteuserdetail_archive (id integer unique not null,userid text not null,provider text not null,ourtoken text not null,created integer not null,remoteuserid text not null);
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
	default_def_DBRemoteUserDetail *DBRemoteUserDetail
)

type DBRemoteUserDetail struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBRemoteUserDetail() *DBRemoteUserDetail {
	if default_def_DBRemoteUserDetail != nil {
		return default_def_DBRemoteUserDetail
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBRemoteUserDetail(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBRemoteUserDetail = res
	return res
}
func NewDBRemoteUserDetail(db *sql.DB) *DBRemoteUserDetail {
	foo := DBRemoteUserDetail{DB: db}
	foo.SQLTablename = "remoteuserdetail"
	foo.SQLArchivetablename = "remoteuserdetail_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBRemoteUserDetail) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBRemoteUserDetail", "insert into "+a.SQLArchivetablename+"+ (id,userid, provider, ourtoken, created, remoteuserid) values ($1,$2, $3, $4, $5, $6) ", p.ID, p.UserID, p.Provider, p.OurToken, p.Created, p.RemoteUserID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBRemoteUserDetail) Save(ctx context.Context, p *savepb.RemoteUserDetail) (uint64, error) {
	qn := "DBRemoteUserDetail_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (userid, provider, ourtoken, created, remoteuserid) values ($1, $2, $3, $4, $5) returning id", p.UserID, p.Provider, p.OurToken, p.Created, p.RemoteUserID)
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
func (a *DBRemoteUserDetail) SaveWithID(ctx context.Context, p *savepb.RemoteUserDetail) error {
	qn := "insert_DBRemoteUserDetail"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,userid, provider, ourtoken, created, remoteuserid) values ($1,$2, $3, $4, $5, $6) ", p.ID, p.UserID, p.Provider, p.OurToken, p.Created, p.RemoteUserID)
	return a.Error(ctx, qn, e)
}

func (a *DBRemoteUserDetail) Update(ctx context.Context, p *savepb.RemoteUserDetail) error {
	qn := "DBRemoteUserDetail_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set userid=$1, provider=$2, ourtoken=$3, created=$4, remoteuserid=$5 where id = $6", p.UserID, p.Provider, p.OurToken, p.Created, p.RemoteUserID, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBRemoteUserDetail) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBRemoteUserDetail_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBRemoteUserDetail) ByID(ctx context.Context, p uint64) (*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No RemoteUserDetail with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) RemoteUserDetail with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBRemoteUserDetail) All(ctx context.Context) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" order by id")
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

// get all "DBRemoteUserDetail" rows with matching UserID
func (a *DBRemoteUserDetail) ByUserID(ctx context.Context, p string) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where userid = $1", p)
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
func (a *DBRemoteUserDetail) ByLikeUserID(ctx context.Context, p string) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where userid ilike $1", p)
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

// get all "DBRemoteUserDetail" rows with matching Provider
func (a *DBRemoteUserDetail) ByProvider(ctx context.Context, p string) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByProvider"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where provider = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProvider: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProvider: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRemoteUserDetail) ByLikeProvider(ctx context.Context, p string) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByLikeProvider"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where provider ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProvider: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByProvider: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRemoteUserDetail" rows with matching OurToken
func (a *DBRemoteUserDetail) ByOurToken(ctx context.Context, p string) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByOurToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where ourtoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOurToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOurToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRemoteUserDetail) ByLikeOurToken(ctx context.Context, p string) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByLikeOurToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where ourtoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOurToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOurToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRemoteUserDetail" rows with matching Created
func (a *DBRemoteUserDetail) ByCreated(ctx context.Context, p uint32) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where created = $1", p)
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
func (a *DBRemoteUserDetail) ByLikeCreated(ctx context.Context, p uint32) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByLikeCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where created ilike $1", p)
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

// get all "DBRemoteUserDetail" rows with matching RemoteUserID
func (a *DBRemoteUserDetail) ByRemoteUserID(ctx context.Context, p string) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByRemoteUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where remoteuserid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRemoteUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRemoteUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRemoteUserDetail) ByLikeRemoteUserID(ctx context.Context, p string) ([]*savepb.RemoteUserDetail, error) {
	qn := "DBRemoteUserDetail_ByLikeRemoteUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, provider, ourtoken, created, remoteuserid from "+a.SQLTablename+" where remoteuserid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRemoteUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRemoteUserID: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBRemoteUserDetail) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.RemoteUserDetail, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBRemoteUserDetail) Tablename() string {
	return a.SQLTablename
}

func (a *DBRemoteUserDetail) SelectCols() string {
	return "id,userid, provider, ourtoken, created, remoteuserid"
}
func (a *DBRemoteUserDetail) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userid, " + a.SQLTablename + ".provider, " + a.SQLTablename + ".ourtoken, " + a.SQLTablename + ".created, " + a.SQLTablename + ".remoteuserid"
}

func (a *DBRemoteUserDetail) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.RemoteUserDetail, error) {
	var res []*savepb.RemoteUserDetail
	for rows.Next() {
		foo := savepb.RemoteUserDetail{}
		err := rows.Scan(&foo.ID, &foo.UserID, &foo.Provider, &foo.OurToken, &foo.Created, &foo.RemoteUserID)
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
func (a *DBRemoteUserDetail) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid text not null  ,provider text not null  ,ourtoken text not null  ,created integer not null  ,remoteuserid text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid text not null  ,provider text not null  ,ourtoken text not null  ,created integer not null  ,remoteuserid text not null  );`,
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
func (a *DBRemoteUserDetail) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
