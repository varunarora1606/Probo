package memory

type Yes_no string

const (
	Yes Yes_no = "yes"
	No  Yes_no = "no"
)

type OrderDetails struct {
	Total  int            `json:"total"`
	Orders map[string]int `json:"orders"`
}
type StockBook struct {
	Yes map[int]OrderDetails `json:"yes"`
	No  map[int]OrderDetails `json:"no"`
}

type Balance struct {
	Quantity int `json:"quantity"`
	Locked   int `json:"locked"`
}

var OrderBook = make(map[string]StockBook)

var InrBalance = make(map[string]Balance)

var StockBalance = make(map[string]map[string]map[Yes_no]Balance)