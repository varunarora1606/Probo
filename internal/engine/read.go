package engine

import (
	"fmt"

	"github.com/varunarora1606/Probo/internal/memory"
	"github.com/varunarora1606/Probo/internal/types"
)

func GetMarket(symbol string) (memory.SymbolBook, error) {
	memory.MarketBook.Mu.RLock()
	defer memory.MarketBook.Mu.RUnlock()

	if symbolBook, exist := memory.MarketBook.Data[symbol]; !exist {
		return memory.SymbolBook{}, fmt.Errorf("this market does not exist")
	} else {
		return symbolBook, nil
	}
}

func GetMarkets() (map[string]memory.SymbolBook) {
	memory.MarketBook.Mu.RLock()
	defer memory.MarketBook.Mu.RUnlock()
	
	return memory.MarketBook.Data
}

func GetOrderBook(symbol string) (memory.StockBook, error) {
	memory.OrderBook.Mu.RLock()
	defer memory.OrderBook.Mu.RUnlock()

	if stockBook, exist := memory.OrderBook.Data[symbol]; !exist {
		return memory.StockBook{}, fmt.Errorf("this market does not exist")
	} else {
		return stockBook, nil
	}
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

func GetMe(userId string) (memory.Balance, []types.PortfolioItem) {
	memory.StockBalance.Mu.RLock()
	memory.InrBalance.Mu.RLock()
	memory.MarketBook.Mu.RLock()
	defer memory.MarketBook.Mu.RUnlock()
	defer memory.InrBalance.Mu.RUnlock()
	defer memory.StockBalance.Mu.RUnlock()

	inrBalance, exist := memory.InrBalance.Data[userId];
	if !exist {
		inrBalance = memory.Balance{
			Quantity: 0,
			Locked:   0,
		}
	}

	portfolioItems := []types.PortfolioItem{}

	stockBalance, exist := memory.StockBalance.Data[userId]; 
	if !exist {
		stockBalance = nil
	}

	for symbol, sideBalance := range stockBalance {
		yesClosing := memory.MarketBook.Data[symbol].YesClosing
		value := sideBalance[memory.Yes].Quantity*yesClosing + sideBalance[memory.No].Quantity*(100-yesClosing)
		portfolioItems = append(portfolioItems, types.PortfolioItem{
			Symbol: symbol,
			Quantity: sideBalance[memory.Yes].Quantity + sideBalance[memory.No].Quantity,
			Value: value,
		})
	}


	return inrBalance, portfolioItems
}