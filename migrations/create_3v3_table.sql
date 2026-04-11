-- 3v3 player stats table (same shape as 1v1 / 2v2)

CREATE TABLE IF NOT EXISTS rocketleague_3v3 (
    "id" SERIAL PRIMARY KEY,
    "Name" VARCHAR(255) NOT NULL,
    "MMR" DOUBLE PRECISION NOT NULL DEFAULT 1000,
    "Wins" INTEGER NOT NULL DEFAULT 0,
    "Losses" INTEGER NOT NULL DEFAULT 0,
    "MatchUID" UUID,
    "DiscordId" INTEGER UNIQUE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rocketleague_3v3_discordid ON rocketleague_3v3("DiscordId");
CREATE INDEX IF NOT EXISTS idx_rocketleague_3v3_mmr ON rocketleague_3v3("MMR" DESC);
