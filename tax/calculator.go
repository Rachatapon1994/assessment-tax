package tax

import (
	"database/sql"
	"github.com/Rachatapon1994/assessment-tax/db"
)

var (
	PERSONAL = "personal"
)

type Deductor interface {
	get() float64
}
type Calculator struct {
	TotalIncome float64
	Deductors   []Deductor
}

type Personal struct {
	DB *sql.DB
}

func (p *Personal) get() float64 {
	return (&db.Allowance{AllowanceType: PERSONAL}).SearchByType(p.DB).Amount
}

func setDeductors(allowances []Allowance, DB *sql.DB) []Deductor {
	deductors := make([]Deductor, 0)
	for _, allowance := range allowances {
		switch allowance.AllowanceType {
		case PERSONAL:
			deductors = append(deductors, &Personal{DB: DB})
		}
	}
	return deductors
}

func (c *Calculator) sumDeduction() float64 {
	result := 0.0
	for _, deduction := range c.Deductors {
		result += deduction.get()
	}
	return result
}

func calculateTaxLevels(income float64) []float64 {
	result := make([]float64, 0)
	for _, taxLevel := range getLevels() {
		if income > taxLevel.endAmount {
			result = append(result, taxLevel.maxDeduction)
		} else {
			result = append(result, (income-taxLevel.startAmount+1)*(taxLevel.percentage/100))
			break
		}
	}
	return result
}

func (c *Calculator) calculate() float64 {
	result := 0.0
	for _, taxLevel := range calculateTaxLevels(c.TotalIncome - c.sumDeduction()) {
		result += taxLevel
	}
	return result
}
