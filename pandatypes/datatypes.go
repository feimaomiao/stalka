package pandatypes

import (
	"database/sql"
	"time"
)

const (
	twoTeams = 2
)

type PandaDataLike interface {
	ToRow() RowLike
}

type RowLike interface {
	WriteToDB(db *sql.DB) error
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
	GameAdvantage any `json:"game_advantage"`
	StreamsList   []struct {
		Main     bool   `json:"main"`
		Language string `json:"language"`
		EmbedURL string `json:"embed_url"`
		Official bool   `json:"official"`
		RawURL   string `json:"raw_url"`
	} `json:"streams_list"`
	EndAt            time.Time `json:"end_at"`
	VideogameTitle   any       `json:"videogame_title"`
	Slug             string    `json:"slug"`
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

type GameLikes []GameLike

type LeagueLikes []LeagueLike

type SeriesLikes []SeriesLike

type TournamentLikes []TournamentLike

type MatchLikes []MatchLike
type ResultMatchLikes struct {
	Matches MatchLikes
	Err     error
}

type GameRow struct {
	ID   int
	Slug string
	Name string
}

func (game GameLike) ToRow() RowLike {
	return GameRow{
		ID:   game.ID,
		Slug: game.Slug,
		Name: game.Name,
	}
}

func (row GameRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO games (id, slug, name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING;",
		row.ID,
		row.Slug,
		row.Name,
	)
	return err
}

type LeagueRow struct {
	ID        int
	Slug      string
	GameID    int
	Name      string
	ImageLink string
}

func (league LeagueLike) ToRow() RowLike {
	return LeagueRow{
		ID:        league.ID,
		Slug:      league.Slug,
		GameID:    league.Videogame.ID,
		Name:      league.Name,
		ImageLink: league.ImageURL,
	}
}

func (row LeagueRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO leagues (id,slug, game_id, name, image_link) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING;",
		row.ID,
		row.Slug,
		row.GameID,
		row.Name,
		row.ImageLink,
	)
	return err
}

type SeriesRow struct {
	ID       int
	Slug     string
	GameID   int
	LeagueID int
	Name     string
}

func (series SeriesLike) ToRow() RowLike {
	return SeriesRow{
		ID:       series.ID,
		Slug:     series.Slug,
		GameID:   series.Videogame.ID,
		LeagueID: series.League.ID,
		Name:     series.Name,
	}
}

func (row SeriesRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO series (id,slug, game_id, league_id, name) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING;",
		row.ID,
		row.Slug,
		row.GameID,
		row.LeagueID,
		row.Name,
	)
	return err
}

type TournamentRow struct {
	ID       int
	Slug     string
	GameID   int
	SerieID  int
	LeagueID int
	Tier     int
	Name     string
}

func (tournament TournamentLike) ToRow() RowLike {
	var tier int
	switch tournament.Tier {
	case "S":
		tier = 1
	case "A":
		tier = 2
	case "B":
		tier = 3
	case "C":
		tier = 4
	case "D":
		tier = 5
	default:
		tier = 6
	}
	return TournamentRow{
		ID:       tournament.ID,
		Slug:     tournament.Slug,
		GameID:   tournament.Videogame.ID,
		SerieID:  tournament.Serie.ID,
		LeagueID: tournament.League.ID,
		Tier:     tier,
		Name:     tournament.Name,
	}
}

func (row TournamentRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO tournaments (id, slug, game_id, serie_id, league_id, tier, name) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO NOTHING;",
		row.ID,
		row.Slug,
		row.GameID,
		row.SerieID,
		row.LeagueID,
		row.Tier,
		row.Name,
	)
	return err
}

type MatchRow struct {
	ID                int
	Slug              string
	Finished          bool
	GameID            int
	LeagueID          int
	SerieID           int
	TournamentID      int
	Team1ID           int
	Team1Score        int
	Team2ID           int
	Team2Score        int
	Name              string
	ExpectedStartTime time.Time
	AmountOfGames     int
	ActualGameTime    float64
}

func (match MatchLike) ToRow() RowLike {
	actualGT := 0.0
	if match.EndAt != (time.Time{}) {
		actualGT = match.EndAt.Sub(match.BeginAt).Seconds() / float64(match.NumberOfGames)
	}
	t1ID := 0
	t1Score := 0
	t2ID := 0
	t2Score := 0
	if len(match.Opponents) == twoTeams {
		t1ID = match.Opponents[0].Opponent.ID
		t1Score = match.Results[0].Score
		t2ID = match.Opponents[1].Opponent.ID
		t2Score = match.Results[1].Score
	}
	return MatchRow{
		ID:                match.ID,
		Slug:              match.Slug,
		Finished:          match.EndAt != time.Time{},
		GameID:            match.Videogame.ID,
		LeagueID:          match.League.ID,
		SerieID:           match.Serie.ID,
		TournamentID:      match.Tournament.ID,
		Team1ID:           t1ID,
		Team1Score:        t1Score,
		Team2ID:           t2ID,
		Team2Score:        t2Score,
		Name:              match.Name,
		ExpectedStartTime: match.BeginAt,
		AmountOfGames:     match.NumberOfGames,
		ActualGameTime:    actualGT,
	}
}

func (row MatchRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO matches (id,slug, finished,game_id, league_id, series_id, tournament_id, Team1_id, Team1_score, Team2_id, Team2_score, name, expected_start_time, amount_of_games, actual_game_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) ON CONFLICT (id) DO NOTHING;",
		row.ID,
		row.Slug,
		row.Finished,
		row.GameID,
		row.LeagueID,
		row.SerieID,
		row.TournamentID,
		row.Team1ID,
		row.Team1Score,
		row.Team2ID,
		row.Team2Score,
		row.Name,
		row.ExpectedStartTime,
		row.AmountOfGames,
		row.ActualGameTime,
	)
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

func (row TeamRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec(
		"INSERT INTO teams (id,slug, game_id, name, acronym, image_link) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (id) DO NOTHING;",
		row.ID,
		row.Slug,
		row.GameID,
		row.Name,
		row.Acronym,
		row.ImageLink,
	)
	return err
}
