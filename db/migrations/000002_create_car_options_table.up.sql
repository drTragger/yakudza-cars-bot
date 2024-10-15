CREATE TABLE IF NOT EXISTS car_options
(
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT NOT NULL,
    title       VARCHAR(255)                               NOT NULL,
    description VARCHAR(500)                               NOT NULL,
    price       INT                                        NOT NULL,
    year        YEAR                                       NOT NULL,
    photo_id    VARCHAR(255)                               NOT NULL,
    created_at  TIMESTAMP                                  NOT NULL
);