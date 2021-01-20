package user

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/fignocius/echo-api/service/user/auth"
	"gopkg.in/guregu/null.v3"
	"time"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

//Role if the user type for roles
type Role []string

//User is a representation of the table user
type User struct {
	UserID uuid.UUID `db:"user_id" json:"userID"`
	Email  string    `db:"email" json:"email"`
	Password  []byte    `db:"password" json:"-"`
	Role      Role      `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	DeletedAt null.Time `db:"deleted_at" json:"deletedAt"`
}

type Getter struct {
	DB *sqlx.DB
}

func (g *Getter) Run(userID uuid.UUID) (*User, error) {
	tx, err := g.DB.Beginx()
	if err != nil {
		return nil, err
	}
	u, err := fromID(tx, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return u, err
}

// Create a new user
func newUser(u *User, password string) (*User, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating user uuid")
	}

	passHash, err := auth.PasswordGen(password)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to hash user password")
	}
	u.UserID = id
	u.Password = passHash
	return u, nil
}

// Save a user in the database
func saveUser(tx *sqlx.Tx, u *User) (*User, error) {

	query := psql.Insert(`"user"`).
		Columns("user_id", "email", "password", "role", "info").
		Values(u.UserID, u.Email, u.Password, u.Role, u.Info).
		Suffix("RETURNING *")

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating user sql")
	}

	err = tx.Get(u, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error inserting user")
	}
	return u, nil
}

// Update updates a user in the database
func updateUser(tx *sqlx.Tx, u *User) (*User, error) {

	query := psql.Update(`"user"`).
		Set("role", u.Role).
		Suffix("RETURNING *")

	query = query.Where(sq.Eq{"user_id": u.UserID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "Error generating user update sql")
	}

	err = tx.Get(u, qSQL, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Error user update sql")
	}

	return u, nil
}

// Update user email in the database
func updateEmail(tx *sqlx.Tx, email string, userID uuid.UUID) error {

	query := psql.Update(`"user"`).
		Set("email", email)

	query = query.Where(sq.Eq{"user_id": userID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "Error generating user email update sql")
	}

	_, err = tx.Exec(qSQL, args...)
	return errors.Wrap(err, "Error user email update sql")
}

// fromID get an User from the database
func fromID(tx *sqlx.Tx, userID uuid.UUID) (usr *User, err error) {
	usr = &User{}
	query := psql.Select("*").From(`"user"`).Where(sq.Eq{"user_id": userID, "deleted_at": nil})
	qSQL, args, err := query.ToSql()
	if err != nil {
		return usr, err
	}
	err = tx.Get(usr, qSQL, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return usr, &auth.UserNotFoundError{
				Message: "No user whit this id: " + userID.String(),
			}
		}
		return usr, err
	}
	return usr, err
}

// Soft delete user in the database
func softDeleteUser(tx *sqlx.Tx, userID uuid.UUID) error {
	query := psql.Update(`"user"`).Set("deleted_at", time.Now()).Where(sq.Eq{"user_id": userID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "Error generating user sql")
	}
	_, err = tx.Exec(qSQL, args...)
	return errors.Wrap(err, "Error soft deleting user")

}

//Experimentation functions
func fromEmail(db *sqlx.DB, email string) (usr *User, err error) {
	usr = &User{}
	query := psql.Select("*").From(`"user"`).Where(sq.Eq{"email": email, "deleted_at": nil})
	qSQL, args, err := query.ToSql()
	if err != nil {
		return usr, err
	}
	err = db.Get(usr, qSQL, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return usr, &auth.UserNotFoundError{
				Message: "No user with this email: " + email,
			}
		}
		return usr, err
	}
	return usr, err
}
