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
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/sheets/v4"
  "encoding/json"
)

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/", indexHandler)
  r.HandleFunc("/v1/price/summary/{cardname}", priceSummaryHandler).Methods("GET")
  r.HandleFunc("/v1/price/shop/{cardname}", priceShopHandler).Methods("GET")
  r.HandleFunc("/oauth/callback", oauthCallbackHandler)
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

func priceShopGetDataFromWisdomGuild(cardname string) (*goquery.Document, error) {
  name := "card=" + url.QueryEscape(cardname)
  checktime_gt := "checktime_gt=24h"
  mode := "mode=shop"
  stock_gt := "stock_gt=1"
  state_gt := "state_gt=EX"
  query_string := name + "&" + checktime_gt + "&" + mode + "&" + state_gt + "&" + stock_gt
  url := "http://wonder.wisdom-guild.net/search.php?" + query_string
  fmt.Println(query_string)
  return goquery.NewDocument(url)
}

func priceShopWebTableToArray(doc *goquery.Document) []ShopPrice {
  raw_data := make([]ShopPrice, 0)
  doc.Find(".table-main tbody tr").Each(func (i int, tr *goquery.Selection) {
    //fmt.Printf("  tr[%d]: %s\n", i, tr.Text())
    price := ShopPrice{}
    tr.Find("td").Each(func (ii int, td *goquery.Selection) {
      switch ii {
      case 0:
        price.ShopName = td.Find(".shopname").Text()
        //fmt.Println("    td[shopname]: " + price.ShopName)
        price.CardName = td.Find(".cardname").Text()
        //fmt.Println("    td[cardname]: " + price.CardName)
      case 1:
        price.Price = strings.Replace(td.Find("strong").Text(), ",", "", -1)
        //fmt.Println("    td[price]: " + price.Price)
      case 2:
        price.Set = td.Text()
        //fmt.Println("    td[set]: " + price.Set)
      case 3:
        price.Lang = td.Text()
        //fmt.Println("    td[lang]: " + price.Lang)
      case 4:
        price.Stock = td.Text()
        //fmt.Println("    td[stock]: " + price.Stock)
      case 5:
        // skip
      case 6:
        price.State = td.Text()
        //fmt.Println("    td[state]: " + price.State)
      case 7:
        // skip
      case 8:
        lastmodified := td.Text()
        price.LastModified, _ = time.Parse("01/12/23 01:23", lastmodified)
        //fmt.Println("    td[lastmodified]: " + price.LastModified.String())
      case 9:
        lastcheck := td.Text()
        price.LastCheck, _ = time.Parse("01/12/23 01:23", lastcheck)
        //fmt.Println("    td[lastcheck]: " + price.LastCheck.String())
      }
    })
    raw_data = append(raw_data, price)
  })

  return raw_data
}

// ToDo: 自前DB -> Wisdom Guildの順で問い合わせる
// ToDo: Wisdom Guildへの問い合わせをメソッド化する
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

  // Wisdom Guildへ問い合わせ
  doc, err := priceShopGetDataFromWisdomGuild(vars["cardname"])
  if err != nil {
    fmt.Println("failed get")
    http.NotFound(w, r)
    return
  }

  // テーブルを配列に変換する
  raw_data := priceShopWebTableToArray(doc)

  // 配列から特定ショップをフィルタ
  // ToDo: 対応するショップはリスト化してリストに含まれているかで判定する
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

  // カード名,晴れる屋,Cardshop Serra,カードラッシュ,トレトク,ENNDAL GAMES,ドラゴンスターのTSVにする
  output := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", vars["cardname"], unique["晴れる屋"].Price, unique["Cardshop Serra"].Price, unique["カードラッシュ"].Price, unique["トレトク"].Price, unique["ENNDAL GAMES"].Price, unique["ドラゴンスター"].Price)

  fmt.Println(output)

  w.Header().Set("Content-Type", "text/plain")
  w.WriteHeader(http.StatusOK)
  fmt.Fprint(w, output)
}

func oauthHandler() {
  b, err := ioutil.ReadFile("credentials.json")
  if err != nil {
    log.Fatalf("Unable to read client secret file: %v", err)
  }

  // If modifying these scopes, delete your previously saved token.json.
  config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
  if err != nil {
    log.Fatalf("Unable to parse client secret file to config: %v", err)
  }
  client := getClient(config)

  srv, err := sheets.New(client)
  if err != nil {
    log.Fatalf("Unable to retrieve Sheets client: %v", err)
  }

  // Prints the names and majors of students in a sample spreadsheet:
  // https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
  spreadsheetId := "1TtZf6VrEKpKVCxlkkEF0YQxGu8-BxAygassZ7WqIo3E"
  readRange := "input!A1:E"
  resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
  if err != nil {
    log.Fatalf("Unable to retrieve data from sheet: %v", err)
  }

  if len(resp.Values) == 0 {
    fmt.Println("No data found.")
  } else {
    fmt.Println("Name, Major:")
    for _, row := range resp.Values {
      // Print columns A and E, which correspond to indices 0 and 4.
      fmt.Printf("%s, %s\n", row[0], row[4])
    }
  }
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
  // The file token.json stores the user's access and refresh tokens, and is
  // created automatically when the authorization flow completes for the first
  // time.
  tokFile := "token.json"
  tok, err := tokenFromFile(tokFile)
  if err != nil {
    tok = getTokenFromWeb(config)
    saveToken(tokFile, tok)
  }
  return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  fmt.Printf("Go to the following link in your browser then type the "+
  "authorization code: \n%v\n", authURL)

  var authCode string
  if _, err := fmt.Scan(&authCode); err != nil {
    log.Fatalf("Unable to read authorization code: %v", err)
  }

  tok, err := config.Exchange(context.TODO(), authCode)
  if err != nil {
    log.Fatalf("Unable to retrieve token from web: %v", err)
  }
  return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
  f, err := os.Open(file)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  tok := &oauth2.Token{}
  err = json.NewDecoder(f).Decode(tok)
  return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
  fmt.Printf("Saving credential file to: %s\n", path)
  f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
  if err != nil {
    log.Fatalf("Unable to cache oauth token: %v", err)
  }
  defer f.Close()
  json.NewEncoder(f).Encode(token)
}

func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
}
// [END indexHandler]
// [END gae_go111_app]

