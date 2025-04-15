// ranking/scheduler.go
package ranking

import (
	"log"
	"sync"
	"time"
)

// RankingScheduler はランキング更新を定期的に実行するためのスケジューラーである。
type RankingScheduler struct {
	service     *RankingService
	runningJobs map[string]struct{}
	interval    time.Duration
	mu          sync.Mutex
	stopCh      chan struct{}
}

// NewRankingScheduler は新しいランキングスケジューラーを生成する。
func NewRankingScheduler(service *RankingService, interval time.Duration) *RankingScheduler {
	return &RankingScheduler{
		service:     service,
		runningJobs: make(map[string]struct{}),
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

// Start はランキング更新スケジューラーを開始する。
func (s *RankingScheduler) Start() {
	log.Println("ランキング更新スケジューラーを開始します。")

	// 初回実行
	s.updateAllRankings()

	// 定期実行
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateAllRankings()
		case <-s.stopCh:
			log.Println("ランキング更新スケジューラーを停止します。")
			return
		}
	}
}

// Stop はランキング更新スケジューラーを停止する。
func (s *RankingScheduler) Stop() {
	close(s.stopCh)
}

// UpdateRanking は指定したジャンルと期間のランキングを更新する。
func (s *RankingScheduler) UpdateRanking(genre, periodType string) {
	jobID := genre + "-" + periodType

	s.mu.Lock()
	if _, exists := s.runningJobs[jobID]; exists {
		s.mu.Unlock()
		log.Printf("ジョブ %s は既に実行中です。", jobID)
		return
	}
	s.runningJobs[jobID] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.runningJobs, jobID)
		s.mu.Unlock()
	}()

	log.Printf("ランキング更新開始: %s - %s", genre, periodType)
	err := s.service.FetchAndStoreRakutenRanking(genre, periodType)
	if err != nil {
		log.Printf("ランキング更新エラー (%s - %s): %v", genre, periodType, err)
		return
	}
	log.Printf("ランキング更新完了: %s - %s", genre, periodType)
}

// updateAllRankings は全てのジャンルと期間のランキングを更新する。
func (s *RankingScheduler) updateAllRankings() {
	// 更新対象のジャンルリスト
	genres := []string{"business", "computer", "all"}

	// 更新対象の期間リスト
	periodTypes := []string{"daily", "weekly"}

	// 各ジャンルと期間の組み合わせでランキング更新を実行
	var wg sync.WaitGroup
	for _, genre := range genres {
		for _, periodType := range periodTypes {
			wg.Add(1)
			go func(g, p string) {
				defer wg.Done()
				s.UpdateRanking(g, p)
			}(genre, periodType)
		}
	}

	wg.Wait()
}