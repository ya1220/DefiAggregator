package db

import (
	"fmt"
	"math"
)



func calculate_price_return_x_days(hist_date_px_series HistoricalCurrencyData, days int64) float64 {

	max_date := int64(0)
	latest_px := float64(0)

	fmt.Print("--IN CALC HIST RETURN PX: ---")
	fmt.Print(days)
	fmt.Printf("act len: %v\n", len(hist_date_px_series.Date))

		j := 0
		max_date, j = MaxArgSlice(hist_date_px_series.Date)
		latest_px = hist_date_px_series.Price[j] 

	t_ago := max_date - 60 * 60 * 24 * days
	smallest_diff := float64(0) // 60 * 60 * 24 * days

	// if t_ago exists - get it

	for i := 0; i < len(hist_date_px_series.Date); i++ {
		if hist_date_px_series.Date[i] == t_ago {
			fmt.Print("found exact date!!")
			fmt.Printf("returning: %v\n", latest_px / hist_date_px_series.Price[i])
			if hist_date_px_series.Price[i] != 0.0 {
				return latest_px / hist_date_px_series.Price[i]
			} else {
				fmt.Print("ERROR: 0 price in hist")
				return -1
			}
		}
	}

	fmt.Print("exact date not found")
	fmt.Printf("t_ago: %v\n",t_ago)

	closest_i := 0

	for i := 0; i < len(hist_date_px_series.Date); i++ {
	//		fmt.Printf("i: %v",i)
		if i == 0 {
		  smallest_diff = math.Abs(float64(hist_date_px_series.Date[i] - t_ago))
		} else {
			if math.Abs(float64(hist_date_px_series.Date[i] - t_ago)) < float64(smallest_diff) {
			//	fmt.Printf("reducing lkp dist at i: %v\n", i)
			//	fmt.Printf(" | diff abs: %v\n",math.Abs(float64(hist_date_px_series.Date[i] - t_ago)))
				closest_i = i
				smallest_diff = math.Abs(float64(hist_date_px_series.Date[i] - t_ago))

				if smallest_diff == 0 { // if found exact date
					break
				}
			} // if reducing diff
		} // else
	} // i

	fmt.Printf("closest_to_days_ago_date: %v",hist_date_px_series.Date[closest_i])
	fmt.Printf(" | closest_to_days_ago_px: %v\n",hist_date_px_series.Price[closest_i])

	if hist_date_px_series.Date[closest_i] > 0 {
		return latest_px / hist_date_px_series.Price[closest_i]
	}

	fmt.Print("ERROR 654! Could not find a historical starting price")
	return 0.0
}



func calculateROI_hist(interestrate float64, pool_reward_pct float64, pool_sz_hist float64, daily_volume_hist float64, imp_loss_hist_est float64, px_return_hist float64) float64 {
	var ROI float64
	ROI = 0.0676

	if pool_sz_hist == 0.0 {
		return 0.0 // -992
	} else {
		ROI = interestrate + (pool_reward_pct * daily_volume_hist * 365 / pool_sz_hist) + px_return_hist + imp_loss_hist_est
	}

	if math.IsInf(float64(ROI), 0) {
		return -999
	}

	if math.IsNaN(float64(ROI)) {
		return -998
	}

	return float64(ROI)
}

// Sharpe ratio
func calculateROI_vol_adj(ROI_raw_est float64, volatility_est float64) float64 {

	if math.IsInf(float64(ROI_raw_est), 0) {
		return 888.0
	}

	if volatility_est <= 0.01 { // if not volatile - do not adjust by volatility
		return ROI_raw_est
	} else {
		if !math.IsInf(float64(ROI_raw_est/volatility_est), 0) {
			return ROI_raw_est / volatility_est // sharpe ratio - risk free rate assumed to be zero
		} else {
			return 888.8
		}

	}

}

func calculateROI_raw_est(interestrate float64, pool_reward_pct float64, future_pool_sz_est float64, future_daily_volume_est float64, imp_loss_hist_est float64) float64 {

	var ROI float64
	ROI = 0.069

	if future_pool_sz_est > 0.0 {
		ROI = interestrate + (pool_reward_pct * future_daily_volume_est * 365 / future_pool_sz_est) + imp_loss_hist_est
	}

	if math.IsInf(float64(ROI), 0) {
		return -999
	}
	if math.IsNaN(float64(ROI)) {
		return -998
	}

	return float64(ROI)
}

