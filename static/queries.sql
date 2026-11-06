-- name: InsertToGames :exec
INSERT INTO games (id, name, slug) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug
WHERE (
    games.name, games.slug
) IS DISTINCT FROM (
    EXCLUDED.name, EXCLUDED.slug
);

-- name: InsertToLeagues :exec
INSERT INTO leagues (id, name, slug, image_link, game_id) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    image_link = EXCLUDED.image_link,
    game_id = EXCLUDED.game_id
WHERE (
    leagues.name, leagues.slug, leagues.image_link, leagues.game_id
) IS DISTINCT FROM (
    EXCLUDED.name, EXCLUDED.slug, EXCLUDED.image_link, EXCLUDED.game_id
);

-- name: InsertToSeries :exec
INSERT INTO series (id, name, slug, game_id, league_id) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    game_id = EXCLUDED.game_id,
    league_id = EXCLUDED.league_id
WHERE (
    series.name, series.slug, series.game_id, series.league_id
) IS DISTINCT FROM (
    EXCLUDED.name, EXCLUDED.slug, EXCLUDED.game_id, EXCLUDED.league_id
);

-- name: InsertToTournaments :exec
INSERT INTO tournaments (id,name, slug,tier, game_id, league_id, serie_id) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    tier = EXCLUDED.tier,
    game_id = EXCLUDED.game_id,
    league_id = EXCLUDED.league_id,
    serie_id = EXCLUDED.serie_id
WHERE (
    tournaments.name, tournaments.slug, tournaments.tier, tournaments.game_id,
    tournaments.league_id, tournaments.serie_id
) IS DISTINCT FROM (
    EXCLUDED.name, EXCLUDED.slug, EXCLUDED.tier, EXCLUDED.game_id, EXCLUDED.league_id,
    EXCLUDED.serie_id
);

-- name: InsertToMatches :exec
INSERT INTO matches (id, name, slug, finished, expected_start_time, actual_game_time, team1_id, team1_score, team2_id, team2_score, amount_of_games, game_id, league_id, series_id, tournament_id, stream_url) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    finished = EXCLUDED.finished,
    expected_start_time = EXCLUDED.expected_start_time,
    actual_game_time = EXCLUDED.actual_game_time,
    team1_id = EXCLUDED.team1_id,
    team1_score = EXCLUDED.team1_score,
    team2_id = EXCLUDED.team2_id,
    team2_score = EXCLUDED.team2_score,
    amount_of_games = EXCLUDED.amount_of_games,
    game_id = EXCLUDED.game_id,
    league_id = EXCLUDED.league_id,
    series_id = EXCLUDED.series_id,
    tournament_id = EXCLUDED.tournament_id,
    stream_url = EXCLUDED.stream_url
WHERE (
    matches.name, matches.slug, matches.finished, matches.expected_start_time,
    matches.actual_game_time, matches.team1_id, matches.team1_score, matches.team2_id,
    matches.team2_score, matches.amount_of_games, matches.game_id, matches.league_id,
    matches.series_id, matches.tournament_id, matches.stream_url
) IS DISTINCT FROM (
    EXCLUDED.name, EXCLUDED.slug, EXCLUDED.finished, EXCLUDED.expected_start_time,
    EXCLUDED.actual_game_time, EXCLUDED.team1_id, EXCLUDED.team1_score, EXCLUDED.team2_id,
    EXCLUDED.team2_score, EXCLUDED.amount_of_games, EXCLUDED.game_id, EXCLUDED.league_id,
    EXCLUDED.series_id, EXCLUDED.tournament_id, EXCLUDED.stream_url
);

-- name: InsertToTeams :exec
INSERT INTO teams (id, name, slug, acronym, image_link, game_id) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    acronym = EXCLUDED.acronym,
    image_link = EXCLUDED.image_link,
    game_id = EXCLUDED.game_id
WHERE (
    teams.name, teams.slug, teams.acronym, teams.image_link, teams.game_id
) IS DISTINCT FROM (
    EXCLUDED.name, EXCLUDED.slug, EXCLUDED.acronym, EXCLUDED.image_link, EXCLUDED.game_id
);

-- name: GameExist :one
SELECT COUNT(*) FROM games WHERE id = $1;

-- name: LeagueExist :one
SELECT COUNT(*) FROM leagues WHERE id = $1;

-- name: SeriesExist :one
SELECT COUNT(*) FROM series WHERE id = $1;

-- name: TournamentExist :one
SELECT COUNT(*) FROM tournaments WHERE id = $1;

-- name: MatchExist :one
SELECT COUNT(*) FROM matches WHERE id = $1;

-- name: TeamExist :one
SELECT COUNT(*) FROM teams WHERE id = $1;


-- name: GetAllGames :many
SELECT id, name, slug FROM games WHERE (id != 14) ORDER BY id ASC;

-- name: GetSeriesByGameID :many
SELECT id, name, slug, game_id, league_id FROM series WHERE game_id = $1 ORDER BY name ASC;

-- name: GetLeaguesByGameID :many
SELECT l.*
FROM LEAGUES l
LEFT JOIN TOURNAMENTS t ON l.id = t.league_id
WHERE l.game_id = $1
GROUP BY l.id, l.name, l.slug, l.game_id, l.image_link
ORDER BY MIN(t.tier) ASC, l.name ASC;

-- SyncMatchesIsLive sets is_live to true for exactly the given ids. Postgres
-- writes a new row version for every row an UPDATE matches, even when the value
-- is unchanged, so the WHERE must skip rows that are already correct. The
-- upserts above carry the same IS DISTINCT FROM guard for the same reason.
-- COALESCE matters: pgx sends a nil slice as NULL, and id = ANY(NULL) is NULL.
-- name: SyncMatchesIsLive :exec
UPDATE MATCHES SET is_live = (id = ANY(COALESCE(sqlc.arg(live_ids)::int[], '{}')))
WHERE is_live IS DISTINCT FROM (id = ANY(COALESCE(sqlc.arg(live_ids)::int[], '{}')));
