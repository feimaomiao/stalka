CREATE TABLE GAMES(
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255)
);

CREATE TABLE LEAGUES(
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    game_id INT NOT NULL,
    image_link VARCHAR(255),
    FOREIGN KEY (game_id) REFERENCES GAMES(id)
);

CREATE TABLE SERIES(
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    game_id INT NOT NULL,
    league_id INT NOT NULL,
    FOREIGN KEY (game_id) REFERENCES GAMES(id),
    FOREIGN KEY (league_id) REFERENCES LEAGUES(id)
);

CREATE TABLE TOURNAMENTS(
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    tier INT,
    game_id INT NOT NULL,
    serie_id INT NOT NULL,
    league_id INT NOT NULL,
    FOREIGN KEY (game_id) REFERENCES GAMES(id),
    FOREIGN KEY (serie_id) REFERENCES SERIES(id),
    FOREIGN KEY (league_id) REFERENCES LEAGUES(id)
);


CREATE TABLE IF NOT EXISTS MATCHES(
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    finished BOOLEAN NOT NULL,
    expected_start_time TIMESTAMP NOT NULL,
    team1_id INT NOT NULL,
    team1_score INT NOT NULL,
    team2_id INT NOT NULL,
    team2_score INT NOT NULL,
    amount_of_games INT NOT NULL,
    game_id INT NOT NULL,
    league_id INT NOT NULL,
    series_id INT NOT NULL,
    tournament_id INT NOT NULL,
    FOREIGN KEY (game_id) REFERENCES GAMES(id),
    FOREIGN KEY (league_id) REFERENCES LEAGUES(id),
    FOREIGN KEY (series_id) REFERENCES SERIES(id),
    FOREIGN KEY (tournament_id) REFERENCES TOURNAMENTS(id)
);

CREATE TABLE IF NOT EXISTS TEAMS(
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255),
    acronym VARCHAR(255),
    image_link VARCHAR(255),
    game_id INT NOT NULL,
    FOREIGN KEY (game_id) REFERENCES GAMES(id)
);