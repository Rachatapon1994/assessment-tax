package tax

import (
	"database/sql"
	"github.com/Rachatapon1994/assessment-tax/db"
	"github.com/shopspring/decimal"
)

var (
	PERSONAL = "personal"
	DONATION = "donation"
	KRECEIPT = "k-receipt"
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

type KReceipt struct {
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

func (d *KReceipt) get() float64 {
	maximumDonationAmount := (&db.Allowance{AllowanceType: KRECEIPT}).SearchByType(d.DB).Amount
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
		case KRECEIPT:
			deductors = append(deductors, &KReceipt{amount: *allowance.Amount, DB: DB})
		}
	}
	return deductors
}

func (c *Calculator) sumDeduction() float64 {
	result := 0.0
	for _, deduction := range c.Deductors {
		deductionValue, _ := decimal.NewFromFloat(deduction.get()).Add(decimal.NewFromFloat(result)).Float64()
		result = deductionValue
	}
	return result
}

func calculateTaxLevels(income float64) []TaxLevel {
	result := make([]TaxLevel, 0)
	passLastTaxLevel := false
	for _, taxLevel := range getLevels() {
		if income > taxLevel.EndAmount {
			result = append(result, TaxLevel{Tax: taxLevel.MaxDeduction, Level: taxLevel.Name})
		} else {
			if !passLastTaxLevel {
				differenceValue := decimal.NewFromFloat(income).Sub(decimal.NewFromFloat(taxLevel.StartAmount)).Add(decimal.NewFromFloat(1))
				percentageValue := decimal.NewFromFloat(taxLevel.Percentage).Div(decimal.NewFromFloat(100))
				tax, _ := differenceValue.Mul(percentageValue).Float64()
				result = append(result, TaxLevel{Tax: tax, Level: taxLevel.Name})
				passLastTaxLevel = true
			} else {
				result = append(result, TaxLevel{Tax: 0, Level: taxLevel.Name})
			}
		}
	}
	return result
}

func (c *Calculator) calculate() (float64, []TaxLevel) {
	result := 0.0
	taxLevels := calculateTaxLevels(c.TotalIncome - c.sumDeduction())
	for _, taxLevel := range taxLevels {
		result += taxLevel.Tax
	}
	tax, _ := decimal.NewFromFloat(result).Sub(decimal.NewFromFloat(c.Wht)).Float64()
	return tax, taxLevels
}
