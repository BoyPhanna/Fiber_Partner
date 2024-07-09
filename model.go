package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	Name        string  `json:"name"`
	Amount      float32 `json:"amount"`
	AccountId   uint    `json:"accountId"`
	AccountName string  `json:"accountName"`
}
type Account struct {
	AccountId int    `json:"accountId"`
	Name      string `json:"name"`
}

// getBooks retrieves all books
func getPayments(db *gorm.DB, c *fiber.Ctx) error {
	var payments []Payment
	db.Find(&payments)
	return c.JSON(payments)
}

// create new Payment
func createPayment(db *gorm.DB, c *fiber.Ctx) error {
	payment := new(Payment)
	if err := c.BodyParser(payment); err != nil {
		return err
	}
	if err := pay(int(payment.AccountId), payment.Amount); err != nil {
		return err
	}
	db.Create(&payment)
	return c.JSON(payment)
}
func getAccount(c *fiber.Ctx) error {
	var id, err = strconv.Atoi(c.Params("id"))
	if err != nil {
		return err
	}
	var account Account
	checkBeforPay(&account, id)
	return c.JSON(account)
}

func checkBeforPay(person *Account, id int) {
	response, err := http.Get("http://127.0.0.1:8080/account/checkbeforepay/" + fmt.Sprint(id))
	fmt.Println("Test===============")
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))

	if err := json.Unmarshal(body, &person); err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}
	fmt.Printf("Name: %s, Age: %d\n", person.Name, person.AccountId)
}

type Payload struct {
	AccountId int     `json:"accountId"`
	Amount    float32 `json:"amount"`
}

func pay(accountId int, amount float32) error {
	url := "http://127.0.0.1:8080/account/pay"

	// Create the payload
	payload := Payload{
		AccountId: accountId,
		Amount:    amount,
	}

	// Marshal the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
		return err
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request using http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read the response
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Request was successful")

	} else {
		fmt.Printf("Request failed with status: %s\n", resp.Status)
	}
	return nil
}
