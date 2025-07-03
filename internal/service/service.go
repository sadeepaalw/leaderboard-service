package service

import (
	"context"
	"errors"
	"leaderboard-service/internal/model"
	"leaderboard-service/internal/repository"
	"log"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo repository.RepositoryInterface
}

func NewService(repo repository.RepositoryInterface) *Service {
	return &Service{repo: repo}
}

func (s *Service) StartMatchmakingWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		log.Println("[MatchmakingWorker] Started")
		for {
			select {
			case <-ctx.Done():
				log.Println("[MatchmakingWorker] Stopped")
				return
			case <-ticker.C:
				s.runMatchmaking(ctx)
			}
		}
	}()
}

func (s *Service) runMatchmaking(ctx context.Context) {
	waitingPlayers, err := s.repo.GetWaitingPlayers(ctx)
	if err != nil {
		log.Printf("[MatchmakingWorker] Error fetching waiting players: %v", err)
		return
	}
	if len(waitingPlayers) == 0 {
		log.Println("[MatchmakingWorker] No players waiting")
		return
	}
	group := waitingPlayers
	if len(group) > 10 {
		group = group[:10]
	}
	log.Printf("[MatchmakingWorker] Found %d waiting players, grouping %d", len(waitingPlayers), len(group))

	compID := uuid.New()
	now := time.Now()
	endsAt := now.Add(60 * time.Minute)
	comp := &model.Competition{
		CompetitionID: compID,
		StartedAt:     now,
		EndsAt:        endsAt,
		Level:         group[0].Level,
		CountryCode:   group[0].CountryCode,
	}
	if err := s.repo.CreateCompetition(ctx, comp); err != nil {
		log.Printf("[MatchmakingWorker] Error creating competition: %v", err)
		return
	}
	playerIDs := make([]string, len(group))
	for i, p := range group {
		playerIDs[i] = p.PlayerID
	}
	if err := s.repo.UpdatePlayerCompetitionsToActive(ctx, playerIDs, compID, endsAt); err != nil {
		log.Printf("[MatchmakingWorker] Error updating player competitions: %v", err)
		return
	}
	log.Printf("[MatchmakingWorker] Started competition %s with players: %v", compID.String(), playerIDs)
}

func (s *Service) Join(ctx context.Context, playerID string) (string, error) {
	log.Printf("[Service] Player %s attempting to join matchmaking", playerID)
	player, err := s.repo.GetPlayerByID(ctx, playerID)
	if err != nil {
		log.Printf("[Service] Player %s not found", playerID)
		return "", errors.New("player not found")
	}
	_, err = s.repo.GetActivePlayerCompetition(ctx, playerID)
	if err == nil {
		log.Printf("[Service] Player %s already in active competition", playerID)
		return "", errors.New("player already in active competition")
	}
	pc := &model.PlayerCompetition{
		PlayerID:      playerID,
		CompetitionID: nil,
		Status:        model.StatusWaiting,
		Score:         0,
		JoinedAt:      time.Now(),
		UpdatedAt:     time.Now(),
		Level:         player.Level,
		CountryCode:   player.CountryCode,
	}
	err = s.repo.CreatePlayerCompetition(ctx, pc)
	if err != nil {
		log.Printf("[Service] Error adding player %s to matchmaking queue: %v", playerID, err)
		return "", err
	}
	log.Printf("[Service] Player %s added to matchmaking queue", playerID)
	return "", nil
}

func (s *Service) GetPlayerLeaderboard(ctx context.Context, playerID string) (interface{}, error) {
	log.Printf("[Service] Fetching leaderboard for player %s", playerID)
	pc, err := s.repo.GetLatestPlayerCompetition(ctx, playerID)
	if err != nil {
		log.Printf("[Service] No competition found for player %s", playerID)
		return map[string]interface{}{}, nil
	}
	if pc.CompetitionID == nil {
		log.Printf("[Service] No competition ID for player %s", playerID)
		return map[string]interface{}{}, nil
	}
	leaderboard, err := s.repo.GetLeaderboardByCompetitionID(ctx, pc.CompetitionID.String())
	if err != nil {
		log.Printf("[Service] Error fetching leaderboard for competition %v: %v", pc.CompetitionID, err)
		return nil, err
	}
	entries := make([]map[string]interface{}, 0, len(leaderboard))
	for _, entry := range leaderboard {
		entries = append(entries, map[string]interface{}{
			"player_id": entry.PlayerID,
			"score":     entry.Score,
		})
	}
	log.Printf("[Service] Returning leaderboard for competition %v", pc.CompetitionID)
	return map[string]interface{}{
		"leaderboard_id": pc.CompetitionID.String(),
		"ends_at":        pc.UpdatedAt.Unix(),
		"leaderboard":    entries,
	}, nil
}

func (s *Service) GetLeaderboard(ctx context.Context, leaderboardID string) (interface{}, error) {
	log.Printf("[Service] Fetching leaderboard for competition %s", leaderboardID)
	pcs, err := s.repo.GetLeaderboardByCompetitionID(ctx, leaderboardID)
	if err != nil || len(pcs) == 0 {
		log.Printf("[Service] No leaderboard found for competition %s", leaderboardID)
		return nil, errors.New("leaderboard not found")
	}
	entries := make([]map[string]interface{}, 0, len(pcs))
	for _, entry := range pcs {
		entries = append(entries, map[string]interface{}{
			"player_id": entry.PlayerID,
			"score":     entry.Score,
		})
	}
	return map[string]interface{}{
		"leaderboard_id": leaderboardID,
		"leaderboard":    entries,
	}, nil
}

func (s *Service) SubmitScore(ctx context.Context, playerID string, score int) error {
	log.Printf("[Service] Submitting score for player %s", playerID)
	pc, err := s.repo.GetActivePlayerCompetition(ctx, playerID)
	if err != nil {
		log.Printf("[Service] Player %s not in active competition", playerID)
		return errors.New("player not in active competition")
	}
	err = s.repo.AddScoreToPlayer(ctx, playerID, score)
	if err != nil {
		log.Printf("[Service] Error adding score for player %s: %v", playerID, err)
		return err
	}
	log.Printf("[Service] Score %d added to player %s in competition %v", score, playerID, pc.CompetitionID)
	return nil
}

func (s *Service) CreatePlayer(ctx context.Context, playerID string, level int, countryCode string) error {
	player := &model.Player{
		PlayerID:    playerID,
		Level:       level,
		CountryCode: countryCode,
	}
	return s.repo.CreatePlayer(ctx, player)
}

func (s *Service) GetPlayer(ctx context.Context, playerID string) (*model.Player, error) {
	return s.repo.GetPlayerByID(ctx, playerID)
}

func (s *Service) UpdatePlayer(ctx context.Context, playerID string, level int, countryCode string) error {
	player, err := s.repo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return err
	}
	player.Level = level
	player.CountryCode = countryCode
	return s.repo.UpdatePlayer(ctx, player)
}

// // Repository interface for dependency injection
// // (to be implemented in internal/repository)
// type Repository interface {
// 	// Define methods as needed for DB access
// 	GetLatestPlayerCompetition(ctx context.Context, playerID string) (*model.PlayerCompetition, error)
// 	GetLeaderboardByCompetitionID(ctx context.Context, competitionID string) ([]model.PlayerCompetition, error)
// 	GetPlayerByID(ctx context.Context, playerID string) (*model.Player, error)
// 	GetActivePlayerCompetition(ctx context.Context, playerID string) (*model.PlayerCompetition, error)
// 	CreatePlayerCompetition(ctx context.Context, playerCompetition *model.PlayerCompetition) error
// 	CreatePlayer(ctx context.Context, player *model.Player) error
// 	UpdatePlayer(ctx context.Context, player *model.Player) error
// 	AddScoreToPlayer(ctx context.Context, playerID string, score int) error
// }
