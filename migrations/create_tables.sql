-- Migration script to create separate 1v1 and 2v2 tables
-- Run this script in your PostgreSQL database

-- Create 1v1 table
CREATE TABLE IF NOT EXISTS rocketleague_1v1 (
    "id" SERIAL PRIMARY KEY,
    "Name" VARCHAR(255) NOT NULL,
    "MMR" DOUBLE PRECISION NOT NULL DEFAULT 1000,
    "Wins" INTEGER NOT NULL DEFAULT 0,
    "Losses" INTEGER NOT NULL DEFAULT 0,
    "MatchUID" UUID,
    "DiscordId" INTEGER UNIQUE NOT NULL
);

-- Create 2v2 table
CREATE TABLE IF NOT EXISTS rocketleague_2v2 (
    "id" SERIAL PRIMARY KEY,
    "Name" VARCHAR(255) NOT NULL,
    "MMR" DOUBLE PRECISION NOT NULL DEFAULT 1000,
    "Wins" INTEGER NOT NULL DEFAULT 0,
    "Losses" INTEGER NOT NULL DEFAULT 0,
    "MatchUID" UUID,
    "DiscordId" INTEGER UNIQUE NOT NULL
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_rocketleague_1v1_discordid ON rocketleague_1v1("DiscordId");
CREATE INDEX IF NOT EXISTS idx_rocketleague_1v1_mmr ON rocketleague_1v1("MMR" DESC);
CREATE INDEX IF NOT EXISTS idx_rocketleague_2v2_discordid ON rocketleague_2v2("DiscordId");
CREATE INDEX IF NOT EXISTS idx_rocketleague_2v2_mmr ON rocketleague_2v2("MMR" DESC);

-- Note: After running this script, you can manually delete the old 'rocketleague' table:
-- DROP TABLE IF EXISTS rocketleague;

