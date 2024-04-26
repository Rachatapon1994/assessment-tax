package tax

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"
)

func mockCalculatorDb(t *testing.T) *sql.DB {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	mock.MatchExpectationsInOrder(false)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rowsPersonal := mock.NewRows([]string{"id", "allowance_type", "amount"}).
		AddRow(1, "personal", 60000.00)

	SearchByTypeSql := "SELECT id, allowance_type, amount FROM allowance WHERE allowance_type = $1"
	mock.ExpectQuery(SearchByTypeSql).WithArgs("personal").WillReturnRows(rowsPersonal)
	return db
}

func TestPersonal_get(t *testing.T) {
	t.Parallel()
	type fields struct {
		DB *sql.DB
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{"Personal should get allowance correctly", fields{mockCalculatorDb(t)}, 60000.00},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.fields.DB.Close()
			p := &Personal{
				DB: tt.fields.DB,
			}
			if got := p.get(); got != tt.want {
				t.Errorf("Personal.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setDeductors(t *testing.T) {
	t.Parallel()

	type args struct {
		allowances []Allowance
		DB         *sql.DB
	}
	tests := []struct {
		name string
		args args
		want []Deductor
	}{
		{"Should return list of Deductor correctly when Allowance is empty", args{make([]Allowance, 0), mockCalculatorDb(t)}, make([]Deductor, 0)},
		{"Should return list of Deductor correctly when Allowance is not empty", args{[]Allowance{{AllowanceType: "personal"}}, mockCalculatorDb(t)}, []Deductor{&Personal{DB: mockCalculatorDb(t)}}},
		{"Should return list of Deductor correctly when Allowance is not empty", args{[]Allowance{{AllowanceType: "personal"}}, mockCalculatorDb(t)}, []Deductor{&Personal{DB: mockCalculatorDb(t)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.args.DB.Close()
			if got := setDeductors(tt.args.allowances, tt.args.DB); len(got) != len(tt.want) {
				t.Errorf("setDeductors() = %v, want %v", len(got), len(tt.want))
			}
		})
	}
}

func TestCalculator_sumDeduction(t *testing.T) {
	t.Parallel()
	type fields struct {
		TotalIncome float64
		Deductors   []Deductor
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{"Should return sum of deduction = 60000 when allowance has only personal deduction", fields{500000.0, []Deductor{&Personal{DB: mockCalculatorDb(t)}}}, 60000},
		{"Should return sum of deduction = 0 when Deduction is empty", fields{500000.0, make([]Deductor, 0)}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Calculator{
				TotalIncome: tt.fields.TotalIncome,
				Deductors:   tt.fields.Deductors,
			}
			if got := c.sumDeduction(); got != tt.want {
				t.Errorf("Calculator.sumDeduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateTaxLevels(t *testing.T) {
	t.Parallel()
	type args struct {
		income float64
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		{"Should return tax information for in first tier correctly", args{income: 100000}, []float64{0}},
		{"Should return tax information for end of first tier correctly", args{income: 150000}, []float64{0}},
		{"Should return tax information for in second tier correctly", args{income: 300000}, []float64{0, 15000}},
		{"Should return tax information for end of second tier correctly", args{income: 500000}, []float64{0, 35000}},
		{"Should return tax information for in third tier correctly", args{income: 750000}, []float64{0, 35000, 37500}},
		{"Should return tax information for end of third tier correctly", args{income: 1000000}, []float64{0, 35000, 75000}},
		{"Should return tax information for in fourth tier correctly", args{income: 1500000}, []float64{0, 35000, 75000, 100000}},
		{"Should return tax information for end of fourth tier correctly", args{income: 2000000}, []float64{0, 35000, 75000, 200000}},
		{"Should return tax information for fifth tier correctly", args{income: 2500000}, []float64{0, 35000, 75000, 200000, 175000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTaxLevels(tt.args.income); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateTaxLevels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculator_calculate(t *testing.T) {
	t.Parallel()
	type fields struct {
		TotalIncome float64
		Deductors   []Deductor
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{"Should return tax = 29000 when income = 500000 and allowance has only personal deduction", fields{500000.0, []Deductor{&Personal{DB: mockCalculatorDb(t)}}}, 29000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Calculator{
				TotalIncome: tt.fields.TotalIncome,
				Deductors:   tt.fields.Deductors,
			}
			if got := c.calculate(); got != tt.want {
				t.Errorf("Calculator.calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}
