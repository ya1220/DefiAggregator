package db

/*
import (
	"context"
	"log"

	"github.com/machinebox/graphql"
)

type CompoundQuery struct {
	Markets []CompoundMarket `json:"markets"`
}

type CompoundMarket struct {
	BorrowRate              	string `json:"borrowRate"`
	ExchangeRate            	string `json:"exchangeRate"`
	SupplyRate               	string `json:"supplyRate"`
	ID                       	string `json:"id"`
	BlockTimestamp				int `json:"blockTimestamp"`
	UnderlyingName           	string `json:"underlyingName"`
	UnderlyingPrice          	string `json:"underlyingPrice"`
	UnderlyingSymbol         	string `json:"underlyingSymbol"`
	/*Cash                  	string `json:"cash"`
	Name                     	string `json:"name"`
	Symbol                   	string `json:"symbol"`
	CollateralFactor        	string `json:"collateralFactor"`
	Reserves                 	string `json:"reserves"`
	InterestRateModelAddress 	string `json:"interestRateModelAddress"`
	TotalBorrows             	string `json:"totalBorrows"`
	TotalSupply              	string `json:"totalSupply"`
	UnderlyingAddress        	string `json:"underlyingAddress"`
	ReserveFactor            string `json:"reserveFactor"`
	UnderlyingPriceUSD       string `json:"underlyingPriceUSD"`
}

func getCompoundData(database *Database, compoundreqdata CompoundInputStruct) {
	clientCompound := graphql.NewClient("https://api.thegraph.com/subgraphs/name/graphprotocol/compound-v2")

	reqCompoundRates := graphql.NewRequest(`
	query {

		markets(first: 12) {
			borrowRate
			supplyRate
			id
			blockTimestamp
			underlyingName
			underlyingPrice
			underlyingSymbol
			blockTimestamp
		}
	}`)

	reqCompoundRates.Var("key", "value")
	reqCompoundRates.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respCompoundMarkets CompoundQuery

	if err := clientCompound.Run(ctx, reqCompoundRates, &respCompoundMarkets); err != nil {
		log.Fatal(err)
	}

	var CompoundTokenName []string
	var CompoundTokenSymbol []string
	var CompoundTokenID []string
	var CompoundTokenExchangeRate []float64
	var CompoundTokenBorrowRate []float64
	var CompoundTokenSupplyRate []float64
	var CompoundTokenBlockTimestamp []int
	var CompoundTokenPrice []float64

	for i := 0; i < len(respCompoundMarkets.Markets); i++ {

		CompoundTokenName = append(CompoundTokenName, respCompoundMarkets.Markets[i].UnderlyingName)
		CompoundTokenSymbol = append(CompoundTokenSymbol, respCompoundMarkets.Markets[i].UnderlyingSymbol)
		CompoundTokenID = append(CompoundTokenID, respCompoundMarkets.Markets[i].ID)
		CompoundTokenExchangeRate = append(CompoundTokenExchangeRate, respCompoundMarkets.Markets[i].ExchangeRate)
		CompoundTokenBorrowRate = append(CompoundTokenBorrowRate, respCompoundMarkets.Markets[i].BorrowRate)
		CompoundTokenSupplyRate = append(CompoundTokenSupplyRate, respCompoundMarkets.Markets[i].SupplyRate)
		CompoundTokenBlockTimestamp = append(CompoundTokenBlockTimestamp, respCompoundMarkets.Markets[i].BlockTimestamp)
		CompoundTokenPrice = append(CompoundTokenPrice, respCompoundMarkets.Markets[i].CompoundTokenPrice)

		addToCompoundData(CompoundTokenSymbol[i], CompoundTokenName[i], CompoundTokenID[i],
			CompoundTokenSupplyRate[i], CompoundTokenBorrowRate[i],
			CompoundTokenExchangeRate[i], CompoundTokenPrice[i], CompoundTokenBlockTimestamp[i])

		var recordalreadyexists bool
		recordalreadyexists = false

		for j := 0; j < len(database.PoolTokenPairReturns); j++ {
			// Means record already exists - UPDATE IT, DO NOT APPEND
			if database.PoolTokenPairReturns[j].Pair == CompoundTokenSymbol[i]+"/c"+CompoundTokenSymbol[i]
				&& database.PoolTokenPairReturns[j].Pool == "Compound" {
				recordalreadyexists = true
				database.PoolTokenPairReturns[j].PoolSize = float64(future_pool_sz_est)
				database.PoolTokenPairReturns[j].PoolVolume = float64(future_daily_volume_est)

				database.PoolTokenPairReturns[j].ROI_raw_est = ROI_raw_est
				database.PoolTokenPairReturns[j].ROI_vol_adj_est = ROI_vol_adj_est
				database.PoolTokenPairReturns[j].ROI_hist = ROI_hist

				database.PoolTokenPairReturns[j].Volatility = volatility
				database.PoolTokenPairReturns[j].Yield = currentInterestrate
			}
		}

			// APPEND IF NEW
			if !recordalreadyexists {
				//				appendHistPriceDataToDb(PoolTokenPairReturns{token0symbol + "/" + token1symbol, float64(future_pool_sz_est),
				//					float64(future_daily_volume_est), currentInterestrate, "Uniswap", volatility, ROI_raw_est, ROI_vol_adj_est, ROI_hist})
				database.PoolTokenPairReturns = append(database.PoolTokenPairReturns, PoolTokenPairReturns{token0symbol + "/" + token1symbol, float64(future_pool_sz_est),
					float64(future_daily_volume_est), currentInterestrate, "Uniswap", volatility, ROI_raw_est, ROI_vol_adj_est, ROI_hist})
				//				database.PoolTokenPairReturns = append(database.PoolTokenPairReturns, PoolTokenPairReturns{token0symbol + "/" + token1symbol, float64(currentSize), float64(currentVolume), currentInterestrate, "Uniswap", volatility, ROI})
			}
		} // if pool is within pre filtered list ends
		// } // if pool has some tokens ends
	} // Uniswap pair loop closes
	}



}
*/

