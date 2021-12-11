package db


import (
	"fmt"
	"time"
)

func (database *Database) do_records_need_updating(poolname string) bool {
	// check if need to run at all
	// if data last updated less than 1 minute ago, do not recalculate
	recordalreadyexists := false
	latest_update_time := int64(0)

	for k := 0; k < len(database.PoolTokenPairReturns); k++ {
		if database.PoolTokenPairReturns[k].Pool == poolname {
			recordalreadyexists = true // data not completely blank
			if database.PoolTokenPairReturns[k].Last_updated > latest_update_time {
				latest_update_time = database.PoolTokenPairReturns[k].Last_updated
			} // update if newer uniswap data item exists
		}
	}

	fmt.Print("Time of latest update for Balancer: ")
	fmt.Print(latest_update_time)
	fmt.Print(" diff vs now: ")
	fmt.Println(time.Now().Unix() - latest_update_time)

	fmt.Printf("SoD local: %v", BoDi64(time.Now().Unix()))
	fmt.Printf("SoD UTC: %v", BoDi64(time.Now().UTC().Unix()))
	// Do nothing - data is less than x minutes old
	if recordalreadyexists && (time.Now().Unix() - latest_update_time) < database.update_frequency {
		fmt.Print("Data recent - no update!")
		return false
	}

		return true
}
// uniswapreqdata UniswapInputStruct