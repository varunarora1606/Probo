package order

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/varunarora1606/Probo/internal/memory"
)

type HandlerReq struct {
	UserId    string           `json:"userId" binding:"required"`
	Symbol    string           `json:"stockSymbol" binding:"required"` //symbol
	Quantity  int              `json:"quantity" binding:"required"`
	Price     int              `json:"price"`
	StockSide memory.Side      `json:"stockSide" binding:"required,oneof=yes no"`       //side
	StockType memory.OrderType `json:"stockType" binding:"required,oneof=market limit"` //ordertype
}

type CreateMarketHandlerReq struct {
	UserId       string `json:"userId" binding:"required"`
	Question     string `json:"question" binding:"required"`
	EndTimeMilli int64  `json:"endTime" binding:"required"`
	Symbol 		 string `json:"symbol" binding:"required"`
}

func BuyHandler(c *gin.Context) {
	var req HandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	if _, exist := memory.OrderBook[req.Symbol]; !exist {
		c.JSON(http.StatusNotFound, gin.H{"message": "This market does not exist"})
		return
	}

	buyStock(req.Symbol, memory.Side(req.StockSide), req.Price, req.UserId, req.Quantity, memory.OrderType(req.StockType))

	c.JSON(http.StatusCreated, gin.H{"message": "Stock added successfully", "orderBook": memory.OrderBook, "inrBalance": memory.InrBalance, "stockBalance": memory.StockBalance, "betBook": memory.BetBook})
}

func SellHandler(c *gin.Context) {
	var req HandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	if _, exist := memory.OrderBook[req.Symbol]; !exist {
		c.JSON(http.StatusNotFound, gin.H{"message": "This market does not exist"})
		return
	}

	sellStock(req.Symbol, memory.Side(req.StockSide), req.Price, req.UserId, req.Quantity, memory.OrderType(req.StockType))

	c.JSON(http.StatusCreated, gin.H{"message": "Stock sold successfully", "data": memory.OrderBook, "inrBalance": memory.InrBalance, "stockBalance": memory.StockBalance, "betBook": memory.BetBook})
}

func GetMarketHandler(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	if marketBook, exist := memory.OrderBook[req.Symbol]; !exist {
		c.JSON(http.StatusNotFound, gin.H{"message": "This market does not exist"})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Market fetched successfully",
			"data":    marketBook,
		})
	}
}

func GetMarketsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Markets fetched successfully",
		"data":    memory.OrderBook,
	})
}

func CreateMarketHandler(c *gin.Context) {
	var req CreateMarketHandlerReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	// TODO: Check symbol in DB

	endTime := time.Unix(0, req.EndTimeMilli*int64(time.Millisecond))
	serverTime := time.Now()

	if endTime.Before(serverTime.Add(5 * time.Second)) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "EndTime should be in far future"})
		return
	}
	// TODO: Add to DB

	// TODO: Add mutex
	memory.OrderBook[req.Symbol] = memory.StockBook{
		Yes: make(map[int]memory.OrderDetails),
		No:  make(map[int]memory.OrderDetails),
	}

	// TODO: Return Db result.

	c.JSON(http.StatusOK, gin.H{
		"message": "Inr balance fetched successfully",
		"data":    gin.H{"symbol": req.Symbol, "endTime": endTime},
	})
}

func OnRampInrHandler(c *gin.Context) {
	var req struct {
		UserId string `json:"userId" binding:"required"`
		Quantity int `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	// TODO: Add to DB
	// TODO: Add mutex
	userBalance := memory.InrBalance[req.UserId]
	userBalance.Quantity += req.Quantity;
	memory.InrBalance[req.UserId] = userBalance

	c.JSON(http.StatusOK, gin.H{
		"message": "Inr balance ramped successfully",
		"data":    gin.H{"balance": userBalance},
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

	balance, exist := memory.InrBalance[req.UserId]
	if !exist {
		balance = memory.Balance{
			Quantity: 0,
			Locked:   0,
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Inr balance fetched successfully",
		"data":    balance,
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

	if balance, exist := memory.StockBalance[req.UserId]; !exist {
		c.JSON(http.StatusOK, gin.H{"message": "User does not have any stock"})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Stock balance fetched successfully",
			"data":    balance,
		})
	}
}

func GetMeHandler(c *gin.Context) {
	var req struct {
		UserId string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Validation error", "error": err.Error()})
		return
	}

	inrBalance, exist := memory.InrBalance[req.UserId]
	if !exist {
		inrBalance = memory.Balance{
			Quantity: 0,
			Locked:   0,
		}
	}
	if stockBalance, exist := memory.StockBalance[req.UserId]; !exist {
		c.JSON(http.StatusOK, gin.H{
			"message": "User does not have any stock",
			"data":    gin.H{"stockBalance": "", "inrBalance": inrBalance},
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Stock balance fetched successfully",
			"data":    gin.H{"stockBalance": stockBalance, "inrBalance": inrBalance},
		})
	}
}

func CancelBuyOrderHandler(c *gin.Context) {

}

func CancelSellOrderHandler(c *gin.Context) {

}