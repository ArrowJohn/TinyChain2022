package General

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// Response is the http response we send back to the client
type Response struct {
	Code    int    `json:"code"`
	Comment string `json:"comment"`
}

func CurrentTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func FindIndex(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func CalculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func CalculateTransHash(transactions []Transaction) string {
	record := ""
	for _, tran := range transactions {
		record = record + tran.Signature
	}
	return record
}

func CalculateTranHash(transaction Transaction) string {
	str := CalculateHash(transaction.Date + transaction.ID + transaction.From + transaction.To +
		string(rune(transaction.Value)) + transaction.Signature)
	return CalculateHash(str)
}

func ReadInput(input chan<- string) {
	for {
		var u string
		_, err := fmt.Scanf("%s\n", &u)
		if err != nil {
			panic(err)
		}
		input <- u
	}
}

func FormatResponse(code int, comment string) Response {
	response := Response{}
	response = Response{code, comment}
	return response
}

func JudgeInDays(timestamp string, days int) bool {
	currentTime := time.Now()
	oldTime := currentTime.AddDate(0, 0, -1*days)
	transTime, _ := strconv.ParseInt(timestamp, 10, 64)
	wantTime := time.Unix(transTime, 0)
	return wantTime.After(oldTime)
}
