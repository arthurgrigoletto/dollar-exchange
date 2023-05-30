package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type EconomiaAwesomeAPI struct {
	USDBRL USDBRL
}

type USDBRL struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Exchange struct {
	ID    int `gorm:"primaryKey"`
	Name  string
	Value string
	gorm.Model
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", mux)
}

func handler(w http.ResponseWriter, r *http.Request) {
	exchange, err := GetExchange(r.Context())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = CreateExchange(r.Context(), exchange)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(exchange)
}

func GetExchange(ctx context.Context) (*USDBRL, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	var data EconomiaAwesomeAPI
	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, err
	}

	return &data.USDBRL, nil
}

func CreateExchange(ctx context.Context, usdbrl *USDBRL) (*Exchange, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Nanosecond)
	defer cancel()

	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Exchange{})

	exchange := Exchange{
		Name:  usdbrl.Name,
		Value: usdbrl.Bid,
	}

	err = db.WithContext(ctx).Create(&exchange).Error

	if err != nil {
		fmt.Println("Timeout: Não foi possível criar registro no banco de dados")
	}

	return &exchange, nil
}
