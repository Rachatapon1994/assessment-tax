package admin

import (
	"database/sql"
	"github.com/Rachatapon1994/assessment-tax/db"
	"github.com/labstack/echo/v4"
	"net/http"
)

type (
	DeductionPersonal struct {
		Amount *float64 `json:"amount" validate:"required,numeric,gte=10000,lte=100000"`
	}

	DeductionKReceipt struct {
		Amount *float64 `json:"amount" validate:"required,numeric,gt=0,lte=100000"`
	}
)

var (
	PERSONAL = "personal"
	DONATION = "donation"
	KRECEIPT = "k-receipt"
)

type Handler struct {
	DB *sql.DB
}

type Err struct {
	Message string `json:"message"`
}

func (e *Err) Error() string {
	return e.Message
}

type PersonalResult struct {
	PersonalDeduction float64 `json:"personalDeduction"`
}

type KReceiptResult struct {
	KReceipt float64 `json:"kReceipt"`
}

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

func validateInput[T DeductionPersonal | DeductionKReceipt](c echo.Context, t *T) error {
	if err := c.Bind(&t); err != nil {
		return &Err{Message: "Error when binding JSON"}
	}
	if err := c.Validate(t); err != nil {
		return &Err{Message: "Validation fields does not pass"}
	}
	return nil
}

func (h *Handler) DeductionPersonalHandler(c echo.Context) error {
	dp := DeductionPersonal{}
	if err := validateInput(c, &dp); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	if err := (&db.Allowance{AllowanceType: PERSONAL, Amount: *dp.Amount}).UpdateByType(h.DB); err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, PersonalResult{PersonalDeduction: *dp.Amount})
}

func (h *Handler) DeductionKReceiptHandler(c echo.Context) error {
	dkr := DeductionKReceipt{}
	if err := validateInput(c, &dkr); err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	if err := (&db.Allowance{AllowanceType: KRECEIPT, Amount: *dkr.Amount}).UpdateByType(h.DB); err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	return c.JSON(http.StatusOK, KReceiptResult{KReceipt: *dkr.Amount})
}
