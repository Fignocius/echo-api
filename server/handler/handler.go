package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fignocius/echo-api/service/appconf"
	"github.com/fignocius/echo-api/service/user"
	"github.com/fignocius/echo-api/service/user/auth"
	"github.com/fignocius/echo-api/service/user/auth/rolecache"
	amw "github.com/fignocius/echo-api/service/user/auth/rolecache/mw"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	echoSwagger "github.com/pindamonhangaba/echo-swagger"
)

// Based on Google JSONC styleguide
// https://google.github.io/styleguide/jsoncstyleguide.xml

type errorResponse struct {
	Error generalError `json:"error"`
}

type generalError struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Errors  []detailError `json:"errors,omitempty"`
}

type detailError struct {
	Domain       string  `json:"domain"`
	Reason       string  `json:"reason"`
	Message      string  `json:"message"`
	Location     *string `json:"location,omitempty"`
	LocationType *string `json:"locationType,omitempty"`
	ExtendedHelp *string `json:"extendedHelp,omitempty"`
	SendReport   *string `json:"sendReport,omitempty"`
}

type dataResponse struct {
	// Client sets this value and server echos data in the response
	Context string `json:"context,omitempty"`
	Data    dataer `json:"data"`
}

type dataer interface {
	Data()
}

type dataDetail struct {
	// The kind property serves as a guide to what type of information this particular object stores
	Kind string `json:"kind" example:"resource"`
	// Indicates the language of the rest of the properties in this object (BCP 47)
	Language string `json:"lang,omitempty" example:"pt-br"`
}

func (d dataDetail) Data() {}

type singleItemData struct {
	dataDetail
	Item interface{} `json:"item"`
}

func (d singleItemData) Data() {}

type collectionItemData struct {
	dataDetail
	Items []interface{} `json:"items"`
	// The number of items in this result set
	CurrentItemCount int64 `json:"currentItemCount" example:"1"`
	// The number of items in the result
	ItemsPerPage int64 `json:"itemsPerPage" example:"10"`
	// The index of the first item in data.items
	StartIndex int64 `json:"startIndex" example:"1"`
	// The total number of items available in this set
	TotalItems int64 `json:"totalItems" example:"100"`
	// The index of the current page of items
	PageIndex int64 `json:"pageIndex" example:"1"`
	// The total number of pages in the result set.
	TotalPages int64 `json:"totalPages" example:"10"`
}

// HTTPServer create a service to echo server
type HTTPServer struct {
	DB    *sqlx.DB
	Roles *rolecache.RoleCache
}

// Run create a new echo server
func (u *HTTPServer) Run() {
	// Echo instance
	e := echo.New()
	e.Use(mw.Recover())
	e.Use(mw.Logger())

	/// CORS restricted
	// Allows requests from all origins
	// wth GET, PUT, POST or DELETE method.
	e.Use(mw.CORSWithConfig(mw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/", h)

	gAPI := e.Group("/api")
	jwtConfig := mw.JWTConfig{
		Claims:     &auth.Claims{},
		SigningKey: []byte(appconf.Secret),
		ContextKey: "user",
	}
	gAPI.Use(mw.JWTWithConfig(jwtConfig))
	gAPI.Use(amw.EchoMiddleware(u.Roles, amw.JWTConfig{
		RolesCtxKey: "roles",
		TokenCtxKey: "user",
	}))
	Onboarding(u.DB, e)
	RegisterTo(u.DB, e)
	Support(u.DB, e)
	RoutesConfig(u.DB, gAPI, u.Ecom)
	e.HTTPErrorHandler = httpErrorHandler

	fmt.Println("online")
	addr := appconf.App.Address
	e.Logger.Fatal(e.Start(addr))
}

func (d collectionItemData) Data() {}

// Public Routes
func RegisterTo(db *sqlx.DB, e *echo.Echo) error {
	ua := &user.Authenticator{
		DB: db,
		JWTConfig: user.JWTConfig{
			Secret:          appconf.Secret,
			HoursTillExpire: 72 * time.Hour,
			SigningMethod:   jwt.SigningMethodHS256,
		},
	}
	ah := &AuthHandler{signin: ua.Run}
	e.POST("/auth/signin", ah.EmailLogin)

	return nil
}

func Onboarding(db *sqlx.DB, e *echo.Echo) error {
	ua := &user.Authenticator{
		DB: db,
		JWTConfig: user.JWTConfig{
			Secret:          appconf.Secret,
			HoursTillExpire: 72 * time.Hour,
			SigningMethod:   jwt.SigningMethodHS256,
		},
	}
	cd := &user.DoctorCreator{DB: db}
	cdh := &DoctorHandler{create: cd.Run}
	e.POST("/onboarding/doctor", cdh.Create)
	cp := &user.PatientCreator{DB: db}
	cph := &PatientHandler{create: cp.Run, authenticate: ua.Run}
	e.POST("/onboarding/patient", cph.Create)
	return nil
}

// Private Routes
func RoutesConfig(db *sqlx.DB, e *echo.Group, ecom *cielo.Ecommerce) error {

	// Patients
	p := &user.PatientUpdater{DB: db}
	pg := &user.PatientGeter{DB: db}
	ph := &PatientHandler{update: p.Run, get: pg.Run}
	e.PUT("/patients/:pati_id", ph.Update)
	e.GET("/patients/:pati_id", ph.Get)

	return nil
}

func httpErrorHandler(err error, c echo.Context) {

	// since it's an api, it should always be in json
	// won't be using xml anytime soon
	//isJsonRequest := c.Request().Header().Get("Content-Type") == "application/json"
	fmt.Println("error", err)

	if e, ok := err.(*echo.HTTPError); ok {
		c.JSON(e.Code, errorResponse{
			Error: generalError{
				Code:    int64(e.Code),
				Message: e.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, errorResponse{
		Error: generalError{
			Message: err.Error(),
		},
	})
}
func h(c echo.Context) (err error) {

	return c.JSON(http.StatusOK, "It's works")
}
