package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	jd "github.com/josephburnett/jd/lib"
	"github.com/satori/go.uuid"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gopkg.in/guregu/null.v3"
	"testing"
	"time"
)

func jsonCompare(one, two interface{}, excluded ...string) error {
	contains := func(s []string, e string) bool {
		for _, a := range s {
			if a == e {
				return true
			}
		}
		return false
	}
	render := func(d jd.Diff, exclude ...string) string {
		b := bytes.NewBuffer(nil)
		for _, el := range d {
			p, _ := json.Marshal(el.Path)
			if contains(exclude, string(p)) {
				continue
			}
			b.WriteString(el.Render())
		}
		return b.String()
	}
	ajson, err := json.Marshal(one)
	if err != nil {
		return err
	}
	bjson, err := json.Marshal(two)
	if err != nil {
		return err
	}
	fmt.Println(string(ajson))
	fmt.Println(string(bjson))

	a, _ := jd.ReadJsonString(string(ajson))
	b, _ := jd.ReadJsonString(string(bjson))
	diff := a.Diff(b)
	r := render(diff, excluded...)
	if len(r) > 0 {
		return errors.New(r)
	}
	return nil
}

func TestCreateUserValues(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	defer mockDB.Close()

	id, _ := uuid.NewV4()
	u := User{
		UserID:    id,
		Email:     "test@mail.com",
		Password:  []byte("123123"),
		Role:      []string{"admin"},
		Info:      Info{},
		DeletedAt: null.Time{Valid: false},
	}
	// copy struct because createUser modifies pointer
	uInitial := u

	r, _ := u.Role.Value()
	i, _ := u.Info.Value()
	rows := sqlmock.NewRows([]string{"user_id", "email", "password", "role", "info", "created_at", "deleted_at"}).
		// values to return from query
		AddRow(u.UserID.String(), u.Email, u.Password, r, i, time.Now(), nil)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "user" (.*) VALUES (.*) RETURNING *`).
		WithArgs(id, u.Email, u.Password, u.Role, u.Info).
		WillReturnRows(rows)

	// testing
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	tx, err := sqlxDB.Beginx()
	if err != nil {
		t.Errorf("Expected no error, but got %s instead", err)
	}

	// run tested function
	nUser, err := SaveUser(tx, &u)
	if err != nil {
		t.Errorf("Error creating user %s", err)
	}

	// compare user inserted and returned
	if err := jsonCompare(uInitial, nUser, `["created_at"]`); err != nil {
		t.Errorf("Mismatched values of \n %s", err)
	}

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed expectations %s", err)
	}
}

func TestCreateUserQuery(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	defer mockDB.Close()

	id, _ := uuid.NewV4()
	u := User{
		UserID:    id,
		Email:     "test@mail.com",
		Password:  []byte("123123"),
		Role:      []string{"admin"},
		Info:      Info{},
		DeletedAt: null.Time{Valid: false},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "user" (.*) VALUES (.*) RETURNING \*`).
		WithArgs(id, u.Email, u.Password, u.Role, u.Info)

	// testing
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	tx, err := sqlxDB.Beginx()
	if err != nil {
		t.Errorf("Expected no error, but got %s instead", err)
	}

	// ignore returns
	_, _ = createUser(tx, &u)

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("Failed expectations %s", err)
	}
}
