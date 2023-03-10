CREATE TABLE IF NOT EXISTS movies (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  title TEXT NOT NULL,
  year INT NOT NULL,
  runtime INT NOT NULL,
  genres TEXT NOT NULL,
  version INT NOT NULL DEFAULT 1
);