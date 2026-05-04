CREATE TABLE users_events (
  id           CHAR(36)    NOT NULL PRIMARY KEY,
  aggregate_id CHAR(36)    NOT NULL,
  type         VARCHAR(64) NOT NULL,
  payload      JSON        NOT NULL,
  created_at   DATETIME(6) NOT NULL,
  INDEX idx_aggregate_created (aggregate_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE users_view (
  id                 CHAR(36)     NOT NULL PRIMARY KEY,
  name               VARCHAR(255) NOT NULL,
  email              VARCHAR(255) NOT NULL,
  bio                TEXT         NOT NULL,
  status             VARCHAR(32)  NOT NULL,
  verification_token VARCHAR(64)  NOT NULL,
  credit_score       INT          NOT NULL DEFAULT 0,
  credit_approved    BOOLEAN      NOT NULL DEFAULT FALSE,
  created_at         DATETIME(6)  NOT NULL,
  UNIQUE KEY uniq_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
