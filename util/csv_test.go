package util

import (
	"bytes"
	"github.com/Rachatapon1994/assessment-tax/config"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var CSVHEADER = []string{"totalIncome", "wht", "donation"}

type mockHandlerContext struct {
	c echo.Context
	r *httptest.ResponseRecorder
}

func MockCalculationCsvHandler(c echo.Context) ([][]string, error) {
	fileForm, _ := c.FormFile("taxFile")
	csvBody, err := ReadCsvFile(fileForm, CSVHEADER)
	return csvBody, err
}

func mockContextMultipart(fieldName string, fileName string, fileContent string) mockHandlerContext {
	var buf bytes.Buffer
	multipartWriter := multipart.NewWriter(&buf)
	defer multipartWriter.Close()

	filePart, _ := multipartWriter.CreateFormFile(fieldName, fileName)
	filePart.Write([]byte(fileContent))

	e := echo.New()
	e.Validator = &config.CustomValidator{Validator: validator.New(validator.WithRequiredStructEnabled())}
	req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", &buf)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	rec := httptest.NewRecorder()
	return mockHandlerContext{
		e.NewContext(req, rec),
		rec,
	}
}

func TestReadCsvFile(t *testing.T) {

	mockContextMultipartCsvSuccess := mockContextMultipart("taxFile", "taxes.csv", "totalIncome,wht,donation\n500000,0,0\n600000,40000,20000\n750000,50000,15000\n")
	mockContextMultipartCsvErrorWhenCsvIsIncorrectFormat := mockContextMultipart("taxFile", "taxes.csv", "totalIncome,wht,donation\n5000000,0\n600000,40000,20000\n750000,50000,15000\n")
	mockContextMultipartCsvErrorWhenCsvHeaderIsInvalid := mockContextMultipart("taxFile", "taxes.csv", "totalIncome,wht,donation1\n500000,0,0\n600000,40000,20000\n750000,50000,15000\n")

	type args struct {
		c         mockHandlerContext
		csvHeader []string
	}
	tests := []struct {
		name    string
		args    args
		want    [][]string
		wantErr bool
	}{
		{"Should return response when csv is correct format", args{csvHeader: CSVHEADER, c: mockContextMultipartCsvSuccess}, [][]string{{"500000", "0", "0"}, {"600000", "40000", "20000"}, {"750000", "50000", "15000"}}, false},
		{"Should return error response when csv is incorrect format", args{csvHeader: CSVHEADER, c: mockContextMultipartCsvErrorWhenCsvIsIncorrectFormat}, nil, true},
		{"Should return error response when csv header is invalid", args{csvHeader: CSVHEADER, c: mockContextMultipartCsvErrorWhenCsvHeaderIsInvalid}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MockCalculationCsvHandler(tt.args.c.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCsvFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadCsvFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
