package db

import (
	"fmt"
	"strconv"
	"github.com/machinebox/graphql"
)


type RawPortfolioRecord struct {
	Token  string  `json:"token"`
	Amount float64 `json:"amount"`
}

// Constructor
func NewRawPortfolioRecord(token string, amount float64) RawPortfolioRecord { // , pool_sz float64
	return RawPortfolioRecord{token, amount} // , pool_sz
}

// Output of RawPortfolioRecord
type OptimisedPortfolioRecord struct {
	Pool                	string  `json:"pool"`
	Token0       			string `json:"token0"`
	Token1          		string `json:"token1"`
	
	Amount_token0         	float64 `json:"amount_token0"`
	Amount_token1         	float64 `json:"amount_token1"`

	Total_Value_usd       	float64 `json:"total_value_usd"`
	Share_pf_tot_usd_val 	float64 `json:"percentageofportfolio"`
	ROI_raw_est           	float64 `json:"roi_estimate"`
	Volatility           	float64 `json:"volatility"`
	Risksetting           	float64 `json:"risk_setting"`
}

// Output of RawPortfolioRecord
/*
type OptimisedPortfolioRecord struct {
	TokenOrPair           string  `json:"tokenorpair"`
	Pool                  string  `json:"pool"`
	Total_Value_USD       float64 `json:"total_value_usd"`
	Amount_token0         float64 `json:"amount_token0"`
	PercentageOfPortfolio float64 `json:"percentageofportfolio"`
	ROI_raw_est           float64 `json:"roi_estimate"`
	Risksetting           float64 `json:"risk_setting"`
}
*/

type HistoricalCurrencyData struct {
	Date   []int64   `default0:"10099999999999" json:"date"` // `default0:"Mon Jan 2 15:04:05 MST 2006" json:"date"`
	Price  []float64 `default0:"420.69" json:"price"`
	Ticker string    `default0:"ETH" json:"ticker"`
}


func NewHistoricalCurrencyData() HistoricalCurrencyData {
	historicaldata := HistoricalCurrencyData{}
	historicaldata.Date = append(historicaldata.Date, 0.0)
	historicaldata.Price = append(historicaldata.Price, float64(0.0))
	historicaldata.Ticker = "UNDEFINED"
	return historicaldata
}

func NewHistPxDataFromRaw_aboveT(token string, rawhistoricaldata []UniswapDaily, time_min int64) HistoricalCurrencyData {
	var historicaldata HistoricalCurrencyData

	for i := 0; i < len(rawhistoricaldata); i++ {
		if int64(rawhistoricaldata[i].Date) > time_min {
			historicaldata.Date = append(historicaldata.Date, int64(rawhistoricaldata[i].Date))
			price, _ := strconv.ParseFloat(rawhistoricaldata[i].PriceUSD, 64)
			historicaldata.Price = append(historicaldata.Price, price)	
		}
	}

	historicaldata.Ticker = token
	return historicaldata
}

func NewHistoricalCurrencyDataFromRaw(token string, rawhistoricaldata []UniswapDaily) HistoricalCurrencyData {
	var historicaldata HistoricalCurrencyData

	for i := 0; i < len(rawhistoricaldata); i++ {
		historicaldata.Date = append(historicaldata.Date, int64(rawhistoricaldata[i].Date))
		price, _ := strconv.ParseFloat(rawhistoricaldata[i].PriceUSD, 64)
		historicaldata.Price = append(historicaldata.Price, price)
	}

	historicaldata.Ticker = token
	return historicaldata
}


// Current
type PoolTokenPairReturns struct {
	Pair            string  `default0:"ETH/DAI" json:"backend_pair"`
	PoolSize        float64 `default0:"0.0" json:"backend_poolsize"`
	PoolVolume      float64 `default0:"0.0" json:"backend_volume"`
	PoolRatio0		float64 `default0:"0.0" json:"pool_ratio_0"` 
	PoolRatio1		float64 `default0:"0.0" json:"pool_ratio_1"`
	Yield           float64 `default0:"0.0" json:"backend_yield"`
	Pool            string  `default0:"Uniswap" json:"pool_source"`
	Volatility      float64 `default0:"0.0%" json:"volatility"`
	ROI_raw_est     float64 `default0:"0.0%" json:"ROIestimate"`
	ROI_vol_adj_est float64 `default0:"0.0%" json:"ROIvoladjest"`
	ROI_hist        float64 `default0:"0.0%" json:"ROIhist"`
	Last_updated	int64	
}

type RiskWrapper struct {
	Risksettinginput float64 `json:"risk_setting"`
}

type Database struct {
	// Data structure for Optimisation
	ownstartingportfolio []RawPortfolioRecord       // for portfolio optimisation table
	optimisedportfolio   []OptimisedPortfolioRecord // for storing output of ownstartingportfolio
	Risksetting          float64

	// Data structure for Ranking
	PoolTokenPairReturns      []PoolTokenPairReturns      // LATEST currency pair info
	historicalcurrencydata []HistoricalCurrencyData // historical time series

	uniswapreqdata 		UniswapInputStruct
	update_frequency 	int64
}

func (db *Database) Getnumberofhistrecords() int {
	return len(db.PoolTokenPairReturns)
}


func New() Database {
	ownstartingportfolio := make([]RawPortfolioRecord, 0)
	optimisedportfolio := make([]OptimisedPortfolioRecord, 0)
	Risksetting := 0.00

	PoolTokenPairReturns := make([]PoolTokenPairReturns, 0)
	historicalcurrencydata := make([]HistoricalCurrencyData, 0)

	// define uniswap queries here
	// 1 - create clients
	clientUniswap := graphql.NewClient("https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2")
	//	clientCompound := graphql.NewClient("https://api.thegraph.com/subgraphs/name/graphprotocol/compound-v2")

	// 2 - declare queries
	reqUniswapHist := graphql.NewRequest(`
				query ($tokenid:String!){
						tokenDayDatas(first: 30 orderBy: date, orderDirection: desc,
						 where: {
						   token:$tokenid
						 }
						) {
						   date
						   priceUSD
						   token{
							   id
							   symbol
						   }
						}
				  }
			`)

	reqUniswapIDFromTokenTicker := graphql.NewRequest(`
						query ($ticker:String!){
							tokens(where:{symbol:$ticker})
							{
								id
								symbol
							}
						}
			`)

	// 3 - set query headers
	reqUniswapIDFromTokenTicker.Header.Set("Cache-Control", "no-cache")
	reqUniswapHist.Header.Set("Cache-Control", "no-cache")

	// 4 - run data queries on each pool
	U := UniswapInputStruct{clientUniswap, reqUniswapIDFromTokenTicker, reqUniswapHist}

	// 60 is the default update frequency - now 600
	return Database{ownstartingportfolio, optimisedportfolio, float64(Risksetting), PoolTokenPairReturns, historicalcurrencydata,U,600}
}

// Add OWN PORTFOLIO data
func (database *Database) AddOwnStartingPortfolioRecord(r RawPortfolioRecord) {
	already_exists := false
	// check if that token already exists
	for i := 0; i < len(database.ownstartingportfolio);i++ {
		if database.ownstartingportfolio[i].Token == r.Token {
			already_exists = true
			// Record already exists - just add up - do not append
			database.ownstartingportfolio[i].Amount += r.Amount
		}
	}

	if !already_exists && r.Amount != 0.0 && len(r.Token) > 1  {
		database.ownstartingportfolio = append(database.ownstartingportfolio, r)
	}

	//fmt.Print("499 - Added new record - now re-optimising PF..")
	database.OptimisePortfolio()
}

func (database *Database) SetRiskLevel(risk RiskWrapper) {
	fmt.Print("------------SETTING RISK TO: ")
	fmt.Println(risk.Risksettinginput)
	
	if database.Risksetting != risk.Risksettinginput {
		database.Risksetting = risk.Risksettinginput

		database.OptimisePortfolio()
	}

}

// Retrieve pool list
func (database *Database) GetRankedPoolsTable() []PoolTokenPairReturns {
	return database.PoolTokenPairReturns
}

// Retrieve optimised pf
func (database *Database) GetOptimisedPortfolio() []OptimisedPortfolioRecord {
	return database.optimisedportfolio
}

// Retrieve raw pf
func (database *Database) GetRawPortfolio() []RawPortfolioRecord {
	return database.ownstartingportfolio
}

