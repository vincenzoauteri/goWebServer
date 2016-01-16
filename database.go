package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

func initDb(account string) *sql.DB {

    db, err := sql.Open("sqlite3", fmt.Sprintf("./%s.db",account))

    if err != nil {
        log.Fatal(err)
    }

    sqlStmt := `SELECT name FROM sqlite_master WHERE type='table' AND name='position';`

    var name string
    err = db.QueryRow(sqlStmt).Scan(&name);

    if err == sql.ErrNoRows {
        sqlStmt = `
        CREATE table position (stock string not null primary key, owned integer , balance integer) ;
        CREATE table orders (id number not null primary key, stock string, direction string, type string, price number, qty number, filled number, open integer);
        `;
        _ , err = db.Exec(sqlStmt);

        if err != nil {
            log.Fatal(err)
        } else {
            fmt.Printf("Database :%s.db created\n",account);
        }
    } else if err != nil {
        log.Print(err)
    } else {
        fmt.Printf("Database :%s.db exists\n",account);
    }

    //defer db.Close()
    return db
}

    /*
    //Insert reference code
    if tempJson.Ok {
        sqlStmt := fmt.Sprintf(`INSERT INTO orders 
        (id, stock, direction, type, price, qty, filled, open ) 
        VALUES 
        (%d, "%s", "%s", "%s", %d, %d, %d, %d);`,
        tempJson.Id,
        tempJson.Symbol,tempJson.Direction,tempJson.OrderType,
        tempJson.Price,tempJson.Qty,tempJson.TotalFilled, btoi(tempJson.Open));

        _ , err = db.Exec (sqlStmt);

        if err != nil {
            log.Print(err)

        }
    }
    */

func update_position_sql(stock string, change int, price int, db *sql.DB){

    type Position struct {
        Stock string
        Owned int
        Balance int
    }

    var pos Position

    sqlStmt := fmt.Sprintf(`SELECT * FROM position WHERE stock="%s";`,stock);

    err := db.QueryRow(sqlStmt).Scan(&pos.Stock,&pos.Owned,&pos.Balance)

    if err==sql.ErrNoRows{
        sqlStmt := fmt.Sprintf(`INSERT INTO position
        (stock, owned, balance) 
        VALUES 
        ("%s", %d, %d);`,stock,0,0);
        _ , err = db.Exec(sqlStmt);
    }

    if err!=nil {
        log.Fatal(err)
    }

    sqlStmt = fmt.Sprintf(`UPDATE position SET owned=%d, balance=%d  WHERE stock="%s";`,
    pos.Owned+change, pos.Balance+price, pos.Stock);
    _ , err = db.Exec(sqlStmt);

    if err!=nil {
        log.Fatal(err)
    }
}
