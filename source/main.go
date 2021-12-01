package main

import (
	"fmt"
	"pusher/defi_aggregator/source/internal/db"
//	"pusher/defi_aggregator/source/internal/notifier"
//	"pusher/defi_aggregator/source/internal/notifier_new"
	"pusher/defi_aggregator/source/internal/webapp"
)

func main() {
	database := db.New()
	//notifierClient := notifier.New(&database)
	//notifierClient := notifier_new.New(&database)
	notifierClient := db.New_Notifier_new(&database)
	// alternatively create 2 clients
	fmt.Print("About to start server in main..")

	webapp.StartServer(&database, &notifierClient)
}
