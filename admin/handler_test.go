package admin

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Rachatapon1994/assessment-tax/config"
	mw "github.com/Rachatapon1994/assessment-tax/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

type mockHandlerContext struct {
	c echo.Context
	r *httptest.ResponseRecorder
}

func mockPostAdminPersonalContext(body string) mockHandlerContext {
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "secret")

	e := echo.New()
	e.Validator = &config.CustomValidator{Validator: validator.New(validator.WithRequiredStructEnabled())}
	e.Use(middleware.BasicAuth(mw.Authenticate()))
	req := httptest.NewRequest(http.MethodPost, "/admin/deductions/personal", strings.NewReader(body))
	auth := "basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, auth)
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

	updateAllowanceSql := "UPDATE allowance SET amount = $1 WHERE allowance_type = $2"
	mock.ExpectExec(updateAllowanceSql).WithArgs(10000.0, PERSONAL).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(updateAllowanceSql).WithArgs(50000.0, PERSONAL).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(updateAllowanceSql).WithArgs(100000.0, PERSONAL).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(updateAllowanceSql).WithArgs(88888.0, PERSONAL).WillReturnError(sql.ErrConnDone)
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
		tc *DeductionPersonal
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		wantErrorMessage string
	}{
		{"Should validate input failed when JSON is incorrect format", args{mockPostAdminPersonalContext(`{  "amount": 60001.0 `), &DeductionPersonal{}}, true, "Error when binding JSON"},
		{"Should validate input failed when JSON data is not meet validator setup", args{mockPostAdminPersonalContext(`{  "amount": 9999.0}`), &DeductionPersonal{}}, true, "Validation fields does not pass"},
		{"Should validate input success when JSON data is correctly and meet validator setup", args{mockPostAdminPersonalContext(`{  "amount": 60001.0}`), &DeductionPersonal{}}, false, ""},
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

func TestHandler_DeductionPersonalHandler(t *testing.T) {
	t.Parallel()

	mockContext400WhenAmount9999 := mockPostAdminPersonalContext(`{  "amount": 9999.0}`)
	mockContext200WhenAmount10000 := mockPostAdminPersonalContext(`{  "amount": 10000.0}`)
	mockContext200WhenAmount50000 := mockPostAdminPersonalContext(`{  "amount": 50000.0}`)
	mockContext200WhenAmount100000 := mockPostAdminPersonalContext(`{  "amount": 100000.0}`)
	mockContext200WhenAmount100001 := mockPostAdminPersonalContext(`{  "amount": 100001.0}`)
	mockContext500WhenAmount88888 := mockPostAdminPersonalContext(`{  "amount": 88888.0}`)

	type fields struct {
		DB *sql.DB
	}
	type args struct {
		c mockHandlerContext
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		wantResponseBody   interface{}
		wantResponseStatus int
	}{
		{"Should return response with status 400 when amount = 9999", fields{DB: mockHandlerDb(t)}, args{c: mockContext400WhenAmount9999}, Err{Message: "Validation fields does not pass"}, 400},
		{"Should return successful response when amount = 10000", fields{DB: mockHandlerDb(t)}, args{c: mockContext200WhenAmount10000}, PersonalResult{10000}, 200},
		{"Should return successful response when amount = 50000", fields{DB: mockHandlerDb(t)}, args{c: mockContext200WhenAmount50000}, PersonalResult{50000}, 200},
		{"Should return successful response when amount = 100000", fields{DB: mockHandlerDb(t)}, args{c: mockContext200WhenAmount100000}, PersonalResult{100000}, 200},
		{"Should return successful response when amount = 100001", fields{DB: mockHandlerDb(t)}, args{c: mockContext200WhenAmount100001}, Err{Message: "Validation fields does not pass"}, 400},
		{"Should return unsuccessful response when amount = 88888 due to mock response error to ErrConnDone", fields{DB: mockHandlerDb(t)}, args{c: mockContext500WhenAmount88888}, Err{Message: sql.ErrConnDone.Error()}, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.fields.DB.Close()

			h := &Handler{
				DB: tt.fields.DB,
			}

			if err := h.DeductionPersonalHandler(tt.args.c.c); err != nil {
				t.Errorf("Handler.CalculationHandler() error = %v", err)
			}

			if tt.wantResponseStatus == 200 {
				result := PersonalResult{}
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
