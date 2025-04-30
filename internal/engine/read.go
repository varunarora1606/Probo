package engine

import (
	"fmt"

	"github.com/varunarora1606/Probo/internal/memory"
	"github.com/varunarora1606/Probo/internal/types"
)

func GetMarket(symbol string) (types.SymbolBook, error) {
	memory.MarketBook.Mu.RLock()
	defer memory.MarketBook.Mu.RUnlock()

	if symbolBook, exist := memory.MarketBook.Data[symbol]; !exist {
		return types.SymbolBook{}, fmt.Errorf("this market does not exist")
	} else {
		return symbolBook, nil
	}
}

func GetMarkets() (map[string]types.SymbolBook) {
	memory.MarketBook.Mu.RLock()
	defer memory.MarketBook.Mu.RUnlock()
	
	return memory.MarketBook.Data
}

func GetOrderBook(symbol string) (types.StockBook, error) {
	memory.OrderBook.Mu.RLock()
	defer memory.OrderBook.Mu.RUnlock()

	if stockBook, exist := memory.OrderBook.Data[symbol]; !exist {
		return types.StockBook{}, fmt.Errorf("this market does not exist")
	} else {
		return stockBook, nil
	}
}

func GetInrBalance(userId string) (types.Balance) {
	memory.InrBalance.Mu.RLock()
	defer memory.InrBalance.Mu.RUnlock()
	
	if balance, exist := memory.InrBalance.Data[userId]; !exist {
		return types.Balance{
			Quantity: 0,
			Locked:   0,
		}
	} else {
		return balance
	}
}

func GetStockBalance(userId string) (map[string]map[types.Side]types.Balance) {
	memory.StockBalance.Mu.RLock()
	defer memory.StockBalance.Mu.RUnlock()

	if balance, exist := memory.StockBalance.Data[userId]; !exist {
		return nil
	} else {
		return balance
	}
}

func GetMe(userId string) (types.Balance, []types.PortfolioItem) {
	memory.StockBalance.Mu.RLock()
	memory.InrBalance.Mu.RLock()
	memory.MarketBook.Mu.RLock()
	defer memory.MarketBook.Mu.RUnlock()
	defer memory.InrBalance.Mu.RUnlock()
	defer memory.StockBalance.Mu.RUnlock()

	inrBalance, exist := memory.InrBalance.Data[userId];
	if !exist {
		inrBalance = types.Balance{
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
		value := sideBalance[types.Yes].Quantity*yesClosing + sideBalance[types.No].Quantity*(100-yesClosing)
		portfolioItems = append(portfolioItems, types.PortfolioItem{
			Symbol: symbol,
			Quantity: sideBalance[types.Yes].Quantity + sideBalance[types.No].Quantity,
			Value: value,
		})
	}


	return inrBalance, portfolioItems
}