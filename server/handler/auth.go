package handler

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/fignocius/echo-api/service/user"
	"net/http"
)

type AuthHandler struct {
	signin func(email, password string) (*user.AuthResponse, error)
}

// EmailLogin returns an echo handler
// @Summary auth.login
// @Description Login with email & password
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param credentials body handler.loginForm true "Email to login with"
// @Success 200 {object} handler.loginOut
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /doctors [post]
func (handler *AuthHandler) EmailLogin(c echo.Context) error {
	request := loginForm{}
	err := c.Bind(&request)
	if err != nil {
		return err
	}
	r, err := handler.signin(request.Email, request.Password)
	if err != nil {
		return errors.Wrap(err, "Fail to sign in")
	}
	return c.JSON(http.StatusOK, loginOut{
		Kind: "authToken",
		Item: authToken{
			User: r.User,
			JWT:  r.Jwt,
		},
	})
}

type loginForm struct {
	Email    string `json:"email" example:"user@mail.com"`
	Password string `json:"password" example:"mypassword123"`
}

type loginOut struct {
	singleItemData
	Item authToken `json:"item"`
	Kind string    `json:"kind" example:"authToken"`
}

type authToken struct {
	// User auth data
	User user.User `json:"user"`
	// JWT token
	JWT string `json:"jwt" example:"wqeoifjweoifjwef.afoj3204jfdkjf0wjf0wefj0w9fjf..."`
}
