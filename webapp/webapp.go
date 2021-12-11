package webapp

import (
	"DefiAggregator/db"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartServer(database *db.Database, notifierClient *db.Notifier_new) {
	//func StartServer(database *db.Database, notifierClient *notifier.Notifier) {

	fmt.Print("Starting server: ")
	fmt.Print(time.Now())

	r := gin.Default()
	r.Use(cors.Default())

	// POST - risk setting
	// POST - own pf
	// GET  - optimised pf
	// GET  - raw pf - raw_portfolio
	// GET  - data table

	// Add own starting portfolio record
	r.POST("/raw_portfolio_input", func(c *gin.Context) {
		var json db.RawPortfolioRecord //		var json db.Record
		if err := c.BindJSON(&json); err == nil {
			database.AddOwnStartingPortfolioRecord(json)
			c.JSON(http.StatusCreated, json)
			//notifierClient.Notify()
			notifierClient.Notify_raw_and_optimised_pf()
		} else {
			c.JSON(http.StatusBadRequest, gin.H{})
		}
	})

	// Raw portfolio
	r.GET("/raw_portfolio", func(c *gin.Context) {
		raw_portfolio_data := database.GetRawPortfolio()
		c.JSON(http.StatusOK, gin.H{
			"raw_portfolio": raw_portfolio_data,
		})
	})

	// NEW
	r.GET("/optimised_portfolio", func(c *gin.Context) {
		optimised_portfolio_data := database.GetOptimisedPortfolio()
		c.JSON(http.StatusOK, gin.H{
			"optimised_portfolio": optimised_portfolio_data,
		})
	})

	// Post data from slider into db
	r.POST("/risk_setting", func(c *gin.Context) {
		var json db.RiskWrapper
		if err := c.BindJSON(&json); err == nil {
			// fmt.Println("ADDING RISK RECORD FROM BUTTON!!")
			database.SetRiskLevel(json) // json
			c.JSON(http.StatusCreated, json)
			//notifierClient.Notify()
			notifierClient.Notify_raw_and_optimised_pf()
			// fmt.Println(database.Risksetting)
		} else {
			fmt.Println("ERROR IN PARSING JSON RISK SETTING!!")
			c.JSON(http.StatusBadRequest, gin.H{})
		}
	})

	// Summary of returns from pools
	r.GET("/ranked_pools_table", func(c *gin.Context) {
		ranked_pools_table := database.GetRankedPoolsTable()
		c.JSON(http.StatusOK, gin.H{
			"ranked_pools_table": ranked_pools_table,
		})
	})

	go doEvery(10*time.Second, database.UpdateData, notifierClient)

	r.Run()
}

func doEvery(d time.Duration, f func(nc *db.Notifier_new), nc *db.Notifier_new) {
	for _ = range time.Tick(d) {
		f(nc)
	}
}
