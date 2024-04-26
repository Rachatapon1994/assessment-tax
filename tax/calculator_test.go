package tax

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func mockCalculatorDb(t *testing.T) *sql.DB {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	mock.MatchExpectationsInOrder(false)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rowsPersonal := mock.NewRows([]string{"id", "allowance_type", "amount"}).
		AddRow(1, "personal", 60000.00)
	rowsDonation := mock.NewRows([]string{"id", "allowance_type", "amount"}).
		AddRow(2, "donation", 100000.00)

	SearchByTypeSql := "SELECT id, allowance_type, amount FROM allowance WHERE allowance_type = $1"
	mock.ExpectQuery(SearchByTypeSql).WithArgs("personal").WillReturnRows(rowsPersonal)
	mock.ExpectQuery(SearchByTypeSql).WithArgs("donation").WillReturnRows(rowsDonation)
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

func TestDonation_get(t *testing.T) {
	t.Parallel()
	type fields struct {
		DB     *sql.DB
		amount float64
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{"Donation should get allowance correctly when input amount < max value", fields{mockCalculatorDb(t), 50000.00}, 50000.00},
		{"Donation should get allowance correctly when input amount > max value", fields{mockCalculatorDb(t), 110000.00}, 100000.00},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.fields.DB.Close()
			d := &Donation{
				DB:     tt.fields.DB,
				amount: tt.fields.amount,
			}
			if got := d.get(); got != tt.want {
				t.Errorf("Donation.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setDeductors(t *testing.T) {
	t.Parallel()
	mockDonationAmount := 100000.00

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
		{"Should return list of Deductor correctly when Allowance is not empty", args{[]Allowance{{AllowanceType: "donation", Amount: &mockDonationAmount}, {AllowanceType: "personal"}}, mockCalculatorDb(t)}, []Deductor{&Donation{amount: 100000.00, DB: mockCalculatorDb(t)}, &Personal{DB: mockCalculatorDb(t)}}},
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
		Wht         float64
		Deductors   []Deductor
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{"Should return sum of deduction = 60000 when allowance has only personal deduction", fields{500000.0, 0.0, []Deductor{&Personal{DB: mockCalculatorDb(t)}}}, 60000},
		{"Should return sum of deduction = 80000 when allowance has personal deduction and donation = 20000", fields{500000.0, 0.0, []Deductor{&Donation{amount: 20000.0, DB: mockCalculatorDb(t)}, &Personal{DB: mockCalculatorDb(t)}}}, 80000},
		{"Should return sum of deduction = 160000 when allowance has personal deduction and donation = 1000000", fields{500000.0, 0.0, []Deductor{&Donation{amount: 1000000.0, DB: mockCalculatorDb(t)}, &Personal{DB: mockCalculatorDb(t)}}}, 160000},
		{"Should return sum of deduction = 0 when Deduction is empty", fields{500000.0, 0.0, make([]Deductor, 0)}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Calculator{
				TotalIncome: tt.fields.TotalIncome,
				Wht:         tt.fields.Wht,
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
		want []TaxLevel
	}{
		{"Should return tax information for in first tier correctly", args{income: 100000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 0}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax information for end of first tier correctly", args{income: 150000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 0}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax information for in second tier correctly", args{income: 300000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 15000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax information for end of second tier correctly", args{income: 500000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 35000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax information for in third tier correctly", args{income: 750000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 35000}, {"500,001-1,000,000", 37500}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax information for end of third tier correctly", args{income: 1000000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 35000}, {"500,001-1,000,000", 75000}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax information for in fourth tier correctly", args{income: 1500000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 35000}, {"500,001-1,000,000", 75000}, {"1,000,001-2,000,000", 100000}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax information for end of fourth tier correctly", args{income: 2000000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 35000}, {"500,001-1,000,000", 75000}, {"1,000,001-2,000,000", 200000}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax information for fifth tier correctly", args{income: 2500000}, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 35000}, {"500,001-1,000,000", 75000}, {"1,000,001-2,000,000", 200000}, {"2,000,001 ขึ้นไป", 175000}}},
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
		Wht         float64
		Deductors   []Deductor
	}
	tests := []struct {
		name         string
		fields       fields
		want         float64
		wantTaxLevel []TaxLevel
	}{
		{"Should return tax = 29000 when income = 500000 and allowance has only personal deduction", fields{500000.0, 0.0, []Deductor{&Personal{DB: mockCalculatorDb(t)}}}, 29000, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 29000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax = 4000 when income = 500000, wht = 25000 and allowance has only personal deduction", fields{500000.0, 25000.0, []Deductor{&Personal{DB: mockCalculatorDb(t)}}}, 4000, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 29000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
		{"Should return tax = 16500 when income = 500000, wht = 25000 and allowance has personal deduction donation = 20000", fields{500000.0, 2500.0, []Deductor{&Donation{amount: 200000.0, DB: mockCalculatorDb(t)}, &Personal{DB: mockCalculatorDb(t)}}}, 16500, []TaxLevel{{"0-150,000", 0}, {"150,001-500,000", 19000}, {"500,001-1,000,000", 0}, {"1,000,001-2,000,000", 0}, {"2,000,001 ขึ้นไป", 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Calculator{
				TotalIncome: tt.fields.TotalIncome,
				Wht:         tt.fields.Wht,
				Deductors:   tt.fields.Deductors,
			}
			got, taxLevel := c.calculate()
			if tt.want != got {
				t.Errorf("Calculator.calculate() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(taxLevel, tt.wantTaxLevel) {
				t.Errorf("Calculator.calculate() = %v, want %v", taxLevel, tt.wantTaxLevel)
			}
		})
	}
}
