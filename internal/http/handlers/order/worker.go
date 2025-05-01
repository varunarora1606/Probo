package order

import (
	"encoding/json"
	"fmt"

	"github.com/varunarora1606/Probo/internal/database"
	"github.com/varunarora1606/Probo/internal/types"
)

func worker(input types.Input) (types.Output, error) {

	inputJson, err := json.Marshal(input)
	if err != nil {
		return types.Output{}, fmt.Errorf("error during marshalling: %v", err)
	}

	err = database.RClient.LPush(database.Ctx, "input", inputJson).Err()
	if err != nil {
		return types.Output{}, fmt.Errorf("error during LPUSH on 'input': %v", err)
	}

	result, err := database.RClient.BRPop(database.Ctx, 0, "output").Result()
	if err != nil {
		return types.Output{}, fmt.Errorf("error during BRPOP on 'output': %v", err)
	}

	data := result[1]
	var output types.Output

	if err := json.Unmarshal([]byte(data), &output); err != nil {
		return types.Output{}, fmt.Errorf("error during unmarshalling of %s in 'output': %v", data, err)
	}

	return output, nil
}
