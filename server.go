package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"InshortsAssignment/controller"
)


func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.POST("/getstate",controller.Getstatefromlatilongi)
	e.POST("/updatedb",controller.Updatemongodb)

	e.Logger.Fatal(e.Start(":1323"))
}
