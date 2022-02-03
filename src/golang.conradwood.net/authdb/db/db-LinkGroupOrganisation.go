package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBLinkGroupOrganisation
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence linkgrouporganisation_seq;

Main Table:

 CREATE TABLE linkgrouporganisation (id integer primary key default nextval('linkgrouporganisation_seq'),orgid bigint not null  ,groupid bigint not null  );

Alter statements:
ALTER TABLE linkgrouporganisation ADD COLUMN orgid bigint not null default 0;
ALTER TABLE linkgrouporganisation ADD COLUMN groupid bigint not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE linkgrouporganisation_archive (id integer unique not null,orgid bigint not null,groupid bigint not null);
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
	default_def_DBLinkGroupOrganisation *DBLinkGroupOrganisation
)

type DBLinkGroupOrganisation struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBLinkGroupOrganisation() *DBLinkGroupOrganisation {
	if default_def_DBLinkGroupOrganisation != nil {
		return default_def_DBLinkGroupOrganisation
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBLinkGroupOrganisation(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBLinkGroupOrganisation = res
	return res
}
func NewDBLinkGroupOrganisation(db *sql.DB) *DBLinkGroupOrganisation {
	foo := DBLinkGroupOrganisation{DB: db}
	foo.SQLTablename = "linkgrouporganisation"
	foo.SQLArchivetablename = "linkgrouporganisation_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBLinkGroupOrganisation) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBLinkGroupOrganisation", "insert into "+a.SQLArchivetablename+"+ (id,orgid, groupid) values ($1,$2, $3) ", p.ID, p.OrgID, p.GroupID)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBLinkGroupOrganisation) Save(ctx context.Context, p *savepb.LinkGroupOrganisation) (uint64, error) {
	qn := "DBLinkGroupOrganisation_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (orgid, groupid) values ($1, $2) returning id", p.OrgID, p.GroupID)
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
func (a *DBLinkGroupOrganisation) SaveWithID(ctx context.Context, p *savepb.LinkGroupOrganisation) error {
	qn := "insert_DBLinkGroupOrganisation"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,orgid, groupid) values ($1,$2, $3) ", p.ID, p.OrgID, p.GroupID)
	return a.Error(ctx, qn, e)
}

func (a *DBLinkGroupOrganisation) Update(ctx context.Context, p *savepb.LinkGroupOrganisation) error {
	qn := "DBLinkGroupOrganisation_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set orgid=$1, groupid=$2 where id = $3", p.OrgID, p.GroupID, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBLinkGroupOrganisation) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBLinkGroupOrganisation_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBLinkGroupOrganisation) ByID(ctx context.Context, p uint64) (*savepb.LinkGroupOrganisation, error) {
	qn := "DBLinkGroupOrganisation_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,orgid, groupid from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No LinkGroupOrganisation with id %d", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) LinkGroupOrganisation with id %d", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBLinkGroupOrganisation) All(ctx context.Context) ([]*savepb.LinkGroupOrganisation, error) {
	qn := "DBLinkGroupOrganisation_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,orgid, groupid from "+a.SQLTablename+" order by id")
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

// get all "DBLinkGroupOrganisation" rows with matching OrgID
func (a *DBLinkGroupOrganisation) ByOrgID(ctx context.Context, p uint64) ([]*savepb.LinkGroupOrganisation, error) {
	qn := "DBLinkGroupOrganisation_ByOrgID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,orgid, groupid from "+a.SQLTablename+" where orgid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOrgID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOrgID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBLinkGroupOrganisation) ByLikeOrgID(ctx context.Context, p uint64) ([]*savepb.LinkGroupOrganisation, error) {
	qn := "DBLinkGroupOrganisation_ByLikeOrgID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,orgid, groupid from "+a.SQLTablename+" where orgid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOrgID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOrgID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBLinkGroupOrganisation" rows with matching GroupID
func (a *DBLinkGroupOrganisation) ByGroupID(ctx context.Context, p uint64) ([]*savepb.LinkGroupOrganisation, error) {
	qn := "DBLinkGroupOrganisation_ByGroupID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,orgid, groupid from "+a.SQLTablename+" where groupid = $1", p)
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
func (a *DBLinkGroupOrganisation) ByLikeGroupID(ctx context.Context, p uint64) ([]*savepb.LinkGroupOrganisation, error) {
	qn := "DBLinkGroupOrganisation_ByLikeGroupID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,orgid, groupid from "+a.SQLTablename+" where groupid ilike $1", p)
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

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBLinkGroupOrganisation) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.LinkGroupOrganisation, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBLinkGroupOrganisation) Tablename() string {
	return a.SQLTablename
}

func (a *DBLinkGroupOrganisation) SelectCols() string {
	return "id,orgid, groupid"
}
func (a *DBLinkGroupOrganisation) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".orgid, " + a.SQLTablename + ".groupid"
}

func (a *DBLinkGroupOrganisation) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.LinkGroupOrganisation, error) {
	var res []*savepb.LinkGroupOrganisation
	for rows.Next() {
		foo := savepb.LinkGroupOrganisation{}
		err := rows.Scan(&foo.ID, &foo.OrgID, &foo.GroupID)
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
func (a *DBLinkGroupOrganisation) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),orgid bigint not null  ,groupid bigint not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),orgid bigint not null  ,groupid bigint not null  );`,
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
func (a *DBLinkGroupOrganisation) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
