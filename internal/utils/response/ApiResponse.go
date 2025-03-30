package response

// import (
// 	"encoding/json"
// 	"net/http"

// 	"github.com/go-playground/validator/v10"
// )

// func ApiResponse(w http.ResponseWriter, status int, message string, data any) error {

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)

// 	response := map[string]any {
// 		"message": message,
// 	}

// 	if status < 400 {
// 		response["data"] = data
// 	} else {
// 		response["error"] = data
// 	}

// 	return json.NewEncoder(w).Encode(response)
// }

// func ValidateErr(errs validator.ValidationErrors) {
// 	var errMsg []string

// 	for _, err := range errs {

// 	}
// }
