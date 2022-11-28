package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type CIK struct {
	CIK_Str int    `json:"cik_str"`
	Ticker  string `json:"ticker"`
	Title   string `json:"title"`
}

type Financials struct {
	Cik        int    `json:"cik"`
	EntityName string `json:"entityName"`
	Facts      struct {
		UsGaap struct {
			Revenues Revenues `json:"revenues"`
		} `json:"us-gaap"`
	} `json:"facts"`
}

type Revenues struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Units       struct {
		USD []Revenuez `json:"USD"`
	} `json:"units"`
}

type Revenuez struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Val   int    `json:"val"`
	Accn  string `json:"accn"`
	Fy    int    `json:"fy"`
	Fp    string `json:"fp"`
	Form  string `json:"form"`
	Filed string `json:"filed"`
}

var CIK_list []CIK

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/loadCIK", loadCIK).Methods("GET")
	// router.HandleFunc("/matchCIK", matchCIK).Methods("GET")
	router.HandleFunc("/financials/{ticker}", getFinancials).Methods("GET")

	loadCIK_local()
	fmt.Println("server at 8080")

	// connect db
	// connect()

	log.Fatal(http.ListenAndServe("localhost:8080", router))
}

func loadCIK_local() {
	jsonFile, err := os.Open("cik_list.json")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("successfully opened cik_list.json")
	defer jsonFile.Close()

	// var jsonData map[string]CIK
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal([]byte(byteValue), &CIK_list)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("loaded json")

}

func loadCIK(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://www.sec.gov/files/company_tickers.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	// read body resp
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// var jsonData CIK_response
	var jsonData map[string]CIK
	// var CIK_list []CIK

	err = json.Unmarshal([]byte(body), &jsonData)

	if err != nil {
		log.Fatalln(err)
	}

	for _, company := range jsonData {
		CIK_list = append(CIK_list, company)
	}

	json.NewEncoder(w).Encode(CIK_list)
}

func matchCIK(ticker string) CIK {
	// ticker := "AAPL"

	var cik CIK
	for _, company := range CIK_list {
		if company.Ticker == strings.ToUpper(ticker) {
			cik.CIK_Str = company.CIK_Str
			cik.Ticker = company.Ticker
			cik.Title = company.Title
		}
	}
	return cik
}

func getFinancials(w http.ResponseWriter, r *http.Request) {
	ticker := mux.Vars(r)["ticker"]
	fmt.Printf("ticker is: %v \n", ticker)
	company := matchCIK(ticker)
	fmt.Printf("CIK is %v \n", company.CIK_Str)

	cik_length := len(strconv.Itoa(company.CIK_Str))
	fmt.Println(cik_length)

	cik_padded := fmt.Sprintf("%010v", company.CIK_Str)
	fmt.Println(cik_padded)

	url := fmt.Sprintf("https://data.sec.gov/api/xbrl/companyfacts/CIK%v.json", cik_padded)
	fmt.Println(url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println("0")
		log.Fatalln(err)
	}

	req.Header.Add("User-Agent", "kalanco kalan kalan1@ualberta.ca")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("1")
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	// read body
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("2")
		log.Fatalln(err)
	}

	// var jsonData CIK_response
	// var jsonData map[string]interface{}
	var companyRevenue Financials
	// var CIK_list []CIK

	err = json.Unmarshal([]byte(body), &companyRevenue)

	if err != nil {
		fmt.Println("3")
		log.Fatalln(err)
	}

	fmt.Printf("company: %v", companyRevenue.EntityName)
	fmt.Printf("Revenue: %v", companyRevenue.Facts.UsGaap.Revenues)

	json.NewEncoder(w).Encode(companyRevenue)
}

// 10 digits
// ##########
