package main

import (
    "time"
    "fmt"
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


func monitorOrders (account string, venue string, stock string) {
    ok , _, _ := place_order(venue, stock, "buy" , account, 10,0, market)

    if !ok {
        fmt.Printf("Cannot place initial order\n")
        return
    }

    for ;; {
        if (check_venue (venue)) {
            place_order(venue, stock, "buy" , account, 1,0, "market")
            place_order(venue, stock, "sell" , account, 1,0, "market")
            time.Sleep(time.Duration(1000) * time.Millisecond)
        } else {
            return;
        }
    }
}


var templates = template.Must(template.ParseFiles("select.html","monitor.html"))
var validPath = regexp.MustCompile("^/accounts/([a-zA-Z0-9]+)/venues/([a-zA-Z0-9]+)/stocks/([a-zA-Z0-9]+)$")

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
    p:= &Page{account,stock,venue,stock,}
    go monitorOrders(account,venue,stock);
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


func main() {
    content, err := ioutil.ReadFile("./keyfile.dat")
    if err != nil {
        log.Fatal(err)
    }

    //Init globals
    globals.ApiKey = string(content);
    globals.httpClient = http.Client{}
    f, err := os.OpenFile("log", os.O_RDWR | os.O_CREATE, 0666)
    if err != nil {
        log.Fatal("error opening file: %v", err)
    }
    defer f.Close()
    log.SetOutput(f)

    fs := http.FileServer(http.Dir(""))

    http.HandleFunc("/accounts/", monitorHandler)
    http.Handle("/static/",fs)
    http.HandleFunc("/save", saveHandler)
    http.HandleFunc("/select",selectHandler)
    http.ListenAndServe(":8080", nil)
}
