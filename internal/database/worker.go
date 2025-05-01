package database

import (
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/varunarora1606/Probo/internal/models"
	"github.com/varunarora1606/Probo/internal/types"
)

var batchSize int = 20

func Worker() {
	for {
		var openOrders []types.Order = []types.Order{}
		var updateOrders []types.Order = []types.Order{}
		var matchedOrders []types.Order = []types.Order{}
		result, err := RClient.BRPop(Ctx, 0, "order_events").Result()
		if err != nil {
			fmt.Println("error during BRPOP on 'order_events':", err.Error())
			continue
		}
		data := result[1]
		var delta types.Delta
		if err := json.Unmarshal([]byte(data), &delta); err != nil {
			fmt.Printf("error during unmarshalling of %s in 'output': %v", data, err)
			continue
		}

		switch delta.Msg {
		case "open":
			openOrders = append(openOrders, delta.Order)
		case "update":
			updateOrders = append(updateOrders, delta.Order)
		case "matched":
			matchedOrders = append(matchedOrders, delta.Order)
		default:
			fmt.Println("Default delta:", delta.Msg)
		}

		for range batchSize {
			data, err = RClient.RPop(Ctx, "order_events").Result()
			if err == redis.Nil {
				break
			} else if err != nil {
				fmt.Println("RPOP error:", err)
				break
			}
			var delta types.Delta
			if err := json.Unmarshal([]byte(data), &delta); err != nil {
				fmt.Printf("error during unmarshalling of %s in 'order_events': %v", data, err)
				continue
			}
			switch delta.Msg {
			case "open":
				openOrders = append(openOrders, delta.Order)
			case "update":
				updateOrders = append(updateOrders, delta.Order)
			case "matched":
				matchedOrders = append(matchedOrders, delta.Order)
			default:
				fmt.Println("Default delta:", delta.Msg)
			}
		}

		if len(openOrders) > 0 {
			openOrders := convertTypesInModels(openOrders)
			if err := DB.Create(&openOrders).Error; err != nil {
				fmt.Println("Error inserting orders:", err)
			} else {
				fmt.Println("Bulk insert successful")
			}
		}
		if len(updateOrders) > 0 {
			updateOrders := convertTypesInModels(updateOrders)
			for _, order := range updateOrders {
				// TODO: Check if the EventId are different before updating so that any 2 worker might not update it twice.
				if err := DB.Model(&models.Order{}).Where("bet_id = ?", order.BetId).Updates(models.Order{
					Quantity: order.Quantity,
					EventId:  order.EventId,
				}).Error; err != nil {
					fmt.Println("Error inserting orders:", err)
				} else {
					fmt.Println("update successful")
				}
			}
		}
		if len(matchedOrders) > 0 {
			matchedOrders := convertTypesInModels(matchedOrders)
			matchedBetIds := []string{}
			for _, orders := range matchedOrders {
				matchedBetIds = append(matchedBetIds, orders.BetId)
			}
			if err := DB.Where("bet_id IN ?", matchedBetIds).Delete(&models.Order{}).Error; err != nil {
				fmt.Println("Error deleting orders:", err)
			} else {
				fmt.Println("Bulk delete successful")
			}
		}
	}
}

func convertTypesInModels(typesOrders []types.Order) []models.Order {
	var modelOrders []models.Order
	for _, o := range typesOrders {
		modelOrders = append(modelOrders, models.Order{
			BetId:           o.BetId,
			EventId:         o.EventId,
			UserID:          o.UserID,
			Symbol:          o.Symbol,
			Side:            o.Side,
			Price:           o.Price,
			Quantity:        o.Quantity,
			TransactionType: o.TransactionType,
			CreatedAt:       o.CreatedAt,
		})
	}
	return modelOrders
}
