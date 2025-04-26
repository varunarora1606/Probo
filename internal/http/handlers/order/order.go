package order

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/varunarora1606/Probo/internal/memory"
)

type HandlerReq struct {
	UserId    string           `json:"userId" binding:"required"`
	Symbol    string           `json:"symbol" binding:"required"` //symbol
	Quantity  int              `json:"quantity" binding:"required,gt=0"` //Greater than 0 check
	Price     int              `json:"price" binding:"gt=0"` //Greater than 0 check only with limit
	StockSide memory.Side      `json:"stockSide" binding:"required,oneof=yes no"`       //side
	StockType memory.OrderType `json:"stockType" binding:"required,oneof=market limit"` //ordertype
}

type CreateMarketHandlerReq struct {
	UserId       string `json:"userId" binding:"required"`
	Question     string `json:"question" binding:"required"`
	EndTimeMilli int64  `json:"endTime" binding:"required,gt=0"`
	Symbol 		 string `json:"symbol" binding:"required"`
}

func BuyHandler(c *gin.Context) {
	var req HandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	result ,err := worker(Input{
		Fnx: "order_engine",
		Symbol: req.Symbol, 
		StockSide: memory.Side(req.StockSide), 
		Price: req.Price, 
		UserId: req.UserId, 
		Quantity: req.Quantity, 
		StockType: memory.OrderType(req.StockType), 
		TransactionType: memory.Buy,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	if req.StockType == memory.Market {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Buy request completed", 
			"data": gin.H{
				"trades": result.Trade.MicroTrades,
			},
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Buy request completed", "data": gin.H{
			"completed": result.Trade.TotalQuantity,
			"pending": req.Quantity - result.Trade.TotalQuantity,
			},
		})
	}
}

func SellHandler(c *gin.Context) {
	var req HandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	result ,err := worker(Input{
		Fnx: "order_engine",
		Symbol: req.Symbol, 
		StockSide: memory.Side(req.StockSide), 
		Price: req.Price, 
		UserId: req.UserId, 
		Quantity: req.Quantity, 
		StockType: memory.OrderType(req.StockType), 
		TransactionType: memory.Sell,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	fmt.Println(result.Trade)
	
	if req.StockType == memory.Market {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Sell request completed", 
			"data": gin.H{
				"trades": result.Trade.MicroTrades,
			},
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{"message": "Sell request completed", "data": gin.H{
			"completed": result.Trade.TotalQuantity,
			"pending": req.Quantity - result.Trade.TotalQuantity,
			},
		})
	}
}

func GetMarketHandler(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	result, err := worker(Input{
		Fnx: "get_market",
		Symbol: req.Symbol,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Market fetched successfully",
		"data":    gin.H{"market": result.Market},
	})
	
}

func GetMarketsHandler(c *gin.Context) {
	result, err := worker(Input{
		Fnx: "get_markets",
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Markets fetched successfully",
		"data":    gin.H{"markets": result.Markets},
	})
}

func CreateMarketHandler(c *gin.Context) {
	var req CreateMarketHandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	endTime := time.Unix(0, req.EndTimeMilli*int64(time.Millisecond))
	serverTime := time.Now()

	if endTime.Before(serverTime.Add(5 * time.Second)) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "EndTime should be in far future"})
		return
	}
	// TODO: Check and Add to DB

	result, err := worker(Input{
		Fnx: "create_market",
		Symbol: req.Symbol,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	// TODO: Return Db result.

	c.JSON(http.StatusOK, gin.H{
		"message": "Market created successfully",
	})
}

func OnRampInrHandler(c *gin.Context) {
	var req struct {
		UserId string `json:"userId" binding:"required"`
		Quantity int `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	// TODO: Add to DB

	result, err := worker(Input{
		Fnx: "on_ramp_inr",
		Quantity: req.Quantity,
		UserId: req.UserId,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Inr balance ramped successfully",
		"data":    gin.H{"balance": result.InrBalance},
	})
}

func GetInrBalanceHandler(c *gin.Context) {
	var req struct {
		UserId string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	result, err := worker(Input{
		Fnx: "get_inr_balance",
		UserId: req.UserId,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Inr balance fetched successfully",
		"data":    gin.H{"balance": result.InrBalance},
	})
}

func GetStockBalanceHandler(c *gin.Context) {
	var req struct {
		UserId string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	result, err := worker(Input{
		Fnx: "get_stock_balance",
		UserId: req.UserId,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock balance fetched successfully",
		"data":    gin.H{"balance": result.StockBalance},
	})
}

func GetMeHandler(c *gin.Context) {
	var req struct {
		UserId string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	result, err := worker(Input{
		Fnx: "get_me",
		UserId: req.UserId,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": err.Error()})
		return
	}
	if result.Err != "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error", "error": result.Err})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock balance fetched successfully",
		"data":    gin.H{"stockBalance": result.StockBalance, "inrBalance": result.InrBalance},
	})
	
}

func CancelBuyOrderHandler(c *gin.Context) {

}

func CancelSellOrderHandler(c *gin.Context) {

}