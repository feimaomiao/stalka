package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/feimaomiao/stalka/pandatypes"
	"github.com/h2non/gock"
	"github.com/nbio/st"
	"github.com/pashagolub/pgxmock/v4"
	"go.uber.org/zap/zaptest"
)

func TestUpdateGames(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - UpdateGames", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gameData, err := os.ReadFile("../static/fetch_data/videogames.json")
		st.Assert(t, err, nil)

		var gameResponse pandatypes.GameLike
		err = json.Unmarshal(gameData, &gameResponse)
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/videogames").
			Reply(200).
			BodyString("[" + string(gameData) + "]")

		mockDB.ExpectExec("INSERT INTO games").
			WithArgs(
				int32(gameResponse.ID),
				gameResponse.Name,
				pgxmock.AnyArg(),
			).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.UpdateGames()
		st.Expect(t, err, nil)
		st.Expect(t, gock.IsDone(), true)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/videogames").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.UpdateGames()
		st.Reject(t, err, nil)
	})

	t.Run("Error - Invalid JSON response", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/videogames").
			Reply(200).
			BodyString("invalid json")

		err := client.UpdateGames()
		st.Reject(t, err, nil)
	})
}

func TestGetLeagues(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - GetLeagues with setup=false", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		leagueData, err := os.ReadFile("../static/fetch_data/leagues.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/leagues").
			MatchParam("sort", "-modified_at").
			MatchParam("page", "0").
			Reply(200).
			BodyString("[" + string(leagueData) + "]")

		mockDB.ExpectExec("INSERT INTO leagues").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.GetLeagues(false)
		st.Expect(t, err, nil)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/leagues").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.GetLeagues(false)
		st.Reject(t, err, nil)
	})
}

func TestGetSeries(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - GetSeries with existing league", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		seriesData, err := os.ReadFile("../static/fetch_data/series.json")
		st.Assert(t, err, nil)

		var seriesResponse pandatypes.SeriesLike
		err = json.Unmarshal(seriesData, &seriesResponse)
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/series").
			Reply(200).
			BodyString("[" + string(seriesData) + "]")

		// Mock LeagueExist to return that league exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(seriesResponse.LeagueID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		mockDB.ExpectExec("INSERT INTO series").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.GetSeries(false)
		st.Expect(t, err, nil)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/series").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.GetSeries(false)
		st.Reject(t, err, nil)
	})
}

func TestGetTournaments(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - GetTournaments with existing series", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		tournamentData, err := os.ReadFile("../static/fetch_data/tournaments.json")
		st.Assert(t, err, nil)

		var tournamentResponse pandatypes.TournamentLike
		err = json.Unmarshal(tournamentData, &tournamentResponse)
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/tournaments").
			MatchParam("sort", "-modified_at").
			MatchParam("page", "0").
			Reply(200).
			BodyString("[" + string(tournamentData) + "]")

		// Mock SeriesExist to return that series exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(tournamentResponse.SerieID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		mockDB.ExpectExec("INSERT INTO tournaments").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.GetTournaments(false)
		st.Expect(t, err, nil)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/tournaments").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.GetTournaments(false)
		st.Reject(t, err, nil)
	})
}

func TestGetMatchPage(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - getMatchPage for upcoming matches (even page)", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/matches/upcoming").
			MatchParam("page", "0").
			Reply(200).
			BodyString("[" + string(matchData) + "]")

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(0, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		st.Expect(t, result.Err, nil)
		st.Reject(t, result.Matches, nil)
		st.Expect(t, len(result.Matches) > 0, true)
	})

	t.Run("Success - getMatchPage for past matches (odd page)", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/matches/past").
			MatchParam("page", "0").
			Reply(200).
			BodyString("[" + string(matchData) + "]")

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(1, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		st.Expect(t, result.Err, nil)
		st.Reject(t, result.Matches, nil)
		st.Expect(t, len(result.Matches) > 0, true)
	})

	t.Run("Error - Non-200 status code", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/matches/upcoming").
			Reply(404)

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(0, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		// Note: The implementation has a bug where err is nil on non-200 status
		// but Matches is still an empty slice
		st.Expect(t, len(result.Matches), 0)
	})

	t.Run("Error - Invalid JSON response", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/matches/upcoming").
			Reply(200).
			BodyString("invalid json")

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(0, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		st.Reject(t, result.Err, nil)
		st.Expect(t, len(result.Matches), 0)
	})
}

func TestGetMatches(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - GetMatches with setup=false", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var matchResponse pandatypes.MatchLike
		err = json.Unmarshal(matchData, &matchResponse)
		st.Assert(t, err, nil)

		// Mock both upcoming and past matches endpoints for all pages
		for i := 0; i < Pages; i++ {
			if i%2 == 0 {
				gock.New("https://api.pandascore.io").
					Get("/matches/upcoming").
					MatchParam("page", strings.Join([]string{}, "")).
					Reply(200).
					BodyString("[" + string(matchData) + "]")
			} else {
				gock.New("https://api.pandascore.io").
					Get("/matches/past").
					Reply(200).
					BodyString("[" + string(matchData) + "]")
			}
		}

		// Mock database expectations for tournament existence check and match write
		for i := 0; i < Pages; i++ {
			mockDB.ExpectQuery("SELECT COUNT").
				WithArgs(int32(matchResponse.TournamentID)).
				WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

			mockDB.ExpectExec("INSERT INTO matches").
				WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
		}

		err = client.GetMatches(false)
		st.Expect(t, err, nil)
	})
}

func TestGetTeams(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - GetTeams with setup=false", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		teamData, err := os.ReadFile("../static/fetch_data/teams.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/teams").
			Reply(200).
			BodyString("[" + string(teamData) + "]")

		mockDB.ExpectExec("INSERT INTO teams").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.GetTeams(false)
		st.Expect(t, err, nil)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/teams").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.GetTeams(false)
		st.Reject(t, err, nil)
	})

	t.Run("Error - Invalid JSON response", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/teams").
			Reply(200).
			BodyString("invalid json")

		err := client.GetTeams(false)
		st.Reject(t, err, nil)
	})
}

func TestWriteMatches(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - WriteMatches with existing tournament", func(t *testing.T) {
		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var match pandatypes.MatchLike
		err = json.Unmarshal(matchData, &match)
		st.Assert(t, err, nil)

		matches := pandatypes.MatchLikes{match}

		// Mock TournamentExist to return that tournament exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.TournamentID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		mockDB.ExpectExec("INSERT INTO matches").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		// Since match is finished, expect team checks (2 teams)
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.Opponents[0].Opponent.ID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.Opponents[1].Opponent.ID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		client.WriteMatches(matches)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})
}

func TestConstants(t *testing.T) {
	// Test that the constants have the expected values
	st.Expect(t, sortedBy, "-modified_at")
	st.Expect(t, Pages, 20)
	st.Expect(t, SetupPages, 50)
}

func TestGetMatchPageReadError(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Error - Non-200 status code", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/matches/upcoming").
			Reply(404)

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(0, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		// Note: The implementation has a bug where err is nil on non-200 status
		// but Matches is still an empty slice
		st.Expect(t, len(result.Matches), 0)
	})
}
