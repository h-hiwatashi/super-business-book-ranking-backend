package rakuten

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type RakutenHandler struct {
	Client *RakutenClient
}

func NewRakutenHandler() *RakutenHandler {
	return &RakutenHandler{
		Client: NewRakutenClient(),
	}
}

func (h *RakutenHandler) GetRakutenBookRankingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID := vars["categoryId"]
	
	periodType := r.URL.Query().Get("period")
	if periodType == "" {
		periodType = "daily" // デフォルト値
	}
	
	ranking, err := h.Client.GetBookRanking(categoryID, periodType)
	if err != nil {
		http.Error(w, "楽天APIエラー: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ranking)
}
