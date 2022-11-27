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
	log := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	// read config
	file, err := os.ReadFile("./ganiu.json")
	if err != nil {
		log.Println("ERROR: Reading config:", err)
		return
	}

	// parse config
	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		log.Println("ERROR: Parsing config:", err)
		return
	}

	// connect to kraken api
	api := krakenapi.New(cfg.ApiData.Key, cfg.ApiData.Secret)

	// HANDLE BUSINESS
	counter := 0
	for {

		time.Sleep(time.Second * 30)

		// fetch open order
		orders, err := api.OpenOrders(make(map[string]string))
		if err != nil {
			log.Println("ERROR: Fetching open orders:", err)
			continue
		}

		// if no open orders, all gucci
		if len(orders.Open) == 0 {
			log.Println("ERROR: No open orders")
			break
		}

		// for now ganiu will operate with one open order
		if len(orders.Open) > 1 {
			log.Println("ERROR: Ganiu can't handle multiple open orders")
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
			if counter == 60 {
				log.Println("Limit order still pending")
				counter = 0
			}
			counter++
			continue
		}

		// get last price
		ticker, err := api.Ticker(cfg.Pair)
		if err != nil {
			log.Println("ERROR: Getting ticker:", err)
			continue
		}
		lastPrice := ticker.XETHZUSD.Close[0]

		// parse last price into float
		price, err := strconv.ParseFloat(lastPrice, 0)
		if err != nil {
			log.Println("ERROR: Parsing last price:", err)
			continue
		}

		// if order is stop-loss and current price > entry price
		if orderType == "stop-loss" && price > cfg.OrderData.Entry {
			err = HandleOrder(api, &cfg, orderId, orderType)
			if err != nil {
				log.Println(err)
			}
			continue
		}

		// if order is take-prfit and current price < entry price
		if orderType == "take-profit" && price < cfg.OrderData.Entry {
			err = HandleOrder(api, &cfg, orderId, orderType)
			if err != nil {
				log.Println(err)
			}
			continue
		}

	}

}

// HANDLE ORDER FUNCTION
func HandleOrder(api *krakenapi.KrakenApi, cfg *Config, orderId string, orderType string) error {
	// get balance
	balance, err := api.Balance()
	if err != nil {
		log.Println("ERROR: Getting balance:", err)
		return err
	}

	// cancel order by orderId
	_, err = api.CancelOrder(orderId)
	if err != nil {
		log.Println("ERROR: Cancelling order:", err)
		return err
	}
	log.Println(orderType + " order cancelled")

	// wait in cased the cancelation lags
	time.Sleep(time.Second * 1)

	// place order by orderType
	stopPrice := fmt.Sprintf("%v", cfg.OrderData.Stop)
	volume := fmt.Sprintf("%v", balance.XETH)
	args := map[string]string{"price": stopPrice}
	_, err = api.AddOrder(cfg.Pair, "sell", orderType, volume, args)
	if err != nil {
		log.Println("ERROR: Placing order:", err)
		return err
	}
	log.Println(orderType + " order placed")

	return nil
}
