package order

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/varunarora1606/Booking-App-Go/internal/memory"
)

type HandlerReq struct {
	UserId      string `json:"userId" binding:"required"`
	StockSymbol string `json:"stockSymbol" binding:"required"` //symbol
	Quantity    int    `json:"quantity" binding:"required"`
	Price       int    `json:"price"`
	StockSide   memory.Yes_no `json:"stockSide" binding:"required,oneof=yes no"` //side
	StockType memory.OrderType `json:"stockType" binding:"required,oneof=market limit"` //ordertype
}

func BuyHandler(c *gin.Context) {
	var req HandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	buyStock(req.StockSymbol, memory.Yes_no(req.StockSide), req.Price, req.UserId, req.Quantity, memory.OrderType(req.StockType))

	c.JSON(http.StatusCreated, gin.H{"message": "Stock added successfully", "data": memory.OrderBook})
}

func SellHandler(c *gin.Context) {
	var req HandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	sellStock(req.StockSymbol, memory.Yes_no(req.StockSide), req.Price, req.UserId, req.Quantity, memory.OrderType(req.StockType))

	c.JSON(http.StatusCreated, gin.H{"message": "Stock sold successfully", "data": memory.OrderBook})
}
