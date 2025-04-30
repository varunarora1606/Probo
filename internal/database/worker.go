package database

import (
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/varunarora1606/Probo/internal/models"
	"github.com/varunarora1606/Probo/internal/types"
	"golang.org/x/text/cases"
)

var batchSize int = 20;

func Worker() {
	for {
		var openDeltas []types.Delta = []types.Delta{};
		var updateDeltas []types.Delta = []types.Delta{};
		var matchedDeltas []types.Delta = []types.Delta{};
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
			openDeltas = append(openDeltas, delta)
		case "update":
			updateDeltas = append(updateDeltas, delta)
		case "matched":
			matchedDeltas = append(matchedDeltas, delta)
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
				openDeltas = append(openDeltas, delta)
			case "update":
				updateDeltas = append(updateDeltas, delta)
			case "matched":
				matchedDeltas = append(matchedDeltas, delta)
			default:
				fmt.Println("Default delta:", delta.Msg)
			}
		}

		if len(openDeltas) > 0 {
			
		}


	}
}