DO
$do$
BEGIN
    IF EXISTS (
        SELECT FROM pg_catalog.pg_roles
        WHERE  rolname = 'reader') THEN
    RAISE NOTICE 'Role "reader" already exists. Skipping.';
    ELSE
        CREATE ROLE reader LOGIN PASSWORD '%s';
        GRANT pg_read_all_data TO reader;

    END IF;
    IF EXISTS (
        SELECT FROM pg_catalog.pg_roles
        WHERE  rolname = 'writer') THEN
    RAISE NOTICE 'Role "writer" already exists. Skipping.';
    ELSE
        CREATE ROLE writer LOGIN PASSWORD '%s';
        GRANT pg_read_all_data TO writer;
        GRANT pg_write_all_data TO writer;
    END IF;
END
$do$;

CREATE TABLE IF NOT EXISTS GAMES(
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS LEAGUES(
    id INTEGER PRIMARY KEY,
    slug VARCHAR(255),
    game_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    image_link VARCHAR(255) NOT NULL,
    FOREIGN KEY (game_id) REFERENCES GAMES(id)
);

CREATE TABLE IF NOT EXISTS SERIES(
    id INTEGER PRIMARY KEY,
    slug VARCHAR(255),
    game_id INT NOT NULL,
    league_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    FOREIGN KEY (game_id) REFERENCES GAMES(id),
    FOREIGN KEY (league_id) REFERENCES LEAGUES(id)
);

CREATE TABLE IF NOT EXISTS TOURNAMENTS(
    id INTEGER PRIMARY KEY,
    slug VARCHAR(255),
    game_id INT NOT NULL,
    serie_id INT NOT NULL,
    league_id INT NOT NULL,
    tier INT,
    name VARCHAR(255) NOT NULL,
    FOREIGN KEY (game_id) REFERENCES GAMES(id),
    FOREIGN KEY (serie_id) REFERENCES SERIES(id),
    FOREIGN KEY (league_id) REFERENCES LEAGUES(id)
);

CREATE TABLE IF NOT EXISTS MATCHES(
    id INTEGER PRIMARY KEY,
    slug VARCHAR(255),
    finished BOOLEAN NOT NULL,
    game_id INT NOT NULL,
    league_id INT NOT NULL,
    series_id INT NOT NULL,
    tournament_id INT NOT NULL,
    team1_id INT NOT NULL,
    team1_score INT NOT NULL,
    team2_id INT NOT NULL,
    team2_score INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    expected_start_time TIMESTAMP NOT NULL,
    amount_of_games INT NOT NULL,-- analytics
    actual_game_time FLOAT NOT NULL,  
    FOREIGN KEY (game_id) REFERENCES GAMES(id),
    FOREIGN KEY (league_id) REFERENCES LEAGUES(id),
    FOREIGN KEY (series_id) REFERENCES SERIES(id),
    FOREIGN KEY (tournament_id) REFERENCES TOURNAMENTS(id)
);

CREATE TABLE IF NOT EXISTS TEAMS(
    id INTEGER PRIMARY KEY,
    slug VARCHAR(255),
    game_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    acronym VARCHAR(255),
    image_link VARCHAR(255),
    FOREIGN KEY (game_id) REFERENCES GAMES(id)
);
COMMIT;