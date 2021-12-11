package db

import (
	"fmt"
//	"time"
	"sort"
	"github.com/machinebox/graphql"
)

type UniswapInputStruct struct {
	clientUniswap               *graphql.Client
	reqUniswapIDFromTokenTicker *graphql.Request
	reqUniswapHist              *graphql.Request
}

func (database *Database) UpdateData(nc *Notifier_new) {
	/*
	t := BoDi64(time.Now().UTC().Unix()) // - 24 * 60 * 60 *2
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(1, "USD", "WETH", t)) 
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(1, "USD", "DOGE", t))
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(1, "WETH", "USD", t)) 
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(1, "DOGE", "USD", t))
	fmt.Printf("\n------------------------------------------\n")
	fmt.Printf("\n------------------------------------------\n")

	fmt.Println(convert_amt_from_to(100, "WBTC", "USD", t)) 
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(100, "USD", "USDC", t)) 
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(100, "USD", "WBTC", t)) 
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(100, "AGQ", "USD", t)) 
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(100, "WETH", "WBTC", t)) 
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(100, "WBTC", "WETH", t))
	fmt.Printf("\n------------------------------------------\n")
	fmt.Println(convert_amt_from_to(0, "WBTC", "WETH", t))
	*/
//	t2 := BoDi64(time.Now().UTC().Unix()) - 24 * 60 * 60 * 200
//	fmt.Printf("\n------------------------------------------\n")

/*
	var t0 []string
	var t1 []string
	var w []float64
	var ratios []float64
	var conv []float64
	var r_arr 	[]RawPortfolioRecord

	t0 = append(t0,"USDT")
	t0 = append(t0,"DAI")
	t0 = append(t0,"WETH")

	t1 = append(t1,"DOGE")
	t1 = append(t1,"")
	t1 = append(t1,"")

	w = append(w,0.25)
	w = append(w,0.25)
	w = append(w,0.25)

	ratios = append(ratios,0.73)
	ratios = append(ratios,1.00)
	ratios = append(ratios,1.00)

	conv = append(conv,0.54)
	conv = append(conv,2500)
	conv = append(conv,1.00)

	r0 := RawPortfolioRecord{"DOGE",2000}
	r1 := RawPortfolioRecord{"WETH",1}
	r2 := RawPortfolioRecord{"DAI",500}

	r_arr = append(r_arr,r0)
	r_arr = append(r_arr,r1)
	r_arr = append(r_arr,r2)

	res, l_t, l_amt := nrm_pool_wgts(w, t0, t1, ratios, r_arr, conv)
	fmt.Print(res)
	fmt.Print(l_t)
	fmt.Print(l_amt)
	fmt.Print("---ZZZ-----")
*/	
	//fmt.Print("---RUNNING DATA RESCAN CYCLE-----")
	//database.getUniswapData()
	//database.getBalancerData(nc)
 if len(database.ownstartingportfolio) == 0 {
	database.ownstartingportfolio = append(database.ownstartingportfolio,RawPortfolioRecord{"DOGE",2000})
	database.ownstartingportfolio = append(database.ownstartingportfolio,RawPortfolioRecord{"WETH",0.25})
	database.ownstartingportfolio = append(database.ownstartingportfolio,RawPortfolioRecord{"DAI",1000})
 }

	fmt.Print("--------APPENDED RAW STARTING PF---------")
//	database.ownstartingportfolio = append(database.ownstartingportfolio,RawPortfolioRecord{"BTC",0.01})
//	database.ownstartingportfolio = append(database.ownstartingportfolio,RawPortfolioRecord{"USDT",750})

	database.GetTestData(nc)

	database.OptimisePortfolio()
	//database.getAave2Data()
	//database.getCurveData()
	
	sort.Slice(database.PoolTokenPairReturns, func(i, j int) bool {
		return database.PoolTokenPairReturns[i].ROI_raw_est > database.PoolTokenPairReturns[j].ROI_raw_est
	})

//	fmt.Println("Ran all download functions and appended data")
}
