package sql3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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
	/*
		SELECT
			COALESCE(h.user, u.id) AS user_id,
			CASE WHEN (_mask & 2) != 0 THEN h.name ELSE u.name END AS name,
			CASE WHEN (_mask & 4) != 0 THEN h.age ELSE u.age END AS age,
			h._version
		FROM
			_users_history AS h
		LEFT JOIN
			users AS u ON h.user = u.id OR (h.user IS NULL AND u.id = ?)
		WHERE
			(h.user = ? OR h.user IS NULL)
			AND h._version = ?;
	*/

	const q1 = `
select 
    coalesce(h.user, u.id) AS id,
    case when (_mask & 2) != 0 then h.name ELSE u.name end as name,
    case when (_mask & 4) != 0 then h.age ELSE u.age end as age
from 
    _users_history h
left join
    users u ON h.user = u.id or (h.user is null and u.id = @id)
WHERE 
   (h.user = @id or h.user is null) and h._version = @version;
`
	err = d.RWC.QueryRow(ctx, q1, sql.Named("id", id), sql.Named("version", version)).Scan(&u.ID, &u.Name, &u.Age)
	if err != nil {
		return user.User{}, wrap(err)
	}
	return u, nil
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
