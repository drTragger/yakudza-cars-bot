CREATE TABLE feedback
(
    id            INT AUTO_INCREMENT PRIMARY KEY      NOT NULL,
    description   VARCHAR(600)                        NOT NULL,
    video_file_id VARCHAR(255)                        NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);