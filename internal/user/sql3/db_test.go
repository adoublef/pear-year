package sql3_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.adoublef.dev/is"
	"go.adoublef.dev/sdk/time/date"
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
			dob           = must(date.Parse("1818-10-12"))
		)

		_, err := d.SetUser(context.TODO(), ada, dob)
		is.NoErr(err)
	}))
}

func Test_DB_User(t *testing.T) {
	t.Run("OK", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			dob           = must(date.Parse("1818-10-12"))
		)

		uid, err := d.SetUser(context.TODO(), ada, dob)
		is.NoErr(err)

		_, n, err := d.User(context.TODO(), uid)
		is.NoErr(err)

		is.Equal(n, uint(1))
	}))

	t.Run("ErrNotFound", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			dob           = must(date.Parse("1818-10-12"))
		)

		_, err := d.SetUser(context.TODO(), ada, dob)
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
			dob           = must(date.Parse("1818-10-12"))
		)

		uid, err := d.SetUser(context.TODO(), ada, dob)
		is.NoErr(err)

		_, err = d.UserFrom(context.TODO(), uid, 3)
		is.NoErr(err)
	}))

	t.Run("ErrNotFound", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			dob           = must(date.Parse("1818-10-12"))
		)

		uid, err := d.SetUser(context.TODO(), ada, dob)
		is.NoErr(err) // (name=ada,age=27) (version=1)

		_, err = d.UserFrom(context.TODO(), uid, 0)
		is.Err(err, user.ErrNotFound)
	}))

	t.Run("Multi", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			dob           = must(date.Parse("1818-10-12"))

			alan text.Name = "Alan Turing" // June 23, 1912.
			dob2           = must(date.Parse("1912-06-23"))
		)

		uid, err := d.SetUser(context.TODO(), ada, dob)
		is.NoErr(err) //  (version=1)

		err = d.Rename(context.TODO(), alan, uid, 1)
		is.NoErr(err) // (version=2)

		err = d.Birthday(context.TODO(), dob2, uid, 2)
		is.NoErr(err) // (version=3)

		// if version=0 or version=max.Int this still works
		u, err := d.UserFrom(context.TODO(), uid, 2)
		is.NoErr(err)

		is.Equal(u.Name, alan)
		is.Equal(u.DOB, dob)
	}))
}

func Test_DB_Rename(t *testing.T) {
	t.Run("OK", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			dob           = must(date.Parse("1818-10-12"))

			alan text.Name = "Alan Turing"
		)

		uid, err := d.SetUser(context.TODO(), ada, dob)
		is.NoErr(err)

		err = d.Rename(context.TODO(), alan, uid, 1)
		is.NoErr(err) // rename user

		_, n, err := d.User(context.TODO(), uid)
		is.NoErr(err)

		is.Equal(n, uint(2))
	}))

	t.Run("Err", run(func(t *testing.T, d *DB) {
		is := is.NewRelaxed(t)

		// insert user args
		var (
			ada text.Name = "Ada Lovelace"
			dob           = must(date.Parse("1818-10-12"))

			alan text.Name = "Alan Turing"
		)

		uid, err := d.SetUser(context.TODO(), ada, dob)
		is.NoErr(err)

		err = d.Rename(context.TODO(), alan, uid, 2)
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

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
