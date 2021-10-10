package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBGroupDB
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence groups_seq;

Main Table:

 CREATE TABLE groups (id integer primary key default nextval('groups_seq'),name text not null  ,description text not null  );

Alter statements:
ALTER TABLE groups ADD COLUMN name text not null default '';
ALTER TABLE groups ADD COLUMN description text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE groups_archive (id integer unique not null,name text not null,description text not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/auth"
	"golang.conradwood.net/go-easyops/sql"
)

type DBGroupDB struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func NewDBGroupDB(db *sql.DB) *DBGroupDB {
	foo := DBGroupDB{DB: db}
	foo.SQLTablename = "groups"
	foo.SQLArchivetablename = "groups_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBGroupDB) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBGroupDB", "insert into "+a.SQLArchivetablename+"+ (id,name, description) values ($1,$2, $3) ", p.ID, p.Name, p.Description)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBGroupDB) Save(ctx context.Context, p *savepb.GroupDB) (uint64, error) {
	qn := "DBGroupDB_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (name, description) values ($1, $2) returning id", p.Name, p.Description)
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
func (a *DBGroupDB) SaveWithID(ctx context.Context, p *savepb.GroupDB) error {
	qn := "insert_DBGroupDB"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,name, description) values ($1,$2, $3) ", p.ID, p.Name, p.Description)
	return a.Error(ctx, qn, e)
}

func (a *DBGroupDB) Update(ctx context.Context, p *savepb.GroupDB) error {
	qn := "DBGroupDB_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set name=$1, description=$2 where id = $3", p.Name, p.Description, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBGroupDB) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBGroupDB_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBGroupDB) ByID(ctx context.Context, p uint64) (*savepb.GroupDB, error) {
	qn := "DBGroupDB_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, description from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No GroupDB with id %d", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) GroupDB with id %d", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBGroupDB) All(ctx context.Context) ([]*savepb.GroupDB, error) {
	qn := "DBGroupDB_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, description from "+a.SQLTablename+" order by id")
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

// get all "DBGroupDB" rows with matching Name
func (a *DBGroupDB) ByName(ctx context.Context, p string) ([]*savepb.GroupDB, error) {
	qn := "DBGroupDB_ByName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, description from "+a.SQLTablename+" where name = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupDB) ByLikeName(ctx context.Context, p string) ([]*savepb.GroupDB, error) {
	qn := "DBGroupDB_ByLikeName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, description from "+a.SQLTablename+" where name ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupDB" rows with matching Description
func (a *DBGroupDB) ByDescription(ctx context.Context, p string) ([]*savepb.GroupDB, error) {
	qn := "DBGroupDB_ByDescription"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, description from "+a.SQLTablename+" where description = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupDB) ByLikeDescription(ctx context.Context, p string) ([]*savepb.GroupDB, error) {
	qn := "DBGroupDB_ByLikeDescription"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name, description from "+a.SQLTablename+" where description ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBGroupDB) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.GroupDB, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBGroupDB) Tablename() string {
	return a.SQLTablename
}

func (a *DBGroupDB) SelectCols() string {
	return "id,name, description"
}
func (a *DBGroupDB) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".name, " + a.SQLTablename + ".description"
}

func (a *DBGroupDB) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.GroupDB, error) {
	var res []*savepb.GroupDB
	for rows.Next() {
		foo := savepb.GroupDB{}
		err := rows.Scan(&foo.ID, &foo.Name, &foo.Description)
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
func (a *DBGroupDB) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),name text not null  ,description text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),name text not null  ,description text not null  );`,
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
func (a *DBGroupDB) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
