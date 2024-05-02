package sql3

import (
	"context"
	"embed"
	"fmt"

	"github.com/google/uuid"
	"go.adoublef.dev/sdk/database/sql3"
	"go.pear-year.io/internal/user"
	"go.pear-year.io/text"
)

type DB struct {
	RWC *sql3.DB
}

func (d *DB) User(ctx context.Context, id uuid.UUID) (u user.User, n int, err error) {
	const q1 = `
select u.*, h.version
from users u
left join (
    select _rowid as rowid, max(_version) as version
    from _users_history
    group by _rowid
) h on u.rowid = h.rowid;	
	`
	err = d.RWC.QueryRow(ctx, q1, id).Scan(&u.ID, &u.Name, &u.Age, &n)
	if err != nil {
		return user.User{}, 0, err
	}
	return u, n, nil
}

func (d *DB) SetUser(ctx context.Context, name text.Name, age uint8) (uuid.UUID, error) {
	const q1 = `
insert into users (id, name, age) values (?, ?, ?) 
	`
	uid := uuid.Must(uuid.NewV7())
	_, err := d.RWC.Exec(ctx, q1, uid, name, age)
	if err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}

func (d *DB) Rename(ctx context.Context, id uuid.UUID, version int, name text.Name) error {
	const q1 = `
update users 
set name = ? 
where id = ? 
and exists (
	select 1
	from (
		select max(_version) as version
		from _users_history
		where user = ?
	) as h where h.version = ?
)
	`
	rs, err := d.RWC.Exec(ctx, q1, name, id, id, version)
	if err != nil {
		return err
	}
	if n, err := rs.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		return fmt.Errorf("%d rows where affected", n)
	}
	return nil
}

//go:embed all:*.up.sql
var embedFS embed.FS

func Up(ctx context.Context, filename string) (*DB, error) {
	rwc, err := sql3.Up(ctx, filename, embedFS)
	if err != nil {
		return nil, fmt.Errorf("running migration scripts for people domain: %w", err)
	}
	return &DB{rwc}, nil
}
