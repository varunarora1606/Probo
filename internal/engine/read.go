package engine

import (
	"fmt"

	"github.com/varunarora1606/Probo/internal/memory"
)

func GetMarket(symbol string) (memory.StockBook, error) {
	memory.OrderBook.Mu.RLock()
	defer memory.OrderBook.Mu.RUnlock()

	if marketBook, exist := memory.OrderBook.Data[symbol]; !exist {
		return memory.StockBook{}, fmt.Errorf("this market does not exist")
	} else {
		return marketBook, nil
	}
}

func GetMarkets() (map[string]memory.StockBook) {
	memory.OrderBook.Mu.RLock()
	defer memory.OrderBook.Mu.RUnlock()

	return memory.OrderBook.Data
}

func GetInrBalance(userId string) (memory.Balance) {
	memory.InrBalance.Mu.RLock()
	defer memory.InrBalance.Mu.RUnlock()

	if balance, exist := memory.InrBalance.Data[userId]; !exist {
		return memory.Balance{
			Quantity: 0,
			Locked:   0,
		}
	} else {
		return balance
	}
}

func GetStockBalance(userId string) (map[string]map[memory.Side]memory.Balance) {
	memory.StockBalance.Mu.RLock()
	defer memory.StockBalance.Mu.RUnlock()

	if balance, exist := memory.StockBalance.Data[userId]; !exist {
		return nil
	} else {
		return balance
	}
}

func GetMe(userId string) (memory.Balance, map[string]map[memory.Side]memory.Balance) {
	memory.StockBalance.Mu.RLock()
	memory.InrBalance.Mu.RLock()
	defer memory.StockBalance.Mu.RUnlock()
	defer memory.InrBalance.Mu.RUnlock()

	inrBalance, exist := memory.InrBalance.Data[userId];
	if !exist {
		inrBalance = memory.Balance{
			Quantity: 0,
			Locked:   0,
		}
	}

	stockBalance, exist := memory.StockBalance.Data[userId]; 
	if !exist {
		stockBalance = nil
	}

	return inrBalance, stockBalance
}