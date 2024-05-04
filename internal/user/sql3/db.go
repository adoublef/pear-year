package sql3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"iter"
	"log"
	"time"

	"github.com/google/uuid"
	"go.adoublef.dev/sdk/database/sql3"
	"go.adoublef.dev/sdk/time/date"
	"go.adoublef.dev/sdk/time/julian"
	"go.pear-year.io/internal/user"
	"go.pear-year.io/text"
)

type DB struct {
	RWC *sql3.DB
}

func (d *DB) User(ctx context.Context, id uuid.UUID) (user.User, uint, error) {
	const q1 = `
select u.id, u.name, u.dob, u.role, u._version from users u where u.id = ?	
`
	var u User
	var ver uint
	err := d.RWC.QueryRow(ctx, q1, id).Scan(&u.ID, &u.Name, &u.DOB, &u.Role, &ver)
	if err != nil {
		return user.User{}, 0, wrap(err)
	}
	return UserTo(u), ver, nil
}

// get from a particular version
func (d *DB) UserAt(ctx context.Context, id uuid.UUID, ver uint) (user.User, error) {
	const q1 = `
select id, name, dob, role, _mask 
from (
	select id, name, dob, role, _version, (1<<4)-1 as _mask
	from users
	where id = @id

	union

	select user as id, name, dob, role, _version, _mask
	from _users_history
	where _rowid = (select rowid from users where id = @id) and _version >= @version
) as agg
order by _version desc
`
	rs, err := d.RWC.Query(ctx, q1, sql.Named("id", id), sql.Named("version", ver))
	if err != nil {
		return user.User{}, wrap(err)
	}
	defer rs.Close()
	var u User
	for rs.Next() {
		var uid *uuid.UUID
		var name *text.Name
		var dob *julian.Time
		var role *user.Role
		var mask int
		err := rs.Scan(&uid, &name, &dob, &role, &mask)
		if err != nil {
			return user.User{}, wrap(err)
		}
		if mask&1 != 0 {
			u.ID = *uid
		}
		if mask&2 != 0 {
			u.Name = *name
		}
		if mask&4 != 0 {
			u.DOB = *dob
		}
		if mask&8 != 0 {
			u.Role = *role
		}
	}
	if err := rs.Err(); err != nil {
		return user.User{}, wrap(err)
	}
	if (User{}) == u {
		return user.User{}, user.ErrNotFound
	}
	return UserTo(u), nil
}

func (d *DB) History(ctx context.Context, id uuid.UUID, ver uint) ([]user.User, error) {
	const q1 = `
select id, name, dob, role, _mask 
from (
	select id, name, dob, role, _version, (1<<4)-1 as _mask
	from users
	where id = @id

	union

	select user as id, name, dob, role, _version, _mask
	from _users_history
	where _rowid = (select rowid from users where id = @id) and _version >= @version
) as agg
order by _version desc
`
	rs, err := d.RWC.Query(ctx, q1, sql.Named("id", id), sql.Named("version", ver))
	if err != nil {
		return nil, wrap(err)
	}
	defer rs.Close()

	var uu []user.User
	var u User
	for rs.Next() {
		var uid *uuid.UUID
		var name *text.Name
		var dob *julian.Time
		var role *user.Role
		var mask int
		err := rs.Scan(&uid, &name, &dob, &role, &mask)
		if err != nil {
			return nil, wrap(err)
		}
		if mask&1 != 0 {
			u.ID = *uid
		}
		if mask&2 != 0 {
			u.Name = *name
		}
		if mask&4 != 0 {
			u.DOB = *dob
		}
		if mask&8 != 0 {
			u.Role = *role
		}
		uu = append(uu, UserTo(u))
	}
	if err := rs.Err(); err != nil {
		return nil, wrap(err)
	}
	return uu, nil
}

func (d *DB) Iter(ctx context.Context, id uuid.UUID, ver uint) iter.Seq2[user.User, error] {
	panic("todo")
}

func (d *DB) SetUser(ctx context.Context, name text.Name, dob date.Date) (uuid.UUID, error) {
	const q1 = `
insert into users (id, name, dob, role, _version) values (?, ?, ?, ?, ?) 
`
	uid := uuid.Must(uuid.NewV7())
	_, err := d.RWC.Exec(ctx, q1, uid, name, julian.FromTime(dob.In(time.UTC)), user.Guest, 0)
	if err != nil {
		return uuid.Nil, wrap(err)
	}
	return uid, nil
}

func (d *DB) Rename(ctx context.Context, name text.Name, id uuid.UUID, ver uint) error {
	const q1 = `
update users set 
	name = ?
	, _version = _version + 1
where id = ? and _version = ?
`
	rs, err := d.RWC.Exec(ctx, q1, name, id, ver)
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

func (d *DB) SetDOB(ctx context.Context, dob date.Date, id uuid.UUID, ver uint) error {
	const q1 = `
update users set
	dob = ?
	, _version = _version + 1
where id = ? and _version = ?
`
	rs, err := d.RWC.Exec(ctx, q1, julian.FromTime(dob.In(time.UTC)), id, ver)
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

func (d *DB) SetRole(ctx context.Context, role user.Role, id uuid.UUID, ver uint8) error {
	const q1 = `
update users set
	role = ?
	, _version = _version + 1
where id = ? and _version = ?
	`
	rs, err := d.RWC.Exec(ctx, q1, role, id, ver)
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

type User struct {
	ID   uuid.UUID
	Name text.Name
	DOB  julian.Time
	Role user.Role
}

func UserTo(u User) user.User {
	return user.User{u.ID, u.Name, date.DateOf(u.DOB.Time()), u.Role}
}

func UserOf(u user.User) User {
	return User{u.ID, u.Name, julian.FromTime(u.DOB.In(time.UTC)), u.Role}
}
