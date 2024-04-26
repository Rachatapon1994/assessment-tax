package db

import (
	"github.com/DATA-DOG/go-sqlmock"
	"testing"

	_ "github.com/lib/pq"
)

func TestInitDB(t *testing.T) {
	t.Parallel()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	createTableSql := "CREATE TABLE IF NOT EXISTS allowance ( id SERIAL PRIMARY KEY, allowance_type TEXT UNIQUE, amount float)"
	insertAllowanceSql := "INSERT INTO allowance (allowance_type, amount) VALUES ($1,$2)"
	SearchByTypeSql := "SELECT id, allowance_type, amount FROM allowance WHERE allowance_type = $1"
	searchAllAllowanceSql := "SELECT id, allowance_type, amount FROM allowance"
	rowsAll := mock.NewRows([]string{"id", "allowance_type", "amount"}).
		AddRow(1, "personal", 60000.00).
		AddRow(2, "donation", 100000.00)

	mock.ExpectExec(createTableSql).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(SearchByTypeSql).WithArgs("personal").WillReturnRows(mock.NewRows([]string{"id", "allowance_type", "amount"}))
	mock.ExpectExec(insertAllowanceSql).WithArgs("personal", 60000.00).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(SearchByTypeSql).WithArgs("donation").WillReturnRows(mock.NewRows([]string{"id", "allowance_type", "amount"}))
	mock.ExpectExec(insertAllowanceSql).WithArgs("donation", 100000.00).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(searchAllAllowanceSql).WillReturnRows(rowsAll)

	t.Run("Should run dbPreparation correctly", func(t *testing.T) {
		dbPreparation(db)
	})
}
