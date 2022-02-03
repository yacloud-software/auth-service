package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBOrganisation
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence organisation_seq;

Main Table:

 CREATE TABLE organisation (id integer primary key default nextval('organisation_seq'),name text not null  );

Alter statements:
ALTER TABLE organisation ADD COLUMN name text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE organisation_archive (id integer unique not null,name text not null);
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
	default_def_DBOrganisation *DBOrganisation
)

type DBOrganisation struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBOrganisation() *DBOrganisation {
	if default_def_DBOrganisation != nil {
		return default_def_DBOrganisation
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBOrganisation(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBOrganisation = res
	return res
}
func NewDBOrganisation(db *sql.DB) *DBOrganisation {
	foo := DBOrganisation{DB: db}
	foo.SQLTablename = "organisation"
	foo.SQLArchivetablename = "organisation_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBOrganisation) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBOrganisation", "insert into "+a.SQLArchivetablename+"+ (id,name) values ($1,$2) ", p.ID, p.Name)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBOrganisation) Save(ctx context.Context, p *savepb.Organisation) (uint64, error) {
	qn := "DBOrganisation_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (name) values ($1) returning id", p.Name)
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
func (a *DBOrganisation) SaveWithID(ctx context.Context, p *savepb.Organisation) error {
	qn := "insert_DBOrganisation"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,name) values ($1,$2) ", p.ID, p.Name)
	return a.Error(ctx, qn, e)
}

func (a *DBOrganisation) Update(ctx context.Context, p *savepb.Organisation) error {
	qn := "DBOrganisation_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set name=$1 where id = $2", p.Name, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBOrganisation) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBOrganisation_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBOrganisation) ByID(ctx context.Context, p uint64) (*savepb.Organisation, error) {
	qn := "DBOrganisation_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Organisation with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Organisation with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBOrganisation) All(ctx context.Context) ([]*savepb.Organisation, error) {
	qn := "DBOrganisation_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name from "+a.SQLTablename+" order by id")
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

// get all "DBOrganisation" rows with matching Name
func (a *DBOrganisation) ByName(ctx context.Context, p string) ([]*savepb.Organisation, error) {
	qn := "DBOrganisation_ByName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name from "+a.SQLTablename+" where name = $1", p)
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
func (a *DBOrganisation) ByLikeName(ctx context.Context, p string) ([]*savepb.Organisation, error) {
	qn := "DBOrganisation_ByLikeName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,name from "+a.SQLTablename+" where name ilike $1", p)
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

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBOrganisation) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Organisation, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBOrganisation) Tablename() string {
	return a.SQLTablename
}

func (a *DBOrganisation) SelectCols() string {
	return "id,name"
}
func (a *DBOrganisation) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".name"
}

func (a *DBOrganisation) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Organisation, error) {
	var res []*savepb.Organisation
	for rows.Next() {
		foo := savepb.Organisation{}
		err := rows.Scan(&foo.ID, &foo.Name)
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
func (a *DBOrganisation) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),name text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),name text not null  );`,
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
func (a *DBOrganisation) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
