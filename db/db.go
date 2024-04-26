package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func dbPreparation(db *sql.DB) {
	createAllowanceTable(db)

	for _, aw := range getAllowanceDefaultValues() {
		allowance := (&Allowance{AllowanceType: aw.AllowanceType}).SearchByType(db)
		if (Allowance{}) == allowance {
			allowance := &Allowance{AllowanceType: aw.AllowanceType, Amount: aw.Amount}
			if err := allowance.insert(db); err != nil {
				log.Fatal("can't initialize data", err)
			}
		}
	}

	allowances := SearchAllAllowance(db)
	fmt.Println(`Starting Tax calculate application with default fields as below: `)
	for _, allowance := range allowances {
		fmt.Printf("ID: %d, TYPE: %v, AMOUNT: %0.2f\n", allowance.Id, allowance.AllowanceType, allowance.Amount)
	}
}

func InitDB() *sql.DB {
	var err error
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to database error", err)
	}
	dbPreparation(db)
	return db
}
