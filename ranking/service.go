// ranking/service.go
package ranking

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"strings"

	"github.com/google/uuid"

	"super-business-book-ranking-backend/rakuten"
)

// RankingService はランキングの取得と保存を管理するサービスである。
type RankingService struct {
	DB           *sql.DB
	RakutenClient *rakuten.RakutenClient
}

// NewRankingService は新しいランキングサービスを生成する。
func NewRankingService(db *sql.DB, rakutenClient *rakuten.RakutenClient) *RankingService {
	return &RankingService{
		DB:           db,
		RakutenClient: rakutenClient,
	}
}

// FetchAndStoreRakutenRanking は楽天APIから指定ジャンルの書籍ランキングを取得し、DBに保存する。
func (s *RankingService) FetchAndStoreRakutenRanking(genreName string, periodType string) error {
	// ジャンルIDの取得
	genreID, ok := rakuten.BookGenres[genreName]
	if !ok {
		return fmt.Errorf("無効なジャンル名: %s", genreName)
	}

	// 楽天APIパラメータの設定
	param := rakuten.RankingParam{
		Genre:  genreID,
		Period: periodType,
		Page:   1,
	}

	// ランキングデータの取得
	resp, err := s.RakutenClient.GetBookRanking(param)
	if err != nil {
		return fmt.Errorf("楽天ランキング取得エラー: %w", err)
	}

	// 取得した期間の設定
	now := time.Now()
	var dateFrom, dateTo time.Time

	switch periodType {
	case "daily":
		dateFrom = now.AddDate(0, 0, -1)
		dateTo = now
	case "weekly":
		dateFrom = now.AddDate(0, 0, -7)
		dateTo = now
	case "monthly":
		dateFrom = now.AddDate(0, -1, 0)
		dateTo = now
	default:
		dateFrom = now.AddDate(0, 0, -1)
		dateTo = now
	}

	// 楽天サイト情報の取得またはDB登録
	siteID, err := s.getOrCreateSite("rakuten", "https://books.rakuten.co.jp/", "")
	if err != nil {
		return fmt.Errorf("サイト情報の取得・登録エラー: %w", err)
	}

	// カテゴリ情報の取得または登録
	categoryName := fmt.Sprintf("楽天ブックス %s", genreName)
	categoryID, err := s.getOrCreateCategory(categoryName, nil)
	if err != nil {
		return fmt.Errorf("カテゴリの取得・登録エラー: %w", err)
	}

	// サイトカテゴリマッピングの取得または登録
	_, err = s.getOrCreateSiteCategoryMapping(categoryID, siteID, fmt.Sprintf("%d", genreID))
	if err != nil {
		return fmt.Errorf("サイトカテゴリマッピングの取得・登録エラー: %w", err)
	}

	// トランザクション開始
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("トランザクション開始エラー: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("パニックによるロールバック: %v", r)
		}
	}()

	// 各書籍の処理
	for _, item := range resp.Items {
		bookInfo := item.Item

		// 書籍情報の保存
		bookID, err := s.getOrCreateBook(tx, bookInfo)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("書籍情報の保存エラー: %w", err)
		}

		// 書籍サイトマッピングの保存
		bookSiteMappingID, err := s.getOrCreateBookSiteMapping(tx, bookID, siteID, bookInfo)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("書籍サイトマッピングの保存エラー: %w", err)
		}

		// ランキング情報の保存
		err = s.createRanking(tx, bookSiteMappingID, categoryID, bookInfo.Rank, periodType, dateFrom, dateTo)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ランキング情報の保存エラー: %w", err)
		}
	}

	// トランザクションのコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットエラー: %w", err)
	}

	return nil
}

// getOrCreateSite はサイト情報を取得または新規作成する。
func (s *RankingService) getOrCreateSite(name string, baseURL string, affiliateID string) (string, error) {
	var siteID string
	query := "SELECT id FROM sites WHERE name = ?"
	err := s.DB.QueryRow(query, name).Scan(&siteID)

	if err == nil {
		return siteID, nil
	}

	if err != sql.ErrNoRows {
		return "", err
	}

	// 新規サイト情報の作成
	siteID = uuid.New().String()
	query = "INSERT INTO sites (id, name, base_url, affiliate_id) VALUES (?, ?, ?, ?)"
	_, err = s.DB.Exec(query, siteID, name, baseURL, affiliateID)
	if err != nil {
		return "", err
	}

	return siteID, nil
}

// getOrCreateCategory はカテゴリを取得または新規作成する。
func (s *RankingService) getOrCreateCategory(name string, parentID *string) (string, error) {
	var categoryID string
	var query string
	var args []interface{}

	query = "SELECT id FROM categories WHERE name = ?"
	args = append(args, name)
	
	if parentID != nil {
		query += " AND parent_id = ?"
		args = append(args, *parentID)
	} else {
		query += " AND parent_id IS NULL"
	}
	
	err := s.DB.QueryRow(query, args...).Scan(&categoryID)

	if err == nil {
		return categoryID, nil
	}

	if err != sql.ErrNoRows {
		return "", err
	}

	// 新規カテゴリの作成
	categoryID = uuid.New().String()
	if parentID != nil {
		query = "INSERT INTO categories (id, name, parent_id) VALUES (?, ?, ?)"
		_, err = s.DB.Exec(query, categoryID, name, parentID)
	} else {
		query = "INSERT INTO categories (id, name) VALUES (?, ?)"
		_, err = s.DB.Exec(query, categoryID, name)
	}

	if err != nil {
		return "", err
	}

	return categoryID, nil
}

// getOrCreateSiteCategoryMapping はサイトカテゴリマッピングを取得または作成する。
func (s *RankingService) getOrCreateSiteCategoryMapping(categoryID, siteID, siteCategoryID string) (string, error) {
	var mappingID string
	query := "SELECT id FROM site_category_mappings WHERE category_id = ? AND site_id = ? AND site_specific_category_id = ?"
	err := s.DB.QueryRow(query, categoryID, siteID, siteCategoryID).Scan(&mappingID)

	if err == nil {
		return mappingID, nil
	}

	if err != sql.ErrNoRows {
		return "", err
	}

	// 新規マッピングの作成
	mappingID = uuid.New().String()
	query = "INSERT INTO site_category_mappings (id, category_id, site_id, site_specific_category_id) VALUES (?, ?, ?, ?)"
	_, err = s.DB.Exec(query, mappingID, categoryID, siteID, siteCategoryID)
	if err != nil {
		return "", err
	}

	return mappingID, nil
}

// getOrCreateBook は書籍情報を取得または作成する。
func (s *RankingService) getOrCreateBook(tx *sql.Tx, bookInfo rakuten.ItemInfo) (string, error) {
	// ISBNがある場合はISBNで検索
	if bookInfo.ISBN != "" {
		var bookID string
		query := "SELECT id FROM books WHERE isbn = ?"
		err := tx.QueryRow(query, bookInfo.ISBN).Scan(&bookID)
		if err == nil {
			return bookID, nil
		}
		if err != sql.ErrNoRows {
			return "", err
		}
	}

	// タイトルと著者で検索
	if bookInfo.ItemName != "" && bookInfo.Author != "" {
		var bookID string
		query := "SELECT id FROM books WHERE title = ? AND author = ?"
		err := tx.QueryRow(query, bookInfo.ItemName, bookInfo.Author).Scan(&bookID)
		if err == nil {
			return bookID, nil
		}
		if err != sql.ErrNoRows {
			return "", err
		}
	}

	// 新規書籍情報の作成
	bookID := uuid.New().String()
	pubDate := sql.NullString{String: "", Valid: false}
	imageURL := ""
	
	if len(bookInfo.MediumImageURLs) > 0 {
		imageURL = bookInfo.MediumImageURLs[0].ImageURL
	}

	query := `
		INSERT INTO books 
		(id, title, author, publisher, isbn, publication_date, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := tx.Exec(
		query, 
		bookID, 
		bookInfo.ItemName, 
		bookInfo.Author, 
		bookInfo.PublisherName,
		bookInfo.ISBN,
		pubDate,
		imageURL,
	)
	if err != nil {
		return "", err
	}

	return bookID, nil
}

// getOrCreateBookSiteMapping は書籍サイトマッピングを取得または作成する。
func (s *RankingService) getOrCreateBookSiteMapping(tx *sql.Tx, bookID, siteID string, bookInfo rakuten.ItemInfo) (string, error) {
	var mappingID string
	query := "SELECT id FROM book_site_mappings WHERE book_id = ? AND site_id = ? AND site_specific_id = ?"
	err := tx.QueryRow(query, bookID, siteID, bookInfo.ItemCode).Scan(&mappingID)
	
	if err == nil {
		// 既存のマッピングが見つかった場合は価格とURLを更新
		updateQuery := "UPDATE book_site_mappings SET price = ?, url = ? WHERE id = ?"
		_, err = tx.Exec(updateQuery, float64(bookInfo.ItemPrice), bookInfo.ItemURL, mappingID)
		if err != nil {
			return "", err
		}
		return mappingID, nil
	}
	
	if err != sql.ErrNoRows {
		return "", err
	}

	// 新規マッピングの作成
	mappingID = uuid.New().String()
	insertQuery := "INSERT INTO book_site_mappings (id, book_id, site_id, site_specific_id, price, url) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(insertQuery, mappingID, bookID, siteID, bookInfo.ItemCode, float64(bookInfo.ItemPrice), bookInfo.ItemURL)
	if err != nil {
		return "", err
	}

	return mappingID, nil
}

// createRanking はランキング情報を作成する。
func (s *RankingService) createRanking(
	tx *sql.Tx, 
	bookSiteMappingID string, 
	categoryID string, 
	rank int, 
	periodType string, 
	dateFrom time.Time, 
	dateTo time.Time,
) error {
	// 同じ条件のランキングが既にあるか確認（当日の同じジャンル、同じ期間タイプ）
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM rankings 
		WHERE book_site_mapping_id = ? AND category_id = ? AND period_type = ? 
		AND date_from = ? AND date_to = ?
	`
	err := tx.QueryRow(
		checkQuery, 
		bookSiteMappingID, 
		categoryID, 
		periodType, 
		dateFrom.Format("2006-01-02"), 
		dateTo.Format("2006-01-02"),
	).Scan(&count)
	
	if err != nil {
		return err
	}
	
	if count > 0 {
		// 既に存在する場合はランクを更新
		updateQuery := "UPDATE rankings SET rank = ? WHERE book_site_mapping_id = ? AND category_id = ? AND period_type = ? AND date_from = ? AND date_to = ?"
		_, err = tx.Exec(
			updateQuery, 
			rank, 
			bookSiteMappingID, 
			categoryID, 
			periodType, 
			dateFrom.Format("2006-01-02"), 
			dateTo.Format("2006-01-02"),
		)
		return err
	}
	
	// 新規ランキングの作成
	rankingID := uuid.New().String()
	insertQuery := `
		INSERT INTO rankings 
		(id, book_site_mapping_id, category_id, rank, period_type, date_from, date_to) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(
		insertQuery, 
		rankingID, 
		bookSiteMappingID, 
		categoryID, 
		rank, 
		periodType, 
		dateFrom.Format("2006-01-02"), 
		dateTo.Format("2006-01-02"),
	)
	
	return err
}