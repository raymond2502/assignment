package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

var Refval = "01-02-2006T15:04:05Z"

const TransactionError = "Future Transaction Date"

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/transaction", PostTransactions).Methods("POST")
	r.HandleFunc("/statistics", GetStatistics).Methods("GET")
	r.HandleFunc("/delete", DeleteData).Methods("DELETE")
	//http.Handle("/", r)
	fmt.Println("server started")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalln("There's an error with the server,", err)
	}

}
func PostTransactions(w http.ResponseWriter, r *http.Request) {
	var t Transaction
	var resp Response
	var transactions []Transaction
	json.NewDecoder(r.Body).Decode(&t)
	if _, ok := os.Stat("transaction.json"); ok == nil {
		fmt.Println("file Found")
		file, err := os.ReadFile("transaction.json")
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(file, &transactions)
	} else {
		fmt.Println("file not found. File created")
		_, err := os.Create("transaction.json")
		if err != nil {
			fmt.Println(err)
		}
	}
	if ok, err := compTime(t.Timestamp, 60); ok {
		transactions = append(transactions, t)
		data, err := json.MarshalIndent(transactions, "", " ")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("data:", string(data))
		err = os.WriteFile("transaction.json", data, 7777)
		if err != nil {
			fmt.Println(err)
		}
		resp.StatusCode = 201
		resp.Message = "Success"

	} else if err != nil {
		resp.StatusCode = 422
		if err.Error() == TransactionError {
			resp.Message = TransactionError
		} else {
			resp.Message = "Timestamp error"
		}
		fmt.Println(err)

	} else {
		resp.StatusCode = 204
		resp.Message = "Transaction Older than 60s"
	}
	json.NewEncoder(w).Encode(resp)
}
func GetStatistics(w http.ResponseWriter, r *http.Request) {
	var transaction []Transaction
	var stats Statistics
	var sum, max, min float64
	var count int
	json.NewDecoder(r.Body).Decode(&transaction)
	file, err := os.ReadFile("transaction.json")
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(file, &transaction)
	for i, trans := range transaction {
		//fmt.Println("transaction:", i, trans.Amount)
		if i == 0 {
			max = trans.Amount
			min = trans.Amount
		}
		if ok, _ := compTime(trans.Timestamp, 60); !ok {
			continue
		}
		if trans.Amount > max {
			max = trans.Amount
		}
		if trans.Amount < min {
			min = trans.Amount
		}
		sum = sum + trans.Amount
		count++
	}
	stats.Max = max
	stats.Min = min
	stats.Count = count
	stats.Sum = sum
	stats.Avg = sum / float64(count)

	fmt.Println("check Stats:", stats)
	json.NewEncoder(w).Encode(stats)
}
func DeleteData(w http.ResponseWriter, r *http.Request) {
	var resp Response
	err := os.Truncate("transaction.json", 0)
	if err != nil {
		fmt.Println(err)
	}
	resp.StatusCode = 204
	resp.Message = "All Transactions Deleted"
	json.NewEncoder(w).Encode(resp)
}

func compTime(timestamp string, dur float64) (bool, error) {

	t1, err := time.Parse(Refval, timestamp)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	t21, err := time.Parse(Refval, time.Now().Format(Refval))
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if t21.Before(t1) {
		return false, fmt.Errorf(TransactionError)
	}
	durationDiff := t21.Sub(t1)
	if durationDiff.Seconds() > dur {
		return false, err
	}
	return true, nil
}

type Transaction struct {
	Amount    float64 `json:"amount"`
	Timestamp string  `json:"timestamp"`
}

type Response struct {
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message,omitempty"`
	Output     string `json:"output,omitempty"`
}
type Statistics struct {
	Sum   float64 `json:"sum"`
	Avg   float64 `json:"avg"`
	Max   float64 `json:"max"`
	Min   float64 `json:"min"`
	Count int     `json:"count"`
}
