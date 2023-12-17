package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type BankAccount struct {
	ID            string    `json:"id"`
	AccountNumber string    `json:"account_number,omitempty"`
	Status        string    `json:"status"`
	Deadline      time.Time `json:"deadline"`
}

var (
	accountsMutex sync.Mutex
	accounts      = make(map[string]*BankAccount)
	secretCode    = "async-service-secret-code"
)

func main() {
	http.HandleFunc("/generate-account", handleGenerateAccount)
	http.HandleFunc("/update-results", handleUpdateResults)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleGenerateAccount(w http.ResponseWriter, r *http.Request) {
	if !authorize(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var currency struct {
		CurrencyCode string `json:"currency_code"`
	}

	err := json.NewDecoder(r.Body).Decode(&currency)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	id := generateID()

	account := &BankAccount{
		ID:       id,
		Status:   "processing",
		Deadline: time.Now().Add(10 * time.Second),
	}

	account.AccountNumber = generateAccountNumber(currency.CurrencyCode)

	accountsMutex.Lock()
	accounts[id] = account
	accountsMutex.Unlock()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(account.AccountNumber))

	go asyncProcessor(account, currency.CurrencyCode)
}

func asyncProcessor(account *BankAccount, currencyCode string) {
	time.Sleep(2 * time.Second)

	account.Status = "completed"

	account.AccountNumber = generateAccountNumber(currencyCode)

	fmt.Printf("Bank Account ID: %s, Account Number: %s\n", account.ID, account.AccountNumber)

	accountsMutex.Lock()
	delete(accounts, account.ID)
	accountsMutex.Unlock()
}

func handleUpdateResults(w http.ResponseWriter, r *http.Request) {
	if !authorize(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var updateData struct {
		ID     string `json:"id"`
		Result string `json:"result"`
	}
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	accountsMutex.Lock()
	if account, ok := accounts[updateData.ID]; ok && account.Status == "completed" {
		account.Status = updateData.Result
		fmt.Printf("Result updated for Bank Account ID: %s, Status: %s\n", updateData.ID, updateData.Result)
	} else {
		fmt.Printf("Bank Account ID not found or account not completed: %s\n", updateData.ID)
	}
	accountsMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

func authorize(r *http.Request) bool {
	return r.Header.Get("Authorization") == secretCode
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateAccountNumber(currencyCode string) string {
	prefix := "40817"
	branchCode := "2003"

	controlDigit := generateControlDigit(prefix + currencyCode + branchCode)

	accountNumber := generateUniqueAccountNumber()

	return prefix + currencyCode + controlDigit + branchCode + accountNumber
}

func generateControlDigit(input string) string {
	var sum int

	for i, char := range input {
		digit := int(char - '0')
		weight := 2 - i%2
		sum += weight * digit
	}

	controlDigit := (10 - sum%10) % 10

	return fmt.Sprintf("%d", controlDigit)
}

func generateUniqueAccountNumber() string {
	return fmt.Sprintf("%07d", rand.Intn(10000000))
}
