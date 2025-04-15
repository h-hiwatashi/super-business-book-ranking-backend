package rakuten

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestNewRakutenHandler(t *testing.T) {
	handler := NewRakutenHandler()
	if handler == nil {
		t.Error("NewRakutenHandler() returned nil")
	}
	
	if handler.Client == nil {
		t.Error("NewRakutenHandler() returned handler with nil Client")
	}
}

func TestGetRakutenBookRankingHandler(t *testing.T) {
	testCases := []struct {
		name           string
		categoryID     string
		periodType     string
		wantStatusCode int
	}{
		{
			name:           "正常系：ビジネス書カテゴリ、日次ランキング",
			categoryID:     "001",
			periodType:     "daily",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "正常系：存在しないカテゴリID（デフォルトに変換される）",
			categoryID:     "999",
			periodType:     "daily",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "正常系：期間パラメータなし（デフォルト値が使用される）",
			categoryID:     "001",
			periodType:     "",
			wantStatusCode: http.StatusOK,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/rakuten/rankings/"+tc.categoryID, nil)
			if err != nil {
				t.Fatal(err)
			}
			
			if tc.periodType != "" {
				q := req.URL.Query()
				q.Add("period", tc.periodType)
				req.URL.RawQuery = q.Encode()
			}
			
			rr := httptest.NewRecorder()
			
			router := mux.NewRouter()
			handler := NewRakutenHandler()
			router.HandleFunc("/api/rakuten/rankings/{categoryId}", handler.GetRakutenBookRankingHandler).Methods("GET")
			
			router.ServeHTTP(rr, req)
			
			if status := rr.Code; status != tc.wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.wantStatusCode)
			}
			
			contentType := rr.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("handler returned wrong content type: got %v want application/json", contentType)
			}
			
			var response RakutenBookRankingResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Errorf("handler returned invalid JSON: %v", err)
			}
			
			if len(response.Items) == 0 {
				t.Error("handler returned empty Items array")
			}
			
			for i, item := range response.Items {
				if item.Item.Rank != i+1 {
					t.Errorf("Item %d has rank %d, want %d", i, item.Item.Rank, i+1)
				}
				
				if item.Item.Title == "" {
					t.Errorf("Item %d has empty title", i)
				}
				
				if item.Item.Author == "" {
					t.Errorf("Item %d has empty author", i)
				}
			}
		})
	}
}
