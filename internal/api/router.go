package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/hello", handler.HelloHandler).Methods("GET")
	r.HandleFunc("/leaderboard/join", handler.JoinHandler).Methods("POST")
	r.HandleFunc("/leaderboard/player/{player_id}", handler.PlayerLeaderboardHandler).Methods("GET")
	r.HandleFunc("/leaderboard/{leaderboardID}", handler.LeaderboardHandler).Methods("GET")
	r.HandleFunc("/leaderboard/score", handler.ScoreHandler).Methods("POST")

	// Player CRUD
	r.HandleFunc("/player", handler.CreatePlayerHandler).Methods("POST")
	r.HandleFunc("/player/{player_id}", handler.GetPlayerHandler).Methods("GET")
	r.HandleFunc("/player/{player_id}", handler.UpdatePlayerHandler).Methods("PUT")

	return r
}
