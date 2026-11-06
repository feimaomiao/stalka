package pandatypes

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"time"

	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/jackc/pgx/v5/pgtype"
)

// PandaScore match statuses that change how a row is written.
const (
	MatchStatusFinished = "finished"
	MatchStatusCanceled = "canceled"
)

const (
	twoTeams = 2
)

// StreamEntry mirrors a single entry of PandaScore's `streams_list` array.
type StreamEntry struct {
	Main     bool   `json:"main"`
	Language string `json:"language"`
	EmbedURL string `json:"embed_url"`
	Official bool   `json:"official"`
	RawURL   string `json:"raw_url"`
}

type PandaDataLike interface {
	ToRow() RowLike
}

type RowLike interface {
	WriteToDB(ctx context.Context, db *dbtypes.Queries) error
}

type GameLike struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	CurrentVersion any    `json:"current_version"`
	Slug           string `json:"slug"`
	Leagues        []struct {
		ID         int       `json:"id"`
		Name       string    `json:"name"`
		URL        any       `json:"url"`
		Slug       string    `json:"slug"`
		ModifiedAt time.Time `json:"modified_at"`
		Series     []struct {
			ID         int       `json:"id"`
			Name       string    `json:"name"`
			Year       int       `json:"year"`
			BeginAt    time.Time `json:"begin_at"`
			EndAt      time.Time `json:"end_at"`
			WinnerID   int       `json:"winner_id"`
			Slug       string    `json:"slug"`
			WinnerType string    `json:"winner_type"`
			ModifiedAt time.Time `json:"modified_at"`
			LeagueID   int       `json:"league_id"`
			Season     any       `json:"season"`
			FullName   string    `json:"full_name"`
		} `json:"series"`
		ImageURL string `json:"image_url"`
	} `json:"leagues"`
}

type LeagueLike struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	URL        any       `json:"url"`
	Slug       string    `json:"slug"`
	ModifiedAt time.Time `json:"modified_at"`
	Videogame  struct {
		ID             int    `json:"id"`
		Name           string `json:"name"`
		CurrentVersion string `json:"current_version"`
		Slug           string `json:"slug"`
	} `json:"videogame"`
	Series []struct {
		ID         int       `json:"id"`
		Name       string    `json:"name"`
		Year       int       `json:"year"`
		BeginAt    time.Time `json:"begin_at"`
		EndAt      time.Time `json:"end_at"`
		WinnerID   any       `json:"winner_id"`
		WinnerType string    `json:"winner_type"`
		Slug       string    `json:"slug"`
		ModifiedAt time.Time `json:"modified_at"`
		LeagueID   int       `json:"league_id"`
		Season     any       `json:"season"`
		FullName   string    `json:"full_name"`
	} `json:"series"`
	ImageURL string `json:"image_url"`
}

type SeriesLike struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Year       int       `json:"year"`
	Slug       string    `json:"slug"`
	BeginAt    time.Time `json:"begin_at"`
	EndAt      time.Time `json:"end_at"`
	WinnerID   any       `json:"winner_id"`
	WinnerType string    `json:"winner_type"`
	Videogame  struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"videogame"`
	ModifiedAt time.Time `json:"modified_at"`
	LeagueID   int       `json:"league_id"`
	League     struct {
		ID         int       `json:"id"`
		Name       string    `json:"name"`
		URL        string    `json:"url"`
		Slug       string    `json:"slug"`
		ModifiedAt time.Time `json:"modified_at"`
		ImageURL   string    `json:"image_url"`
	} `json:"league"`
	Tournaments []struct {
		ID            int       `json:"id"`
		Name          string    `json:"name"`
		Type          string    `json:"type"`
		Country       string    `json:"country"`
		Slug          string    `json:"slug"`
		BeginAt       time.Time `json:"begin_at"`
		DetailedStats bool      `json:"detailed_stats"`
		EndAt         time.Time `json:"end_at"`
		WinnerID      any       `json:"winner_id"`
		WinnerType    string    `json:"winner_type"`
		SerieID       int       `json:"serie_id"`
		ModifiedAt    time.Time `json:"modified_at"`
		LeagueID      int       `json:"league_id"`
		Prizepool     any       `json:"prizepool"`
		Tier          string    `json:"tier"`
		HasBracket    bool      `json:"has_bracket"`
		Region        string    `json:"region"`
		LiveSupported bool      `json:"live_supported"`
	} `json:"tournaments"`
	Season         string `json:"season"`
	VideogameTitle any    `json:"videogame_title"`
	FullName       string `json:"full_name"`
}

type TournamentLike struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Matches []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
		Live   struct {
			Supported bool `json:"supported"`
			URL       any  `json:"url"`
			OpensAt   any  `json:"opens_at"`
		} `json:"live"`
		BeginAt             time.Time `json:"begin_at"`
		DetailedStats       bool      `json:"detailed_stats"`
		EndAt               any       `json:"end_at"`
		Forfeit             bool      `json:"forfeit"`
		WinnerID            any       `json:"winner_id"`
		WinnerType          string    `json:"winner_type"`
		Draw                bool      `json:"draw"`
		Slug                string    `json:"slug"`
		ModifiedAt          time.Time `json:"modified_at"`
		TournamentID        int       `json:"tournament_id"`
		MatchType           string    `json:"match_type"`
		NumberOfGames       int       `json:"number_of_games"`
		ScheduledAt         time.Time `json:"scheduled_at"`
		OriginalScheduledAt time.Time `json:"original_scheduled_at"`
		GameAdvantage       any       `json:"game_advantage"`
		StreamsList         []struct {
			Main     bool   `json:"main"`
			Language string `json:"language"`
			EmbedURL any    `json:"embed_url"`
			Official bool   `json:"official"`
			RawURL   string `json:"raw_url"`
		} `json:"streams_list"`
		Rescheduled bool `json:"rescheduled"`
	} `json:"matches"`
	Country       string    `json:"country"`
	BeginAt       time.Time `json:"begin_at"`
	DetailedStats bool      `json:"detailed_stats"`
	EndAt         time.Time `json:"end_at"`
	WinnerID      any       `json:"winner_id"`
	WinnerType    string    `json:"winner_type"`
	Teams         []any     `json:"teams"`
	Slug          string    `json:"slug"`
	ModifiedAt    time.Time `json:"modified_at"`
	Videogame     struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"videogame"`
	SerieID int `json:"serie_id"`
	Serie   struct {
		ID         int       `json:"id"`
		Name       string    `json:"name"`
		Year       int       `json:"year"`
		BeginAt    time.Time `json:"begin_at"`
		EndAt      time.Time `json:"end_at"`
		WinnerID   any       `json:"winner_id"`
		WinnerType string    `json:"winner_type"`
		Slug       string    `json:"slug"`
		ModifiedAt time.Time `json:"modified_at"`
		LeagueID   int       `json:"league_id"`
		Season     string    `json:"season"`
		FullName   string    `json:"full_name"`
	} `json:"serie"`
	LeagueID int `json:"league_id"`
	League   struct {
		ID         int       `json:"id"`
		Name       string    `json:"name"`
		URL        string    `json:"url"`
		Slug       string    `json:"slug"`
		ModifiedAt time.Time `json:"modified_at"`
		ImageURL   string    `json:"image_url"`
	} `json:"league"`
	Prizepool      any    `json:"prizepool"`
	Tier           string `json:"tier"`
	VideogameTitle any    `json:"videogame_title"`
	HasBracket     bool   `json:"has_bracket"`
	Region         string `json:"region"`
	LiveSupported  bool   `json:"live_supported"`
	ExpectedRoster []any  `json:"expected_roster"`
}

type MatchLike struct {
	Results []struct {
		TeamID int `json:"team_id"`
		Score  int `json:"score"`
	} `json:"results"`
	Tournament struct {
		ID            int       `json:"id"`
		Name          string    `json:"name"`
		Type          string    `json:"type"`
		Country       string    `json:"country"`
		BeginAt       time.Time `json:"begin_at"`
		DetailedStats bool      `json:"detailed_stats"`
		EndAt         time.Time `json:"end_at"`
		WinnerID      any       `json:"winner_id"`
		WinnerType    string    `json:"winner_type"`
		Slug          string    `json:"slug"`
		ModifiedAt    time.Time `json:"modified_at"`
		SerieID       int       `json:"serie_id"`
		LeagueID      int       `json:"league_id"`
		Prizepool     any       `json:"prizepool"`
		Tier          string    `json:"tier"`
		HasBracket    bool      `json:"has_bracket"`
		Region        string    `json:"region"`
		LiveSupported bool      `json:"live_supported"`
	} `json:"tournament"`
	Live struct {
		Supported bool      `json:"supported"`
		URL       string    `json:"url"`
		OpensAt   time.Time `json:"opens_at"`
	} `json:"live"`
	WinnerType   string `json:"winner_type"`
	TournamentID int    `json:"tournament_id"`
	League       struct {
		ID         int       `json:"id"`
		Name       string    `json:"name"`
		URL        any       `json:"url"`
		Slug       string    `json:"slug"`
		ModifiedAt time.Time `json:"modified_at"`
		ImageURL   string    `json:"image_url"`
	} `json:"league"`
	GameAdvantage    any           `json:"game_advantage"`
	StreamsList      []StreamEntry `json:"streams_list"`
	EndAt            time.Time     `json:"end_at"`
	VideogameTitle   any           `json:"videogame_title"`
	Slug             string        `json:"slug"`
	VideogameVersion struct {
		Name    string `json:"name"`
		Current bool   `json:"current"`
	} `json:"videogame_version"`
	ScheduledAt time.Time `json:"scheduled_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	Rescheduled bool      `json:"rescheduled"`
	MatchType   string    `json:"match_type"`
	Forfeit     bool      `json:"forfeit"`
	Opponents   []struct {
		Type     string `json:"type"`
		Opponent struct {
			ID         int       `json:"id"`
			Name       string    `json:"name"`
			Location   string    `json:"location"`
			Slug       string    `json:"slug"`
			ModifiedAt time.Time `json:"modified_at"`
			Acronym    string    `json:"acronym"`
			ImageURL   string    `json:"image_url"`
		} `json:"opponent"`
	} `json:"opponents"`
	Status        string    `json:"status"`
	BeginAt       time.Time `json:"begin_at"`
	DetailedStats bool      `json:"detailed_stats"`
	Draw          bool      `json:"draw"`
	Serie         struct {
		ID         int       `json:"id"`
		Name       string    `json:"name"`
		Year       int       `json:"year"`
		BeginAt    time.Time `json:"begin_at"`
		EndAt      time.Time `json:"end_at"`
		WinnerID   any       `json:"winner_id"`
		WinnerType string    `json:"winner_type"`
		Slug       string    `json:"slug"`
		ModifiedAt time.Time `json:"modified_at"`
		LeagueID   int       `json:"league_id"`
		Season     string    `json:"season"`
		FullName   string    `json:"full_name"`
	} `json:"serie"`
	LeagueID            int       `json:"league_id"`
	WinnerID            int       `json:"winner_id"`
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	OriginalScheduledAt time.Time `json:"original_scheduled_at"`
	Videogame           struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"videogame"`
	Games []struct {
		Complete      bool      `json:"complete"`
		ID            int       `json:"id"`
		Position      int       `json:"position"`
		Status        string    `json:"status"`
		Length        int       `json:"length"`
		Finished      bool      `json:"finished"`
		MatchID       int       `json:"match_id"`
		BeginAt       time.Time `json:"begin_at"`
		DetailedStats bool      `json:"detailed_stats"`
		EndAt         time.Time `json:"end_at"`
		Forfeit       bool      `json:"forfeit"`
		WinnerType    string    `json:"winner_type"`
		Winner        struct {
			ID   int    `json:"id"`
			Type string `json:"type"`
		} `json:"winner"`
	} `json:"games"`
	NumberOfGames int `json:"number_of_games"`
	SerieID       int `json:"serie_id"`
	Winner        struct {
		ID         int       `json:"id"`
		Name       string    `json:"name"`
		Location   string    `json:"location"`
		Slug       string    `json:"slug"`
		ModifiedAt time.Time `json:"modified_at"`
		Acronym    string    `json:"acronym"`
		ImageURL   string    `json:"image_url"`
	} `json:"winner"`
}

type TeamLike struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Location         any       `json:"location"`
	Players          []any     `json:"players"`
	Slug             string    `json:"slug"`
	ModifiedAt       time.Time `json:"modified_at"`
	Acronym          string    `json:"acronym"`
	ImageURL         string    `json:"image_url"`
	CurrentVideogame struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"current_videogame"`
}

type GameLikes []GameLike

type LeagueLikes []LeagueLike

type SeriesLikes []SeriesLike

type TournamentLikes []TournamentLike

type MatchLikes []MatchLike

type TeamLikes []TeamLike
type ResultMatchLikes struct {
	Matches MatchLikes
	Err     error
}

type GameRow struct {
	ID   int
	Name string
	Slug string
}

// SafeIntToInt32 converts an int to int32 safely, returning an error if the value overflows.
// @param value - the int value to convert.
// @returns the int32 value and an error if the conversion fails.
func SafeIntToInt32(value int) (int32, error) {
	if value > math.MaxInt32 || value < math.MinInt32 {
		err := fmt.Errorf("value %d overflows int32 range [%d, %d]", value, math.MinInt32, math.MaxInt32)
		return 0, err
	}
	return int32(value), nil
}

func (game GameLike) ToRow() RowLike {
	return GameRow{
		ID:   game.ID,
		Name: game.Name,
		Slug: game.Slug,
	}
}

func (row GameRow) WriteToDB(ctx context.Context, db *dbtypes.Queries) error {
	id, err := SafeIntToInt32(row.ID)
	if err != nil {
		return err
	}
	err = db.InsertToGames(ctx, dbtypes.InsertToGamesParams{
		ID:   id,
		Name: row.Name,
		Slug: pgtype.Text{String: row.Slug, Valid: row.Slug != ""},
	})
	return err
}

type LeagueRow struct {
	ID        int
	Name      string
	Slug      string
	GameID    int
	ImageLink string
}

func (league LeagueLike) ToRow() RowLike {
	return LeagueRow{
		ID:        league.ID,
		Name:      league.Name,
		Slug:      league.Slug,
		GameID:    league.Videogame.ID,
		ImageLink: league.ImageURL,
	}
}

func (row LeagueRow) WriteToDB(ctx context.Context, db *dbtypes.Queries) error {
	id, err := SafeIntToInt32(row.ID)
	if err != nil {
		return err
	}
	gameID, err := SafeIntToInt32(row.GameID)
	if err != nil {
		return err
	}
	err = db.InsertToLeagues(ctx, dbtypes.InsertToLeaguesParams{
		ID:        id,
		Slug:      pgtype.Text{String: row.Slug, Valid: row.Slug != ""},
		Name:      row.Name,
		GameID:    gameID,
		ImageLink: pgtype.Text{String: row.ImageLink, Valid: row.ImageLink != ""},
	})
	return err
}

type SeriesRow struct {
	ID       int
	Name     string
	Slug     string
	GameID   int
	LeagueID int
}

func (series SeriesLike) ToRow() RowLike {
	return SeriesRow{
		ID:       series.ID,
		Name:     series.Name,
		Slug:     series.Slug,
		GameID:   series.Videogame.ID,
		LeagueID: series.League.ID,
	}
}

func (row SeriesRow) WriteToDB(ctx context.Context, db *dbtypes.Queries) error {
	id, err := SafeIntToInt32(row.ID)
	if err != nil {
		return err
	}
	gameID, err := SafeIntToInt32(row.GameID)
	if err != nil {
		return err
	}
	leagueID, err := SafeIntToInt32(row.LeagueID)
	if err != nil {
		return err
	}
	err = db.InsertToSeries(ctx, dbtypes.InsertToSeriesParams{
		ID:       id,
		Name:     row.Name,
		Slug:     pgtype.Text{String: row.Slug, Valid: row.Slug != ""},
		GameID:   gameID,
		LeagueID: leagueID,
	})
	return err
}

type TournamentRow struct {
	ID       int
	Name     string
	Slug     string
	Tier     int
	GameID   int
	LeagueID int
	SerieID  int
}

func (tournament TournamentLike) ToRow() RowLike {
	var tier int
	switch tournament.Tier {
	case "S", "s":
		tier = 1
	case "A", "a":
		tier = 2
	case "B", "b":
		tier = 3
	case "C", "c":
		tier = 4
	case "D", "d":
		tier = 5
	default:
		tier = 6
	}
	return TournamentRow{
		ID:       tournament.ID,
		Name:     tournament.Name,
		Slug:     tournament.Slug,
		Tier:     tier,
		GameID:   tournament.Videogame.ID,
		SerieID:  tournament.Serie.ID,
		LeagueID: tournament.League.ID,
	}
}

func (row TournamentRow) WriteToDB(ctx context.Context, db *dbtypes.Queries) error {
	id, err := SafeIntToInt32(row.ID)
	if err != nil {
		return err
	}
	gameID, err := SafeIntToInt32(row.GameID)
	if err != nil {
		return err
	}
	leagueID, err := SafeIntToInt32(row.LeagueID)
	if err != nil {
		return err
	}
	serieID, err := SafeIntToInt32(row.SerieID)
	if err != nil {
		return err
	}
	//nolint:gosec // tier is fixed by the switch statement previously
	tier := int32(row.Tier)
	err = db.InsertToTournaments(ctx, dbtypes.InsertToTournamentsParams{
		ID:       id,
		Name:     row.Name,
		Slug:     pgtype.Text{String: row.Slug, Valid: row.Slug != ""},
		Tier:     pgtype.Int4{Int32: tier, Valid: row.Tier != 0},
		GameID:   gameID,
		SerieID:  serieID,
		LeagueID: leagueID,
	})
	return err
}

type MatchRow struct {
	ID                int
	Name              string
	Slug              string
	Finished          bool
	ExpectedStartTime time.Time
	Team1ID           int
	Team1Score        int
	Team2ID           int
	Team2Score        int
	AmountOfGames     int
	ActualGameTime    float64
	GameID            int
	LeagueID          int
	SerieID           int
	TournamentID      int
	StreamURL         string
	Status            string
	ModifiedAt        time.Time
	Forfeit           bool
	WinnerID          int
}

// IsFinished reports whether the match was actually played to a result. A
// canceled match can never count, whatever its timestamps say, because a row
// marked finished is shown with its 0-0 scores as a final result.
func (match MatchLike) IsFinished() bool {
	if match.Status == MatchStatusCanceled {
		return false
	}
	return match.Status == MatchStatusFinished || !match.EndAt.IsZero()
}

// opponentSlot returns the id and score for one side of the match. Each side
// fills in on its own: a bracket match reads "X vs TBD" until the other team is
// known, and X must still be findable by team filters in the meantime.
func (match MatchLike) opponentSlot(index int) (int, int) {
	if index >= len(match.Opponents) || len(match.Opponents) > twoTeams {
		return 0, 0
	}
	id := match.Opponents[index].Opponent.ID
	for _, result := range match.Results {
		if result.TeamID == id {
			return id, result.Score
		}
	}
	// Player matches key their results by player_id, so fall back to position.
	if index < len(match.Results) {
		return id, match.Results[index].Score
	}
	return id, 0
}

func (match MatchLike) ToRow() RowLike {
	actualGT := 0.0
	// Canceled and malformed matches can lack either, and a float column would
	// happily store the infinity that dividing by zero games produces.
	if !match.EndAt.IsZero() && !match.BeginAt.IsZero() && match.NumberOfGames > 0 {
		actualGT = match.EndAt.Sub(match.BeginAt).Seconds() / float64(match.NumberOfGames)
	}
	t1ID, t1Score := match.opponentSlot(0)
	t2ID, t2Score := match.opponentSlot(1)
	streamURL := ExtractPrimaryStreamURL(match.StreamsList)
	return MatchRow{
		ID:                match.ID,
		Slug:              match.Slug,
		Finished:          match.IsFinished(),
		Status:            match.Status,
		ModifiedAt:        match.ModifiedAt,
		Forfeit:           match.Forfeit,
		WinnerID:          match.WinnerID,
		GameID:            match.Videogame.ID,
		LeagueID:          match.League.ID,
		SerieID:           match.Serie.ID,
		TournamentID:      match.Tournament.ID,
		Team1ID:           t1ID,
		Team1Score:        t1Score,
		Team2ID:           t2ID,
		Team2Score:        t2Score,
		Name:              match.Name,
		ExpectedStartTime: match.ScheduledAt,
		AmountOfGames:     match.NumberOfGames,
		ActualGameTime:    actualGT,
		StreamURL:         streamURL,
	}
}

func (row MatchRow) WriteToDB(ctx context.Context, db *dbtypes.Queries) error {
	id, err := SafeIntToInt32(row.ID)
	if err != nil {
		return err
	}
	team1Id, err := SafeIntToInt32(row.Team1ID)
	if err != nil {
		return err
	}
	team2Id, err := SafeIntToInt32(row.Team2ID)
	if err != nil {
		return err
	}
	team1Score, err := SafeIntToInt32(row.Team1Score)
	if err != nil {
		return err
	}
	team2Score, err := SafeIntToInt32(row.Team2Score)
	if err != nil {
		return err
	}
	gameAmount, err := SafeIntToInt32(row.AmountOfGames)
	if err != nil {
		return err
	}
	gameID, err := SafeIntToInt32(row.GameID)
	if err != nil {
		return err
	}
	leagueID, err := SafeIntToInt32(row.LeagueID)
	if err != nil {
		return err
	}
	serieID, err := SafeIntToInt32(row.SerieID)
	if err != nil {
		return err
	}
	tournamentID, err := SafeIntToInt32(row.TournamentID)
	if err != nil {
		return err
	}
	winnerID, err := SafeIntToInt32(row.WinnerID)
	if err != nil {
		return err
	}
	err = db.InsertToMatches(ctx, dbtypes.InsertToMatchesParams{
		ID:       id,
		Name:     row.Name,
		Slug:     pgtype.Text{String: row.Slug, Valid: row.Slug != ""},
		Finished: row.Finished,
		ExpectedStartTime: pgtype.Timestamp{
			Time:             row.ExpectedStartTime,
			Valid:            !row.ExpectedStartTime.IsZero(),
			InfinityModifier: 0,
		},
		ActualGameTime: row.ActualGameTime,
		Team1ID:        team1Id,
		Team1Score:     team1Score,
		Team2ID:        team2Id,
		Team2Score:     team2Score,
		AmountOfGames:  gameAmount,
		GameID:         gameID,
		LeagueID:       leagueID,
		SeriesID:       serieID,
		TournamentID:   tournamentID,
		StreamURL:      pgtype.Text{String: row.StreamURL, Valid: row.StreamURL != ""},
		Status:         pgtype.Text{String: row.Status, Valid: row.Status != ""},
		ModifiedAt: pgtype.Timestamp{
			Time:             row.ModifiedAt,
			Valid:            !row.ModifiedAt.IsZero(),
			InfinityModifier: 0,
		},
		Forfeit:  row.Forfeit,
		WinnerID: winnerID,
	})
	return err
}

type TeamRow struct {
	ID        int
	GameID    int
	Name      string
	Acronym   string
	Slug      string
	ImageLink string
}

func (team TeamLike) ToRow() RowLike {
	return TeamRow{
		ID:        team.ID,
		GameID:    team.CurrentVideogame.ID,
		Name:      team.Name,
		Acronym:   team.Acronym,
		Slug:      team.Slug,
		ImageLink: team.ImageURL,
	}
}

func (row TeamRow) WriteToDB(ctx context.Context, db *dbtypes.Queries) error {
	id, err := SafeIntToInt32(row.ID)
	if err != nil {
		return err
	}
	gameID, err := SafeIntToInt32(row.GameID)
	if err != nil {
		return err
	}
	err = db.InsertToTeams(ctx, dbtypes.InsertToTeamsParams{
		ID:        id,
		Name:      row.Name,
		Slug:      pgtype.Text{String: row.Slug, Valid: row.Slug != ""},
		Acronym:   pgtype.Text{String: row.Acronym, Valid: row.Acronym != ""},
		ImageLink: pgtype.Text{String: row.ImageLink, Valid: row.ImageLink != ""},
		GameID:    gameID,
	})
	return err
}

// ExtractPrimaryStreamURL returns the primary stream URL from a PandaScore match response.
// Priority: main=true stream, then English language, then first entry.
// Returns empty string if no valid https:// URL is found.
func ExtractPrimaryStreamURL(streams []StreamEntry) string {
	for _, s := range streams {
		if s.Main && isValidHTTPSURL(s.RawURL) {
			return s.RawURL
		}
	}
	for _, s := range streams {
		if s.Language == "en" && isValidHTTPSURL(s.RawURL) {
			return s.RawURL
		}
	}
	if len(streams) > 0 && isValidHTTPSURL(streams[0].RawURL) {
		return streams[0].RawURL
	}
	return ""
}

// isValidHTTPSURL parses the input and verifies it is an absolute https:// URL with a host.
func isValidHTTPSURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "https" && u.Host != ""
}
