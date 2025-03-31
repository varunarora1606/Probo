package order

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type OrderDetails struct {
	Total  int            `json:"total"`
	Orders map[string]int `json:"orders"`
}
type StockBook struct {
	Yes map[int]OrderDetails `json:"yes"`
	No  map[int]OrderDetails `json:"no"`
}

var OrderBook = make(map[string]StockBook)

type yes_no string

const (
	Yes yes_no = "yes"
	No  yes_no = "no"
)

type buyHandlerReq struct {
	UserId      string `json:"userId" binding:"required"`
	StockSymbol string `json:"stockSymbol" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required"`
	Price       int    `json:"price" binding:"required"`
	StockType   string `json:"stockType" binding:"required"`
}
type sellHandlerReq struct {
	UserId      string `json:"userId" binding:"required"`
	StockSymbol string `json:"stockSymbol" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required"`
	StockType   string `json:"stockType" binding:"required"`
}

func buyStock(symbol string, orderSide yes_no, price int, user string, quantity int) {
	stockBook, exists := OrderBook[symbol]
	if !exists {
		// TODO: return error or do it in the handler function
		stockBook = StockBook{
			Yes: make(map[int]OrderDetails),
			No:  make(map[int]OrderDetails),
		}
		OrderBook[symbol] = stockBook
	}

	var oppositeSide, currentSide *map[int]OrderDetails

	if orderSide == Yes {
		oppositeSide = &stockBook.No
		currentSide = &stockBook.Yes
	} else {
		oppositeSide = &stockBook.Yes
		currentSide = &stockBook.No
	}

	addToOrderBook := func(addQuantity int) {
		if orderDetails, exists := (*currentSide)[price]; !exists {
			(*currentSide)[price] = OrderDetails{
				Total:  addQuantity,
				Orders: map[string]int{user: addQuantity},
			}
		} else {
			orderDetails.Total += addQuantity
			orderDetails.Orders[user] += addQuantity
			(*currentSide)[price] = orderDetails
		}
	}

	if orderDetails, exists := (*oppositeSide)[100-price]; !exists {
		addToOrderBook(quantity)
	} else {
		if orderDetails.Total <= quantity {
			quantity -= orderDetails.Total
			delete(*oppositeSide, 100-price)
			addToOrderBook(quantity)
		} else {
			orderDetails.Total -= quantity
			for user := range orderDetails.Orders {
				if quantity == 0 {
					return
				}
				if orderDetails.Orders[user] <= quantity {
					quantity -= orderDetails.Orders[user]
					delete(orderDetails.Orders, user)
				} else {
					orderDetails.Orders[user] -= quantity
				}
			}
			(*oppositeSide)[100-price] = orderDetails
		}
	}

	OrderBook[symbol] = stockBook
	fmt.Println(OrderBook)
}

func sellStock(symbol string, orderSide yes_no, user string, quantity int) {
	stockBook, exists := OrderBook[symbol]
	if !exists {
		// TODO: return error or do it in the handler function
		stockBook = StockBook{
			Yes: make(map[int]OrderDetails),
			No:  make(map[int]OrderDetails),
		}
		OrderBook[symbol] = stockBook
	}

	var currentSide *map[int]OrderDetails

	if orderSide == Yes {
		currentSide = &stockBook.Yes
	} else {
		currentSide = &stockBook.No
	}

	prices := make([]int, 0, len(*currentSide))
	for price := range *currentSide {
		prices = append(prices, price)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(prices)))

	for _, price := range prices {
		orderDetails := (*currentSide)[price]
		if orderDetails.Total <= quantity {
			quantity -= orderDetails.Total
			delete(*currentSide, price)
		} else {
			orderDetails.Total -= quantity
			for user := range orderDetails.Orders {
				if quantity == 0 {
					break
				}
				if orderDetails.Orders[user] <= quantity {
					quantity -= orderDetails.Orders[user]
					delete(orderDetails.Orders, user)
				} else {
					orderDetails.Orders[user] -= quantity
				}
			}
			(*currentSide)[price] = orderDetails
		}
		if quantity == 0 {break}
	}

}

func BuyHandler(c *gin.Context) {
	var req buyHandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	buyStock(req.StockSymbol, yes_no(req.StockType), req.Price, req.UserId, req.Quantity)

	c.JSON(http.StatusCreated, gin.H{"message": "Stock added successfully", "data": OrderBook})
}

func SellHandler(c *gin.Context) {
	var req sellHandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	sellStock(req.StockSymbol, yes_no(req.StockType), req.UserId, req.Quantity)

	c.JSON(http.StatusCreated, gin.H{"message": "Stock sold successfully", "data": OrderBook})
}
