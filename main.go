package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	kraken "github.com/beldur/kraken-go-api-client"
)

// init logger
var logErr = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
var logOk = log.New(os.Stdout, "OK: ", log.Ldate|log.Ltime)

func main() {

	// read config
	file, err := os.ReadFile("./ganiu.json")
	if err != nil {
		logErr.Println("Reading config:", err)
		return
	}

	// parse config
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		logErr.Println("Parsing config:", err)
		return
	}

	// connect to kraken api
	api := kraken.New(cfg.ApiData.Key, cfg.ApiData.Secret)

	// HANDLE BUSINESS
	logOk.Println("Ganiu is listening..")
	for {

		time.Sleep(time.Second * 30)

		// fetch open order
		orders, err := api.OpenOrders(make(map[string]string))
		if err != nil {
			logErr.Println("Fetching open orders:", err)
			continue
		}

		// if no open orders, all gucci
		if len(orders.Open) == 0 {
			logErr.Println("No open orders")
			break
		}

		// for now ganiu will operate with one open order
		if len(orders.Open) > 1 {
			logErr.Println("Ganiu can't handle multiple open orders")
			break
		}

		// take order id and type
		var orderId string
		var orderType string

		for txid, order := range orders.Open {
			orderId = txid
			orderType = order.Description.OrderType
		}

		// handle limit order
		if orderType == "limit" {
			continue
		}

		// get last price
		ticker, err := api.Ticker("XETHZUSD")
		if err != nil {
			logErr.Println("Getting ticker:", err)
			continue
		}

		lastPrice, err := strconv.ParseFloat(ticker.XETHZUSD.Close[0], 0)
		if err != nil {
			logErr.Println("Parsing last price:", err)
			continue
		}

		// if order is stop-loss and current price > entry price
		if orderType == "stop-loss" && lastPrice > cfg.OrderData.Entry {
			newOrder := NewOrder{
				Type:  "take-profit",
				Entry: cfg.OrderData.Take,
			}

			err = HandleOrder(api, orderId, orderType, &newOrder)
			if err != nil {
				logErr.Println(err)
			}
			continue
		}

		// if order is take-prfit and current price < entry price
		if orderType == "take-profit" && lastPrice < cfg.OrderData.Entry {
			newOrder := NewOrder{
				Type:  "stop-loss",
				Entry: cfg.OrderData.Stop,
			}

			err = HandleOrder(api, orderId, orderType, &newOrder)
			if err != nil {
				logErr.Println(err)
			}
			continue
		}

	}

}

// HANDLE ORDER FUNCTION
func HandleOrder(api *kraken.KrakenApi, orderId, orderType string, newOrder *NewOrder) error {
	// get balance
	balance, err := api.Balance()
	if err != nil {
		logErr.Println("Getting balance:", err)
		return err
	}

	// cancel order by orderId
	_, err = api.CancelOrder(orderId)
	if err != nil {
		logErr.Println("Cancelling order:", err)
		return err
	}
	logOk.Println(orderType + " order cancelled")

	// wait in cased the cancelation lags
	time.Sleep(time.Second * 1)

	// parse order type
	volume := fmt.Sprintf("%v", balance.XETH)
	args := map[string]string{
		"price": fmt.Sprintf("%v", newOrder.Entry),
	}

	_, err = api.AddOrder("XETHZUSD", "sell", newOrder.Type, volume, args)
	if err != nil {
		logErr.Println("Placing order:", err)
		return err
	}
	logOk.Println(newOrder.Type + " order placed")

	return nil
}
