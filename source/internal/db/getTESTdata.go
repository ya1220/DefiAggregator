package db

import (
	"fmt"
	"time"
	"math/rand"
//	"pusher/defi_aggregator/source/internal/notifier_new"
//	"pusher/defi_aggregator/source/internal/db/data_structures"
)

func (database *Database) GetTestData(nc *Notifier_new) {
	fmt.Printf("---RUNNING TEST DATA: %v..",time.Now().Unix())

	poolname := "TESTPOOL"

	if !database.do_records_need_updating(poolname) {
		fmt.Print(poolname)
		fmt.Print(" - Records are recent - no need to update..returning..")
		return
	}

	// create some random values here
	num_records := 25

	for i:= 0; i < num_records; i++ {

		//fmt.Printf("RUNNING ITERATION i:%v",i)

		ROI_raw_est := float64(0.03)
		ROI_vol_adj_est := float64(0.08)                                                                                                               // Sharpe ratio
		ROI_hist := float64(0.99)
	
		tokens := [4]string{"WETH","WBTC","DOGE","USDT"} // ,"DAI"
	
		randtoken0 := tokens[rand.Intn(len(tokens))]
		randtoken1 := tokens[rand.Intn(len(tokens))]
	
		var pool_full_name_str string
		var ratios []float64

		if randtoken0 != randtoken1 {
			pool_full_name_str = randtoken0 + "/" + randtoken1
	
			ratios = append(ratios, 0.73)
			ratios = append(ratios, 0.27)
	
			} else {
			pool_full_name_str = randtoken0

			ratios = append(ratios, 1.00)
			ratios = append(ratios, 0.0)
	
		}
	
		poolsize := float64(100)
		tradingvolume := float64(10)
		volatility := float64(0.25)
		currentInterestrate := float64(0.0124)
	
		// append some records
		recordalreadyexists := false
	
		// CHECK IF NOT DUPLICATING RECORD - IF ALREADY EXISTS - UPDATE NOT APPEND
		for k := 0; k < len(database.PoolTokenPairReturns); k++ {
			// Means record already exists - UPDATE IT, DO NOT APPEND
			if database.PoolTokenPairReturns[k].Pair == pool_full_name_str && database.PoolTokenPairReturns[k].Pool == poolname {
				recordalreadyexists = true
				database.PoolTokenPairReturns[k].PoolSize = poolsize // float64(currentSize) 
				database.PoolTokenPairReturns[k].PoolVolume =  tradingvolume // float64(currentVolume)

				database.PoolTokenPairReturns[k].PoolRatio0 = ratios[0]
				database.PoolTokenPairReturns[k].PoolRatio1 = ratios[1]

				database.PoolTokenPairReturns[k].ROI_raw_est = ROI_raw_est
				database.PoolTokenPairReturns[k].ROI_vol_adj_est = ROI_vol_adj_est
				database.PoolTokenPairReturns[k].ROI_hist = ROI_hist
	
				database.PoolTokenPairReturns[k].Volatility = volatility
				database.PoolTokenPairReturns[k].Yield = currentInterestrate
				// notify here
				//fmt.Print("About to notify 0..")
		
		//**************************************************//
			//	nc.Notify_pooltable()
			//	nc.Notify_raw_and_optimised_pf()
		//**************************************************//
		
		
		//fmt.Print("Notified UPDATE OF EXISTING record!!")
				break
			}
		}
	
		// APPEND IF NEW
		if !recordalreadyexists {
			database.PoolTokenPairReturns = append(database.PoolTokenPairReturns, PoolTokenPairReturns{pool_full_name_str, poolsize,
				tradingvolume, ratios[0],ratios[1], currentInterestrate, poolname, volatility, ROI_raw_est, 0.0, 0.0, time.Now().Unix()})
				// notify here
				//fmt.Print("About to notify 1..")

		//**************************************************//
			//	nc.Notify_pooltable()
			//	nc.Notify_raw_and_optimised_pf()
		//**************************************************//

		//fmt.Print("Notified NEW record!!")
			}
	
			//time.Sleep(1 * time.Second)
	} // loop
		fmt.Println("-------------------RAN TEST DATA ITERATION-------------------------")


		nc.Notify_pooltable()
		nc.Notify_raw_and_optimised_pf()

	}