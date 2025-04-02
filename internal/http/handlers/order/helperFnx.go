package order

import (
	"sort"

	"github.com/varunarora1606/Booking-App-Go/internal/memory"
)

type OrderType string

const (
	Limit OrderType = "limit"
	Market OrderType = "market"
)

func order(currentSide *map[int]memory.OrderDetails, oppositeSide *map[int]memory.OrderDetails, price int, user string, quantity int, orderType OrderType) {
	if orderType == Market {
		prices := make([]int, 0, len(*oppositeSide))

		for price := range *oppositeSide {
			prices = append(prices, price)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(prices)))

		for _, price := range prices {
			orderDetails := (*oppositeSide)[price]
			if orderDetails.Total <= quantity {
				quantity -= orderDetails.Total
				delete(*oppositeSide, price)
			} else {
				orderDetails.Total -= quantity
				for user := range orderDetails.Orders {
					if quantity == 0 {
						break
					}
					if orderDetails.Orders[user] <= quantity {
						quantity -= orderDetails.Orders[user]
						delete(orderDetails.Orders, user)
					} else {
						orderDetails.Orders[user] -= quantity
						quantity = 0
					}
				}
				(*oppositeSide)[price] = orderDetails
			}
			if quantity == 0 {
				break
			}
		}
	} else {
	//NOTE: Here currentSide and oppositeSide is opposite in case of sell And price is also 100 - price
	addToOrderBook := func(addQuantity int) {
		if orderDetails, exists := (*currentSide)[price]; !exists {
			(*currentSide)[price] = memory.OrderDetails{
				Total:  addQuantity,
				Orders: map[string]int{user: addQuantity},
			}
		} else {
			orderDetails.Total += addQuantity
			orderDetails.Orders[user] += addQuantity
			(*currentSide)[price] = orderDetails
		}
	}

	if orderDetails, exists := (*oppositeSide)[100-price]; !exists {
		addToOrderBook(quantity)
	} else {
		if orderDetails.Total <= quantity {
			quantity -= orderDetails.Total
			delete(*oppositeSide, 100-price)
			if quantity > 0 {
				addToOrderBook(quantity)
			}
		} else {
			orderDetails.Total -= quantity
			for user := range orderDetails.Orders {
				if quantity == 0 {
					return
				}
				if orderDetails.Orders[user] <= quantity {
					quantity -= orderDetails.Orders[user]
					delete(orderDetails.Orders, user)
				} else {
					orderDetails.Orders[user] -= quantity
				}
			}
			(*oppositeSide)[100-price] = orderDetails
		}
	}}
}

func buyStock(symbol string, orderSide memory.Yes_no, price int, user string, quantity int, orderType OrderType) {
	stockBook, exists := memory.OrderBook[symbol]
	if !exists {
		// TODO: return error or do it in the handler function
		stockBook = memory.StockBook{
			Yes: make(map[int]memory.OrderDetails),
			No:  make(map[int]memory.OrderDetails),
		}
		memory.OrderBook[symbol] = stockBook
	}

	var oppositeSide, currentSide *map[int]memory.OrderDetails

	if orderSide == memory.Yes {
		oppositeSide = &stockBook.No
		currentSide = &stockBook.Yes
	} else {
		oppositeSide = &stockBook.Yes
		currentSide = &stockBook.No
	}

	order(currentSide, oppositeSide, price, user, quantity, orderType)

	memory.OrderBook[symbol] = stockBook
}

func sellStock(symbol string, orderSide memory.Yes_no, price int, user string, quantity int, orderType OrderType) {
	stockBook, exists := memory.OrderBook[symbol]
	if !exists {
		// TODO: return error or do it in the handler function
		stockBook = memory.StockBook{
			Yes: make(map[int]memory.OrderDetails),
			No:  make(map[int]memory.OrderDetails),
		}
		memory.OrderBook[symbol] = stockBook
	}

	var currentSide, oppositeSide *map[int]memory.OrderDetails

	if orderSide == memory.Yes {
		oppositeSide = &stockBook.No
		currentSide = &stockBook.Yes
	} else {
		oppositeSide = &stockBook.Yes
		currentSide = &stockBook.No
	}

	order(oppositeSide, currentSide, 100-price, user, quantity, orderType)
	memory.OrderBook[symbol] = stockBook
}