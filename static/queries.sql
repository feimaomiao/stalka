-- name: InsertToGames :exec
INSERT INTO games (id, name, slug) VALUES ($1, $2, $3);

-- name: InsertToLeagues :exec
INSERT INTO leagues (id, name, slug, game_id, image_link) VALUES ($1, $2, $3, $4, $5);

-- name: InsertToSeries :exec
INSERT INTO series (id, name, slug, game_id, league_id) VALUES ($1, $2, $3, $4, $5);

-- name: InsertToTournaments :exec
INSERT INTO tournaments (id,name, slug,tier, game_id, serie_id, league_id) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: InsertToMatches :exec
INSERT INTO matches (id, name, slug, finished, expected_start_time, team1_id, team1_score, team2_id, team2_score, amount_of_games, game_id, league_id, series_id, tournament_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: InsertToTeams :exec
INSERT INTO teams (id, name, slug, acronym, image_link, game_id) VALUES ($1, $2, $3, $4, $5, $6);