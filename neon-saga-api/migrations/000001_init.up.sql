-- ============================================================
-- USERS
-- ============================================================
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT UNIQUE NOT NULL,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    avatar        TEXT,
    role          TEXT NOT NULL DEFAULT 'USER' CHECK (role IN ('USER', 'MODERATOR', 'ADMIN')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- ============================================================
-- ANIME
-- ============================================================
CREATE TABLE anime (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT UNIQUE NOT NULL,
    title       TEXT NOT NULL,
    title_en    TEXT,
    description TEXT NOT NULL,
    cover       TEXT,
    banner      TEXT,
    year        INT NOT NULL,
    rating      REAL NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'ONGOING' CHECK (status IN ('ONGOING', 'COMPLETED', 'ANNOUNCED')),
    genres      TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_anime_slug ON anime(slug);
CREATE INDEX idx_anime_year ON anime(year);
CREATE INDEX idx_anime_rating ON anime(rating DESC);
CREATE INDEX idx_anime_genres ON anime USING GIN(genres);

-- ============================================================
-- EPISODES
-- ============================================================
CREATE TABLE episodes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    anime_id    UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    number      INT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT,
    duration    INT NOT NULL DEFAULT 0,
    video_url   TEXT NOT NULL,
    thumbnail   TEXT,
    aired_at    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (anime_id, number)
);

CREATE INDEX idx_episodes_anime ON episodes(anime_id);

-- ============================================================
-- COMMENTS
-- ============================================================
CREATE TABLE comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id   UUID REFERENCES anime(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (anime_id IS NOT NULL OR episode_id IS NOT NULL)
);

CREATE INDEX idx_comments_anime ON comments(anime_id);
CREATE INDEX idx_comments_episode ON comments(episode_id);
CREATE INDEX idx_comments_user ON comments(user_id);

-- ============================================================
-- RATINGS
-- ============================================================
CREATE TABLE ratings (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anime_id   UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    score      INT NOT NULL CHECK (score BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, anime_id)
);

CREATE INDEX idx_ratings_anime ON ratings(anime_id);

-- ============================================================
-- WATCH HISTORY
-- ============================================================
CREATE TABLE watch_history (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    progress   INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, episode_id)
);

CREATE INDEX idx_watch_history_user ON watch_history(user_id);

-- ============================================================
-- TRIGGER для updated_at
-- ============================================================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER anime_updated_at BEFORE UPDATE ON anime
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER comments_updated_at BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();