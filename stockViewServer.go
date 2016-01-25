package main
import (
    "time"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "html/template"
    "net/http"
    "regexp"
    "log"
    "errors"
    "os"
)

type Page struct {
    Title string
    Account string
    Venue string
    Stock string
}




var templates = template.Must(template.ParseFiles("select.html","monitor.html"))
var validPath = regexp.MustCompile("^/accounts/([a-zA-Z0-9]+)/venues/([a-zA-Z0-9]+)/stocks/([a-zA-Z0-9]+)$")
var ajaxPath = regexp.MustCompile("^/update/([a-zA-Z0-9]+)$")

func selectHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "select" ,nil)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
    account:= r.FormValue("account")
    venue:= r.FormValue("venue")
    stock := r.FormValue("stock")
    if venue == "" ||  stock == "" || account == ""{
        http.Redirect(w, r, "/select",http.StatusFound)
        return
    }
    http.Redirect(w, r, "/accounts/"+account+"/venues/"+venue+"/stocks/"+stock, http.StatusFound)
}

func monitorHandler(w http.ResponseWriter, r *http.Request) {
    account, venue, stock  ,err  := getDataFromUrl(w, r)

    if err != nil {
        http.Redirect(w, r, "/select",http.StatusFound)
        return
    }
     _, err = r.Cookie("account");
    if err != nil {
        expiration := time.Now().Add(1*time.Hour)
        accountCookie := http.Cookie{Name: "account", Value: account, Expires: expiration}
        venueCookie := http.Cookie{Name: "venue", Value: venue, Expires: expiration}
        stockCookie := http.Cookie{Name: "stock", Value: stock, Expires: expiration}
        http.SetCookie(w,&accountCookie);
        http.SetCookie(w,&venueCookie);
        http.SetCookie(w,&stockCookie);
    }
    p:= &Page{stock, account, venue, stock,}
    solve_level6(account, venue, stock)
    renderTemplate(w, "monitor" , p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html",p)
    if err != nil {
        fmt.Println(err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func getDataFromUrl(w http.ResponseWriter, r *http.Request) (string,string, string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        return "","","", errors.New("Invalid Page Title")
    }
    return m[1], m[2], m[3], nil // The title is the second subexpression.
}

func testHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`[{"id": 1, "author": "Pete Hunt", "text": "This is one comment"}, {"id": 2, "author": "Jordan Walke", "text": "This is *another* comment"}]`));
}


func ajaxHandler(w http.ResponseWriter, r *http.Request) {
    m := ajaxPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        fmt.Println("Invalid Ajax Request")
        http.Error(w, "Invalid Request", http.StatusInternalServerError)
    }
    venueId := m[1]
    var posArray []PositionT
    for _, account:= range gameData.Venues[venueId].Accounts {
        posArray = append(posArray, account.Position)
    }
    response, _ := json.Marshal(posArray)
    w.Write(response)
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
    gameData.Venues = make(map[string]*VenueT);

    f, err := os.OpenFile("log", os.O_RDWR | os.O_CREATE, 0666)
    if err != nil {
        log.Fatal("error opening file: %v", err)
    }
    defer f.Close()
    log.SetOutput(f)

    fs := http.FileServer(http.Dir(""))

    http.HandleFunc("/test", testHandler)
    http.HandleFunc("/update/", ajaxHandler)
    http.HandleFunc("/accounts/", monitorHandler)
    http.Handle("/static/",fs)
    http.HandleFunc("/save", saveHandler)
    http.HandleFunc("/select",selectHandler)
    fmt.Println("Server starting now")
    http.ListenAndServe(":8080", nil)
}
