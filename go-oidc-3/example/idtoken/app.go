/*
This is an example application to demonstrate parsing an ID Token.
*/
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<body>

<h2>RDAP Search</h2>

<form action="/search">
  <input type="text" id="query" name="query">
  <input type="submit" value="Submit">
</form> 

</body>
</html>
`

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) { //設定cookie
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}
func convertNewlinesToHTML(input string) string { //將\n轉成<br>
	return strings.ReplaceAll(input, "\n", "<br>")
}
func main() {
	godotenv.Load() //讀取.env檔案
	var (
		clientID     = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID") //設定clientID
		clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
	)

	ctx := context.Background() //設定context

	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com") //設定provider
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &oidc.Config{ //設定oidc的config
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig) //設定verifier

	config := oauth2.Config{ //設定oauth2的config
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://127.0.0.1:5556/auth/google/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	// 創建一個身份對應表
	identityMap := map[string]string{
		"115199660478099487343": "registrar",
		"default":               "guest",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { //首頁
		state, err := randString(16)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		nonce, err := randString(16)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		setCallbackCookie(w, r, "state", state) //將state和nonce存入cookie
		setCallbackCookie(w, r, "nonce", nonce) //將state和nonce存入cookie

		http.Redirect(w, r, config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound) // 重定向用戶到 OAuth 2.0 授權服務器，並在 URL 中包含 "state" 和 "nonce" 參數
	})

	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) { //callback頁面 r代表請求  w代表請求的回應
		tmpl, err := template.New("searchForm").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		state, err := r.Cookie("state") // 從 HTTP 請求的 cookie 中讀取名為 "state" 的 cookie
		if err != nil {
			http.Error(w, "state not found", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("state") != state.Value { //檢查state是否相符
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code")) //使用授權碼來取得token
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError) //檢查token是否存在
			return
		}
		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token)) //獲得userinfo
		if err != nil {
			http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string) //取得id_token
		if !ok {
			http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError) //檢查id_token是否存在
			return
		}
		idToken, err := verifier.Verify(ctx, rawIDToken) //驗證id_token
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError) //檢查id_token是否正確
			return
		}

		nonce, err := r.Cookie("nonce") //檢查nonce是否存在
		if err != nil {
			http.Error(w, "nonce not found", http.StatusBadRequest) //檢查nonce是否存在
			return
		}
		if idToken.Nonce != nonce.Value { //檢查nonce是否相符
			http.Error(w, "nonce did not match", http.StatusBadRequest) //檢查nonce是否相符
			return
		}

		// oauth2Token.AccessToken = "*REDACTED*" //隱藏access_token

		resp := struct { //回傳資料 resp = response
			OAuth2Token   *oauth2.Token //隱藏token
			UserInfo      *oidc.UserInfo
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, userInfo, new(json.RawMessage)} //設定回傳資料
		//fmt.Printf("%s\n", resp)
		resp2 := struct { //回傳資料 resp = response
			OAuth2Token *oauth2.Token //隱藏token
		}{oauth2Token} //設定回傳資料
		data2, err := json.MarshalIndent(resp2, "", "    ") //將resp轉成json格式
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) //檢查resp是否正確
			return
		}
		//把data2傳送到127.0.0.1:8080
		url2 := "http://127.0.0.1:8080"
		abcd, err := http.Post(url2, "application/json", bytes.NewBuffer(data2))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer abcd.Body.Close() //關閉連線

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil { //將id_token的payload存入resp.IDTokenClaims
			http.Error(w, err.Error(), http.StatusInternalServerError) //檢查id_token是否正確
			return
		}

		data, err := json.MarshalIndent(resp, "", "    ") //將resp轉成json格式
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) //檢查resp是否正確
			return
		}
		stringdata := convertNewlinesToHTML(string(data))
		// 假設 resp.IDTokenClaims 是一個 map[string]interface{}
		claims := make(map[string]interface{})
		err = json.Unmarshal(*resp.IDTokenClaims, &claims)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 從 IDTokenClaims 獲取 sub 值
		sub := claims["sub"].(string)
		// 從 sub 值獲取對應的身份
		// 從身份對應表中獲取身份
		identity, ok := identityMap[sub]
		if !ok {
			// 如果 sub 值在身份對應表中不存在，則使用默認身份
			identity = identityMap["default"]
		}
		w.Write([]byte("Your identity is " + identity + "<br>")) //將level回傳 顯示在網頁上

		w.Write([]byte(stringdata)) //將resp回傳 顯示在網頁上

		//把data傳送到127.0.0.1:8080
		url := "http://127.0.0.1:8080"
		abc, err := http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer abc.Body.Close() //關閉連線

		// 將發送的內容寫入到一個檔案中
		err = os.WriteFile("sender.txt", data, 0644) //寫入檔案
		if err != nil {
			http.Error(w, "Error writing to file", http.StatusInternalServerError)
			return
		}

	})
	http.HandleFunc("/s", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.New("searchForm").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		// 顯示查詢表單以便進行下一次查詢
		tmpl, err := template.New("searchForm").Parse(htmlTemplate) //將htmlTemplate轉成html格式
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		query := r.URL.Query().Get("query")
		// 這裡處理查詢邏輯
		w.Write([]byte("You searched for: " + query + "<br>"))

		urlsearch := "http://127.0.0.1:8080"
		search, err := http.Post(urlsearch, "text/plain", bytes.NewBuffer([]byte(query)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer search.Body.Close() //關閉連線

		// 讀取回應體
		body, err := io.ReadAll(search.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// re := regexp.MustCompile(`:\s.*`) //正則表達式
		// redactedBody := re.ReplaceAllString(string(body), ": *REDACTED*")

		// // 將響應體寫入 HTTP 響應
		// w.Write([]byte(redactedBody))

		stringbody := convertNewlinesToHTML(string(body)) //將\n轉成<br>
		w.Write([]byte(stringbody))                       //將body回傳 顯示在網頁上
		//w.Write(body)
	})

	log.Printf("listening on http://%s/", "127.0.0.1:5556") //顯示網址
	log.Fatal(http.ListenAndServe("127.0.0.1:5556", nil))   //開啟網頁
}
