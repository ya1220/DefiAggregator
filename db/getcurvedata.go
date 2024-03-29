package db

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"DefiAggregator/db/curveRegistry"
	"DefiAggregator/db/token"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type CurvePoolData struct {
	poolAddress common.Address
	//poolCurrentBalances [8]*big.Int
	assetAddresses     [8]common.Address
	assetDecimals      [8]*big.Int
	volumes            []*[8]*big.Int
	fees               []*[8]*big.Float
	balances           []*[8]*big.Int
	normalisedBalances []*big.Float
}

func conv_curve_token_to_uniswap(curve_token string) string {
	assetName := curve_token

	if curve_token == "yDAI" {
		assetName = "DAI"
	} else if curve_token == "cDAI" {
		assetName = "DAI"
	} else if curve_token == "ycDAI" {
		assetName = "DAI"
	} else if curve_token == "ycUSDC" {
		assetName = "USDC"
	} else if curve_token == "yBUSD" {
		assetName = "USDC"
	} else if curve_token == "oBTC" {
		assetName = "WBTC"
	} else if curve_token == "HBTC" {
		assetName = "WBTC"
	} else if curve_token == "yUSDT" {
		assetName = "USDT"
	} else if curve_token == "oBTC" {
		assetName = "WBTC"
	} else if curve_token == "TBTC" {
		assetName = "WBTC"
	} else if curve_token == "BBTC" {
		assetName = "WBTC"
	} else if curve_token == "pBTC" {
		assetName = "WBTC"
	} else if curve_token == "oBTC" {
		assetName = "WBTC"
	} else if curve_token == "renBTC" {
		assetName = "WBTC"
	} else if curve_token == "GUSD" {
		assetName = "USDC"
	} else if curve_token == "HUSD" {
		assetName = "USDC"
	} else if curve_token == "mUSD" {
		assetName = "USDC"
	} else if curve_token == "sUSD" {
		assetName = "USDC"
	} else if curve_token == "cUSDC" {
		assetName = "USDC"
	} else if curve_token == "yUSDC" {
		assetName = "USDC"
	}

	return assetName
}

func (database *Database) getCurveData() {

	poolname := "Curve"

	if !database.do_records_need_updating(poolname) {
		fmt.Print(poolname)
		fmt.Print(" - Records are recent - no need to update..returning..")
		return
	}

	fmt.Println("GETTING CURVE DATA!!!")
	// Connecting to client
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/e009cbb4a2bd4c28a3174ac7884f4b42")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	var respUniswapTicker UniswapTickerQuery // Used to look up Uniswap IDs of 'ETH' etc
	var respUniswapHist UniswapHistQuery
	var Histrecord HistoricalCurrencyData
	// Creating a contract instance
	var curveRegistryAddress = common.HexToAddress("0x7D86446dDb609eD0F5f8684AcF30380a356b2B4c")
	provider, err := curveRegistry.NewMain(curveRegistryAddress, client)

	if err != nil {
		log.Fatal(err)
	}

	pools_to_pull := int64(16)
	var pool_addresses []common.Address
	var i int64

	// Main loop
	for i = 0; i < pools_to_pull; i++ {
		pool_address, _ := provider.PoolList(&bind.CallOpts{}, big.NewInt(i))
		pool_addresses = append(pool_addresses, pool_address)
		// got list of pools

		fmt.Print("Pool Address: ")
		fmt.Println(pool_address)

		//Getting token Adresses:
		coin_addresses, err := provider.GetCoins(&bind.CallOpts{}, pool_address)
		if err != nil {
			log.Fatal(err)
		}

		var tokenqueue []string
		var decimals []int64
		var balances []float64

		var dates []int64
		var tradingvolumes []float64
		var fees []float64
		var poolsizes []float64
		var interest []float64
		var utilization []float64
		if len(utilization) > 0 || len(interest) > 0 {
			// Golang 'usage' for otherwise un-used variables
			fmt.Print("placeholder")
		}

		//skips pool if token not in filter.
		skip_pool := false
		// get actual token names goes here
		for j := 0; j < len(coin_addresses); j++ {
			fmt.Print("j: ")
			fmt.Print(j)
			fmt.Print(" | ")
			if coin_addresses[j] == common.HexToAddress("0x0000000000000000000000000000000000000000") {
				continue
			}
			if coin_addresses[j] == common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE") {
				continue
			}

			instance, err := token.NewToken(coin_addresses[j], client)
			if err != nil {
				fmt.Println("failed getting instance")
				log.Fatal(err)
			}
			name, err := instance.Symbol(&bind.CallOpts{})
			if err != nil {
				fmt.Print("failed getting symbol for: ")
				fmt.Println(coin_addresses[j])
				log.Fatal(err)
			}
			name = conv_curve_token_to_uniswap(name)

			fmt.Print("token: ")
			fmt.Println(name)

			if !isCoinPartOfFilter(name) {
				// coin not part of our filter - skip this whole pool
				skip_pool = true
				fmt.Print("Token: ")
				fmt.Print(name)
				fmt.Print(" --- not part of filter..SKIPPING POOL..")
				break
			}
			tokenqueue = append(tokenqueue, name)

			is_avail, number_of_days, newest_t := isHistDataAlreadyDownloadedDatabase(tokenqueue[j])

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
			//if !isHistDataAlreadyDownloadedDatabase(tokenqueue[j]) {
			if need_to_download_data {
				// uniswap check
				//if !isHistDataAlreadyDownloadedDatabase(tokenqueue[j]) {
				fmt.Print("In Uniswap hist data check..")
				// Get Uniswap Ids of these tokens
				database.uniswapreqdata.reqUniswapIDFromTokenTicker.Var("ticker", tokenqueue[j])
				if err := database.uniswapreqdata.clientUniswap.Run(ctx, database.uniswapreqdata.reqUniswapIDFromTokenTicker, &respUniswapTicker); err != nil {
					log.Fatal(err)
				}
				// Download historical data for each token for which data is missing
				if len(respUniswapTicker.IDsforticker) >= 1 {
					// request data from uniswap using this queried ticker
					database.uniswapreqdata.reqUniswapHist.Var("tokenid", setUniswapQueryIDForToken(tokenqueue[j], respUniswapTicker.IDsforticker[0].ID))
					fmt.Print("Querying historical (in GETCURVE) data from UNISWAP for: ")
					fmt.Print(tokenqueue[j])
					if err := database.uniswapreqdata.clientUniswap.Run(ctx, database.uniswapreqdata.reqUniswapHist, &respUniswapHist); err != nil {
						log.Fatal(err)
					}

					fmt.Print("| returned days: ")
					fmt.Println(len(respUniswapHist.DailyTimeSeries))

					// if returned data - append it to database
					if len(respUniswapHist.DailyTimeSeries) > 0 {
						Histrecord = NewHistoricalCurrencyDataFromRaw(tokenqueue[j], respUniswapHist.DailyTimeSeries)
						appendHistPriceDataToDb(Histrecord)
					}
				} // if managed to find some IDs for this TOKEN
			} // if historical data needs updating

		} // j loop through coins

		coin_decimals, err := provider.GetDecimals(&bind.CallOpts{}, pool_address)
		// fmt.Print(len(coin_decimals))
		if err != nil {
			log.Fatal(err)
		}

		// copy decimals to a regular int format
		for jz := 0; jz < len(tokenqueue); jz++ {
			decimals = append(decimals, coin_decimals[jz].Int64())
		}

		current_coin_balances, err := provider.GetBalances(&bind.CallOpts{}, pool_address)

		fmt.Print("len of decimals: ")
		fmt.Print(len(decimals))
		fmt.Print("len of curr coin balances: ")
		fmt.Print(len(current_coin_balances))

		for jx := 0; jx < len(decimals); jx++ {
			balance := new(big.Float).SetInt(current_coin_balances[jx])
			fmt.Print("jx: ")
			fmt.Print(jx)
			fmt.Print("balance: ")
			fmt.Print(balance)
			balance = negPow(balance, decimals[jx])
			b, _ := balance.Float64()
			balances = append(balances, b)
		}

		fmt.Print("len of balances: ")
		fmt.Print(len(balances))

		fmt.Print("sz of tokenqueue: ")
		fmt.Print(len(tokenqueue))

		for jj := 0; jj < len(tokenqueue); jj++ {
			fmt.Print("jj: ")
			fmt.Print(tokenqueue[jj])
			fmt.Print(" | ")
		}

		fmt.Print(" | number of tokenqueue items: ")
		fmt.Println(len(tokenqueue))

		// if the next pool flag is true, go on to the next pool
		if skip_pool {
			fmt.Println("Skipping pool! Found tokens not in our permitted list")
			continue
		}

		if len(tokenqueue) > 1 && len(tokenqueue) < 5 {
			fmt.Print("getting oldest record: ")
			newest_db_vlm_record := time.Unix(get_newest_timestamp_from_db(poolname, tokenqueue, ""), 0)
			fmt.Print(" OLDEST RECORD IS AT: ")
			fmt.Print(newest_db_vlm_record)

			data_is_old := false

			if (time.Since(newest_db_vlm_record).Hours()) > 24 {
				data_is_old = true
			}
			fmt.Print("is data old: ")
			fmt.Print(data_is_old)

			if data_is_old { // download it
				//3)  Query between oldest and current block for Curve-specific addresses
				dates, tradingvolumes, poolsizes, fees, interest = curveGetPoolVolume(pool_address, client, balances, tokenqueue, decimals)

				fmt.Print("dates retrieved from curveGetPoolVolume: ")
				fmt.Print(len(dates))
				fmt.Print(" | tradingvolumes : ")
				fmt.Print(len(tradingvolumes))
				fmt.Print(" | poolsizes: ")
				fmt.Print(len(poolsizes))
				fmt.Print(" | fees: ")
				fmt.Print(len(fees))
				fmt.Print(" | interest: ")
				fmt.Print(len(interest))

				for jjj := 0; jjj < len(dates); jjj++ {
					// interest[jjj]
					recordID := appendHistVolumeDataToDb(poolname, tokenqueue, "", dates[jjj], tradingvolumes[jjj], fees[jjj], poolsizes[jjj], 0.0, float64(0), float64(1), float64(0))
					if len(recordID) == 0 {
						fmt.Print(recordID)
					}
				}

			} // if data is old

			if !data_is_old { // else: data is not old
				dates, tradingvolumes, poolsizes, fees, interest, utilization = retrieve_hist_pool_sizes_volumes_fees_ir(poolname, tokenqueue, "")
			}

			// ROI calculation
			currentSize := float64(0)
			for jkx := 0; jkx < len(poolsizes); jkx++ {
				currentSize += float64(poolsizes[jkx])
			}
			/*
				if len(poolsizes) > 0 {
					currentSize = float64(poolsizes[len(poolsizes)-1])
					fmt.Print("poolsizes: ")
					fmt.Print(poolsizes[0])
				}*/

			currentVolume := 0.0
			for jkx := 0; jkx < len(tradingvolumes); jkx++ {
				currentVolume += float64(tradingvolumes[jkx])
				//	fmt.Print("volumes: ")
				//	fmt.Print(tradingvolumes[jkx])
			}

			// Use balancer function - also applicable for curve
			future_daily_volume_est, future_pool_sz_est := estimate_future_balancer_volume_and_pool_sz(dates, tradingvolumes, poolsizes)
			historical_pool_sz_avg, historical_pool_daily_volume_avg := future_pool_sz_est, future_daily_volume_est
			currentInterestrate := float64(0.00) // POPULATE

			CurveRewardPercentage := 0.02 // 2% standard across curve pools
			volatility := calculatehistoricalvolatility(getHistPriceDataForTokenPairFromDB(tokenqueue[0], tokenqueue[1]), 30)
			volatility1 := float64(0.0)
			volatility2 := float64(0.0)

			if len(tokenqueue) >= 3 {
				volatility1 = calculatehistoricalvolatility(getHistPriceDataForTokenPairFromDB(tokenqueue[1], tokenqueue[2]), 30)
			}
			if len(tokenqueue) >= 4 {
				volatility2 = calculatehistoricalvolatility(getHistPriceDataForTokenPairFromDB(tokenqueue[2], tokenqueue[3]), 30)
			}

			totl := volatility
			vol_count := float64(1.0)
			if volatility1 > 0 {
				totl += volatility1
				vol_count++
			}

			if volatility2 > 0 {
				totl += volatility2
				vol_count++
			}

			//AVG volatility
			volatility = totl / vol_count

			fmt.Print("avg vol: ")

			imp_loss_hist := estimate_impermanent_loss_hist(volatility, 1, "Curve")
			px_return_hist := calculate_price_return_x_days(Histrecord, 30)

			ROI_raw_est := calculateROI_raw_est(currentInterestrate, float64(CurveRewardPercentage), float64(future_pool_sz_est), float64(future_daily_volume_est), imp_loss_hist) // + imp
			if ROI_raw_est > 50 {
				ROI_raw_est = ROI_raw_est / 20
			}
			ROI_vol_adj_est := calculateROI_vol_adj(ROI_raw_est, volatility)                                                                                                            // Sharpe ratio
			ROI_hist := calculateROI_hist(currentInterestrate, float64(CurveRewardPercentage), historical_pool_sz_avg, historical_pool_daily_volume_avg, imp_loss_hist, px_return_hist) // + imp + hist

			var ratios []float64
			ratios = append(ratios, 1)
			ratios = append(ratios, 0)
			/*
				fmt.Print("| ROI_raw_est: ")
				fmt.Print(ROI_raw_est)
				fmt.Print("| ROI_vol_adj_est: ")
				fmt.Print(ROI_vol_adj_est)
				fmt.Print("| ROI_hist: ")
				fmt.Print(ROI_hist)
			*/
			var recordalreadyexists bool
			recordalreadyexists = false

			str := tokenqueue[0] + "/" + tokenqueue[1]
			if len(tokenqueue) == 3 {
				str = str + "/" + tokenqueue[2]
			}
			if len(tokenqueue) == 4 {
				str = str + "/" + tokenqueue[3]
			}

			// CHECK IF NOT DUPLICATING RECORD - IF ALREADY EXISTS - UPDATE NOT APPEND
			for k := 0; k < len(database.PoolTokenPairReturns); k++ {
				// Means record already exists - UPDATE IT, DO NOT APPEND
				if database.PoolTokenPairReturns[k].Pair == str && database.PoolTokenPairReturns[k].Pool == "Curve" {
					recordalreadyexists = true
					database.PoolTokenPairReturns[k].PoolSize = float64(currentSize)
					database.PoolTokenPairReturns[k].PoolVolume = float64(currentVolume)

					database.PoolTokenPairReturns[k].PoolRatio0 = ratios[0]
					database.PoolTokenPairReturns[k].PoolRatio1 = ratios[1]

					database.PoolTokenPairReturns[k].ROI_raw_est = ROI_raw_est
					database.PoolTokenPairReturns[k].ROI_vol_adj_est = ROI_vol_adj_est
					database.PoolTokenPairReturns[k].ROI_hist = ROI_hist

					database.PoolTokenPairReturns[k].Volatility = volatility
					database.PoolTokenPairReturns[k].Yield = currentInterestrate
					break
				}
			} // k loop ends

			// APPEND IF NEW
			if !recordalreadyexists {
				database.PoolTokenPairReturns = append(database.PoolTokenPairReturns, PoolTokenPairReturns{str, float64(currentSize),
					float64(currentVolume), ratios[0], ratios[1], currentInterestrate, "Curve", volatility, ROI_raw_est, 0.0, 0.0, time.Now().Unix()})
			}
			fmt.Println("APPENDED CURVE DATA ITERATION")

		} // if len > 1
	} // pools to pull loop ends
	fmt.Println("CURVE COMPLETED!!!!!")
} // Get Curve Data closes

func curveGetPoolVolume(pool_address common.Address, client *ethclient.Client, balances []float64, tokenqueue []string, decimals []int64) ([]int64, []float64, []float64, []float64, []float64) {
	poolTopics := []string{"0x8b3e96f2b889fa771c53c981b40daf005f63f637f1869f707052d15a3dd97140" /* "0xd013ca23e77a65003c2c659c5442c00c805371b7fc1ebd4c206c41d1536bd90b"*/}

	var current_block *big.Int
	var oldest_block *big.Int
	current_block = big.NewInt(0)

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	current_block = header.Number

	fmt.Print("current block: ")
	fmt.Print(current_block)

	oldest_block = new(big.Int).Set(current_block)
	days_ago := 1
	oldest_lookup_time := time.Now()
	oldest_lookup_time = oldest_lookup_time.AddDate(0, 0, -days_ago)
	//fmt.Print("oldest_lookup_time: ")
	//fmt.Print(oldest_lookup_time)

	j := int64(0) // compute block id [days_ago] days away from now
	for {
		j -= 2000
		oldest_block.Add(oldest_block, big.NewInt(j))

		block, err := client.BlockByNumber(context.Background(), oldest_block)
		if err != nil {
			log.Fatal(err)
		}

		if block.Time() <= uint64(oldest_lookup_time.Unix()) {
			/*
				fmt.Print("oldest lkp block: ")
				fmt.Println(oldest_block)
				fmt.Print(" | t: ")
				fmt.Print(block.Time())
				fmt.Print("| curr blk - oldest blk: ")
			*/
			diff := current_block.Sub(current_block, oldest_block)
			fmt.Println(diff)
			break
		}
	}

	//3)  Query between oldest and current block for Curve-specific addresses
	fmt.Print("Querying pool address: ")
	fmt.Println(pool_address)

	query := ethereum.FilterQuery{
		FromBlock: oldest_block,
		ToBlock:   nil, // = latest block
		Addresses: []common.Address{pool_address},
	}

	logsX, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("DOWNLOADED logsX: ")
	fmt.Print(len(logsX))
	//4)  Loop through received data and filter it again
	// For each transaction in logsX - check if it matches lookup criteria - add volume if does:
	var fees []float64
	var tradingvolumes []float64
	var dates []int64
	var poolsizes []float64
	var interest []float64
	cumulative_for_day_fees := float64(0)
	cumulative_for_day_volume := float64(0)
	t_prev := uint64(0)
	t_new := uint64(0)

	// which symbols do we need to get here??
	Histrecord_2 := getHistPriceDataForTokenPairFromDB(tokenqueue[0], tokenqueue[1])

	fmt.Print(" len(logsX): ")
	fmt.Print(len(logsX))

	// can append balances out of loop
	totalBal := float64(0)
	//fmt.Print("in ballen: ")
	//fmt.Print(len(balances))
	for balLen := 0; balLen < len(balances); balLen++ {
		/*	fmt.Print("tokenqueue[balLen]: ")
			fmt.Print(tokenqueue[balLen])
			fmt.Print(" | b[balLen]: ")
			fmt.Print(balances[balLen])
		*/
		totalBal += balances[balLen]
	}
	//	fmt.Print(" | total bal: ")
	//	fmt.Print(totalBal)
	//	fmt.Print("..")

	// loop through whole log
	for i := 0; i < len(logsX); i++ {
		if logsX[i].Topics[0] != common.HexToHash(poolTopics[0]) /*&& logsX[i].Topics[0] != common.HexToHash(poolTopics[1])*/ {
			continue
		}

		fmt.Print(i)
		fmt.Print("..")
		// Get date from block number
		if math.Mod(float64(i), 20) == 0 {
			block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(logsX[i].BlockNumber)))
			if err != nil {
				log.Fatal(err)
			}
			// Here we have to add summing them up by day - not just total
			t_prev = t_new       // uint
			t_new = block.Time() // uint

			//fmt.Print(t_new)
			//fmt.Print("..")
		}

		txlog, err := client.TransactionReceipt(context.Background(), logsX[i].TxHash)

		if err != nil {
			log.Fatal(err)
		}

		if t_prev == 0 || (t_new-uint64(math.Mod(float64(t_new), 86400)))/86400 !=
			(t_prev-uint64(math.Mod(float64(t_prev), 86400)))/86400 { // 1 day

			dates = append(dates, int64(BoD(time.Unix(int64(t_new), 0)).Unix()))
			fees = append(fees, cumulative_for_day_fees)
			tradingvolumes = append(tradingvolumes, cumulative_for_day_volume)

			if len(Histrecord_2.Price) >= 1 {
			} // convert to token1

			poolsizes = append(poolsizes, totalBal) // bal.Int64()
			//fmt.Print("appended to poolsizes..")
			//fmt.Print(poolsizes[len(poolsizes)-1])
			cumulative_for_day_fees = 0
			cumulative_for_day_volume = 0 // reset days tally if day threshold crossed
		} else { //-------------
			asset0_index, asset0_volume, asset1_index, asset1_volume := getTradingVolumeFromTxLogCurve(txlog.Logs, poolTopics)
			//	fmt.Print("asset idx: ")
			//	fmt.Print(asset0_index)
			//	fmt.Print(" | ")
			//	fmt.Print(asset1_index)

			//	fmt.Print("asset0 vlm raw: ")
			//	fmt.Print(asset0_volume)
			//	fmt.Print(" | asset1 vlm raw: ")
			//	fmt.Print(asset1_volume)

			exch0 := float64(1.0) // assumed to be stablecoin
			exch1 := float64(1.0) // assumed to be stablecoin
			if asset0_index < int64(len(tokenqueue)) && asset1_index < int64(len(tokenqueue)) {
				if isTokenStableCoin(tokenqueue[asset0_index]) {
					exch0 = float64(1.0)
				}
				if isTokenStableCoin(tokenqueue[asset1_index]) {
					exch1 = float64(1.0)
				}
			}

			// sz_0 := int64(float64(asset0_volume) * exch0)
			// sz_1 := int64(float64(asset1_volume) * exch1)
			if asset0_index < int64(len(decimals)) && asset1_index < int64(len(decimals)) {
				sz_0 := negPowF(float64(asset0_volume), decimals[asset0_index]) * exch0
				sz_1 := negPowF(float64(asset1_volume), decimals[asset1_index]) * exch1

				if sz_0 < 0 {
					sz_0 = -sz_0
				}
				if sz_1 < 0 {
					sz_1 = -sz_1
				}
				/*
					fmt.Print("decimals[asset0_index]: ")
					fmt.Print(decimals[asset0_index])
					fmt.Print(" sz0: ")
					fmt.Print(sz_0)
					fmt.Print(" | sz1: ")
					fmt.Print(sz_1)
					fmt.Print(" | ")
				*/
				pool_fee := 0.02 // Standard curve fee

				f0 := float64(sz_0) * pool_fee
				f1 := float64(sz_1) * pool_fee
				// Add it to tally for that day
				cumulative_for_day_fees += (f0 + f1)
				cumulative_for_day_volume += sz_1 // sz_0
				cumulative_for_day_volume += sz_0 // sz_0
			}
		} // else - if not a new day
	} // loop through logs ends

	return dates, tradingvolumes, poolsizes, fees, interest
}

func decodeBytesCurve(log *types.Log) (int, *big.Int, int, *big.Int) {

	asset0_index, _ := strconv.Atoi((new(big.Int).SetBytes(log.Data[0:32])).String())
	asset0_volume := new(big.Int).SetBytes(log.Data[32:64])

	asset1_index, _ := strconv.Atoi((new(big.Int).SetBytes(log.Data[64:96])).String())
	asset1_volume := new(big.Int).SetBytes(log.Data[96:128])

	return asset0_index, asset0_volume, asset1_index, asset1_volume
}

func getTradingVolumeFromTxLogCurve(logs []*types.Log, pooltopics []string) (int64, float64, int64, float64) { //Should return an int not big int

	var firstLog *types.Log
	//var lastLog *types.Log

	for _, log := range logs {
		if log.Topics[0] != common.HexToHash(pooltopics[0]) /*&& log.Topics[0] != common.HexToHash(pooltopics[1])*/ {
			continue
		}
		if firstLog == nil {
			firstLog = log
		}
		//lastLog = log
	}

	if firstLog == nil { // could not find any valid swaps, thus the transaction failed
		return -1, 0, -1, 0
	}
	asset0_index, asset0_volume, asset1_index, asset1_volume := decodeBytesCurve(firstLog)

	v0, _ := new(big.Float).SetInt(asset0_volume).Float64()
	v1, _ := new(big.Float).SetInt(asset1_volume).Float64()

	return int64(asset0_index), v0, int64(asset1_index), v1
}
