package db

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/machinebox/graphql"
)

func calculate_historical_uniswap_volume_and_pool_sz(histvolume UniswapHistVolumeQuery) (float64, float64) {

	// implement
	return 0.0, 0.0
}

func estimate_future_uniswap_volume_and_pool_sz(histvolume UniswapHistVolumeQuery) (float64, float64) {
	future_volume_est := 0.0
	future_sz_est := 0.0

	var count float64
	var count_sz float64
	count = 0
	count_sz = 0

	for i := 0; i < len(histvolume.DailyTimeSeries); i++ {
		/*
			fmt.Print("daily volume usd: ")
			fmt.Print(histvolume.DailyTimeSeries[i].DailyVolumeUSD)
			fmt.Print(" | ")
			fmt.Print("Reserve USD: ")
			fmt.Println(histvolume.DailyTimeSeries[i].ReserveUSD)
		*/
		v, _ := strconv.ParseFloat(histvolume.DailyTimeSeries[i].DailyVolumeUSD, 64)
		sz, _ := strconv.ParseFloat(histvolume.DailyTimeSeries[i].ReserveUSD, 64)

		future_volume_est += v
		future_sz_est += sz

		if v != 0.0 {
			count++
		}

		if sz != 0.0 {
			count_sz++
		}

	}

	// APPLY ADJUSTOR?
	// MEDIAN?
	// TAKE OUT EXTREME VALUES TO NORMALISE?
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

	fmt.Print("Future volume est: ")
	fmt.Print(future_volume_est)
	fmt.Print(" | ")
	fmt.Print("Future sz est: ")
	fmt.Print(future_sz_est)

	return float64(future_volume_est), float64(future_sz_est) // USD
}
// uniswapreqdata UniswapInputStruct
func (database *Database) getUniswapData() {
/*
	// check if need to run at all
	// if data last updated less than 1 minute ago, do not recalculate
	recordalreadyexists := false
	latest_update_time := int64(0)
	poolname := "Uniswap"

	for k := 0; k < len(database.PoolTokenPairReturns); k++ {
		if database.PoolTokenPairReturns[k].Pool == poolname {
			recordalreadyexists = true // data not completely blank
			if database.PoolTokenPairReturns[k].Last_updated > latest_update_time {
				latest_update_time = database.PoolTokenPairReturns[k].Last_updated
			} // update if newer uniswap data item exists
		}
	}

	fmt.Print("Time of latest update for Uniswap: ")
	fmt.Print(latest_update_time)
	fmt.Print(" diff vs now: ")
	fmt.Println(time.Now().Unix() - latest_update_time)

	if recordalreadyexists && (time.Now().Unix() - latest_update_time) < 600 {
		// Do nothing - data is less than x minutes old
		fmt.Print("Uniswap data is recent - nothing to update!!")
		return
	}

	// otherwise recalculate:
*/

	poolname := "Uniswap"

	if !database.do_records_need_updating(poolname) {
		fmt.Print(poolname)
		fmt.Print(" - Records are recent - no need to update..returning..")
		return
	}

	reqUniswapListOfPools := graphql.NewRequest(`
	query{
		pairs(first: 100, orderBy: volumeUSD, orderDirection: desc) {
			id
			untrackedVolumeUSD
			volumeUSD
			token0 {
				id
				symbol
			}
			token1 {
				id
				symbol
			}
		}
	}
`)

	reqUniswapByPoolID := graphql.NewRequest(`
			query ($poolid:String!){
				pair(id:$poolid) {
					id
					untrackedVolumeUSD
					volumeUSD
					token0Price
					token1Price			
					token0 {
						id
						symbol
					}
					token1 {
						id
						symbol
					}
				}
			}
	`)

	// New

	reqUniswapAllPairs := graphql.NewRequest(`
query{				
	pairs(first: 1000, orderBy: reserveUSD, orderDirection: desc) {
	  id
	 token0 {
	   id
	   symbol
	 }
	 token1 {
	   id
	   symbol
	 }
	}
 }
 `)

	reqUniswapHistVolume := graphql.NewRequest(`
		query($pairid:String!){				
			pairDayDatas(first:30, orderBy: date, orderDirection: desc, 
				where: {pairAddress:$pairid}				
				) {
			id 
			date
			token0 {
				id
				symbol
			}
			token1 {
				id
				symbol
			}
			dailyVolumeToken0
			dailyVolumeToken1
			dailyVolumeUSD
			totalSupply
			reserveUSD
			}
		}
`)

	var respUniswapPoolList UniswapPoolList
	var respUniswapHist UniswapHistQuery
	var respUniswapById UniswapCurrentQuery
	var respUniswapHistVolume UniswapHistVolumeQuery
	var respUniswapPairList UniswapPairList

	reqUniswapListOfPools.Var("key", "value")
	reqUniswapListOfPools.Header.Set("Cache-Control", "no-cache")

	reqUniswapHistVolume.Var("key", "value")
	reqUniswapHistVolume.Header.Set("Cache-Control", "no-cache")

	ctx := context.Background()

	// 7b - UNISWAP
	var UniswapFilteredPoolList []string      // Pairs - IDS - 0x124145
	var UniswapFilteredPoolListPairs []string // Pairs - Tokens ETH/DAI
	var UniswapFilteredTokenList []string     // Tokens - ETH, DAI

	var Histrecord HistoricalCurrencyData

	// request all pairs
	if err := database.uniswapreqdata.clientUniswap.Run(ctx, reqUniswapAllPairs, &respUniswapPairList); err != nil {
		fmt.Print("Failed to connect to the API - Uniswap")
		return
		// log.Fatal(err)
	}

	// arrays to hold interesting pairs
	var UniswapPairOfInterestIDList []string
	var UniswapPairOfInterestIDToken0 []string
	var UniswapPairOfInterestIDToken1 []string

	for i := 0; i < len(respUniswapPairList.Pairs); i++ {
		// check if both tokens are of interest
		if isPoolPartOfFilter(respUniswapPairList.Pairs[i].Token0.Symbol, respUniswapPairList.Pairs[i].Token1.Symbol) {
			// save pair id to database
			UniswapPairOfInterestIDList = append(UniswapPairOfInterestIDList, respUniswapPairList.Pairs[i].ID)
			UniswapPairOfInterestIDToken0 = append(UniswapPairOfInterestIDToken0, respUniswapPairList.Pairs[i].Token0.Symbol)
			UniswapPairOfInterestIDToken1 = append(UniswapPairOfInterestIDToken1, respUniswapPairList.Pairs[i].Token1.Symbol)
		}
	}

	if err := database.uniswapreqdata.clientUniswap.Run(ctx, reqUniswapListOfPools, &respUniswapPoolList); err != nil {
		log.Fatal(err)
	}

	
		for i := 0; i < len(respUniswapPoolList.Pools); i++ {
			fmt.Print("i: ")
			fmt.Print(i)
			fmt.Print(" | ")
			fmt.Print(respUniswapPoolList.Pools[i].Token0.Symbol)
			fmt.Print(" | ")
			fmt.Println(respUniswapPoolList.Pools[i].Token1.Symbol)
		}
	

	for i := 0; i < len(respUniswapPoolList.Pools); i++ {
		// if len(respUniswapPoolList.Pools[i].Token0) > 1 {
		token0symbol := respUniswapPoolList.Pools[i].Token0.Symbol
		token1symbol := respUniswapPoolList.Pools[i].Token1.Symbol

		if isPoolPartOfFilter(token0symbol, token1symbol) {
			// Filter pools to allowed components (WETH, DAI, USDC, USDT)
			UniswapFilteredPoolList = append(UniswapFilteredPoolList, respUniswapPoolList.Pools[i].ID)
			UniswapFilteredPoolListPairs = append(UniswapFilteredPoolListPairs, token0symbol+"/"+token1symbol)

			var tokenqueue []string
			var tokenqueueIDs []string

			// Split list of pairs into single tokens
			if !stringInSlice(token0symbol, UniswapFilteredTokenList) {
				UniswapFilteredTokenList = append(UniswapFilteredTokenList, token0symbol)
				tokenqueue = append(tokenqueue, token0symbol)
				tokenqueueIDs = append(tokenqueueIDs, respUniswapPoolList.Pools[i].Token0.ID)
			}
			if !stringInSlice(token1symbol, UniswapFilteredTokenList) {
				UniswapFilteredTokenList = append(UniswapFilteredTokenList, token1symbol)
				tokenqueue = append(tokenqueue, token1symbol)
				tokenqueueIDs = append(tokenqueueIDs, respUniswapPoolList.Pools[i].Token1.ID)
			}

			// Check if database already has historical data
			for j := 0; j < len(tokenqueueIDs); j++ {
				is_avail, number_of_days, newest_t := isHistDataAlreadyDownloadedDatabase(tokenqueue[j])

				need_to_download_data := false

				if !is_avail {need_to_download_data = true}
				if time.Now().Unix() - newest_t > 60 * 60 * 24 {need_to_download_data = true}
				if number_of_days <= 3 {need_to_download_data = true}
				//if !isHistDataAlreadyDownloadedDatabase(tokenqueue[j]) {
					if need_to_download_data {
					// No need to get uniswap ids of these tokens
					// Download historical data for each token for which data is missing
					// request data from uniswap using this queried ticker
					//fmt.Print("setting token ids (shld be long hex value):: ")
					//fmt.Println(tokenqueueIDs[j])
					database.uniswapreqdata.reqUniswapHist.Var("tokenid", tokenqueueIDs[j])
					fmt.Print("Querying hist data: ")
					fmt.Print(tokenqueueIDs[j])
					fmt.Print(" | ")
					fmt.Print(tokenqueue[j])
					if err := database.uniswapreqdata.clientUniswap.Run(ctx, database.uniswapreqdata.reqUniswapHist, &respUniswapHist); err != nil {
						//log.Fatal(err)
						fmt.Print("Could not connect to Uniswap to query historical data..")
						return
					}
					fmt.Print("| downloaded days: ")
					fmt.Println(len(respUniswapHist.DailyTimeSeries))
					// if returned data - append it to database
					if len(respUniswapHist.DailyTimeSeries) > 0 {
						Histrecord = NewHistoricalCurrencyDataFromRaw(tokenqueue[j], respUniswapHist.DailyTimeSeries)
						appendHistPriceDataToDb(Histrecord)
						fmt.Println("....APPENDED TO DATABASE!..")
					}
				} // if historical data needs updating
			} // tokenqueue loop ends

			// if historical data is in order - get current data
			reqUniswapByPoolID.Var("poolid", respUniswapPoolList.Pools[i].ID)

			if err := database.uniswapreqdata.clientUniswap.Run(ctx, reqUniswapByPoolID, &respUniswapById); err != nil {
				log.Fatal(err)
			}

			// currentVolume, _ := strconv.ParseFloat(respUniswapById.Pair.VolumeUSD, 32) //
			currentPrice0, _ := strconv.ParseFloat(respUniswapById.Pair.Token0Price, 32) //
			currentPrice1, _ := strconv.ParseFloat(respUniswapById.Pair.Token1Price, 32) //
			currentPricePair := currentPrice0 / currentPrice1                            // which order is correct?
			if math.IsInf(currentPricePair, 0) {
				currentPricePair = -99.0
			}
			if math.IsNaN(currentPricePair) {
				currentPricePair = -99.9
			}
			currentInterestrate := float64(0.00)      // Zero for liquidity pool
			UniswapRewardPercentage := float64(0.003) // Placeholder

			var pairid string

			// find the pair id from 2 tokens
			// fmt.Println("TRYING TO MATCH TO PAIR ID: ")
			for jjj := 0; jjj < len(UniswapPairOfInterestIDList); jjj++ {
				matches := 0
				// respUniswapPoolList.Pools[i].Token1.Symbol
				//	fmt.Println(jjj)
				//	fmt.Println("Pair of interest: ")
				//	fmt.Println(UniswapPairOfInterestIDToken0[jjj] + "/" + UniswapPairOfInterestIDToken1[jjj])
				//	fmt.Println("Vs: ")
				//	fmt.Println(respUniswapPoolList.Pools[i].Token0.Symbol + "/" + respUniswapPoolList.Pools[i].Token1.Symbol)

				if UniswapPairOfInterestIDToken0[jjj] == respUniswapPoolList.Pools[i].Token0.Symbol {
					matches += 1
				}
				if UniswapPairOfInterestIDToken0[jjj] == respUniswapPoolList.Pools[i].Token1.Symbol {
					matches += 1
				}
				if UniswapPairOfInterestIDToken1[jjj] == respUniswapPoolList.Pools[i].Token0.Symbol {
					matches += 1
				}
				if UniswapPairOfInterestIDToken1[jjj] == respUniswapPoolList.Pools[i].Token1.Symbol {
					matches += 1
				}
				// if match on both
				if matches == 2 {
					// then found
					pairid = UniswapPairOfInterestIDList[jjj] //	fmt.Println("Matched!")
					break
				}

			}

			//fmt.Println("FOUND PAIR ID FOR: ")
			//fmt.Println(respUniswapPoolList.Pools[i].Token0.Symbol)
			//fmt.Println(respUniswapPoolList.Pools[i].Token1.Symbol)
			//fmt.Println(pairid)

			// find the dash - remove it
			pairid = strings.TrimRight(pairid, "-")
			fmt.Println(strings.TrimRight(pairid, "-"))

			reqUniswapHistVolume.Var("pairid", pairid) // respUniswapPoolList.Pools[i].ID
			if err := database.uniswapreqdata.clientUniswap.Run(ctx, reqUniswapHistVolume, &respUniswapHistVolume); err != nil {
				log.Fatal(err)
			}

			//fmt.Print("CURRENT VOLUME: ")
			currentVolume, _ := strconv.ParseFloat(respUniswapById.Pair.VolumeUSD, 32)
			fmt.Println(currentVolume)

			if len(respUniswapHistVolume.DailyTimeSeries) > 0 {
				fmt.Print("NOW PRINTING HISTORICAL VOLUME: ")
				fmt.Print(respUniswapHistVolume.DailyTimeSeries[0].Token0.Symbol)
				fmt.Print(" | ")
				fmt.Println(respUniswapHistVolume.DailyTimeSeries[0].Token1.Symbol)
			}

			future_daily_volume_est, future_pool_sz_est := estimate_future_uniswap_volume_and_pool_sz(respUniswapHistVolume)
			historical_pool_sz_avg, historical_pool_daily_volume_avg := future_pool_sz_est, future_daily_volume_est

			//fmt.Println("-----------ABOUT TO RUN NEW DATABASE RETRIEVAL FUNC---------------------------")
			//			xxx := getHistPriceDataForTokenPairFromDB(token0symbol, token1symbol)
			//			fmt.Println(xxx)
			//			volatility := calculatehistoricalvolatility(retrieveDataForTokensFromDatabase(token0symbol, token1symbol, database), 30)
			volatility := calculatehistoricalvolatility(getHistPriceDataForTokenPairFromDB(token0symbol, token1symbol), 30)
			//fmt.Println("-----------RAN NEW DATABASE RETRIEVAL FUNC---------------------------")

			fmt.Print("volatility hist for: ")
			fmt.Print(token0symbol)
			fmt.Print(" | ")
			fmt.Print(token1symbol)
			fmt.Print(" : ")
			fmt.Println(volatility)

			imp_loss_hist := estimate_impermanent_loss_hist(volatility, 1, poolname)
			fmt.Print("imp return hist: ")
			fmt.Print(imp_loss_hist)
			fmt.Print(" | ")

			px_return_hist := calculate_price_return_x_days(Histrecord, 30)

			fmt.Print("px return hist: ")
			fmt.Print(px_return_hist)
			fmt.Print(" | ")

			ROI_raw_est := calculateROI_raw_est(currentInterestrate, UniswapRewardPercentage, float64(future_pool_sz_est), float64(future_daily_volume_est), imp_loss_hist)      // + imp
			ROI_vol_adj_est := calculateROI_vol_adj(ROI_raw_est, volatility)                                                                                                     // Sharpe ratio
			ROI_hist := calculateROI_hist(currentInterestrate, UniswapRewardPercentage, historical_pool_sz_avg, historical_pool_daily_volume_avg, imp_loss_hist, px_return_hist) // + imp + hist
			
			var ratios []float64
			ratios = append(ratios,1)
			ratios = append(ratios,0)

					fmt.Print("ROI results are raw / sharpe / hist: ")
					fmt.Print(ROI_raw_est)
					fmt.Print(" | ")
					fmt.Print(ROI_vol_adj_est)
					fmt.Print(" | ")
					fmt.Print(ROI_hist)
					fmt.Println(" | ")
			
			var recordalreadyexists bool
			recordalreadyexists = false

			for k := 0; k < len(database.PoolTokenPairReturns); k++ {
				// Means record already exists - UPDATE IT, DO NOT APPEND
				if database.PoolTokenPairReturns[k].Pair == token0symbol+"/"+token1symbol && database.PoolTokenPairReturns[k].Pool == poolname {
					recordalreadyexists = true
					// Update
					database.PoolTokenPairReturns[k].PoolSize = float64(future_pool_sz_est)
					database.PoolTokenPairReturns[k].PoolVolume = float64(future_daily_volume_est)

					database.PoolTokenPairReturns[k].PoolRatio0 = ratios[0]
					database.PoolTokenPairReturns[k].PoolRatio1 = ratios[1]

					database.PoolTokenPairReturns[k].ROI_raw_est = ROI_raw_est
					database.PoolTokenPairReturns[k].ROI_vol_adj_est = ROI_vol_adj_est
					database.PoolTokenPairReturns[k].ROI_hist = ROI_hist

					database.PoolTokenPairReturns[k].Volatility = volatility
					database.PoolTokenPairReturns[k].Yield = currentInterestrate
					break
				}
			}

			// APPEND IF NEW
			if !recordalreadyexists {
				database.PoolTokenPairReturns = append(database.PoolTokenPairReturns, PoolTokenPairReturns{token0symbol + "/" + token1symbol, float64(future_pool_sz_est),
					float64(future_daily_volume_est), ratios[0], ratios[1], currentInterestrate, poolname, volatility, ROI_raw_est, ROI_vol_adj_est, ROI_hist, time.Now().Unix()})
			}
		} // if pool is within pre filtered list ends
		// } // if pool has some tokens ends
	} // Uniswap pair loop closes

	fmt.Print("--------------------UPDATING UNISWAP DATA COMPLTE-----------------------")
}
