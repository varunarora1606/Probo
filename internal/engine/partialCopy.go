package engine

import (
	"fmt"
	"time"

	"github.com/varunarora1606/Probo/internal/types"
)

func partialCopyStockBook(symbol string, original map[string]types.StockBook) (types.StockBook, error) {
	stockBook, exist := original[symbol]
	if !exist {
		return types.StockBook{}, fmt.Errorf("no such symbol exists")
	}

	copy := types.StockBook{
		Yes: make(map[int]types.OrderDetails),
		No: make(map[int]types.OrderDetails),
	}

	for price, orderDetails := range stockBook.Yes {
		newOrderDetails := types.OrderDetails{}
		newOrderDetails.Orders = append(newOrderDetails.Orders, orderDetails.Orders...)
		newOrderDetails.Total = orderDetails.Total
		copy.Yes[price] = newOrderDetails
	}

	for price, orderDetails := range stockBook.No {
		newOrderDetails := types.OrderDetails{}
		newOrderDetails.Orders = append(newOrderDetails.Orders, orderDetails.Orders...)
		newOrderDetails.Total = orderDetails.Total
		copy.No[price] = newOrderDetails
	}

	return copy, nil

}

func partialCopyInrBalance(original map[string]types.Balance) map[string]types.Balance {
	copy := make(map[string]types.Balance)
	for userId, balance := range original {
		copy[userId] = types.Balance{
			Quantity: balance.Quantity,
			Locked: balance.Locked,
		}
	}
	return copy
}

func partialCopyStockBalance(original map[string]map[string]map[types.Side]types.Balance) map[string]map[string]map[types.Side]types.Balance {
	copy := make(map[string]map[string]map[types.Side]types.Balance)

	for userId, symbolDetails := range original {
		newSymbolDetails := make(map[string]map[types.Side]types.Balance)
		for symbol, sideDetails := range symbolDetails {
			newSideDetails := make(map[types.Side]types.Balance)
			for side, balance := range sideDetails {
				newSideDetails[side] = types.Balance{
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

func partialCopyBetBook(original map[string]types.BetDetails) map[string]types.BetDetails {
	copy := make(map[string]types.BetDetails)

	for betId, betDetails := range original {
		copy[betId] = types.BetDetails{
			UserId: betDetails.UserId,
			Price: betDetails.Price,
			Quantity: betDetails.Quantity,
			Side: betDetails.Side,
			TransactionType: betDetails.TransactionType,
		}
	}

	return copy
}

func partialCopySymbolBook(symbol string, original map[string]types.SymbolBook) (types.SymbolBook, error) {
	symbolBook, exist := original[symbol]
	if !exist {
		return types.SymbolBook{}, fmt.Errorf("no such symbol exists")
	}

	if symbolBook.EndTime < time.Now().UnixNano() {
		return types.SymbolBook{}, fmt.Errorf("market have expired")
	}


	return types.SymbolBook{
		Question: symbolBook.Question,
		EndTime: symbolBook.EndTime,
		YesClosing: symbolBook.YesClosing,
		Volume: symbolBook.Volume,
	}, nil
}