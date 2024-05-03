package sql3

import (
	"context"
	"database/sql"
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
select u.id, u.name, u.age, u._version from users u where u.id = ?	
	`
	err = d.RWC.QueryRow(ctx, q1, id).Scan(&u.ID, &u.Name, &u.Age, &n)
	if err != nil {
		return user.User{}, 0, err
	}
	return u, n, nil
}

func (d *DB) SetUser(ctx context.Context, name text.Name, age uint8) (uuid.UUID, error) {
	const q1 = `
insert into users (id, name, age, _version) values (?, ?, ?, ?) 
	`
	uid := uuid.Must(uuid.NewV7())
	_, err := d.RWC.Exec(ctx, q1, uid, name, age, 1)
	if err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}

func (d *DB) Rename(ctx context.Context, id uuid.UUID, version int, name text.Name) error {
	const q1 = `
update users set name = @name
	, _version = _version + 1
where id = @id 
and _version = @version
	`
	rs, err := d.RWC.Exec(ctx, q1, sql.Named("name", name), sql.Named("id", id), sql.Named("version", version))
	if err != nil {
		return err
	}
	if n, err := rs.RowsAffected(); err != nil {
		return err
	} else if n != 1 {
		// not found error if a row could not be updated
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
