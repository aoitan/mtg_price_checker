/* vim:set ts=2 sw=2 sts=2 fdm=indent: */
// Copyright 2019 September Soft
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "time"
  "net/url"
  "crypto/sha256"
  "crypto/hmac"
  "encoding/hex"
  "io/ioutil"
  "strconv"
  "github.com/gorilla/mux"
  "strings"
  "github.com/PuerkitoBio/goquery"
  "github.com/thoas/go-funk"
)

func main() {
    r := mux.NewRouter()
  r.HandleFunc("/", indexHandler)
  r.HandleFunc("/v1/price/summary/{cardname}", priceSummaryHandler).Methods("GET")
  r.HandleFunc("/v1/price/shop/{cardname}", priceShopHandler).Methods("GET")
  http.Handle("/", r)

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
    log.Printf("Defaulting to port %s\n", port)
  }

  log.Printf("Listening on port %s\n", port)
  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func toHmac(message, key string) string {
  mac := hmac.New(sha256.New, []byte(key))
  mac.Write([]byte(message))
  return hex.EncodeToString(mac.Sum(nil))
}

// indexHandler responds to requests with our greeting.
func indexHandler(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path != "/" {
    http.NotFound(w, r)
    return
  }

  name := "name=" + url.QueryEscape("宝石鉱山")
  api_key := "api_key=" + url.QueryEscape("testuser") // ToDo: 環境変数から読む
  stock_gt := "stock_gt=1"
  timestamp := "timestamp=" + url.QueryEscape(strconv.FormatInt(time.Now().Unix(), 10))
  message := api_key + "\n" + name + "\n" + stock_gt + "\n" + timestamp
  fmt.Println(message)
  api_secret := "0123456789" // ToDo: 環境変数から読む
  api_sig := "api_sig=" + toHmac(message, api_secret)
  fmt.Println(api_sig)
  query_string := api_key + "&" + name + "&" + stock_gt + "&" + timestamp + "&" + api_sig
  fmt.Println(query_string)
  url := "http://wonder.wisdom-guild.net/api/card-price/v1/?" + query_string
  resp, _ := http.Get(url)
  defer resp.Body.Close()
  bytes, _ := ioutil.ReadAll(resp.Body)

  fmt.Fprint(w, string(bytes))
}

func priceSummaryHandler(w http.ResponseWriter, r *http.Request) {
  if strings.Index(r.URL.Path, "/v1/price/summary") == -1 {
    http.NotFound(w, r)
    return
  }

  vars := mux.Vars(r)

  if vars["cardname"] == "" {
    fmt.Println("cardname is empty")
    http.NotFound(w, r)
    return
  }

  name := "name=" + url.QueryEscape(vars["cardname"])
  api_key := "api_key=" + url.QueryEscape("testuser") // ToDo: 環境変数から読む
  stock_gt := "stock_gt=1"
  timestamp := "timestamp=" + url.QueryEscape(strconv.FormatInt(time.Now().Unix(), 10))
  message := api_key + "\n" + name + "\n" + stock_gt + "\n" + timestamp
  fmt.Println(message)
  api_secret := "0123456789" // ToDo: 環境変数から読む
  api_sig := "api_sig=" + toHmac(message, api_secret)
  fmt.Println(api_sig)
  query_string := api_key + "&" + name + "&" + stock_gt + "&" + timestamp + "&" + api_sig
  fmt.Println(query_string)
  url := "http://wonder.wisdom-guild.net/api/card-price/v1/?" + query_string
  resp, _ := http.Get(url)
  defer resp.Body.Close()
  bytes, _ := ioutil.ReadAll(resp.Body)

  fmt.Fprint(w, string(bytes))
}

type ShopPrice struct {
  CardName string
  ShopName string
  Price string
  Set string
  Lang string
  Stock string
  State string
  LastModified time.Time
  LastCheck time.Time
}

func priceShopHandler(w http.ResponseWriter, r *http.Request) {
  if strings.Index(r.URL.Path, "/v1/price/shop") == -1 {
    http.NotFound(w, r)
    return
  }

  vars := mux.Vars(r)

  if vars["cardname"] == "" {
    fmt.Println("cardname is empty")
    http.NotFound(w, r)
    return
  }

  name := "card=" + url.QueryEscape(vars["cardname"])
  checktime_gt := "checktime_gt=24h"
  mode := "mode=shop"
  stock_gt := "stock_gt=1"
  state_gt := "state_gt=EX"
  query_string := name + "&" + checktime_gt + "&" + mode + "&" + state_gt + "&" + stock_gt
  url := "http://wonder.wisdom-guild.net/search.php?" + query_string
  doc, err := goquery.NewDocument(url)
  if err != nil {
    fmt.Println("failed get")
    http.NotFound(w, r)
    return
  }

  // テーブルを配列に変換する
  raw_data := make([]ShopPrice, 0)
  doc.Find(".table-main tbody tr").Each(func (i int, tr *goquery.Selection) {
    //fmt.Printf("  tr[%d]: %s\n", i, tr.Text())
    price := ShopPrice{}
    tr.Find("td").Each(func (ii int, td *goquery.Selection) {
      switch ii {
      case 0:
        price.ShopName = td.Find(".shopname").Text()
        fmt.Println("    td[shopname]: " + price.ShopName)
        price.CardName = td.Find(".cardname").Text()
        fmt.Println("    td[cardname]: " + price.CardName)
      case 1:
        price.Price = strings.Replace(td.Find("strong").Text(), ",", "", -1)
        fmt.Println("    td[price]: " + price.Price)
      case 2:
        price.Set = td.Text()
        fmt.Println("    td[set]: " + price.Set)
      case 3:
        price.Lang = td.Text()
        fmt.Println("    td[lang]: " + price.Lang)
      case 4:
        price.Stock = td.Text()
        fmt.Println("    td[stock]: " + price.Stock)
      case 5:
        // skip
      case 6:
        price.State = td.Text()
        fmt.Println("    td[state]: " + price.State)
      case 7:
        // skip
      case 8:
        lastmodified := td.Text()
        price.LastModified, _ = time.Parse("01/12/23 01:23", lastmodified)
        fmt.Println("    td[lastmodified]: " + price.LastModified.String())
      case 9:
        lastcheck := td.Text()
        price.LastCheck, _ = time.Parse("01/12/23 01:23", lastcheck)
        fmt.Println("    td[lastcheck]: " + price.LastCheck.String())
      }
    })
    raw_data = append(raw_data, price)
  })

  // 配列から特定ショップをフィルタ
  filtered, _ := funk.Filter(raw_data, func (price ShopPrice) bool {
    return price.ShopName == "晴れる屋" || price.ShopName == "Cardshop Serra" || price.ShopName == "カードラッシュ" || price.ShopName == "トレトク" || price.ShopName == "ENNDAL GAMES" || price.ShopName == "ドラゴンスター"
  }).([]ShopPrice)

  // ショップ毎にユニークにする
  encount := make(map[string]bool)
  unique := make(map[string]ShopPrice, 0)
  for _, price := range filtered {
    if !encount[price.ShopName] {
      encount[price.ShopName] = true
      unique[price.ShopName] = price
    }
  }

  // カード名,晴れる屋,Cardshop Serra,カードラッシュ,トレトク,ENNDAL GAMES,ドラゴンスターのCSVにする
  output := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", vars["cardname"], unique["晴れる屋"].Price, unique["Cardshop Serra"].Price, unique["カードラッシュ"].Price, unique["トレトク"].Price, unique["ENNDAL GAMES"].Price, unique["ドラゴンスター"].Price)

  fmt.Println(output)
  fmt.Println(query_string)
  fmt.Fprint(w, output)
}

// [END indexHandler]
// [END gae_go111_app]

