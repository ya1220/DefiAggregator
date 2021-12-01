package db

import (
	"fmt"
	"math"
)

// pool_sizes_usd
func nrm_pool_wgts(pool_weights_raw []float64, available_pool_tkn0s []string, available_pool_tkn1s []string, pool_ratios []float64, own_pf []RawPortfolioRecord, conversion_to_usd_px_arr []float64) ([]float64,[]string,[]float64) {
	//fmt.Printf("\n500 - NORMALIZING: Pool weights raw: %v\n", pool_weights_raw)

	var weights_norm []float64
	var leftovertokens []string
	var leftoveramounts []float64

	// 0 - ELIMINATE < 0 and > 1 values
	for j := 0; j < len(pool_weights_raw); j++ {
		if pool_weights_raw[j] < 0 {
			weights_norm = append(weights_norm, 0.0)
		} else if pool_weights_raw[j] > 1 {
			weights_norm = append(weights_norm, 1.00)
		} else {
			weights_norm = append(weights_norm, pool_weights_raw[j])
		}
	}

	// 1 - CALCULATE USD PF VALUE
	total_pf_val_usd := 0.0 
	for i := 0; i < len(own_pf); i++ {
		total_pf_val_usd += conversion_to_usd_px_arr[i] * own_pf[i].Amount
		fmt.Printf(".adding to TOTAL: %v\n", conversion_to_usd_px_arr[i] * own_pf[i].Amount)
	}

	// 2 - ENSURE WEIGHTS SUM TO 100%
	tot_wgt := sum(weights_norm)
	if tot_wgt > 1.00 {
		for j := 0; j < len(weights_norm); j++ {
			weights_norm[j] = weights_norm[j] / tot_wgt
		}
	} 

	//tot_test := sum(weights_norm)
	fmt.Print("NORMALIZED WEIGHTS: ")
	for i:=0; i < len(weights_norm); i++ {
		fmt.Print("i: ")
		fmt.Print(i)
		fmt.Print(" | ")
		fmt.Println(weights_norm[i])
	}
	fmt.Printf("new total: %v\n", sum(weights_norm))

	token_balances_resulting_from_entered_weights := token_balances_from_weights_usd(weights_norm, total_pf_val_usd, available_pool_tkn0s, available_pool_tkn1s, pool_ratios, own_pf)

	//fmt.Printf("..IN NORM..token balances IN USD VALs: %v\n", token_balances_resulting_from_entered_weights)

	violation_count := 0
	for i := 0; i < len(own_pf); i++ {
		if token_balances_resulting_from_entered_weights[i] > own_pf[i].Amount * conversion_to_usd_px_arr[i] {
			fmt.Printf("VIOLATING token: %v", own_pf[i].Token)
			fmt.Printf(" amt: %v", own_pf[i].Amount)
			fmt.Printf(" conv mult: %v", conversion_to_usd_px_arr[i])
			fmt.Printf(" vs calced: %v\n", token_balances_resulting_from_entered_weights[i])

			violation_count++
		}
	}

	first_pass := true

	// Tokens which appear in weights BUT DO NOT APPEAR IN OWN PF

	fmt.Printf("violation count: %v..entering while loop to eliminate them ..\n", violation_count)

	for (violation_count > 0 || first_pass) { // make sure they sum to NO GREATER THAN individual token balances
		for j := 0; j < len(weights_norm); j++ {
			fmt.Printf("available_pool_tkn0s[j]: %v", available_pool_tkn0s[j])
			fmt.Printf(" | available_pool_tkn1s[j]: %v\n", available_pool_tkn1s[j])

			idx0 := -1 // find idx of available_pool_tkn0s[j] in own pf
			idx1 := -1 // find idx of available_pool_tkn1s[j] in own pf

			for jj := 0; jj < len(own_pf); jj++ {
				if available_pool_tkn0s[j] == own_pf[jj].Token {
					idx0 = jj
					break
				}
			} // find idx0

			for jj := 0; jj < len(own_pf); jj++ {
				if available_pool_tkn1s[j] == own_pf[jj].Token {
					idx1 = jj
					break
				}
			} // find idx1

			fmt.Printf("idx0: %v ", idx0)
			fmt.Printf("idx1: %v ", idx1)

			// if token completely not found - ZERO this pool
			if len(available_pool_tkn0s[j]) > 0 && idx0 == -1 {
				fmt.Printf("..TOKEN at idx0 not found in own pf - ZEROING POOL WEIGHT..\n\n")
				weights_norm[j] = 0
				first_pass = false
				continue
			}

			if len(available_pool_tkn1s[j]) > 0 && idx1 == -1 {
				fmt.Printf("..TOKEN at idx1 not found in own pf - ZEROING POOL WEIGHT..\n\n")
				weights_norm[j] = 0
				first_pass = false
				continue
			}

			own_pf_token0_amt_usd := float64(0.0)
			own_pf_token1_amt_usd := float64(0.0)

			if len(available_pool_tkn0s[j]) > 0 {
				own_pf_token0_amt_usd = own_pf[idx0].Amount * conversion_to_usd_px_arr[idx0]
			}

			if len(available_pool_tkn1s[j]) > 0 {
				own_pf_token1_amt_usd = own_pf[idx1].Amount * conversion_to_usd_px_arr[idx1]	
			}

			rat0 := 0.0 // downscaling ratio to fit within available amounts
			rat1 := 0.0 // downscaling ratio to fit within available amounts

			if own_pf_token0_amt_usd > 0 {
				fmt.Printf("token_balances_resulting_from_entered_weights[idx0]: %v\n", token_balances_resulting_from_entered_weights[idx0])
				fmt.Printf("own_pf_token0_amt_usd: %v\n", own_pf_token0_amt_usd)

				if token_balances_resulting_from_entered_weights[idx0] > own_pf_token0_amt_usd  { // > 0.01
					rat0 = token_balances_resulting_from_entered_weights[idx0] / own_pf_token0_amt_usd
					fmt.Printf("\nRat0: %v\n", rat0)
				}
			}
			if own_pf_token1_amt_usd > 0 {
				if token_balances_resulting_from_entered_weights[idx1] > own_pf_token1_amt_usd { 
					rat1 = token_balances_resulting_from_entered_weights[idx1] / own_pf_token1_amt_usd
					fmt.Printf("\nRat1: %v\n", rat1)
				}
			}

			rat := float64(0.0)

			if rat0 > 0.0 && rat1 > 0.0 {
				rat = math.Min(rat0, rat1) // resultant divided by available
				fmt.Printf("\nSCALING DOWN POOL BY: %v\n", 1/rat)
				if rat > 1 { // scale pool % by rat
					fmt.Printf("..Ratio > 1 at pool idx: %v\n", j)
					weights_norm[j] = weights_norm[j] / rat
				} // if rat	

			}
			if len(available_pool_tkn1s[j]) == 0 {
				rat = rat0
		
				if rat > 1 { // scale pool % by rat
					fmt.Printf("\nSCALING DOWN LENDING POOL BY: %v\n", 1/rat)
					fmt.Printf("..Ratio > 1 at pool idx: %v", j)
					weights_norm[j] = weights_norm[j] / rat
					fmt.Printf("..new norm: %v\n\n", weights_norm[j])
				} // if rat
			
			}
		} // for len pool weights raw

		token_balances_resulting_from_entered_weights = token_balances_from_weights_usd(weights_norm, total_pf_val_usd, available_pool_tkn0s, available_pool_tkn1s, pool_ratios, own_pf)
		violation_count = 0

		for i := 0; i < len(own_pf); i++ {
			if token_balances_resulting_from_entered_weights[i] > float64(own_pf[i].Amount)*conversion_to_usd_px_arr[i] {
				violation_count++
			}
		}

		first_pass = false
	} // violation count loop ends

	for ii := 0; ii < len(own_pf); ii++ {
		fmt.Printf("token: %v", own_pf[ii].Token)
		fmt.Printf(" amt: %v", own_pf[ii].Amount)
		fmt.Printf(" conv px: %v", conversion_to_usd_px_arr[ii])
		fmt.Printf(" bal res: %v\n", token_balances_resulting_from_entered_weights[ii])
	}

	// Update return matrix with actual token returns
	for ii := 0; ii < len(own_pf); ii++ {
		diff := own_pf[ii].Amount * conversion_to_usd_px_arr[ii] - token_balances_resulting_from_entered_weights[ii]
		if diff < 0 {
			fmt.Print("ERROR: negative resultant balance")
			//log.Fatal()
		}
		fmt.Printf("ADDING LEFTOVER token: %v ",own_pf[ii].Token)
		fmt.Printf(" | AMT: %v \n",diff)

		leftovertokens = append(leftovertokens,own_pf[ii].Token)
		leftoveramounts = append(leftoveramounts,diff)
	}

	return weights_norm, leftovertokens,leftoveramounts
}


func token_balances_from_weights_usd(weights []float64, total float64, available_pool_tkn0s []string, available_pool_tkn1s []string, pool_ratios []float64, own_pf []RawPortfolioRecord) []float64 {

		fmt.Print("\nIN RECALC BALANCES: ")
		fmt.Println(len(weights))
		fmt.Println(len(available_pool_tkn0s))
		fmt.Println(len(available_pool_tkn1s))
		fmt.Println(len(pool_ratios))
		fmt.Println(len(own_pf))

		fmt.Printf("total USD VAL: %v\n", total)

	token_balances_resulting_from_entered_weights := make([]float64, len(own_pf), len(own_pf))

	for j := 0; j < len(weights); j++ {
		idx0 := -1 // find idx of available_pool_tkn0s[j] in own pf
		idx1 := -1 // find idx of available_pool_tkn1s[j] in own pf

		for jj := 0; jj < len(own_pf); jj++ {
			if available_pool_tkn0s[j] == own_pf[jj].Token {
				idx0 = jj
				fmt.Printf("X tkn0: %v", own_pf[jj].Token)
				fmt.Printf(" | idx0: %v\n", idx0)
				break
			}
		} // find idx0

		for jj := 0; jj < len(own_pf); jj++ {
			if available_pool_tkn1s[j] == own_pf[jj].Token {
				idx1 = jj
				fmt.Printf("Y tkn0: %v", own_pf[jj].Token)
				fmt.Printf(" | idx0: %v\n", idx1)
				break
			}
		} // find idx1

		fmt.Printf(" | j: %v ", j)
		fmt.Printf("weights j: %v ", weights[j])
		fmt.Printf(" | pool ratios j: %v ", pool_ratios[j])
		fmt.Printf(" | total: %v \n", total)

		token_balances_resulting_from_entered_weights[idx0] += (weights[j] * total * (pool_ratios[j]))
		if available_pool_tkn1s[j] != "" {
			token_balances_resulting_from_entered_weights[idx1] += (weights[j] * total * (1-pool_ratios[j]))
		}

	} // translate pool weights to total token balances

	fmt.Printf("\n\n GDX Resultant token balances: \n")
		for ii := 0; ii < len(token_balances_resulting_from_entered_weights);ii++{		
			fmt.Print("ii: ")
			fmt.Print(ii)
			fmt.Print(" | tkn: ")
			fmt.Print(own_pf[ii].Token)
			fmt.Print(" | ")
			fmt.Print(own_pf[ii].Amount)
			fmt.Print(" | tkn bal res: ")
			fmt.Println(token_balances_resulting_from_entered_weights[ii])	
		}

	return token_balances_resulting_from_entered_weights
}









		//fmt.Print(" |t oken balances resulting: ")
		//fmt.Println(token_balances_resulting_from_entered_weights[idx0])
		//fmt.Println(token_balances_resulting_from_entered_weights[idx1])







		/*
			fmt.Print("Checkpoint 994")
			fmt.Print("ii: ")
			fmt.Print(ii)
			fmt.Print(" | tkn: ")
			fmt.Print(own_pf[ii].Token)
			fmt.Print(" | ")
			fmt.Print(own_pf[ii].Amount)
			fmt.Print(" | tkn bal res: ")
			fmt.Print(token_balances_resulting_from_entered_weights[ii])
		*/


		/*
	// if nothing to deploy into - return whole pf as to NA 
	if len(available_pool_tkn0s) == 0 && len(available_pool_tkn1s) == 0 {
		for jjj := 0; jjj < len(own_pf);jjj++ {
			weights_optimised = append(weights_optimised,0.0)		
		}

		total := 0.0
		for i := 0; i < len(own_pf); i++ {
			total += conversion_to_usd_px_arr[i] * float64(own_pf[i].Amount)
		}
	
		token_balances_resulting_from_entered_weights := token_balances_from_weights_usd(weights_optimised, total, available_pool_tkn0s, available_pool_tkn1s, pool_ratios, own_pf)

		for ii := 0; ii < len(own_pf); ii++ {
			diff := float64(own_pf[ii].Amount)*conversion_to_usd_px_arr[ii] - token_balances_resulting_from_entered_weights[ii]
			leftovertokens = append(leftovertokens,own_pf[ii].Token)
			leftoveramounts = append(leftoveramounts,diff)
		}
		return weights_optimised, leftovertokens,leftoveramounts
	}
*/

	/*

	fmt.Println(len(pool_weights_raw))
	fmt.Println(len(available_pool_tkn0s))
	fmt.Println(len(available_pool_tkn1s))
	fmt.Printf("ratios: %v\n", len(pool_ratios))
	fmt.Printf("own pf: %v\n", len(own_pf))
	fmt.Printf("available px: %v\n",len(conversion_to_usd_px_arr))
	
	fmt.Print("RAW WEIGHTS: ")
	for i:=0; i < len(pool_weights_raw); i++ {
		fmt.Print("i: ")
		fmt.Print(i)
		fmt.Print(" | ")
		fmt.Print(pool_weights_raw[i])
		fmt.Print(" | ")
		fmt.Print(available_pool_tkn0s[i])
		fmt.Print(" | ")
		fmt.Println(available_pool_tkn1s[i])
	}
*/


	/*
	for i:=0; i < len(own_pf); i++ {
		sum up across all pools 
		max_pool_sz_for_token[i] = % weight of that token in pool 
	}
	*/


