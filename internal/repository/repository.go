package repository

import (
	"context"
	"database/sql"
	"leaderboard-service/internal/model"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreatePlayer(ctx context.Context, player *model.Player) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO players (player_id, level, country_code) VALUES ($1, $2, $3)`,
		player.PlayerID, player.Level, player.CountryCode,
	)
	if err != nil {
		log.Printf("[Repository] Error creating player %s: %v", player.PlayerID, err)
		return err
	}
	log.Printf("[Repository] Successfully created player %s", player.PlayerID)
	return nil
}

func (r *Repository) GetPlayerByID(ctx context.Context, playerID string) (*model.Player, error) {
	var player model.Player
	err := r.db.QueryRowContext(ctx,
		`SELECT player_id, level, country_code FROM players WHERE player_id = $1`,
		playerID,
	).Scan(&player.PlayerID, &player.Level, &player.CountryCode)
	if err != nil {
		log.Printf("[Repository] Error fetching player %s: %v", playerID, err)
		return nil, err
	}
	log.Printf("[Repository] Successfully fetched player %s", playerID)
	return &player, nil
}

func (r *Repository) UpdatePlayer(ctx context.Context, player *model.Player) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE players SET level = $2, country_code = $3 WHERE player_id = $1`,
		player.PlayerID, player.Level, player.CountryCode,
	)
	if err != nil {
		log.Printf("[Repository] Error updating player %s: %v", player.PlayerID, err)
		return err
	}
	log.Printf("[Repository] Successfully updated player %s", player.PlayerID)
	return nil
}

// Competition methods
func (r *Repository) GetActiveCompetition(ctx context.Context) (*model.Competition, error) {
	var comp model.Competition
	err := r.db.QueryRowContext(ctx,
		`SELECT competition_id, started_at, ends_at, level, country_code, status FROM competitions WHERE status = 'ACTIVE' LIMIT 1`,
	).Scan(&comp.CompetitionID, &comp.StartedAt, &comp.EndsAt, &comp.Level, &comp.CountryCode, &comp.Status)
	if err != nil {
		return nil, err
	}
	return &comp, nil
}

func (r *Repository) CreateCompetition(ctx context.Context, comp *model.Competition) error {
	log.Printf("[Repository] Creating competition %s", comp.CompetitionID.String())
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO competitions (competition_id, started_at, ends_at, level, country_code, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, comp.CompetitionID, comp.StartedAt, comp.EndsAt, comp.Level, comp.CountryCode, comp.Status)
	if err != nil {
		log.Printf("[Repository] Error creating competition: %v", err)
	}
	return err
}

func (r *Repository) GetCompetitionByID(ctx context.Context, competitionID string) (*model.Competition, error) {
	var comp model.Competition
	err := r.db.QueryRowContext(ctx,
		`SELECT competition_id, started_at, ends_at, level, country_code, status FROM competitions WHERE competition_id = $1`,
		competitionID,
	).Scan(&comp.CompetitionID, &comp.StartedAt, &comp.EndsAt, &comp.Level, &comp.CountryCode, &comp.Status)
	if err != nil {
		return nil, err
	}
	return &comp, nil
}

func (r *Repository) UpdateCompetition(ctx context.Context, comp *model.Competition) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE competitions SET started_at = $2, ends_at = $3, level = $4, country_code = $5, status = $6 WHERE competition_id = $1`,
		comp.CompetitionID, comp.StartedAt, comp.EndsAt, comp.Level, comp.CountryCode, comp.Status,
	)
	return err
}

// PlayerCompetition methods
func (r *Repository) CreatePlayerCompetition(ctx context.Context, pc *model.PlayerCompetition) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO player_competitions (player_id, competition_id, status, score, joined_at, updated_at, level, country_code) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		pc.PlayerID, pc.CompetitionID, pc.Status, pc.Score, pc.JoinedAt, pc.UpdatedAt, pc.Level, pc.CountryCode,
	)
	if err != nil {
		log.Printf("[Repository] Error creating player_competition for player %s: %v", pc.PlayerID, err)
		return err
	}
	log.Printf("[Repository] Successfully created player_competition for player %s", pc.PlayerID)
	return nil
}

func (r *Repository) GetPlayerCompetitionByID(ctx context.Context, id int) (*model.PlayerCompetition, error) {
	var pc model.PlayerCompetition
	err := r.db.QueryRowContext(ctx,
		`SELECT id, player_id, competition_id, status, score, joined_at, updated_at, level, country_code FROM player_competitions WHERE id = $1`,
		id,
	).Scan(&pc.ID, &pc.PlayerID, &pc.CompetitionID, &pc.Status, &pc.Score, &pc.JoinedAt, &pc.UpdatedAt, &pc.Level, &pc.CountryCode)
	if err != nil {
		log.Printf("[Repository] Error fetching player_competition %d: %v", id, err)
		return nil, err
	}
	log.Printf("[Repository] Successfully fetched player_competition %d", id)
	return &pc, nil
}

func (r *Repository) UpdatePlayerCompetition(ctx context.Context, pc *model.PlayerCompetition) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE player_competitions SET player_id = $2, competition_id = $3, status = $4, score = $5, joined_at = $6, updated_at = $7, level = $8, country_code = $9 WHERE id = $1`,
		pc.ID, pc.PlayerID, pc.CompetitionID, pc.Status, pc.Score, pc.JoinedAt, pc.UpdatedAt, pc.Level, pc.CountryCode,
	)
	if err != nil {
		log.Printf("[Repository] Error updating player_competition %d: %v", pc.ID, err)
		return err
	}
	log.Printf("[Repository] Successfully updated player_competition %d", pc.ID)
	return nil
}

func (r *Repository) GetLatestPlayerCompetition(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
	var pc model.PlayerCompetition
	err := r.db.QueryRowContext(ctx, `
		SELECT id, player_id, competition_id, status, score, joined_at, updated_at, level, country_code
		FROM player_competitions
		WHERE player_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, playerID).Scan(&pc.ID, &pc.PlayerID, &pc.CompetitionID, &pc.Status, &pc.Score, &pc.JoinedAt, &pc.UpdatedAt, &pc.Level, &pc.CountryCode)
	if err != nil {
		log.Printf("[Repository] Error fetching latest player_competition for player %s: %v", playerID, err)
		return nil, err
	}
	log.Printf("[Repository] Successfully fetched latest player_competition for player %s", playerID)
	return &pc, nil
}

func (r *Repository) GetLeaderboardByCompetitionID(ctx context.Context, competitionID string) ([]model.PlayerCompetition, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, player_id, competition_id, status, score, joined_at, updated_at, level, country_code
		FROM player_competitions
		WHERE competition_id = $1
		ORDER BY score DESC, player_id ASC
	`, competitionID)
	if err != nil {
		log.Printf("[Repository] Error fetching leaderboard for competition %s: %v", competitionID, err)
		return nil, err
	}
	defer rows.Close()

	var pcs []model.PlayerCompetition
	for rows.Next() {
		var pc model.PlayerCompetition
		if err := rows.Scan(&pc.ID, &pc.PlayerID, &pc.CompetitionID, &pc.Status, &pc.Score, &pc.JoinedAt, &pc.UpdatedAt, &pc.Level, &pc.CountryCode); err != nil {
			log.Printf("[Repository] Error scanning leaderboard entry for competition %s: %v", competitionID, err)
			return nil, err
		}
		pcs = append(pcs, pc)
	}
	log.Printf("[Repository] Successfully fetched leaderboard for competition %s", competitionID)
	return pcs, nil
}

func (r *Repository) GetActivePlayerCompetition(ctx context.Context, playerID string) (*model.PlayerCompetition, error) {
	var pc model.PlayerCompetition
	err := r.db.QueryRowContext(ctx, `
		SELECT pc.id, pc.player_id, pc.competition_id, pc.status, pc.score, pc.joined_at, pc.updated_at, pc.level, pc.country_code
		FROM player_competitions pc
		JOIN competitions c ON pc.competition_id = c.competition_id
		WHERE pc.player_id = $1 AND pc.status = 'ACTIVE' AND c.ends_at > NOW()
		LIMIT 1
	`, playerID).Scan(&pc.ID, &pc.PlayerID, &pc.CompetitionID, &pc.Status, &pc.Score, &pc.JoinedAt, &pc.UpdatedAt, &pc.Level, &pc.CountryCode)
	if err != nil {
		return nil, err
	}
	return &pc, nil
}

func (r *Repository) GetWaitingPlayers(ctx context.Context) ([]model.PlayerCompetition, error) {
	log.Println("[Repository] Fetching waiting players")
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, player_id, competition_id, status, score, joined_at, updated_at, level, country_code
		FROM player_competitions
		WHERE status = 'WAITING'
		ORDER BY joined_at
		LIMIT 10
	`)
	if err != nil {
		log.Printf("[Repository] Error fetching waiting players: %v", err)
		return nil, err
	}
	defer rows.Close()

	var pcs []model.PlayerCompetition
	for rows.Next() {
		var pc model.PlayerCompetition
		if err := rows.Scan(&pc.ID, &pc.PlayerID, &pc.CompetitionID, &pc.Status, &pc.Score, &pc.JoinedAt, &pc.UpdatedAt, &pc.Level, &pc.CountryCode); err != nil {
			log.Printf("[Repository] Error scanning waiting player: %v", err)
			return nil, err
		}
		pcs = append(pcs, pc)
	}
	return pcs, nil
}

func (r *Repository) UpdatePlayerCompetitionsToActive(ctx context.Context, playerIDs []string, competitionID uuid.UUID, endsAt time.Time) error {
	log.Printf("[Repository] Updating %d players to ACTIVE for competition %s", len(playerIDs), competitionID.String())
	if len(playerIDs) == 0 {
		return nil
	}
	// Build the IN clause
	placeholders := make([]string, len(playerIDs))
	args := make([]interface{}, len(playerIDs)+3)
	for i, id := range playerIDs {
		placeholders[i] = "$" + strconv.Itoa(i+1)
		args[i] = id
	}
	args[len(playerIDs)] = competitionID
	args[len(playerIDs)+1] = "ACTIVE"
	args[len(playerIDs)+2] = time.Now()

	query := `
		UPDATE player_competitions
		SET status = $` + strconv.Itoa(len(playerIDs)+2) + `, competition_id = $` + strconv.Itoa(len(playerIDs)+1) + `, updated_at = $` + strconv.Itoa(len(playerIDs)+3) + `
		WHERE player_id IN (` + strings.Join(placeholders, ",") + `) AND status = 'WAITING'
	`
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		log.Printf("[Repository] Error updating player competitions: %v", err)
	}
	return err
}

func (r *Repository) AddScoreToPlayer(ctx context.Context, playerID string, score int) error {
	log.Printf("[Repository] Adding score %d to player %s", score, playerID)
	_, err := r.db.ExecContext(ctx, `
		UPDATE player_competitions pc
		SET score = score + $1, updated_at = NOW()
		FROM competitions c
		WHERE pc.player_id = $2
		  AND pc.status = 'ACTIVE'
		  AND pc.competition_id = c.competition_id
		  AND c.ends_at > NOW()
	`, score, playerID)
	if err != nil {
		log.Printf("[Repository] Error adding score: %v", err)
	}
	return err
}

func (r *Repository) CompleteFinishedCompetitions(ctx context.Context) error {
	// 1. Mark competitions as COMPLETED
	res, err := r.db.ExecContext(ctx, `
		UPDATE competitions
		SET status = 'COMPLETED'
		WHERE ends_at <= NOW() AND status = 'ACTIVE'
	`)
	if err != nil {
		log.Printf("[Repository] Error completing finished competitions: %v", err)
		return err
	}
	count, _ := res.RowsAffected()
	log.Printf("[Repository] Marked %d competitions as COMPLETED", count)

	// 2. Mark related player_competitions as COMPLETED
	_, err = r.db.ExecContext(ctx, `
		UPDATE player_competitions
		SET status = 'COMPLETED'
		WHERE competition_id IN (
			SELECT competition_id FROM competitions WHERE status = 'COMPLETED' AND ends_at <= NOW()
		) AND status = 'ACTIVE'
	`)
	if err != nil {
		log.Printf("[Repository] Error completing player_competitions: %v", err)
		return err
	}
	return nil
}

func (r *Repository) IsPlayerInWaitingQueue(ctx context.Context, playerID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM player_competitions WHERE player_id = $1 AND status = 'WAITING'
	`, playerID).Scan(&count)
	if err != nil {
		log.Printf("[Repository] Error checking waiting queue for player %s: %v", playerID, err)
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) RunMatchmakingTransactional(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	return fn(tx)
}

// Repository interface for dependency injection
// (should match the one in service)
type RepositoryInterface interface {
	CreatePlayer(ctx context.Context, player *model.Player) error
	GetPlayerByID(ctx context.Context, playerID string) (*model.Player, error)
	UpdatePlayer(ctx context.Context, player *model.Player) error

	CreateCompetition(ctx context.Context, comp *model.Competition) error
	GetCompetitionByID(ctx context.Context, competitionID string) (*model.Competition, error)
	UpdateCompetition(ctx context.Context, comp *model.Competition) error

	GetActiveCompetition(ctx context.Context) (*model.Competition, error)

	CreatePlayerCompetition(ctx context.Context, pc *model.PlayerCompetition) error
	GetPlayerCompetitionByID(ctx context.Context, id int) (*model.PlayerCompetition, error)
	UpdatePlayerCompetition(ctx context.Context, pc *model.PlayerCompetition) error

	GetLatestPlayerCompetition(ctx context.Context, playerID string) (*model.PlayerCompetition, error)
	GetLeaderboardByCompetitionID(ctx context.Context, competitionID string) ([]model.PlayerCompetition, error)
	GetActivePlayerCompetition(ctx context.Context, playerID string) (*model.PlayerCompetition, error)

	GetWaitingPlayers(ctx context.Context) ([]model.PlayerCompetition, error)
	UpdatePlayerCompetitionsToActive(ctx context.Context, playerIDs []string, competitionID uuid.UUID, endsAt time.Time) error

	AddScoreToPlayer(ctx context.Context, playerID string, score int) error

	CompleteFinishedCompetitions(ctx context.Context) error

	IsPlayerInWaitingQueue(ctx context.Context, playerID string) (bool, error)

	RunMatchmakingTransactional(ctx context.Context, fn func(tx *sql.Tx) error) error
}
