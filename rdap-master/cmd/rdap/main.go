package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/openrdap/rdap"
	"golang.org/x/oauth2"
)

type responseWriter struct {
	http.ResponseWriter
	body bytes.Buffer
}

func redactResponse(response string) string {
	// 創建一個正則表達式來匹配冒號後的內容
	re := regexp.MustCompile(`:.*`)
	// 將冒號後的內容替換為 *REDACTED*
	redactedBody := re.ReplaceAllString(response, ": *REDACTED*")
	return redactedBody
}
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}
func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost { //如果是post就執行
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			ctx := context.Background()     //設定context
			body, err := io.ReadAll(r.Body) //讀取body
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
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
		case "text/plain":
			body, err := io.ReadAll(r.Body) //讀取body
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}
			// Convert body to []string
			args := []string{string(body)}
			// 創建一個新的 responseWriter 來捕獲回應
			rw := &responseWriter{ResponseWriter: w}

			// 執行 RDAP 查詢並將結果寫入 HTTP 回應
			exitCode := rdap.RunCLI(args, rw, rw, rdap.CLIOptions{})
			// 如果 RDAP 查詢失敗，則返回一個錯誤
			if exitCode != 0 {
				http.Error(w, "RDAP command failed", http.StatusInternalServerError)
			}

			redactedBody := redactResponse(rw.body.String()) //將回應內容轉換為*REDACTED*

			// 將修改後的回應寫入原始的 w
			w.Write([]byte(redactedBody))
		default:
			http.Error(w, "Unsupported content type", http.StatusBadRequest)
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
}

func main() {
	// 註冊 HTTP 處理器函數
	http.HandleFunc("/", handler)
	// 啟動 HTTP 伺服器
	http.ListenAndServe(":8080", nil)
}

// func main() {
// 	exitCode := rdap.RunCLI(os.Args[1:], os.Stdout, os.Stderr, rdap.CLIOptions{}) //os.Stdout, os.Stderr, rdap.CLIOptions{})

// 	os.Exit(exitCode)
// }
