-- Database setup untuk Telegram AI Bot
-- PolarDB MySQL 8.0 Compatible

-- Buat database jika belum ada
CREATE DATABASE IF NOT EXISTS telegram_bot 
CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;

USE telegram_bot;

-- Tabel untuk menyimpan riwayat percakapan
CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    user_name VARCHAR(255),
    message TEXT NOT NULL,
    response TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabel untuk menyimpan session data
CREATE TABLE IF NOT EXISTS chat_sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL UNIQUE,
    session_data JSON,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_last_activity (last_activity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabel untuk menyimpan memory permanen user (informasi personal)
CREATE TABLE IF NOT EXISTS user_memories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    memory_key VARCHAR(100) NOT NULL,
    memory_value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_memory (user_id, memory_key),
    INDEX idx_user_id (user_id),
    INDEX idx_memory_key (memory_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Contoh data untuk testing (opsional)
-- INSERT INTO conversations (user_id, user_name, message, response) VALUES
-- ('12345', 'TestUser', 'Halo', 'Halo juga! Ada yang bisa saya bantu?'),
-- ('12345', 'TestUser', 'Apa kabar?', 'Baik! Terima kasih sudah bertanya. Bagaimana dengan Anda?');

-- Tampilkan informasi tabel
SHOW TABLES;
DESCRIBE conversations;
DESCRIBE chat_sessions;
