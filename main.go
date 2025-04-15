package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// 環境変数の設定とデフォルト値
var (
	port     = getEnv("PORT", "8080")
	dbHost   = getEnv("DB_HOST", "localhost")
	dbPort   = getEnv("DB_PORT", "3306")
	dbUser   = getEnv("DB_USER", "root")
	dbPass   = getEnv("DB_PASS", "password")
	dbName   = getEnv("DB_NAME", "book_ranking")
	dbParams = getEnv("DB_PARAMS", "parseTime=true&loc=Asia%2FTokyo")
)

// ヘルスチェックレスポンス
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Time    string `json:"time"`
}

// ランキング書籍情報
type RankedBook struct {
	ID             string    `json:"id"`
	Rank           int       `json:"rank"`
	Title          string    `json:"title"`
	Author         string    `json:"author"`
	Publisher      string    `json:"publisher"`
	ISBN           string    `json:"isbn"`
	PublicationDate string   `json:"publicationDate"`
	ImageURL       string    `json:"imageUrl"`
	Price          float64   `json:"price"`
	URL            string    `json:"url"`
}

// ランキングリスト
type RankingResponse struct {
	CategoryID   string       `json:"categoryId"`
	CategoryName string       `json:"categoryName"`
	PeriodType   string       `json:"periodType"`
	DateFrom     string       `json:"dateFrom"`
	DateTo       string       `json:"dateTo"`
	Books        []RankedBook `json:"books"`
}

// データベース接続用のグローバル変数
var db *sql.DB

func main() {

	// データベース接続
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", 
		dbUser, dbPass, dbHost, dbPort, dbName, dbParams)
	
	var err error
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("データベース接続エラー: %v", err)
	}
	defer db.Close()

	// 接続テスト
	err = db.Ping()
	if err != nil {
		log.Fatalf("データベースPingエラー: %v", err)
	}
	log.Println("データベース接続成功")

	// ルーターの設定
	r := mux.NewRouter()
	
	// APIエンドポイント
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")
	r.HandleFunc("/api/rankings/{categoryId}", getRankingsHandler).Methods("GET")
	r.HandleFunc("/api/books/{bookId}", getBookDetailsHandler).Methods("GET")
	r.HandleFunc("/api/categories", getCategoriesHandler).Methods("GET")

	// サーバー起動
	log.Printf("サーバーを起動しています。ポート: %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// 環境変数を取得（デフォルト値付き）
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// ヘルスチェック用ハンドラー
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
		Time:    time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ランキング取得ハンドラー
func getRankingsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["categoryId"]
	
	// クエリパラメータ
	periodType := r.URL.Query().Get("period")
	if periodType == "" {
		periodType = "daily" // デフォルト値
	}
	
	limit := 10 // デフォルト取得数
	
	// ランキングデータ取得のSQLクエリ
	query := `
		SELECT 
			r.rank, b.id, b.title, b.author, b.publisher, 
			b.isbn, b.publication_date, b.image_url, 
			bsm.price, bsm.url, c.id, c.name, 
			r.period_type, r.date_from, r.date_to
		FROM rankings r
		JOIN book_site_mappings bsm ON r.book_site_mapping_id = bsm.id
		JOIN books b ON bsm.book_id = b.id
		JOIN categories c ON r.category_id = c.id
		WHERE r.category_id = ? AND r.period_type = ?
		ORDER BY r.rank
		LIMIT ?
	`
	
	rows, err := db.Query(query, categoryID, periodType, limit)
	if err != nil {
		http.Error(w, "データベースクエリエラー", http.StatusInternalServerError)
		log.Printf("クエリエラー: %v", err)
		return
	}
	defer rows.Close()
	
	var books []RankedBook
	var response RankingResponse
	
	for rows.Next() {
		var book RankedBook
		var categoryID, categoryName, periodType, dateFrom, dateTo string
		var publicationDate sql.NullString
		
		err := rows.Scan(
			&book.Rank, &book.ID, &book.Title, &book.Author, &book.Publisher,
			&book.ISBN, &publicationDate, &book.ImageURL,
			&book.Price, &book.URL, &categoryID, &categoryName,
			&periodType, &dateFrom, &dateTo,
		)
		
		if err != nil {
			http.Error(w, "データスキャンエラー", http.StatusInternalServerError)
			log.Printf("スキャンエラー: %v", err)
			return
		}
		
		if publicationDate.Valid {
			book.PublicationDate = publicationDate.String
		}
		
		books = append(books, book)
		
		// 最初の行からカテゴリ情報などを設定
		if response.CategoryID == "" {
			response.CategoryID = categoryID
			response.CategoryName = categoryName
			response.PeriodType = periodType
			response.DateFrom = dateFrom
			response.DateTo = dateTo
		}
	}
	
	response.Books = books
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 書籍詳細取得ハンドラー
func getBookDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["bookId"]
	
	query := `
		SELECT 
			b.id, b.title, b.author, b.publisher, 
			b.isbn, b.publication_date, b.image_url
		FROM books b
		WHERE b.id = ?
	`
	
	var book RankedBook
	var publicationDate sql.NullString
	
	err := db.QueryRow(query, bookID).Scan(
		&book.ID, &book.Title, &book.Author, &book.Publisher,
		&book.ISBN, &publicationDate, &book.ImageURL,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "書籍が見つかりません", http.StatusNotFound)
		} else {
			http.Error(w, "データベースクエリエラー", http.StatusInternalServerError)
			log.Printf("クエリエラー: %v", err)
		}
		return
	}
	
	if publicationDate.Valid {
		book.PublicationDate = publicationDate.String
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

// カテゴリ一覧取得ハンドラー
func getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, name, parent_id
		FROM categories
		ORDER BY name
	`
	
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "データベースクエリエラー", http.StatusInternalServerError)
		log.Printf("クエリエラー: %v", err)
		return
	}
	defer rows.Close()
	
	type Category struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		ParentID *string `json:"parentId"`
	}
	
	var categories []Category
	
	for rows.Next() {
		var category Category
		var parentID sql.NullString
		
		err := rows.Scan(&category.ID, &category.Name, &parentID)
		if err != nil {
			http.Error(w, "データスキャンエラー", http.StatusInternalServerError)
			log.Printf("スキャンエラー: %v", err)
			return
		}
		
		if parentID.Valid {
			category.ParentID = &parentID.String
		}
		
		categories = append(categories, category)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
