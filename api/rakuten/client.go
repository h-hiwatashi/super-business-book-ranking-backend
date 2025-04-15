package rakuten

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type RakutenBookRankingResponse struct {
	Items []RakutenBookItem `json:"Items"`
	Count int               `json:"count"`
	Page  int               `json:"page"`
	First int               `json:"first"`
	Last  int               `json:"last"`
	Hits  int               `json:"hits"`
}

type RakutenBookItem struct {
	Item RakutenBook `json:"Item"`
}

type RakutenBook struct {
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	PublisherName string  `json:"publisherName"`
	ISBN          string  `json:"isbn"`
	ItemPrice     int     `json:"itemPrice"`
	ItemURL       string  `json:"itemUrl"`
	LargeImageURL string  `json:"largeImageUrl"`
	SalesDate     string  `json:"salesDate"`
	Rank          int     `json:"rank"`
	ReviewCount   int     `json:"reviewCount"`
	ReviewAverage float64 `json:"reviewAverage"`
	ItemCaption   string  `json:"itemCaption"`
}

type RakutenClient struct {
}

func NewRakutenClient() *RakutenClient {
	return &RakutenClient{}
}

func (c *RakutenClient) GetBookRanking(categoryID string, periodType string) (*RakutenBookRankingResponse, error) {
	mockData := generateMockBookRanking(categoryID, periodType)
	return mockData, nil
}

func (c *RakutenClient) GetBookRankingJSON(categoryID string, periodType string) (string, error) {
	ranking, err := c.GetBookRanking(categoryID, periodType)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(ranking, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSONエンコードエラー: %v", err)
	}

	return string(jsonData), nil
}

func generateMockBookRanking(categoryID string, periodType string) *RakutenBookRankingResponse {
	rand.Seed(time.Now().UnixNano())
	
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	
	if !validPeriods[periodType] {
		periodType = "daily" // デフォルト値
	}
	
	categoryNames := map[string]string{
		"001": "ビジネス書",
		"002": "自己啓発",
		"003": "マーケティング",
		"004": "経済・金融",
		"005": "IT・テクノロジー",
	}
	
	if _, exists := categoryNames[categoryID]; !exists {
		categoryID = "001"
	}
	
	items := []RakutenBookItem{}
	
	businessBooks := []RakutenBook{
		{
			Title:         "成功する習慣",
			Author:        "山田太郎",
			PublisherName: "ビジネス出版",
			ISBN:          "9784123456789",
			ItemPrice:     1650,
			ItemURL:       "https://books.rakuten.co.jp/mock/book1",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book1.jpg",
			SalesDate:     "2024-03-15",
			ReviewCount:   42,
			ReviewAverage: 4.5,
			ItemCaption:   "ビジネスで成功するための習慣について解説した一冊",
		},
		{
			Title:         "リーダーシップの極意",
			Author:        "佐藤次郎",
			PublisherName: "リーダーシップ社",
			ISBN:          "9784123456790",
			ItemPrice:     1980,
			ItemURL:       "https://books.rakuten.co.jp/mock/book2",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book2.jpg",
			SalesDate:     "2024-02-20",
			ReviewCount:   38,
			ReviewAverage: 4.2,
			ItemCaption:   "現代のリーダーに必要なスキルを解説",
		},
		{
			Title:         "効率的な時間管理術",
			Author:        "鈴木花子",
			PublisherName: "タイムマネジメント出版",
			ISBN:          "9784123456791",
			ItemPrice:     1540,
			ItemURL:       "https://books.rakuten.co.jp/mock/book3",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book3.jpg",
			SalesDate:     "2024-01-10",
			ReviewCount:   56,
			ReviewAverage: 4.7,
			ItemCaption:   "忙しいビジネスパーソンのための時間管理術",
		},
		{
			Title:         "ビジネス交渉術",
			Author:        "田中一郎",
			PublisherName: "ネゴシエーション社",
			ISBN:          "9784123456792",
			ItemPrice:     2200,
			ItemURL:       "https://books.rakuten.co.jp/mock/book4",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book4.jpg",
			SalesDate:     "2023-12-05",
			ReviewCount:   29,
			ReviewAverage: 4.0,
			ItemCaption:   "ビジネス交渉で成功するためのテクニック",
		},
		{
			Title:         "マインドフルネス経営",
			Author:        "高橋誠",
			PublisherName: "マインド出版",
			ISBN:          "9784123456793",
			ItemPrice:     1870,
			ItemURL:       "https://books.rakuten.co.jp/mock/book5",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book5.jpg",
			SalesDate:     "2023-11-20",
			ReviewCount:   45,
			ReviewAverage: 4.6,
			ItemCaption:   "マインドフルネスを取り入れた新しい経営手法",
		},
		{
			Title:         "デジタルトランスフォーメーション入門",
			Author:        "伊藤健太",
			PublisherName: "DX出版",
			ISBN:          "9784123456794",
			ItemPrice:     2420,
			ItemURL:       "https://books.rakuten.co.jp/mock/book6",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book6.jpg",
			SalesDate:     "2023-10-15",
			ReviewCount:   33,
			ReviewAverage: 4.3,
			ItemCaption:   "DXの基礎から応用までを解説",
		},
		{
			Title:         "起業家精神の育て方",
			Author:        "中村起業",
			PublisherName: "スタートアップ出版",
			ISBN:          "9784123456795",
			ItemPrice:     1760,
			ItemURL:       "https://books.rakuten.co.jp/mock/book7",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book7.jpg",
			SalesDate:     "2023-09-10",
			ReviewCount:   27,
			ReviewAverage: 4.1,
			ItemCaption:   "起業家に必要なマインドセットを解説",
		},
		{
			Title:         "財務諸表の読み方",
			Author:        "小林会計",
			PublisherName: "ファイナンス社",
			ISBN:          "9784123456796",
			ItemPrice:     2090,
			ItemURL:       "https://books.rakuten.co.jp/mock/book8",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book8.jpg",
			SalesDate:     "2023-08-20",
			ReviewCount:   41,
			ReviewAverage: 4.4,
			ItemCaption:   "初心者でもわかる財務諸表の読み方",
		},
		{
			Title:         "マーケティング戦略の立て方",
			Author:        "山本マーケ",
			PublisherName: "マーケティング出版",
			ISBN:          "9784123456797",
			ItemPrice:     1980,
			ItemURL:       "https://books.rakuten.co.jp/mock/book9",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book9.jpg",
			SalesDate:     "2023-07-15",
			ReviewCount:   36,
			ReviewAverage: 4.2,
			ItemCaption:   "効果的なマーケティング戦略の立て方",
		},
		{
			Title:         "チームビルディングの秘訣",
			Author:        "佐々木チーム",
			PublisherName: "チーム出版",
			ISBN:          "9784123456798",
			ItemPrice:     1870,
			ItemURL:       "https://books.rakuten.co.jp/mock/book10",
			LargeImageURL: "https://thumbnail.image.rakuten.co.jp/mock/book10.jpg",
			SalesDate:     "2023-06-10",
			ReviewCount:   31,
			ReviewAverage: 4.3,
			ItemCaption:   "強いチームを作るためのリーダーシップ論",
		},
	}
	
	switch periodType {
	case "daily":
	case "weekly":
		rand.Shuffle(len(businessBooks), func(i, j int) {
			if rand.Intn(3) == 0 { // 33%の確率で入れ替え
				businessBooks[i], businessBooks[j] = businessBooks[j], businessBooks[i]
			}
		})
	case "monthly":
		rand.Shuffle(len(businessBooks), func(i, j int) {
			if rand.Intn(2) == 0 { // 50%の確率で入れ替え
				businessBooks[i], businessBooks[j] = businessBooks[j], businessBooks[i]
			}
		})
	}
	
	for i, book := range businessBooks {
		book.Rank = i + 1
		items = append(items, RakutenBookItem{Item: book})
	}
	
	
	response := &RakutenBookRankingResponse{
		Items: items,
		Count: len(items),
		Page:  1,
		First: 1,
		Last:  len(items),
		Hits:  len(items),
	}
	
	return response
}
