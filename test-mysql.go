package main

import (
	"context"
	"net/http"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// Container holds info about docker container
type Container struct {
	ID      string `db:"id" json:"id"`
	Image   string `db:"image" json:"image"`
	ImageID string `db:"image_id" json:"imageId"`
	Name    string `db:"name" json:"name"`
	Command string `db:"command" json:"command"`
	Created int64  `db:"created" json:"created"`
	State   string `db:"state" json:"state"`
	Status  string `db:"status" json:"status"`
}

var db *sqlx.DB

func checkAPIKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiKey := c.Request().Header.Get("X-Api-Key")

		if apiKey == "" {
			return nil
		}
		return next(c)
	}
}

func listContainers(c echo.Context) error {
	containers := []Container{}

	// 'Select' is an sqlx statement that allows a direct reading from columns into an array of struct instances,
	// or any other type
	err := db.Select(&containers, `SELECT * FROM containers`)

	if err != nil {
		panic(err)
	}

	return c.JSON(http.StatusOK, &containers)
}

func insertContainers(c echo.Context) error {
	cli, err := docker.NewEnvClient()

	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		panic(err)
	}

	// begins a transaction, this is to ensure we're using the same connection from the pool throughout the
	// transaction's duration
	tx := db.MustBegin()
	statement := `
		INSERT INTO containers(id, image, image_id, name, command, created, state, status) 
		VALUES(:id, :image, :image_id, :name, :command, :created, :state, :status)`

	// NamedExec is a named statement transaction, evidenced by the :value scheme in the VALUES above. This allows
	// for direct use of structs in the Exec call
	for _, c := range containers {

		container := Container{c.ID[:10], c.Image, c.ImageID, c.Names[0], c.Command, c.Created, c.State, c.Status}
		_, err := tx.NamedExec(statement, container)

		if err != nil {
			panic(err)
		}
	}

	// commit 'saves' the executed transactions to the db and closes the open connection
	err = tx.Commit()

	if err != nil {
		panic(err)
	}

	return nil
}

func main() {
	var err error
	e := echo.New()

	db, err = sqlx.Open("mysql", "root:password@tcp(mysqlDB)/test")

	if err != nil {
		panic(err)
	}

	e.Use(middleware.Logger())

	protected := e.Group("/containers", checkAPIKey)

	protected.GET("/add", insertContainers)

	protected.GET("/list", listContainers)

	e.Logger.Fatal(e.Start(":3000"))

	defer db.Close()

}
