package db

import (
	"fmt"
	//	"log"
	"math"
	"strings"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/optimize"
	"gonum.org/v1/gonum/stat"
)

// Main optimiser function
func (database *Database) OptimisePortfolio() {
	fmt.Printf("\n\n\n----ENTERING PORTFOLIO OPTIMISATION----------\n")

	if len(database.ownstartingportfolio) == 0 {
		return
	}

	var raw_pf_tokens_unfiltered []string    // Unique own portfolio token list
	var raw_pf_records_with_pools_to_deploy_into []RawPortfolioRecord

	var pool_ratios []float64

	var h_array []HistoricalCurrencyData // For storing historical px data of pools to deploy into

	var available_pool_names []string
	var available_pool_tkn0s []string   // Pool list token0
	var available_pool_tkn1s []string   // Pool list token1

	var available_pool_returns_est []float64
	var available_pool_returns_hist []float64
	
	var conversion_to_usd_px_arr []float64
	var avg_returns []float64

	var Available bool

	var optimised_pf []OptimisedPortfolioRecord // return value
	// gas price

	// 0 - Create copy of OWN PF ticker list
	for i := 0; i < len(database.ownstartingportfolio); i++ {
			raw_pf_tokens_unfiltered = append(raw_pf_tokens_unfiltered, database.ownstartingportfolio[i].Token)
	}
	/*
	fmt.Print("Unfiltered own raw pf token list: ")
	for x, y := range raw_pf_tokens_unfiltered {
		fmt.Print(x)
		fmt.Print(": ")
		fmt.Println(y)
	}
	*/

	fmt.Printf("\n-----------991-------------------\n")
	// 1 - Calculate available pools for deployment + PULL THEIR DATA
	// 3 & 4 - PREPARE VARIABLES - TO FEED INTO FUNC TO NORMALIZE WEIGHTS
	for i := 0; i < len(database.PoolTokenPairReturns); i++ {
		//fmt.Printf("Looping through POOL LIST i: %v\n", i)
		ss := strings.Split(database.PoolTokenPairReturns[i].Pair, "/")
		//fmt.Printf("len of s: %v", len(ss))
		// get rid of all the zero strings
		var s []string
		for jx := 0; jx < len(ss); jx++{
			if len(ss[jx]) > 0 {
				s = append(s, ss[jx])
			}
		}
		// fmt.Printf("len s : %v\n\n", len(s))
		// Loop through pools - if pairs in pool made up of portfolio tokens - add to AVAILABLE pools
		if len(s) == 2 {
				//fmt.Print(s[0])
				//fmt.Print(" | ")
				//fmt.Println(s[1])
				Available = stringInSlice(s[0], raw_pf_tokens_unfiltered) && stringInSlice(s[1], raw_pf_tokens_unfiltered)
				//fmt.Print(Available)
				//fmt.Printf("avail0: %v\n", Available)
			} else if len(s) == 1 {
				//fmt.Print("len 1: ")
				//fmt.Print(s[0])
				//fmt.Print(" ")
				Available = stringInSlice(s[0], raw_pf_tokens_unfiltered)
				//fmt.Print(Available)
				//fmt.Printf("avail1: %v\n", Available)
			} else {
				Available = false
				//fmt.Printf("avail2: %v\n", Available)
			}

		if Available {
			available_pool_names = append(available_pool_names, database.PoolTokenPairReturns[i].Pool)

			available_pool_returns_est = append(available_pool_returns_est, database.PoolTokenPairReturns[i].ROI_raw_est)
			available_pool_returns_hist = append(available_pool_returns_hist, database.PoolTokenPairReturns[i].ROI_hist)

			pool_ratios = append(pool_ratios,database.PoolTokenPairReturns[i].PoolRatio0)
			available_pool_tkn0s = append(available_pool_tkn0s, s[0])	

			if len(s) == 2 {
				if s[0] != s[1] {
					h_array = append(h_array,getHistPriceDataForTokenPairFromDB(s[0], s[1]))
					available_pool_tkn1s = append(available_pool_tkn1s, s[1])
				}
			}
			if len(s) == 2 {
				if s[0] == s[1] {
					h_array = append(h_array,getHistPriceDataForTokenPairFromDB(s[0], "USD"))
					available_pool_tkn1s = append(available_pool_tkn1s, "")
					// available_pool_tkn1s = append(available_pool_tkn1s, "USD")
				}
			}
			if len(s) != 2 {
				h_array = append(h_array, getHistPriceDataForTokenPairFromDB(s[0], "USD"))
				available_pool_tkn1s = append(available_pool_tkn1s, "")
				//available_pool_tkn1s = append(available_pool_tkn1s, "USD")
			}
		} // if Available
	} // Calculate available pools for deployment + PULL THEIR HIST PX DATA - end

	fmt.Printf("\n-----------992-------------------\n")

	// print everything computed to far

	fmt.Printf("---------------COMPUTED SO FAR - AVAILABLE DESTINATIONS: --------------------\n")
	fmt.Println(len(available_pool_tkn0s))
	fmt.Println(len(available_pool_tkn1s))
	fmt.Println(len(h_array))
	fmt.Println(len(pool_ratios))
	fmt.Println(len(available_pool_names))

	fmt.Println("Available destination pools: ")
	for i := 0; i < len(available_pool_tkn0s);i++{
		fmt.Print(available_pool_tkn0s[i])
		fmt.Print(" | ")
		fmt.Print(available_pool_tkn1s[i])
		fmt.Print(" | ")
		fmt.Print(h_array[i].Ticker)
		fmt.Print(" | ")
		fmt.Print(pool_ratios[i])
		fmt.Print(" | ")
		fmt.Print(available_pool_names[i])
		fmt.Print(" | ")
		fmt.Print(available_pool_returns_est[i])			
		fmt.Print(" | ")
		fmt.Println(available_pool_returns_hist[i])			
	}

	// calculate total pf val in usd px as well
	total_pf_val_usd := 0.0
	for i:=0; i < len(database.ownstartingportfolio);i++ {
		usd_amt := convert_amt_from_to_using_latest_exch_rate(database.ownstartingportfolio[i].Amount, database.ownstartingportfolio[i].Token,"USD")
		fmt.Printf(".adding to TOTAL: %v\n", usd_amt)
		total_pf_val_usd += usd_amt
	}

	fmt.Printf("\nCALCULATED TOTAL PF VALUE AS: %v\n", total_pf_val_usd)

 	if len(available_pool_tkn0s) == 0 { // if nothing available - all own pf items go to leftovers - return
		fmt.Print("..401 - Nothing available!!..everything to be kept at hand..")
		for i:=0; i < len(database.ownstartingportfolio);i++ {
			yield := float64(0.0)
			volatility := float64(0.0)
			pool := "Own Hardware Wallet"
			amt_usd := convert_amt_from_to_using_latest_exch_rate(database.ownstartingportfolio[i].Amount, database.ownstartingportfolio[i].Token,"USD")
			pct := 0.0
			if total_pf_val_usd > 0 {
				pct = amt_usd / total_pf_val_usd
			}
			fmt.Printf("Appending optimised record: %v\n", database.ownstartingportfolio[i].Amount)
			optimised_pf = append(optimised_pf, OptimisedPortfolioRecord{pool, database.ownstartingportfolio[i].Token, "", database.ownstartingportfolio[i].Amount, 0.0, amt_usd, pct, yield, volatility, database.Risksetting})
		}
		database.optimisedportfolio = optimised_pf
		return
	} // if nothing available - everything goes to leftover 

	// 4 - Filter out OWN portfolio tokens - throw out these which have ZERO pools containing them
	for i := 0; i < len(database.ownstartingportfolio); i++ {
		if stringInSlice(database.ownstartingportfolio[i].Token, available_pool_tkn0s) || stringInSlice(database.ownstartingportfolio[i].Token, available_pool_tkn1s) {
			raw_pf_records_with_pools_to_deploy_into = append(raw_pf_records_with_pools_to_deploy_into, database.ownstartingportfolio[i])
			conversion_to_usd_px_arr = append(conversion_to_usd_px_arr, get_latest_token_price(database.ownstartingportfolio[i].Token))
		}
	}

	fmt.Printf("\n-----------992-------------------\n")
	fmt.Printf("NUMBER OF RECORDS IN H ARRAY 0 PRICE: %v\n",len(h_array[0].Price))

	// 2 -  pulled pool price data to return matrix (prices first)
	ret_mat_xxx := mat.NewDense(1, 1, nil)
	max_len := 2

	if len(h_array) > 0 {
		max_len = int( math.Min(30,float64(len(h_array[0].Price))) )
		ret_mat_xxx = mat.NewDense(max_len, len(h_array), nil)
		// h_array[jj]
			for jj := 0; jj < len(h_array); jj++ {
				for ii := 0; ii < max_len; ii++ { // row?
					if len(h_array[jj].Price) - 1 >= ii {
						ret_mat_xxx.Set(ii, jj, float64(h_array[jj].Price[ii]))
					} else {
						fmt.Print("...WARNING: DATA MISSING..FILLING last known...")
						max_idx := len(h_array[jj].Price) - 1
						ret_mat_xxx.Set(ii, jj, float64(h_array[jj].Price[max_idx]))
					}
				} // ii  
			} // jj
	} else {
		ret_mat_xxx.Zero()
	} // if len(h_array) > 0

	fmt.Printf("\n-----------993-------------------\n")
	//fmt.Print(ret_mat_xxx)
	// 5 - Define dimensions
	number_of_pools := int(math.Max(float64(len(h_array)), 1)) // Number of pools to deploy into
	number_of_days := 2                                         // starting value to prevent errors in sizing
	if len(h_array) > 0 {
		number_of_days = max_len // int(math.Max(float64(len(h_array[0].Price)), 2))
	}

	// 6 - Declare matrix for returns data in %
	ret_mat_pct := mat.NewDense(number_of_days - 1, number_of_pools, nil)

	// 7 - Populate this matrix with returns in %
	for ii := 0; ii < number_of_days - 1; ii++ { // row
		for jj := 0; jj < number_of_pools; jj++ { // col
			if ret_mat_xxx.At(ii, jj) != 0.0 {
				ret_mat_pct.Set(ii, jj, ret_mat_xxx.At(ii+1, jj)/ret_mat_xxx.At(ii, jj)-1.0)
			} else {
				ret_mat_pct.Set(ii, jj, 0.0)
			}
		} // jj
	} // ii

	fmt.Printf("\n-----------994-------------------\n")
	// 8 - Calculate average returns by POOL to be deployed into
	for jj := 0; jj < number_of_pools; jj++ {	
		avg_returns = append(avg_returns, available_pool_returns_est[jj])
	}
	
	ret := mat.NewVecDense(number_of_pools, avg_returns) // vector of returns

	fmt.Printf("\n-----------999-------------------\n")

	// 9 - Calculate covariance matrix
	var cov *mat.SymDense = mat.NewSymDense(number_of_pools, nil)
	cov.Reset()
	stat.CovarianceMatrix(cov, ret_mat_pct, nil) // use ret mat pct for covariance
	var cov2 *mat.SymDense = mat.NewSymDense(number_of_pools, nil)

	for ii := 0; ii < number_of_pools; ii++ { // row
		for jj := 0; jj < number_of_pools; jj++ { // col
			cov2.SetSym(ii, jj, cov.At(ii, jj)*252) // annualise them
		} // ii
	} // jj
	fmt.Printf("\n-----------1000-------------------\n")

	// 10 - Define optimization function
	fcn := func(weights_in_opt_func []float64) float64 {	
		weights_in_opt_func, _, _ = nrm_pool_wgts(weights_in_opt_func, available_pool_tkn0s, available_pool_tkn1s, pool_ratios, raw_pf_records_with_pools_to_deploy_into, conversion_to_usd_px_arr)
		weights := mat.NewVecDense(number_of_pools, weights_in_opt_func)

		blended_return := mat.Dot(ret, weights) // XXX - add actual returns from pools

		risk_step0 := mat.NewVecDense(number_of_pools, nil)
		risk_step0.MulVec(cov2, weights)
		risk := math.Sqrt(mat.Dot(weights, risk_step0))
		sharpe := -blended_return / risk // Return sharpe ratio

		if math.IsNaN(sharpe) || math.IsInf(sharpe, 0) {
			return 0.0
		}

		return sharpe
	} // fcn definition complete

	fmt.Printf("\n-----------1001-------------------\n")
	// 11 - Call the optimizer
	var p0 []float64

	if len(h_array) > 0 { // sized same as number of tokens
		for i := 0; i < len(h_array); i++ {
			p0 = append(p0, 1/float64(len(h_array)))
		}
	} else {
		p0 = append(p0, 0.0)
	} // 1/number_of_tokens

	// 12 - Feed fcn into optimizer
	p := optimize.Problem{
		Func: fcn,
	}

	fmt.Printf("\n-----------1002-------------------\n")
	
	result, err := optimize.Minimize(p, p0, nil, nil)
	if err != nil {
		fmt.Print("405 - ERROR: optimisation FAILED..")
		//log.Fatal(err)
	}

	fmt.Printf("\n-----------1003-------------------\n")	
	fmt.Print("RAW WEIGHTS OPTIMIZED: ")
	fmt.Println(result)

	for i := 0; i < len(available_pool_tkn0s);i++{
		fmt.Print(available_pool_tkn0s[i])
		fmt.Print(" | ")
		fmt.Print(available_pool_tkn1s[i])
		fmt.Print(" | ")
		fmt.Print(h_array[i].Ticker)
		fmt.Print(" | ")
		fmt.Print(pool_ratios[i])
		fmt.Print(" | ")
		fmt.Print(available_pool_names[i])
		fmt.Print(" | ")
		fmt.Print(available_pool_returns_est[i])			
		fmt.Print(" | ")
		fmt.Println(available_pool_returns_hist[i])			
	}

	result_norm, leftovertokens, leftoveramounts := nrm_pool_wgts(result.X, available_pool_tkn0s, available_pool_tkn1s, pool_ratios, raw_pf_records_with_pools_to_deploy_into, conversion_to_usd_px_arr)

//	fmt.Printf("\n-----------1004-------------------\n")	
//	fmt.Print("..NORMALIZED WEIGHTS OPTIMIZED: ")
//	fmt.Println(result_norm)
	fmt.Printf("\n-----------1005 PACKING RESULTS INTO OPTIMISED PF ARRAY-------------------\n")	
	// pack results into optimised pf array
	for i := 0; i < len(result_norm); i++ {
		val_usd := total_pf_val_usd * result_norm[i]

		fmt.Printf("at i: %v", i)
		fmt.Printf(" val usd i: %v", val_usd)
		fmt.Printf(" | available_pool_tkn0s[i]: %v", available_pool_tkn0s[i])
		fmt.Printf(" | available_pool_tkn1s[i]: %v\n", available_pool_tkn1s[i])
		fmt.Printf(" | pool ratios[i]: %v\n", pool_ratios[i])

		//ratio := 0.0 
		roi_est := 0.0
		volatility := 0.0

		if val_usd > 0 {
		for ii := 0; ii < len(database.PoolTokenPairReturns); ii++ {
			if database.PoolTokenPairReturns[ii].Pair == (available_pool_tkn0s[i]) && database.PoolTokenPairReturns[ii].Pool == available_pool_names[i] {
				//fmt.Printf("\nEquating a LENDING POOL..t1 = %v\n",available_pool_tkn1s[i])
				roi_est = database.PoolTokenPairReturns[ii].ROI_raw_est
				volatility = database.PoolTokenPairReturns[ii].Volatility
				//ratio = database.PoolTokenPairReturns[ii].PoolRatio0
			}
			if database.PoolTokenPairReturns[ii].Pair == (available_pool_tkn0s[i] + "/" + available_pool_tkn1s[i]) && database.PoolTokenPairReturns[ii].Pool == available_pool_names[i] {
				roi_est = database.PoolTokenPairReturns[ii].ROI_raw_est
				volatility = database.PoolTokenPairReturns[ii].Volatility
				//ratio = database.PoolTokenPairReturns[ii].PoolRatio0
			}
			if available_pool_tkn1s[i] == "USD" && database.PoolTokenPairReturns[ii].Pair == (available_pool_tkn0s[i] + "/" + available_pool_tkn0s[i]) && database.PoolTokenPairReturns[ii].Pool == available_pool_names[i] {
				roi_est = database.PoolTokenPairReturns[ii].ROI_raw_est
				volatility = database.PoolTokenPairReturns[ii].Volatility
				// ratio = database.PoolTokenPairReturns[ii].PoolRatio0
			}
		}

//		fmt.Printf("\nRATIO0: %v\n",ratio)  ratio
		amount_token0 := convert_amt_from_to_using_latest_exch_rate(val_usd * pool_ratios[i],"USD",available_pool_tkn0s[i])
		amount_token1 := float64(0.0)

		if available_pool_tkn1s[i] == "" {
			amount_token1 = 0.0
		} else {
			//fmt.Print("XXXXXXXXXXXXXXXXXXX2222222")
			amount_token1 = convert_amt_from_to_using_latest_exch_rate(val_usd * (1-pool_ratios[i]),"USD",available_pool_tkn1s[i])
			//fmt.Print("YYYYYYYYYYYYYYYYYYY2222222")
		}

		fmt.Print("CHECK IF NOT BLANK available_pool_tkn1s[i]: %v", available_pool_tkn1s[i])

		fmt.Printf("..APPENDING NEW PF record: %v", amount_token0)
		optimised_pf = append(optimised_pf, OptimisedPortfolioRecord{available_pool_names[i], available_pool_tkn0s[i], available_pool_tkn1s[i], amount_token0, amount_token1, val_usd, result_norm[i], roi_est, volatility, database.Risksetting})
		}
	}

	fmt.Printf("\n-----------1006-------------------\n")	

	for i:= 0; i < len(leftovertokens); i++ { 	// pack leftover tokens and amounts
		roi_est := float64(0.0)
		pool := "Keep at Hand"
		pct := float64(0.0)
		volatility := float64(0.0)

		for j := 0; j < len(database.PoolTokenPairReturns); j++{
			// search if anything available for that token
			ss := strings.Split(database.PoolTokenPairReturns[j].Pair, "/")	
			var s []string
			for jx := 0; jx < len(ss); jx++{
				if len(ss[jx]) > 0 {
					s = append(s, ss[jx])
				}
			}
			
			if len(s) == 1 {
				if s[0] == leftovertokens[i] {
					roi_est = database.PoolTokenPairReturns[j].Yield
					volatility = database.PoolTokenPairReturns[j].Volatility
				}
			}
			if len(s) == 2 {
				if s[0] == leftovertokens[i] && s[1] == leftovertokens[i] {
					roi_est = database.PoolTokenPairReturns[j].ROI_raw_est
					volatility = database.PoolTokenPairReturns[j].Volatility
				}
			}
		} // j 

		if total_pf_val_usd > 0 {
			pct = leftoveramounts[i] / total_pf_val_usd		
		}
		
		fmt.Printf("leftoveramounts[i]: %v ", leftoveramounts[i])
		fmt.Printf("leftovertokens[i]: %v \n", leftovertokens[i])
		amount_token0 := convert_amt_from_to_using_latest_exch_rate(leftoveramounts[i], "USD", leftovertokens[i])

		if leftoveramounts[i] > 1 { // Discard small amounts
			fmt.Printf("..APPENDING leftover AMT to OPTIMISED PF t0: %v \n",amount_token0)			
			optimised_pf = append(optimised_pf, OptimisedPortfolioRecord{pool, leftovertokens[i], "", amount_token0, 0.0, leftoveramounts[i], pct, roi_est, volatility, database.Risksetting})
		}
	}

	fmt.Printf("\n-----------1007-------------------\n")
	fmt.Println("..OPTIMIZATION COMPLETE..")

	database.optimisedportfolio = optimised_pf
	return
}






	/*
	// Pack results into output struct array
	for i := 0; i < len(result_norm); i++ {
		amount_usd := total_pf_val_usd * result_norm[i]
		amount_token0 := amount_usd / conversion_to_usd_px_arr[i]
		roi_est := 0.0

		if amount_usd > 0 {
		for ii := 0; ii < len(database.PoolTokenPairReturns); ii++ {
			if database.PoolTokenPairReturns[ii].Pair == (available_pool_tkn0s[i] + "/" + available_pool_tkn1s[i]) && database.PoolTokenPairReturns[ii].Pool == available_pool_names[i] {
				roi_est = database.PoolTokenPairReturns[ii].ROI_raw_est
			}
			if available_pool_tkn1s[i] == "USD" && database.PoolTokenPairReturns[ii].Pair == (available_pool_tkn0s[i] + "/" + available_pool_tkn0s[i]) && database.PoolTokenPairReturns[ii].Pool == available_pool_names[i] {
				roi_est = database.PoolTokenPairReturns[ii].ROI_raw_est
			}
		}

		optimised_pf = append(optimised_pf, OptimisedPortfolioRecord{available_pool_tkn0s[i] + "/" + available_pool_tkn1s[i], available_pool_names[i], amount_usd,amount_token0, result_norm[i], roi_est, database.Risksetting})
		}
	}
	*/



/*
	for i := 0; i < len(optimised_pf); i++ {
		fmt.Print(optimised_pf[i].TokenOrPair)
		fmt.Print(" | ")
		fmt.Print(optimised_pf[i].PercentageOfPortfolio)
		fmt.Print(" | ")
		fmt.Print(optimised_pf[i].Pool)
		fmt.Print(" | ")
		fmt.Println(optimised_pf[i].Amount_token0)
		fmt.Print(" | ")
		fmt.Println(optimised_pf[i].Total_Value_USD)
	}
*/





/*
	if len(leftovertokens) == 0 && len(result_norm) == 0 && len(database.ownstartingportfolio) > 0 {
		// populate leftovers here
		for i:= 0; i < len(database.ownstartingportfolio); i++ {
			leftovertokens = append(leftovertokens,database.ownstartingportfolio[i].Token)
			leftoveramounts = append(leftoveramounts,database.ownstartingportfolio[i].Amount)
		}
	}
*/
//	total_pf_val := 0.0
//	for i:=0; i < len(raw_pf_records_with_pools_to_deploy_into); i++ {
//		total_pf_val += raw_pf_records_with_pools_to_deploy_into[i].Amount*conversion_to_usd_px_arr[i]
//	}


			/*
			fmt.Print("SEARCHING FOR YIELD IN DB!!!: ")
			fmt.Print(ii)
			fmt.Print(" pair: ")
			fmt.Print(database.PoolTokenPairReturns[ii].Pair)
			fmt.Print(" | ")
			fmt.Print(database.PoolTokenPairReturns[ii].ROI_raw_est)
			fmt.Print(" | name: ")
			fmt.Print(database.PoolTokenPairReturns[ii].Pool)
			fmt.Print(" | tryna match: ")
			fmt.Print(available_pool_tkn0s[i] + "/" + available_pool_tkn1s[i])
			fmt.Print(" | our pool name: ")
			fmt.Print(available_pool_names[i])
			fmt.Println(" | ")
			*/

/*
	fmt.Print("len len(h_array): ")
	fmt.Println(len(h_array))
	fmt.Println(number_of_days)
	fmt.Println(number_of_tokens)
*/

/*
	for i := 0; i < len(raw_pf_tokens_unfiltered); i++ {
		fmt.Println(raw_pf_tokens_unfiltered[i])
	}
*/
		//	database.ownstartingportfolio = append(database.ownstartingportfolio, RawPortfolioRecord{"WETH", 0.01})
		//	database.ownstartingportfolio = append(database.ownstartingportfolio, RawPortfolioRecord{"DAI", 200})
		//	database.ownstartingportfolio = append(database.ownstartingportfolio, RawPortfolioRecord{"USDC", 200})
		//fmt.Print("Empty raw portfolio - returning blank")

				//if !stringInSlice(database.ownstartingportfolio[i].Token, raw_pf_tokens_unfiltered) {

/*
	for j := 0; j < len(h_array); j++ {
		s := strings.Split(h_array[j].Ticker, "/")
		available_pool_tkn0s = append(available_pool_tkn1s, s[0])
		
		if len(s) == 1  {
			available_pool_tkn1s = append(available_pool_tkn1s, "USD")
			filtered_ratios_array = append(filtered_ratios_array,pool_ratios[j])
		} 
		if len(s) == 2 {
			if s[0] == s[1] {
				available_pool_tkn1s = append(available_pool_tkn1s, "USD")
				filtered_ratios_array = append(filtered_ratios_array,pool_ratios[j])
			}
		} 
		if len(s) == 2 {
			if s[0] == s[1] {
				available_pool_tkn1s = append(available_pool_tkn1s, s[1])
				filtered_ratios_array = append(filtered_ratios_array,pool_ratios[j])
			}
		}
	}
*/

/*
fmt.Printf("\nPOOLS available to DEPLOY INTO: ")
for i:=0; i < len(available_pool_tkn0s); i++ {
	fmt.Printf("i: %v", i)
	fmt.Printf(" | t0: %v ", available_pool_tkn0s[i])
	fmt.Printf(" | t1: %v ", available_pool_tkn1s[i])
	fmt.Printf(" | ratios: %v ", filtered_ratios_array[i])
}
*/

/*
	fmt.Printf(" | deployable portfolio i : %v", i)
	fmt.Print(" | ")
	fmt.Print(raw_pf_records_with_pools_to_deploy_into[int(len(raw_pf_records_with_pools_to_deploy_into)-1)])
	fmt.Print(" | ")
	fmt.Print(" | px: ")
	fmt.Print(conversion_to_usd_px_arr[int(len(raw_pf_records_with_pools_to_deploy_into)-1)])
*/


		/*
			total := 0.0
			for ii := 0; ii < number_of_days-1; ii++ {
				total += ret_mat_pct.At(ii, jj)
			}
			
			px_ret := 252*total/float64((number_of_days-1))

			fmt.Printf("\n\n PX RET: %v", px_ret)
			fmt.Printf(" POOL: %v", available_pool_names[jj] + " " + available_pool_tkn0s[jj] + " | " + available_pool_tkn1s[jj])
			fmt.Printf(" RET EST: %v", available_pool_returns_est[jj])
			fmt.Printf(" RET HIST: %v\n\n", available_pool_returns_hist[jj])

			avg_returns = append(avg_returns, 252*total/float64((number_of_days-1)))
		*/