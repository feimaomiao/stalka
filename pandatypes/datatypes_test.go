package pandatypes

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nbio/st"

	"github.com/pashagolub/pgxmock/v4"
)

// tests SafeIntToInt32 function
func TestSafeIntToInt32(t *testing.T) {
	value := 1234567890
	expected := int32(1234567890)

	result, err := SafeIntToInt32(value)
	st.Expect(t, err, nil)
	st.Expect(t, result, expected)
	// Test overflow case
	result, err = SafeIntToInt32(2147483648)
	st.Expect(t, err, fmt.Errorf("value 2147483648 overflows int32 range [-2147483648, 2147483647]"))
	st.Expect(t, result, int32(0))
}

// tests ToRow function for all types
func TestToRow(t *testing.T) {

	t.Run("GameLike To Row", func(t *testing.T) {

		// open the file static/fetch_data/videogames.json with os.readfile
		data, err := os.ReadFile("../static/fetch_data/videogames.json")
		st.Assert(t, err, nil)

		var game GameLike
		err = json.Unmarshal(data, &game)
		st.Assert(t, err, nil)

		st.Assert(t, game.ID, 34)
		st.Assert(t, game.Name, "Mobile Legends: Bang Bang")
		st.Assert(t, game.Slug, "mlbb")

		// assert that gameR can be converted to GameRow
		gameR := game.ToRow().(GameRow)
		st.Assert(t, recover(), nil)

		// Assert the row data
		st.Assert(t, gameR.ID, game.ID)
		st.Assert(t, gameR.Name, game.Name)
		st.Assert(t, gameR.Slug, game.Slug)
	})

	t.Run("LeagueLike To Row", func(t *testing.T) {
		// open the file static/fetch_data/leagues.json with os.readfile
		data, err := os.ReadFile("../static/fetch_data/leagues.json")
		st.Assert(t, err, nil)

		var league LeagueLike
		err = json.Unmarshal(data, &league)
		st.Assert(t, err, nil)

		st.Assert(t, league.ID, 289)
		st.Assert(t, league.Name, "NA LCS")
		st.Assert(t, league.Slug, "league-of-legends-na-lcs")

		// assert that leagueR can be converted to LeagueRow
		leagueR := league.ToRow().(LeagueRow)
		st.Assert(t, recover(), nil)

		// Assert the row data
		st.Assert(t, leagueR.ID, league.ID)
		st.Assert(t, leagueR.Name, league.Name)
		st.Assert(t, leagueR.Slug, league.Slug)
		st.Assert(t, leagueR.GameID, league.Videogame.ID)
		st.Assert(t, leagueR.ImageLink, league.ImageURL)
	})

	t.Run("SeriesLike To Row", func(t *testing.T) {

		// Test series
		data, err := os.ReadFile("../static/fetch_data/series.json")
		st.Assert(t, err, nil)

		var series SeriesLike
		err = json.Unmarshal(data, &series)
		st.Assert(t, err, nil)

		st.Assert(t, series.ID, 346)
		st.Assert(t, series.Name, "")
		st.Assert(t, series.Slug, "league-of-legends-international-wildcard-msi-qualifier-2016")

		// assert that series can be converted to SeriesRow
		seriesR := series.ToRow().(SeriesRow)
		st.Assert(t, recover(), nil)

		// Assert the row data
		st.Assert(t, seriesR.ID, series.ID)
		st.Assert(t, seriesR.Name, series.Name)
		st.Assert(t, seriesR.Slug, series.Slug)
		st.Assert(t, seriesR.GameID, series.Videogame.ID)
		st.Assert(t, seriesR.LeagueID, series.League.ID)
	})

	t.Run("TournamentLike To Row", func(t *testing.T) {
		// Test tournaments
		data, err := os.ReadFile("../static/fetch_data/tournaments.json")
		st.Assert(t, err, nil)

		var tournament TournamentLike
		err = json.Unmarshal(data, &tournament)
		st.Assert(t, err, nil)

		st.Assert(t, tournament.ID, 17283)
		st.Assert(t, tournament.Name, "Group Stage")
		st.Assert(t, tournament.Slug, "the-international-2025-group-stage")

		// assert that tournament can be converted to TournamentRow
		tournamentR := tournament.ToRow().(TournamentRow)
		st.Assert(t, recover(), nil)

		// Assert the row data
		st.Assert(t, tournamentR.ID, tournament.ID)
		st.Assert(t, tournamentR.Name, tournament.Name)
		st.Assert(t, tournamentR.Slug, tournament.Slug)
		st.Assert(t, tournamentR.GameID, tournament.Videogame.ID)
		st.Assert(t, tournamentR.SerieID, tournament.Serie.ID)
		st.Assert(t, tournamentR.LeagueID, tournament.League.ID)

		tournament.Tier = "A"
		st.Assert(t, tournament.ToRow().(TournamentRow).Tier, 2)

		tournament.Tier = "B"
		st.Assert(t, tournament.ToRow().(TournamentRow).Tier, 3)

		tournament.Tier = "C"
		st.Assert(t, tournament.ToRow().(TournamentRow).Tier, 4)

		tournament.Tier = "D"
		st.Assert(t, tournament.ToRow().(TournamentRow).Tier, 5)

		tournament.Tier = "Whatever"
		st.Assert(t, tournament.ToRow().(TournamentRow).Tier, 6)
	})

	t.Run("TeamLike to Row", func(t *testing.T) {
		// Test teams
		data, err := os.ReadFile("../static/fetch_data/teams.json")
		st.Assert(t, err, nil)

		var team TeamLike
		err = json.Unmarshal(data, &team)
		st.Assert(t, err, nil)

		st.Assert(t, team.ID, 127652)
		st.Assert(t, team.Name, "Ares Gaming")
		st.Assert(t, team.Slug, "ares-gaming")

		// assert that team can be converted to TeamRow
		teamR := team.ToRow().(TeamRow)
		st.Assert(t, recover(), nil)

		// Assert the row data
		st.Assert(t, teamR.ID, team.ID)
		st.Assert(t, teamR.Name, team.Name)
		st.Assert(t, teamR.Slug, team.Slug)
		st.Assert(t, teamR.GameID, team.CurrentVideogame.ID)
		st.Assert(t, teamR.Acronym, team.Acronym)
		st.Assert(t, teamR.ImageLink, team.ImageURL)
	})

	t.Run("MatchLike to Row", func(t *testing.T) {

		// Test matches
		data, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var match MatchLike
		err = json.Unmarshal(data, &match)
		st.Assert(t, err, nil)

		st.Assert(t, match.ID, 21655)
		st.Assert(t, match.Name, "Qf 2")
		st.Assert(t, match.Slug, "team-lolpro-vs-area-of-effect-esports-2014-06-01")

		// assert that match can be converted to MatchRow
		matchR := match.ToRow().(MatchRow)
		st.Assert(t, recover(), nil)

		// Assert the row data
		st.Assert(t, matchR.ID, match.ID)
		st.Assert(t, matchR.Name, match.Name)
		st.Assert(t, matchR.Slug, match.Slug)
		st.Assert(t, matchR.GameID, match.Videogame.ID)
		st.Assert(t, matchR.LeagueID, match.League.ID)
		st.Assert(t, matchR.SerieID, match.Serie.ID)
		st.Assert(t, matchR.TournamentID, match.Tournament.ID)
	})

}

// tests writetodb for all types
func TestWriteToDB(t *testing.T) {
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQuery := dbtypes.New(mockDB)
	t.Run("Write Game", func(t *testing.T) {
		// open the file static/fetch_data/videogames.json with os.readfile
		data, err := os.ReadFile("../static/fetch_data/videogames.json")
		st.Assert(t, err, nil)

		var game GameLike
		err = json.Unmarshal(data, &game)
		st.Assert(t, err, nil)
		row := game.ToRow()
		mockDB.ExpectExec("INSERT INTO games").
			WithArgs(
				int32(game.ID),
				game.Name,
				pgtype.Text{
					String: game.Slug,
					Valid:  true,
				}).
			WillReturnResult(
				pgxmock.NewResult("INSERT", 1),
			)
		err = row.WriteToDB(t.Context(), mockQuery)
		st.Expect(t, err, nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)

		// this following test should fail because the game ID is out of range
		game.ID = math.MaxInt32 + 1
		err = game.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
	})

	t.Run("Write League", func(t *testing.T) {
		// open the file static/fetch_data/leagues.json with os.readfile
		data, err := os.ReadFile("../static/fetch_data/leagues.json")
		st.Assert(t, err, nil)

		var league LeagueLike
		err = json.Unmarshal(data, &league)
		st.Assert(t, err, nil)
		row := league.ToRow()
		mockDB.ExpectExec("INSERT INTO leagues").
			WithArgs(
				int32(league.ID),
				league.Name,
				pgtype.Text{
					String: league.Slug,
					Valid:  true,
				},
				pgtype.Text{
					String: league.ImageURL,
					Valid:  true,
				},
				int32(league.Videogame.ID),
			).
			WillReturnResult(
				pgxmock.NewResult("INSERT", 1),
			)
		err = row.WriteToDB(t.Context(), mockQuery)
		st.Expect(t, err, nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)

		// this should fail because the game ID is out of range
		league.Videogame.ID = math.MaxInt32 + 1
		err = league.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		league.Videogame.ID = 1

		//this should fail because the league ID is out of range
		league.ID = math.MaxInt32 + 1
		err = league.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		league.ID = 1
	})

	t.Run("Write Series", func(t *testing.T) {
		// open the file static/fetch_data/series.json with os.readfile
		data, err := os.ReadFile("../static/fetch_data/series.json")
		st.Assert(t, err, nil)

		var series SeriesLike
		err = json.Unmarshal(data, &series)
		st.Assert(t, err, nil)
		row := series.ToRow()

		mockDB.ExpectExec("INSERT INTO series").
			WithArgs(
				int32(series.ID),
				series.Name,
				pgtype.Text{
					String: series.Slug,
					Valid:  true,
				},
				int32(series.Videogame.ID),
				int32(series.League.ID),
			).
			WillReturnResult(
				pgxmock.NewResult("INSERT", 1),
			)
		err = row.WriteToDB(t.Context(), mockQuery)
		st.Expect(t, err, nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)

		//this should fail because gameID is out of range
		series.Videogame.ID = math.MaxInt32 + 1
		err = series.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		series.Videogame.ID = 1

		// this should fail because the league ID is out of range
		series.League.ID = math.MaxInt32 + 1
		err = series.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		series.League.ID = 1
		// this following test should fail because the series ID is out of range
		series.ID = math.MaxInt32 + 1
		err = series.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		series.ID = 1
	})

	t.Run("Write Tournament", func(t *testing.T) {
		// open the file static/fetch_data/tournaments.json with os.readfile
		data, err := os.ReadFile("../static/fetch_data/tournaments.json")
		st.Assert(t, err, nil)

		var tournament TournamentLike
		err = json.Unmarshal(data, &tournament)
		st.Assert(t, err, nil)
		row := tournament.ToRow()
		mockDB.ExpectExec("INSERT INTO tournaments").
			WithArgs(
				int32(tournament.ID),
				tournament.Name,
				pgtype.Text{
					String: tournament.Slug,
					Valid:  true,
				},
				pgtype.Int4{
					Int32: 1,
					Valid: true,
				},
				int32(tournament.Videogame.ID),
				int32(tournament.League.ID),
				int32(tournament.Serie.ID),
			).
			WillReturnResult(
				pgxmock.NewResult("INSERT", 1),
			)
		err = row.WriteToDB(t.Context(), mockQuery)
		st.Expect(t, err, nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)

		// this following test should fail because the tournament ID is out of range
		tournament.ID = math.MaxInt32 + 1
		err = tournament.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		tournament.ID = 1

		//this should fail because gameID is out of range
		tournament.Videogame.ID = math.MaxInt32 + 1
		err = tournament.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		tournament.Videogame.ID = 1
		temp := tournament.League.ID

		// this should fail because the league ID is out of range
		tournament.League.ID = math.MaxInt32 + 1
		err = tournament.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)

		tournament.League.ID = temp

		//this should fail because serie ID is out of range
		tournament.Serie.ID = math.MaxInt32 + 1
		err = tournament.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
	})

	t.Run("Write Match", func(t *testing.T) {
		// open the file static/fetch_data/matches.json with os.readfile
		data, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var match MatchLike
		err = json.Unmarshal(data, &match)
		st.Assert(t, err, nil)
		row := match.ToRow()

		mockDB.ExpectExec("INSERT INTO matches").
			WithArgs(
				int32(match.ID),
				match.Name,
				pgtype.Text{
					String: match.Slug,
					Valid:  true,
				},
				match.EndAt != time.Time{},
				pgtype.Timestamp{Time: match.BeginAt, Valid: true},
				match.EndAt.Sub(match.BeginAt).Seconds()/float64(match.NumberOfGames),
				int32(match.Opponents[0].Opponent.ID),
				int32(match.Results[0].Score),
				int32(match.Opponents[1].Opponent.ID),
				int32(match.Results[1].Score),
				int32(match.NumberOfGames),
				int32(match.Videogame.ID),
				int32(match.LeagueID),
				int32(match.SerieID),
				int32(match.TournamentID),
				pgtype.Text{String: "", Valid: false},
				pgtype.Text{String: match.Status, Valid: match.Status != ""},
				pgtype.Timestamp{Time: match.ModifiedAt, Valid: !match.ModifiedAt.IsZero()},
				match.Forfeit,
				int32(match.WinnerID),
			).
			WillReturnResult(
				pgxmock.NewResult("INSERT", 1),
			)

		err = row.WriteToDB(t.Context(), mockQuery)
		st.Expect(t, err, nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)

		//this should fail because the match id is out of range
		temp := match.ID
		match.ID = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.ID = temp

		// the following tests fail because the team id and score are out of range
		temp = match.Opponents[0].Opponent.ID
		match.Opponents[0].Opponent.ID = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.Opponents[0].Opponent.ID = temp

		temp = match.Results[0].Score
		match.Results[0].Score = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.Results[0].Score = temp

		temp = match.Opponents[1].Opponent.ID
		match.Opponents[1].Opponent.ID = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.Opponents[1].Opponent.ID = temp

		temp = match.Results[1].Score
		match.Results[1].Score = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.Results[1].Score = temp

		//the following should fail because the amount of games is out of range
		temp = match.NumberOfGames
		match.NumberOfGames = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.NumberOfGames = temp

		// the following should fail because any of the IDs are out of ragne
		temp = match.Videogame.ID
		match.Videogame.ID = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.Videogame.ID = temp

		temp = match.League.ID
		match.League.ID = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.League.ID = temp

		temp = match.Serie.ID
		match.Serie.ID = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.Serie.ID = temp

		temp = match.Tournament.ID
		match.Tournament.ID = math.MaxInt32 + 1
		err = match.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		match.Tournament.ID = temp
	})

	t.Run("Write Team", func(t *testing.T) {
		// open the file static/fetch_data/teams.json with os.readfile
		data, err := os.ReadFile("../static/fetch_data/teams.json")
		st.Assert(t, err, nil)

		var team TeamLike
		err = json.Unmarshal(data, &team)
		st.Assert(t, err, nil)
		row := team.ToRow()
		mockDB.ExpectExec("INSERT INTO teams").
			WithArgs(
				int32(team.ID),
				team.Name,
				pgtype.Text{
					String: team.Slug,
					Valid:  team.Slug != "",
				},
				pgtype.Text{
					String: team.Acronym,
					Valid:  team.Acronym != "",
				},
				pgtype.Text{
					String: team.ImageURL,
					Valid:  team.ImageURL != "",
				},
				int32(team.CurrentVideogame.ID),
			).
			WillReturnResult(
				pgxmock.NewResult("INSERT", 1),
			)
		err = row.WriteToDB(t.Context(), mockQuery)
		st.Expect(t, err, nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)

		// this following test should fail because the team ID is out of range
		team.ID = math.MaxInt32 + 1
		err = team.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
		team.ID = 1

		//the following test should fail because the gameID is out of range
		team.CurrentVideogame.ID = math.MaxInt32 + 1
		err = team.ToRow().WriteToDB(t.Context(), mockQuery)
		st.Reject(t, err, nil)
	})
}

func TestMatchIsFinished(t *testing.T) {
	ended := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		status string
		endAt  time.Time
		want   bool
	}{
		{name: "not started", status: "not_started"},
		{name: "running", status: "running"},
		{name: "postponed", status: "postponed"},
		{name: "canceled", status: MatchStatusCanceled},
		{name: "canceled beats a stray end time", status: MatchStatusCanceled, endAt: ended},
		{name: "finished", status: MatchStatusFinished, endAt: ended, want: true},
		{name: "finished with no end time", status: MatchStatusFinished, want: true},
		{name: "end time with no status", endAt: ended, want: true},
		{name: "nothing known"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			match := MatchLike{Status: tc.status, EndAt: tc.endAt, NumberOfGames: 3}
			row, ok := match.ToRow().(MatchRow)
			st.Assert(t, ok, true)
			st.Expect(t, row.Finished, tc.want)
			st.Expect(t, row.Status, tc.status)
		})
	}
}

// matchWithOpponents builds a match from opponent ids and {team_id, score} pairs.
func matchWithOpponents(t *testing.T, opponents []int, results [][2]int) MatchRow {
	t.Helper()
	var match MatchLike
	for _, id := range opponents {
		var decoded MatchLike
		body := fmt.Sprintf(`{"opponents":[{"opponent":{"id":%d}}]}`, id)
		st.Assert(t, json.Unmarshal([]byte(body), &decoded), nil)
		match.Opponents = append(match.Opponents, decoded.Opponents...)
	}
	for _, result := range results {
		var decoded MatchLike
		body := fmt.Sprintf(`{"results":[{"team_id":%d,"score":%d}]}`, result[0], result[1])
		st.Assert(t, json.Unmarshal([]byte(body), &decoded), nil)
		match.Results = append(match.Results, decoded.Results...)
	}
	row, ok := match.ToRow().(MatchRow)
	st.Assert(t, ok, true)
	return row
}

func TestMatchOpponentSlots(t *testing.T) {
	cases := []struct {
		name      string
		opponents []int
		results   [][2]int
		want      [4]int // team1 id, team1 score, team2 id, team2 score
	}{
		{name: "both TBD"},
		{name: "both known", opponents: []int{7, 8}, results: [][2]int{{7, 2}, {8, 1}}, want: [4]int{7, 2, 8, 1}},
		{name: "one known", opponents: []int{7}, results: [][2]int{{7, 0}}, want: [4]int{7, 0, 0, 0}},
		{name: "one known with no results", opponents: []int{7}, want: [4]int{7, 0, 0, 0}},
		{name: "two known with no results", opponents: []int{7, 8}, want: [4]int{7, 0, 8, 0}},
		{name: "two known with short results", opponents: []int{7, 8}, results: [][2]int{{7, 1}}, want: [4]int{7, 1, 8, 0}},
		{name: "scores follow the team, not the order", opponents: []int{7, 8}, results: [][2]int{{8, 1}, {7, 2}}, want: [4]int{7, 2, 8, 1}},
		{name: "player results have no team id", opponents: []int{7, 8}, results: [][2]int{{0, 2}, {0, 1}}, want: [4]int{7, 2, 8, 1}},
		{name: "more than two sides is not a versus match", opponents: []int{7, 8, 9}, results: [][2]int{{7, 1}, {8, 2}, {9, 3}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			row := matchWithOpponents(t, tc.opponents, tc.results)
			st.Expect(t, [4]int{row.Team1ID, row.Team1Score, row.Team2ID, row.Team2Score}, tc.want)
		})
	}
}

func TestMatchActualGameTimeStaysFinite(t *testing.T) {
	begin := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	end := begin.Add(3 * time.Hour)
	cases := []struct {
		name       string
		begin, end time.Time
		games      int
		want       float64
	}{
		{name: "played", begin: begin, end: end, games: 3, want: 3600},
		{name: "no games", begin: begin, end: end, games: 0},
		{name: "no begin", end: end, games: 3},
		{name: "no end", begin: begin, games: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			match := MatchLike{BeginAt: tc.begin, EndAt: tc.end, NumberOfGames: tc.games}
			row, ok := match.ToRow().(MatchRow)
			st.Assert(t, ok, true)
			st.Expect(t, row.ActualGameTime, tc.want)
		})
	}
}

func TestMatchCarriesForfeitAndModifiedAt(t *testing.T) {
	var match MatchLike
	body := `{"status": "canceled", "forfeit": true, "winner_id": 7, "modified_at": "2026-11-06T12:00:00Z"}`
	st.Assert(t, json.Unmarshal([]byte(body), &match), nil)
	row, ok := match.ToRow().(MatchRow)
	st.Assert(t, ok, true)
	st.Expect(t, row.Forfeit, true)
	st.Expect(t, row.WinnerID, 7)
	st.Expect(t, row.ModifiedAt, time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC))
	st.Expect(t, row.Finished, false)
}
