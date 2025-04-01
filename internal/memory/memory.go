package memory

type OrderDetails struct {
	Total  int            `json:"total"`
	Orders map[string]int `json:"orders"`
}
type StockBook struct {
	Yes map[int]OrderDetails `json:"yes"`
	No  map[int]OrderDetails `json:"no"`
}

var OrderBook = make(map[string]StockBook)