package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// Permissions is a list of permissions for api usage
type Permissions []string

// Can checks if a permission is given
func (p Permissions) Can(perm ...string) bool {
	found := 0
	for _, e := range p {
		for _, o := range perm {
			if o == e {
				found++
			}
		}
	}
	return found == len(perm) && len(perm) > 0
}

// Claims is the claims for a JWT
type Claims struct {
	UserID string  `json:"userID"`
	Email  string  `json:"email"`
	jwt.StandardClaims
}

// Valid implement jwt.Claims
func (c Claims) Valid() error {
	return nil
}

func (c *Claims) String() string {
	return c.UserID
}

// Extract builds a Claims struct from an jwt.Token
func Extract(i interface{}) (c *Claims, err error) {
	userToken, ok := i.(*jwt.Token)
	if !ok {
		return c, errors.New("No token")
	}
	claims, ok := userToken.Claims.(*Claims)
	if !ok {
		return c, errors.New("No claims in token")
	}
	return claims, nil
}

// FromUnknown builds a Claims struct from an interface{}
func FromUnknown(i interface{}) (c Claims, err error) {
	var claimsMap map[string]interface{}
	if val, ok := i.(map[string]interface{}); ok {
		claimsMap = val
	} else if val, ok := i.(Claims); ok {
		return val, nil
	} else {
		err = errors.New("Couldn't parse user claims")
		return
	}

	if id, ok := claimsMap["user_id"].(string); ok {
		c.UserID = id
	} else {
		err = errors.New("Couldn't parse user ID")
	}

	return
}
