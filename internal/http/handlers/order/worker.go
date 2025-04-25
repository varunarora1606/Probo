package order

import (
	"encoding/json"
	"fmt"

	"github.com/varunarora1606/Probo/internal/database"
	"github.com/varunarora1606/Probo/internal/memory"
)

type Input struct {
	ApiId           string
	Fnx             string
	UserId          string
	Symbol          string
	Quantity        int
	Price           int
	StockSide       memory.Side
	StockType       memory.OrderType
	TransactionType memory.TransactionType
}

type Output struct {
	ForWs bool
	ApiId string
	Err error
	Market memory.StockBook
	Markets map[string]memory.StockBook
	InrBalance memory.Balance
	StockBalance map[string]map[memory.Side]memory.Balance
	Deltas []memory.Delta
}

func Worker(input Input) (Output, error) {

	inputJson, err := json.Marshal(input);
	if  err != nil {
		return Output{}, fmt.Errorf("Error during marshalling: %v", err)
	}

	err = database.RClient.LPush(database.Ctx, "input", inputJson).Err()
	if err != nil {
		return Output{}, fmt.Errorf("Error during LPUSH on 'input': %v", err)
	}

	result, err := database.RClient.BRPop(database.Ctx, 0, "output").Result()
	if err != nil {
		return Output{}, fmt.Errorf("Error during BRPOP on 'output': %v", err)
	}

	data := result[1]
	var output Output

	if err := json.Unmarshal([]byte(data), &output); err != nil {
		return Output{}, fmt.Errorf("Error during unmarshalling of %s in 'output': %v", data, err)
	}

	return output, nil

}