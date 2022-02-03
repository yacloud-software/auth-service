package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBSudoStatus
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence sudostatus_seq;

Main Table:

 CREATE TABLE sudostatus (id integer primary key default nextval('sudostatus_seq'),userid bigint not null  ,groupid text not null  ,expiry integer not null  );

Alter statements:
ALTER TABLE sudostatus ADD COLUMN userid bigint not null default 0;
ALTER TABLE sudostatus ADD COLUMN groupid text not null default '';
ALTER TABLE sudostatus ADD COLUMN expiry integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE sudostatus_archive (id integer unique not null,userid bigint not null,groupid text not null,expiry integer not null);
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
	default_def_DBSudoStatus *DBSudoStatus
)

type DBSudoStatus struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBSudoStatus() *DBSudoStatus {
	if default_def_DBSudoStatus != nil {
		return default_def_DBSudoStatus
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBSudoStatus(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBSudoStatus = res
	return res
}
func NewDBSudoStatus(db *sql.DB) *DBSudoStatus {
	foo := DBSudoStatus{DB: db}
	foo.SQLTablename = "sudostatus"
	foo.SQLArchivetablename = "sudostatus_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBSudoStatus) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBSudoStatus", "insert into "+a.SQLArchivetablename+"+ (id,userid, groupid, expiry) values ($1,$2, $3, $4) ", p.ID, p.UserID, p.GroupID, p.Expiry)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBSudoStatus) Save(ctx context.Context, p *savepb.SudoStatus) (uint64, error) {
	qn := "DBSudoStatus_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (userid, groupid, expiry) values ($1, $2, $3) returning id", p.UserID, p.GroupID, p.Expiry)
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
func (a *DBSudoStatus) SaveWithID(ctx context.Context, p *savepb.SudoStatus) error {
	qn := "insert_DBSudoStatus"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,userid, groupid, expiry) values ($1,$2, $3, $4) ", p.ID, p.UserID, p.GroupID, p.Expiry)
	return a.Error(ctx, qn, e)
}

func (a *DBSudoStatus) Update(ctx context.Context, p *savepb.SudoStatus) error {
	qn := "DBSudoStatus_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set userid=$1, groupid=$2, expiry=$3 where id = $4", p.UserID, p.GroupID, p.Expiry, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBSudoStatus) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBSudoStatus_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBSudoStatus) ByID(ctx context.Context, p uint64) (*savepb.SudoStatus, error) {
	qn := "DBSudoStatus_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, groupid, expiry from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No SudoStatus with id %d", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) SudoStatus with id %d", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBSudoStatus) All(ctx context.Context) ([]*savepb.SudoStatus, error) {
	qn := "DBSudoStatus_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, groupid, expiry from "+a.SQLTablename+" order by id")
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

// get all "DBSudoStatus" rows with matching UserID
func (a *DBSudoStatus) ByUserID(ctx context.Context, p uint64) ([]*savepb.SudoStatus, error) {
	qn := "DBSudoStatus_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, groupid, expiry from "+a.SQLTablename+" where userid = $1", p)
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
func (a *DBSudoStatus) ByLikeUserID(ctx context.Context, p uint64) ([]*savepb.SudoStatus, error) {
	qn := "DBSudoStatus_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, groupid, expiry from "+a.SQLTablename+" where userid ilike $1", p)
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

// get all "DBSudoStatus" rows with matching GroupID
func (a *DBSudoStatus) ByGroupID(ctx context.Context, p string) ([]*savepb.SudoStatus, error) {
	qn := "DBSudoStatus_ByGroupID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, groupid, expiry from "+a.SQLTablename+" where groupid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSudoStatus) ByLikeGroupID(ctx context.Context, p string) ([]*savepb.SudoStatus, error) {
	qn := "DBSudoStatus_ByLikeGroupID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, groupid, expiry from "+a.SQLTablename+" where groupid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSudoStatus" rows with matching Expiry
func (a *DBSudoStatus) ByExpiry(ctx context.Context, p uint32) ([]*savepb.SudoStatus, error) {
	qn := "DBSudoStatus_ByExpiry"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, groupid, expiry from "+a.SQLTablename+" where expiry = $1", p)
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
func (a *DBSudoStatus) ByLikeExpiry(ctx context.Context, p uint32) ([]*savepb.SudoStatus, error) {
	qn := "DBSudoStatus_ByLikeExpiry"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, groupid, expiry from "+a.SQLTablename+" where expiry ilike $1", p)
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

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBSudoStatus) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.SudoStatus, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBSudoStatus) Tablename() string {
	return a.SQLTablename
}

func (a *DBSudoStatus) SelectCols() string {
	return "id,userid, groupid, expiry"
}
func (a *DBSudoStatus) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userid, " + a.SQLTablename + ".groupid, " + a.SQLTablename + ".expiry"
}

func (a *DBSudoStatus) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.SudoStatus, error) {
	var res []*savepb.SudoStatus
	for rows.Next() {
		foo := savepb.SudoStatus{}
		err := rows.Scan(&foo.ID, &foo.UserID, &foo.GroupID, &foo.Expiry)
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
func (a *DBSudoStatus) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid bigint not null  ,groupid text not null  ,expiry integer not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid bigint not null  ,groupid text not null  ,expiry integer not null  );`,
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
func (a *DBSudoStatus) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
