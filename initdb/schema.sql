-- Enum for player status
CREATE TYPE player_status AS ENUM ('WAITING', 'ACTIVE', 'COMPLETED', 'CANCELLED');

-- Players table
CREATE TABLE IF NOT EXISTS players (
    player_id      TEXT PRIMARY KEY,
    level          INT NOT NULL,
    country_code   TEXT
);

-- Competitions table
CREATE TABLE IF NOT EXISTS competitions (
    competition_id UUID PRIMARY KEY,
    started_at     TIMESTAMP NOT NULL,
    ends_at        TIMESTAMP NOT NULL,
    level          INT,
    country_code   TEXT,
    status         TEXT NOT NULL DEFAULT 'ACTIVE'
);

-- Player competitions table
CREATE TABLE IF NOT EXISTS player_competitions (
    id             SERIAL PRIMARY KEY,
    player_id      TEXT REFERENCES players(player_id),
    competition_id UUID REFERENCES competitions(competition_id),
    status         player_status NOT NULL,
    score          INT DEFAULT 0,
    joined_at      TIMESTAMP NOT NULL,
    updated_at     TIMESTAMP NOT NULL,
    level          INT NOT NULL,
    country_code   TEXT
);

CREATE INDEX IF NOT EXISTS idx_player_competitions_status ON player_competitions(status);
CREATE INDEX IF NOT EXISTS idx_player_competitions_competition_id ON player_competitions(competition_id);
CREATE INDEX IF NOT EXISTS idx_player_competitions_player_id ON player_competitions(player_id); 