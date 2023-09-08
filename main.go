package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/chzyer/readline"
	"github.com/gorilla/websocket"
	"github.com/nxtrace/wscat-go/pow"
)

type IpInfo struct {
	IP          string  `json:"ip"`
	AsNumber    string  `json:"asnumber"`
	Country     string  `json:"country"`
	CountryEn   string  `json:"country_en"`
	CountryCode string  `json:"country_code"`
	Prov        string  `json:"prov"`
	ProvEn      string  `json:"prov_en"`
	City        string  `json:"city"`
	CityEn      string  `json:"city_en"`
	District    string  `json:"district"`
	Owner       string  `json:"owner"`
	ISP         string  `json:"isp"`
	Domain      string  `json:"domain"`
	Whois       string  `json:"whois"`
	Prefix      string  `json:"prefix"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Router      string  `json:"router"`
	Source      string  `json:"source"`
}

func main() {
	fmt.Println("PoW Start")
	jwtToken, err := pow.GetToken()

	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}

	requestHeader := http.Header{
		"Authorization": []string{"Bearer " + jwtToken},
	}

	c, _, err := websocket.DefaultDialer.Dial("wss://api.leo.moe/v3/ipGeoWs", requestHeader)
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer c.Close()

	fmt.Println("LeoMoeAPI V2 连接成功！")

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			<-ticker.C
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Println("发送心跳失败:", err)
				return
			}
		}
	}()

	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		err = c.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			fmt.Println("发送失败:", err)
			break
		}

		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("接收失败:", err)
			break
		}

		// var ipInfo IpInfo
		// err = json.Unmarshal(message, &ipInfo)
		// if err != nil {
		// 	fmt.Println("JSON解析失败:", err)
		// 	break
		// }

		// color.Cyan("ip: %s", ipInfo.IP)
		// color.Green("ASN: %s", ipInfo.AsNumber)
		// color.Yellow("Geo: %s %s %s", ipInfo.Country, ipInfo.Prov, ipInfo.City)
		// color.Magenta("Owner: %s", ipInfo.Owner)

		var ipObj map[string]interface{}
		json.Unmarshal([]byte(message), &ipObj)

		// New colorjson Formatter
		f := colorjson.NewFormatter()
		f.Indent = 2

		s, _ := f.Marshal(ipObj)
		fmt.Println(string(s))
	}
}
