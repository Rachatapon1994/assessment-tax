package tax

import (
	"database/sql"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Rachatapon1994/assessment-tax/config"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type mockHandlerContext struct {
	c echo.Context
	r *httptest.ResponseRecorder
}

func mockPostTaxCalculationContext(body string) mockHandlerContext {
	e := echo.New()
	e.Validator = &config.CustomValidator{Validator: validator.New(validator.WithRequiredStructEnabled())}
	req := httptest.NewRequest(http.MethodPost, "/tax/calculations", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return mockHandlerContext{
		e.NewContext(req, rec),
		rec,
	}
}

func mockHandlerDb(t *testing.T) *sql.DB {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	mock.MatchExpectationsInOrder(false)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rowsDonation := mock.NewRows([]string{"id", "allowance_type", "amount"}).
		AddRow(2, "donation", 100000.00)

	rowsPersonal := mock.NewRows([]string{"id", "allowance_type", "amount"}).
		AddRow(1, "personal", 60000.00)

	SearchByTypeSql := "SELECT id, allowance_type, amount FROM allowance WHERE allowance_type = $1"
	mock.ExpectQuery(SearchByTypeSql).WithArgs("personal").WillReturnRows(rowsPersonal)
	mock.ExpectQuery(SearchByTypeSql).WithArgs("donation").WillReturnRows(rowsDonation)
	return db
}

func TestErr_Error(t *testing.T) {
	t.Parallel()
	type fields struct {
		Message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Should return error message correctly", fields{"TEST ERROR MESSAGE"}, "TEST ERROR MESSAGE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Err{
				Message: tt.fields.Message,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Err.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateInput(t *testing.T) {
	t.Parallel()
	type args struct {
		c  mockHandlerContext
		tc *Calculation
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		wantErrorMessage string
	}{
		{"Should validate input failed when JSON is incorrect format", args{mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 0.0,  "allowances":     {      "allowanceType": "donation",      "amount": 0.0    }  ]}`), &Calculation{}}, true, "Error when binding JSON"},
		{"Should validate input failed when JSON data is not meet validator setup", args{mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 500001.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`), &Calculation{}}, true, "Validation fields does not pass"},
		{"Should validate input success when JSON data is correctly and meet validator setup", args{mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 0.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`), &Calculation{}}, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInput(tt.args.c.c, tt.args.tc)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if err.Error() != tt.wantErrorMessage {
					t.Errorf("validateInput() error message = %v, wantErrorMessage %v", err.Error(), tt.wantErrorMessage)
				}
			}
		})
	}
}

func TestHandler_CalculationHandler(t *testing.T) {
	t.Parallel()
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		c mockHandlerContext
	}

	mockContext400WhenInputFieldsNotMeetValidator := mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 500001.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`)
	mockContextSuccessWhenWhtZeroAndNotAllowance := mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 0.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`)
	mockContextSuccessWhenWht5000AndNotAllowance := mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 5000.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`)
	mockContextSuccessWhenWht5000AndDonation10000 := mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 5000.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 10000.0    }  ]}`)
	mockContextSuccessWhenWht28000AndDonation10000 := mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 28000.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 10000.0    }  ]}`)
	mockContextSuccessWhenWht30000AndDonation10000 := mockPostTaxCalculationContext(`{  "totalIncome": 500000.0,  "wht": 30000.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 10000.0    }  ]}`)

	tests := []struct {
		name               string
		fields             fields
		args               args
		wantResponseBody   interface{}
		wantResponseStatus int
	}{
		{"Should return response with status 400 input failed when JSON data is not meet validator setup", fields{DB: mockHandlerDb(t)}, args{c: mockContext400WhenInputFieldsNotMeetValidator}, Err{Message: "Validation fields does not pass"}, 400},
		{"Should return successful response when WHT = 0 and no allowance", fields{DB: mockHandlerDb(t)}, args{c: mockContextSuccessWhenWhtZeroAndNotAllowance}, Result{29000, 0, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 29000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}}, 200},
		{"Should return successful response when WHT = 5000 and no allowance", fields{DB: mockHandlerDb(t)}, args{c: mockContextSuccessWhenWht5000AndNotAllowance}, Result{24000, 0, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 29000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}}, 200},
		{"Should return successful response when WHT = 5000 and Donation = 10000", fields{DB: mockHandlerDb(t)}, args{c: mockContextSuccessWhenWht5000AndDonation10000}, Result{23000, 0, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 28000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}}, 200},
		{"Should return successful response when WHT = 28000 and Donation = 10000", fields{DB: mockHandlerDb(t)}, args{c: mockContextSuccessWhenWht28000AndDonation10000}, Result{0, 0, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 28000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}}, 200},
		{"Should return successful response when WHT = 30000 and Donation = 10000", fields{DB: mockHandlerDb(t)}, args{c: mockContextSuccessWhenWht30000AndDonation10000}, Result{0, 2000, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 28000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}}, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.fields.DB.Close()

			h := &Handler{
				DB: tt.fields.DB,
			}

			if err := h.CalculationHandler(tt.args.c.c); err != nil {
				t.Errorf("Handler.CalculationHandler() error = %v", err)
			}

			if tt.wantResponseStatus == 200 {
				result := Result{}
				if err := json.Unmarshal(tt.args.c.r.Body.Bytes(), &result); err != nil {
					t.Errorf("unable to unmarshal json: %v", err)
				}

				if !reflect.DeepEqual(result, tt.wantResponseBody) {
					t.Errorf("expected (%v), got (%v)", tt.wantResponseBody, result)
				}
			} else {
				result := Err{}
				if err := json.Unmarshal(tt.args.c.r.Body.Bytes(), &result); err != nil {
					t.Errorf("unable to unmarshal json: %v", err)
				}

				if !reflect.DeepEqual(result, tt.wantResponseBody) {
					t.Errorf("expected (%v), got (%v)", tt.wantResponseBody, result)
				}
			}

			if tt.args.c.r.Code != tt.wantResponseStatus {
				t.Errorf("expected (%v), got (%v)", tt.wantResponseStatus, tt.args.c.r.Code)
			}
		})
	}
}
