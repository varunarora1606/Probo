package order

import (
	"github.com/varunarora1606/Probo/internal/memory"
)

func buyStock(symbol string, orderSide memory.Side, price int, user string, quantity int, orderType memory.OrderType) {
	stockBook := memory.OrderBook[symbol]

	var currentSide, oppositeSide *map[int]memory.OrderDetails
	if orderSide == memory.Yes {
		oppositeSide = &stockBook.No
		currentSide = &stockBook.Yes
	} else {
		oppositeSide = &stockBook.Yes
		currentSide = &stockBook.No
	}

	orderEngine(symbol, currentSide, oppositeSide, price, user, quantity, orderType, memory.Buy, orderSide)
	memory.OrderBook[symbol] = stockBook
}

func sellStock(symbol string, orderSide memory.Side, price int, user string, quantity int, orderType memory.OrderType) {
	stockBook := memory.OrderBook[symbol]

	var currentSide, oppositeSide *map[int]memory.OrderDetails
	if orderSide == memory.Yes {
		oppositeSide = &stockBook.No
		currentSide = &stockBook.Yes
	} else {
		oppositeSide = &stockBook.Yes
		currentSide = &stockBook.No
	}

	orderEngine(symbol, oppositeSide, currentSide, 100-price, user, quantity, orderType, memory.Sell, orderSide)
	memory.OrderBook[symbol] = stockBook
}
