package engine

import (
	"encoding/json"
	"log"

	"github.com/varunarora1606/Probo/internal/database"
	"github.com/varunarora1606/Probo/internal/memory"
)

type Task struct {
	ApiId string
	Fnx string
	UserId    string          
	Symbol    string          
	Quantity  int             
	Price     int             
	StockSide memory.Side     
	StockType memory.OrderType
	TransactionType memory.TransactionType
}

type Output struct {
	ForWs bool
	ApiId string
	Err string
	Market memory.StockBook
	Markets map[string]memory.StockBook
	InrBalance memory.Balance
	StockBalance map[string]map[memory.Side]memory.Balance
	Deltas []memory.Delta
	Trade memory.Trade
}

func Worker() {
	for {
		result, err := database.RClient.BRPop(database.Ctx, 0, "input").Result()
		if err != nil {
			log.Printf("Error during BRPOP on 'input': %v", err)
			continue
		}

		data := result[1]
		var task Task

		if err := json.Unmarshal([]byte(data), &task); err != nil {
			log.Printf("Error during unmarshalling of %s in 'input': %v", data, err)
			continue
		}

		var output Output;

		switch task.Fnx {
		case "order_engine":
			trade, err := OrderEngine(task.Symbol, task.StockSide, task.Price, task.UserId, task.Quantity, task.StockType, task.TransactionType)
			output.Trade = trade
			if err != nil {
				output.Err = err.Error()
			}
		case "create_market":
			err := CreateMarket(task.Symbol)
			if err != nil {
				output.Err = err.Error()
			}
		case "on_ramp_inr":
			balance := OnRampInr(task.UserId, task.Quantity)
			output.InrBalance = balance
		case "get_market":
			market, err := GetMarket(task.Symbol)
			output.Market = market
			if err != nil {
				output.Err = err.Error()
			}
		case "get_markets":
			markets := GetMarkets()
			output.Markets =  markets
		case "get_inr_balance":
			balance := GetInrBalance(task.UserId)
			output.InrBalance = balance
		case "get_stock_balance":
			balance := GetStockBalance(task.UserId)
			output.StockBalance = balance
		case "get_me":
			inrBalance, stockBalance := GetMe(task.UserId)
			output.InrBalance = inrBalance
			output.StockBalance = stockBalance
		}

		outputJson, err := json.Marshal(output);
		if  err != nil {
			log.Printf("Error during marshalling of %v in 'input': %v", output, err)
			continue
		}

		err = database.RClient.LPush(database.Ctx, "output", outputJson).Err()
		if err != nil {
			log.Printf("Error during LPUSH on 'output': %v", err)
			continue
		}
		
	}
}