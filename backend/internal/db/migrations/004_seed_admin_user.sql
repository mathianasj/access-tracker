-- Migration: Seed default admin user
-- Version: 004
-- Description: Create default admin user for initial access

INSERT INTO users (username, password_hash)
VALUES ('admin', '$2a$10$iWDneHqhkAC6ie3CkT3U7ePZbD2sWrVGj9QrjsCukkLdSayNHXUMq')
ON CONFLICT (username) DO NOTHING;
