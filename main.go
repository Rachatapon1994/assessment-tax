package main

import (
	"context"
	"fmt"
	"github.com/Rachatapon1994/assessment-tax/admin"
	mw "github.com/Rachatapon1994/assessment-tax/middleware"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rachatapon1994/assessment-tax/config"
	"github.com/Rachatapon1994/assessment-tax/db"
	"github.com/Rachatapon1994/assessment-tax/tax"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func main() {
	if os.Getenv("PORT") != "8080" {
		log.Fatal(fmt.Sprintf("Port :%v could not be run, this program allow only port :8080", os.Getenv("PORT")))
	}

	db := db.InitDB()
	e := echo.New()
	e.Validator = &config.CustomValidator{Validator: validator.New(validator.WithRequiredStructEnabled())}

	tg := e.Group("/tax")
	taxHandler := tax.Handler{DB: db}
	tg.POST("/calculations", taxHandler.CalculationHandler)
	tg.POST("/calculations/upload-csv", taxHandler.CalculationCsvHandler)

	ag := e.Group("/admin")
	adminHandler := admin.Handler{DB: db}
	ag.Use(middleware.BasicAuth(mw.Authenticate()))
	ag.POST("/deductions/personal", adminHandler.DeductionPersonalHandler)
	ag.POST("/deductions/k-receipt", adminHandler.DeductionKReceiptHandler)

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
