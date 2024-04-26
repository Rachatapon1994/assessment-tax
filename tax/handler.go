package tax

import (
	"database/sql"
	"fmt"
	"github.com/Rachatapon1994/assessment-tax/util"
	"github.com/labstack/echo/v4"
	"math"
	"net/http"
	"strconv"
)

var (
	CSVFILEKEY  = "taxFile"
	CSVFILENAME = "taxes.csv"
	CSVHEADER   = []string{"totalIncome", "wht", "donation"}
)

type (
	Calculation struct {
		TotalIncome *float64    `json:"totalIncome" validate:"required,numeric,gte=0"`
		Wht         *float64    `json:"wht" validate:"required,numeric,gte=0,ltefield=TotalIncome"`
		Allowances  []Allowance `json:"allowances" validate:"dive"`
	}

	Allowance struct {
		AllowanceType string   `json:"allowanceType" validate:"oneof=donation k-receipt"`
		Amount        *float64 `json:"amount" validate:"required,numeric,gte=0"`
	}
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

type Result struct {
	Tax       float64    `json:"tax"`
	TaxRefund float64    `json:"taxRefund"`
	TaxLevel  []TaxLevel `json:"taxLevel"`
}

type TaxLevel struct {
	Level string  `json:"level"`
	Tax   float64 `json:"tax"`
}

type CsvResult struct {
	Taxes []CsvTaxesResult `json:"taxes"`
}

type CsvTaxesResult struct {
	TotalIncome float64 `json:"totalIncome"`
	Tax         float64 `json:"tax"`
	TaxRefund   float64 `json:"taxRefund"`
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
	var result Result
	if math.Signbit(taxAmount) {
		result = Result{TaxRefund: math.Abs(taxAmount), TaxLevel: taxLevels}
	} else {
		result = Result{Tax: taxAmount, TaxLevel: taxLevels}
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) CalculationCsvHandler(c echo.Context) error {
	var err error
	fileForm, err := c.FormFile(CSVFILEKEY)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: fmt.Sprintf("No file key: %v in form-data", CSVFILEKEY)})
	}
	if fileForm.Filename != CSVFILENAME {
		return c.JSON(http.StatusBadRequest, Err{Message: fmt.Sprintf("File name must be %v", CSVFILENAME)})
	}
	csvBody, err := util.ReadCsvFile(fileForm, CSVHEADER)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	csvTaxesResultList := make([]CsvTaxesResult, 0)
	for _, bodys := range csvBody {
		totalIncome, err := strconv.ParseFloat(bodys[0], 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Err{Message: fmt.Sprintf("Cannot convert CSV data to float64 : %v", err)})
		}
		wht, err := strconv.ParseFloat(bodys[1], 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Err{Message: fmt.Sprintf("Cannot convert CSV data to float64 : %v", err)})
		}
		donation, err := strconv.ParseFloat(bodys[2], 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, Err{Message: fmt.Sprintf("Cannot convert CSV data to float64 : %v", err)})
		}
		allowances := []Allowance{{AllowanceType: PERSONAL}, {AllowanceType: DONATION, Amount: &donation}}
		calculator := &Calculator{TotalIncome: totalIncome, Wht: wht, Deductors: setDeductors(allowances, h.DB)}
		taxAmount, _ := calculator.calculate()
		var csvTaxesResult CsvTaxesResult
		if math.Signbit(taxAmount) {
			csvTaxesResult = CsvTaxesResult{TaxRefund: math.Abs(taxAmount), TotalIncome: totalIncome}
		} else {
			csvTaxesResult = CsvTaxesResult{Tax: taxAmount, TotalIncome: totalIncome}
		}
		csvTaxesResultList = append(csvTaxesResultList, csvTaxesResult)
	}
	return c.JSON(http.StatusOK, CsvResult{Taxes: csvTaxesResultList})
}
