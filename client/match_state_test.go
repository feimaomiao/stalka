package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/nbio/st"
	"github.com/pashagolub/pgxmock/v4"
)

const (
	reconcileSelect = "SELECT id FROM MATCHES"
	removedUpdate   = "UPDATE MATCHES SET status = 'removed'"
	liveUpdate      = "UPDATE MATCHES SET is_live"
	matchBody       = `{"id": 5, "name": "M", "status": "finished", "tournament_id": 1000,
		"number_of_games": 3, "scheduled_at": "2026-05-01T14:00:00Z", "end_at": "2026-05-01T16:00:00Z"}`
)

func expectMatchWrite(mockDB pgxmock.PgxPoolIface) {
	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(int32(1000)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))
	args := make([]interface{}, 20)
	for i := range args {
		args[i] = pgxmock.AnyArg()
	}
	mockDB.ExpectExec("INSERT INTO matches").WithArgs(args...).WillReturnResult(pgxmock.NewResult("INSERT", 1))
}

func expectReconcileIDs(mockDB pgxmock.PgxPoolIface, ids ...int32) {
	rows := pgxmock.NewRows([]string{"id"})
	for _, id := range ids {
		rows.AddRow(id)
	}
	mockDB.ExpectQuery(reconcileSelect).WithArgs(int32(ReconcileBatch)).WillReturnRows(rows)
}

// Each list fetcher used to return a nil error on a non-200, so a failed fetch
// read as a page with nothing on it.
func TestListFetchersReportNon200(t *testing.T) {
	cases := []struct {
		name string
		path string
		run  func(*PandaClient) error
	}{
		{"games", "/videogames/", (*PandaClient).UpdateGames},
		{"leagues", "/leagues/", func(c *PandaClient) error { return c.GetLeagues(true) }},
		{"series", "/series/", func(c *PandaClient) error { return c.GetSeries(true) }},
		{"tournaments", "/tournaments/", func(c *PandaClient) error { return c.GetTournaments(true) }},
		{"teams", "/teams/", func(c *PandaClient) error { return c.GetTeams(true) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := newLivesClient(t)
			gock.New(livesTestBaseURL).Get(tc.path).Reply(http.StatusForbidden)

			var statusErr *StatusError
			st.Expect(t, errors.As(tc.run(client), &statusErr), true)
			st.Expect(t, statusErr.Code, http.StatusForbidden)
		})
	}
}

func TestGetMatchesErrors(t *testing.T) {
	t.Run("every page failing is an error", func(t *testing.T) {
		client, _ := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/upcoming/").Persist().Reply(http.StatusForbidden)
		gock.New(livesTestBaseURL).Get("/matches/past/").Persist().Reply(http.StatusForbidden)

		st.Reject(t, client.GetMatches(false), nil)
	})

	t.Run("some pages failing is not", func(t *testing.T) {
		client, _ := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/upcoming/").Persist().Reply(http.StatusOK).BodyString("[]")
		gock.New(livesTestBaseURL).Get("/matches/past/").Persist().Reply(http.StatusForbidden)

		st.Expect(t, client.GetMatches(false), nil)
	})
}

// A team with no game can't be stored, and one bad row mustn't lose the page.
func TestGetTeamsSkipsTeamWithoutGame(t *testing.T) {
	client, mockDB := newLivesClient(t)
	gock.New(livesTestBaseURL).Get("/teams/").Reply(http.StatusOK).BodyString(`[
		{"id": 1, "name": "No game", "current_videogame": null},
		{"id": 2, "name": "Has game", "current_videogame": {"id": 1}}
	]`)
	mockDB.ExpectExec("INSERT INTO teams").
		WithArgs(int32(2), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), int32(1)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	st.Expect(t, client.GetTeams(false), nil)
	st.Expect(t, mockDB.ExpectationsWereMet(), nil)
}

func TestListWriteFailureKeepsGoing(t *testing.T) {
	client, mockDB := newLivesClient(t)
	gock.New(livesTestBaseURL).Get("/teams/").Reply(http.StatusOK).BodyString(`[
		{"id": 1, "name": "Bad", "current_videogame": {"id": 1}},
		{"id": 2, "name": "Good", "current_videogame": {"id": 1}}
	]`)
	anyTeam := []interface{}{
		pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
	}
	mockDB.ExpectExec("INSERT INTO teams").WithArgs(anyTeam...).WillReturnError(io.ErrUnexpectedEOF)
	mockDB.ExpectExec("INSERT INTO teams").WithArgs(anyTeam...).WillReturnResult(pgxmock.NewResult("INSERT", 1))

	st.Expect(t, client.GetTeams(false), nil)
	st.Expect(t, mockDB.ExpectationsWereMet(), nil)
}

// A failed step reports through the returned error but can't stop the rest.
func TestStartupRunsEveryStep(t *testing.T) {
	client, _ := newLivesClient(t)
	gock.New(livesTestBaseURL).Persist().Reply(http.StatusForbidden)

	st.Reject(t, client.Startup(), nil)
	// One request each for games, leagues, series, tournaments, teams and lives,
	// plus the concurrent match pages.
	st.Expect(t, client.Run.Load(), int64(6+SetupPages))
}

func TestGetLivesRefreshesEndedMatch(t *testing.T) {
	flipped := func() *pgxmock.Rows {
		return pgxmock.NewRows([]string{"id", "is_live"}).AddRow(int32(5), false)
	}

	t.Run("match that left the feed is fetched by id", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/running/").Reply(http.StatusOK).BodyString("[]")
		gock.New(livesTestBaseURL).Get("/matches/5/").Reply(http.StatusOK).BodyString(matchBody)
		mockDB.ExpectQuery(liveUpdate).WithArgs([]int32{}).WillReturnRows(flipped())
		expectMatchWrite(mockDB)

		st.Expect(t, client.GetLives(), nil)
		st.Expect(t, gock.IsDone(), true)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	t.Run("match that just went live is not", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/running/").Reply(http.StatusOK).BodyString("[]")
		mockDB.ExpectQuery(liveUpdate).WithArgs([]int32{}).
			WillReturnRows(pgxmock.NewRows([]string{"id", "is_live"}).AddRow(int32(5), true))

		st.Expect(t, client.GetLives(), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	t.Run("failed refresh doesn't fail the poll", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/running/").Reply(http.StatusOK).BodyString("[]")
		gock.New(livesTestBaseURL).Get("/matches/5/").Reply(http.StatusForbidden)
		mockDB.ExpectQuery(liveUpdate).WithArgs([]int32{}).WillReturnRows(flipped())

		st.Expect(t, client.GetLives(), nil)
	})
}

func TestGetLivesPagination(t *testing.T) {
	fullPage := func() string {
		rows := make([]string, perPage)
		for i := range rows {
			rows[i] = fmt.Sprintf(`{"id": %d, "tournament_id": 1000, "number_of_games": 1}`, i+1)
		}
		return "[" + strings.Join(rows, ",") + "]"
	}

	t.Run("a full page reads the next one", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/running/").MatchParam("page", "1").
			Reply(http.StatusOK).BodyString(fullPage())
		gock.New(livesTestBaseURL).Get("/matches/running/").MatchParam("page", "2").
			Reply(http.StatusOK).BodyString(`[{"id": 101, "tournament_id": 1000, "number_of_games": 1}]`)
		for range perPage + 1 {
			expectMatchWrite(mockDB)
		}
		want := make([]int32, perPage+1)
		for i := range want {
			want[i] = int32(i + 1)
		}
		mockDB.ExpectQuery(liveUpdate).WithArgs(want).WillReturnRows(pgxmock.NewRows([]string{"id", "is_live"}))

		st.Expect(t, client.GetLives(), nil)
		st.Expect(t, gock.IsDone(), true)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	// A partial list would clear is_live on every match on the missing page.
	t.Run("a failed page leaves is_live alone", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/running/").MatchParam("page", "1").
			Reply(http.StatusOK).BodyString(fullPage())
		gock.New(livesTestBaseURL).Get("/matches/running/").MatchParam("page", "2").
			Reply(http.StatusForbidden)

		st.Reject(t, client.GetLives(), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})
}

// A finished match still on the running feed must not be flagged live.
func TestGetLivesSkipsFinishedMatch(t *testing.T) {
	client, mockDB := newLivesClient(t)
	gock.New(livesTestBaseURL).Get("/matches/running/").Reply(http.StatusOK).BodyString("[" + matchBody + "]")
	expectMatchWrite(mockDB)
	mockDB.ExpectQuery(liveUpdate).WithArgs([]int32{}).WillReturnRows(pgxmock.NewRows([]string{"id", "is_live"}))

	st.Expect(t, client.GetLives(), nil)
	st.Expect(t, mockDB.ExpectationsWereMet(), nil)
}

func TestRefreshMatch(t *testing.T) {
	t.Run("404 on the match marks it removed", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/5/").Reply(http.StatusNotFound)
		mockDB.ExpectExec(removedUpdate).WithArgs(int32(5)).WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		st.Expect(t, client.refreshMatch(5), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	// The match exists, so calling it removed here would hide a real fixture.
	t.Run("404 on a parent does not", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/5/").Reply(http.StatusOK).BodyString(matchBody)
		gock.New(livesTestBaseURL).Get("/tournaments/1000/").Reply(http.StatusNotFound)
		mockDB.ExpectQuery("SELECT COUNT").WithArgs(int32(1000)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

		st.Reject(t, client.refreshMatch(5), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	t.Run("other failures are returned", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/5/").Reply(http.StatusForbidden)

		st.Reject(t, client.refreshMatch(5), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})
}

func TestReconcileMatches(t *testing.T) {
	t.Run("refreshes the batch", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		expectReconcileIDs(mockDB, 5)
		gock.New(livesTestBaseURL).Get("/matches/5/").Reply(http.StatusOK).BodyString(matchBody)
		expectMatchWrite(mockDB)

		st.Expect(t, client.ReconcileMatches(), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	t.Run("one failure doesn't stop it", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		expectReconcileIDs(mockDB, 4, 5)
		gock.New(livesTestBaseURL).Get("/matches/4/").Reply(http.StatusForbidden)
		gock.New(livesTestBaseURL).Get("/matches/5/").Reply(http.StatusOK).BodyString(matchBody)
		expectMatchWrite(mockDB)

		st.Expect(t, client.ReconcileMatches(), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	t.Run("Pandascore being down stops it", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		expectReconcileIDs(mockDB, 1, 2, 3, 4)
		gock.New(livesTestBaseURL).Get("/matches/1/").ReplyError(io.ErrUnexpectedEOF)
		gock.New(livesTestBaseURL).Get("/matches/2/").ReplyError(io.ErrUnexpectedEOF)
		gock.New(livesTestBaseURL).Get("/matches/3/").ReplyError(io.ErrUnexpectedEOF)
		gock.New(livesTestBaseURL).Get("/matches/4/").Reply(http.StatusOK).BodyString(matchBody)

		st.Reject(t, client.ReconcileMatches(), nil)
		st.Expect(t, gock.IsPending(), true) // match 4 was never requested
	})

	// Broken rows sort first every run, so stopping for them would block the batch.
	t.Run("failures on single matches don't stop it", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		expectReconcileIDs(mockDB, 1, 2, 3, 5)
		gock.New(livesTestBaseURL).Get("/matches/1/").Reply(http.StatusForbidden)
		gock.New(livesTestBaseURL).Get("/matches/2/").Reply(http.StatusForbidden)
		gock.New(livesTestBaseURL).Get("/matches/3/").Reply(http.StatusOK).BodyString(`not json`)
		gock.New(livesTestBaseURL).Get("/matches/5/").Reply(http.StatusOK).BodyString(matchBody)
		expectMatchWrite(mockDB)

		st.Expect(t, client.ReconcileMatches(), nil)
		st.Expect(t, gock.IsDone(), true)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})
}

func TestMakeRequestRetries(t *testing.T) {
	t.Run("5xx is retried", func(t *testing.T) {
		client, _ := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/videogames/").Reply(http.StatusServiceUnavailable)
		gock.New(livesTestBaseURL).Get("/videogames/").Reply(http.StatusOK).BodyString("[]")

		st.Expect(t, client.UpdateGames(), nil)
		st.Expect(t, client.Run.Load(), int64(2))
	})

	t.Run("4xx is not", func(t *testing.T) {
		client, _ := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/videogames/").Reply(http.StatusForbidden)

		st.Reject(t, client.UpdateGames(), nil)
		st.Expect(t, client.Run.Load(), int64(1))
	})
}

func TestRetryAfter(t *testing.T) {
	cases := []struct {
		name   string
		header string
		base   time.Duration
	}{
		{"missing", "", defaultRetryAfter},
		{"garbage", "soon", defaultRetryAfter},
		{"seconds", "5", 5 * time.Second},
		{"capped", "86400", maxRetryAfter},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wait := retryAfter(tc.header)
			if wait < tc.base || wait >= 2*tc.base {
				t.Fatalf("retryAfter(%q) = %s, want in [%s, %s)", tc.header, wait, tc.base, 2*tc.base)
			}
		})
	}
}
