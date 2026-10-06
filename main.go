package main

// 使うパッケージ（道具）を読み込む
import (
	"database/sql"  // データベースを操作するための道具
	"encoding/json" // JSONを読み書きするための道具
	"errors"        // エラーの種類を調べるための道具
	"fmt"           // 文字を出力するための道具
	"log"           // ログ（記録）を出すための道具
	"net/http"      // Webサーバーを作るための道具
	"os"            // 環境変数を読むための道具
	"strconv"       // 文字列を数値に変換するための道具
	"strings"       // 文字列の前後の空白を取るための道具
	"time"          // 日付の形式をチェックするための道具

	"github.com/joho/godotenv" // .env ファイルを読み込むための道具
	"github.com/lib/pq"        // PostgreSQL ドライバ（エラーの種類を調べるのにも使う）
)

// Category は categories テーブルの1行を表す
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Item は shopping_items テーブルの1行を表す
type Item struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	CategoryID   int    `json:"category_id"`
	PurchaseDate string `json:"purchase_date"` // "2026-10-05" の形
	IsPurchased  bool   `json:"is_purchased"`
}

// itemColumns は Item を読み出すときに使う列（日付は "YYYY-MM-DD" の文字にして取り出す）
const itemColumns = "id, name, category_id, to_char(purchase_date, 'YYYY-MM-DD'), is_purchased"

// db はアプリ全体で使うデータベース接続
var db *sql.DB

// writeJSON はデータをJSONにしてブラウザへ返す
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError はエラーメッセージをJSONで返す
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// scanItem は1行分のデータを Item に読み込む
func scanItem(row interface{ Scan(...any) error }) (Item, error) {
	var it Item
	err := row.Scan(&it.ID, &it.Name, &it.CategoryID, &it.PurchaseDate, &it.IsPurchased)
	return it, err
}

// getCategories は GET /api/categories : カテゴリー一覧を返す
func getCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name FROM categories ORDER BY id")
	if err != nil {
		log.Println("カテゴリー取得エラー:", err)
		writeError(w, http.StatusInternalServerError, "カテゴリーの取得に失敗しました")
		return
	}
	defer rows.Close()

	// 0件のときも null ではなく [] を返すため、空のスライスで始める
	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			log.Println("読み取りエラー:", err)
			writeError(w, http.StatusInternalServerError, "カテゴリーの取得に失敗しました")
			return
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		log.Println("カテゴリー取得エラー:", err)
		writeError(w, http.StatusInternalServerError, "カテゴリーの取得に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

// getItems は GET /api/items : 買い物リストを返す
func getItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT " + itemColumns + " FROM shopping_items ORDER BY purchase_date, id")
	if err != nil {
		log.Println("一覧取得エラー:", err)
		writeError(w, http.StatusInternalServerError, "買い物リストの取得に失敗しました")
		return
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			log.Println("読み取りエラー:", err)
			writeError(w, http.StatusInternalServerError, "買い物リストの取得に失敗しました")
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		log.Println("一覧取得エラー:", err)
		writeError(w, http.StatusInternalServerError, "買い物リストの取得に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// createItem は POST /api/items : 商品を追加する
// リクエスト例: {"name": "牛乳", "category_id": 1, "purchase_date": "2026-10-05"}
func createItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name         string `json:"name"`
		CategoryID   int    `json:"category_id"`
		PurchaseDate string `json:"purchase_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "リクエストの形式が正しくありません")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.CategoryID <= 0 {
		writeError(w, http.StatusBadRequest, "name と category_id は必須です")
		return
	}
	// 日付が "2026-10-05" の形になっているかチェックする
	if _, err := time.Parse("2006-01-02", input.PurchaseDate); err != nil {
		writeError(w, http.StatusBadRequest, "purchase_date は 2026-10-05 の形で指定してください")
		return
	}

	// 追加した行を RETURNING でそのまま受け取る
	// is_purchased は書かないので、初期値の false が入る
	it, err := scanItem(db.QueryRow(
		"INSERT INTO shopping_items (name, category_id, purchase_date) VALUES ($1, $2, $3) RETURNING "+itemColumns,
		input.Name, input.CategoryID, input.PurchaseDate,
	))
	if err != nil {
		// 23503 = 外部キー違反（存在しないカテゴリーIDが指定された）
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			writeError(w, http.StatusBadRequest, "指定されたカテゴリーは存在しません")
			return
		}
		log.Println("追加エラー:", err)
		writeError(w, http.StatusInternalServerError, "商品の追加に失敗しました")
		return
	}

	writeJSON(w, http.StatusCreated, it)
}

// updateItem は PATCH /api/items/{id} : 購入済みかどうかを切り替える
// リクエスト例: {"is_purchased": true}
func updateItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "IDが正しくありません")
		return
	}

	// 送られなかったことを判別するため、ポインタで受け取る
	var input struct {
		IsPurchased *bool `json:"is_purchased"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "リクエストの形式が正しくありません")
		return
	}
	if input.IsPurchased == nil {
		writeError(w, http.StatusBadRequest, "is_purchased は必須です")
		return
	}

	it, err := scanItem(db.QueryRow(
		"UPDATE shopping_items SET is_purchased = $1 WHERE id = $2 RETURNING "+itemColumns,
		*input.IsPurchased, id,
	))
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "商品が見つかりません")
		return
	}
	if err != nil {
		log.Println("更新エラー:", err)
		writeError(w, http.StatusInternalServerError, "商品の更新に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, it)
}

// deleteItem は DELETE /api/items/{id} : 商品を削除する
func deleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "IDが正しくありません")
		return
	}

	result, err := db.Exec("DELETE FROM shopping_items WHERE id = $1", id)
	if err != nil {
		log.Println("削除エラー:", err)
		writeError(w, http.StatusInternalServerError, "商品の削除に失敗しました")
		return
	}
	// 削除された行が0件なら、そのIDは存在しない
	if n, _ := result.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "商品が見つかりません")
		return
	}

	// 削除成功：返す中身はないので 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// allowCORS は、HTMLを直接開いたとき（別の場所の画面）からでもAPIを使えるようにする
// ブラウザは、別の場所からのリクエストを、サーバーが許可したときだけ通す
func allowCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 本番のリクエストの前に、ブラウザが「送っていい？」と確認してくる（OPTIONS）ので、OKと返す
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// .env ファイルを読み込んで環境変数にセットする
	if err := godotenv.Load(); err != nil {
		log.Fatal(".env ファイルの読み込みに失敗しました: ", err)
	}

	// 接続情報は .env の DATABASE_URL から取得する（コードには直接書かない）
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL が設定されていません")
	}

	// データベースに接続する
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("データベースの設定に失敗しました: ", err)
	}
	defer db.Close()

	// 本当に接続できるか確認する
	if err := db.Ping(); err != nil {
		log.Fatal("データベースに接続できません: ", err)
	}
	fmt.Println("データベースに接続しました")

	// API：「メソッド + パス」ごとに処理する関数を登録する
	http.HandleFunc("GET /api/categories", getCategories)
	http.HandleFunc("GET /api/items", getItems)
	http.HandleFunc("POST /api/items", createItem)
	http.HandleFunc("PATCH /api/items/{id}", updateItem)
	http.HandleFunc("DELETE /api/items/{id}", deleteItem)

	// 画面：static フォルダの中のファイル（HTML・CSS・JS）をそのまま返す
	http.Handle("/", http.FileServer(http.Dir("static")))

	// サーバーを起動したことをターミナルに表示する
	fmt.Println("サーバーを起動しました: http://localhost:8080")

	// 8080番ポートでサーバーを起動する（エラーが起きたら終了する）
	log.Fatal(http.ListenAndServe(":8080", allowCORS(http.DefaultServeMux)))
}
