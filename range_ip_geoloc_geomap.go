package main

import (
    "bytes"
    "database/sql"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "sync"
    "time"
    "strings"

    _ "github.com/go-sql-driver/mysql"
)

var token string = "*************"

type locResult struct {
    IP       string
    Location string
    Error    error
}

func doPostRequest(url string, data []byte) ([]byte, error) {
    client := &http.Client{
        Timeout: 10 * time.Second,
    }

    req, err := http.NewRequest("GET", url, bytes.NewBuffer(data))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return ioutil.ReadAll(resp.Body)
}

func checkIp(ip string, resultsChan chan<- locResult, wg *sync.WaitGroup) {
    defer wg.Done()
    url := "https://ipinfo.io/" + ip + "?token=" + token

    data, err := doPostRequest(url, nil)
    if err != nil {
        resultsChan <- locResult{IP: ip, Error: err}
        return
    }

    var info struct {
        IP       string `json:"ip"`
        Location string `json:"loc"`
    }
    err = json.Unmarshal(data, &info)
    if err != nil {
        resultsChan <- locResult{IP: ip, Error: err}
        return
    }

    resultsChan <- locResult{IP: ip, Location: info.Location}
}

func main() {
    ipAddresses := []string{"63.32.210.222", "37.00.90.00", "194.163.179.176"}

    resultsChan := make(chan locResult, len(ipAddresses))
    var wg sync.WaitGroup

    for _, ip := range ipAddresses {
        wg.Add(1)
        go checkIp(ip, resultsChan, &wg)
    }

    wg.Wait()
    close(resultsChan)

    // connect with BD 
    db, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/iploc")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // createBD-tabl
    createTableSQL := `
        CREATE TABLE IF NOT EXISTS ip_coordinates (
            id INT AUTO_INCREMENT PRIMARY KEY,
            ip_address VARCHAR(45) NOT NULL,
            latitude DECIMAL(9, 6) NOT NULL,
            longitude DECIMAL(9, 6) NOT NULL
        )
    `

    // DO SQL 
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("ok")

    // save bd
    for result := range resultsChan {
        if result.Error != nil {
            fmt.Printf("err IP %s: %v\n", result.IP, result.Error)
        } else {
            // lat & long
            parts := strings.Split(result.Location, ",")
            if len(parts) != 2 {
                fmt.Printf("err Location for IP %s: %v\n", result.IP, result.Location)
                continue
            }

            // to SQL BD
            _, err := db.Exec("INSERT INTO ip_coordinates (ip_address, latitude, longitude) VALUES (?, ?, ?)", result.IP, parts[0], parts[1])
            if err != nil {
                log.Fatal(err)
            }
        }
    }
}
