package db

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"
)

func mockAllowanceDb(t *testing.T) *sql.DB {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	mock.MatchExpectationsInOrder(false)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rowsPersonal := mock.NewRows([]string{"id", "allowance_type", "amount"}).
		AddRow(1, "personal", 60000.00)
	rowsAll := mock.NewRows([]string{"id", "allowance_type", "amount"}).
		AddRow(1, "personal", 60000.00)
	insertAllowanceSql := "INSERT INTO allowance (allowance_type, amount) VALUES ($1,$2)"
	createTableSql := "CREATE TABLE IF NOT EXISTS allowance ( id SERIAL PRIMARY KEY, allowance_type TEXT UNIQUE, amount float)"

	SearchByTypeSql := "SELECT id, allowance_type, amount FROM allowance WHERE allowance_type = $1"
	searchAllAllowanceSql := "SELECT id, allowance_type, amount FROM allowance"
	mock.ExpectQuery(SearchByTypeSql).WithArgs("personal").WillReturnRows(rowsPersonal)
	mock.ExpectQuery(SearchByTypeSql).WithArgs("insurance").WillReturnRows(mock.NewRows([]string{"id", "allowance_type", "amount"}))
	mock.ExpectQuery(searchAllAllowanceSql).WillReturnRows(rowsAll)
	mock.ExpectExec(insertAllowanceSql).WithArgs("personal", 60000.00).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(insertAllowanceSql).WithArgs("mockError", 60000.00).WillReturnError(sql.ErrConnDone)
	mock.ExpectExec(createTableSql).WillReturnResult(sqlmock.NewResult(0, 0))
	return db
}

func Test_getAllowanceDefaultValues(t *testing.T) {
	tests := []struct {
		name string
		want []Allowance
	}{
		{"Should return list of allowance correctly", []Allowance{{AllowanceType: "personal", Amount: 60000}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAllowanceDefaultValues(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllowanceDefaultValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchAllAllowance(t *testing.T) {
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
		want []Allowance
	}{
		{"Should return all allowances correctly", args{db: mockAllowanceDb(t)}, []Allowance{{Id: 1, AllowanceType: "personal", Amount: 60000}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchAllAllowance(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchAllAllowance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllowance_SearchByType(t *testing.T) {
	type fields struct {
		Id            int
		AllowanceType string
		Amount        float64
	}
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Allowance
	}{
		{"Should return empty allowance for any type that does not exist in database", fields{AllowanceType: "insurance"}, args{db: mockAllowanceDb(t)}, Allowance{}},
		{"Should return allowance for 'personal' type correctly", fields{AllowanceType: "personal"}, args{db: mockAllowanceDb(t)}, Allowance{Id: 1, AllowanceType: "personal", Amount: 60000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Allowance{
				Id:            tt.fields.Id,
				AllowanceType: tt.fields.AllowanceType,
				Amount:        tt.fields.Amount,
			}
			if got := a.SearchByType(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Allowance.SearchByType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllowance_insert(t *testing.T) {
	type fields struct {
		Id            int
		AllowanceType string
		Amount        float64
	}
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   error
	}{
		{"Should return nil when inserting allowance successfully", fields{AllowanceType: "personal" +
			"", Amount: 60000.00}, args{db: mockAllowanceDb(t)}, nil},
		{"Should return error when inserting allowance unsuccessfully", fields{AllowanceType: "mockError", Amount: 60000.00}, args{db: mockAllowanceDb(t)}, sql.ErrConnDone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Allowance{
				Id:            tt.fields.Id,
				AllowanceType: tt.fields.AllowanceType,
				Amount:        tt.fields.Amount,
			}
			if got := a.insert(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Allowance.insert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllowance_createAllowanceTable(t *testing.T) {
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{"Should return nil when creating allowance table successfully", args{db: mockAllowanceDb(t)}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createAllowanceTable(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Allowance.createAllowanceTable() = %v, want %v", got, tt.want)
			}
		})
	}
}
