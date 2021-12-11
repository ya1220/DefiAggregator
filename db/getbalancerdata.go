package db

import (
	"DefiAggregator/db/token"
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"time"

	//	"math/rand"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/machinebox/graphql"
)

func (database *Database) getBalancerData(nc *Notifier_new) {
	//func (database *Database) getBalancerData() {

	poolname := "Balancer"

	if !database.do_records_need_updating(poolname) {
		fmt.Print(poolname)
		fmt.Print(" - Records are recent - no need to update..returning..")
		return
	}

	clientBalancer := graphql.NewClient("https://api.thegraph.com/subgraphs/name/balancer-labs/balancer")
	balancertopic := "0x908fb5ee8f16c6bc9bc3690973819f32a4d4b10188134543c88706e0e1d43378"

	// 0) Connect to client
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/e009cbb4a2bd4c28a3174ac7884f4b42")
	if err != nil {
		//log.Fatal(err)
		fmt.Print("Failed to connect to the API - ")
		fmt.Println(poolname)
		return
	}

	// 2 - declare queries
	reqBalancerListOfPools := graphql.NewRequest(`
	query {
		pools(first: 5, orderDirection: desc, orderBy: liquidity, where: {publicSwap: true}) {
		  id
		  tokensList
		  tokens {
			id
			address
			balance
			symbol
			decimals
		  }
		}
	  }
	`)

	reqBalancerByPoolID := graphql.NewRequest(`
		query ($poolid:String!){
			pool(id:$poolid) {
				id
				swapFee
				totalSwapVolume
				liquidity
				totalWeight
				tokensList
				tokens {
					id
					address
					balance
					symbol
					decimals
				}	
			}
		}
 	`)

	// get historical volume
	/*
			reqBalancerHistVolume := graphql.NewRequest(`
			query($pairid:String!){
				pool(id:$pairid) {
					swaps(first: 1000, skip: 0, orderBy: timestamp, orderDirection: desc){
						timestamp
						feeValue
						tokenInSym
						tokenOutSym
						tokenIn
						tokenOut
						tokenAmountIn
						tokenAmountOut
						poolLiquidity
					}
				}
			}
		`)
	*/
	// get TVL
	// get this pool % TVL
	// get BAL token price

	reqBalancerListOfPools.Var("key", "value")
	reqBalancerListOfPools.Header.Set("Cache-Control", "no-cache")
	reqBalancerByPoolID.Header.Set("Cache-Control", "no-cache")

	//	reqBalancerHistVolume.Var("key", "value")
	//	reqBalancerHistVolume.Header.Set("Cache-Control", "no-cache")

	ctx := context.Background()

	var respBalancerPoolList BalancerPoolList
	var respBalancerById BalancerById
	//	var respBalancerHistVolume BalancerHistVolumeQuery

	var respUniswapTicker UniswapTickerQuery // Used in Balancer to look up Uniswap IDs of 'ETH' etc
	var respUniswapHist UniswapHistQuery

	var Histrecord_2 HistoricalCurrencyData

	if err := clientBalancer.Run(ctx, reqBalancerListOfPools, &respBalancerPoolList); err != nil {
		//log.Fatal(err)
		fmt.Print("Could not connect to API..returning..")
		return
	}

	// PRINT - comment out later
	for i := 0; i < len(respBalancerPoolList.Pools); i++ {
		fmt.Print("i: ")
		fmt.Print(i)
		fmt.Print(" ")
		for j, tkn := range respBalancerPoolList.Pools[i].Tokens {
			fmt.Print(j)
			fmt.Print(": ")
			fmt.Print(tkn.Symbol) // respBalancerPoolList.Pools[i].Tokens[j].Symbol
			fmt.Print(" | ")
		}
		fmt.Print(" | n tkn: ")
		fmt.Println(len(respBalancerPoolList.Pools[i].Tokens))
	} // printing pools

	// GET CURRENT BLOCK NUMBER
	var current_block *big.Int
	var oldest_block *big.Int
	current_block = big.NewInt(0)

	// Get current block
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		//log.Fatal(err)
		fmt.Print("ERROR: could not get current block number")
		return
	}

	current_block = header.Number
	fmt.Printf("Current block number: %v\n\n", current_block)

	// Process received list of pools (PAIRS)
	for i := 0; i < len(respBalancerPoolList.Pools); i++ {

		number_of_pool_tokens := len(respBalancerPoolList.Pools[i].Tokens)
		pool_passes_filter := true

		// Remove
		if number_of_pool_tokens > 1 && number_of_pool_tokens < 3 {
			token0symbol := respBalancerPoolList.Pools[i].Tokens[0].Symbol
			token1symbol := respBalancerPoolList.Pools[i].Tokens[1].Symbol

			// if coin part of filter for all constituent tokens
			for tkn := range respBalancerPoolList.Pools[i].Tokens {
				if !isCoinPartOfFilter(respBalancerPoolList.Pools[i].Tokens[tkn].Symbol) {
					pool_passes_filter = false
				}
			}

			if pool_passes_filter {
				fmt.Printf("\n\n\n NEW pool passes filter i: %v", i)

				var tokenqueue []string

				pool_full_name_str := ""
				for _, tkn := range respBalancerPoolList.Pools[i].Tokens {
					if len(pool_full_name_str) > 0 {
						pool_full_name_str += "/"
					}
					pool_full_name_str += tkn.Symbol
					tokenqueue = append(tokenqueue, tkn.Symbol)
				}

				fmt.Printf(" name: %v", pool_full_name_str)
				fmt.Printf(" id: %v\n", respBalancerPoolList.Pools[i].ID)

				// BalancerFilteredPoolListPairs = append(BalancerFilteredPoolListPairs, pool_full_name_str)
				//fmt.Print(pool_full_name_str)
				//fmt.Print(" vs old method: ")
				//fmt.Print(token0symbol+"/"+token1symbol)

				pool_ratio_tkn0 := float64(0.0)
				pool_ratio_tkn1 := float64(0.0)

				for j := 0; j < len(tokenqueue); j++ {
					is_avail, number_of_days, newest_t := isHistDataAlreadyDownloadedDatabase(tokenqueue[j])

					fmt.Print("..checking if need to update HIST PX: ")
					fmt.Print(tokenqueue[j])
					fmt.Print(is_avail)
					fmt.Print(" | days: ")
					fmt.Print(number_of_days)
					fmt.Print(" | newest t: ")
					fmt.Print(newest_t)

					need_to_download_data := false

					if !is_avail {
						need_to_download_data = true
					}
					if time.Now().Unix()-newest_t > 60*60*24 {
						need_to_download_data = true
					}
					if number_of_days <= 3 {
						need_to_download_data = true
					}
					fmt.Print("..NEED TO UPDATE HIST PX Y/N: ")
					fmt.Println(need_to_download_data)

					if need_to_download_data {
						// Check if database already has historical data
						// Get Uniswap Ids of these tokens
						fmt.Print("...677 - Need to download HIST PX data for: ")
						fmt.Println(tokenqueue[j])
						fmt.Println(" | ")
						fmt.Print(convBalancerToken(tokenqueue[j]))

						database.uniswapreqdata.reqUniswapIDFromTokenTicker.Var("ticker", convBalancerToken(tokenqueue[j]))
						if err := database.uniswapreqdata.clientUniswap.Run(ctx, database.uniswapreqdata.reqUniswapIDFromTokenTicker, &respUniswapTicker); err != nil {
							fmt.Print("Failed connecting to Uniswap")
							return
							// log.Fatal(err)
						}
						// Download historical data for each token for which data is missing
						if len(respUniswapTicker.IDsforticker) >= 1 {
							// request data from uniswap using this queried ticker
							database.uniswapreqdata.reqUniswapHist.Var("tokenid", setUniswapQueryIDForToken(tokenqueue[j], respUniswapTicker.IDsforticker[0].ID))

							fmt.Print("Querying historical (in GETBALANCER) data from UNISWAP for: ")
							fmt.Print(tokenqueue[j])
							if err := database.uniswapreqdata.clientUniswap.Run(ctx, database.uniswapreqdata.reqUniswapHist, &respUniswapHist); err != nil {
								fmt.Print("Failed connecting to Uniswap")
								return
								//log.Fatal(err)
							}

							fmt.Printf("| returned days: %v", len(respUniswapHist.DailyTimeSeries))

							// if returned data - append it to database
							if len(respUniswapHist.DailyTimeSeries) > 0 {
								appendHistPriceDataToDb(NewHistPxDataFromRaw_aboveT(tokenqueue[j], respUniswapHist.DailyTimeSeries, newest_t))
							}
						} // if managed to find some IDs for this TOKEN
					} // if historical data needs updating
				} // tokenqueue loop ends

				// if historical data is in order - get current data
				reqBalancerByPoolID.Var("poolid", respBalancerPoolList.Pools[i].ID)

				if err := clientBalancer.Run(ctx, reqBalancerByPoolID, &respBalancerById); err != nil {
					//log.Fatal(err)
					fmt.Println("Failed to connect to API..")
					return
				}

				fmt.Printf("..FINISHED HIST PX DATA CHECK..MOVING ON TO VOLUME CHECKS--\n")

				days_ago := 5
				newest_t_raw := get_newest_timestamp_from_db(poolname, tokenqueue, respBalancerPoolList.Pools[i].ID)
				newest_db_vlm_record := time.Unix(newest_t_raw, 0)
				data_is_old := false
				oldest_lookup_time := time.Now().UTC()

				var dates []int64
				var tradingvolumes []float64
				var poolsizes []float64
				var fees []float64
				var interest []float64
				var utilization []float64

				// for checking
				var bal0f_arr []float64
				var bal1f_arr []float64

				fmt.Printf("Hrs since update of HIST VOLUME data: %v", time.Since(newest_db_vlm_record).Hours())
				fmt.Printf(" Newest record: %v\n", newest_db_vlm_record.Unix())

				if (time.Since(newest_db_vlm_record).Hours()) > 48 {
					data_is_old = true
					lookup_days := math.Min((time.Since(newest_db_vlm_record).Hours()-24.0)/24.0, float64(days_ago))
					fmt.Printf(" Lookup days: %v\n", lookup_days)

					oldest_lookup_time = oldest_lookup_time.AddDate(0, 0, -int(lookup_days))
				}
				fmt.Printf(" is HIST VOLUME data in Balancer db old: %v", data_is_old)

				if len(utilization) > 0 || len(fees) > 0 { /**/
				}

				// 1) If data is old and need to update it - Define pool specific parameters
				if data_is_old {
					fmt.Printf(" POOL addr hex: %v", respBalancerById.Pool.ID)

					BalancerpoolAddress := common.HexToAddress(respBalancerById.Pool.ID)
					// for len(tokenqueue)
					//bal_array_big
					//ball_array_float
					//var tokenaddresses []string

					tokenAddress0 := common.HexToAddress(respBalancerById.Pool.Tokens[0].Address)
					tokenAddress1 := common.HexToAddress(respBalancerById.Pool.Tokens[1].Address)

					fmt.Printf("type: %v", reflect.TypeOf(respBalancerById.Pool.Tokens[0].Address).String())

					instance0, err := token.NewToken(tokenAddress0, client)
					instance1, err := token.NewToken(tokenAddress1, client)

					if err != nil {
						fmt.Print("ERROR: failed to get token address in Balancer func")
						//log.Fatal(err)
						return
					}

					bal0, err := instance0.BalanceOf(&bind.CallOpts{}, BalancerpoolAddress)
					bal1, err := instance1.BalanceOf(&bind.CallOpts{}, BalancerpoolAddress)

					if err != nil {
						fmt.Print("ERROR: failed to get pool balance in Balancer func")
						return
					}

					fmt.Printf("| BAL0 BIG: %s", bal0)
					fmt.Printf("| BAL1 BIG: %s", bal1)

					fmt.Printf(" | t0: %s", tokenqueue[0])
					fmt.Printf(" | t1: %s", tokenqueue[1])

					bal0_float, _ := negPowI(bal0, int64(respBalancerById.Pool.Tokens[0].Decimals)).Float64()
					bal1_float, _ := negPowI(bal1, int64(respBalancerById.Pool.Tokens[1].Decimals)).Float64()

					fmt.Printf(" BALs POST CONV: %v", bal0_float)
					fmt.Printf(" | 1: %v\n", bal1_float)

					Histrecord_2 = getHistPriceDataForTokenPairFromDB(token0symbol, token1symbol)

					//2)  Find oldest block in our lookup date range
					oldest_block = new(big.Int).Set(current_block)

					time_to_start_accumulating := int64(0)
					time_to_stop_accumulating := int64(0)

					j := int64(0) // compute block id [days_ago] days away from now
					for {
						j -= 2000
						oldest_block.Add(oldest_block, big.NewInt(j))

						block, err := client.BlockByNumber(context.Background(), oldest_block)
						if err != nil {
							// log.Fatal(err)
							fmt.Print("ERROR: could not get oldest block")
							return
						}

						if block.Time() <= uint64(oldest_lookup_time.Unix()) {
							fmt.Print("OLDEST BLOCK IS AT: %v", block.Time())

							time_to_start_accumulating = BoDui64(block.Time()) + 24*60*60
							time_to_stop_accumulating = BoDi64(time.Now().UTC().Unix())

							fmt.Printf("time_to_start_accumulating: %v", time_to_start_accumulating)
							fmt.Printf("time_to_stop_accumulating: %v", time_to_stop_accumulating)

							break
						}
					}

					// time has to be > bod(block.time()) + 24H to start accumulating
					//3)  Query between oldest and current block for Balancer-specific addresses
					query := ethereum.FilterQuery{
						FromBlock: oldest_block,
						ToBlock:   nil, // = latest block
						Addresses: []common.Address{BalancerpoolAddress},
					}

					logsX, err := client.FilterLogs(context.Background(), query)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Printf("..N block logs: %v\n", len(logsX))

					cumulative_for_day := float64(0)
					t_prev := uint64(0)
					t_new := uint64(0)
					day_crossed := false

					last_idx := len(logsX) - 1
					skip := float64(5.0)

					//4)  Loop through received data and filter it again
					// For each transaction in logsX - check if it matches lookup criteria - add volume if does:
					for ii := 0; ii < len(logsX); ii++ {
						if logsX[ii].Topics[0] != common.HexToHash(balancertopic) {
							continue
						}

						recheck_block := false
						/*
							if int64(t_new) - BoDui64(t_new) < 60 * 60 * 18  {
								skip = 50
							} else {
								skip = 5
							}
						*/

						if math.Mod(float64(ii), skip) == 0 {
							recheck_block = true
						}
						if t_prev == 0 && t_new == 0 {
							recheck_block = true
						}
						if day_crossed {
							fmt.Print("Day X FLAG on..rechecking block")
							recheck_block = true
						}
						/*
							if math.Mod(float64(ii), 100) == 0 {
								fmt.Print(ii)
								fmt.Print("..")
							}
						*/
						if recheck_block { // Just for time
							block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(logsX[ii].BlockNumber)))
							if err != nil {
								log.Fatal(err)
							}
							t_prev = t_new       // uint
							t_new = block.Time() // uint

							if t_prev == 0 { // ?
								t_prev = t_new
							}
						} // check every [n] tx logs

						// Get the transaction
						txlog, err := client.TransactionReceipt(context.Background(), logsX[ii].TxHash)
						if err != nil {
							fmt.Print("ERROR: could not get transaction..")
							continue
							// log.Fatal(err)
						}

						//fmt.Println(" ")
						//fmt.Printf("t_prev: %v", t_prev)
						//fmt.Printf("| t_new: %v", t_new)

						start_of_day_for_block_time_prev := BoDui64(t_prev)
						start_of_day_for_block_time_new := BoDui64(t_new)
						fmt.Printf("ii: %v", ii)
						fmt.Printf(" | t raw: %v", t_new)
						fmt.Printf(" | SoD t_prev: %v", start_of_day_for_block_time_prev)
						fmt.Printf(" | SoD t_new: %v ", start_of_day_for_block_time_new)

						if start_of_day_for_block_time_new > start_of_day_for_block_time_prev {
							day_crossed = true
							fmt.Print("Day X! 345..")
						} else {
							day_crossed = false
						}

						t_within_limits := false

						if start_of_day_for_block_time_prev >= time_to_start_accumulating && start_of_day_for_block_time_prev < time_to_stop_accumulating {
							t_within_limits = true
						}

						fmt.Printf("t within limits?: %v", t_within_limits)

						// or if last
						if (day_crossed || ii == last_idx) && t_within_limits {
							fmt.Print("--DAY X'D!!!--")
							dates = append(dates, start_of_day_for_block_time_prev)
							v := convert_amt_from_to(cumulative_for_day, token1symbol, "USD", start_of_day_for_block_time_prev)
							tradingvolumes = append(tradingvolumes, v)

							//exch_rate_tkn0_into_tkn1 := float64(0.0)
							//_,k := MaxArgSlice(Histrecord_2.Date)
							//exch_rate_tkn0_into_tkn1 = Histrecord_2.Price[k] // Get most recent
							//fmt.Printf("exch: %v\n", exch_rate_tkn0_into_tkn1)
							// convert everything to token1

							sz0 := convert_amt_from_to(bal0_float, token0symbol, "USD", start_of_day_for_block_time_prev)
							sz1 := convert_amt_from_to(bal1_float, token1symbol, "USD", start_of_day_for_block_time_prev)

							pool_ratio_tkn0 = sz0 / (sz0 + sz1)
							pool_ratio_tkn1 = sz1 / (sz0 + sz1)

							fmt.Printf("ratio0: %v\n", pool_ratio_tkn0)
							fmt.Printf("ratio1: %v\n", pool_ratio_tkn1)

							poolsizes = append(poolsizes, sz0+sz1) // int64(bal0_float * exch_rate_tkn0_into_tkn1 + bal1_float)
							bal0f_arr = append(bal0f_arr, bal0_float)
							bal1f_arr = append(bal1f_arr, bal1_float)

							cumulative_for_day = 0.0
						} else if !day_crossed && t_within_limits {
							token0AD := respBalancerById.Pool.Tokens[0].Address //"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH
							token1AD := respBalancerById.Pool.Tokens[1].Address
							dec := respBalancerById.Pool.Tokens[1].Decimals

							// Get volume in tkn1 terms
							tkn1_volume := getVolumeFromTxLogBalancer(txlog.Logs, balancertopic, token0AD, token1AD, dec)
							cumulative_for_day += tkn1_volume // convert to usd
							day_crossed = false

							fmt.Print("+ @ ii: ")
							fmt.Print(ii)
							fmt.Print(" vlm: ")
							fmt.Print(tkn1_volume)

							fmt.Print(" cum: ")
							fmt.Println(cumulative_for_day)
						} else {
							fmt.Println(" ")
						} // if not day x - just keep accumulating
						// else - do nothing
					} // loop through log finished

					fmt.Printf("\n--------------SUMMARY DAILY VOLUME HIST DATA: ---------------\n")
					fmt.Printf("Got num DATES: %v\n", len(dates))

					for ii := 0; ii < len(dates); ii++ {
						fmt.Print("ii: ")
						fmt.Print(ii)
						interest = append(interest, 0.0)
						fmt.Print("| t: ")
						fmt.Print(dates[ii])
						fmt.Print("| volumes: ")
						fmt.Println(tradingvolumes[ii])

						vlm := tradingvolumes[ii]
						interest_x := float64(0)
						fees_x := float64(0)

						if dates[ii] > newest_t_raw {
							fmt.Printf("\nAppending HIST VOLUME data to db: %v\n", dates[ii])
							_ = appendHistVolumeDataToDb(poolname, tokenqueue, respBalancerPoolList.Pools[i].ID, dates[ii], vlm, poolsizes[ii], fees_x, interest_x, float64(0), pool_ratio_tkn0, pool_ratio_tkn1)
						}

					} // print loop
					fmt.Printf("\n--------------SUMMARY DAILY VOLUME HIST DATA END ---------------\n")
				} // if need to update data

				// Get this anyway - not just if data is old
				tkn0_pct_pool, tkn1_pct_pool := retrieve_pool_ratios(poolname, tokenqueue, respBalancerPoolList.Pools[i].ID)
				dates, tradingvolumes, poolsizes, fees, interest, utilization = retrieve_hist_pool_sizes_volumes_fees_ir(poolname, tokenqueue, respBalancerPoolList.Pools[i].ID)

				fmt.Printf("\n-------------------GETTING HISTORICAL VOLUME DATA COMPLETE----------------\n")

				fmt.Printf("--Volume data available: \n")
				//	fmt.Printf("len dt: %v\n", len(dates))
				//	fmt.Printf("len b0: %v\n", len(bal0f_arr))
				//	fmt.Printf("len b1: %v\n", len(bal1f_arr))

				for jjj := 0; jjj < len(dates); jjj++ {
					fmt.Printf("dt: %v", dates[jjj])
					fmt.Printf("| vlm (tkn1?): %v", tradingvolumes[jjj])
					fmt.Printf(" | sz: %v", poolsizes[jjj])
					if len(bal0f_arr) > jjj {
						fmt.Printf(" | bal0: %v", bal0f_arr[jjj])
						fmt.Printf(" | bal1: %v", bal1f_arr[jjj])
					}

					fmt.Printf("\n")
				}

				fmt.Print("--555--\n")
				//fmt.Printf("\n")

				future_daily_volume_est, future_pool_sz_est := estimate_future_balancer_volume_and_pool_sz(dates, tradingvolumes, poolsizes)
				historical_pool_sz_avg, historical_pool_daily_volume_avg := future_pool_sz_est, future_daily_volume_est

				_, max_date_i := MaxArgSlice(dates)

				//currentSize, _ := strconv.ParseFloat(respBalancerById.Pool.Liquidity, 64)
				//currentVolume, _ := strconv.ParseFloat(respBalancerById.Pool.TotalSwapVolume, 64) // No historical for now

				//fmt.Printf("Current pool sz UNISWAP: %v : ",currentSize) // what is this?
				//fmt.Printf(" | vlm UNISWAP: %v\n",currentVolume) // what timeframe?

				currentInterestrate := float64(0.00) // Zero for liquidity pool
				BalancerRewardPercentage, _ := strconv.ParseFloat(respBalancerById.Pool.SwapFee, 64)

				fmt.Printf("Reward pct: %v", BalancerRewardPercentage)

				if len(Histrecord_2.Date) == 0 || Histrecord_2.Ticker != pool_full_name_str {
					fmt.Print("currently loaded data: %v", Histrecord_2.Ticker)
					fmt.Print("Reloading histrecord2: token mismatch or 0 len: %v", pool_full_name_str)
					Histrecord_2 = getHistPriceDataForTokenPairFromDB(token0symbol, token1symbol)
				}

				volatility := calculatehistoricalvolatility(Histrecord_2, 30)
				imp_loss_hist := estimate_impermanent_loss_hist(volatility, 1, poolname)

				t_ago := BoDi64(time.Now().UTC().Unix()) - 24*60*60*30
				px_ago0 := convert_amt_from_to(1, token0symbol, "USD", t_ago)
				px_now0 := convert_amt_from_to(1, token0symbol, "USD", BoDi64(time.Now().UTC().Unix()))

				px_ago1 := convert_amt_from_to(1, token1symbol, "USD", t_ago)
				px_now1 := convert_amt_from_to(1, token1symbol, "USD", BoDi64(time.Now().UTC().Unix()))
				px_return_hist := tkn0_pct_pool[0]*px_now0/px_ago0 + tkn1_pct_pool[0]*px_now1/px_ago1
				fmt.Printf("px_return_hist: %v\n", px_return_hist)
				if px_ago0 == 0.0 || px_ago1 == 0.0 {
					Histrecord_3 := getHistPriceDataForTokenPairFromDB(token0symbol, "USD")
					Histrecord_4 := getHistPriceDataForTokenPairFromDB(token1symbol, "USD")

					p0 := calculate_price_return_x_days(Histrecord_3, 30)
					p1 := calculate_price_return_x_days(Histrecord_4, 30)
					px_return_hist = tkn0_pct_pool[0]*p0 + tkn1_pct_pool[0]*p1
					fmt.Printf("PX RETURN HIST NEW ALT: %v\n", px_return_hist) // BOTH WORK
				}

				if px_return_hist <= -1 {
					fmt.Printf("ERROR 893: WRONG PX RETURN HISTORICAL\n")
					px_return_hist = 0.0
					log.Fatal(err)
				}

				ROI_raw_est := calculateROI_raw_est(currentInterestrate, float64(BalancerRewardPercentage), float64(future_pool_sz_est), float64(future_daily_volume_est), imp_loss_hist)      // + imp
				ROI_vol_adj_est := calculateROI_vol_adj(ROI_raw_est, volatility)                                                                                                               // Sharpe ratio
				ROI_hist := calculateROI_hist(currentInterestrate, float64(BalancerRewardPercentage), historical_pool_sz_avg, historical_pool_daily_volume_avg, imp_loss_hist, px_return_hist) // + imp + hist

				var ratios []float64
				ratios = append(ratios, tkn0_pct_pool[0])
				ratios = append(ratios, tkn1_pct_pool[0])

				fmt.Print("| ROI_raw_est: ")
				fmt.Print(ROI_raw_est)
				fmt.Print("| ROI_vol_adj_est: ")
				fmt.Print(ROI_vol_adj_est)
				fmt.Print("| ROI_hist: ")
				fmt.Print(ROI_hist)
				/*
					fmt.Print("DECIMALS t0: ")
					fmt.Print(respBalancerById.Pool.Tokens[0].Symbol)
					fmt.Print(respBalancerById.Pool.Tokens[0].ID)
					fmt.Print(" | ")
					fmt.Print(respBalancerById.Pool.Tokens[0].Address)
					fmt.Print(" | ")
					fmt.Print(respBalancerById.Pool.Tokens[0].Decimals)
					fmt.Print(" | t1: ")
					fmt.Print(respBalancerById.Pool.Tokens[1].Symbol)
					fmt.Print(respBalancerById.Pool.Tokens[1].ID)
					fmt.Print(" | ")
					fmt.Print(respBalancerById.Pool.Tokens[1].Address)
					fmt.Print(" | ")
					fmt.Print(respBalancerById.Pool.Tokens[1].Decimals)
					fmt.Print(" | ")
				*/

				recordalreadyexists := false

				// CHECK IF NOT DUPLICATING RECORD - IF ALREADY EXISTS - UPDATE NOT APPEND
				for k := 0; k < len(database.PoolTokenPairReturns); k++ {
					// Means record already exists - UPDATE IT, DO NOT APPEND
					if database.PoolTokenPairReturns[k].Pair == pool_full_name_str && database.PoolTokenPairReturns[k].Pool == "Balancer" {
						recordalreadyexists = true
						database.PoolTokenPairReturns[k].PoolSize = poolsizes[max_date_i]        // float64(currentSize)
						database.PoolTokenPairReturns[k].PoolVolume = tradingvolumes[max_date_i] // float64(currentVolume)

						//database.PoolTokenPairReturns[k].PoolRatios = ratios
						database.PoolTokenPairReturns[k].PoolRatio0 = ratios[0]
						database.PoolTokenPairReturns[k].PoolRatio1 = ratios[1]

						database.PoolTokenPairReturns[k].ROI_raw_est = ROI_raw_est
						database.PoolTokenPairReturns[k].ROI_vol_adj_est = ROI_vol_adj_est
						database.PoolTokenPairReturns[k].ROI_hist = ROI_hist

						database.PoolTokenPairReturns[k].Volatility = volatility
						database.PoolTokenPairReturns[k].Yield = currentInterestrate

						nc.Notify_pooltable()
						nc.Notify_raw_and_optimised_pf()

						break
					}
				}

				// APPEND IF NEW
				if !recordalreadyexists {
					database.PoolTokenPairReturns = append(database.PoolTokenPairReturns, PoolTokenPairReturns{pool_full_name_str, poolsizes[max_date_i],
						tradingvolumes[max_date_i], ratios[0], ratios[1], currentInterestrate, poolname, volatility, ROI_raw_est, 0.0, 0.0, time.Now().Unix()})

					nc.Notify_pooltable()
					nc.Notify_raw_and_optimised_pf()
				}

				fmt.Printf("\n--------CURRENT i CYCLE COMPLETE-----------\n\n\n")
			} // if pool is within pre filtered list ends
		} // if pool has some tokens ends
	} // balancer pair loop closes
	// if pool len is == 2
	fmt.Println("BALANCER COMPLETED!!!!!")

} // balancer get data close

func getVolumeFromTxLogBalancer(logs []*types.Log, pooltopic string, token0AD string, token1AD string, token1decimals int) float64 {
	// func always gets token1 amounts
	//token0AD := token0 // "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599" // BTC
	//token1AD := token1 // "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" // WETH
	decimals := token1decimals
	varbytes := big.NewInt(0)
	token1_found := false

	if len(logs) <= 10 {
		for i := 0; i < len(logs); i++ {
			if len(logs[i].Topics) >= 3 {
				if logs[i].Topics[2] == common.HexToHash(token1AD) {
					token1_found = true
					if len(logs[i].Data) >= 32 {
						varbytes = new(big.Int).SetBytes(logs[i].Data[0:32])
						//fmt.Print(" | 0-32: ")
						//fmt.Print(varbytes)
					}
				}
			}
		}
		if !token1_found {
			for i := 0; i < len(logs); i++ {
				if len(logs[i].Topics) >= 3 {
					if logs[i].Topics[2] == common.HexToHash(token0AD) {
						if len(logs[i].Data) >= 64 {
							varbytes = new(big.Int).SetBytes(logs[i].Data[32:64])
							//fmt.Print(" | 0-32: ")
							//fmt.Print(varbytes)
						}
					}
				}
			}
		} // not found ether
		//end Short list
	} else {
		for i := 0; i < len(logs); i++ {
			if len(logs[i].Topics) >= 3 {
				if logs[i].Topics[2] == common.HexToHash(token1AD) {
					if len(logs[i].Data) >= 32 {
						varbytes = varbytes.Add(varbytes, new(big.Int).SetBytes(logs[i].Data[0:32]))
						// fmt.Print(varbytes)
					}
				}
			}
		}
	} // end long lists

	/*
		for i := 0; i < len(logs); i++ {
			//fmt.Print(" |  i:::: ")
			//fmt.Print(i)
			if len(logs[i].Data) >= 32 {
				//fmt.Print(" | 0-32: ")
				//fmt.Print(new(big.Int).SetBytes(logs[i].Data[0:32]))
			}
			if len(logs[i].Data) >= 64 {
				//fmt.Print(" | len: ")
				//fmt.Print(len(logs[i].Data))
				//fmt.Print(" | 32-64: ")
				//fmt.Print(new(big.Int).SetBytes(logs[i].Data[32:64]))
			}
			if len(logs[i].Topics) > 0 {
				//fmt.Print(" | t0: ")
				//fmt.Print(logs[i].Topics[0])
			}
			if len(logs[i].Topics) > 1 {
				//fmt.Print(" | t1: ")
				//fmt.Print(logs[i].Topics[1])
			}
			if len(logs[i].Topics) > 2 {
				//fmt.Print(" | t2: ")
				//fmt.Print(logs[i].Topics[2])
				//fmt.Print(" | ")
			}
			//fmt.Println(" ")
		}
	*/

	ten := big.NewInt(10)
	ten.Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	varbytes.Div(varbytes, ten)
	num, _ := new(big.Float).SetInt(varbytes).Float64()
	return num
}

func estimate_future_balancer_volume_and_pool_sz(dates []int64, tradingvolumes []float64, poolsizes []float64) (float64, float64) {
	future_volume_est := 0.0
	future_sz_est := 0.0

	var count float64
	var count_sz float64
	count = 0
	count_sz = 0

	for i := 0; i < len(dates); i++ {
		/*
			fmt.Print("ESTIMATING BALANCER FUTURE VOLUME + POOL SZ: ")
			fmt.Print("transaction volume: ")
			fmt.Print(histvolume.Pool.Swaps[i].TokenAmountIn)
			fmt.Print(" | pool sz (liquidity): ")
			fmt.Println(histvolume.Pool.Swaps[i].PoolLiquidity)
		*/
		//v, _ := strconv.ParseFloat(histvolume.Pool.Swaps[i].TokenAmountIn, 64) // double check the TokenAmounIn, only used to compile;
		//sz, _ := strconv.ParseFloat(histvolume.Pool.Swaps[i].PoolLiquidity, 64)

		v := tradingvolumes[i]
		sz := poolsizes[i]

		future_volume_est += float64(tradingvolumes[i])
		future_sz_est += float64(poolsizes[i]) // sz

		if v != 0.0 {
			count++
		}

		if sz != 0.0 {
			count_sz++
		}

	}

	// APPLY ADJUSTOR? 	// MEDIAN?	// TAKE OUT EXTREME VALUES TO NORMALISE?
	if count > 0 {
		future_volume_est = future_volume_est / count
	} else {
		future_volume_est = 0.0
	}

	if count_sz > 0 {
		future_sz_est = future_sz_est / count_sz
	} else {
		future_sz_est = 0.0
	}

	if math.IsNaN(float64(future_volume_est)) {
		// should never happen
		fmt.Println("ERROR IN FUTURE VOLUME - 999999999999999999555555555555555555")
		future_volume_est = -995.0
	}
	if math.IsNaN(float64(future_sz_est)) {
		// should never happen
		fmt.Println("ERROR IN FUTURE SZ - 999999999999999999666666666666666666")
		future_sz_est = -996.0
	}

	if math.IsInf(float64(future_volume_est), 0) {
		fmt.Println("ERROR IN FUTURE VOLUME - 999999999999999999555555555555555555")
		future_volume_est = -993.0
	}
	if math.IsInf(float64(future_sz_est), 0) {
		fmt.Println("ERROR IN FUTURE SZ - 999999999999999999666666666666666666")
		future_sz_est = -994.0
	}

	fmt.Printf("\nFuture volume est: %v | ", future_volume_est)
	fmt.Printf("Future sz est: %v | ", future_sz_est)

	return float64(future_volume_est), float64(future_sz_est) // USD
}
