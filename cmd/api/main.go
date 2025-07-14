package main

import (
	"fmt"
	"microser/db"
	"microser/handlers"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	_, err := db.InitDB()

	if err != nil {
		fmt.Println("Error al conectar con la DB", err)
	}

	api := echo.New()
	api.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"*"},
	}))
	api.Use(middleware.Recover())
	api.Use(handlers.TracingMiddleware("chedul-service"))

	api.GET("/users", handlers.HandleGetUsers)
	api.GET("/users/:id", handlers.HandleGetUser)
	api.POST("/users", handlers.HandlePostUser)
	api.PUT("/users", handlers.HandlePutUser)
	api.DELETE("/users/:id", handlers.HandleDeleteUser)

	api.Logger.Fatal(api.Start(os.Getenv("LISTEN_ADDR")))
}
