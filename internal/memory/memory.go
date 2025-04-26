package memory

import "sync"

type Side string

const (
	Yes Side = "yes"
	No  Side = "no"
)

type OrderType string

const (
	Limit  OrderType = "limit"
	Market OrderType = "market"
)

type TransactionType string

const (
	Buy  TransactionType = "buy"
	Sell TransactionType = "sell"
)

type OrderDetails struct {
	Total  int      `json:"total"`
	Orders []string `json:"orders"` // int to BetDetails
}
type StockBook struct {
	Yes map[int]OrderDetails `json:"yes"`
	No  map[int]OrderDetails `json:"no"`
}

type Balance struct {
	Quantity int `json:"quantity"`
	Locked   int `json:"locked"`
}

type BetDetails struct {
	UserId          string          `json:"userId"`
	Price           int             `json:"price"`
	Quantity        int             `json:"quantity"`
	Side            Side            `json:"side"`
	TransactionType TransactionType `json:"transactionType"`
}

type Delta struct {
	Msg string      // remove, update, add  // Partially filled/ Fully filled/Unfilled (market)
	OrderId string
	Symbol string
	Price int
	Side Side
	Quantity int
}

type MicroTrade struct {
	Quantity int
	Price int
}

type Trade struct {
	TotalQuantity int
	MicroTrades []MicroTrade
}

type OrderBookType struct {
	Mu sync.RWMutex
	Data map[string]StockBook
}
type InrBalanceType struct {
	Mu sync.RWMutex
	Data map[string]Balance
}
type StockBalanceType struct {
	Mu sync.RWMutex
	Data map[string]map[string]map[Side]Balance
}
type BetBookType struct {
	Mu sync.RWMutex
	Data map[string]BetDetails
}

var OrderBook = OrderBookType {
	Data: make(map[string]StockBook),
}

var InrBalance = InrBalanceType {
	Data: make(map[string]Balance),
}

var StockBalance = StockBalanceType {
	Data: make(map[string]map[string]map[Side]Balance),
}

var BetBook = BetBookType {
	Data: make(map[string]BetDetails),
}
