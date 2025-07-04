package repository

import (
	"context"
	"database/sql"
	"leaderboard-service/internal/model"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var testDSN = "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	return db
}

func cleanupPlayerCompetitionByPlayerID(t *testing.T, db *sql.DB, playerID string) {
	_, err := db.Exec("DELETE FROM player_competitions WHERE player_id = $1", playerID)
	if err != nil {
		t.Fatalf("failed to cleanup player_competition: %v", err)
	}
}

func cleanupPlayerCompetitionByCompetitionID(t *testing.T, db *sql.DB, competitionID string) {
	_, err := db.Exec("DELETE FROM player_competitions WHERE competition_id = $1", competitionID)
	if err != nil {
		t.Fatalf("failed to cleanup player_competition: %v", err)
	}
}

func cleanupPlayerCompetition(t *testing.T, db *sql.DB, id int) {
	_, err := db.Exec("DELETE FROM player_competitions WHERE id = $1", id)
	if err != nil {
		t.Fatalf("failed to cleanup player_competition: %v", err)
	}
}

func cleanupPlayer(t *testing.T, db *sql.DB, playerID string) {
	_, err := db.Exec("DELETE FROM players WHERE player_id = $1", playerID)
	if err != nil {
		t.Fatalf("failed to cleanup player: %v", err)
	}
}

func cleanupCompetition(t *testing.T, db *sql.DB, competitionID string) {
	_, err := db.Exec("DELETE FROM competitions WHERE competition_id = $1", competitionID)
	if err != nil {
		t.Fatalf("failed to cleanup competition: %v", err)
	}
}

func cleanupPlayerCompetitionByPlayerAndCompetition(t *testing.T, db *sql.DB, playerID, competitionID string) {
	_, err := db.Exec("DELETE FROM player_competitions WHERE player_id = $1 OR competition_id = $2", playerID, competitionID)
	if err != nil {
		t.Fatalf("failed to cleanup player_competition: %v", err)
	}
}

func TestCreateAndGetPlayer(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	player := &model.Player{PlayerID: "testplayer1", Level: 5, CountryCode: "US"}
	defer cleanupPlayer(t, db, player.PlayerID)

	err := repo.CreatePlayer(context.Background(), player)
	if err != nil {
		t.Fatalf("CreatePlayer failed: %v", err)
	}

	got, err := repo.GetPlayerByID(context.Background(), player.PlayerID)
	if err != nil {
		t.Fatalf("GetPlayerByID failed: %v", err)
	}
	if got.PlayerID != player.PlayerID || got.Level != player.Level || got.CountryCode != player.CountryCode {
		t.Errorf("GetPlayerByID returned wrong player: got %+v, want %+v", got, player)
	}
}

func TestUpdatePlayer(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	player := &model.Player{PlayerID: "testplayer2", Level: 1, CountryCode: "CA"}
	defer cleanupPlayerCompetitionByPlayerID(t, db, player.PlayerID)
	defer cleanupPlayer(t, db, player.PlayerID)

	err := repo.CreatePlayer(context.Background(), player)
	if err != nil {
		t.Fatalf("CreatePlayer failed: %v", err)
	}

	player.Level = 10
	player.CountryCode = "GB"
	err = repo.UpdatePlayer(context.Background(), player)
	if err != nil {
		t.Fatalf("UpdatePlayer failed: %v", err)
	}

	got, err := repo.GetPlayerByID(context.Background(), player.PlayerID)
	if err != nil {
		t.Fatalf("GetPlayerByID failed: %v", err)
	}
	if got.Level != 10 || got.CountryCode != "GB" {
		t.Errorf("UpdatePlayer did not update fields: got %+v", got)
	}
}

func TestCreateAndGetCompetition(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	compID := uuid.New()
	comp := &model.Competition{
		CompetitionID: compID,
		StartedAt:     time.Now(),
		EndsAt:        time.Now().Add(time.Hour),
		Level:         1,
		CountryCode:   "US",
	}
	defer cleanupPlayerCompetitionByCompetitionID(t, db, compID.String())
	defer cleanupCompetition(t, db, compID.String())

	err := repo.CreateCompetition(context.Background(), comp)
	if err != nil {
		t.Fatalf("CreateCompetition failed: %v", err)
	}

	got, err := repo.GetCompetitionByID(context.Background(), compID.String())
	if err != nil {
		t.Fatalf("GetCompetitionByID failed: %v", err)
	}
	if got.CompetitionID != comp.CompetitionID || got.Level != comp.Level || got.CountryCode != comp.CountryCode {
		t.Errorf("GetCompetitionByID returned wrong competition: got %+v, want %+v", got, comp)
	}
}

func TestUpdateCompetition(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	compID := uuid.New()
	comp := &model.Competition{
		CompetitionID: compID,
		StartedAt:     time.Now(),
		EndsAt:        time.Now().Add(time.Hour),
		Level:         1,
		CountryCode:   "US",
	}
	defer cleanupPlayerCompetitionByCompetitionID(t, db, compID.String())
	defer cleanupCompetition(t, db, compID.String())

	err := repo.CreateCompetition(context.Background(), comp)
	if err != nil {
		t.Fatalf("CreateCompetition failed: %v", err)
	}

	comp.Level = 2
	comp.CountryCode = "GB"
	err = repo.UpdateCompetition(context.Background(), comp)
	if err != nil {
		t.Fatalf("UpdateCompetition failed: %v", err)
	}

	got, err := repo.GetCompetitionByID(context.Background(), compID.String())
	if err != nil {
		t.Fatalf("GetCompetitionByID failed: %v", err)
	}
	if got.Level != 2 || got.CountryCode != "GB" {
		t.Errorf("UpdateCompetition did not update fields: got %+v", got)
	}
}

func TestCreateAndGetPlayerCompetition(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	compID := uuid.New()
	playerID := "testplayer3"
	pc := &model.PlayerCompetition{
		PlayerID:      playerID,
		CompetitionID: &compID,
		Status:        "WAITING",
		Score:         0,
		JoinedAt:      time.Now(),
		UpdatedAt:     time.Now(),
		Level:         1,
		CountryCode:   "US",
	}
	_, _ = db.Exec("INSERT INTO players (player_id, level, country_code) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", playerID, 1, "US")
	_, _ = db.Exec("INSERT INTO competitions (competition_id, started_at, ends_at, level, country_code) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING", compID, time.Now(), time.Now().Add(time.Hour), 1, "US")

	idBefore := 0
	err := repo.CreatePlayerCompetition(context.Background(), pc)
	if err != nil {
		t.Fatalf("CreatePlayerCompetition failed: %v", err)
	}
	row := db.QueryRow("SELECT max(id) FROM player_competitions WHERE player_id = $1", playerID)
	row.Scan(&idBefore)
	defer cleanupPlayer(t, db, playerID)
	defer cleanupCompetition(t, db, compID.String())
	defer cleanupPlayerCompetitionByPlayerAndCompetition(t, db, playerID, compID.String())

	got, err := repo.GetPlayerCompetitionByID(context.Background(), idBefore)
	if err != nil {
		t.Fatalf("GetPlayerCompetitionByID failed: %v", err)
	}
	if got.PlayerID != pc.PlayerID || got.Level != pc.Level || got.CountryCode != pc.CountryCode {
		t.Errorf("GetPlayerCompetitionByID returned wrong player_competition: got %+v, want %+v", got, pc)
	}
}

func TestUpdatePlayerCompetition(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	compID := uuid.New()
	playerID := "testplayer4"
	pc := &model.PlayerCompetition{
		PlayerID:      playerID,
		CompetitionID: &compID,
		Status:        "WAITING",
		Score:         0,
		JoinedAt:      time.Now(),
		UpdatedAt:     time.Now(),
		Level:         1,
		CountryCode:   "US",
	}
	_, _ = db.Exec("INSERT INTO players (player_id, level, country_code) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", playerID, 1, "US")
	_, _ = db.Exec("INSERT INTO competitions (competition_id, started_at, ends_at, level, country_code) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING", compID, time.Now(), time.Now().Add(time.Hour), 1, "US")

	err := repo.CreatePlayerCompetition(context.Background(), pc)
	if err != nil {
		t.Fatalf("CreatePlayerCompetition failed: %v", err)
	}
	row := db.QueryRow("SELECT max(id) FROM player_competitions WHERE player_id = $1", playerID)
	var id int
	row.Scan(&id)
	defer cleanupPlayer(t, db, playerID)
	defer cleanupCompetition(t, db, compID.String())
	defer cleanupPlayerCompetitionByPlayerAndCompetition(t, db, playerID, compID.String())

	pc.ID = id
	pc.Status = "ACTIVE"
	pc.Score = 50
	pc.Level = 2
	pc.CountryCode = "GB"
	pc.UpdatedAt = time.Now()
	err = repo.UpdatePlayerCompetition(context.Background(), pc)
	if err != nil {
		t.Fatalf("UpdatePlayerCompetition failed: %v", err)
	}

	got, err := repo.GetPlayerCompetitionByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetPlayerCompetitionByID failed: %v", err)
	}
	if got.Status != "ACTIVE" || got.Score != 50 || got.Level != 2 || got.CountryCode != "GB" {
		t.Errorf("UpdatePlayerCompetition did not update fields: got %+v", got)
	}
}
