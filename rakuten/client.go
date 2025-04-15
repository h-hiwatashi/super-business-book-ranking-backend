// rakuten/client.go
package rakuten

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// APIのエンドポイントURL
const (
	RankingAPIEndpoint = "https://app.rakuten.co.jp/services/api/IchibaItem/Ranking/20170628"
)

// RakutenClient は楽天APIへのアクセスを管理する構造体である。
type RakutenClient struct {
	ApplicationID string
	HTTPClient    *http.Client
}

// NewClient は新しい楽天APIクライアントを生成する。
// applicationIDが指定されていない場合は環境変数から取得する。
func NewClient(applicationID string) *RakutenClient {
	if applicationID == "" {
		applicationID = os.Getenv("RAKUTEN_APPLICATION_ID")
		if applicationID == "" {
			panic("楽天アプリケーションIDが設定されていません。")
		}
	}

	return &RakutenClient{
		ApplicationID: applicationID,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

// RankingParam は楽天商品ランキングAPIのパラメータを表す。
type RankingParam struct {
	Genre       int    // ジャンルID (必須)
	Page        int    // ページ番号 (任意, デフォルト: 1)
	CacheUpdate bool   // キャッシュ更新フラグ (任意)
	Period      string // 集計期間 (任意, デフォルト: "daily")
}

// RankingResponse は楽天商品ランキングAPIのレスポンスを表す。
type RankingResponse struct {
	Items      []RankingItem `json:"Items"`
	PageCount  int           `json:"pageCount"`
	LastBuild  string        `json:"lastBuild"`
	TotalCount int           `json:"count"`
}

// RankingItem はランキング内の商品を表す。
type RankingItem struct {
	Item ItemInfo `json:"Item"`
}

// ItemInfo は商品の詳細情報を表す。
type ItemInfo struct {
	ItemName        string  `json:"itemName"`
	ItemCode        string  `json:"itemCode"`
	ItemPrice       int     `json:"itemPrice"`
	ItemCaption     string  `json:"itemCaption"`
	ItemURL         string  `json:"itemUrl"`
	ShopName        string  `json:"shopName"`
	MediumImageURLs []Image `json:"mediumImageUrls"`
	Rank            int     `json:"rank"`
	GenreID         string  `json:"genreId"`
	// 書籍関連の追加情報
	Author          string `json:"author"`
	PublisherName   string `json:"publisherName"`
	ISBN            string `json:"isbn"`
	Size            string `json:"size"`
	// 楽天ブックス用に拡張できる項目を必要に応じて追加
}

// Image は商品画像情報を表す。
type Image struct {
	ImageURL string `json:"imageUrl"`
}

// GetBookRanking は指定されたパラメータで書籍のランキングを取得する。
// ジャンルIDは書籍関連のものを指定する必要がある。
func (c *RakutenClient) GetBookRanking(param RankingParam) (*RankingResponse, error) {
	// 楽天ブックスのトップレベルジャンルID
	// 200162: 本・雑誌・コミック
	if param.Genre == 0 {
		param.Genre = 200162
	}

	// URLクエリパラメータの構築
	values := url.Values{}
	values.Set("applicationId", c.ApplicationID)
	values.Set("format", "json")
	values.Set("genreId", strconv.Itoa(param.Genre))

	// オプションパラメータ
	if param.Page > 0 {
		values.Set("page", strconv.Itoa(param.Page))
	}
	if param.CacheUpdate {
		values.Set("formatVersion", "2")
	}
	if param.Period != "" {
		values.Set("period", param.Period)
	}

	// リクエストURLの構築
	reqURL := fmt.Sprintf("%s?%s", RankingAPIEndpoint, values.Encode())

	// HTTP GETリクエストの実行
	resp, err := c.HTTPClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("APIリクエストエラー: %w", err)
	}
	defer resp.Body.Close()

	// エラーレスポンスの確認
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("APIレスポンスエラー: %s", resp.Status)
	}

	// レスポンスのパース
	var result RankingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("レスポンスのパースエラー: %w", err)
	}

	return &result, nil
}

// 書籍のジャンルIDマッピング（楽天の書籍ジャンル）
var BookGenres = map[string]int{
	"all":         200162, // 本・雑誌・コミック
	"books":       200163, // 本
	"magazine":    200164, // 雑誌
	"comics":      200165, // コミック
	"lightNovel":  200302, // ライトノベル
	"business":    200166, // ビジネス・経済・就職
	"computer":    200167, // コンピュータ・IT
	"science":     200168, // 科学・医学・技術
	"humanities":  200169, // 人文・思想
	"entertainment": 200170, // エンタメ・ゲーム
}