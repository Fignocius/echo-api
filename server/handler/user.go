package handler

import (
	"net/http"

	"github.com/labstack/echo"
	//"gitlab.com/falqon/matchmed/backend/service/user"
)

type UserHandler struct {
	signup func(email, password string) (*interface{}, error)
}

// Signup godoc
// @Summary Signup to service
// @Description User signup with email & password
// @Accept  json
// @Produce  json
// @Success 200 {object} handler.User
// @Router /signup [post]
func (handler *UserHandler) Signup(c echo.Context) error {
	return c.JSON(http.StatusOK, "ok")
}

// GetSignup godoc
// @Summary Show a account
// @Description get string by ID
// @ID get-string-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} handler.User
// @Router /signup/{id} [get]
func (handler *UserHandler) GetSignup(c echo.Context) error {
	return c.JSON(http.StatusOK, "okok")
}

type User struct {
	ID   int    `json:"id" example:"1"`
	Name string `json:"name" example:"account name"`
}
