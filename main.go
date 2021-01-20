package main

import (
	_ "database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/satori/go.uuid"
	"github.com/tidwall/buntdb"
	_ "github.com/fignocius/echo-api/docs" // docs is generated by Swag CLI, you have to import it.
	"github.com/fignocius/echo-api/server/handler"
	"github.com/fignocius/echo-api/service/appconf"
	"github.com/fignocius/echo-api/service/cielo"
	"github.com/fignocius/echo-api/service/user"
	"github.com/fignocius/echo-api/service/user/auth/rolecache"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 192.168.1.41:1323
// @BasePath /

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		appconf.DB.Host, appconf.DB.Port, appconf.DB.User, appconf.DB.Password, appconf.DB.Name)
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// in-memory cache for roles
	memDB, err := buntdb.Open(":memory:")
	if err != nil {
		panic(err)
	}
	defer memDB.Close()

	rcServ := &rolecache.RoleCache{
		DB: memDB,
		GetUserRoles: func(userID string) ([]string, error) {
			roles := []string{}
			UID, err := uuid.FromString(userID)
			if err != nil {
				return roles, err
			}
			g := user.Getter{DB: db}

			u, err := g.Run(UID)
			if err != nil {
				return roles, err
			}
			return u.Role, nil
		},
	}
	
	server := handler.HTTPServer{DB: db, Roles: rcServ}
	server.Run()
}