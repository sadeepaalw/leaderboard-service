package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"leaderboard-service/internal/model"
	"leaderboard-service/internal/repository"

	"github.com/google/uuid"
)

type mockRepo struct {
	repository.RepositoryInterface
	GetPlayerByIDFunc                 func(ctx context.Context, playerID string) (*model.Player, error)
	GetActivePlayerCompetitionFunc    func(ctx context.Context, playerID string) (*model.PlayerCompetition, error)
	IsPlayerInWaitingQueueFunc        func(ctx context.Context, playerID string) (bool, error)
	CreatePlayerCompetitionFunc       func(ctx context.Context, pc *model.PlayerCompetition) error
	AddScoreToPlayerFunc              func(ctx context.Context, playerID string, score int) error
	GetLeaderboardByCompetitionIDFunc func(ctx context.Context, competitionID string) ([]model.PlayerCompetition, error)
	GetLatestPlayerCompetitionFunc    func(ctx context.Context, playerID string) (*model.PlayerCompetition, error)
	CreatePlayerFunc                  func(ctx context.Context, player *model.Player) error
	UpdatePlayerFunc                  func(ctx context.Context, player *model.Player) error
}

func (m *mockRepo) GetPlayerByID(ctx context.Context, playerID string) (*model.Player, error) {
	return m.GetPlayerByIDFunc(ctx, playerID)
}
func (m *mockRepo) GetActivePlayerCompetition(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
	return m.GetActivePlayerCompetitionFunc(ctx, playerID)
}
func (m *mockRepo) IsPlayerInWaitingQueue(ctx context.Context, playerID string) (bool, error) {
	return m.IsPlayerInWaitingQueueFunc(ctx, playerID)
}
func (m *mockRepo) CreatePlayerCompetition(ctx context.Context, pc *model.PlayerCompetition) error {
	return m.CreatePlayerCompetitionFunc(ctx, pc)
}
func (m *mockRepo) AddScoreToPlayer(ctx context.Context, playerID string, score int) error {
	if m.AddScoreToPlayerFunc != nil {
		return m.AddScoreToPlayerFunc(ctx, playerID, score)
	}
	return nil
}
func (m *mockRepo) GetLeaderboardByCompetitionID(ctx context.Context, competitionID string) ([]model.PlayerCompetition, error) {
	if m.GetLeaderboardByCompetitionIDFunc != nil {
		return m.GetLeaderboardByCompetitionIDFunc(ctx, competitionID)
	}
	return nil, nil
}
func (m *mockRepo) GetLatestPlayerCompetition(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
	if m.GetLatestPlayerCompetitionFunc != nil {
		return m.GetLatestPlayerCompetitionFunc(ctx, playerID)
	}
	return nil, nil
}
func (m *mockRepo) CreatePlayer(ctx context.Context, player *model.Player) error {
	if m.CreatePlayerFunc != nil {
		return m.CreatePlayerFunc(ctx, player)
	}
	return nil
}
func (m *mockRepo) UpdatePlayer(ctx context.Context, player *model.Player) error {
	if m.UpdatePlayerFunc != nil {
		return m.UpdatePlayerFunc(ctx, player)
	}
	return nil
}

func TestService_Join_PlayerNotFound(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewService(repo, Config{})
	_, err := svc.Join(context.Background(), "p1")
	if err == nil || err.Error() != "player not found" {
		t.Errorf("expected player not found error, got %v", err)
	}
}

func TestService_Join_AlreadyInActiveCompetition(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID, Level: 1, CountryCode: "US"}, nil
		},
		GetActivePlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			return &model.PlayerCompetition{}, nil
		},
	}
	svc := NewService(repo, Config{})
	_, err := svc.Join(context.Background(), "p2")
	if err == nil || err.Error() != "player already in active competition" {
		t.Errorf("expected already in active competition error, got %v", err)
	}
}

func TestService_Join_AlreadyInWaitingQueue(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID, Level: 1, CountryCode: "US"}, nil
		},
		GetActivePlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			return nil, errors.New("not found")
		},
		IsPlayerInWaitingQueueFunc: func(ctx context.Context, playerID string) (bool, error) {
			return true, nil
		},
	}
	svc := NewService(repo, Config{})
	_, err := svc.Join(context.Background(), "p3")
	if err == nil || err.Error() != "player already in waiting queue" {
		t.Errorf("expected already in waiting queue error, got %v", err)
	}
}

func TestService_Join_Success(t *testing.T) {
	called := false
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID, Level: 2, CountryCode: "GB"}, nil
		},
		GetActivePlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			return nil, errors.New("not found")
		},
		IsPlayerInWaitingQueueFunc: func(ctx context.Context, playerID string) (bool, error) {
			return false, nil
		},
		CreatePlayerCompetitionFunc: func(ctx context.Context, pc *model.PlayerCompetition) error {
			called = true
			if pc.PlayerID != "p4" || pc.Level != 2 || pc.CountryCode != "GB" || pc.Status != model.StatusWaiting {
				t.Errorf("unexpected player competition: %+v", pc)
			}
			return nil
		},
	}
	svc := NewService(repo, Config{})
	_, err := svc.Join(context.Background(), "p4")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Errorf("expected CreatePlayerCompetition to be called")
	}
}

func TestService_SubmitScore_PlayerNotFound(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewService(repo, Config{})
	err := svc.SubmitScore(context.Background(), "p1", 10)
	if err == nil || err.Error() != "player not found" {
		t.Errorf("expected player not found error, got %v", err)
	}
}

func TestService_SubmitScore_NotInActiveCompetition(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID}, nil
		},
		GetActivePlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewService(repo, Config{})
	err := svc.SubmitScore(context.Background(), "p2", 10)
	if err == nil || err.Error() != "player not in active competition" {
		t.Errorf("expected player not in active competition error, got %v", err)
	}
}

func TestService_SubmitScore_Success(t *testing.T) {
	called := false
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID}, nil
		},
		GetActivePlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			return &model.PlayerCompetition{PlayerID: playerID}, nil
		},
		AddScoreToPlayerFunc: func(ctx context.Context, playerID string, score int) error {
			called = true
			if playerID != "p3" || score != 99 {
				t.Errorf("unexpected AddScoreToPlayer args: %s, %d", playerID, score)
			}
			return nil
		},
	}
	svc := NewService(repo, Config{})
	err := svc.SubmitScore(context.Background(), "p3", 99)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Errorf("expected AddScoreToPlayer to be called")
	}
}

func TestService_GetPlayerLeaderboard_NoCompetition(t *testing.T) {
	repo := &mockRepo{
		GetLatestPlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewService(repo, Config{})
	resp, err := svc.GetPlayerLeaderboard(context.Background(), "p1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if m, ok := resp.(map[string]interface{}); !ok || len(m) != 0 {
		t.Errorf("expected empty map, got %v", resp)
	}
}

func TestService_GetPlayerLeaderboard_NoCompetitionID(t *testing.T) {
	repo := &mockRepo{
		GetLatestPlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			return &model.PlayerCompetition{CompetitionID: nil}, nil
		},
	}
	svc := NewService(repo, Config{})
	resp, err := svc.GetPlayerLeaderboard(context.Background(), "p2")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if m, ok := resp.(map[string]interface{}); !ok || len(m) != 0 {
		t.Errorf("expected empty map, got %v", resp)
	}
}

func TestService_GetPlayerLeaderboard_LeaderboardFetchError(t *testing.T) {
	repo := &mockRepo{
		GetLatestPlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			id := uuid.New()
			return &model.PlayerCompetition{CompetitionID: &id}, nil
		},
		GetLeaderboardByCompetitionIDFunc: func(ctx context.Context, competitionID string) ([]model.PlayerCompetition, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewService(repo, Config{})
	_, err := svc.GetPlayerLeaderboard(context.Background(), "p3")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestService_GetPlayerLeaderboard_Success(t *testing.T) {
	compID := uuid.New()
	entries := []model.PlayerCompetition{{PlayerID: "p1", Score: 10}, {PlayerID: "p2", Score: 5}}
	repo := &mockRepo{
		GetLatestPlayerCompetitionFunc: func(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
			return &model.PlayerCompetition{CompetitionID: &compID, UpdatedAt: time.Unix(123, 0)}, nil
		},
		GetLeaderboardByCompetitionIDFunc: func(ctx context.Context, competitionID string) ([]model.PlayerCompetition, error) {
			return entries, nil
		},
	}
	svc := NewService(repo, Config{})
	resp, err := svc.GetPlayerLeaderboard(context.Background(), "p1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	m, ok := resp.(map[string]interface{})
	if !ok || m["leaderboard_id"] != compID.String() {
		t.Errorf("expected leaderboard_id %s, got %v", compID.String(), m["leaderboard_id"])
	}
}

func TestService_CreatePlayer_Error(t *testing.T) {
	repo := &mockRepo{
		CreatePlayerFunc: func(ctx context.Context, player *model.Player) error {
			return errors.New("fail")
		},
	}
	svc := NewService(repo, Config{})
	err := svc.CreatePlayer(context.Background(), "p1", 1, "US")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestService_CreatePlayer_Success(t *testing.T) {
	called := false
	repo := &mockRepo{
		CreatePlayerFunc: func(ctx context.Context, player *model.Player) error {
			called = true
			if player.PlayerID != "p2" || player.Level != 2 || player.CountryCode != "GB" {
				t.Errorf("unexpected player: %+v", player)
			}
			return nil
		},
	}
	svc := NewService(repo, Config{})
	err := svc.CreatePlayer(context.Background(), "p2", 2, "GB")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Errorf("expected CreatePlayer to be called")
	}
}

func TestService_GetPlayer_Error(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return nil, errors.New("fail")
		},
	}
	svc := NewService(repo, Config{})
	_, err := svc.GetPlayer(context.Background(), "p1")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestService_GetPlayer_Success(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID, Level: 1, CountryCode: "US"}, nil
		},
	}
	svc := NewService(repo, Config{})
	p, err := svc.GetPlayer(context.Background(), "p1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if p.PlayerID != "p1" {
		t.Errorf("expected player_id p1, got %s", p.PlayerID)
	}
}

func TestService_UpdatePlayer_ErrorFetching(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return nil, errors.New("fail")
		},
	}
	svc := NewService(repo, Config{})
	err := svc.UpdatePlayer(context.Background(), "p1", 2, "GB")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestService_UpdatePlayer_ErrorUpdating(t *testing.T) {
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID, Level: 1, CountryCode: "US"}, nil
		},
		UpdatePlayerFunc: func(ctx context.Context, player *model.Player) error {
			return errors.New("fail")
		},
	}
	svc := NewService(repo, Config{})
	err := svc.UpdatePlayer(context.Background(), "p1", 2, "GB")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestService_UpdatePlayer_Success(t *testing.T) {
	called := false
	repo := &mockRepo{
		GetPlayerByIDFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID, Level: 1, CountryCode: "US"}, nil
		},
		UpdatePlayerFunc: func(ctx context.Context, player *model.Player) error {
			called = true
			if player.Level != 2 || player.CountryCode != "GB" {
				t.Errorf("unexpected update: %+v", player)
			}
			return nil
		},
	}
	svc := NewService(repo, Config{})
	err := svc.UpdatePlayer(context.Background(), "p1", 2, "GB")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Errorf("expected UpdatePlayer to be called")
	}
}
