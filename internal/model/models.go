package model

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	PlayerID    string `db:"player_id"`
	Level       int    `db:"level"`
	CountryCode string `db:"country_code"`
}

type CompetitionStatus string

const (
	CompetitionActive    CompetitionStatus = "ACTIVE"
	CompetitionCompleted CompetitionStatus = "COMPLETED"
	CompetitionCancelled CompetitionStatus = "CANCELLED"
)

type Competition struct {
	CompetitionID uuid.UUID         `db:"competition_id"`
	StartedAt     time.Time         `db:"started_at"`
	EndsAt        time.Time         `db:"ends_at"`
	Level         int               `db:"level"`
	CountryCode   string            `db:"country_code"`
	Status        CompetitionStatus `db:"status"`
}

type PlayerStatus string

const (
	StatusWaiting   PlayerStatus = "WAITING"
	StatusActive    PlayerStatus = "ACTIVE"
	StatusCompleted PlayerStatus = "COMPLETED"
	StatusCancelled PlayerStatus = "CANCELLED"
)

type PlayerCompetition struct {
	ID            int          `db:"id"`
	PlayerID      string       `db:"player_id"`
	CompetitionID *uuid.UUID   `db:"competition_id"` // nullable
	Status        PlayerStatus `db:"status"`
	Score         int          `db:"score"`
	JoinedAt      time.Time    `db:"joined_at"`
	UpdatedAt     time.Time    `db:"updated_at"`
	Level         int          `db:"level"`
	CountryCode   string       `db:"country_code"`
}
