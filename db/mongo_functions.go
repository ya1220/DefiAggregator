package db

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"math"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database
func appendHistPriceDataToDb(Histrecord HistoricalCurrencyData) {
	// append non-overlapping dates
	for i := 0; i < len(Histrecord.Date); i++ {
		// if date not already in collec
		id := addHistoricalCurrencyData(Histrecord.Date[i], Histrecord.Price[i], Histrecord.Ticker)
		if len(id) == 0 {fmt.Println(id)}
	}
}


func appendHistVolumeDataToDb(pool string, tokens []string, poolid string, date int64, trading_volume_usd float64, pool_sz_usd float64, fees float64, weighted_av_ir float64, util_rate float64, tkn0_pct float64,tkn1_pct float64) string {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("De-Fi_Aggregator")

	v := generate_pool_collection_name_in_db(pool,tokens,poolid)

//	fmt.Printf("v: %v\n", v)
//	fmt.Printf("generated through func: %v\n", generate_pool_collection_name_in_db(pool,tokens,poolid))
//	fmt.Printf("ARE THEY EQUAL: %v\n", v == generate_pool_collection_name_in_db(pool,tokens,poolid))

	optimisedportfolio := Database.Collection(v)

	// Check if it does not already exist
	// check if date exists in that collection - if yes return "already exists"
	cursor, err := optimisedportfolio.Find(ctx, bson.M{"Date": date})
	if err != nil {
		log.Fatal(err)
	}

	var collectionFiltered []bson.M
	err = cursor.All(ctx, &collectionFiltered)
	if err != nil {
		log.Fatal(err)
	}

	if len(collectionFiltered) > 0 {
		fmt.Print("..appending new hist VLM data to: ")
		fmt.Print(v)
		fmt.Print("..date already exists in collection..")
		return " date already in database..not added anything"
	}

	tokens_modified := tokens

	fmt.Print("CHECKPOINT 23421452")

	for i:=0; i < 8; i++ {
		if len(tokens_modified) < i + 1 {
			tokens_modified = append(tokens_modified,"token" + strconv.Itoa(i))
		} 
	}

	for i:=0; i < 8; i++ {
		fmt.Println(tokens_modified[i])
	}

	//log.Fatal(err)

	new_portfolio, err := optimisedportfolio.InsertOne(ctx, bson.D{
		{Key: "pool", Value: pool},
		{Key: "token0", Value: tokens_modified[0]}, // token0
		{Key: "token1", Value: tokens_modified[1]}, // token1 
		{Key: "token2", Value: tokens_modified[2]}, // token2
		{Key: "token3", Value: tokens_modified[3]}, // token3
		{Key: "token4", Value: tokens_modified[4]}, // token4
		{Key: "token5", Value: tokens_modified[5]}, // token5
		{Key: "token6", Value: tokens_modified[6]}, // token6
		{Key: "token7", Value: tokens_modified[7]}, // token7
		{Key: "date", Value: date},
		{Key: "trading_Volume_USD", Value: trading_volume_usd},
		{Key: "pool_sz_USD", Value: pool_sz_usd},
		{Key: "fees", Value: fees},
		{Key: "weighted_av_IR", Value: weighted_av_ir},
		{Key: "token0_pct", Value: tkn0_pct},
		{Key: "token1_pct", Value: tkn1_pct},
	})

	if err != nil {
		log.Fatal(err)
	}

	newID := new_portfolio.InsertedID
	hexID := newID.(primitive.ObjectID).Hex()

	return hexID
}

func get_latest_token_price(token string) float64 {
//	if isTokenStableCoin(token) {
//		return 1.0
//	}
	//data_exists,_,_ := isHistDataAlreadyDownloadedDatabase(token)
//	if data_exists {
		//date, _ := MaxArgSlice(returnDatesInCollection)
		return convert_amt_from_to_using_latest_exch_rate(1, token,"USD")
//	}
//	return 1.0
}



func convert_amt_from_to_using_latest_exch_rate(token_from_amt float64, token_from string, token_to string) float64 {

	return convert_amt_from_to(token_from_amt,token_from,token_to,get_newest_timestamp_for_token_from_db(token_from))

}

func convert_amt_from_to(token_from_amt float64, token_from string, token_to string, date int64) float64 {
	// if both stablecoins --> return token_from_Amt
	if isTokenStableCoin(token_from) && isTokenStableCoin(token_to) {
		return token_from_amt
	}

	px_from := float64(1.0)
	px_to := float64(1.0)
	newest_act_dt := int64(0)

	if token_from == "ETH" {token_from = "WETH"}
	if token_from == "BTC" {token_from = "WBTC"}

	if token_to == "ETH" {token_to = "WETH"}
	if token_to == "BTC" {token_to = "WBTC"}

	// otherwise query db
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("test2")
	
	token_from_data := Database.Collection(token_from)
	token_to_data := Database.Collection(token_to)

	var token_data_filtered_from []bson.M
	var token_data_filtered_to []bson.M


	if !isTokenStableCoin(token_from){
		filterCursor, err := token_from_data.Find(ctx, bson.M{"Date": date})
		if err != nil {
			//log.Fatal(err)
			fmt.Print("..403 - Date not found in db!..")
		}
	
		if err = filterCursor.All(ctx, &token_data_filtered_from); err != nil {
			log.Fatal(err)
		}

	if len(token_data_filtered_from) == 0 {
		// fmt.Print("IN LEN FROM = 0")
		px_from = 0.0

			//////////
			newest_act_dt = get_newest_timestamp_for_token_from_db(token_from)
			filterCursor, err = token_from_data.Find(ctx, bson.M{"Date": newest_act_dt})
			if err = filterCursor.All(ctx, &token_data_filtered_from); err != nil {
				log.Fatal(err)
			}

			// fmt.Printf("NEW LEN: %v\n", len(token_data_filtered_from))

			for _, token_data_filtered_f_r := range token_data_filtered_from {
				date_f := token_data_filtered_f_r["Date"]
				if date_f == newest_act_dt {
					px_from_i := token_data_filtered_f_r["Price"]
					px_from = px_from_i.(float64)
				} 
			}
			/////////
	} else {
		for _, token_data_filtered_f_r := range token_data_filtered_from {
				date_f := token_data_filtered_f_r["Date"]
				if date_f == date {
					px_from_i := token_data_filtered_f_r["Price"]
					px_from = px_from_i.(float64)
			} 
		}	
	}
	} // token_from


if !isTokenStableCoin(token_to){
	filterCursor_to, err := token_to_data.Find(ctx, bson.M{"Date": date})
	if err != nil {
		log.Fatal(err)
	}

	if err = filterCursor_to.All(ctx, &token_data_filtered_to); err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("len of data pulled: %v\n", len(token_data_filtered_to))

	if len(token_data_filtered_to) == 0 {
	//	fmt.Print("..in TOKEN DATA LEN 0..")
		px_to = 0.0

		newest_act_dt = get_newest_timestamp_for_token_from_db(token_to)
		// fmt.Printf("newest_act_dt: %v", newest_act_dt)

		filterCursor_to, err = token_to_data.Find(ctx, bson.M{"Date": newest_act_dt})
		if err != nil {
			log.Fatal(err)
		}

		if err = filterCursor_to.All(ctx, &token_data_filtered_to); err != nil {
			log.Fatal(err)
		}

		//fmt.Printf("len of data pulled ON TRY 2: %v\n", len(token_data_filtered_to))		

		for _, token_data_filtered_t_r := range token_data_filtered_to {
			date_t := token_data_filtered_t_r["Date"]
			if date_t == newest_act_dt {
				px_to_i := token_data_filtered_t_r["Price"]
				px_to = px_to_i.(float64)
			} 
		}
		//fmt.Print("After checking for latest act available: %v\n", px_to)

	} else {
		for _, token_data_filtered_t_r := range token_data_filtered_to {
			date_t := token_data_filtered_t_r["Date"]
			if date_t == date {
				px_to_i := token_data_filtered_t_r["Price"]
				px_to = px_to_i.(float64)
			} 
		}
	// if exact date not found - use latest available
		} // else 
} // token_from


	if px_to != 0.0 && px_from != 0.0 {
		//fmt.Printf("converted amt: %v\n\n", float64(token_from_amt * px_from / px_to))
		return token_from_amt * (px_from / px_to)
	} else {
		fmt.Printf("..404 - ERROR: price not found DB..")

		fmt.Printf("\nt from: %v | ", token_from)
		fmt.Printf("t to: %v | ", token_to)
		fmt.Printf("px from: %v | ", px_from)
		fmt.Printf("px to: %v\n", px_to)

		return 0.0
	}

}


func get_newest_timestamp_for_token_from_db(token string) int64 {
	dates := returnDatesInCollection(token)

	max := int64(0)
	for _, v := range dates {
		if v > max {
			max = v
		}
	}

	return max
}


func get_newest_timestamp_from_db(pool string, tokens []string, poolid string) int64 {
/*
	var name []string
	name = append(name, pool)
	for i := 0; i < len(tokens); i++ {
		name = append(name, tokens[i])
	}

	for i := len(tokens); i < 8; i++ {
		var tokenNum []string
		tokenNum = append(tokenNum, "token")
		tokenNum = append(tokenNum, strconv.Itoa(i))
		tokenJoined := strings.Join(tokenNum, "")
		name = append(name, tokenJoined)
	}
	v := strings.Join(name, " ")
	
	//fmt.Print(v)

	truncpoolid := ""
	if len(poolid) > 4 {
		truncpoolid = poolid[len(poolid)-3:]
		xx := v + " "+ truncpoolid
		fmt.Printf("xx in GET TIMESTAMP: %v\n",xx)
		v = xx
	}
*/
	v := generate_pool_collection_name_in_db(pool,tokens,poolid)
/*
	fmt.Printf("v: %v\n", v)
	fmt.Printf("generated through func: %v\n", vv)
	fmt.Printf("ARE THEY EQUAL: %v\n", v == vv)
*/
	dates := returnDatesInCollection_pools(v)

	max := int64(0)
	for _, v := range dates {
		if v > max {
			max = v
		}
	}

	return max
}


func generate_pool_collection_name_in_db(pool string, tokens []string, poolid string) string {
	var name []string
	name = append(name, pool)
	for i := 0; i < len(tokens); i++ {
		name = append(name, tokens[i])
	}

	for i := len(tokens); i < 8; i++ {
		var tokenNum []string
		tokenNum = append(tokenNum, "token")
		tokenNum = append(tokenNum, strconv.Itoa(i))
		tokenJoined := strings.Join(tokenNum, "")
		name = append(name, tokenJoined)
	}
	v := strings.Join(name, " ")

	truncpoolid := ""
	if len(poolid) > 4 {
		truncpoolid = poolid[len(poolid)-3:]
		xx := v + " "+ truncpoolid
		fmt.Printf("xx in GET TIMESTAMP: %v\n",xx)
		v = xx
	}
	return v
}


func retrieve_pool_ratios(pool string, tokens []string, poolid string) (token0_pct_pool []float64, token1_pct_pool []float64) {
/*
	var name []string
	name = append(name, pool)
	for i := 0; i < len(tokens); i++ {
		name = append(name, tokens[i])
	}

	for i := len(tokens); i < 8; i++ {
		var tokenNum []string
		tokenNum = append(tokenNum, "token")
		tokenNum = append(tokenNum, strconv.Itoa(i))
		tokenJoined := strings.Join(tokenNum, "")
		name = append(name, tokenJoined)
	}
	v := strings.Join(name, " ")

	truncpoolid := ""
	if len(poolid) > 4 {
		truncpoolid = poolid[len(poolid)-3:]
		xx := v + " "+ truncpoolid
		fmt.Printf("xx in GET TIMESTAMP: %v\n",xx)
		v = xx
	}
*/
	v := generate_pool_collection_name_in_db(pool,tokens,poolid)

/*
	fmt.Printf("v: %v\n", v)
	fmt.Printf("generated through func: %v\n", vv)
	fmt.Printf("ARE THEY EQUAL: %v\n", v == vv)
*/
	token0_pct_pool = returnAttributeInCollectionAsFloat64(v, "token0_pct")
	token1_pct_pool = returnAttributeInCollectionAsFloat64(v, "token1_pct")
	return token0_pct_pool,token1_pct_pool
}

func retrieve_hist_pool_sizes_volumes_fees_ir(pool string, tokens []string, poolid string) (dates []int64, tradingvolumes []float64, poolsizes []float64, fees []float64, ir []float64, util_rate []float64) {
/*
	var name []string
	name = append(name, pool)
	for i := 0; i < len(tokens); i++ {
		name = append(name, tokens[i])
	}

	for i := len(tokens); i < 8; i++ {
		var tokenNum []string
		tokenNum = append(tokenNum, "token")
		tokenNum = append(tokenNum, strconv.Itoa(i))
		tokenJoined := strings.Join(tokenNum, "")
		name = append(name, tokenJoined)
	}
	v := strings.Join(name, " ")

	truncpoolid := ""
	if len(poolid) > 4 {
		truncpoolid = poolid[len(poolid)-3:]
		xx := v + " "+ truncpoolid
		fmt.Printf("xx in GET TIMESTAMP: %v\n",xx)
		v = xx
	}
*/
	v := generate_pool_collection_name_in_db(pool,tokens,poolid)
/*
	fmt.Printf("v: %v\n", v)
	fmt.Printf("generated through func: %v\n", vv)
	fmt.Printf("ARE THEY EQUAL: %v\n", v == vv)
*/

	dates = returnAttributeInCollectionAsInt64(v, "date")
	tradingvolumes = returnAttributeInCollectionAsFloat64(v, "trading_Volume_USD")
	poolsizes = returnAttributeInCollectionAsFloat64(v, "pool_sz_USD")
	fees = returnAttributeInCollectionAsFloat64(v, "fees")
	ir = returnAttributeInCollectionAsFloat64(v, "weighted_av_IR")
	util_rate = returnAttributeInCollectionAsFloat64(v, "weighted_av_IR")
	return dates, tradingvolumes, poolsizes, fees, ir, util_rate
}

func isAave1RecordsInDb() bool {
	names := getCollectionNames("De-Fi Aggregator")

	for _, name := range names {
		fmt.Println(name)
		if name == "Aave1 USDC USDC" {
			return true
		}
	}
	return false
}

func getCollectionNames(database string) []string {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("De-Fi_Aggregator")
	collection, err := Database.ListCollections(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	var collectionNames []bson.M
	err = collection.All(ctx, &collectionNames)
	if err != nil {
		log.Fatal(err)
	}

	var allNames []string
	for _, name := range collectionNames {
		aName := name["name"]
		allNames = append(allNames, fmt.Sprint(aName))
	}

	return allNames
}

func returnAttributeInCollectionAsFloat64(collectionName string, attribute string) []float64 {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("De-Fi_Aggregator")
	collection := Database.Collection(collectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var records []bson.M

	if err = cursor.All(ctx, &records); err != nil {
		log.Fatal(err)
	}

	var array []float64
	for _, record := range records {
		//fmt.Println(record)
		//fmt.Println(reflect.TypeOf(record["Date"]))
		val := record[attribute]
		//fmt.Println(date)
		//fmt.Println(reflect.TypeOf(date))
		//attributes = append(attributes, fmt.Sprint(attribute_value))
		array = append(array, val.(float64))
	}

	return array
}

func returnAttributeInCollectionAsInt64(collectionName string, attribute string) []int64 {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("De-Fi_Aggregator")
	collection := Database.Collection(collectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var records []bson.M

	if err = cursor.All(ctx, &records); err != nil {
		log.Fatal(err)
	}

	var array []int64
	for _, record := range records {
		val := record[attribute]
		array = append(array, val.(int64))
	}

	return array
}

func addHistoricalCurrencyData(date int64, price float64, CollectionOrTicker string) string {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)

	}
	defer client.Disconnect(ctx)

	Database := client.Database("test2")
	historicaldata := Database.Collection(CollectionOrTicker)

	// check if date exists in that collection - if yes return "already exists"
	cursor, err := historicaldata.Find(ctx, bson.M{"Date": date})
	if err != nil {
		log.Fatal(err)
	}

	var collectionFiltered []bson.M
	err = cursor.All(ctx, &collectionFiltered)
	if err != nil {
		log.Fatal(err)
	}

//	fmt.Print("Collection Filtered: ")
//	fmt.Println(collectionFiltered)

	if len(collectionFiltered) > 0 {
		fmt.Print("..appending new hist PX data for: ")
		fmt.Print(CollectionOrTicker)
		fmt.Print("..date ALREADY EXISTS!!..")
		return " date already in database..not added anything"
	}

	new_data, err := historicaldata.InsertOne(ctx, bson.D{
		{Key: "Date", Value: date},
		{Key: "Price", Value: price},
	})

	if err != nil {
		log.Fatal(err)
	}

	newID := new_data.InsertedID
	hexID := newID.(primitive.ObjectID).Hex()

	return hexID
}

func addPoolTokenPairReturns(pair string, poolsize float64, poolvolume float64, yield float64, pool string,
	volatility float64, roi_raw_est float64, roi_vol_adj_est float64, roi_hist float64) string {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("test2")
	ownstartingportfolio := Database.Collection("Currency Input Data")

	new_data, err := ownstartingportfolio.InsertOne(ctx, bson.D{
		{Key: "Pair", Value: pair},
		{Key: "Pool Size", Value: poolsize},
		{Key: "Pool Volume", Value: poolvolume},
		{Key: "Yield", Value: yield},
		{Key: "Pool", Value: pool},
		{Key: "Volatility", Value: volatility},
		{Key: "ROI Raw Estimation", Value: roi_raw_est},
		{Key: "ROI Vol Adjusted Estimation", Value: roi_vol_adj_est},
		{Key: "ROI History", Value: roi_hist},
	})

	if err != nil {
		log.Fatal(err)
	}

	newID := new_data.InsertedID
	hexID := newID.(primitive.ObjectID).Hex()

	return hexID
}



func returnDatesInCollection_pools(collectionName string) []int64 {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("De-Fi_Aggregator")
	collection := Database.Collection(collectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var records []bson.M

	if err = cursor.All(ctx, &records); err != nil {
		log.Fatal(err)
	}

	var dates []int64
	for _, record := range records {
		date := record["date"]
		dates = append(dates, date.(int64))
	}

	return dates
}


func returnDatesInCollection(collectionName string) []int64 {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("test2")
	collection := Database.Collection(collectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var records []bson.M

	if err = cursor.All(ctx, &records); err != nil {
		log.Fatal(err)
	}

	var dates []int64
	for _, record := range records {
		date := record["Date"]
		dates = append(dates, date.(int64))
	}

	return dates
}

func returnPricesInCollection(collectionName string) []float64 {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("test2")
	collection := Database.Collection(collectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var records []bson.M

	if err = cursor.All(ctx, &records); err != nil {
		log.Fatal(err)
	}

	var prices []float64
	for _, record := range records {
		price := record["Price"]
		prices = append(prices, price.(float64))
	}

	return prices
}

func returnAttributeInCollection(collectionName string, attribute string) []string {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("test2")
	collection := Database.Collection(collectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	var records []bson.M
	if err = cursor.All(ctx, &records); err != nil {
		log.Fatal(err)
	}

	var attributes []string
	for _, record := range records {
		attribute_value := record[attribute]
		//fmt.Println(attribute_value)
		//fmt.Println(reflect.TypeOf(attribute_value))
		attributes = append(attributes, fmt.Sprint(attribute_value))
	}

	return attributes
}

func ones(length int) []float64{
	var x []float64
	for i:=0; i < length;i++ {
		x = append(x,1.0)
	}
	return x
}


func getHistPriceDataForTokenPairFromDB(token0 string, token1 string) HistoricalCurrencyData {

	token0dataishere := true
	token1dataishere := true

	if token0 != "USD" {
		token0dataishere,_,_ = isHistDataAlreadyDownloadedDatabase(token0)
	}

	if token1 != "USD" {
		token1dataishere,_,_ = isHistDataAlreadyDownloadedDatabase(token1)
	}

	if !token0dataishere || !token1dataishere {
		fmt.Println("ERROR 899: ticker combo not found in database..returning blank object")
		return NewHistoricalCurrencyData()
	}

	var token0datesarray []int64
	var token0pricesarray []float64

	var token1datesarray []int64
	var token1pricesarray []float64

	if token0dataishere {
		token0datesarray = returnDatesInCollection(token0)
		token0pricesarray = returnPricesInCollection(token0)
	}

	// otherwise populate 1

	if token1dataishere {
		token1datesarray = returnDatesInCollection(token1)
		token1pricesarray = returnPricesInCollection(token1)
	}



	if token0 == "USD" && token1 != "USD" {
		token0datesarray = token1datesarray
		token0pricesarray = ones(len(token1datesarray))
	}

	if token0 != "USD" && token1 == "USD" {
		token1datesarray = token0datesarray
		token1pricesarray = ones(len(token0datesarray))
	}

	if len(token1datesarray) == 0 || len(token0datesarray) == 0 {
		fmt.Print("..retrieving HIST PX data from DB for: ")
		fmt.Println(token0 + "/" + token1 + " : ")	

		fmt.Print("t0 n dates: ")
		fmt.Print(len(token0datesarray))
		fmt.Print("| t1 n dates: ")
		fmt.Println(len(token1datesarray))
		log.Fatal()
	}

	var histcombo HistoricalCurrencyData
	histcombo.Ticker = token0 + "/" + token1

	newest_record_in_either := math.Max(float64(MaxIntSlice(token0datesarray)), float64(MaxIntSlice(token1datesarray))) 
	oldest_record_in_either := math.Min(float64(MinIntSlice(token0datesarray)),float64(MinIntSlice(token1datesarray)))

	if token1 == "USD" {
		for i:=0; i < len(token0datesarray);i++ {
			histcombo.Date = append(histcombo.Date, token0datesarray[i])
			histcombo.Price = append(histcombo.Price, float64(token0pricesarray[i]))
		}
	}

	if token1 != "USD" {
		for {
			newest_record_in_either -= 60 * 60 * 24
			if newest_record_in_either <= oldest_record_in_either {
				break
			} 

			// check if record exists in both
			e0,i0 := int64InSlice(int64(newest_record_in_either),token0datesarray)
			e1,i1 := int64InSlice(int64(newest_record_in_either),token1datesarray)

			// append it if it does
			if  e0 && e1 {
				histcombo.Date = append(histcombo.Date, token0datesarray[i0])
			//	fmt.Print("..337..")
				var price float64
				if token1pricesarray[i1] > 0 {
					price = token0pricesarray[i0] / token1pricesarray[i1]
					//fmt.Print("..338..")
				} else {
					price = 0.0
				}
				if math.IsInf(float64(price), 0) {
					price = 0.0
					fmt.Println("WARNING 987: Inf in calculating token combo price")
				}
				if math.IsNaN(float64(price)) {
					price = 0.0
					fmt.Println("WARNING 987: Nan in calculating token combo price")
				}

				histcombo.Price = append(histcombo.Price, float64(price))			
			} // if both in

		} // for loop
	} // if not usd
/*
	fmt.Printf("\nSummary of retrieved px hist data: \n")
	for i:= 0; i < len(histcombo.Price);i++ {
		fmt.Print(histcombo.Date[i])
		fmt.Print(" | ")
		fmt.Println(histcombo.Price[i])
	}
*/
	return histcombo
}

func isHistDataAlreadyDownloadedDatabase(token string) (bool, int, int64) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://admin:highyield4me@cluster0.tmmmg.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	Database := client.Database("test2")

	array, err := Database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	t := time.Unix(get_newest_timestamp_for_token_from_db(token), 0).Unix()

	for i := 0; i < len(array); i++ {
		if array[i] == token {
		//	fmt.Print("Found token in collections!! - ")
		//	fmt.Print(array[i])
				return true, len(returnDatesInCollection(token)), t
		}
	}
//	fmt.Print("NOT FOUND ANY DATA FOR TOKEN: ")
//	fmt.Print(token)
	return false, 0, 0
}


	/*
		fmt.Print("SIZE of returned combo for ticker: ")
		fmt.Print(histcombo.Ticker)
		fmt.Print(": ")
		fmt.Println(len(histcombo.Price))
	*/
	/*
		if len(histcombo.Price) >= 2 {
			fmt.Print(histcombo.Date[0])
			fmt.Print(" | ")
			fmt.Println(histcombo.Price[0])
			fmt.Print(histcombo.Date[1])
			fmt.Print(" | ")
			fmt.Print(histcombo.Price[1])
		}
	*/
	//	fmt.Print("returning sz of histcombo: ")
	//	fmt.Print(len(histcombo.Date))

			//	fmt.Print("i: ")
		//	fmt.Print(i)
		//	fmt.Print(" | t0: ")
		//	fmt.Print(token0datesarray[i])
		//	fmt.Print(" | px0: ")
		//	fmt.Print(token0pricesarray[i])

			//	fmt.Print(" | t1: ")
			//	fmt.Print(token1datesarray[i])
			//	fmt.Print(" | px1: ")
			//	fmt.Print(token1pricesarray[i])



			
/*
	ago0 := time.Since(time.Unix(MaxIntSlice(token0datesarray), 0))
	ago1 := time.Since(time.Unix(MaxIntSlice(token1datesarray), 0))

	if ago0.Hours() < 24 {
		fmt.Print("Data 0 is recent - no need to update! ")
		fmt.Println(ago0.Hours())
	}

	if ago1.Hours() < 24 {
		fmt.Println("Data 1 is recent - no need to update! ")
		fmt.Println(ago1.Hours())
	}
*/

/*
	for i := lengthoflookbackhist - 1; i >= 0; i-- {
		if token1 != "USD" {
			// Synchronise indices i and j in token1
			if token0datesarray[i] == token1datesarray[i] {
				histcombo.Date = append(histcombo.Date, token0datesarray[i])

				var price float64
				if token1pricesarray[i] > 0 {
					price = token0pricesarray[i] / token1pricesarray[i]
				} else {
					price = 0.0
				}
				if math.IsInf(float64(price), 0) {
					price = 0.0
					fmt.Println("WARNING 987: Inf in calculating token combo price")
				}
				if math.IsNaN(float64(price)) {
					price = 0.0
					fmt.Println("WARNING 987: Nan in calculating token combo price")
				}

				histcombo.Price = append(histcombo.Price, float64(price))
			} else if token0datesarray[i] != token1datesarray[i] { // Find if matching date exists
				fmt.Print(" | WARNING: dates do not match!..trying to loop through whole ")
				for j, dts := range token1datesarray {
					if token0datesarray[i] == dts { // if yes - append that j

						histcombo.Date = append(histcombo.Date, token0datesarray[i])
						var price float64
						if token1pricesarray[j] > 0 {
							price = token0pricesarray[i] / token1pricesarray[j]
						} else {
							price = 0.0
						}
						if math.IsInf(float64(price), 0) {
							price = 0.0
							fmt.Println("WARNING 987.5: Inf in calculating token combo price")
						}
						if math.IsNaN(float64(price)) {
							price = 0.0
							fmt.Println("WARNING 987.5: Nan in calculating token combo price")
						}
		
						histcombo.Price = append(histcombo.Price, float64(price))

					} else { // if no - skip this date
						// do nothing
					}
				}
			}
		} else if token1 == "USD" {
			histcombo.Date = append(histcombo.Date, token0datesarray[i])
			histcombo.Price = append(histcombo.Price, float64(token0pricesarray[i]))
		}

	}
*/

/*
	lengthoflookbackhist := len(token0datesarray)
	if token1 != "USD" {
		lengthoflookbackhist2 := len(token1datesarray)
		lengthoflookbackhist = int(math.Min(float64(lengthoflookbackhist), float64(lengthoflookbackhist2)))

			fmt.Print("length of lookback = ")
			fmt.Println(lengthoflookbackhist)
	} else {
		lengthoflookbackhist2 := lengthoflookbackhist
		lengthoflookbackhist = int(math.Min(float64(lengthoflookbackhist), float64(lengthoflookbackhist2)))
	}
*/


/*
	var name []string
	token0 := "N/A"
	token1 := "N/A"
	token2 := "N/A"
	token3 := "N/A"
	token4 := "N/A"
	token5 := "N/A"
	token6 := "N/A"
	token7 := "N/A"

	name = append(name, pool)
	for i := 0; i < len(tokens); i++ {
		name = append(name, tokens[i])
	}

	for i := 0; i < 8; i++ {
		switch i {
		case 0:
			if i < len(tokens) {
				token0 = tokens[i]
			} else {
				name = append(name, "token0")
				// name += " " + "token" + strconv.Itoa(i)
			}
		case 1:
			if i < len(tokens) {
				token1 = tokens[i]
			} else {
				name = append(name, "token1")
			}
		case 2:
			if i < len(tokens) {
				token2 = tokens[i]
			} else {
				name = append(name, "token2")
			}
		case 3:
			if i < len(tokens) {
				token3 = tokens[i]
			} else {
				name = append(name, "token3")
			}
		case 4:
			if i < len(tokens) {
				token4 = tokens[i]
			} else {
				name = append(name, "token4")
			}
		case 5:
			if i < len(tokens) {
				token5 = tokens[i]
			} else {
				name = append(name, "token5")
			}
		case 6:
			if i < len(tokens) {
				token6 = tokens[i]
			} else {
				name = append(name, "token6")
			}
		case 7:
			if i < len(tokens) {
				token7 = tokens[i]
			} else {
				name = append(name, "token7")
			}
		}
	}
	v := strings.Join(name, " ")

	truncpoolid := ""
	if len(poolid) > 4 {
		truncpoolid = poolid[len(poolid)-3:]
		xx := v + " "+ truncpoolid
		fmt.Printf("xx in GET TIMESTAMP: %v\n",xx)
		v = xx	
	}
*/