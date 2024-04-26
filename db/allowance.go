package db

import (
	"database/sql"
	"log"
)

type Allowance struct {
	Id            int     `json:"id"`
	AllowanceType string  `json:"allowanceType"`
	Amount        float64 `json:"amount"`
}

func getAllowanceDefaultValues() []Allowance {
	return []Allowance{
		{AllowanceType: "personal", Amount: 60000.00},
		{AllowanceType: "donation", Amount: 100000.00},
	}
}

func createAllowanceTable(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS allowance ( id SERIAL PRIMARY KEY, allowance_type TEXT UNIQUE, amount float)`); err != nil {
		return err
	}
	return nil
}

func (a *Allowance) Insert(db *sql.DB) error {
	if _, err := db.Exec("INSERT INTO allowance (allowance_type, amount) VALUES ($1,$2)", a.AllowanceType, a.Amount); err != nil {
		return err
	}
	return nil
}

func (a *Allowance) UpdateByType(db *sql.DB) error {
	if _, err := db.Exec("UPDATE allowance SET amount = $1 WHERE allowance_type = $2", a.Amount, a.AllowanceType); err != nil {
		return err
	}
	return nil
}

func (a *Allowance) SearchByType(db *sql.DB) Allowance {
	result := Allowance{}
	selectAllowance := "SELECT id, allowance_type, amount FROM allowance WHERE allowance_type = $1"
	rows, err := db.Query(selectAllowance, a.AllowanceType)
	if err != nil {
		log.Fatal("can't select allowance list", err)
	}

	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&result.Id, &result.AllowanceType, &result.Amount); err != nil {
			log.Fatal("can't Scan row into variable", err)
		}
	}

	return result
}

func SearchAllAllowance(db *sql.DB) []Allowance {
	results := make([]Allowance, 0)
	selectAllAllowance := "SELECT id, allowance_type, amount FROM allowance"
	rows, err := db.Query(selectAllAllowance)
	if err != nil {
		log.Fatal("can't select allowance list", err)
	}

	defer rows.Close()

	for rows.Next() {
		allowance := Allowance{}
		if err := rows.Scan(&allowance.Id, &allowance.AllowanceType, &allowance.Amount); err != nil {
			log.Fatal("can't Scan row into variable", err)
		}
		results = append(results, allowance)
	}
	return results
}
