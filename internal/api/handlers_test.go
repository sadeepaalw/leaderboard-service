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

	"github.com/gorilla/mux"
)

type mockService struct {
	CreatePlayerFunc         func(ctx context.Context, playerID string, level int, countryCode string) error
	JoinFunc                 func(ctx context.Context, playerID string) (string, error)
	GetPlayerLeaderboardFunc func(ctx context.Context, playerID string) (interface{}, error)
	GetLeaderboardFunc       func(ctx context.Context, leaderboardID string) (interface{}, error)
	SubmitScoreFunc          func(ctx context.Context, playerID string, score int) error
	GetPlayerFunc            func(ctx context.Context, playerID string) (*model.Player, error)
	UpdatePlayerFunc         func(ctx context.Context, playerID string, level int, countryCode string) error
}

func (m *mockService) CreatePlayer(ctx context.Context, playerID string, level int, countryCode string) error {
	return m.CreatePlayerFunc(ctx, playerID, level, countryCode)
}
func (m *mockService) Join(ctx context.Context, playerID string) (string, error) {
	return m.JoinFunc(ctx, playerID)
}
func (m *mockService) GetPlayerLeaderboard(ctx context.Context, playerID string) (interface{}, error) {
	if m.GetPlayerLeaderboardFunc != nil {
		return m.GetPlayerLeaderboardFunc(ctx, playerID)
	}
	return nil, nil
}
func (m *mockService) GetLeaderboard(ctx context.Context, leaderboardID string) (interface{}, error) {
	if m.GetLeaderboardFunc != nil {
		return m.GetLeaderboardFunc(ctx, leaderboardID)
	}
	return nil, nil
}
func (m *mockService) SubmitScore(ctx context.Context, playerID string, score int) error {
	if m.SubmitScoreFunc != nil {
		return m.SubmitScoreFunc(ctx, playerID, score)
	}
	return nil
}
func (m *mockService) GetPlayer(ctx context.Context, playerID string) (*model.Player, error) {
	if m.GetPlayerFunc != nil {
		return m.GetPlayerFunc(ctx, playerID)
	}
	return nil, nil
}
func (m *mockService) UpdatePlayer(ctx context.Context, playerID string, level int, countryCode string) error {
	if m.UpdatePlayerFunc != nil {
		return m.UpdatePlayerFunc(ctx, playerID, level, countryCode)
	}
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

func TestHelloHandler(t *testing.T) {
	h := NewHandler(&mockService{})
	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()

	h.HelloHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("Hello, World!")) {
		t.Errorf("expected Hello, World! message, got %s", string(body))
	}
}

func TestPlayerLeaderboardHandler_Success(t *testing.T) {
	svc := &mockService{
		GetPlayerLeaderboardFunc: func(ctx context.Context, playerID string) (interface{}, error) {
			return map[string]interface{}{"leaderboard_id": "lid", "leaderboard": []interface{}{}}, nil
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/leaderboard/player/p1", nil)
	rec := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"player_id": "p1"})
	h.PlayerLeaderboardHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("leaderboard_id")) {
		t.Errorf("expected leaderboard_id in response, got %s", string(body))
	}
}

func TestPlayerLeaderboardHandler_Error(t *testing.T) {
	svc := &mockService{
		GetPlayerLeaderboardFunc: func(ctx context.Context, playerID string) (interface{}, error) {
			return nil, errors.New("fail")
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/leaderboard/player/p1", nil)
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"player_id": "p1"})
	h.PlayerLeaderboardHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestLeaderboardHandler_Success(t *testing.T) {
	svc := &mockService{
		GetLeaderboardFunc: func(ctx context.Context, leaderboardID string) (interface{}, error) {
			return map[string]interface{}{"leaderboard_id": leaderboardID, "leaderboard": []interface{}{}}, nil
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/leaderboard/lid", nil)
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"leaderboardID": "lid"})
	h.LeaderboardHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLeaderboardHandler_NotFound(t *testing.T) {
	svc := &mockService{
		GetLeaderboardFunc: func(ctx context.Context, leaderboardID string) (interface{}, error) {
			return nil, errors.New("not found")
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/leaderboard/lid", nil)
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"leaderboardID": "lid"})
	h.LeaderboardHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestScoreHandler_Success(t *testing.T) {
	svc := &mockService{
		SubmitScoreFunc: func(ctx context.Context, playerID string, score int) error {
			if playerID != "p1" || score != 42 {
				t.Errorf("unexpected args: %s, %d", playerID, score)
			}
			return nil
		},
	}
	h := NewHandler(svc)
	b, _ := json.Marshal(map[string]interface{}{"player_id": "p1", "score": 42})
	req := httptest.NewRequest("POST", "/leaderboard/score", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.ScoreHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestScoreHandler_PlayerNotFound(t *testing.T) {
	svc := &mockService{
		SubmitScoreFunc: func(ctx context.Context, playerID string, score int) error {
			return errors.New("player not found")
		},
	}
	h := NewHandler(svc)
	b, _ := json.Marshal(map[string]interface{}{"player_id": "p2", "score": 10})
	req := httptest.NewRequest("POST", "/leaderboard/score", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.ScoreHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestScoreHandler_BadRequest(t *testing.T) {
	h := NewHandler(&mockService{})
	req := httptest.NewRequest("POST", "/leaderboard/score", bytes.NewReader([]byte("notjson")))
	rec := httptest.NewRecorder()

	h.ScoreHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetPlayerHandler_Success(t *testing.T) {
	svc := &mockService{
		GetPlayerFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return &model.Player{PlayerID: playerID, Level: 1, CountryCode: "US"}, nil
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/player/p1", nil)
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"player_id": "p1"})
	h.GetPlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetPlayerHandler_NotFound(t *testing.T) {
	svc := &mockService{
		GetPlayerFunc: func(ctx context.Context, playerID string) (*model.Player, error) {
			return nil, errors.New("not found")
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/player/p2", nil)
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"player_id": "p2"})
	h.GetPlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdatePlayerHandler_Success(t *testing.T) {
	svc := &mockService{
		UpdatePlayerFunc: func(ctx context.Context, playerID string, level int, countryCode string) error {
			if playerID != "p1" || level != 2 || countryCode != "GB" {
				t.Errorf("unexpected update: %s, %d, %s", playerID, level, countryCode)
			}
			return nil
		},
	}
	h := NewHandler(svc)
	b, _ := json.Marshal(map[string]interface{}{"level": 2, "country_code": "GB"})
	req := httptest.NewRequest("PUT", "/player/p1", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"player_id": "p1"})
	h.UpdatePlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUpdatePlayerHandler_BadRequest(t *testing.T) {
	h := NewHandler(&mockService{})
	req := httptest.NewRequest("PUT", "/player/p1", bytes.NewReader([]byte("notjson")))
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"player_id": "p1"})
	h.UpdatePlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdatePlayerHandler_Error(t *testing.T) {
	svc := &mockService{
		UpdatePlayerFunc: func(ctx context.Context, playerID string, level int, countryCode string) error {
			return errors.New("fail")
		},
	}
	h := NewHandler(svc)
	b, _ := json.Marshal(map[string]interface{}{"level": 2, "country_code": "GB"})
	req := httptest.NewRequest("PUT", "/player/p1", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"player_id": "p1"})
	h.UpdatePlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestJoinHandler_AlreadyInActiveCompetition(t *testing.T) {
	svc := &mockService{
		JoinFunc: func(ctx context.Context, playerID string) (string, error) {
			return "", errors.New("player already in active competition")
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("POST", "/leaderboard/join?player_id=p5", nil)
	rec := httptest.NewRecorder()

	h.JoinHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", resp.StatusCode)
	}
}

func TestJoinHandler_AlreadyInWaitingQueue(t *testing.T) {
	svc := &mockService{
		JoinFunc: func(ctx context.Context, playerID string) (string, error) {
			return "", errors.New("player already in waiting queue")
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("POST", "/leaderboard/join?player_id=p6", nil)
	rec := httptest.NewRecorder()

	h.JoinHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", resp.StatusCode)
	}
}

func TestJoinHandler_InternalError(t *testing.T) {
	svc := &mockService{
		JoinFunc: func(ctx context.Context, playerID string) (string, error) {
			return "", errors.New("some internal error")
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("POST", "/leaderboard/join?player_id=p7", nil)
	rec := httptest.NewRecorder()

	h.JoinHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestCreatePlayerHandler_BadRequest(t *testing.T) {
	h := NewHandler(&mockService{})
	req := httptest.NewRequest("POST", "/player", bytes.NewReader([]byte("notjson")))
	rec := httptest.NewRecorder()

	h.CreatePlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestScoreHandler_PlayerNotInActiveCompetition(t *testing.T) {
	svc := &mockService{
		SubmitScoreFunc: func(ctx context.Context, playerID string, score int) error {
			return errors.New("player not in active competition")
		},
	}
	h := NewHandler(svc)
	b, _ := json.Marshal(map[string]interface{}{"player_id": "p3", "score": 10})
	req := httptest.NewRequest("POST", "/leaderboard/score", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.ScoreHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", resp.StatusCode)
	}
}

func TestScoreHandler_InternalError(t *testing.T) {
	svc := &mockService{
		SubmitScoreFunc: func(ctx context.Context, playerID string, score int) error {
			return errors.New("db error")
		},
	}
	h := NewHandler(svc)
	b, _ := json.Marshal(map[string]interface{}{"player_id": "p4", "score": 10})
	req := httptest.NewRequest("POST", "/leaderboard/score", bytes.NewReader(b))
	rec := httptest.NewRecorder()

	h.ScoreHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestUpdatePlayerHandler_InternalError(t *testing.T) {
	svc := &mockService{
		UpdatePlayerFunc: func(ctx context.Context, playerID string, level int, countryCode string) error {
			return errors.New("db error")
		},
	}
	h := NewHandler(svc)
	b, _ := json.Marshal(map[string]interface{}{"level": 2, "country_code": "GB"})
	req := httptest.NewRequest("PUT", "/player/p1", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"player_id": "p1"})
	h.UpdatePlayerHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestPlayerLeaderboardHandler_EmptyLeaderboard(t *testing.T) {
	svc := &mockService{
		GetPlayerLeaderboardFunc: func(ctx context.Context, playerID string) (interface{}, error) {
			return map[string]interface{}{}, nil
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/leaderboard/player/p8", nil)
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"player_id": "p8"})
	h.PlayerLeaderboardHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != "{}" {
		t.Errorf("expected empty JSON object, got %s", string(body))
	}
}

func TestLeaderboardHandler_InternalError(t *testing.T) {
	svc := &mockService{
		GetLeaderboardFunc: func(ctx context.Context, leaderboardID string) (interface{}, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest("GET", "/leaderboard/lid", nil)
	rec := httptest.NewRecorder()
	req = mux.SetURLVars(req, map[string]string{"leaderboardID": "lid"})
	h.LeaderboardHandler(rec, req)
	resp := rec.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
