package notifier

import (
	"pusher/defi_aggregator/source/internal/db"
	"github.com/pusher/pusher-http-go"
)

type Notifier struct {
	notifyChannel chan<- bool
}

func notifier(database *db.Database, notifyChannel <-chan bool) {
	client := pusher.Client{
		AppID:   "1139323",
		Key:     "7885860875bb513c3e34",
		Secret:  "3633fcf50bba02630b0c",
		Cluster: "eu",
		Secure:  true,
	}

	// infinite loop for both results and ranked_pools_table
	for {
		<-notifyChannel
		raw_portfolio_data := map[string][]db.RawPortfolioRecord{"raw_portfolio": database.GetRawPortfolio()}
		optimised_portfolio_data := map[string][]db.OptimisedPortfolioRecord{"optimised_portfolio": database.GetOptimisedPortfolio()}
		//ranked_pools_table_data := map[string][]db.PoolTokenPairReturns{"ranked_pools_table": database.GetRankedPoolsTable()}

		client.Trigger("raw_portfolio", "raw_portfolio", raw_portfolio_data) // raw
		client.Trigger("optimised_portfolio", "optimised_portfolio", optimised_portfolio_data) // optimised
		//client.Trigger("ranked_pools_table", "ranked_pools_table", ranked_pools_table_data) // data table
	}

}

func New(database *db.Database) Notifier {
	notifyChannel := make(chan bool)
	go notifier(database, notifyChannel)
	return Notifier{
		notifyChannel,
	}
}

func (notifier *Notifier) Notify() {
	notifier.notifyChannel <- true
}
