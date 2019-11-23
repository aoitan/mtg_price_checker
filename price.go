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
	"crypto/hmac"
	"encode/hex"
	"io/ioutil"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/v1/price", priceHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func toHmac(message, key string) string {
    mac := hmac.New(sha256.New, key)
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
}

// indexHandler responds to requests with our greeting.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	name := url.QueryEscape("name=" + "宝石鉱山")
	api_key := "api_key=" + "testuser" // ToDo: 環境変数から読む
	timestamp := "timestamp=" + time.Now().Unix()
	api_secret := "0123456789" // ToDo: 環境変数から読む
	message := api_key + "\n" + name + "\n" + timestamp
	api_sig := "api_sig" + toHmac(message, api_secret)
	query_string := api_key + "&" + name + "&" + timestamp + "&" + api_sig
	url := "http://wonder.wisdom-guild.net/api/card-price/v1/?" + query_string
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	bytes := ioutil.ReadAll(resp.Body)

	fmt.Fprint(w, string(bytes))
}

func priceHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/price" {
		http.NotFound(w, r)
		return
	}
	w.Header().set("Content-Type", "application/json")
	w.Write(w, "{status:200}")
}

// [END indexHandler]
// [END gae_go111_app]

