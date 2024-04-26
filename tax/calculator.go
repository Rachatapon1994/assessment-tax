package tax

import (
	"database/sql"

	"github.com/Rachatapon1994/assessment-tax/db"
)

var (
	PERSONAL = "personal"
	DONATION = "donation"
)

type Deductor interface {
	get() float64
}
type Calculator struct {
	TotalIncome float64
	Wht         float64
	Deductors   []Deductor
}

type Personal struct {
	DB *sql.DB
}

type Donation struct {
	DB     *sql.DB
	amount float64
}

func (p *Personal) get() float64 {
	return (&db.Allowance{AllowanceType: PERSONAL}).SearchByType(p.DB).Amount
}

func (d *Donation) get() float64 {
	maximumDonationAmount := (&db.Allowance{AllowanceType: DONATION}).SearchByType(d.DB).Amount
	if d.amount > maximumDonationAmount {
		return maximumDonationAmount
	}
	return d.amount
}

func setDeductors(allowances []Allowance, DB *sql.DB) []Deductor {
	deductors := make([]Deductor, 0)
	for _, allowance := range allowances {
		switch allowance.AllowanceType {
		case DONATION:
			deductors = append(deductors, &Donation{amount: *allowance.Amount, DB: DB})
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
	return result - c.Wht
}
