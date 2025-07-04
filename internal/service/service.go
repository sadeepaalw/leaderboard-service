package service

import (
	"context"
	"errors"
	"fmt"
	"leaderboard-service/internal/model"
	"leaderboard-service/internal/repository"
	"log"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	MatchmakingInterval time.Duration
	CompetitionDuration time.Duration
}

type Service struct {
	repo   repository.RepositoryInterface
	config Config
}

type ServiceInterface interface {
	Join(ctx context.Context, playerID string) (string, error)
	SubmitScore(ctx context.Context, playerID string, score int) error
	GetPlayerLeaderboard(ctx context.Context, playerID string) (interface{}, error)
	GetLeaderboard(ctx context.Context, leaderboardID string) (interface{}, error)
	CreatePlayer(ctx context.Context, playerID string, level int, countryCode string) error
	GetPlayer(ctx context.Context, playerID string) (*model.Player, error)
	UpdatePlayer(ctx context.Context, playerID string, level int, countryCode string) error
}

func NewService(repo repository.RepositoryInterface, config Config) *Service {
	return &Service{repo: repo, config: config}
}

func (s *Service) StartMatchmakingWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(s.config.MatchmakingInterval)
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
	// 1. Mark finished competitions as COMPLETED
	if err := s.repo.CompleteFinishedCompetitions(ctx); err != nil {
		log.Printf("[MatchmakingWorker] Error completing finished competitions: %v", err)
	}

	// Check for existing active competition
	activeComp, err := s.repo.GetActiveCompetition(ctx)
	if err == nil && activeComp != nil {
		log.Printf("[MatchmakingWorker] Active competition %s already exists, skipping creation", activeComp.CompetitionID.String())
		return
	}

	// 2. Fetch all waiting players
	waitingPlayers, err := s.repo.GetWaitingPlayers(ctx)
	if err != nil {
		log.Printf("[MatchmakingWorker] Error fetching waiting players: %v", err)
		return
	}
	if len(waitingPlayers) < 2 {
		log.Println("[MatchmakingWorker] Not enough players waiting")
		return
	}

	// 3. Try to find the best group to match
	var bestGroup []model.PlayerCompetition
	var matchType string

	// 3a. Level-based matching
	levelGroups := make(map[int][]model.PlayerCompetition)
	for _, p := range waitingPlayers {
		levelGroups[p.Level] = append(levelGroups[p.Level], p)
	}
	for level, group := range levelGroups {
		if len(group) >= 2 {
			bestGroup = group
			matchType = fmt.Sprintf("level %d", level)
			break
		}
	}

	// 3b. Country-based matching (if no level group found)
	if len(bestGroup) == 0 {
		countryGroups := make(map[string][]model.PlayerCompetition)
		for _, p := range waitingPlayers {
			countryGroups[p.CountryCode] = append(countryGroups[p.CountryCode], p)
		}
		for country, group := range countryGroups {
			if len(group) >= 2 {
				bestGroup = group
				matchType = fmt.Sprintf("country %s", country)
				break
			}
		}
	}

	// 3c. Fallback: all waiting players
	if len(bestGroup) == 0 {
		bestGroup = waitingPlayers
		matchType = "fallback (all waiting players)"
	}

	// 4. Create the competition for the best group

	compID := uuid.New()
	now := time.Now()
	endsAt := now.Add(s.config.CompetitionDuration)
	comp := &model.Competition{
		CompetitionID: compID,
		StartedAt:     now,
		EndsAt:        endsAt,
		Level:         bestGroup[0].Level,
		CountryCode:   bestGroup[0].CountryCode,
		Status:        model.CompetitionActive,
	}
	if err := s.repo.CreateCompetition(ctx, comp); err != nil {
		log.Printf("[MatchmakingWorker] Error creating competition: %v", err)
		return
	}
	playerIDs := make([]string, len(bestGroup))
	for i, p := range bestGroup {
		playerIDs[i] = p.PlayerID
	}
	if err := s.repo.UpdatePlayerCompetitionsToActive(ctx, playerIDs, compID, endsAt); err != nil {
		log.Printf("[MatchmakingWorker] Error updating player competitions: %v", err)
		return
	}
	log.Printf("[MatchmakingWorker] Started competition %s (%s) with players: %v", compID.String(), matchType, playerIDs)
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
	inQueue, err := s.repo.IsPlayerInWaitingQueue(ctx, playerID)
	if err != nil {
		log.Printf("[Service] Error checking waiting queue for player %s: %v", playerID, err)
		return "", err
	}
	if inQueue {
		log.Printf("[Service] Player %s is already in the waiting queue", playerID)
		return "", errors.New("player already in waiting queue")
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
	// Check if player exists
	_, err := s.repo.GetPlayerByID(ctx, playerID)
	if err != nil {
		log.Printf("[Service] Player %s not found when submitting score", playerID)
		return errors.New("player not found")
	}
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
	err := s.repo.CreatePlayer(ctx, player)
	if err != nil {
		log.Printf("[Service] Error creating player %s: %v", playerID, err)
		return err
	}
	log.Printf("[Service] Successfully created player %s", playerID)
	return nil
}

func (s *Service) GetPlayer(ctx context.Context, playerID string) (*model.Player, error) {
	player, err := s.repo.GetPlayerByID(ctx, playerID)
	if err != nil {
		log.Printf("[Service] Error fetching player %s: %v", playerID, err)
		return nil, err
	}
	log.Printf("[Service] Successfully fetched player %s", playerID)
	return player, nil
}

func (s *Service) UpdatePlayer(ctx context.Context, playerID string, level int, countryCode string) error {
	player, err := s.repo.GetPlayerByID(ctx, playerID)
	if err != nil {
		log.Printf("[Service] Error fetching player %s for update: %v", playerID, err)
		return err
	}
	player.Level = level
	player.CountryCode = countryCode
	err = s.repo.UpdatePlayer(ctx, player)
	if err != nil {
		log.Printf("[Service] Error updating player %s: %v", playerID, err)
		return err
	}
	log.Printf("[Service] Successfully updated player %s", playerID)
	return nil
}
