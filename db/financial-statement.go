package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// func (company *company) CreateIncomeStatement() {
// 	fmt.Println("creating income statement...")
// }

// type company struct {
// 	name   string
// 	ticker string
// }

type CompanyInformation struct {
	Name          string
	Ticker        string
	Cik           string
	EntityName    string
	Revenues      []IncomeStatementEntry `json:"revenues"`
	CostOfRevenue []IncomeStatementEntry `json:"costOfRevenue"`
}

type IncomeStatementEntry struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Val   int    `json:"val"`
	Accn  string `json:"accn"`
	Fy    int    `json:"fy"`
	Fp    string `json:"fp"`
	Form  string `json:"form"`
	Filed string `json:"filed"`
	Frame string `json:"frame"`
}

// var companies = []company{{"apple", "aapl"}, {"microsoft", "msft"}}

var companies = []CompanyInformation{{"apple", "aapl", "", "", []IncomeStatementEntry{}, []IncomeStatementEntry{}}, {"microsoft", "msft", "", "", []IncomeStatementEntry{}, []IncomeStatementEntry{}}}

func assignCompany(name string) (CompanyInformation, error) {
	for _, entry := range companies {
		if entry.Name == name || entry.Ticker == name {
			return entry, nil
		}
	}
	return CompanyInformation{}, fmt.Errorf("company not found")
}

func (company *CompanyInformation) CreateCompanyTable(db *sql.DB) {
	sqlStatement := fmt.Sprintf(`
	CREATE TABLE %v (
			id SERIAL PRIMARY KEY,
			ticker TEXT,
			entity_name TEXT,
			cik TEXT,
			item_type TEXT,
			item_type_alias TEXT,
			start_date TEXT,
			end_date TEXT,
			val BIGINT,
			accn TEXT,
			fy INT,
			fp TEXT,
			form TEXT,
			filed TEXT,
			frame TEXT
		  );
		`, company.Ticker)

	// create table
	_, err := db.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}
}

func (company *CompanyInformation) CreateCompanyFinancials(db *sql.DB) {
	sqlStatement := `
	SELECT 
		j->>'cik' AS cik,
		j->>'entityName' AS name,
		j->'facts'->'us-gaap'->'RevenueFromContractWithCustomerExcludingAssessedTax'->'units'->>'USD'  AS revenues,
		j->'facts'->'us-gaap'->'CostOfGoodsAndServicesSold' -> 'units' ->> 'USD' AS costOfRevenue
	from t;
	`
	var p, q []byte

	row := db.QueryRow(sqlStatement)

	switch err := row.Scan(&company.Cik, &company.EntityName, &p, &q); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		fmt.Printf("%+v\n", company)
	default:
		panic(err)
	}

	err := json.Unmarshal(p, &company.Revenues)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(q, &company.CostOfRevenue)
	if err != nil {
		panic(err)
	}

	// drop table and start fresh
	_, err = db.Exec(fmt.Sprintf("DROP TABLE %v;", company.Ticker))
	if err != nil {
		panic(err)
	}

	sqlStatement = fmt.Sprintf(`
	CREATE TABLE %v (
		id SERIAL PRIMARY KEY,
		ticker TEXT,
		entity_name TEXT,
		cik TEXT,
		item_type TEXT,
		item_type_alias TEXT,
		start_date TEXT,
		end_date TEXT,
		val BIGINT,
		accn TEXT,
		fy INT,
		fp TEXT,
		form TEXT,
		filed TEXT,
		frame TEXT,
		modified TEXT
	);
`, company.Ticker)

	_, err = db.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%+v\n", aapl)

	insertStatement := fmt.Sprintf(`
	INSERT INTO %v (ticker, entity_name, cik, item_type, item_type_alias, start_date, end_date, val, accn, fy, fp, form, filed, frame, modified)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, company.Ticker)

	for index, entry := range company.Revenues {

		_, err = db.Exec(insertStatement, company.Ticker, company.EntityName, company.Cik, "RevenueFromContractWithCustomerExcludingAssessedTax", "revenue", entry.Start, entry.End, entry.Val, entry.Accn, entry.Fy, entry.Fp, entry.Form, entry.Filed, entry.Frame, "false")
		if err != nil {
			panic(err)
		}

		// Q4 entries are not available, have to create on our own
		if entry.Fp == "FY" && index >= 5 {
			var Q4Rev int
			// look behind to find previous Q3FY TTM to get Q4. Q4 = FYTTM - Q3TTM
			prevEntries := index - 5
			for i := prevEntries; i <= index; i++ {
				if company.Revenues[i].Fp == "Q3" && company.Revenues[i].Start == entry.Start {
					Q4Rev = entry.Val - company.Revenues[i].Val
				}
			}

			if Q4Rev != 0 {
				_, err = db.Exec(insertStatement, company.Ticker, company.EntityName, company.Cik, "RevenueFromContractWithCustomerExcludingAssessedTax", "revenue", entry.Start, entry.End, Q4Rev, entry.Accn, entry.Fy, "Q4", entry.Form, entry.Filed, entry.Frame, "true")
				if err != nil {
					panic(err)
				}
			}

		}
	}

	for index, entry := range company.CostOfRevenue {
		_, err = db.Exec(insertStatement, company.Ticker, company.EntityName, company.Cik, "CostOfGoodsAndServicesSold", "cost_of_revenue", entry.Start, entry.End, entry.Val, entry.Accn, entry.Fy, entry.Fp, entry.Form, entry.Filed, entry.Frame, "false")
		if err != nil {
			panic(err)
		}

		// Q4 entries are not available, have to create on our own
		if entry.Fp == "FY" && index >= 5 {
			var Q4COGS int
			// look behind to find previous Q3FY TTM to get Q4. Q4 = FYTTM - Q3TTM
			prevEntries := index - 5
			for i := prevEntries; i <= index; i++ {
				if company.CostOfRevenue[i].Fp == "Q3" && company.CostOfRevenue[i].Start == entry.Start {
					Q4COGS = entry.Val - company.CostOfRevenue[i].Val
				}
			}

			if Q4COGS != 0 {
				_, err = db.Exec(insertStatement, company.Ticker, company.EntityName, company.Cik, "CostOfGoodsAndServicesSold", "cost_of_revenue", entry.Start, entry.End, Q4COGS, entry.Accn, entry.Fy, "Q4", entry.Form, entry.Filed, entry.Frame, "true")
				if err != nil {
					panic(err)
				}
			}

		}
	}

}

// _, err = db.Exec(insertStatement, "aapl", aapl.EntityName, aapl.Cik, "RevenueFromContractWithCustomerExcludingAssessedTax", "revenue", aapl.Revenues[0].Start, aapl.Revenues[0].End, aapl.Revenues[0].Val, aapl.Revenues[0].Accn, aapl.Revenues[0].Fy, aapl.Revenues[0].Fp, aapl.Revenues[0].Form, aapl.Revenues[0].Filed)
// 	if err != nil {
// 		panic(err)
// 	}

// {
// 	"start": "2021-09-26",
// 	"end": "2022-09-24",
// 	"val": 394328000000,
// 	"accn": "0000320193-22-000108",
// 	"fy": 2022,
// 	"fp": "FY",
// 	"form": "10-K",
// 	"filed": "2022-10-28",
// 	"frame": "CY2022"
//   }

// CREATE TABLE aapl (
// 	id SERIAL PRIMARY KEY,
// 	ticker TEXT,
// 	entity_name TEXT,
// 	cik TEXT,
// 	item_type TEXT,
// 	item_type_alias TEXT,
// 	start_date TEXT,
// 	end_date TEXT,
// 	val BIGINT,
// 	accn TEXT,
// 	fy INT,
// 	fp TEXT,
// 	form TEXT,
// 	filed TEXT,
// 	frame TEXT
//   );

// query for quarterly revenue
// SELECT *
// FROM (select
// 	  * ,
// 	 (end_date::DATE) - (start_date::DATE) AS date_difference
// 	 from aapl
// 	 WHERE (end_date::DATE) - (start_date::DATE) = 90 OR (end_date::DATE) - (start_date::DATE) > 360) s
// WHERE
// val != 0
// AND
// fp != 'FY';

// query for annual revenue
// SELECT *
// FROM (select
// 	  * ,
// 	 (end_date::DATE) - (start_date::DATE) AS date_difference
// 	 from aapl
// 	 WHERE (end_date::DATE) - (start_date::DATE) > 360) s
// WHERE
// frame <> ''
// AND
// fp = 'FY';
