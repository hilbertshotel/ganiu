package main

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

type Config struct {
	OrderData OrderData
	ApiData   ApiData
}

// NEW ORDER
type NewOrder struct {
	Type  string
	Entry float64
}
