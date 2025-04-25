package engine

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/varunarora1606/Probo/internal/memory"
)

var tempOrderBook map[string]memory.StockBook
var tempInrBalance map[string]memory.Balance
var tempStockBalance map[string]map[string]map[memory.Side]memory.Balance
var tempBetBook map[string]memory.BetDetails

// For WS
var deltas []memory.Delta

// TODO: Agr ek bhi process me problem hui toh pooraa revert back karna chahiye. Starting me hi sell karta hai tu kuch toh error aata hai but order and betbook me add ho jata hai tera order(Fixed it by moving addToOrderbook above) but still it is imp do it.

func flipSide(side memory.Side) memory.Side {
	sideFlip := map[memory.Side]memory.Side{
		memory.Yes: memory.No,
		memory.No:  memory.Yes,
	}
	return sideFlip[side]
}

func OrderEngine(
	symbol string,
	side memory.Side,
	price int,
	user string,
	quantity int,
	orderType memory.OrderType,
	transactionType memory.TransactionType,
) error {
	// mutex
	memory.OrderBook.Mu.Lock()
	memory.InrBalance.Mu.Lock()
	memory.StockBalance.Mu.Lock()
	memory.BetBook.Mu.Lock()
	defer memory.OrderBook.Mu.Unlock()
	defer memory.InrBalance.Mu.Unlock()
	defer memory.StockBalance.Mu.Unlock()
	defer memory.BetBook.Mu.Unlock()

	tempOrderBook = memory.OrderBook.Data
	tempInrBalance = memory.InrBalance.Data
	tempStockBalance = memory.StockBalance.Data
	tempBetBook = memory.BetBook.Data

	if orderType == memory.Limit && transactionType == memory.Buy && tempInrBalance[user].Quantity < price * quantity {
		return fmt.Errorf("insufficient balance")
	}
	if transactionType == memory.Sell && tempStockBalance[user][symbol][side].Quantity < quantity {
		return fmt.Errorf("insufficient stocks balance")
	}

	stockBook, exist := tempOrderBook[symbol]
	if !exist {
		return fmt.Errorf("no such symbol exists")
	}

	var currentSide, oppositeSide *map[int]memory.OrderDetails
	if side == memory.Yes {
		oppositeSide = &stockBook.No
		currentSide = &stockBook.Yes
	} else {
		oppositeSide = &stockBook.Yes
		currentSide = &stockBook.No
	}

	if transactionType == memory.Sell {
		tempSide := oppositeSide
		oppositeSide = currentSide
		currentSide = tempSide
		price = 100 - price
	}

	if orderType == memory.Market {
		if err := executeMarketOrder(symbol, oppositeSide, user, quantity, transactionType, side); err != nil {
			return err
		}
	} else {
		// Mostly isme error nhi aega bcoz this have aalreaady been checked above
		if err := executeLimitOrder(symbol, currentSide, oppositeSide, price, user, quantity, transactionType, side); err != nil {
			return err
		}
	}
	tempOrderBook[symbol] = stockBook

	memory.OrderBook.Data = tempOrderBook
	memory.InrBalance.Data = tempInrBalance
	memory.StockBalance.Data = tempStockBalance
	memory.BetBook.Data = tempBetBook

	return nil
}

func executeMarketOrder(
	symbol string,
	oppositeSide *map[int]memory.OrderDetails,
	user string,
	quantity int,
	transactionType memory.TransactionType,
	side memory.Side,
) error {
	prices := make([]int, 0, len(*oppositeSide))
	for p := range *oppositeSide {
		prices = append(prices, p)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(prices)))

	for _, p := range prices {
		orderDetails := (*oppositeSide)[p]

		if orderDetails.Total <= quantity {
			quantity -= orderDetails.Total
			if err := executeTransaction(symbol, user, p, orderDetails.Total, transactionType, side); err != nil {
				return err
			}
			delete(*oppositeSide, p)
			dissolveOrders(symbol, orderDetails.Orders)
		} else {
			orderDetails.Total -= quantity
			if err := executeTransaction(symbol, user, p, quantity, transactionType, side); err != nil {
				return err
			}
			orderDetails.Orders = updateOrders(symbol, orderDetails.Orders, &quantity)
			(*oppositeSide)[p] = orderDetails
		}

		if quantity == 0 {
			break
		}
	}
	return nil
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
) error {
	addToOrderBook := func(addQuantity int) {
		orderID := uuid.NewString()
		side = flipSide(side)
		tempBetBook[orderID] = memory.BetDetails{
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
		if err := lockUserFunds(symbol, user, price, quantity, transactionType, side); err != nil {
			return err
		}
		addToOrderBook(quantity)
		return nil
	}

	transactionPrice := price
	if transactionType == memory.Sell {
		transactionPrice = 100 - price
	}

	if orderDetails.Total <= quantity {
		quantity -= orderDetails.Total
		if err := executeTransaction(symbol, user, transactionPrice, orderDetails.Total, transactionType, side); err != nil {
			return err
		}
		delete(*oppositeSide, oppPrice)
		dissolveOrders(symbol, orderDetails.Orders)
	} else {
		orderDetails.Total -= quantity
		if err := executeTransaction(symbol, user, transactionPrice, quantity, transactionType, side); err != nil {
			return err
		}
		orderDetails.Orders = updateOrders(symbol, orderDetails.Orders, &quantity)
		(*oppositeSide)[oppPrice] = orderDetails
		quantity = 0
	}

	if quantity > 0 {
		addToOrderBook(quantity)
		if err := lockUserFunds(symbol, user, price, quantity, transactionType, side); err != nil {
			return err
		}
	}
	return nil
}

func lockUserFunds(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Side) error {
	if transactionType == memory.Sell {
		stockBalance := tempStockBalance[user][symbol][side]
		stockBalance.Quantity -= quantity
		stockBalance.Locked += quantity
		tempStockBalance[user][symbol][side] = stockBalance
	} else {
		userInrBalance := tempInrBalance[user]
		amount := price * quantity
		if userInrBalance.Quantity < amount {
			return fmt.Errorf("insufficient balance")
		}
		userInrBalance.Quantity -= amount
		userInrBalance.Locked += amount
		tempInrBalance[user] = userInrBalance
	}
	return nil
}

func executeTransaction(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Side) error {
	amount := price * quantity
	if transactionType == memory.Sell {
		userInrBalance := tempInrBalance[user]
		userInrBalance.Quantity += amount
		tempInrBalance[user] = userInrBalance

		if _, ok := tempStockBalance[user]; !ok {
			tempStockBalance[user] = make(map[string]map[memory.Side]memory.Balance)
		}

		if _, ok := tempStockBalance[user][symbol]; !ok {
			tempStockBalance[user][symbol] = make(map[memory.Side]memory.Balance)
		}

		stockBalance := tempStockBalance[user][symbol][side]
		stockBalance.Quantity -= quantity
		tempStockBalance[user][symbol][side] = stockBalance
	} else {
		userInrBalance := tempInrBalance[user]
		if userInrBalance.Quantity < amount {
			return fmt.Errorf("insufficient balance")
		}
		userInrBalance.Quantity -= amount
		tempInrBalance[user] = userInrBalance

		if _, ok := tempStockBalance[user]; !ok {
			tempStockBalance[user] = make(map[string]map[memory.Side]memory.Balance)
		}

		if _, ok := tempStockBalance[user][symbol]; !ok {
			tempStockBalance[user][symbol] = make(map[memory.Side]memory.Balance)
		}

		stockBalance := tempStockBalance[user][symbol][side]
		stockBalance.Quantity += quantity
		tempStockBalance[user][symbol][side] = stockBalance
	}
	return nil
}

func dissolveOrders(symbol string, orders []string) {
	for _, orderId := range orders {
		bet := tempBetBook[orderId]
		delete(tempBetBook, orderId)
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

		bet := tempBetBook[orderId]
		if bet.Quantity <= *quantity {
			*quantity -= bet.Quantity
			delete(tempBetBook, orderId)
			// executeTransaction(symbol, bet.UserId, bet.Price, bet.Quantity, bet.TransactionType, bet.Side)
			adjustLockedBalance(symbol, bet.UserId, bet.Price, bet.Quantity, bet.TransactionType, flipSide(bet.Side))
		} else {
			bet.Quantity -= *quantity
			// executeTransaction(symbol, bet.UserId, bet.Price, *quantity, bet.TransactionType, bet.Side)
			adjustLockedBalance(symbol, bet.UserId, bet.Price, *quantity, bet.TransactionType, flipSide(bet.Side))
			tempBetBook[orderId] = bet
			*quantity = 0
			newOrders = append(newOrders, orderId)
		}
	}
	return newOrders
}

func adjustLockedBalance(symbol, user string, price, quantity int, transactionType memory.TransactionType, side memory.Side) {
	if transactionType == memory.Sell {
		userInrBalance := tempInrBalance[user]
		userInrBalance.Quantity += (100 - price) * quantity
		tempInrBalance[user] = userInrBalance

		if _, ok := tempStockBalance[user]; !ok {
			tempStockBalance[user] = make(map[string]map[memory.Side]memory.Balance)
		}

		if _, ok := tempStockBalance[user][symbol]; !ok {
			tempStockBalance[user][symbol] = make(map[memory.Side]memory.Balance)
		}

		stockBalance := tempStockBalance[user][symbol][side]
		stockBalance.Locked -= quantity
		tempStockBalance[user][symbol][side] = stockBalance
	} else {
		inrBalance := tempInrBalance[user]
		inrBalance.Locked -= price * quantity
		tempInrBalance[user] = inrBalance

		if _, ok := tempStockBalance[user]; !ok {
			tempStockBalance[user] = make(map[string]map[memory.Side]memory.Balance)
		}

		if _, ok := tempStockBalance[user][symbol]; !ok {
			tempStockBalance[user][symbol] = make(map[memory.Side]memory.Balance)
		}

		stockBalance := tempStockBalance[user][symbol][side]
		stockBalance.Quantity += quantity
		tempStockBalance[user][symbol][side] = stockBalance
	}
}
