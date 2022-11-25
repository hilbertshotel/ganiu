package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	krakenapi "github.com/beldur/kraken-go-api-client"
)

func main() {

	// init logger
	log := log.New(os.Stdout, "ERROR: ", log.Lshortfile)

	// read config
	file, err := os.ReadFile("./ganiu.json")
	if err != nil {
		log.Println("Reading config:", err)
		return
	}

	// parse config
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Println("Parsing config:", err)
		return
	}

	// connect to kraken api
	api := krakenapi.New(cfg.ApiData.Key, cfg.ApiData.Secret)

	// HANDLE BUSINESS
	for {

		time.Sleep(time.Second * 15)

		// fetch open order
		orders, err := api.OpenOrders(make(map[string]string))
		if err != nil {
			log.Println("Fetching open orders:", err)
			continue
		}

		// if no open orders, all gucci
		if len(orders.Open) == 0 {
			log.Println("No open orders")
			break
		}

		// for now ganiu will operate with one open order
		if len(orders.Open) > 1 {
			log.Println("Ganiu can't handle multiple open orders")
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
		ticker, err := api.Ticker(cfg.Pair)
		if err != nil {
			log.Println("Getting ticker:", err)
			continue
		}
		lastPrice := ticker.XETHZUSD.Close[0]

		// parse last price into float
		price, err := strconv.ParseFloat(lastPrice, 0)
		if err != nil {
			log.Println("Parsing last price:", err)
			continue
		}

		// if order is stop-loss and current price > entry price
		if orderType == "stop-loss" && price > cfg.OrderData.Entry {

			// get balance
			balance, err := api.Balance()
			if err != nil {
				log.Println("Getting balance:", err)
				continue
			}

			// cancel stop-loss order
			_, err = api.CancelOrder(orderId)
			if err != nil {
				log.Println("Cancelling stop-loss order:", err)
				continue
			}
			fmt.Println("Stop-loss canceled")

			// place take-profit order
			takePrice := fmt.Sprintf("%v", cfg.OrderData.Take)
			volume := fmt.Sprintf("%v", balance.XETH)
			args := map[string]string{"price": takePrice}
			_, err = api.AddOrder(cfg.Pair, "sell", "take-profit", volume, args)
			if err != nil {
				log.Println("Adding order:", err)
				continue
			}
			fmt.Println("Take-profit placed")

			continue
		}

		// if order is take-prfit and current price < entry price
		if orderType == "take-profit" && price < cfg.OrderData.Entry {

			// get balance
			balance, err := api.Balance()
			if err != nil {
				log.Println("Getting balance:", err)
				continue
			}

			// cancel take-profit order
			_, err = api.CancelOrder(orderId)
			if err != nil {
				log.Println("Cancelling take-profit order:", err)
				continue
			}
			fmt.Println("Take-proft canceled")

			// place stop-loss order
			stopPrice := fmt.Sprintf("%v", cfg.OrderData.Stop)
			volume := fmt.Sprintf("%v", balance.XETH)
			args := map[string]string{"price": stopPrice}
			_, err = api.AddOrder(cfg.Pair, "sell", "stop-loss", volume, args)
			if err != nil {
				log.Println("Adding order:", err)
				continue
			}
			fmt.Println("Stop-loss placed")

			continue
		}

	}

}