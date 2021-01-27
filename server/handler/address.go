package handler

import (
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"github.com/Fignocius/echo-api/backend/service/user"
	"net/http"
)

type AddressHandler struct {
	list   func(doctID uuid.UUID) (*user.Addresses, error)
	create func(*user.Address) (*user.Address, error)
	remove func(*user.Address) (string, error)
}

// List doctor address
// @Summary Address.List
// @Description Return a list of doctor address
// @Accept  json
// @Produce  json
// @Param doct_id path string true "Doctor id"
// @Success 200 {object} handler.listAddresses
// @Router /doctors/{doct_id}/addresses/ [get]
func (handler *AddressHandler) List(c echo.Context) error {
	did, err := uuid.FromString(c.Param("doct_id"))
	if err != nil {
		return err
	}
	r, err := handler.list(did)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, listAddresses{Kind: "Addresses", TotalItems: int64(len(*r)), Items: r})
}

// Create doctor address
// @Summary Address.Create
// @Description Create doctor address
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param credentials body handler.createAddress true "Create new address"
// @Success 200 {object} handler.singleAddress
// @Router /doctors/{doct_id}/addresses/ [post]
func (handler *AddressHandler) Create(c echo.Context) error {
	did, err := uuid.FromString(c.Param("doct_id"))
	if err != nil {
		return err
	}
	req := createAddress{}
	err = c.Bind(&req)
	if err != nil {
		return err
	}
	r, err := handler.create(&user.Address{DoctID: did, Description: req.Description, Location: req.Location})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, singleAddress{Kind: "Address", Item: r})
}

// Remove doctor address
// @Summary Address.Remove
// @Description Remove doctor address
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param credentials body handler.removeAddress true "Remove address"
// @Success 200 {object} handler.textResponse
// @Router /doctors/{doct_id}/addresses [put]
func (handler *AddressHandler) Remove(c echo.Context) error {
	did, err := uuid.FromString(c.Param("doct_id"))
	if err != nil {
		return err
	}
	req := removeAddress{}
	err = c.Bind(&req)
	if err != nil {
		return err
	}
	r, err := handler.remove(&user.Address{DoctID: did, AddrID: req.AddrID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, textResponse{Res: r})
}

type singleAddress struct {
	singleItemData
	Item *user.Address `json:"item"`
	Kind string        `json:"kind"`
}

type createAddress struct {
	Description string `json:"description"`
	Location    string `json:"location"`
}

type listAddresses struct {
	collectionItemData
	TotalItems int64           `json:"totalItems"`
	Items      *user.Addresses `json:"items"`
	Kind       string          `json:"kind"`
}

type removeAddress struct {
	AddrID int `json:"addrID"`
}
