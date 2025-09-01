-- Database creation
CREATE DATABASE IF NOT EXISTS secure_comm
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE secure_comm;

-- === Tables ===

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(150) NOT NULL UNIQUE,
    email VARCHAR(254) NOT NULL UNIQUE,
    password_hmac CHAR(64) NOT NULL,      -- HMAC-SHA256 hex (64)
    salt VARBINARY(16) NOT NULL,          -- 16 random bytes
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,  -- Email verification
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS password_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    password_hmac CHAR(64) NOT NULL,
    password_fp  CHAR(64) NULL,            -- SHA-256 hex of the password (for quick lookup)
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_ph_user_changed (user_id, changed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    token_sha1 CHAR(40) NOT NULL,         -- SHA-1 hex (40)
    expires_at DATETIME NOT NULL,
    used_at DATETIME NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_prt_token (token_sha1),
    INDEX idx_prt_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Email verification after registration
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    token_sha1 CHAR(40) NOT NULL,         -- Only the hash of the token is stored
    expires_at DATETIME NOT NULL,
    used_at DATETIME NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY uq_evt_token (token_sha1),
    INDEX idx_evt_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(40) NULL,
    notes TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS login_attempts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NULL,
    username VARCHAR(150) NOT NULL,
    ip VARCHAR(45) NULL,
    attempt_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_la_user_time (user_id, attempt_time),
    INDEX idx_la_username_time (username, attempt_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- === Email OTP for 2FA ===
CREATE TABLE IF NOT EXISTS login_otp_challenges (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  code_sha256 CHAR(64) NOT NULL,          -- store hash, never raw code
  expires_at DATETIME NOT NULL,
  consumed_at DATETIME NULL,
  attempts INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_login_otp_user (user_id),
  INDEX idx_login_otp_exp (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Pending password change requests (email confirmation flow)
CREATE TABLE IF NOT EXISTS password_change_requests (
  id               INT AUTO_INCREMENT PRIMARY KEY,
  user_id          INT NOT NULL,
  new_password_hmac CHAR(64) NOT NULL,   -- HMAC-SHA256 hex of the *new* password
  new_salt         VARBINARY(16) NOT NULL,
  token_sha1       CHAR(40) NOT NULL,    -- SHA-1 hex of confirmation token
  expires_at       DATETIME NOT NULL,
  used_at          DATETIME NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  UNIQUE KEY uq_pcr_token (token_sha1),
  INDEX idx_pcr_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- === Demo data (for development only!) ===
INSERT INTO users (username, email, password_hmac, salt)
VALUES ('admin', 'admin@example.com',
        REPEAT('a',64),
        UNHEX('00112233445566778899AABBCCDDEEFF'));

INSERT INTO customers (name, email, phone, notes)
VALUES ('First Customer', 'customer1@example.com', '050-1234567', 'Demo notes');
