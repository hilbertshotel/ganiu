package main

import "time"

// CONFIG
type OrderData struct {
	Entry float64
	Stop  float64
	Take  float64
}

type ApiData struct {
	Key    string
	Secret string
}

type Currency struct {
	Base  string
	Quote string
}

type Config struct {
	OrderData OrderData
	ApiData   ApiData
	WaitTime  time.Duration
	Currency  Currency
}

// TICKER
type PairInfo struct {
	Ask                []string `json:"a"`
	Bid                []string `json:"b"`
	Close              []string `json:"c"`
	Volume             []string `json:"v"`
	VolumeAveragePrice []string `json:"p"`
	Trades             []int    `json:"t"`
	Low                []string `json:"l"`
	High               []string `json:"h"`
	OpeningPrice       string   `json:"o"`
}

type Ticker struct {
	Error  []string            `json:"error"`
	Result map[string]PairInfo `json:"result"`
}

// NEW ORDER
type NewOrder struct {
	Pair         string
	Type         string
	Entry        float64
	BaseCurrency string
}
