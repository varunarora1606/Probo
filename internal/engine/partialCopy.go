package engine

import (
	"fmt"
	"time"

	"github.com/varunarora1606/Probo/internal/memory"
)

func partialCopyStockBook(symbol string, original map[string]memory.StockBook) (memory.StockBook, error) {
	stockBook, exist := original[symbol]
	if !exist {
		return memory.StockBook{}, fmt.Errorf("no such symbol exists")
	}

	copy := memory.StockBook{
		Yes: make(map[int]memory.OrderDetails),
		No: make(map[int]memory.OrderDetails),
	}

	for price, orderDetails := range stockBook.Yes {
		newOrderDetails := memory.OrderDetails{}
		newOrderDetails.Orders = append(newOrderDetails.Orders, orderDetails.Orders...)
		newOrderDetails.Total = orderDetails.Total
		copy.Yes[price] = newOrderDetails
	}

	for price, orderDetails := range stockBook.No {
		newOrderDetails := memory.OrderDetails{}
		newOrderDetails.Orders = append(newOrderDetails.Orders, orderDetails.Orders...)
		newOrderDetails.Total = orderDetails.Total
		copy.No[price] = newOrderDetails
	}

	return copy, nil

}

func partialCopyInrBalance(original map[string]memory.Balance) map[string]memory.Balance {
	copy := make(map[string]memory.Balance)
	for userId, balance := range original {
		copy[userId] = memory.Balance{
			Quantity: balance.Quantity,
			Locked: balance.Locked,
		}
	}
	return copy
}

func partialCopyStockBalance(original map[string]map[string]map[memory.Side]memory.Balance) map[string]map[string]map[memory.Side]memory.Balance {
	copy := make(map[string]map[string]map[memory.Side]memory.Balance)

	for userId, symbolDetails := range original {
		newSymbolDetails := make(map[string]map[memory.Side]memory.Balance)
		for symbol, sideDetails := range symbolDetails {
			newSideDetails := make(map[memory.Side]memory.Balance)
			for side, balance := range sideDetails {
				newSideDetails[side] = memory.Balance{
					Quantity: balance.Quantity,
					Locked: balance.Locked,
				}
			}
			newSymbolDetails[symbol] = newSideDetails
		}
		copy[userId] = newSymbolDetails
	}

	return copy
}

func partialCopyBetBook(original map[string]memory.BetDetails) map[string]memory.BetDetails {
	copy := make(map[string]memory.BetDetails)

	for betId, betDetails := range original {
		copy[betId] = memory.BetDetails{
			UserId: betDetails.UserId,
			Price: betDetails.Price,
			Quantity: betDetails.Quantity,
			Side: betDetails.Side,
			TransactionType: betDetails.TransactionType,
		}
	}

	return copy
}

func partialCopySymbolBook(symbol string, original map[string]memory.SymbolBook) (memory.SymbolBook, error) {
	symbolBook, exist := original[symbol]
	if !exist {
		return memory.SymbolBook{}, fmt.Errorf("no such symbol exists")
	}

	if symbolBook.EndTime < time.Now().UnixNano() {
		return memory.SymbolBook{}, fmt.Errorf("market have expired")
	}


	return memory.SymbolBook{
		Question: symbolBook.Question,
		EndTime: symbolBook.EndTime,
		YesClosing: symbolBook.YesClosing,
		Volume: symbolBook.Volume,
	}, nil
}