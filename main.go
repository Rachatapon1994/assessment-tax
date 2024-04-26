package main

import (
	"context"
	"fmt"
	"github.com/Rachatapon1994/assessment-tax/config"
	"github.com/Rachatapon1994/assessment-tax/db"
	"github.com/Rachatapon1994/assessment-tax/tax"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if os.Getenv("PORT") != "8080" {
		log.Fatal(fmt.Sprintf("Port :%v could not be run, this program allow only port :8080", os.Getenv("PORT")))
	}

	db := db.InitDB()
	e := echo.New()
	e.Validator = &config.CustomValidator{Validator: validator.New(validator.WithRequiredStructEnabled())}

	taxHandler := tax.Handler{DB: db}
	e.POST("/tax/calculations", taxHandler.CalculationHandler)

	go func() {
		if err := e.Start(fmt.Sprintf(":%v", os.Getenv("PORT"))); err != nil && err != http.ErrServerClosed { // Start server
			e.Logger.Fatal(fmt.Sprintf("shutting down the server caused from :%v", err.Error()))
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	fmt.Println("Graceful shutting down the server process")
	if err := e.Shutdown(context.Background()); err != nil {
		e.Logger.Fatal(err)
	}
}
