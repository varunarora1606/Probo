package order

import (
	"sort"

	"github.com/google/uuid"
	"github.com/varunarora1606/Booking-App-Go/internal/memory"
)

func order(
	symbol string,
	currentSide *map[int]memory.OrderDetails,
	oppositeSide *map[int]memory.OrderDetails,
	price int,
	user string,
	quantity int,
	orderType memory.OrderType,
	transactionType memory.TransactionType,
	side memory.Yes_no,
) {
	if orderType == memory.Market {
		executeMarketOrder(symbol, currentSide, oppositeSide, price, user, quantity, transactionType, side)
	} else {
		executeLimitOrder(symbol, currentSide, oppositeSide, price, user, quantity, transactionType, side)
	}
}

func executeMarketOrder(
	symbol string,
	currentSide *map[int]memory.OrderDetails,
	oppositeSide *map[int]memory.OrderDetails,
	price int,
	user string,
	quantity int,
	transactionType memory.TransactionType,
	side memory.Yes_no,
) {
	prices := make([]int, 0, len(*oppositeSide))
	for p := range *oppositeSide {
		prices = append(prices, p)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(prices)))

	for _, p := range prices {
		orderDetails := (*oppositeSide)[p]

		if orderDetails.Total <= quantity {
			quantity -= orderDetails.Total
			executeTransaction(symbol, user, p, orderDetails.Total, transactionType, side)
			delete(*oppositeSide, p)
			dissolveOrders(symbol, orderDetails.Orders, p)
		} else {
			orderDetails.Total -= quantity
			executeTransaction(symbol, user, p, quantity, transactionType, side)
			orderDetails.Orders = updateOrders(symbol, orderDetails.Orders, p, &quantity)
			(*oppositeSide)[p] = orderDetails
		}

		if quantity == 0 {
			break
		}
	}
}

func executeLimitOrder(
	symbol string,
	currentSide *map[int]memory.OrderDetails,
	oppositeSide *map[int]memory.OrderDetails,
	price int,
	user string,
	quantity int,
	transactionType memory.TransactionType,
	side memory.Yes_no,
) {
	addToOrderBook := func(addQuantity int) {
		orderID := uuid.NewString()
		if orderDetails, exists := (*currentSide)[price]; !exists {
			(*currentSide)[price] = memory.OrderDetails{
				Total:  addQuantity,
				Orders: []string{orderID},
			}
		} else {
			orderDetails.Total += addQuantity
			orderDetails.Orders = append(orderDetails.Orders, orderID)
			(*currentSide)[price] = orderDetails
		}
	}

	oppPrice := 100 - price
	orderDetails, exists := (*oppositeSide)[oppPrice]

	if !exists {
		addToOrderBook(quantity)
		lockUserFunds(symbol, user, price, quantity, transactionType, side)
		return
	}

	if orderDetails.Total <= quantity {
		quantity -= orderDetails.Total
		executeTransaction(symbol, user, price, orderDetails.Total, transactionType, side)
		delete(*oppositeSide, oppPrice)
		dissolveOrders(symbol, orderDetails.Orders, price)
	} else {
		orderDetails.Total -= quantity
		executeTransaction(symbol, user, price, quantity, transactionType, side)
		orderDetails.Orders = updateOrders(symbol, orderDetails.Orders, price, &quantity)
		(*oppositeSide)[oppPrice] = orderDetails
	}

	if quantity > 0 {
		addToOrderBook(quantity)
		lockUserFunds(symbol, user, price, quantity, transactionType, side)
	}
}

func lockUserFunds(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Yes_no) {
	if transactionType == memory.Sell {
		stockBalance := memory.StockBalance[symbol][user][side]
		stockBalance.Quantity -= quantity
		stockBalance.Locked += quantity
		memory.StockBalance[symbol][user][side] = stockBalance
	} else {
		userInrBalance := memory.InrBalance[user]
		amount := price * quantity
		userInrBalance.Quantity -= amount
		userInrBalance.Locked += amount
		memory.InrBalance[user] = userInrBalance
	}
}

func executeTransaction(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Yes_no) {
	amount := price * quantity
	if transactionType == memory.Sell {
		userInrBalance := memory.InrBalance[user]
		userInrBalance.Quantity += amount
		memory.InrBalance[user] = userInrBalance

		if _, ok := memory.StockBalance[symbol]; !ok {
			memory.StockBalance[symbol] = make(map[string]map[memory.Yes_no]memory.Balance)
		}
		
		if _, ok := memory.StockBalance[symbol][user]; !ok {
			memory.StockBalance[symbol][user] = make(map[memory.Yes_no]memory.Balance)
		}

		stockBalance := memory.StockBalance[symbol][user][side]
		stockBalance.Quantity -= quantity
		memory.StockBalance[symbol][user][side] = stockBalance
	} else {
		userInrBalance := memory.InrBalance[user]
		userInrBalance.Quantity -= amount
		memory.InrBalance[user] = userInrBalance

		if _, ok := memory.StockBalance[symbol]; !ok {
			memory.StockBalance[symbol] = make(map[string]map[memory.Yes_no]memory.Balance)
		}
		
		if _, ok := memory.StockBalance[symbol][user]; !ok {
			memory.StockBalance[symbol][user] = make(map[memory.Yes_no]memory.Balance)
		}

		stockBalance := memory.StockBalance[symbol][user][side]
		stockBalance.Quantity += quantity
		memory.StockBalance[symbol][user][side] = stockBalance
	}
}

func dissolveOrders(symbol string, orders []string, price int) {
	for _, orderId := range orders {
		bet := memory.BetBook[orderId]
		delete(memory.BetBook, orderId)
		executeTransaction(symbol, bet.UserId, price, bet.Quantity, bet.TransactionType, bet.Side)
		adjustLockedBalance(symbol, bet.UserId, price, bet.Quantity, bet.TransactionType, bet.Side)
	}
}

func updateOrders(symbol string, orders []string, price int, quantity *int) []string {
	newOrders := []string{}
	for _, orderId := range orders {
		if *quantity == 0 {
			newOrders = append(newOrders, orderId)
			continue
		}

		bet := memory.BetBook[orderId]
		if bet.Quantity <= *quantity {
			*quantity -= bet.Quantity
			delete(memory.BetBook, orderId)
			executeTransaction(symbol, bet.UserId, price, bet.Quantity, bet.TransactionType, bet.Side)
			adjustLockedBalance(symbol, bet.UserId, price, bet.Quantity, bet.TransactionType, bet.Side)
		} else {
			bet.Quantity -= *quantity
			executeTransaction(symbol, bet.UserId, price, *quantity, bet.TransactionType, bet.Side)
			adjustLockedBalance(symbol, bet.UserId, price, *quantity, bet.TransactionType, bet.Side)
			memory.BetBook[orderId] = bet
			*quantity = 0
			newOrders = append(newOrders, orderId)
		}
	}
	return newOrders
}

func adjustLockedBalance(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Yes_no) {
	if transactionType == memory.Sell {
		stockBalance := memory.StockBalance[symbol][user][side]
		stockBalance.Locked -= quantity
		memory.StockBalance[symbol][user][side] = stockBalance
	} else {
		inrBalance := memory.InrBalance[user]
		inrBalance.Locked -= price * quantity
		memory.InrBalance[user] = inrBalance
	}
}

func buyStock(symbol string, orderSide memory.Yes_no, price int, user string, quantity int, orderType memory.OrderType) {
	stockBook, exists := memory.OrderBook[symbol]
	if !exists {
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

	order(symbol, currentSide, oppositeSide, price, user, quantity, orderType, memory.Buy, orderSide)
	memory.OrderBook[symbol] = stockBook
}

func sellStock(symbol string, orderSide memory.Yes_no, price int, user string, quantity int, orderType memory.OrderType) {
	stockBook, exists := memory.OrderBook[symbol]
	if !exists {
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

	order(symbol, oppositeSide, currentSide, 100-price, user, quantity, orderType, memory.Sell, orderSide)
	memory.OrderBook[symbol] = stockBook
}