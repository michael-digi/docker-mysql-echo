package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

// Config holds sql connection and docker connection
type Config struct {
	SQL    *sqlx.DB
	Docker *docker.Client
	Echo   *echo.Echo
}

func checkAPIKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiKey := c.Request().Header.Get("X-Api-Key")

		if apiKey == "" {
			return nil
		}
		return next(c)
	}
}

// First get list of all containers (if you don't specify 'ALL: true' in types.ContainerListOptions,
// it defaults to only containers currently running). Remove the '/' from each c.Names[0] in the loop
// and then compare against the 'name' param passed in. Once found, grab the ID and use it in ContainerStart.
func (config *Config) startContainer(c echo.Context) error {
	containerToStart := c.Param("name")

	containers, err := config.Docker.ContainerList(context.Background(), types.ContainerListOptions{All: true})

	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		containerName := strings.Replace(c.Names[0], "/", "", -1)
		words := strings.Fields(c.Status)

		if containerName == containerToStart {
			if words[0] == "Exited" {
				config.Docker.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
				fmt.Println("Found container", containerToStart, "and started it")
			} else {
				fmt.Println("Container", containerToStart, "already running")
			}
		}
	}

	return nil
}

// First get list of all running containers. remove the '/' from each c.Names[0] in the loop
// and then compare against the 'name' param passed in. Once found, grab the ID and use it in ContainerStop.
func (config *Config) stopContainer(c echo.Context) error {
	containerToStop := c.Param("name")

	fmt.Println(containerToStop, "name")

	containers, err := config.Docker.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		containerName := strings.Replace(c.Names[0], "/", "", -1)
		words := strings.Fields(c.Status)
		if containerName == containerToStop {
			if words[0] == "Up" {
				config.Docker.ContainerStop(context.Background(), c.ID, nil)
				fmt.Println("Found container", containerToStop, "and stopped it")
			} else {
				fmt.Println("Container", containerToStop, "already stopped")
			}
		}
	}

	return nil
}

func (config *Config) listContainers(c echo.Context) error {
	containers := []Container{}

	// 'Select' is an sqlx statement that allows a direct reading from columns into an array of struct instances,
	// or any other type
	err := config.SQL.Select(&containers, `SELECT * FROM containers`)

	fmt.Println(containers, "these are containers")

	if err != nil {
		fmt.Println("this panicked")
		panic(err)
	}

	return c.JSON(http.StatusOK, &containers)
}

func (config *Config) insertContainers(c echo.Context) error {
	containers, err := config.Docker.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		panic(err)
	}

	// Begins a transaction, this is to ensure we're using the same connection from the pool throughout the
	// transaction's duration
	tx := config.SQL.MustBegin()
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

	// Commit 'saves' the executed transactions to the db and closes the open connection
	err = tx.Commit()

	if err != nil {
		panic(err)
	}

	return nil
}

func main() {
	var err error
	config := &Config{}

	config.Echo = echo.New()

	config.Docker, err = docker.NewEnvClient()

	config.SQL, err = sqlx.Open("mysql", "root:password@tcp(localhost)/test")

	if err != nil {
		fmt.Println("this panicked")
		panic(err)
	}

	config.Echo.Use(middleware.Logger())

	protected := config.Echo.Group("/containers", checkAPIKey)

	protected.GET("/add", config.insertContainers)

	protected.GET("/list", config.listContainers)

	protected.GET("/stop/:name", config.stopContainer)

	protected.GET("/start/:name", config.startContainer)

	config.Echo.Logger.Fatal(config.Echo.Start(":3000"))

	defer config.SQL.Close()

}
