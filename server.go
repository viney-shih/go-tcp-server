package main

import (
    "bufio"
    "fmt"
    "net"
    "time"
    "strings"
    "encoding/json"
    "net/http"
    "net/url"
    "io/ioutil"
    "sync/atomic"
)

var connCounter int64
var reqCounter int64

func printHelp() string{
    help := []string {}

    help = append(help, " Commands:\n")
    help = append(help, "\tweather[,city]\t\t(Temperature of city)\n")
    help = append(help, "\ttime\t\t\t(local time in the server)\n")
    help = append(help, "\tstatus\t\t\t(statistics)\n")
    help = append(help, "\tquit\n")
    help = append(help, "\thelp\n")

    return strings.Join(help, "")
}

func getWeatherURL(url string) string {
    defer func() {
        atomic.AddInt64(&reqCounter, -1)
    }()
    // Timeout mechanism
    client := http.Client{
        Timeout: time.Duration(10 * time.Second),
    }

    res, err := client.Get(url)
    if err != nil {
        // panic(err.Error())
        fmt.Println(" [x] Network error")
        return ""
    }

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        // panic(err.Error())
        fmt.Println(" [x] Body format error")
        return ""
    }

    var data map[string]interface{}
    err = json.Unmarshal(body, &data)
    if err != nil {
        // panic(err.Error())
        fmt.Println(" [x] JSON format error")
        return ""
    }

    if location, ok := data["location"].(map[string]interface{}); ok {
        current := data["current"].(map[string]interface{})

        name := location["name"].(string)
        country := location["country"].(string)
        temp_c := current["temp_c"].(float64)
        feelslike_c := current["feelslike_c"].(float64)

        // fmt.Printf("Results: %v\n", data)
        return fmt.Sprintf("Name: %s, Country:%s, Temperature:%.2f, Feeling of temperature:%.2f\n",
            name, country, temp_c, feelslike_c)
    }
    return ""
}

func tcpPipe(conn *net.TCPConn, thro <-chan time.Time) {
    ipStr := conn.RemoteAddr().String()
    defer func() {
        atomic.AddInt64(&connCounter, -1)
        fmt.Println("<- disconnected: " + ipStr)
        conn.Close()
    }()
    reader := bufio.NewReader(conn)

    loop:
    for {
        // Timeout mechanism
        conn.SetReadDeadline(time.Now().Add(30 * time.Second))

        message, _, err := reader.ReadLine()
        // message, err := reader.ReadString('\n')
        if err != nil {
            return
        }

        s := string(message)
        fmt.Printf(" [cmd] %s ... (%s)\n", s, ipStr)

        m := strings.Split(s, ",")
        resp := ""

        switch m[0] {
            case "quit":
                break loop
            case "help":
                resp = printHelp()
            case "status":
                resp = fmt.Sprintf("Current connection:%d, API request rate:%d\n",
                    atomic.LoadInt64(&connCounter),
                    atomic.LoadInt64(&reqCounter),
                )
            case "time":
                resp = time.Now().String() + "\n"
            case "weather":
                q := "Taipei"
                if len(m) > 1 && m[1] != "" {
                    q = m[1]
                }

                var Url *url.URL
                Url, err := url.Parse(WEATHER_URL)
                if err != nil {
                    panic(err.Error())
                }

                p := url.Values{}
                p.Add("q", q)
                p.Add("key", WEATHER_API_KEY)
                Url.RawQuery = p.Encode()

                <-thro
                atomic.AddInt64(&reqCounter, 1)
                resp = getWeatherURL(Url.String())
            default:
                fmt.Println(" [x] Unknown command: " + s)
                resp = printHelp()
        }

        if resp != "" {
            b := []byte(resp)
            conn.Write(b)
        }
    }
}

func main() {
    tick := time.NewTicker(RATE)
    defer tick.Stop()
    throttle := make(chan time.Time, BURST_LIMIT)

    go func() {
        for t := range tick.C {
            select {
                case throttle <- t:
                    // fmt.Println("time's up: ", t)
                default:
            }
        }
    }()

    go func() {
        http.HandleFunc("/status", func (w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "Current connection:%d, API request rate:%d",
                atomic.LoadInt64(&connCounter),
                atomic.LoadInt64(&reqCounter),
            )
        })

        http.ListenAndServe(HTTP_HOST + ":" + HTTP_PORT, nil)
    }()

    var tcpAddr *net.TCPAddr
    tcpAddr, _ = net.ResolveTCPAddr(CONN_TYPE, CONN_HOST + ":" + CONN_PORT)
    tcpListener, _ := net.ListenTCP(CONN_TYPE, tcpAddr)
    defer tcpListener.Close()

    for {
        tcpConn, err := tcpListener.AcceptTCP()
        if err != nil {
            continue
        }

        atomic.AddInt64(&connCounter, 1)
        fmt.Println("-> connected: " + tcpConn.RemoteAddr().String())
        go tcpPipe(tcpConn, throttle)
    }

}
