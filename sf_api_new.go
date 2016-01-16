package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "golang.org/x/net/websocket"
    "regexp"
    "time"
)

type OrderBookT struct {
    Ok     bool   `json:"ok"`
    Venue  string `json:"venue"`
    Symbol string `json:"symbol"`
    Bids   []struct {
        Price int  `json:"price"`
        Qty   int  `json:"qty"`
        IsBuy bool `json:"isBuy"`
    } `json:"bids"`
    Asks []struct {
        Price int  `json:"price"`
        Qty   int  `json:"qty"`
        IsBuy bool `json:"isBuy"`
    } `json:"asks"`
    Ts string `json:ts`
}


type StockQuoteWsT struct {
    Ok bool  `json:"ok"`
    Quote struct {
        Symbol string `json:"symbol"`
        Venue string  `json:"venue"`
        Bid int  `json:"bid"`
        Ask int `json:"ask"`
        BidSize int  `json:"bidSize"`
        AskSize int `json:"askSize"`
        BidDepth int  `json:"bidDepth"`
        AskDepth int `json:"askDepth"`
        Last int `json:"last"`
        LastSize int `json:"lastSize"`
        LastTrade string  `json:"lastTrade"`
        QuoteTime string `json:"quoteTime"`
    } `json:quote`
}


type QuoteHistoryT struct {
    ready bool
    history []StockQuoteWsT

    lastTopBidPrice int
    lastTopAskPrice int
    avgTopBidQty float64
    avgTopAskQty float64

    avgTopBidPrice float64
    avgTopAskPrice float64

    minTopBidPrice int
    maxTopBidPrice int

    minTopAskPrice int
    maxTopAskPrice int

    lastBidPrice int
    lastAskPrice int

    lastBidId int
    lastAskId int
}


type OrderBookHistoryT struct {
    ready bool
    history []OrderBookT
    lastTopBidPrice int
    lastTopAskPrice int
    avgTopBidQty float64
    avgTopAskQty float64

    avgTopBidPrice float64
    avgTopAskPrice float64

    minTopBidPrice int
    maxTopBidPrice int

    minTopAskPrice int
    maxTopAskPrice int
    quotedAsk int
    quotedBid int
}

var globals struct {
    ApiKey string
    httpClient http.Client
    wsExecutions *websocket.Conn
    ownAccountId string
}

var gameData struct {
    Venues map[string]VenueT
}

type PositionT struct {
    Owned map[string]int
    Cash int
    NAV int
}

type StockT struct {
    Id string
    Quote int
    OrderBookHistory OrderBookHistoryT
    QuoteHistory QuoteHistoryT
}

type VenueT struct {
    Id string
    Stocks map[string]StockT
    Accounts map[string]AccountT
    Orders map[int]OrderT
    wsQuote *websocket.Conn
}

type AccountT struct {
    Id string
    VenueId string
    Position PositionT
    wsExecutions *websocket.Conn
}


type AllOrdersT struct {
    Ok     bool     `json:"ok"`
    Venue  string   `json:"venue"`
    Orders []OrderT  `json:"orders"`
}


type OrderT struct {
    Ok          bool   `json:"ok"`
    Symbol      string `json:"symbol"`
    Venue       string `json:"venue"`
    Direction   string `json:"direction"`
    OriginalQty int    `json:"originalQty"`
    Qty         int    `json:"qty"`
    Price       int    `json:"price"`
    OrderType   string `json:"orderType"`
    Id          int    `json:"id"`
    Account     string `json:"account"`
    Ts          string `json:ts`

    Fills []struct {
        Price int    `json:"price"`
        Qty   int    `json:"qty"`
        Ts    string `json:ts`
    } `json:"fills"`
    TotalFilled int  `json:"totalFilled"`
    Open        bool `json:"open"`
}

type ExecutionsT struct {
    Ok          bool   `json:"ok"`
    Account     string `json:"account"`
    Venue       string `json:"venue"`
    Symbol      string `json:"symbol"`
    Order OrderT `json:"order"`
    StandingId int `json:"standingId"`
    IncomingId int `json:"incomingId"`
    Price       int    `json:"price"`
    Filled int  `json:"filled"`
    FilledAt string `json:"filledAt"`
    StandingComplete bool `json:"standingComplete"`
    IncomingComplete bool `json:"incomingComplete"`
}


func heartbeat() bool {
    httpRequest, err := http.NewRequest("GET", "https://api.stockfighter.io/ob/api/heartbeat", nil)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
        return false
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)
    httpResponse.Body.Close()

    type HeartbeatResponseT struct {
        Ok    bool   `json:"ok"`
        Error string `json:"error"`
    }

    var tempJson HeartbeatResponseT

    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
    }
    //fmt.Printf("%+v\n", tempJson)
    return tempJson.Ok
}


func check_venue_request(venueId string) bool {
    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/heartbeat", venueId)
    httpRequest, err := http.NewRequest("GET", requestUrl, nil)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
        return false
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)

    type VenueResponseT struct {
        Ok    bool   `json:"ok"`
        Venue string `json:"venue"`
    }

    var tempJson VenueResponseT

    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
    }

    return tempJson.Ok
}

func check_stocks_request(venueId string) bool {
    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/stocks", venueId)
    httpRequest, err := http.NewRequest("GET", requestUrl, nil)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print (err)
        return false
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)
    //fmt.Printf("%s\n", responseData)

    type stockResponse struct {
        Ok      bool `json:"ok"`
        Symbols []struct {
            Name   string `json:"name"`
            Symbol string `json:"symbol"`
        } `json:"symbols"`
    }
    var tempJson stockResponse

    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
    } else if tempJson.Ok {
        for i := range tempJson.Symbols {
            stockId:= tempJson.Symbols[i].Symbol
            _, exists := gameData.Venues[venueId].Stocks[stockId]
            if !exists {
                gameData.Venues[venueId].Stocks[stockId] =  StockT{stockId, 0, OrderBookHistoryT{}, QuoteHistoryT{}};
            }
        }
    }

    //fmt.Printf("%+v\n", tempJson)
    return tempJson.Ok
}

func quote_stock_request(venueId string, stockId string) bool {
    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/quote", venueId, stockId)
    httpRequest, err := http.NewRequest("GET", requestUrl, nil)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
        return false
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)
    //fmt.Printf("%s\n",responseData)

    type StockQuoteT struct {
        Ok bool  `json:"ok"`
        Symbol string `json:"symbol"`
        Venue string  `json:"venue"`
        Bid int  `json:"bid"`
        Ask int `json:"ask"`
        BidSize int  `json:"bidSize"`
        AskSize int `json:"askSize"`
        BidDepth int  `json:"bidDepth"`
        AskDepth int `json:"askDepth"`
        Last int `json:"last"`
        LastSize int `json:"lastSize"`
        LastTrade string  `json:"lastTrade"`
        QuoteTime string `json:"quoteTime"`
    }

    var tempJson StockQuoteT

    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
    } else if tempJson.Ok {
        _, exists := gameData.Venues[venueId].Stocks[stockId]
        if !exists {
            gameData.Venues[venueId].Stocks[stockId] =  StockT{stockId,tempJson.Last, OrderBookHistoryT{}, QuoteHistoryT{}};
        } else {
            stock := gameData.Venues[venueId].Stocks[stockId]
            stock.Quote = tempJson.Last
            gameData.Venues[venueId].Stocks[stockId] = stock
        }
        //fmt.Printf("%+v\n",gameData.Venues[venueId].Stocks[stockId])
    }
    return tempJson.Ok
}

func place_order_request(venue string, stock string, direction string, account string, qty int, price int, orderType string) (bool, int) {


    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/orders", venue, stock)

    //POST data here
    var jsonStr = []byte(fmt.Sprintf(" { \"venue\":\"%s\",\"stock\":\"%s\",\"account\": \"%s\",\"price\":%d, \"qty\":%d,\"direction\":\"%s\", \"ordertype\":\"%s\" }", venue, stock, account, price, qty, direction, orderType))

    httpRequest, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonStr))
    httpRequest.Header.Add("X-Starfighter-Authorization", globals.ApiKey)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
        return false, 0
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)

    var tempJson OrderT

    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
    }

    if tempJson.Ok {
    } else {
        log.Print("Error %s",string(responseData))
    }

    return tempJson.Ok, tempJson.Id
}

func cancel_order_request(venue string, stock string, id int) bool {

    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/orders/%d", venue, stock, id)

    httpRequest, err := http.NewRequest("DELETE", requestUrl, nil)
    if err != nil {
        log.Print(err)
        return false
    }
    httpRequest.Header.Add("X-Starfighter-Authorization", globals.ApiKey)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
        return false
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)
    //fmt.Printf("%s\n", responseData)

    var tempJson OrderT

    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
    }

    return tempJson.Ok
}

func order_book_request(venueId string, stockId string) (bool) {

    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/stocks/%s", venueId, stockId)
    httpRequest, err := http.NewRequest("GET", requestUrl, nil)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
        return false
    }

    var tempJson OrderBookT

    responseData, err := ioutil.ReadAll(httpResponse.Body)


    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
    } else if tempJson.Ok {
        _ , exists  := gameData.Venues[venueId].Stocks[stockId]
        if (!exists) {
            return false
        } else {
            stock := gameData.Venues[venueId].Stocks[stockId]
            stock.OrderBookHistory.history = append(stock.OrderBookHistory.history ,tempJson)
            gameData.Venues[venueId].Stocks[stockId]= stock
        }

    }

    return tempJson.Ok
}

func process_order_book(orderBookHistory OrderBookHistoryT) {
    MOVING_AVERAGE := 100
    historyLength := len(orderBookHistory.history)
    orderBookHistory.avgTopBidQty = 0;
    orderBookHistory.avgTopBidPrice= 0;
    orderBookHistory.avgTopAskQty= 0;
    orderBookHistory.avgTopAskPrice= 0;
    orderBookHistory.minTopBidPrice= 100000;
    orderBookHistory.maxTopBidPrice= 0;
    orderBookHistory.minTopAskPrice= 100000;
    orderBookHistory.maxTopAskPrice= 0;
    if historyLength > MOVING_AVERAGE {
        orderBookHistory.ready = true
        bidAvg:=.0
        askAvg:=.0
        for _, ob := range orderBookHistory.history[historyLength - MOVING_AVERAGE -1 : historyLength -1] {
            if  len(ob.Bids) > 0 {
                orderBookHistory.avgTopBidQty += float64(ob.Bids[0].Qty)
                orderBookHistory.avgTopBidPrice += float64(ob.Bids[0].Price)
                orderBookHistory.lastTopBidPrice = ob.Bids[0].Price
                if  ob.Bids[0].Price >  orderBookHistory.maxTopBidPrice {
                    orderBookHistory.maxTopBidPrice  = ob.Bids[0].Price
                }
                if  ob.Bids[0].Price <  orderBookHistory.minTopBidPrice {
                    orderBookHistory.minTopBidPrice  = ob.Bids[0].Price
                }
                bidAvg+=1
            }

            if  len(ob.Asks) > 0 {
                orderBookHistory.avgTopAskQty += float64(ob.Asks[0].Qty)
                orderBookHistory.avgTopAskPrice+= float64(ob.Asks[0].Price)
                orderBookHistory.lastTopAskPrice = ob.Asks[0].Price
                if  ob.Asks[0].Price >  orderBookHistory.maxTopAskPrice {
                    orderBookHistory.maxTopAskPrice= ob.Asks[0].Price
                }
                if  ob.Asks[0].Price <  orderBookHistory.minTopAskPrice {
                    orderBookHistory.minTopAskPrice= ob.Asks[0].Price
                }
                askAvg+=1
            }
        }
        if bidAvg > 0 {
            orderBookHistory.avgTopBidQty /= bidAvg;
            orderBookHistory.avgTopBidPrice /= bidAvg;
        }
        if askAvg > 0 {
            orderBookHistory.avgTopAskQty /= askAvg;
            orderBookHistory.avgTopAskPrice /= askAvg;
        }

    }

}


func (stock *StockT) process_tickertape() {
    quoteHistory := stock.QuoteHistory;
    MOVING_AVERAGE := 50
    quoteHistory.avgTopBidQty = 0;
    quoteHistory.avgTopBidPrice= 0;
    quoteHistory.avgTopAskQty= 0;
    quoteHistory.avgTopAskPrice= 0;
    quoteHistory.minTopBidPrice= 100000;
    quoteHistory.maxTopBidPrice= 0;
    quoteHistory.minTopAskPrice= 100000;
    quoteHistory.maxTopAskPrice= 0;
    historyLength := len(quoteHistory.history)
    //fmt.Printf("History Length: %d\n\n", historyLength)
    if historyLength > MOVING_AVERAGE {
        quoteHistory.history = quoteHistory.history[historyLength - MOVING_AVERAGE -1 : historyLength -1]
        quoteHistory.ready = true
        bidAvg:=.0
        askAvg:=.0
        for _, stockQuoteIter := range quoteHistory.history {
            quote:= stockQuoteIter.Quote
            quoteHistory.avgTopBidQty += float64(quote.BidSize)
            quoteHistory.avgTopBidPrice += float64(quote.Bid)
            quoteHistory.lastTopBidPrice = quote.Bid
            if  quote.Bid>  quoteHistory.maxTopBidPrice {
                quoteHistory.maxTopBidPrice  = quote.Bid
            }
            if  quote.Bid <  quoteHistory.minTopBidPrice {
                quoteHistory.minTopBidPrice  = quote.Bid
            }
            quoteHistory.lastBidPrice = quote.Bid;
            bidAvg+=1

            quoteHistory.avgTopAskQty += float64(quote.AskSize)
            quoteHistory.avgTopAskPrice+= float64(quote.Ask)
            quoteHistory.lastTopAskPrice = quote.Ask
            if  quote.Ask >  quoteHistory.maxTopAskPrice {
                quoteHistory.maxTopAskPrice= quote.Ask
            }
            if  quote.Ask <  quoteHistory.minTopAskPrice {
                quoteHistory.minTopAskPrice= quote.Ask
            }
            quoteHistory.lastAskPrice = quote.Ask;
            askAvg+=1
        }
        if bidAvg > 0 {
            quoteHistory.avgTopBidQty /= bidAvg;
            quoteHistory.avgTopBidPrice /= bidAvg;
        }
        if askAvg > 0 {
            quoteHistory.avgTopAskQty /= askAvg;
            quoteHistory.avgTopAskPrice /= askAvg;
        }
    }
}

func check_order_status_request(orderId int, venueId string, stockId string) bool {

    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/orders/%d", venueId, stockId, orderId)

    httpRequest, err := http.NewRequest("GET", requestUrl, nil)
    httpRequest.Header.Add("X-Starfighter-Authorization", globals.ApiKey)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
        return false
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)
    //fmt.Printf("%s\n", responseData)

    var tempJson OrderT

    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
        return false
    }

    if tempJson.Ok {
        gameData.Venues[venueId].Orders[orderId] = tempJson
    }

    return tempJson.Ok
}

func (account *AccountT) init_executions_websocket() bool{
    origin := "http://localhost/"

    urlExecutions := fmt.Sprintf("wss://api.stockfighter.io/ob/api/ws/%s/venues/%s/executions",
    account.Id, account.VenueId);

    wsExecutions, err:= websocket.Dial(urlExecutions, "", origin)

    if err!= nil {
        log.Print(err)
        return false
    }

    account.wsExecutions = wsExecutions
    return true
}

func (venue *VenueT) init_tickertape_websocket() bool {
    origin := "http://localhost/"
    urlQuote := fmt.Sprintf("wss://api.stockfighter.io/ob/api/ws/%s/venues/%s/tickertape",
    "ABC00000000", venue.Id);

    wsQuote, err:= websocket.Dial(urlQuote, "", origin)

    if err!= nil {
        log.Print(err)
        return false
    }
    venue.wsQuote = wsQuote

    return true
}


func (venue *VenueT) update_quotes_ws() {
    for ;; {
        var stockQuoteWs StockQuoteWsT
        err:= websocket.JSON.Receive(venue.wsQuote, &stockQuoteWs)
        if err!= nil {
            log.Print(err)
            return
        }
        stock,exists := venue.Stocks[stockQuoteWs.Quote.Symbol]
        if exists {
           stock.QuoteHistory.history = append(stock.QuoteHistory.history,stockQuoteWs)
           venue.Stocks[stockQuoteWs.Quote.Symbol] = stock
           //fmt.Printf("%+v\n",venue.Stocks[stockQuoteWs.Quote.Symbol].QuoteHistory)
        }
        time.Sleep(100*time.Millisecond);
    }
}

func (account *AccountT) update_executions_ws() {
    for ;; {
        var executions ExecutionsT
        err:= websocket.JSON.Receive(account.wsExecutions, &executions)
        if err!= nil {
            log.Print(err)
            return
        }
        account.update_position_from_executions(executions)
        time.Sleep(100*time.Millisecond);
    }
}


func (account *AccountT) update_position_from_executions(executions ExecutionsT) {
    position := account.Position
    order := executions.Order

    if order.Direction ==  "buy" {
        for i:= range order.Fills {
            fill := order.Fills[i]
            position.Cash -= fill.Qty*fill.Price
            position.Owned[order.Symbol] += fill.Qty
        }
    }
    if order.Direction ==  "sell" {
        for i:= range order.Fills {
            fill := order.Fills[i]
            position.Cash += fill.Qty*fill.Price
            position.Owned[order.Symbol] -= fill.Qty
        }
    }
    account.Position = position
}

func get_all_orders_request(accountId string, venueId string, stockId string) bool {

    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/accounts/%s/stocks/%s/orders", venueId, accountId, stockId)

    httpRequest, err := http.NewRequest("GET", requestUrl, nil)
    httpRequest.Header.Add("X-Starfighter-Authorization", globals.ApiKey)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)
    //fmt.Printf("%s\n", responseData)

    var tempJson AllOrdersT

    err = json.Unmarshal([]byte(responseData), &tempJson)

    if err != nil {
        log.Print(err)
        return false
    }

    if tempJson.Ok {
        _, exists := gameData.Venues[venueId]
        if exists {
            venue:= gameData.Venues[venueId]
            for _, order := range tempJson.Orders {
                venue.Orders[order.Id] = order
            }
            gameData.Venues[venueId] = venue
        }
    }
    return tempJson.Ok
}

func add_venue(venueId string) {
    gameData.Venues[venueId] =  VenueT{venueId, make(map[string]StockT), make(map[string]AccountT), make(map[int]OrderT), nil}

    venue:= gameData.Venues[venueId]
    venue.init_tickertape_websocket()
    go venue.update_quotes_ws()
}

func (venue *VenueT) add_account(accountId string) {
    venue.Accounts[accountId] = AccountT{accountId, venue.Id, PositionT{}, nil}

    account := venue.Accounts[accountId]

    account.Position.Owned = make(map[string]int)
    for _ , stock := range venue.Stocks {
        account.Position.Owned[stock.Id] = 0
    }
    fmt.Printf("%+v\n",account.Position)

    account.init_executions_websocket()
    go account.update_executions_ws()
    venue.Accounts[accountId] = account
}

func (account* AccountT) update_position() {
    position := account.Position
    //fmt.Printf("%+v\n",position)
    totalValueOwnedStocks := 0
    for stockId , owned := range position.Owned {
        totalValueOwnedStocks += owned * gameData.Venues[account.VenueId].Stocks[stockId].Quote
    }
    position.NAV = position.Cash + totalValueOwnedStocks
    account.Position = position
}

func (venue *VenueT) update_all_positions() {
    for _, account:= range venue.Accounts {
         account.update_position()
    }
}

var validPath = regexp.MustCompile("account ([A-Z0-9]+)")
func get_account_from_cancel_order(venueId string, stockId string, orderId int) string{

    requestUrl := fmt.Sprintf("https://api.stockfighter.io/ob/api/venues/%s/stocks/%s/orders/%d", venueId, stockId, orderId)

    httpRequest, err := http.NewRequest("DELETE", requestUrl, nil)
    if err != nil {
        log.Print(err)
        return ""
    }
    httpRequest.Header.Add("X-Starfighter-Authorization", globals.ApiKey)
    httpResponse, err := globals.httpClient.Do(httpRequest)

    if err != nil {
        log.Print(err)
        return ""
    }

    responseData, err := ioutil.ReadAll(httpResponse.Body)
    //fmt.Printf("%s\n", responseData)

    var tempJson OrderT

    err = json.Unmarshal([]byte(responseData), &tempJson)

    var matches []string
    if err != nil {
        log.Print(err)
    } else if !tempJson.Ok   {
        matches = validPath.FindStringSubmatch(string(responseData))
    }


    if len(matches) > 0 {
        //fmt.Printf("%s\n", matches[1])
        return matches[1]
    } else {
        return ""
    }
}


func collect_accounts (venueId string, stockId string, accountId string) {
    lastOrder := 0;
    accountCounter :=0;
    venue := gameData.Venues[venueId]
    for ;; {
        ok , id  := place_order_request(venueId, stockId, "buy" , accountId, 1, 0, "market")
        if ok {
            for ;lastOrder < id ; lastOrder++ {
                accountId:= get_account_from_cancel_order(venue.Id , stockId , lastOrder);
                _,accountExists:= venue.Accounts[accountId]
                if !accountExists {
                    venue.add_account(accountId)
                    accountCounter +=1;
                    //fmt.Printf("Total accounts %d\n",accountCounter)
                }
            }
        }
        time.Sleep(time.Duration(1000) * time.Millisecond)
    }
}


func main() {
    content, err := ioutil.ReadFile("./keyfile.dat")
    if err != nil {
        log.Fatal(err)
    }

    //Init globals
    globals.ApiKey = string(content);
    globals.httpClient = http.Client{}
    //Init globals
    globals.ApiKey = string(content);
    globals.httpClient = http.Client{}

    //Init game data
    gameData.Venues = make(map[string]VenueT);
    heartbeat()
    currentVenue := "MWNBEX"

    if check_venue_request(currentVenue) {
        add_venue(currentVenue)
        check_stocks_request(currentVenue)
    }

    var myAccount = "CES13776159"
    venue, exists := gameData.Venues[currentVenue]
    if  exists {
        venue.add_account(myAccount)
    }

    tick :=0
    for ;; {
        //fmt.Printf("Tick  %d\n",tick);
        tick+=1;
        for _ , venue := range gameData.Venues {
            check_stocks_request(venue.Id);
            for _ , stock := range gameData.Venues[venue.Id].Stocks {
                quote_stock_request(venue.Id,stock.Id)
                //order_book_request(venue.Id,stock.Id)
                //get_all_orders_request(myAccount,venue.Id,stock.Id);
                //process_order_book(gameData.Venues[venue.Id].Stocks[stock.Id].OrderBookHistory)
                go collect_accounts(venue.Id, stock.Id, myAccount)
            }
            //fmt.Printf(" %+v\n",gameData.Venues[venue.Id].Orders);
            venue.update_all_positions()
            for _,account:= range venue.Accounts {
                fmt.Printf("ACCOUNT %s CASH %d NAV %d OWNED %+v\n",
                account.Id, account.Position.Cash, account.Position.NAV, account.Position.Owned);
            }
        }
        time.Sleep(10000*time.Millisecond)
    }

}
