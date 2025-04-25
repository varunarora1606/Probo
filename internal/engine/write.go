package engine

import (
	"fmt"

	"github.com/varunarora1606/Probo/internal/memory"
)

func CreateMarket(symbol string) error {
	memory.OrderBook.Mu.Lock()
	defer memory.OrderBook.Mu.Unlock()

	if _, exist := memory.OrderBook.Data[symbol]; exist {
		return fmt.Errorf("symbol's market already exists")
	}
	memory.OrderBook.Data[symbol] = memory.StockBook{
		Yes: make(map[int]memory.OrderDetails),
		No:  make(map[int]memory.OrderDetails),
	}
	
	return nil
}

func OnRampInr(userId string, quantity int) memory.Balance {
	memory.InrBalance.Mu.Lock()
	defer memory.InrBalance.Mu.Unlock()

	userBalance := memory.InrBalance.Data[userId]
	userBalance.Quantity += quantity;
	memory.InrBalance.Data[userId] = userBalance

	return userBalance
}