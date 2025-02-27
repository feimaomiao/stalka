package JsonTypes

import (
	"database/sql"
	"time"
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

func (match MatchLike) WriteToDB(db *sql.DB) error {
	row := match.ToRow()
	_, err := db.Exec("INSERT INTO matches (id, finished, league_id, series_id, tournament_id, team1_id, team1_score, team2_id, team2_score, name, expected_start_time, amount_of_games, actual_game_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) ON CONFLICT (id) DO NOTHING;", row.ID, row.Finished, row.League_id, row.Series_id, row.Tournament_id,
		row.Team1_id, row.Team1_score, row.Team2_id, row.Team2_score, row.Name, row.Expected_start_time, row.Amount_of_games, row.Actual_game_time)
	return err

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
	Name string
	Slug string
}

func (game GameLike) ToRow() GameRow {
	return GameRow{
		ID:   game.ID,
		Name: game.Name,
		Slug: game.Slug,
	}
}

func (row GameRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO games (id, name, slug) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING;", row.ID, row.Name, row.Slug)
	return err
}

type LeagueRow struct {
	ID         int
	Game_id    int
	Name       string
	image_link string
}

func (league LeagueLike) ToRow() LeagueRow {
	return LeagueRow{
		ID:         league.ID,
		Game_id:    league.Videogame.ID,
		Name:       league.Name,
		image_link: league.ImageURL,
	}
}

func (row LeagueRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO leagues (id, game_id, name, image_link) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING;", row.ID, row.Game_id, row.Name, row.image_link)
	return err
}

type SeriesRow struct {
	ID        int
	Game_id   int
	League_id int
	Name      string
}

func (series SeriesLike) ToRow() SeriesRow {
	return SeriesRow{
		ID:        series.ID,
		Game_id:   series.Videogame.ID,
		League_id: series.League.ID,
		Name:      series.Name,
	}
}

func (row SeriesRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO series (id, game_id, league_id, name) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING;", row.ID, row.Game_id, row.League_id, row.Name)
	return err
}

type TournamentRow struct {
	ID        int
	Game_id   int
	Serie_id  int
	League_id int
	Name      string
}

func (series TournamentLike) ToRow() TournamentRow {
	return TournamentRow{
		ID:        series.ID,
		Game_id:   series.Videogame.ID,
		Serie_id:  series.Serie.ID,
		League_id: series.League.ID,
		Name:      series.Name,
	}
}

func (row TournamentRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO tournaments (id, game_id, serie_id, league_id, name) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING;", row.ID, row.Game_id, row.Serie_id, row.League_id, row.Name)
	return err
}

type MatchRow struct {
	ID                  int
	Finished            bool
	Game_id             int
	League_id           int
	Series_id           int
	Tournament_id       int
	Team1_id            int
	Team1_score         int
	Team2_id            int
	Team2_score         int
	Name                string
	Expected_start_time time.Time
	Amount_of_games     int
	Actual_game_time    float64
}

func (match MatchLike) ToRow() MatchRow {
	actualGT := 0.0
	if match.EndAt != (time.Time{}) {
		actualGT = match.EndAt.Sub(match.BeginAt).Seconds() / float64(match.NumberOfGames)
	}
	t1_id := 0
	t1_score := 0
	t2_id := 0
	t2_score := 0
	if len(match.Opponents) == 2 {
		t1_id = match.Opponents[0].Opponent.ID
		t1_score = match.Results[0].Score
		t2_id = match.Opponents[1].Opponent.ID
		t2_score = match.Results[1].Score
	}
	return MatchRow{
		ID:                  match.ID,
		Finished:            match.EndAt != time.Time{},
		Game_id:             match.Videogame.ID,
		League_id:           match.League.ID,
		Series_id:           match.Serie.ID,
		Tournament_id:       match.Tournament.ID,
		Team1_id:            t1_id,
		Team1_score:         t1_score,
		Team2_id:            t2_id,
		Team2_score:         t2_score,
		Name:                match.Name,
		Expected_start_time: match.BeginAt,
		Amount_of_games:     match.NumberOfGames,
		Actual_game_time:    actualGT,
	}
}

func (row MatchRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO matches (id, finished,game_id, league_id, series_id, tournament_id, Team1_id, Team1_score, Team2_id, Team2_score, name, expected_start_time, amount_of_games, actual_game_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) ON CONFLICT (id) DO NOTHING;",
		row.ID,
		row.Finished,
		row.Game_id,
		row.League_id,
		row.Series_id,
		row.Tournament_id,
		row.Team1_id,
		row.Team1_score,
		row.Team2_id,
		row.Team2_score,
		row.Name,
		row.Expected_start_time,
		row.Amount_of_games,
		row.Actual_game_time)
	return err
}

type TeamRow struct {
	ID         int
	Name       string
	Acronym    string
	Slug       string
	Image_link string
}

func (row TeamRow) WriteToDB(db *sql.DB) error {
	_, err := db.Exec("INSERT INTO teams (id, name,acronym, slug, image_link) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING;", row.ID, row.Name, row.Acronym, row.Slug, row.Image_link)
	return err
}
