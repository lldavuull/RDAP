package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background() //設定context
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost { //如果是post就執行
			body, err := io.ReadAll(r.Body) //讀取body
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			// 這裡你可以處理 body 的內容
			// fmt.Println("Received: ", string(body))

			// 將收到的內容寫入到一個檔案中
			err = os.WriteFile("received.txt", body, 0644) //寫入檔案
			if err != nil {
				http.Error(w, "Error writing to file", http.StatusInternalServerError)
				return
			}

			provider, err := oidc.NewProvider(ctx, "https://accounts.google.com") //設定provider
			if err != nil {
				http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
				return
			}
			//將收到的json轉回Token
			var resp2 struct {
				OAuth2Token *oauth2.Token
			}
			json.Unmarshal(body, &resp2)
			// 現在，resp2.OAuth2Token 包含了解析後的 token
			token := resp2.OAuth2Token
			userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(token)) //獲得userinfo
			if err != nil {
				http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
				return
			}
			resp := struct { //回傳資料 resp = response
				UserInfo *oidc.UserInfo
			}{userInfo} //設定回傳資料
			data, err := json.MarshalIndent(resp, "", "    ") //將resp轉成json格式
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError) //檢查resp是否正確
				return
			}
			w.Write(data) //將resp回傳 顯示在網頁上
			// 將收到的內容寫入到一個檔案中
			err = os.WriteFile("received_userinfo.txt", data, 0644) //寫入檔案
			if err != nil {
				http.Error(w, "Error writing to file", http.StatusInternalServerError)
				return
			}
		}

		if r.Method == http.MethodGet { //如果是get就執行
			// 讀取檔案的內容
			content, err := os.ReadFile("received.txt")
			if err != nil {
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}

			// 將檔案的內容寫入到 HTTP 響應中
			fmt.Fprint(w, string(content))

			// 讀取檔案的內容
			content_userinfo, err := os.ReadFile("received_userinfo.txt")
			if err != nil {
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}
			// 將檔案的內容寫入到 HTTP 響應中
			fmt.Fprint(w, string(content_userinfo))
			return
		}

		if r.Method != http.MethodPost { //如果不是post就回傳錯誤
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token)) //獲得userinfo
		// if err != nil {
		// 	http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		// 	return
		// }

	})
	log.Fatal(http.ListenAndServe("127.0.0.1:5557", nil))
}
