package engine

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/varunarora1606/Probo/internal/memory"
	"github.com/varunarora1606/Probo/internal/types"
)

var tempSymbolBook types.SymbolBook
var tempStockBook types.StockBook
var tempInrBalance map[string]types.Balance
var tempStockBalance map[string]map[string]map[types.Side]types.Balance
var tempBetBook map[string]types.BetDetails

// For WS
var deltas []types.Delta

// For Api
var trade types.Trade

// func flipSide(side types.Side) types.Side {
// 	sideFlip := map[types.Side]types.Side{
// 		types.Yes: types.No,
// 		types.No:  types.Yes,
// 	}
// 	return sideFlip[side]
// }

func OrderEngine(
	symbol string,
	side types.Side,
	price int,
	user string,
	quantity int,
	orderType types.OrderType,
	transactionType types.TransactionType,
) (types.Trade, []types.Delta, error) {
	// mutex
	memory.MarketBook.Mu.Lock()
	memory.OrderBook.Mu.Lock()
	memory.InrBalance.Mu.Lock()
	memory.StockBalance.Mu.Lock()
	memory.BetBook.Mu.Lock()

	defer func() {
		memory.MarketBook.Mu.Unlock()
		memory.OrderBook.Mu.Unlock()
		memory.InrBalance.Mu.Unlock()
		memory.StockBalance.Mu.Unlock()
		memory.BetBook.Mu.Unlock()
		fmt.Println("tempBetBook:", tempSymbolBook)
		fmt.Println("tempBetBook:", tempStockBook)
		fmt.Println("tempInrBalance:", tempInrBalance)
		fmt.Println("tempStockBalance:", tempStockBalance)
		fmt.Println("tempBetBook:", tempBetBook)
		tempSymbolBook = types.SymbolBook{}
		tempStockBook = types.StockBook{}
		tempInrBalance = make(map[string]types.Balance)
		tempStockBalance = make(map[string]map[string]map[types.Side]types.Balance)
		tempBetBook = make(map[string]types.BetDetails)
		fmt.Println("defer called")
		trade = types.Trade{}
		deltas = []types.Delta{}
	}()
	var err error = nil

	tempSymbolBook, err = partialCopySymbolBook(symbol, memory.MarketBook.Data)
	if err != nil {
		return types.Trade{}, []types.Delta{}, err
	}
	tempStockBook, err = partialCopyStockBook(symbol, memory.OrderBook.Data)
	if err != nil {
		return types.Trade{}, []types.Delta{}, err
	}
	// TODO: Can be improved by adding only the necessary info at the time of requirement in below fnxs.
	tempInrBalance = partialCopyInrBalance(memory.InrBalance.Data)
	tempStockBalance = partialCopyStockBalance(memory.StockBalance.Data)
	tempBetBook = partialCopyBetBook(memory.BetBook.Data)

	if orderType == types.Limit && transactionType == types.Buy && tempInrBalance[user].Quantity < price*quantity {
		return types.Trade{}, []types.Delta{}, fmt.Errorf("insufficient balance")
	}
	fmt.Println("transactionType:", transactionType)
	if transactionType == types.Sell && tempStockBalance[user][symbol][side].Quantity < quantity {
		return types.Trade{}, []types.Delta{}, fmt.Errorf("insufficient stocks balance")
	}

	var currentSide, oppositeSide *map[int]types.OrderDetails
	if side == types.Yes {
		oppositeSide = &tempStockBook.No
		currentSide = &tempStockBook.Yes
	} else {
		oppositeSide = &tempStockBook.Yes
		currentSide = &tempStockBook.No
	}

	if transactionType == types.Sell {
		tempSide := oppositeSide
		oppositeSide = currentSide
		currentSide = tempSide
		price = 100 - price
	}

	if orderType == types.Market {
		if err := executeMarketOrder(symbol, oppositeSide, user, quantity, transactionType, side, types.Market, 0); err != nil {
			return types.Trade{}, []types.Delta{}, err
		}

		if trade.TotalQuantity < quantity {
			return types.Trade{}, []types.Delta{}, fmt.Errorf("not enough liquidity available in the market")
		}

	} else {
		// Mostly isme error nhi aega bcoz this have aalreaady been checked above
		if err := executeLimitOrder(symbol, currentSide, oppositeSide, price, user, quantity, transactionType, side); err != nil {
			return types.Trade{}, []types.Delta{}, err
		}
	}

	// TODO: Add volume to symbolBook using "trade"

	memory.MarketBook.Data[symbol] = tempSymbolBook
	memory.OrderBook.Data[symbol] = tempStockBook
	memory.InrBalance.Data = tempInrBalance
	memory.StockBalance.Data = tempStockBalance
	memory.BetBook.Data = tempBetBook

	return trade, deltas, nil
}

func executeMarketOrder(
	symbol string,
	oppositeSide *map[int]types.OrderDetails,
	user string,
	quantity int,
	transactionType types.TransactionType,
	side types.Side,
	orderType types.OrderType,
	price int,
) error {
	prices := make([]int, 0, len(*oppositeSide))
	for p := range *oppositeSide {
		prices = append(prices, p)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(prices)))

	for _, p := range prices {
		if orderType == types.Limit && 100 - p > price {
			return nil
		}

		orderDetails := (*oppositeSide)[p]

		if orderDetails.Total <= quantity {
			quantity -= orderDetails.Total
			if err := executeTransaction(symbol, user, 100-p, orderDetails.Total, transactionType, side); err != nil {
				return err
			}
			dissolveOrders(symbol, orderDetails.Orders)
			delete(*oppositeSide, p)
		} else {
			orderDetails.Total -= quantity
			if err := executeTransaction(symbol, user, 100-p, quantity, transactionType, side); err != nil {
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
	currentSide *map[int]types.OrderDetails,
	oppositeSide *map[int]types.OrderDetails,
	price int,
	user string,
	quantity int,
	transactionType types.TransactionType,
	side types.Side,
) error {
	originalPrice := price
	if transactionType == types.Sell {
		originalPrice = 100 - price
	} // keep

	addToOrderBook := func(addQuantity int) {
		orderId := uuid.NewString()
		// flipedSide := flipSide(side)
		tempBetBook[orderId] = types.BetDetails{
			UserId:          user,
			Price:           originalPrice,
			Quantity:        quantity,
			Side:            side,
			TransactionType: transactionType,
		}
		deltas = append(deltas, types.Delta{
			Msg: "open",
			Order: types.Order{
				BetId:           orderId,
				EventId:         uuid.NewString(),
				UserID:          user,
				Symbol:          symbol,
				Side:            side,
				TransactionType: transactionType,
				Price:           originalPrice,
				Quantity:        quantity,
				CreatedAt:       time.Now(),
			},
		})
		if orderDetails, exists := (*currentSide)[price]; !exists {
			(*currentSide)[price] = types.OrderDetails{
				Total:  addQuantity,
				Orders: []string{orderId},
			}
		} else {
			orderDetails.Total += addQuantity
			orderDetails.Orders = append(orderDetails.Orders, orderId)
			(*currentSide)[price] = orderDetails
		}
	}

	// Limit order is also amrket order but with a limit.
	if err := executeMarketOrder(symbol, oppositeSide, user, quantity, transactionType, side, types.Limit, originalPrice); err != nil {
		return err
	}
	quantity = quantity - trade.TotalQuantity

	// Limit order
	// oppPrice := 100 - price
	// orderDetails, exists := (*oppositeSide)[oppPrice]

	// if !exists {
	// 	if err := lockUserFunds(symbol, user, price, quantity, transactionType, side); err != nil {
	// 		return err
	// 	}
	// 	addToOrderBook(quantity)
	// 	return nil
	// }

	// if orderDetails.Total <= quantity {
	// 	quantity -= orderDetails.Total
	// 	if err := executeTransaction(symbol, user, transactionPrice, orderDetails.Total, transactionType, side); err != nil {
	// 		return err
	// 	}
	// 	delete(*oppositeSide, oppPrice)
	// 	dissolveOrders(symbol, orderDetails.Orders)
	// } else {
	// 	orderDetails.Total -= quantity
	// 	if err := executeTransaction(symbol, user, transactionPrice, quantity, transactionType, side); err != nil {
	// 		return err
	// 	}
	// 	orderDetails.Orders = updateOrders(symbol, orderDetails.Orders, &quantity)
	// 	(*oppositeSide)[oppPrice] = orderDetails
	// 	quantity = 0
	// }

	if quantity > 0 {
		if err := lockUserFunds(symbol, user, price, quantity, transactionType, side); err != nil {
			return err
		}
		addToOrderBook(quantity)
	} //keep
	return nil
}

func lockUserFunds(symbol, user string, price int, quantity int, transactionType types.TransactionType, side types.Side) error {
	if transactionType == types.Sell {
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

func executeTransaction(symbol string, user string, price int, quantity int, transactionType types.TransactionType, side types.Side) error {
	amount := price * quantity
	if transactionType == types.Sell {
		userInrBalance := tempInrBalance[user]
		userInrBalance.Quantity += amount
		tempInrBalance[user] = userInrBalance

		if _, ok := tempStockBalance[user]; !ok {
			tempStockBalance[user] = make(map[string]map[types.Side]types.Balance)
		}

		if _, ok := tempStockBalance[user][symbol]; !ok {
			tempStockBalance[user][symbol] = make(map[types.Side]types.Balance)
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
			tempStockBalance[user] = make(map[string]map[types.Side]types.Balance)
		}

		if _, ok := tempStockBalance[user][symbol]; !ok {
			tempStockBalance[user][symbol] = make(map[types.Side]types.Balance)
		}

		stockBalance := tempStockBalance[user][symbol][side]
		stockBalance.Quantity += quantity
		tempStockBalance[user][symbol][side] = stockBalance
	}
	if side == types.No {
		tempSymbolBook.YesClosing = 100 - price
	} else {
		tempSymbolBook.YesClosing = price
	}
	trade.TotalQuantity += quantity
	trade.MicroTrades = append(trade.MicroTrades, types.MicroTrade{
		Quantity: quantity,
		Price:    price,
	})
	return nil
}

func dissolveOrders(symbol string, orders []string) {
	for _, orderId := range orders {
		bet := tempBetBook[orderId]
		delete(tempBetBook, orderId)
		// executeTransaction(symbol, bet.UserId, price, bet.Quantity, bet.TransactionType, bet.Side)
		adjustLockedBalance(symbol, bet.UserId, bet.Price, bet.Quantity, bet.TransactionType, (bet.Side))
		deltas = append(deltas, types.Delta{
			Msg: "matched",
			Order: types.Order{
				BetId:           orderId,
				EventId:         uuid.NewString(),
				UserID:          bet.UserId,
				Symbol:          symbol,
				Side:            bet.Side,
				TransactionType: bet.TransactionType,
				Price:           bet.Price,
				Quantity:        bet.Quantity,
				CreatedAt:       time.Now(),
			},
		})
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
			adjustLockedBalance(symbol, bet.UserId, bet.Price, bet.Quantity, bet.TransactionType, (bet.Side))
			deltas = append(deltas, types.Delta{
				Msg: "matched",
				Order: types.Order{
					BetId:           orderId,
					EventId:         uuid.NewString(),
					UserID:          bet.UserId,
					Symbol:          symbol,
					Side:            bet.Side,
					TransactionType: bet.TransactionType,
					Price:           bet.Price,
					Quantity:        bet.Quantity,
					CreatedAt:       time.Now(),
				},
			})
		} else {
			bet.Quantity -= *quantity
			// executeTransaction(symbol, bet.UserId, bet.Price, *quantity, bet.TransactionType, bet.Side)
			adjustLockedBalance(symbol, bet.UserId, bet.Price, *quantity, bet.TransactionType, (bet.Side))
			deltas = append(deltas, types.Delta{
				Msg: "update",
				Order: types.Order{
					BetId:           orderId,
					EventId:         uuid.NewString(),
					UserID:          bet.UserId,
					Symbol:          symbol,
					Side:            bet.Side,
					TransactionType: bet.TransactionType,
					Price:           bet.Price,
					Quantity:        bet.Quantity,
					CreatedAt:       time.Now(),
				},
			})
			tempBetBook[orderId] = bet
			*quantity = 0
			newOrders = append(newOrders, orderId)
		}
	}
	return newOrders
}

func adjustLockedBalance(symbol, user string, price, quantity int, transactionType types.TransactionType, side types.Side) {
	if transactionType == types.Sell {
		userInrBalance := tempInrBalance[user]
		userInrBalance.Quantity += (price) * quantity
		tempInrBalance[user] = userInrBalance

		if _, ok := tempStockBalance[user]; !ok {
			tempStockBalance[user] = make(map[string]map[types.Side]types.Balance)
		}

		if _, ok := tempStockBalance[user][symbol]; !ok {
			tempStockBalance[user][symbol] = make(map[types.Side]types.Balance)
		}

		stockBalance := tempStockBalance[user][symbol][side]
		stockBalance.Locked -= quantity
		tempStockBalance[user][symbol][side] = stockBalance
	} else {
		inrBalance := tempInrBalance[user]
		inrBalance.Locked -= price * quantity
		tempInrBalance[user] = inrBalance

		if _, ok := tempStockBalance[user]; !ok {
			tempStockBalance[user] = make(map[string]map[types.Side]types.Balance)
		}

		if _, ok := tempStockBalance[user][symbol]; !ok {
			tempStockBalance[user][symbol] = make(map[types.Side]types.Balance)
		}

		stockBalance := tempStockBalance[user][symbol][side]
		stockBalance.Quantity += quantity
		tempStockBalance[user][symbol][side] = stockBalance
	}
}
