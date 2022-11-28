package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	// START LISTENING
	logOk.Println("Ganiu is listening..")
	for {

		time.Sleep(time.Second * cfg.WaitTime)

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

		// get currency pair
		pair := cfg.Currency.Base + cfg.Currency.Quote

		// get last price
		lastPrice, err := GetLastPrice(pair)
		if err != nil {
			logErr.Println("Getting last price:", err)
			continue
		}

		// if order is stop-loss and current price > entry price
		if orderType == "stop-loss" && lastPrice > cfg.OrderData.Entry {
			newOrder := NewOrder{
				Pair:         pair,
				Type:         "take-profit",
				Entry:        cfg.OrderData.Take,
				BaseCurrency: cfg.Currency.Base,
			}

			err = HandleOrder(api, orderId, orderType, &newOrder)
			if err != nil {
				logErr.Println(err)
			}
			continue
		}

		// if order is take-profit and current price < entry price
		if orderType == "take-profit" && lastPrice < cfg.OrderData.Entry {
			newOrder := NewOrder{
				Pair:         pair,
				Type:         "stop-loss",
				Entry:        cfg.OrderData.Stop,
				BaseCurrency: cfg.Currency.Base,
			}

			err = HandleOrder(api, orderId, orderType, &newOrder)
			if err != nil {
				logErr.Println(err)
			}
			continue
		}

	}

}

// GET LAST PRICE
func GetLastPrice(pair string) (float64, error) {
	url := "https://api.kraken.com/0/public/Ticker?pair=" + pair
	resp, err := http.Get(url)
	if err != nil {
		logErr.Println("Requesting last price:", err)
		return 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logErr.Println("Reading response body:", err)
		return 0, err
	}

	var ticker Ticker
	err = json.Unmarshal(body, &ticker)
	if err != nil {
		logErr.Println("Parsing response body:", err)
		return 0, err
	}

	lastPrice, err := strconv.ParseFloat(ticker.Result[pair].Close[0], 0)
	if err != nil {
		logErr.Println("Parsing last price:", err)
		return 0, err
	}

	return lastPrice, nil
}

// HANDLE ORDER
func HandleOrder(api *kraken.KrakenApi, orderId, orderType string, newOrder *NewOrder) error {
	// get base currency volume to sell
	volume, err := GetVolume(api, newOrder.BaseCurrency)
	if err != nil {
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
	args := map[string]string{
		"price": fmt.Sprintf("%v", newOrder.Entry),
	}

	_, err = api.AddOrder(newOrder.Pair, "sell", newOrder.Type, volume, args)
	if err != nil {
		logErr.Println("Placing order:", err)
		return err
	}
	logOk.Println(newOrder.Type + " order placed")

	return nil
}

// GET BALANCE
func GetVolume(api *kraken.KrakenAPI, base string) (string, error) {
	// get balance
	balance, err := api.Balance()
	if err != nil {
		logErr.Println("Getting balance:", err)
		return "", err
	}

	// marshal balance struct
	data, err := json.Marshal(balance)
	if err != nil {
		logErr.Println("Marshalling balance:", err)
		return "", err
	}

	// unmarshal balance struct into interface map
	var bmap map[string]interface{}
	err = json.Unmarshal(data, &bmap)
	if err != nil {
		logErr.Println("Unmarshalling balance:", err)
		return "", err
	}

	// parse float to string in return value
	return fmt.Sprintf("%v", bmap[base]), nil
}
