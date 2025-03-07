package main

// 必要なパッケージのインポート
import (
	"crypto/rand"     // 暗号学的に安全な乱数を生成するためのパッケージ
	"encoding/base64" // バイト列を文字列にエンコードするためのパッケージ
	"fmt"             // フォーマット付きの出力のためのパッケージ
	"html/template"   // HTMLテンプレートを扱うためのパッケージ
	"log"             // ログ出力のためのパッケージ
	"net/http"        // HTTPサーバーとクライアントのためのパッケージ
	"strings"         // 文字列操作のためのパッケージ
)

// URLの保存用のマップ
// キー: 短縮URL（例: "abc123"）
// 値: 元のURL（例: "https://example.com"）
var urlMap = make(map[string]string)

// 短縮URLを生成する関数
// 6バイトのランダムなデータを生成し、それをbase64エンコードして6文字の文字列を作成
func generateShortURL() string {
	b := make([]byte, 6)                            // 6バイトのスライスを作成
	rand.Read(b)                                    // ランダムなデータを生成
	return base64.URLEncoding.EncodeToString(b)[:6] // base64エンコードして最初の6文字を返す
}

// メインページのHTMLテンプレート
// {{if .ShortURL}} は、ShortURLが空でない場合にその部分を表示する条件分岐
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>URL短縮サービス</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .form-group { margin-bottom: 20px; }
        input[type="text"] { width: 100%; padding: 8px; margin: 5px 0; }
        button { padding: 10px 20px; background-color: #4CAF50; color: white; border: none; cursor: pointer; }
        .result { margin-top: 20px; padding: 10px; background-color: #f0f0f0; }
    </style>
</head>
<body>
    <h1>URL短縮サービス</h1>
    <form method="POST" action="/shorten">
        <div class="form-group">
            <label for="url">URLを入力してください：</label>
            <input type="text" id="url" name="url" required>
        </div>
        <button type="submit">短縮URLを生成</button>
    </form>
    {{if .ShortURL}}
    <div class="result">
        <p>短縮URL: <a href="{{.ShortURL}}">{{.ShortURL}}</a></p>
    </div>
    {{end}}
</body>
</html>
`

// テンプレートに渡すデータの構造体
type PageData struct {
	ShortURL string // 生成された短縮URLを保持
}

func main() {
	// HTMLテンプレートを解析
	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	// メインページのハンドラー（GETリクエスト用）
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			data := PageData{}    // 空のデータを作成
			tmpl.Execute(w, data) // テンプレートを実行してHTMLを生成
		}
	})

	// URL短縮処理のハンドラー（POSTリクエスト用）
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			url := r.FormValue("url") // フォームからURLを取得
			if url == "" {
				http.Error(w, "URLを入力してください", http.StatusBadRequest)
				return
			}

			// URLの形式チェック（http:// または https:// が付いていない場合は追加）
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				url = "https://" + url
			}

			// 短縮URLを生成して保存
			shortURL := generateShortURL()
			urlMap[shortURL] = url

			// 結果ページを表示
			data := PageData{
				ShortURL: fmt.Sprintf("http://localhost:8080/%s", shortURL),
			}
			tmpl.Execute(w, data)
		}
	})

	// 短縮URLへのアクセス処理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/") // URLから先頭の/を除去
		if path != "" {
			// 短縮URLが存在する場合は元のURLにリダイレクト
			if originalURL, exists := urlMap[path]; exists {
				http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
				return
			}
			http.Error(w, "URLが見つかりません", http.StatusNotFound)
		}
	})

	// サーバーを起動（ポート8080で待ち受け）
	fmt.Println("サーバーを起動します: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
