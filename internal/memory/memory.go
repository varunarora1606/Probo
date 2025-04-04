package memory

type Yes_no string

const (
	Yes Yes_no = "yes"
	No  Yes_no = "no"
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
	Side            Yes_no          `json:"side"`
	TransactionType TransactionType `json:"transactionType"`
}

var OrderBook = make(map[string]StockBook)

var InrBalance = make(map[string]Balance)

var StockBalance = make(map[string]map[string]map[Yes_no]Balance)

var BetBook = make(map[string]BetDetails)