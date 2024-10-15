CREATE TABLE IF NOT EXISTS users
(
    id         BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT NOT NULL,
    chat_id    BIGINT UNSIGNED                            NOT NULL,
    phone      varchar(12)                                NOT NULL,
    created_at TIMESTAMP                                  NOT NULL
);