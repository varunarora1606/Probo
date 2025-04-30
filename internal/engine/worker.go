package engine

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/varunarora1606/Probo/internal/database"
	"github.com/varunarora1606/Probo/internal/types"
)

func Worker() {
	for {
		result, err := database.RClient.BRPop(database.Ctx, 0, "input").Result()
		if err != nil {
			log.Printf("Error during BRPOP on 'input': %v", err)
			continue
		}

		data := result[1]
		var input types.Input

		if err := json.Unmarshal([]byte(data), &input); err != nil {
			log.Printf("Error during unmarshalling of %s in 'input': %v", data, err)
			continue
		}

		var output types.Output;
		var orderEvents []types.Delta;

		switch input.Fnx {
		case "order_engine":
			trade, deltas, err := OrderEngine(input.Symbol, input.StockSide, input.Price, input.UserId, input.Quantity, input.StockType, input.TransactionType)
			output.Trade = trade
			if err != nil {
				output.Err = err.Error()
				fmt.Println(err.Error())
			} else {
				orderEvents = deltas
			}
		case "create_market":
			err := CreateMarket(input.Symbol, input.Question, input.EndTime)
			if err != nil {
				output.Err = err.Error()
			}
		case "on_ramp_inr":
			balance := OnRampInr(input.UserId, input.Quantity)
			output.InrBalance = balance
		case "get_market":
			market, err := GetMarket(input.Symbol)
			output.Market = market
			if err != nil {
				output.Err = err.Error()
			}
		case "get_markets":
			markets := GetMarkets()
			output.Markets =  markets
		case "get_orderbook":
			stockBook, err := GetOrderBook(input.Symbol)
			output.StockBook = stockBook
			if err != nil {
				output.Err = err.Error()
			}
		case "get_inr_balance":
			balance := GetInrBalance(input.UserId)
			output.InrBalance = balance
		case "get_stock_balance":
			balance := GetStockBalance(input.UserId)
			output.StockBalance = balance
		case "get_me":
			inrBalance, portfolioItems := GetMe(input.UserId)
			output.InrBalance = inrBalance
			output.PortfolioItems = portfolioItems
		}

		outputJson, err := json.Marshal(output);
		if  err != nil {
			log.Printf("Error during marshalling of %v in 'output': %v", output, err)
			continue
		}

		err = database.RClient.LPush(database.Ctx, "output", outputJson).Err()
		if err != nil {
			log.Printf("Error during LPUSH on 'output': %v", err)
			continue
		}
		
		if len(orderEvents) > 0 {
			var redisPayload []interface{}
			for _, event := range orderEvents {
				jsonEvent, err := json.Marshal(event)
				if err != nil {
					log.Printf("Error marshalling order event: %v", err)
					continue
				}
				redisPayload = append(redisPayload, jsonEvent)
			}
	
			err = database.RClient.LPush(database.Ctx, "order_events", redisPayload...).Err()
			if err != nil {
				log.Printf("Error during LPUSH on 'order_events': %v", err)
				continue
			}
		}
	}
}