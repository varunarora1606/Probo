package order

import (
	"sort"

	"github.com/google/uuid"
	"github.com/varunarora1606/Booking-App-Go/internal/memory"
)

func order(symbol string, currentSide *map[int]memory.OrderDetails, oppositeSide *map[int]memory.OrderDetails, price int, user string, quantity int, orderType memory.OrderType, transactionType memory.TransactionType, side memory.Yes_no) {
	if orderType == memory.Market {
		prices := make([]int, 0, len(*oppositeSide))

		for price := range *oppositeSide {
			prices = append(prices, price)
		}

		sort.Sort(sort.Reverse(sort.IntSlice(prices)))

		for _, price := range prices {
			orderDetails := (*oppositeSide)[price]
			if orderDetails.Total <= quantity {
				quantity -= orderDetails.Total
				if transactionType == memory.Sell {
					userInrBalance := memory.InrBalance[user]
					userInrBalance.Quantity += (100 - price) * orderDetails.Total
					memory.InrBalance[user] = userInrBalance

					stockBalance := memory.StockBalance[symbol][user][side]
					stockBalance.Quantity -= orderDetails.Total
					memory.StockBalance[symbol][user][side] = stockBalance
				} else {
					userInrBalance := memory.InrBalance[user]
					userInrBalance.Quantity -= (price) * orderDetails.Total
					memory.InrBalance[user] = userInrBalance

					stockBalance := memory.StockBalance[symbol][user][side]
					stockBalance.Quantity += orderDetails.Total
					memory.StockBalance[symbol][user][side] = stockBalance
				}
				// TODO: Add to the user balance
				// if transactionType == memory.Sell {

				// } else {

				// }
				// transaction[price] = orderDetails.Total
				delete(*oppositeSide, price)
				// TODO: Delete all order
				for _, orderId := range orderDetails.Orders {
					// Remove the order from orderBook
					bet := memory.BetBook[orderId]
					delete(memory.BetBook, orderId)
					if bet.TransactionType == memory.Sell {
						// Return the INR to his account for that bet
						inrBalance := memory.InrBalance[bet.UserId]
						inrBalance.Quantity += (100 - price) * bet.Quantity
						memory.InrBalance[bet.UserId] = inrBalance

						// Remove the stock from his stockBalance
						stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
						stockBalance.Locked -= bet.Quantity
						memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance

					} else {
						// Remove the INR from his account for that bet
						inrBalance := memory.InrBalance[bet.UserId]
						inrBalance.Locked -= (100 - price) * bet.Quantity
						memory.InrBalance[bet.UserId] = inrBalance

						// Add the stock to his stockBalance
						stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
						stockBalance.Quantity += bet.Quantity
						memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance
					}
				}

			} else {
				orderDetails.Total -= quantity
				if transactionType == memory.Sell {
					userInrBalance := memory.InrBalance[user]
					userInrBalance.Quantity += (100 - price) * quantity
					memory.InrBalance[user] = userInrBalance

					stockBalance := memory.StockBalance[symbol][user][side]
					stockBalance.Quantity -= quantity
					memory.StockBalance[symbol][user][side] = stockBalance
				} else {
					userInrBalance := memory.InrBalance[user]
					userInrBalance.Quantity -= (price) * quantity
					memory.InrBalance[user] = userInrBalance

					stockBalance := memory.StockBalance[symbol][user][side]
					stockBalance.Quantity += quantity
					memory.StockBalance[symbol][user][side] = stockBalance
				}
				for _, orderId := range orderDetails.Orders {
					// TODO: Desolve the Orderbook (Left)
					if quantity == 0 {
						break
					}
					bet := memory.BetBook[orderId]
					if bet.Quantity <= quantity {
						quantity -= bet.Quantity
						// Disolve the transactions
						delete(memory.BetBook, orderId)
						if bet.TransactionType == memory.Sell {
							// Return the INR to his account for that bet
							inrBalance := memory.InrBalance[bet.UserId]
							inrBalance.Quantity += (100 - price) * bet.Quantity
							memory.InrBalance[bet.UserId] = inrBalance

							// Remove the stock from his stockBalance
							stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
							stockBalance.Locked -= bet.Quantity
							memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance

						} else {
							// Remove the INR from his account for that bet
							inrBalance := memory.InrBalance[bet.UserId]
							inrBalance.Locked -= (100 - price) * bet.Quantity
							memory.InrBalance[bet.UserId] = inrBalance

							// Add the stock to his stockBalance
							stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
							stockBalance.Quantity += bet.Quantity
							memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance
						}
						// TODO: Delete order
						// delete(orderDetails.Orders, user)
					} else {
						// TODO: Free the quantity
						bet.Quantity -= quantity
						quantity = 0
						// Disolve the transactions
						if bet.TransactionType == memory.Sell {
							// Return the INR to his account for that bet
							inrBalance := memory.InrBalance[bet.UserId]
							inrBalance.Quantity += (100 - price) * quantity
							memory.InrBalance[bet.UserId] = inrBalance

							// Remove the stock from his stockBalance
							stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
							stockBalance.Locked -= quantity
							memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance

						} else {
							// Remove the INR from his account for that bet
							inrBalance := memory.InrBalance[bet.UserId]
							inrBalance.Locked -= (100 - price) * quantity
							memory.InrBalance[bet.UserId] = inrBalance

							// Add the stock to his stockBalance
							stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
							stockBalance.Quantity += quantity
							memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance
						}
						memory.BetBook[orderId] = bet
					}
				}
				(*oppositeSide)[price] = orderDetails
			}
			if quantity == 0 {
				break
			}
		}
	} else { //TODO: Do from here
		//NOTE: Here currentSide and oppositeSide is opposite in case of sell And price is also 100 - price
		addToOrderBook := func(addQuantity int) {
			if orderDetails, exists := (*currentSide)[price]; !exists {
				(*currentSide)[price] = memory.OrderDetails{
					Total:  addQuantity,
					Orders: []string{uuid.NewString()},
				}
			} else {
				orderDetails.Total += addQuantity
				orderDetails.Orders = append(orderDetails.Orders, uuid.NewString())
				// orderDetails.Orders[user] += addQuantity
				(*currentSide)[price] = orderDetails
			}
		}

		if orderDetails, exists := (*oppositeSide)[100-price]; !exists {
			addToOrderBook(quantity)
			if transactionType == memory.Sell {
				stockBalance := memory.StockBalance[symbol][user][side]
				stockBalance.Quantity -= quantity
				stockBalance.Locked += quantity
				memory.StockBalance[symbol][user][side] = stockBalance
			} else {
				userInrBalance := memory.InrBalance[user]
				userInrBalance.Quantity -= (price) * quantity
				userInrBalance.Locked += (price) * quantity
				memory.InrBalance[user] = userInrBalance
			}
		} else {
			// TODO: Same as above If function (Make it into a single fnx)
			if orderDetails.Total <= quantity {
				quantity -= orderDetails.Total
				if transactionType == memory.Sell {
					userInrBalance := memory.InrBalance[user]
					userInrBalance.Quantity += (100 - price) * orderDetails.Total
					memory.InrBalance[user] = userInrBalance

					stockBalance := memory.StockBalance[symbol][user][side]
					stockBalance.Quantity -= orderDetails.Total
					memory.StockBalance[symbol][user][side] = stockBalance
				} else {
					userInrBalance := memory.InrBalance[user]
					userInrBalance.Quantity -= (price) * orderDetails.Total
					memory.InrBalance[user] = userInrBalance

					stockBalance := memory.StockBalance[symbol][user][side]
					stockBalance.Quantity += orderDetails.Total
					memory.StockBalance[symbol][user][side] = stockBalance
				}
				delete(*oppositeSide, 100-price)
				for _, orderId := range orderDetails.Orders {
					// Remove the order from orderBook
					bet := memory.BetBook[orderId]
					delete(memory.BetBook, orderId)
					if bet.TransactionType == memory.Sell {
						// Return the INR to his account for that bet
						inrBalance := memory.InrBalance[bet.UserId]
						inrBalance.Quantity += (100 - price) * bet.Quantity
						memory.InrBalance[bet.UserId] = inrBalance

						// Remove the stock from his stockBalance
						stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
						stockBalance.Locked -= bet.Quantity
						memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance

					} else {
						// Remove the INR from his account for that bet
						inrBalance := memory.InrBalance[bet.UserId]
						inrBalance.Locked -= (100 - price) * bet.Quantity
						memory.InrBalance[bet.UserId] = inrBalance

						// Add the stock to his stockBalance
						stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
						stockBalance.Quantity += bet.Quantity
						memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance
					}
				}
			} else {
				orderDetails.Total -= quantity
				if transactionType == memory.Sell {
					userInrBalance := memory.InrBalance[user]
					userInrBalance.Quantity += (100 - price) * quantity
					memory.InrBalance[user] = userInrBalance

					stockBalance := memory.StockBalance[symbol][user][side]
					stockBalance.Quantity -= quantity
					memory.StockBalance[symbol][user][side] = stockBalance
				} else {
					userInrBalance := memory.InrBalance[user]
					userInrBalance.Quantity -= (price) * quantity
					memory.InrBalance[user] = userInrBalance

					stockBalance := memory.StockBalance[symbol][user][side]
					stockBalance.Quantity += quantity
					memory.StockBalance[symbol][user][side] = stockBalance
				}
				for _, orderId := range orderDetails.Orders {
					bet := memory.BetBook[orderId]
					if bet.Quantity <= quantity {
						quantity -= bet.Quantity
						// Disolve the transactions
						delete(memory.BetBook, orderId)
						if bet.TransactionType == memory.Sell {
							// Return the INR to his account for that bet
							inrBalance := memory.InrBalance[bet.UserId]
							inrBalance.Quantity += (100 - price) * bet.Quantity
							memory.InrBalance[bet.UserId] = inrBalance

							// Remove the stock from his stockBalance
							stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
							stockBalance.Locked -= bet.Quantity
							memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance

						} else {
							// Remove the INR from his account for that bet
							inrBalance := memory.InrBalance[bet.UserId]
							inrBalance.Locked -= (100 - price) * bet.Quantity
							memory.InrBalance[bet.UserId] = inrBalance

							// Add the stock to his stockBalance
							stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
							stockBalance.Quantity += bet.Quantity
							memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance
						}
						// TODO: Delete order
						// delete(orderDetails.Orders, user)
					} else {
						// TODO: Free the quantity
						bet.Quantity -= quantity
						quantity = 0
						// Disolve the transactions
						if bet.TransactionType == memory.Sell {
							// Return the INR to his account for that bet
							inrBalance := memory.InrBalance[bet.UserId]
							inrBalance.Quantity += (100 - price) * quantity
							memory.InrBalance[bet.UserId] = inrBalance

							// Remove the stock from his stockBalance
							stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
							stockBalance.Locked -= quantity
							memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance

						} else {
							// Remove the INR from his account for that bet
							inrBalance := memory.InrBalance[bet.UserId]
							inrBalance.Locked -= (100 - price) * quantity
							memory.InrBalance[bet.UserId] = inrBalance

							// Add the stock to his stockBalance
							if _, ok := memory.StockBalance[symbol]; !ok {
								memory.StockBalance[symbol] = make(map[string]map[memory.Yes_no]memory.Balance)
							}
							
							// Ensure the symbol map has the user ID map
							if _, ok := memory.StockBalance[symbol][bet.UserId]; !ok {
								memory.StockBalance[symbol][bet.UserId] = make(map[memory.Yes_no]memory.Balance)
							}
							stockBalance := memory.StockBalance[symbol][bet.UserId][bet.Side]
							stockBalance.Quantity += quantity
							memory.StockBalance[symbol][bet.UserId][bet.Side] = stockBalance
						}
						memory.BetBook[orderId] = bet
					}
				}
				(*oppositeSide)[100-price] = orderDetails
			}
			if quantity > 0 {
				addToOrderBook(quantity)
			}
		}
	}
}

func buyStock(symbol string, orderSide memory.Yes_no, price int, user string, quantity int, orderType memory.OrderType) {
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

	order(symbol, currentSide, oppositeSide, price, user, quantity, orderType, memory.Buy, orderSide)

	memory.OrderBook[symbol] = stockBook
}

func sellStock(symbol string, orderSide memory.Yes_no, price int, user string, quantity int, orderType memory.OrderType) {
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

	order(symbol, oppositeSide, currentSide, 100-price, user, quantity, orderType, memory.Sell, orderSide)
	memory.OrderBook[symbol] = stockBook
}
