package user

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/fignocius/echo-api/service/user/auth"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type confirmationType string

const (
	vEmail = confirmationType("email")
	vPwd   = confirmationType("password")
)

type actionConfirmation struct {
	AcveID       uuid.UUID `db:"acve_id"`
	UserID       uuid.UUID `db:"user_id"`
	Type         confirmationType
	Verification string
	CreatedAt    time.Time
	DeletedAt    *time.Time
}

func newActConfirmation(u uuid.UUID, t confirmationType) (*actionConfirmation, string, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, "", err
	}

	v, err := auth.PasswordGen(uid.String())
	if err != nil {
		return nil, "", err
	}

	resetUUID, err := uuid.NewV4()
	if err != nil {
		return nil, "", err
	}

	return &actionConfirmation{AcveID: resetUUID, UserID: u, Verification: string(v)}, uid.String(), nil
}

func confirmationSave(db *sqlx.Tx, u *actionConfirmation) error {

	ins := psql.Insert("action_verification").
		Columns(
			"acve_id",
			"user_id",
			"verification",
			"type",
			"created_at").
		Values(
			u.AcveID,
			u.UserID,
			u.Verification,
			u.Type,
			time.Now())

	qSQL, args, err := ins.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(qSQL, args...)
	return err
}

func confirmationDelete(db *sqlx.Tx, u *actionConfirmation) error {
	del := psql.Update("action_verification").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"acve_id": u.AcveID, "user_id": u.UserID})

	qSQL, args, err := del.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(qSQL, args...)
	return err
}

func confirmationFromID(tx *sqlx.Tx, acveID string) (actionConfirmation, error) {
	psrt := actionConfirmation{}
	query := psql.Select("*").
		From("action_verification").
		Where(sq.Eq{"acveID": acveID, "deleted_at": nil})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return psrt, errors.Wrap(err, "Error generating user password update sql")
	}

	err = tx.Get(psrt, qSQL, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return psrt, &auth.PwdResetInvalidError{
				Message: "No such reset token: " + acveID,
			}
		}
		return psrt, err
	}
	return psrt, nil
}
