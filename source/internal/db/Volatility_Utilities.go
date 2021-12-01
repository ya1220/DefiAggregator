package db

import (
	"math/big"
	"fmt"
	"math"
	"time"

	"strconv"
)

func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}


func sum(array []float64) float64 {
	result := 0.0
	for _, v := range array {
		result += v
	}
	return result
}

// Current Exchange Rate Comes from the protocol
// Standard deviation comes from the volatility estimate, can be a 30 days estimate
// returned estimate is -x% loss in liquidity
func estimate_impermanent_loss_hist(standard_deviation float64, current_exchange_rate float64, protocol string) float64 {

	impermanent_loss := float64(0)

	if protocol == "Uniswap" {

		forecasted_exchage_rate := current_exchange_rate + standard_deviation
		price_ratio := forecasted_exchage_rate / current_exchange_rate
		impermanent_loss = 2*math.Sqrt(float64(price_ratio))/(1+float64(price_ratio)) - 1

	}

	return float64(impermanent_loss)

}




func calculatehistoricalvolatility(H HistoricalCurrencyData, days int) float64 {
	fmt.Print("CALCULATING HISTORICAL VOLATILITY: ")
	fmt.Println(H.Ticker)

	// Ensure it is sorted from newest = 0 oldest = max int

	var vol float64
	vol = 0.05

	if len(H.Price) == 0 {
		fmt.Print("Error: no historical data found..returning 0: ")
		return 0
	}

	var vol_period int32

	// return -1 if no historical data available
	if len(H.Date) == 0 {
		vol = -1.00
	}

	vol_period = int32(math.Min(float64(len(H.Price)), float64(days))) // lower of days or available data

	// NOTE: oldest = 0
	var total float64
	total = 0.00

	fmt.Print("Vol period: ")
	fmt.Println(vol_period)

	var changes_in_price []float64
	var differencesvsmean []float64 // size = actual vol period
	var squaresofdifferencesvsmean []float64

	var actual_vol_period int32 // days with data
	actual_vol_period = 0

	if vol_period < 2 {
		return 0.0
	}

	for i := 1; i < int(vol_period); i++ {
		if !math.IsNaN(float64(H.Price[i])) && float64(H.Price[i]) > 0 && float64(H.Price[i-1]) > 0 {
			changes_in_price = append(changes_in_price, H.Price[i]/H.Price[i-1]-1)
			total = total + (H.Price[i]/H.Price[i-1] - 1) // calculate average price change
			actual_vol_period++
		}
	}

	//fmt.Print("vol calc - total deviation: ")
	//fmt.Println(total)

	mean := total / float64(actual_vol_period) // actual days?

	//fmt.Print("vol calc - mean: ")
	//fmt.Println(mean)

	for i := 1; i < int(vol_period); i++ {
		if !math.IsNaN(float64(H.Price[i])) && float64(H.Price[i]) > 0 && float64(H.Price[i-1]) > 0 {
			differencesvsmean = append(differencesvsmean, H.Price[i]/H.Price[i-1]-1-mean) // calculate difference between each value and mean
			squaresofdifferencesvsmean = append(squaresofdifferencesvsmean, float64(math.Pow(float64(H.Price[i]/H.Price[i-1]-1-mean), 2.0)))
			/*
				fmt.Print("Date: ")
				fmt.Print(H.Date[i])
				fmt.Print(" | ")
				fmt.Print("Price: ")
				fmt.Print(H.Price[i])
				fmt.Print(" | ")
				fmt.Print("Price - mean: ")
				fmt.Print(H.Price[i] - mean)
				fmt.Print(" | Sqr: ")
				fmt.Println(float64(math.Pow(float64(H.Price[i]/H.Price[i-1]-1-mean), 2.0)))
			*/
		}
	}

	var avg float64
	avg = 0.0

	for i := 0; i < len(squaresofdifferencesvsmean); i++ {
		avg += squaresofdifferencesvsmean[i]
	}

	//fmt.Print("Total squares: ")
	//fmt.Println(avg)

	avg = avg / float64(len(squaresofdifferencesvsmean))    // average them
	vol = float64(math.Sqrt(float64(avg)) * math.Sqrt(252)) // is this the right adjustment for days?

	fmt.Print("VOLATILITY = ")
	fmt.Println(vol)

	if math.IsInf(float64(vol), 0) {
		return -0.99
	}
	if math.IsNaN(float64(vol)) {
		return -0.98
	}

	return float64(vol)
}



// For looking up Uniswap IDs of tokens
func setUniswapQueryIDForToken(token string, ID string) string {
	if token == "DAI" {
		return "0x6b175474e89094c44da98b954eedeac495271d0f"
	}
	if token == "USDC" {
		// TO CHECK
		return "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" // "check the uniswap id for dai"
	}
	if token == "ETH" || token == "WETH" {
		// TO CHECK
		return "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	}
	if token == "BAL" {
		// TO CHECK
		return "0xba100000625a3754423978a60c9317c58a424e3d"
	}

	return ID
}

// Convert Balancer token to a Uniswap token format
func convBalancerToken(t string) string {
	if t == "ETH" {
		return "WETH"
	}
	return t
}

// Checks if string is already in a vector
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func int64InSlice(a int64, list []int64) (bool,int) {
	for i, b := range list {
		if b == a {
			return true, i
		}
	}
	return false,0
}


// Checks if a pair is within a pre-set list
func isPoolPartOfFilter(token0 string, token1 string) bool {
	if isCoinPartOfFilter(token0) && isCoinPartOfFilter(token1) {return true}
	return false
}

func isCoinPartOfFilter(token0 string) bool {

	var tokens []string

	tokens = append(tokens,"DAI")
	tokens = append(tokens,"USDC")
	tokens = append(tokens,"USDT")
	tokens = append(tokens,"WETH")
	tokens = append(tokens,"WBTC")
	tokens = append(tokens,"DOGE")
	tokens = append(tokens,"BAL")
	//tokens = append(tokens,"SNX")
	tokens = append(tokens,"UNI")
	tokens = append(tokens,"RLY")
	tokens = append(tokens,"LINK")

	for i := 0; i < len(tokens);i++ {
		if token0 == tokens[i] {
			return true
		}
	}

	return false
}

func isTokenStableCoin(coinName string) bool {
	if coinName == "USDT" {
		return true
	} else if coinName == "USDC" {
		return true
	} else if coinName == "USD" {
		return true
	} else if coinName == "TUSD" {
		return true
	} else if coinName == "DAI" {
		return true
	} else if coinName == "GUSD" {
		return true
	} else if coinName == "BUSD" {
		return true
	} else {
		return false
	}
}

func BoD(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func BoDi64(t int64) int64 {
	if t == 0 {return 0}
	return int64((t-int64(math.Mod(float64(t), 86400))))
}

func BoDui64(t uint64) int64 {
	if t == 0 {return 0}
	return int64((t-uint64(math.Mod(float64(t), 86400))))
}


func Mul(a, b *big.Float) *big.Float {
	return Zero().Mul(a, b)
}

func negPowF(a float64, e int64) float64 {
	result := a
	for i := int64(0); i < e; i++ {
		result = result / float64(10)
	}
	return result
}

func negPow(a *big.Float, e int64) *big.Float {
	result := Zero().Copy(a)
	divTen := big.NewFloat(0.1)
	for i := int64(0); i < e; i++ {
		result = Mul(result, divTen)
	}
	return result
}

func negPowI(a *big.Int, e int64) *big.Float {
	result := new(big.Float).SetInt(a)
	divTen := big.NewFloat(0.1)
	for i := int64(0); i < e; i++ {
		result = Mul(result, divTen)
	}
	return result
}

func Zero() *big.Float {
	r := big.NewFloat(0.0)
	r.SetPrec(256)
	return r
}


func Div(a, b *big.Float) *big.Float {
	return Zero().Quo(a, b)
}



// ROI Ranking Function
/*
func (database *Database) RankBestCurrencies() {
}
*/

func MaxIntSliceD(v []int64) (m int64) {
	if len(v) > 0 {
		m = v[0]
	} else {
		return 0
	}
	for i := 1; i < len(v); i++ {
		if v[i] > m {
			m = v[i]
		}
	}
	return
}

func MaxIntSlice(v []int64) (m int64) {
	if len(v) > 0 {
		m = v[0]
	}
	for i := 1; i < len(v); i++ {
		if v[i] > m {
			m = v[i]
		}
	}
	return
}

func MaxArgSlice(v []int64) (m int64, j int) {
	
	if len(v) > 0 {
		m = v[0]
	} else {
		return 0,-1
	}
	for i := 1; i < len(v); i++ {
		if v[i] > m {
			m = v[i]
			j = i
		}
	}
	return m,j
}


func MinIntSlice(v []int64) (m int64) {
	if len(v) > 0 {
		m = v[0]
	}
	for i := 1; i < len(v); i++ {
		if v[i] < m {
			m = v[i]
		}
	}
	return
}
