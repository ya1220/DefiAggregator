package db // notifier_new

import (
//	"pusher/defi_aggregator/source/internal/db"
	"github.com/pusher/pusher-http-go"
//	"fmt"
)

type Notifier_new struct {
	notifyChannel_raw_and_optimised_pf chan<- bool
	notifyChannel_pooltable chan<- bool
}

func notifier_new(database *Database, notifyChannel_raw_and_optimised_pf <-chan bool, notifyChannel_pooltable <-chan bool) {
	client := pusher.Client{
		AppID:   "1139323",
		Key:     "7885860875bb513c3e34",
		Secret:  "3633fcf50bba02630b0c",
		Cluster: "eu",
		Secure:  true,
	}

	f1 := func(){
		for {
			//fmt.Print("..007..")
			<-notifyChannel_pooltable
			//fmt.Print("..008..")
			ranked_pools_table_data := map[string][]PoolTokenPairReturns{"ranked_pools_table": database.GetRankedPoolsTable()}
			client.Trigger("ranked_pools_table", "ranked_pools_table", ranked_pools_table_data) // data table
			//fmt.Print("..009..")
		}	
	}

	go f1()
	
	for {
		<-notifyChannel_raw_and_optimised_pf
		raw_portfolio_data := map[string][]RawPortfolioRecord{"raw_portfolio": database.GetRawPortfolio()}
		optimised_portfolio_data := map[string][]OptimisedPortfolioRecord{"optimised_portfolio": database.GetOptimisedPortfolio()}

		client.Trigger("raw_portfolio", "raw_portfolio", raw_portfolio_data) // raw
		client.Trigger("optimised_portfolio", "optimised_portfolio", optimised_portfolio_data) // optimised
	}

}

func New_Notifier_new(database *Database) Notifier_new {
	notifyChannel_raw_and_optimised_pf := make(chan bool)
	notifyChannel_pooltable := make(chan bool)

	go notifier_new(database, notifyChannel_raw_and_optimised_pf, notifyChannel_pooltable)

	return Notifier_new{
		notifyChannel_raw_and_optimised_pf,
		notifyChannel_pooltable,
	}
}

func (notifier_new *Notifier_new) Notify_raw_and_optimised_pf() {
	//fmt.Print("0.10..")
	notifier_new.notifyChannel_raw_and_optimised_pf <- true
	//fmt.Print("reset raw + optimised pf chan to true")
}


func (notifier_new *Notifier_new) Notify_pooltable() {
	//fmt.Print("0.11..")
	notifier_new.notifyChannel_pooltable <- true
	//fmt.Print("reset pool table chan to true")
}
