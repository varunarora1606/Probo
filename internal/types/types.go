package types

import "github.com/varunarora1606/Probo/internal/memory"

type Input struct {
	ApiId           string
	Fnx             string
	UserId          string
	Symbol          string
	Question string
	EndTime int64
	Quantity        int
	Price           int
	StockSide       memory.Side
	StockType       memory.OrderType
	TransactionType memory.TransactionType
}

type Output struct {
	ForWs bool
	ApiId string
	Err string
	Market memory.SymbolBook
	Markets map[string]memory.SymbolBook
	StockBook memory.StockBook
	InrBalance memory.Balance
	StockBalance map[string]map[memory.Side]memory.Balance
	PortfolioItems []PortfolioItem
	Deltas []memory.Delta
	Trade memory.Trade
}

type PortfolioItem struct {
	Symbol string
	Value int
	Quantity int
}