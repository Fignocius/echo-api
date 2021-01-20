package user

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/dgrijalva/jwt-go"
	"github.com/fignocius/echo-api/service"
	"github.com/fignocius/echo-api/service/mailer"
	"github.com/fignocius/echo-api/service/user/auth"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

//Service Object Authentication
type AuthResponse struct {
	User    User
	Patient *Patient
	Doctor  *Doctor
	Jwt     string
}

type Authenticator struct {
	DB        *sqlx.DB
	JWTConfig JWTConfig
}

type JWTConfig struct {
	Secret          string
	HoursTillExpire time.Duration
	SigningMethod   *jwt.SigningMethodHMAC
}

func (u *Authenticator) Run(email, password string) (a *AuthResponse, err error) {
	usr, err := fromEmail(u.DB, email)
	if err != nil {
		return nil, err
	}

	p, d, err := getPatientOrDoctor(u.DB, usr.UserID)
	if err != nil {
		return nil, err
	}
	if p == nil && d == nil {
		return nil, errors.New("User with no attribution (not a doctor or a patient)")
	}

	var docid *string
	if d != nil {
		s := d.ID.String()
		docid = &s
	}
	var patid *string
	if p != nil {
		s := p.ID.String()
		patid = &s
	}

	opts := authOptions{
		user:      *usr,
		doctID:    docid,
		patiID:    patid,
		password:  password,
		jwtConfig: u.JWTConfig,
	}

	jwt, err := authenticate(opts)
	if err != nil {
		return nil, err
	}
	a = &AuthResponse{User: *usr, Jwt: jwt, Doctor: d, Patient: p}
	return a, err
}

func getPatientOrDoctor(db *sqlx.DB, userID uuid.UUID) (*Patient, *Doctor, error) {
	tx, err := db.Beginx()
	if err != nil {
		return nil, nil, err
	}

	d, err := getDoctorByUserID(tx, userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	p, err := getPatientByUserID(tx, userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, err
	}

	return p, d, nil
}

type authOptions struct {
	user      User
	doctID    *string
	patiID    *string
	password  string
	jwtConfig JWTConfig
}

func authenticate(c authOptions) (jwttoken string, err error) {
	err = bcrypt.CompareHashAndPassword(c.user.Password, []byte(c.password))
	if err != nil {
		return jwttoken, &auth.ValidationError{
			Messages: map[string]string{"password": "Wrong email/password combination"},
		}
	}

	claims := auth.Claims{
		UserID: c.user.UserID.String(),
		DoctID: c.doctID,
		PatiID: c.patiID,
		Email:  c.user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(c.jwtConfig.HoursTillExpire).UTC().Unix(),
		},
	}

	token := jwt.NewWithClaims(c.jwtConfig.SigningMethod, claims)
	jwttoken, err = token.SignedString([]byte(c.jwtConfig.Secret))

	return jwttoken, err
}

// PwdRecoverer starts an User`s password reset flow
type PwdRecoverer struct {
	DB     *sqlx.DB
	Config *service.ServicesConfig
}

func (p *PwdRecoverer) Run(email string) error {
	u, err := fromEmail(p.DB, email)
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve user for email "+email)
	}

	ac, secret, err := newActConfirmation(u.UserID, vPwd)
	if err != nil {
		return errors.Wrap(err, "Failed to create action confirmation")
	}

	tx, err := p.DB.Beginx()
	if err != nil {
		return errors.Wrap(err, "Failed to begin transaction")
	}

	err = confirmationSave(tx, ac)
	if err != nil {
		return errors.Wrap(err, "Failed to insert action confirmation")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Failed to commit")
	}

	e := struct {
		Name            string
		ConfirmationURL string
	}{
		Name:            "",
		ConfirmationURL: p.Config.APPURL + "/verification/" + ac.AcveID.String() + "/" + secret,
	}
	err = p.Mailer.SendPwdResetRequest(e)
	if err != nil {
		return errors.Wrap(err, "Failed to send")
	}
	return nil
}

// PwdReseter resets an User`s password
type PwdReseter struct {
	DB     *sqlx.DB
	Mailer *mailer.Mailer
}

func (p *PwdReseter) Run(acveID, verification, password string) error {
	tx, err := p.DB.Beginx()
	if err != nil {
		return err
	}

	psrt, err := confirmationFromID(tx, acveID)
	if err != nil {
		return err
	}

	// validate verification
	err = bcrypt.CompareHashAndPassword([]byte(psrt.Verification), []byte(verification))
	if err != nil {
		return &auth.ValidationError{
			Messages: map[string]string{"verification": "Invalid verification id"},
		}
	}

	// update user password
	passHash, err := auth.PasswordGen(password)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = updatePassword(tx, psrt.UserID, passHash)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "Failed to update user password")
	}

	// remove verification
	err = confirmationDelete(tx, &psrt)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "Failed to update user password")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Failed to commit password reset")
	}

	// TODO: send user name, email
	err = p.Mailer.SendPwdResetAlert()
	if err != nil {
		return errors.Wrap(err, "Failed to insert action confirmation")
	}
	return nil
}

func updatePassword(tx *sqlx.Tx, userID uuid.UUID, pass []byte) error {
	query := psql.Update(`"user"`).
		Set("password", pass).
		Where(sq.Eq{"user_id": userID})

	qSQL, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "Error generating user password update sql")
	}

	_, err = tx.Exec(qSQL, args...)
	if err != nil {
		return errors.Wrap(err, "Error updating user password")
	}
	return nil
}
