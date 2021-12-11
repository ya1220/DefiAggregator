package main

import (
	"DefiAggregator/db"
	"DefiAggregator/webapp"
	"fmt"
)

func main() {
	database := db.New()
	//notifierClient := notifier.New(&database)
	//notifierClient := notifier_new.New(&database)
	notifierClient := db.New_Notifier_new(&database)
	fmt.Print("About to start server in main..")
	webapp.StartServer(&database, &notifierClient)
}
