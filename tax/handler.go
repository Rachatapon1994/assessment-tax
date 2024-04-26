package tax

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"net/http"
)

type (
	Calculation struct {
		TotalIncome *float64    `json:"totalIncome" validate:"required,numeric,gte=0"`
		Wht         *float64    `json:"wht" validate:"required,numeric,gte=0,ltefield=TotalIncome"`
		Allowances  []Allowance `json:"allowances" validate:"dive"`
	}

	Allowance struct {
		AllowanceType string   `json:"allowanceType" validate:"oneof=donation"`
		Amount        *float64 `json:"amount" validate:"required,numeric,gte=0"`
	}
)

type HandlerFunction interface {
	CalculationHandler(c echo.Context) error
}

type Handler struct {
	DB *sql.DB
}

type Err struct {
	Message string `json:"message"`
}

func (e *Err) Error() string {
	return e.Message
}

type Result struct {
	Tax      float64    `json:"tax"`
	TaxLevel []TaxLevel `json:"taxLevel"`
}

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

func validateInput(c echo.Context, tc *Calculation) error {
	if err := c.Bind(&tc); err != nil {
		return &Err{Message: "Error when binding JSON"}
	}
	if err := c.Validate(tc); err != nil {
		return &Err{Message: "Validation fields does not pass"}
	}
	return nil
}

func (h *Handler) CalculationHandler(c echo.Context) error {
	tc := Calculation{}
	if err := validateInput(c, &tc); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	tc.Allowances = append(tc.Allowances, Allowance{AllowanceType: PERSONAL})
	calculator := &Calculator{TotalIncome: *tc.TotalIncome, Wht: *tc.Wht, Deductors: setDeductors(tc.Allowances, h.DB)}
	taxAmount, taxLevels := calculator.calculate()
	return c.JSON(http.StatusOK, Result{Tax: taxAmount, TaxLevel: taxLevels})
}
