package engine

import (
	"fmt"

	"github.com/varunarora1606/Probo/internal/memory"
	"github.com/varunarora1606/Probo/internal/types"
)

func CreateMarket(symbol string, question string, endTime int64) error {
	memory.OrderBook.Mu.Lock()
	memory.MarketBook.Mu.Lock()
	defer memory.MarketBook.Mu.Unlock()
	defer memory.OrderBook.Mu.Unlock()

	_, exist := memory.MarketBook.Data[symbol]
	if exist {
		return fmt.Errorf("symbol's market already exists")
	}

	if _, exist := memory.OrderBook.Data[symbol]; exist {
		return fmt.Errorf("symbol's market already exists")
	}

	memory.MarketBook.Data[symbol] = types.SymbolBook{
		Question: question,
		EndTime:  endTime,
	}

	memory.OrderBook.Data[symbol] = types.StockBook{
		Yes: make(map[int]types.OrderDetails),
		No:  make(map[int]types.OrderDetails),
	}

	return nil
}

func OnRampInr(userId string, quantity int) types.Balance {
	memory.InrBalance.Mu.Lock()
	defer memory.InrBalance.Mu.Unlock()

	userBalance := memory.InrBalance.Data[userId]
	userBalance.Quantity += quantity
	memory.InrBalance.Data[userId] = userBalance

	return userBalance
}
