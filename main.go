package main

import (
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type item struct {
	ShortDescription string `json:"ShortDescription"`
	Price            string `json:"Price"`
}

type receipt struct {
	Retailer     string `json:"Retailer"`
	PurchaseDate string `json:"PurchaseDate"`
	PurchaseTime string `json:"PurchaseTime"`
	Items        []item `json:"Items"`
	Total        string `json:"Total"`
}

var pointsMap = make(map[string]int)

func main() {
	router := gin.Default()
	router.POST("/receipts/process", postReceipts)
	router.GET("/receipts/:id/points", getPoints)
	router.Run("localhost:8080")
}

func clearString(str string) string {
	var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}

func countPoints(receipt receipt) int {
	points := 0

	//Points for retailer length
	points += len(clearString(receipt.Retailer))

	//Parse Date to get day and time
	dateTime := receipt.PurchaseDate + " " + receipt.PurchaseTime
	parseTime, err := time.Parse("2006-01-02 15:04", dateTime)
	if err != nil {
		panic(err)
	}

	//Points for odd day
	if parseTime.Day()%2 == 1 {
		points += 6
	}

	
	//Getting time checkpoints
	layout := "15:04"
	start, _ := time.Parse(layout, "14:00")
	check, _ := time.Parse(layout, receipt.PurchaseTime)
	end, _ := time.Parse(layout, "16:00")

	//Points for right time
	if start.Before(check) && end.After(check) {
		points += 10
	}

	//Points for item pairs
	points += (len(receipt.Items) * 2)

	//Points for short description
	for i := 0; i < len(receipt.Items); i++ {
		length := len(strings.TrimSpace(receipt.Items[i].ShortDescription))
		if length%3 == 0 {
			price, _ := strconv.ParseFloat(receipt.Items[i].Price, 64)
			points += int(math.Ceil(price * .2))
		}

	}

	//Points for .25 divisible
	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if math.Remainder(total, .25) == 0 {
		points += 25
	}

	//Points for no cents
	if math.Remainder(total, 1) == 0 {
		points += 50
	}

	return points
}

func getPoints(c *gin.Context) {
	id := c.Param("id")
	points := pointsMap[id]
	data := map[string]interface{}{
		"points": points,
	}

	c.IndentedJSON(http.StatusOK, data)
}

func postReceipts(c *gin.Context) {
	var newReceipt receipt

	if err := c.BindJSON(&newReceipt); err != nil {
		return
	}

	newID := uuid.New().String()

	data := map[string]interface{}{
		"id": newID,
	}

	c.IndentedJSON(http.StatusCreated, data)
	pointsMap[newID] = countPoints(newReceipt)

}
