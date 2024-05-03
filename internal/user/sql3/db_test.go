package sql3_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.adoublef.dev/is"
	"go.pear-year.io/internal/user"
	. "go.pear-year.io/internal/user/sql3"
	"go.pear-year.io/text"
)

func Test_Up(t *testing.T) {
	is := is.NewRelaxed(t)

	_, err := Up(context.TODO(), t.TempDir()+"/test.db")
	is.NoErr(err)
}

func Test_DB_SetUser(t *testing.T) {
	t.Run("OK", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			age uint8     = 27
		)

		_, err := d.SetUser(context.TODO(), ada, age)
		is.NoErr(err)
	}))
}

func Test_DB_User(t *testing.T) {
	t.Run("OK", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			age uint8     = 27
		)

		uid, err := d.SetUser(context.TODO(), ada, age)
		is.NoErr(err)

		_, n, err := d.User(context.TODO(), uid)
		is.NoErr(err)

		is.Equal(n, 1)
	}))

	t.Run("ErrNotFound", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			age uint8     = 27
		)

		_, err := d.SetUser(context.TODO(), ada, age)
		is.NoErr(err)

		_, _, err = d.User(context.TODO(), uuid.New())
		is.Err(err, user.ErrNotFound)
	}))
}

func Test_DB_UserFrom(t *testing.T) {
	t.Run("OK", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			age uint8     = 27
		)

		_, err := d.SetUser(context.TODO(), ada, age)
		is.NoErr(err)

		// u, err := d.UserFrom(context.TODO(), uid, 1)
		// is.NoErr(err)

		// if testing.Verbose() {
		// 	t.Logf("u@1: %v\n", u)
		// }
	}))

	t.Run("Rename", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			age uint8     = 27

			alan text.Name = "Alan Turing"
		)

		uid, err := d.SetUser(context.TODO(), ada, age)
		is.NoErr(err)

		err = d.Rename(context.TODO(), uid, 1, alan)
		is.NoErr(err) // ada changed name to alan

		// err = d.Birthday(context.TODO(), uid, 2)
		// is.NoErr(err) // alan is 28

		// u, err := d.UserFrom(context.TODO(), uid, 2)
		// is.NoErr(err)

		// if testing.Verbose() {
		// 	t.Logf("u after 3: %v\n", u) // adam is 27
		// }
	}))
}

func Test_DB_Rename(t *testing.T) {
	t.Run("OK", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			age uint8     = 27

			alan text.Name = "Alan Turing"
		)

		uid, err := d.SetUser(context.TODO(), ada, age)
		is.NoErr(err)

		err = d.Rename(context.TODO(), uid, 1, alan)
		is.NoErr(err) // rename user

		_, n, err := d.User(context.TODO(), uid)
		is.NoErr(err)

		is.Equal(n, 2)
	}))

	t.Run("Err", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			age uint8     = 27

			alan text.Name = "Alan Turing"
		)

		uid, err := d.SetUser(context.TODO(), ada, age)
		is.NoErr(err)

		err = d.Rename(context.TODO(), uid, 2, alan)
		is.Err(err, user.ErrNotFound)
	}))
}

func run(f func(*testing.T, *DB)) func(*testing.T) {
	return func(t *testing.T) {
		db, err := Up(context.TODO(), t.TempDir()+"/test.db")
		if err != nil {
			t.Fatalf("running migrations scripts for %q", t.Name())
		}
		t.Cleanup(func() { db.RWC.Close() })
		f(t, db)
	}
}
