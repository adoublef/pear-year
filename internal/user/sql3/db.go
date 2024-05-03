package sql3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

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
		return user.User{}, 0, wrap(err)
	}
	return u, n, nil
}

// get from a particular version
func (d *DB) UserFrom(ctx context.Context, id uuid.UUID, version int) (u user.User, err error) {
	const q1 = `
select user, name, age, _mask
from _users_history
where _rowid = (select rowid from users where id = ?) and _version <= ?
order by _version asc;	
`
	// if no values then?
	rs, err := d.RWC.Query(ctx, q1, id, version)
	if err != nil {
		return user.User{}, wrap(err)
	}
	defer rs.Close()

	for rs.Next() {
		var uid *uuid.UUID
		var name *text.Name
		var age *uint8
		var mask int
		err := rs.Scan(&uid, &name, &age, &mask)
		if err != nil {
			return user.User{}, wrap(err)
		}

		switch mask {
		case 1:
			u.ID = *uid
		case 2:
			u.Name = *name
		case 4:
			u.Age = *age
		case 1 | 2:
			u.ID = *uid
			u.Name = *name
		case 1 | 4:
			u.ID = *uid
			u.Age = *age
		case 2 | 4:
			u.Name = *name
			u.Age = *age
		case 1 | 2 | 4:
			u.ID = *uid
			u.Name = *name
			u.Age = *age
		}
	}
	if err := rs.Err(); err != nil {
		return user.User{}, wrap(err)
	}
	return u, nil
}

type User struct {
	ID   *uuid.UUID
	Name *text.Name
	Age  *uint8
}

func (d *DB) SetUser(ctx context.Context, name text.Name, age uint8) (uuid.UUID, error) {
	const q1 = `
insert into users (id, name, age, _version) values (?, ?, ?, ?) 
	`
	uid := uuid.Must(uuid.NewV7())
	_, err := d.RWC.Exec(ctx, q1, uid, name, age, 1)
	if err != nil {
		return uuid.Nil, wrap(err)
	}
	return uid, nil
}

func (d *DB) Rename(ctx context.Context, id uuid.UUID, version int, name text.Name) error {
	const q1 = `
update users set 
	name = ?
	, _version = _version + 1
where id = ? 
and _version = ?
	`
	rs, err := d.RWC.Exec(ctx, q1, name, id, version)
	if err != nil {
		return wrap(err)
	}
	if n, err := rs.RowsAffected(); err != nil {
		return wrap(err)
	} else if n != 1 {
		return user.ErrNotFound
	}
	return nil
}

func (d *DB) Birthday(ctx context.Context, id uuid.UUID, version int) error {
	const q1 = `
update users set
	age = age + 1
	, _version = _version + 1
where id = ?
and _version = ?
	`
	rs, err := d.RWC.Exec(ctx, q1, id, version)
	if err != nil {
		return wrap(err)
	}
	if n, err := rs.RowsAffected(); err != nil {
		return wrap(err)
	} else if n != 1 {
		return user.ErrNotFound
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

func wrap(err error) error {
	log.Printf("user/sql3: external error occurred %v", err)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return user.ErrNotFound
	}
	return fmt.Errorf("user/sql3: unexpected error: %w", err)
}
