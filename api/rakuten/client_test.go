package rakuten

import (
	"encoding/json"
	"testing"
)

func TestNewRakutenClient(t *testing.T) {
	client := NewRakutenClient()
	if client == nil {
		t.Error("NewRakutenClient() returned nil")
	}
}

func TestGetBookRanking(t *testing.T) {
	client := NewRakutenClient()
	
	testCases := []struct {
		name       string
		categoryID string
		periodType string
		wantCount  int
	}{
		{
			name:       "ビジネス書カテゴリ（デフォルト）",
			categoryID: "001",
			periodType: "daily",
			wantCount:  10,
		},
		{
			name:       "存在しないカテゴリID（デフォルトに変換される）",
			categoryID: "999",
			periodType: "daily",
			wantCount:  10,
		},
		{
			name:       "週間ランキング",
			categoryID: "001",
			periodType: "weekly",
			wantCount:  10,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.GetBookRanking(tc.categoryID, tc.periodType)
			
			if err != nil {
				t.Errorf("GetBookRanking() error = %v", err)
				return
			}
			
			if result == nil {
				t.Error("GetBookRanking() returned nil")
				return
			}
			
			if len(result.Items) != tc.wantCount {
				t.Errorf("GetBookRanking() returned %d items, want %d", len(result.Items), tc.wantCount)
			}
			
			for i, item := range result.Items {
				if item.Item.Rank != i+1 {
					t.Errorf("Item %d has rank %d, want %d", i, item.Item.Rank, i+1)
				}
				
				if item.Item.Title == "" {
					t.Errorf("Item %d has empty title", i)
				}
				
				if item.Item.Author == "" {
					t.Errorf("Item %d has empty author", i)
				}
				
				if item.Item.ISBN == "" {
					t.Errorf("Item %d has empty ISBN", i)
				}
				
				if item.Item.ItemURL == "" {
					t.Errorf("Item %d has empty URL", i)
				}
			}
		})
	}
}

func TestGetBookRankingJSON(t *testing.T) {
	client := NewRakutenClient()
	
	jsonStr, err := client.GetBookRankingJSON("001", "daily")
	
	if err != nil {
		t.Errorf("GetBookRankingJSON() error = %v", err)
		return
	}
	
	if jsonStr == "" {
		t.Error("GetBookRankingJSON() returned empty string")
		return
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Errorf("GetBookRankingJSON() returned invalid JSON: %v", err)
	}
	
	items, ok := result["Items"].([]interface{})
	if !ok {
		t.Error("GetBookRankingJSON() JSON does not contain Items array")
		return
	}
	
	if len(items) == 0 {
		t.Error("GetBookRankingJSON() returned empty Items array")
	}
}
