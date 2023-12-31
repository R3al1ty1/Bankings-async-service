package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const secretKey = "secret-async-key"

type AccountApplication struct {
	AccountID     int64  `json:"account_id"`
	ApplicationID int64  `json:"application_id"`
	Number        int64  `json:"number"`
	Currency      string `json:"currency"`
}

func main() {
	r := gin.Default()

	r.POST("/get_number", func(c *gin.Context) {
		var account AccountApplication

		if err := c.ShouldBindJSON(&account); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		apiKey := c.GetHeader("Authorization")
		if apiKey != secretKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid secret key"})
			return
		}

		go func() {
			time.Sleep(5 * time.Second)
			SendStatus(account)
		}()

		c.JSON(http.StatusOK, gin.H{"message": "Status update initiated"})
	})

	r.Run(":8080")
}

func SendStatus(accApp AccountApplication) bool {
	fmt.Println(accApp.Number)
	accApp.Number = generateAccountNumber(accApp.Currency)
	fmt.Println(accApp.Number)
	url := "http://localhost:8000/api/apps_accs/" + fmt.Sprint(accApp.AccountID) + "/" + fmt.Sprint(accApp.ApplicationID) + "/put/"
	response, err := performPUTRequest(url, accApp)
	if err != nil {
		fmt.Println("Error sending status:", err)
		return false
	}

	if response.StatusCode == http.StatusOK {
		fmt.Println("Status sent successfully for pk:", accApp.ApplicationID)
		return true
	} else {
		fmt.Println("Failed to process PUT request")
		return false
	}
}

func generateAccountNumber(currencyCode string) int64 {
	if currencyCode == "" {
		fmt.Println("Currency code is empty")
		return 0
	}

	prefix := "40817"
	branchCode := "001"

	currencyInt, err := strconv.Atoi(currencyCode)
	if err != nil {
		fmt.Println("Error converting currency code to integer:", err)
		return 0
	}

	controlDigit := generateControlDigit(prefix + strconv.Itoa(currencyInt) + branchCode)

	accountNumber := generateUniqueAccountNumber()

	result, err := strconv.ParseInt(prefix+strconv.Itoa(currencyInt)+controlDigit+branchCode+accountNumber, 10, 64)
	if err != nil {
		fmt.Println("Error converting account number to int64:", err)
		return 0
	}

	return result
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
	return fmt.Sprintf("%07d", rand.Intn(1000000))
}

func performPUTRequest(url string, data AccountApplication) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}
