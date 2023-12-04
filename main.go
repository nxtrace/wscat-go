package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/nxtrace/NTrace-core/util"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
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
	var fastIp string
	host, port := util.GetHostAndPort()
	// 如果 host 是一个 IP 使用默认域名
	if valid := net.ParseIP(host); valid != nil {
		fastIp = host
		host = "origin-fallback.nxtrace.org"
	} else {
		// 默认配置完成，开始寻找最优 IP
		fastIp = util.GetFastIP(host, port, true)
	}
	jwtToken, ua := util.EnvToken, []string{"Privileged Client"}
	err := error(nil)
	if jwtToken == "" {
		if util.GetPowProvider() == "" {
			jwtToken, err = pow.GetToken(fastIp, host, port)
		} else {
			jwtToken, err = pow.GetToken(util.GetPowProvider(), util.GetPowProvider(), port)
		}
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		ua = []string{"wscat-go"}
	}
	fmt.Println("PoW Start")

	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}

	requestHeader := http.Header{
		"Authorization": []string{"Bearer " + jwtToken},
		"Host":          []string{host},
		"User-Agent":    ua,
	}
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		ServerName: host,
	}
	proxyUrl := util.GetProxy()
	if proxyUrl != nil {
		dialer.Proxy = http.ProxyURL(proxyUrl)
	}
	u := url.URL{Scheme: "wss", Host: fastIp + ":" + port, Path: "/v3/ipGeoWs"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), requestHeader)
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(c)

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
	defer func(rl *readline.Instance) {
		err := rl.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(rl)

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
		err = json.Unmarshal(message, &ipObj)
		if err != nil {
			fmt.Println("JSON解析失败:", err)
		}

		// New colorjson Formatter
		f := colorjson.NewFormatter()
		f.Indent = 2

		s, _ := f.Marshal(ipObj)
		fmt.Println(string(s))
	}
}
