package tax

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"net/http"
)

type (
	Calculation struct {
		TotalIncome *float64    `json:"totalIncome" validate:"required,numeric,gte=0"`
		Wht         *float64    `json:"wht"`
		Allowances  []Allowance `json:"allowances" validate:"dive"`
	}

	Allowance struct {
		AllowanceType string   `json:"allowanceType"`
		Amount        *float64 `json:"amount"`
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
	Tax float64 `json:"tax"`
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
	calculator := &Calculator{TotalIncome: *tc.TotalIncome, Deductors: setDeductors(tc.Allowances, h.DB)}
	return c.JSON(http.StatusOK, Result{Tax: calculator.calculate()})
}
