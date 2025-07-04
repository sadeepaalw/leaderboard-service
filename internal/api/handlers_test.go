package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"leaderboard-service/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockService struct {
	CreatePlayerFunc func(ctx context.Context, playerID string, level int, countryCode string) error
	JoinFunc         func(ctx context.Context, playerID string) (string, error)
}

func (m *mockService) CreatePlayer(ctx context.Context, playerID string, level int, countryCode string) error {
	return m.CreatePlayerFunc(ctx, playerID, level, countryCode)
}
func (m *mockService) Join(ctx context.Context, playerID string) (string, error) {
	return m.JoinFunc(ctx, playerID)
}

// Dummy implementations for unused interface methods
func (m *mockService) SubmitScore(ctx context.Context, playerID string, score int) error { return nil }
func (m *mockService) GetPlayerLeaderboard(ctx context.Context, playerID string) (interface{}, error) {
	return nil, nil
}
func (m *mockService) GetLeaderboard(ctx context.Context, leaderboardID string) (interface{}, error) {
	return nil, nil
}
func (m *mockService) GetPlayer(ctx context.Context, playerID string) (*model.Player, error) {
	return nil, nil
}
func (m *mockService) UpdatePlayer(ctx context.Context, playerID string, level int, countryCode string) error {
	return nil
}

func TestCreatePlayerHandler_Success(t *testing.T) {
	svc := &mockService{
		CreatePlayerFunc: func(ctx context.Context, playerID string, level int, countryCode string) error {
			if playerID != "p1" || level != 2 || countryCode != "US" {
				t.Errorf("unexpected args: %s, %d, %s", playerID, level, countryCode)
			}
			return nil
		},
	}
	h := NewHandler(svc)
	reqBody := map[string]interface{}{"player_id": "p1", "level": 2, "country_code": "US"}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/player", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.CreatePlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("Player created")) {
		t.Errorf("expected Player created message, got %s", string(body))
	}
}

func TestCreatePlayerHandler_Error(t *testing.T) {
	svc := &mockService{
		CreatePlayerFunc: func(ctx context.Context, playerID string, level int, countryCode string) error {
			return errors.New("fail")
		},
	}
	h := NewHandler(svc)
	reqBody := map[string]interface{}{"player_id": "p2", "level": 1, "country_code": "GB"}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/player", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.CreatePlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("fail")) {
		t.Errorf("expected fail message, got %s", string(body))
	}
}

func TestJoinHandler_Success(t *testing.T) {
	svc := &mockService{
		JoinFunc: func(ctx context.Context, playerID string) (string, error) {
			if playerID != "p3" {
				t.Errorf("unexpected playerID: %s", playerID)
			}
			return "leaderboard123", nil
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("POST", "/leaderboard/join?player_id=p3", nil)
	rec := httptest.NewRecorder()

	h.JoinHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected 202, got %d", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("Player added to matchmaking queue")) {
		t.Errorf("expected success message, got %s", string(body))
	}
}

func TestJoinHandler_PlayerNotFound(t *testing.T) {
	svc := &mockService{
		JoinFunc: func(ctx context.Context, playerID string) (string, error) {
			return "", errors.New("player not found")
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("POST", "/leaderboard/join?player_id=p4", nil)
	rec := httptest.NewRecorder()

	h.JoinHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("player not found")) {
		t.Errorf("expected player not found message, got %s", string(body))
	}
}
