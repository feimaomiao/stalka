//go:build dbtest

package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/feimaomiao/stalka/internal/testdb"
	"github.com/feimaomiao/stalka/pandatypes"
	"github.com/h2non/gock"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	dbTestBaseURL = "https://api.pandascore.test/"
	seededMatches = 500
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) { testdb.RunMain(m, &pool) }

// seedMatches resets the DB to one game/league/series/tournament chain plus
// seededMatches idle matches with ids 1..seededMatches.
func seedMatches(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	for _, stmt := range []string{
		`TRUNCATE games, leagues, series, tournaments, matches, teams CASCADE`,
		`INSERT INTO games (id, name, slug) VALUES (1, 'Game', 'game')`,
		`INSERT INTO leagues (id, name, slug, game_id) VALUES (10, 'League', 'league', 1)`,
		`INSERT INTO series (id, name, slug, game_id, league_id) VALUES (100, 'Series', 'series', 1, 10)`,
		`INSERT INTO tournaments (id, name, slug, tier, game_id, league_id, serie_id)
			VALUES (1000, 'Tournament', 'tournament', 1, 1, 10, 100)`,
		fmt.Sprintf(`INSERT INTO matches (id, name, slug, finished, expected_start_time, actual_game_time,
			team1_id, team1_score, team2_id, team2_score, amount_of_games, game_id, league_id, series_id,
			tournament_id)
			SELECT g, 'Match', 'match', false, '2026-11-06 12:00:00', 0, 0, 0, 0, 0, 3, 1, 10, 100, 1000
			FROM generate_series(1, %d) g`, seededMatches),
	} {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func newDBClient(t *testing.T) *PandaClient {
	t.Helper()
	client := &PandaClient{
		BaseURL:     dbTestBaseURL,
		Pandasecret: "test_secret",
		Logger:      zap.NewNop().Sugar(),
		HTTPClient:  &http.Client{},
		DBConnector: dbtypes.New(pool),
		Ctx:         context.Background(),
	}
	gock.InterceptClient(client.HTTPClient)
	t.Cleanup(gock.Off)
	return client
}

// rowVersions returns each row's xmin, the transaction that wrote its current
// version. Postgres writes a new version on every UPDATE even when the values
// are unchanged, so a stable xmin is proof the row was left alone.
func rowVersions(t *testing.T, table string) map[int32]string {
	t.Helper()
	rows, err := pool.Query(context.Background(), `SELECT id, xmin::text FROM `+table)
	if err != nil {
		t.Fatalf("read versions: %v", err)
	}
	defer rows.Close()
	versions := make(map[int32]string)
	for rows.Next() {
		var id int32
		var xmin string
		if err = rows.Scan(&id, &xmin); err != nil {
			t.Fatalf("scan versions: %v", err)
		}
		versions[id] = xmin
	}
	return versions
}

func changedRows(before, after map[int32]string) map[int32]bool {
	changed := make(map[int32]bool)
	for id, xmin := range after {
		if before[id] != xmin {
			changed[id] = true
		}
	}
	return changed
}

func assertRowsChanged(t *testing.T, before, after map[int32]string, want ...int32) {
	t.Helper()
	changed := changedRows(before, after)
	for _, id := range want {
		if !changed[id] {
			t.Errorf("row %d was not rewritten, want it rewritten", id)
		}
		delete(changed, id)
	}
	if len(changed) != 0 {
		t.Errorf("%d unrelated rows rewritten, want 0", len(changed))
	}
}

func liveIDs(t *testing.T) map[int32]bool {
	t.Helper()
	rows, err := pool.Query(context.Background(), `SELECT id FROM matches WHERE is_live`)
	if err != nil {
		t.Fatalf("read live ids: %v", err)
	}
	defer rows.Close()
	ids := make(map[int32]bool)
	for rows.Next() {
		var id int32
		if err = rows.Scan(&id); err != nil {
			t.Fatalf("scan live id: %v", err)
		}
		ids[id] = true
	}
	return ids
}

func assertLive(t *testing.T, want ...int32) {
	t.Helper()
	got := liveIDs(t)
	if len(got) != len(want) {
		t.Errorf("live ids = %v, want %v", got, want)
	}
	for _, id := range want {
		if !got[id] {
			t.Errorf("match %d is not live, want live (live ids = %v)", id, got)
		}
	}
}

func runningMatchJSON(id int) string {
	return fmt.Sprintf(`{
		"id": %d, "name": "Running %d", "slug": "running-%d", "number_of_games": 3,
		"scheduled_at": "2026-11-06T12:00:00Z", "tournament_id": 1000,
		"videogame": {"id": 1}, "league": {"id": 10}, "serie": {"id": 100}, "tournament": {"id": 1000}
	}`, id, id, id)
}

// pollLives runs one GetLives against a mocked /matches/running listing ids.
func pollLives(t *testing.T, client *PandaClient, ids ...int) {
	t.Helper()
	matches := make([]string, 0, len(ids))
	for _, id := range ids {
		matches = append(matches, runningMatchJSON(id))
	}
	gock.New(dbTestBaseURL).
		Get("/matches/running/").
		Reply(http.StatusOK).
		BodyString("[" + strings.Join(matches, ",") + "]")
	if err := client.GetLives(); err != nil {
		t.Fatalf("GetLives: %v", err)
	}
}

func TestGetLivesLeavesIdleRowsAlone(t *testing.T) {
	seedMatches(t)
	client := newDBClient(t)

	before := rowVersions(t, "matches")
	pollLives(t, client, 1, 2)
	assertLive(t, 1, 2)
	assertRowsChanged(t, before, rowVersions(t, "matches"), 1, 2)
}

func TestGetLivesRepeatedPollWritesNothing(t *testing.T) {
	seedMatches(t)
	client := newDBClient(t)
	pollLives(t, client, 1, 2)

	before := rowVersions(t, "matches")
	pollLives(t, client, 1, 2)
	assertLive(t, 1, 2)
	assertRowsChanged(t, before, rowVersions(t, "matches"))
}

func TestGetLivesOnlyTouchesTransitions(t *testing.T) {
	seedMatches(t)
	client := newDBClient(t)
	pollLives(t, client, 1, 2)
	pollLives(t, client, 1, 2)

	// 1 leaves, 3 joins, and 2 keeps running with identical data.
	before := rowVersions(t, "matches")
	pollLives(t, client, 2, 3)
	assertLive(t, 2, 3)
	assertRowsChanged(t, before, rowVersions(t, "matches"), 1, 3)
}

func TestGetLivesEmptyPollClearsFlags(t *testing.T) {
	seedMatches(t)
	client := newDBClient(t)
	pollLives(t, client, 1, 2)

	before := rowVersions(t, "matches")
	pollLives(t, client)
	assertLive(t)
	assertRowsChanged(t, before, rowVersions(t, "matches"), 1, 2)
}

func TestGetLivesEmptyPollOnIdleTableWritesNothing(t *testing.T) {
	seedMatches(t)
	client := newDBClient(t)

	before := rowVersions(t, "matches")
	pollLives(t, client)
	assertLive(t)
	assertRowsChanged(t, before, rowVersions(t, "matches"))
}

// A nil slice reaches Postgres as NULL. Without the COALESCE in the query that
// either matches nothing or writes NULL into a NOT NULL column.
func TestSyncMatchesIsLiveNilClearsFlags(t *testing.T) {
	seedMatches(t)
	client := newDBClient(t)
	pollLives(t, client, 1, 2)

	if _, err := dbtypes.New(pool).SyncMatchesIsLive(context.Background(), nil); err != nil {
		t.Fatalf("SyncMatchesIsLive: %v", err)
	}
	assertLive(t)
}

// TestUpsertsSkipUnchangedRows pins every upsert to the same contract: writing
// identical data leaves the row version alone, and a real change still lands.
func TestUpsertsSkipUnchangedRows(t *testing.T) {
	start := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	match := pandatypes.MatchRow{
		ID: 9001, Name: "Upsert", Slug: "upsert", ExpectedStartTime: start, AmountOfGames: 3,
		GameID: 1, LeagueID: 10, SerieID: 100, TournamentID: 1000,
	}
	renamedMatch := match
	renamedMatch.Name = "Renamed"
	scoredMatch := match
	scoredMatch.Team1Score = 2
	canceledMatch := match
	canceledMatch.Status = pandatypes.MatchStatusCanceled

	cases := []struct {
		name    string
		table   string
		id      int32
		row     pandatypes.RowLike
		changed pandatypes.RowLike
	}{
		{
			name: "game", table: "games", id: 2,
			row:     pandatypes.GameRow{ID: 2, Name: "Game 2", Slug: "game-2"},
			changed: pandatypes.GameRow{ID: 2, Name: "Renamed", Slug: "game-2"},
		},
		{
			name: "league", table: "leagues", id: 20,
			row:     pandatypes.LeagueRow{ID: 20, Name: "League 2", Slug: "league-2", GameID: 1},
			changed: pandatypes.LeagueRow{ID: 20, Name: "League 2", Slug: "league-2", GameID: 1, ImageLink: "img"},
		},
		{
			name: "series", table: "series", id: 200,
			row:     pandatypes.SeriesRow{ID: 200, Name: "Series 2", Slug: "series-2", GameID: 1, LeagueID: 10},
			changed: pandatypes.SeriesRow{ID: 200, Name: "Renamed", Slug: "series-2", GameID: 1, LeagueID: 10},
		},
		{
			name: "tournament", table: "tournaments", id: 2000,
			row: pandatypes.TournamentRow{
				ID: 2000, Name: "T2", Slug: "t-2", Tier: 1, GameID: 1, LeagueID: 10, SerieID: 100,
			},
			changed: pandatypes.TournamentRow{
				ID: 2000, Name: "T2", Slug: "t-2", Tier: 2, GameID: 1, LeagueID: 10, SerieID: 100,
			},
		},
		{name: "match name", table: "matches", id: 9001, row: match, changed: renamedMatch},
		{name: "match score", table: "matches", id: 9001, row: match, changed: scoredMatch},
		{name: "match status", table: "matches", id: 9001, row: match, changed: canceledMatch},
		{
			name: "team", table: "teams", id: 700,
			row:     pandatypes.TeamRow{ID: 700, Name: "Team", Slug: "team", Acronym: "T", GameID: 1},
			changed: pandatypes.TeamRow{ID: 700, Name: "Team", Slug: "team", Acronym: "TM", GameID: 1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seedMatches(t)
			ctx := context.Background()
			queries := dbtypes.New(pool)
			if err := tc.row.WriteToDB(ctx, queries); err != nil {
				t.Fatalf("first write: %v", err)
			}

			before := rowVersions(t, tc.table)
			if err := tc.row.WriteToDB(ctx, queries); err != nil {
				t.Fatalf("identical write: %v", err)
			}
			assertRowsChanged(t, before, rowVersions(t, tc.table))

			if err := tc.changed.WriteToDB(ctx, queries); err != nil {
				t.Fatalf("changed write: %v", err)
			}
			assertRowsChanged(t, before, rowVersions(t, tc.table), tc.id)
		})
	}
}

// TestSyncMatchesIsLiveReturnsFlips pins what GetLives relies on to spot a match
// that just ended: both directions come back, and untouched rows don't.
func TestSyncMatchesIsLiveReturnsFlips(t *testing.T) {
	seedMatches(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `UPDATE matches SET is_live = true WHERE id IN (1, 2)`); err != nil {
		t.Fatalf("seed live rows: %v", err)
	}

	rows, err := dbtypes.New(pool).SyncMatchesIsLive(ctx, []int32{2, 3})
	if err != nil {
		t.Fatalf("SyncMatchesIsLive: %v", err)
	}
	got := make(map[int32]bool)
	for _, row := range rows {
		got[row.ID] = row.IsLive
	}
	want := map[int32]bool{1: false, 3: true}
	if len(got) != len(want) || got[1] != want[1] || got[3] != want[3] {
		t.Fatalf("flipped rows = %v, want %v", got, want)
	}
}

func TestGetMatchesToReconcile(t *testing.T) {
	const (
		recent = `NOW() - INTERVAL '1 day'`
		older  = `NOW() - INTERVAL '30 days'`
		soon   = `NOW() - INTERVAL '30 minutes'`
		future = `NOW() + INTERVAL '1 day'`
	)
	cases := []struct {
		name   string
		status string
		start  string
		live   bool
		want   bool
	}{
		{name: "no status yet", status: `NULL`, start: recent, want: true},
		{name: "no status, however old", status: `NULL`, start: older, want: true},
		{name: "stuck before starting", status: `'not_started'`, start: recent, want: true},
		{name: "stuck running", status: `'running'`, start: recent, want: true},
		{name: "postponed", status: `'postponed'`, start: recent, want: true},
		{name: "unknown upstream status", status: `'new_status'`, start: recent, want: true},
		{name: "stuck past the age limit", status: `'not_started'`, start: older},
		{name: "finished", status: `'finished'`, start: recent},
		{name: "canceled", status: `'canceled'`, start: recent},
		{name: "removed", status: `'removed'`, start: recent},
		{name: "live", status: `NULL`, start: recent, live: true},
		{name: "inside the grace period", status: `'not_started'`, start: soon},
		{name: "future", status: `NULL`, start: future},
		{name: "no start time", status: `NULL`, start: `NULL`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seedMatches(t)
			ctx := context.Background()
			if _, err := pool.Exec(ctx, `UPDATE matches SET status = 'finished'`); err != nil {
				t.Fatalf("settle rows: %v", err)
			}
			stmt := fmt.Sprintf(`UPDATE matches SET status = %s, expected_start_time = %s, is_live = %t WHERE id = 1`,
				tc.status, tc.start, tc.live)
			if _, err := pool.Exec(ctx, stmt); err != nil {
				t.Fatalf("seed row: %v", err)
			}

			ids, err := dbtypes.New(pool).GetMatchesToReconcile(ctx, ReconcileBatch)
			if err != nil {
				t.Fatalf("GetMatchesToReconcile: %v", err)
			}
			if got := len(ids) == 1; got != tc.want {
				t.Fatalf("selected %v, want selected = %t", ids, tc.want)
			}
		})
	}
}

func TestGetMatchesToReconcileNewestFirst(t *testing.T) {
	seedMatches(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `UPDATE matches SET expected_start_time = NOW() - id * INTERVAL '1 day'`); err != nil {
		t.Fatalf("seed start times: %v", err)
	}

	ids, err := dbtypes.New(pool).GetMatchesToReconcile(ctx, 3)
	if err != nil {
		t.Fatalf("GetMatchesToReconcile: %v", err)
	}
	if fmt.Sprint(ids) != "[1 2 3]" {
		t.Fatalf("ids = %v, want the 3 most recent", ids)
	}
}

// A wrongly finished 0-0 row has to come back out of the finished state.
func TestMarkMatchRemovedClearsFinished(t *testing.T) {
	seedMatches(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `UPDATE matches SET finished = true, is_live = true WHERE id = 1`); err != nil {
		t.Fatalf("seed row: %v", err)
	}
	if err := dbtypes.New(pool).MarkMatchRemoved(ctx, 1); err != nil {
		t.Fatalf("MarkMatchRemoved: %v", err)
	}

	var status string
	var finished, isLive bool
	row := pool.QueryRow(ctx, `SELECT status, finished, is_live FROM matches WHERE id = 1`)
	if err := row.Scan(&status, &finished, &isLive); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if status != "removed" || finished || isLive {
		t.Fatalf("got status=%s finished=%t is_live=%t", status, finished, isLive)
	}
}

func readMatchState(t *testing.T, id int32) (revision int32, isLive bool) {
	t.Helper()
	row := pool.QueryRow(context.Background(), `SELECT revision, is_live FROM matches WHERE id = $1`, id)
	if err := row.Scan(&revision, &isLive); err != nil {
		t.Fatalf("read match: %v", err)
	}
	return revision, isLive
}

// TestMatchUpsertIgnoresOlderData pins the guard against a page fetched before
// the row's latest write, which would otherwise briefly undo a newer score.
func TestMatchUpsertIgnoresOlderData(t *testing.T) {
	seedMatches(t)
	ctx := context.Background()
	queries := dbtypes.New(pool)
	newer := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	base := pandatypes.MatchRow{
		ID: 9001, Name: "Upsert", Slug: "upsert", AmountOfGames: 3,
		GameID: 1, LeagueID: 10, SerieID: 100, TournamentID: 1000, Status: "running",
	}
	current := base
	current.ModifiedAt = newer
	current.Team1Score = 2
	if err := current.WriteToDB(ctx, queries); err != nil {
		t.Fatalf("write current: %v", err)
	}

	stale := base
	stale.ModifiedAt = newer.Add(-time.Minute)
	stale.Team1Score = 1
	if err := stale.WriteToDB(ctx, queries); err != nil {
		t.Fatalf("write stale: %v", err)
	}
	var score int32
	if err := pool.QueryRow(ctx, `SELECT team1_score FROM matches WHERE id = 9001`).Scan(&score); err != nil {
		t.Fatalf("read score: %v", err)
	}
	if score != 2 {
		t.Fatalf("score = %d, older data overwrote it", score)
	}

	fresh := base
	fresh.ModifiedAt = newer.Add(time.Minute)
	fresh.Team1Score = 3
	if err := fresh.WriteToDB(ctx, queries); err != nil {
		t.Fatalf("write fresh: %v", err)
	}
	if err := pool.QueryRow(ctx, `SELECT team1_score FROM matches WHERE id = 9001`).Scan(&score); err != nil {
		t.Fatalf("read score: %v", err)
	}
	if score != 3 {
		t.Fatalf("score = %d, newer data was dropped", score)
	}
}

func TestMatchRevisionCountsRealChanges(t *testing.T) {
	seedMatches(t)
	ctx := context.Background()
	queries := dbtypes.New(pool)
	match := pandatypes.MatchRow{
		ID: 9001, Name: "Upsert", Slug: "upsert", AmountOfGames: 3,
		GameID: 1, LeagueID: 10, SerieID: 100, TournamentID: 1000, Status: "not_started",
	}
	write := func(row pandatypes.MatchRow) {
		t.Helper()
		if err := row.WriteToDB(ctx, queries); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	write(match)
	write(match)
	if revision, _ := readMatchState(t, 9001); revision != 0 {
		t.Fatalf("revision = %d after identical write, want 0", revision)
	}

	// An upstream touch that changes nothing shown is not a revision.
	touched := match
	touched.ModifiedAt = time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	write(touched)
	if revision, _ := readMatchState(t, 9001); revision != 0 {
		t.Fatalf("revision = %d after a no-op touch, want 0", revision)
	}

	moved := match
	moved.ExpectedStartTime = time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	write(moved)
	if revision, _ := readMatchState(t, 9001); revision != 1 {
		t.Fatalf("revision = %d after a reschedule, want 1", revision)
	}

	if err := queries.MarkMatchRemoved(ctx, 9001); err != nil {
		t.Fatalf("MarkMatchRemoved: %v", err)
	}
	if err := queries.MarkMatchRemoved(ctx, 9001); err != nil {
		t.Fatalf("MarkMatchRemoved again: %v", err)
	}
	if revision, _ := readMatchState(t, 9001); revision != 2 {
		t.Fatalf("revision = %d after removal, want 2: repeats must not bump it", revision)
	}
}

func TestSettledMatchIsNotLive(t *testing.T) {
	cases := []struct {
		name string
		row  pandatypes.MatchRow
		want bool
	}{
		{name: "finished", row: pandatypes.MatchRow{Finished: true, Status: "finished"}},
		{name: "canceled", row: pandatypes.MatchRow{Status: pandatypes.MatchStatusCanceled}},
		{name: "still running", row: pandatypes.MatchRow{Status: "running"}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seedMatches(t)
			ctx := context.Background()
			if _, err := pool.Exec(ctx, `UPDATE matches SET is_live = true WHERE id = 1`); err != nil {
				t.Fatalf("seed: %v", err)
			}
			row := tc.row
			row.ID, row.Name, row.Slug, row.AmountOfGames = 1, "Match", "match", 3
			row.GameID, row.LeagueID, row.SerieID, row.TournamentID = 1, 10, 100, 1000
			if err := row.WriteToDB(ctx, dbtypes.New(pool)); err != nil {
				t.Fatalf("write: %v", err)
			}
			if _, isLive := readMatchState(t, 1); isLive != tc.want {
				t.Fatalf("is_live = %t, want %t", isLive, tc.want)
			}
		})
	}
}
