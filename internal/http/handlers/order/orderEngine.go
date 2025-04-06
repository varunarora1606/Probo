package order

import (
	"sort"

	"github.com/google/uuid"
	"github.com/varunarora1606/Probo/internal/memory"
)

// TODO: Agr ek bhi process me problem hui toh pooraa revert back karna chahiye. Starting me hi sell karta hai tu kuch toh error aata hai but order and betbook me add ho jata hai tera order(Fixed it by moving addToOrderbook above) but still it is imp do it.

func flipSide(side memory.Side) memory.Side {
	sideFlip := map[memory.Side]memory.Side{
		memory.Yes: memory.No,
		memory.No:  memory.Yes,
	}
	return sideFlip[side]
}

func orderEngine(
	symbol string,
	currentSide *map[int]memory.OrderDetails,
	oppositeSide *map[int]memory.OrderDetails,
	price int,
	user string,
	quantity int,
	orderType memory.OrderType,
	transactionType memory.TransactionType,
	side memory.Side,
) {
	if orderType == memory.Market {
		executeMarketOrder(symbol, oppositeSide, user, quantity, transactionType, side)
	} else {
		executeLimitOrder(symbol, currentSide, oppositeSide, price, user, quantity, transactionType, side)
	}
}

func executeMarketOrder(
	symbol string,
	oppositeSide *map[int]memory.OrderDetails,
	user string,
	quantity int,
	transactionType memory.TransactionType,
	side memory.Side,
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
			dissolveOrders(symbol, orderDetails.Orders)
		} else {
			orderDetails.Total -= quantity
			executeTransaction(symbol, user, p, quantity, transactionType, side)
			orderDetails.Orders = updateOrders(symbol, orderDetails.Orders, &quantity)
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
	side memory.Side,
) {
	addToOrderBook := func(addQuantity int) {
		orderID := uuid.NewString()
		side = flipSide(side)
		memory.BetBook[orderID] = memory.BetDetails{
			UserId:          user,
			Price:           price,
			Quantity:        quantity,
			Side:            side, //TODO: iski side flip krni hai and then jha jha bet.Side use hua hai uski bhi flip karni hai
			TransactionType: transactionType,
		}
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
		lockUserFunds(symbol, user, price, quantity, transactionType, side)
		addToOrderBook(quantity)
		return
	}

	transactionPrice := price
	if transactionType == memory.Sell {
		transactionPrice = 100 - price
	}

	if orderDetails.Total <= quantity {
		quantity -= orderDetails.Total
		executeTransaction(symbol, user, transactionPrice, orderDetails.Total, transactionType, side)
		delete(*oppositeSide, oppPrice)
		dissolveOrders(symbol, orderDetails.Orders)
	} else {
		orderDetails.Total -= quantity
		executeTransaction(symbol, user, transactionPrice, quantity, transactionType, side)
		orderDetails.Orders = updateOrders(symbol, orderDetails.Orders, &quantity)
		(*oppositeSide)[oppPrice] = orderDetails
		quantity = 0
	}

	if quantity > 0 {
		addToOrderBook(quantity)
		lockUserFunds(symbol, user, price, quantity, transactionType, side)
	}
}

func lockUserFunds(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Side) {
	if transactionType == memory.Sell {
		stockBalance := memory.StockBalance[user][symbol][side]
		stockBalance.Quantity -= quantity
		stockBalance.Locked += quantity
		memory.StockBalance[user][symbol][side] = stockBalance
	} else {
		userInrBalance := memory.InrBalance[user]
		amount := price * quantity
		userInrBalance.Quantity -= amount
		userInrBalance.Locked += amount
		memory.InrBalance[user] = userInrBalance
	}
}

func executeTransaction(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Side) {
	amount := price * quantity
	if transactionType == memory.Sell {
		userInrBalance := memory.InrBalance[user]
		userInrBalance.Quantity += amount
		memory.InrBalance[user] = userInrBalance

		if _, ok := memory.StockBalance[user]; !ok {
			memory.StockBalance[user] = make(map[string]map[memory.Side]memory.Balance)
		}

		if _, ok := memory.StockBalance[user][symbol]; !ok {
			memory.StockBalance[user][symbol] = make(map[memory.Side]memory.Balance)
		}

		stockBalance := memory.StockBalance[user][symbol][side]
		stockBalance.Quantity -= quantity
		memory.StockBalance[user][symbol][side] = stockBalance
	} else {
		userInrBalance := memory.InrBalance[user]
		userInrBalance.Quantity -= amount
		memory.InrBalance[user] = userInrBalance

		if _, ok := memory.StockBalance[user]; !ok {
			memory.StockBalance[user] = make(map[string]map[memory.Side]memory.Balance)
		}

		if _, ok := memory.StockBalance[user][symbol]; !ok {
			memory.StockBalance[user][symbol] = make(map[memory.Side]memory.Balance)
		}

		stockBalance := memory.StockBalance[user][symbol][side]
		stockBalance.Quantity += quantity
		memory.StockBalance[user][symbol][side] = stockBalance
	}
}

func dissolveOrders(symbol string, orders []string) {
	for _, orderId := range orders {
		bet := memory.BetBook[orderId]
		delete(memory.BetBook, orderId)
		// executeTransaction(symbol, bet.UserId, price, bet.Quantity, bet.TransactionType, bet.Side)
		adjustLockedBalance(symbol, bet.UserId, bet.Price, bet.Quantity, bet.TransactionType, flipSide(bet.Side))
	}
}

func updateOrders(symbol string, orders []string, quantity *int) []string {
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
			// executeTransaction(symbol, bet.UserId, bet.Price, bet.Quantity, bet.TransactionType, bet.Side)
			adjustLockedBalance(symbol, bet.UserId, bet.Price, bet.Quantity, bet.TransactionType, flipSide(bet.Side))
		} else {
			bet.Quantity -= *quantity
			// executeTransaction(symbol, bet.UserId, bet.Price, *quantity, bet.TransactionType, bet.Side)
			adjustLockedBalance(symbol, bet.UserId, bet.Price, *quantity, bet.TransactionType, flipSide(bet.Side))
			memory.BetBook[orderId] = bet
			*quantity = 0
			newOrders = append(newOrders, orderId)
		}
	}
	return newOrders
}

func adjustLockedBalance(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Side) {
	if transactionType == memory.Sell {
		userInrBalance := memory.InrBalance[user]
		userInrBalance.Quantity += (100 - price) * quantity
		memory.InrBalance[user] = userInrBalance

		if _, ok := memory.StockBalance[user]; !ok {
			memory.StockBalance[user] = make(map[string]map[memory.Side]memory.Balance)
		}

		if _, ok := memory.StockBalance[user][symbol]; !ok {
			memory.StockBalance[user][symbol] = make(map[memory.Side]memory.Balance)
		}

		stockBalance := memory.StockBalance[user][symbol][side]
		stockBalance.Locked -= quantity
		memory.StockBalance[user][symbol][side] = stockBalance
	} else {
		inrBalance := memory.InrBalance[user]
		inrBalance.Locked -= price * quantity
		memory.InrBalance[user] = inrBalance

		if _, ok := memory.StockBalance[user]; !ok {
			memory.StockBalance[user] = make(map[string]map[memory.Side]memory.Balance)
		}

		if _, ok := memory.StockBalance[user][symbol]; !ok {
			memory.StockBalance[user][symbol] = make(map[memory.Side]memory.Balance)
		}

		stockBalance := memory.StockBalance[user][symbol][side]
		stockBalance.Quantity += quantity
		memory.StockBalance[user][symbol][side] = stockBalance
	}
}
