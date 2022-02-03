package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBEmailVerifyPins
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence emailverifypins_seq;

Main Table:

 CREATE TABLE emailverifypins (id integer primary key default nextval('emailverifypins_seq'),userid bigint not null  ,pin text not null  ,created bigint not null  ,accepted bigint not null  );

Alter statements:
ALTER TABLE emailverifypins ADD COLUMN userid bigint not null default 0;
ALTER TABLE emailverifypins ADD COLUMN pin text not null default '';
ALTER TABLE emailverifypins ADD COLUMN created bigint not null default 0;
ALTER TABLE emailverifypins ADD COLUMN accepted bigint not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE emailverifypins_archive (id integer unique not null,userid bigint not null,pin text not null,created bigint not null,accepted bigint not null);
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
	default_def_DBEmailVerifyPins *DBEmailVerifyPins
)

type DBEmailVerifyPins struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBEmailVerifyPins() *DBEmailVerifyPins {
	if default_def_DBEmailVerifyPins != nil {
		return default_def_DBEmailVerifyPins
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBEmailVerifyPins(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBEmailVerifyPins = res
	return res
}
func NewDBEmailVerifyPins(db *sql.DB) *DBEmailVerifyPins {
	foo := DBEmailVerifyPins{DB: db}
	foo.SQLTablename = "emailverifypins"
	foo.SQLArchivetablename = "emailverifypins_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBEmailVerifyPins) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBEmailVerifyPins", "insert into "+a.SQLArchivetablename+"+ (id,userid, pin, created, accepted) values ($1,$2, $3, $4, $5) ", p.ID, p.UserID, p.Pin, p.Created, p.Accepted)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBEmailVerifyPins) Save(ctx context.Context, p *savepb.EmailVerifyPins) (uint64, error) {
	qn := "DBEmailVerifyPins_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (userid, pin, created, accepted) values ($1, $2, $3, $4) returning id", p.UserID, p.Pin, p.Created, p.Accepted)
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
func (a *DBEmailVerifyPins) SaveWithID(ctx context.Context, p *savepb.EmailVerifyPins) error {
	qn := "insert_DBEmailVerifyPins"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,userid, pin, created, accepted) values ($1,$2, $3, $4, $5) ", p.ID, p.UserID, p.Pin, p.Created, p.Accepted)
	return a.Error(ctx, qn, e)
}

func (a *DBEmailVerifyPins) Update(ctx context.Context, p *savepb.EmailVerifyPins) error {
	qn := "DBEmailVerifyPins_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set userid=$1, pin=$2, created=$3, accepted=$4 where id = $5", p.UserID, p.Pin, p.Created, p.Accepted, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBEmailVerifyPins) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBEmailVerifyPins_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBEmailVerifyPins) ByID(ctx context.Context, p uint64) (*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No EmailVerifyPins with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) EmailVerifyPins with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBEmailVerifyPins) All(ctx context.Context) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" order by id")
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

// get all "DBEmailVerifyPins" rows with matching UserID
func (a *DBEmailVerifyPins) ByUserID(ctx context.Context, p uint64) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where userid = $1", p)
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
func (a *DBEmailVerifyPins) ByLikeUserID(ctx context.Context, p uint64) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where userid ilike $1", p)
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

// get all "DBEmailVerifyPins" rows with matching Pin
func (a *DBEmailVerifyPins) ByPin(ctx context.Context, p string) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByPin"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where pin = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPin: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPin: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBEmailVerifyPins) ByLikePin(ctx context.Context, p string) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByLikePin"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where pin ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPin: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPin: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBEmailVerifyPins" rows with matching Created
func (a *DBEmailVerifyPins) ByCreated(ctx context.Context, p uint64) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where created = $1", p)
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
func (a *DBEmailVerifyPins) ByLikeCreated(ctx context.Context, p uint64) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByLikeCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where created ilike $1", p)
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

// get all "DBEmailVerifyPins" rows with matching Accepted
func (a *DBEmailVerifyPins) ByAccepted(ctx context.Context, p uint64) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByAccepted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where accepted = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAccepted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAccepted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBEmailVerifyPins) ByLikeAccepted(ctx context.Context, p uint64) ([]*savepb.EmailVerifyPins, error) {
	qn := "DBEmailVerifyPins_ByLikeAccepted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, pin, created, accepted from "+a.SQLTablename+" where accepted ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAccepted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAccepted: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBEmailVerifyPins) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.EmailVerifyPins, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBEmailVerifyPins) Tablename() string {
	return a.SQLTablename
}

func (a *DBEmailVerifyPins) SelectCols() string {
	return "id,userid, pin, created, accepted"
}
func (a *DBEmailVerifyPins) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userid, " + a.SQLTablename + ".pin, " + a.SQLTablename + ".created, " + a.SQLTablename + ".accepted"
}

func (a *DBEmailVerifyPins) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.EmailVerifyPins, error) {
	var res []*savepb.EmailVerifyPins
	for rows.Next() {
		foo := savepb.EmailVerifyPins{}
		err := rows.Scan(&foo.ID, &foo.UserID, &foo.Pin, &foo.Created, &foo.Accepted)
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
func (a *DBEmailVerifyPins) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid bigint not null  ,pin text not null  ,created bigint not null  ,accepted bigint not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid bigint not null  ,pin text not null  ,created bigint not null  ,accepted bigint not null  );`,
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
func (a *DBEmailVerifyPins) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
