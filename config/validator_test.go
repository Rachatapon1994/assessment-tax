package config

import (
	"encoding/json"
	"testing"

	"github.com/Rachatapon1994/assessment-tax/tax"

	"github.com/go-playground/validator/v10"
)

func TestValidateError_Error(t *testing.T) {
	t.Parallel()
	type fields struct {
		Message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Should return error message correctly", fields{Message: "mock error message"}, "mock error message"},
		{"Should return error message correctly", fields{Message: "mock error message1"}, "mock error message1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ValidateError{
				Message: tt.fields.Message,
			}
			if got := m.Error(); got != tt.want {
				t.Errorf("ValidateError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomValidator_Validate(t *testing.T) {
	t.Parallel()

	mockJsonSuccess := &tax.Calculation{}
	json.Unmarshal([]byte(`{  "totalIncome": 500000.0,  "wht": 40000.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`), mockJsonSuccess)

	mockJsonWhtMoreThanTotalIncome := &tax.Calculation{}
	json.Unmarshal([]byte(`{  "totalIncome": 500000.0,  "wht": 500001.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`), mockJsonWhtMoreThanTotalIncome)

	mockJsonTotalIncomeLessThanZero := &tax.Calculation{}
	json.Unmarshal([]byte(`{  "totalIncome": -1.0,  "wht": 40000.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`), mockJsonTotalIncomeLessThanZero)

	mockJsonWhtLessThanZero := &tax.Calculation{}
	json.Unmarshal([]byte(`{  "totalIncome": 500000.0,  "wht": -1.0,  "allowances": [    {      "allowanceType": "donation",      "amount": 0.0    }  ]}`), mockJsonWhtLessThanZero)

	mockJsonAllowanceTypeNotInTheList := &tax.Calculation{}
	json.Unmarshal([]byte(`{  "totalIncome": 500000.0,  "wht": 40000.0,  "allowances": [    {      "allowanceType": "insurance",      "amount": 0.0    }  ]}`), mockJsonAllowanceTypeNotInTheList)

	mockJsonAmountLessThanZero := &tax.Calculation{}
	json.Unmarshal([]byte(`{  "totalIncome": 500000.0,  "wht": 40000.0,  "allowances": [    {      "allowanceType": "donation",      "amount": -10.0    }  ]}`), mockJsonAmountLessThanZero)

	type fields struct {
		Validator *validator.Validate
	}
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Should validate success when JSON data meet validator", fields{Validator: validator.New(validator.WithRequiredStructEnabled())}, args{i: mockJsonSuccess}, false},
		{"Should validate unsuccessful when Wht  > Total Income", fields{Validator: validator.New(validator.WithRequiredStructEnabled())}, args{i: mockJsonWhtMoreThanTotalIncome}, true},
		{"Should validate unsuccessful when Total Income < 0", fields{Validator: validator.New(validator.WithRequiredStructEnabled())}, args{i: mockJsonTotalIncomeLessThanZero}, true},
		{"Should validate unsuccessful when Wht < 0", fields{Validator: validator.New(validator.WithRequiredStructEnabled())}, args{i: mockJsonWhtLessThanZero}, true},
		{"Should validate unsuccessful when Allowance Type is not in the validator list", fields{Validator: validator.New(validator.WithRequiredStructEnabled())}, args{i: mockJsonAllowanceTypeNotInTheList}, true},
		{"Should validate unsuccessful when Amount < 0", fields{Validator: validator.New(validator.WithRequiredStructEnabled())}, args{i: mockJsonAmountLessThanZero}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := &CustomValidator{
				Validator: tt.fields.Validator,
			}
			if err := cv.Validate(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("CustomValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
