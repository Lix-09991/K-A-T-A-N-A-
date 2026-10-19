-- ============================================================
-- FRANCHISES
-- ============================================================
CREATE TABLE franchises (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT UNIQUE NOT NULL,
    title       TEXT NOT NULL,
    description TEXT,
    cover       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- ALTER ANIME — новые поля
-- ============================================================
ALTER TABLE anime
    ADD COLUMN country        TEXT   NOT NULL DEFAULT 'JP',
    ADD COLUMN type           TEXT   NOT NULL DEFAULT 'TV',
    ADD COLUMN views          BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN comments_count INT    NOT NULL DEFAULT 0,
    ADD COLUMN franchise_id   UUID   REFERENCES franchises(id) ON DELETE SET NULL;

-- ============================================================
-- FRANCHISE ITEMS (M:N с position — порядок просмотра)
-- ============================================================
CREATE TABLE franchise_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    franchise_id UUID NOT NULL REFERENCES franchises(id) ON DELETE CASCADE,
    anime_id     UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    position     INT  NOT NULL,
    note         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (franchise_id, anime_id),
    UNIQUE (franchise_id, position)
);

-- ============================================================
-- COLLECTIONS (подборки)
-- ============================================================
CREATE TABLE collections (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        TEXT UNIQUE NOT NULL,
    title       TEXT NOT NULL,
    description TEXT,
    cover       TEXT,
    is_featured BOOLEAN NOT NULL DEFAULT false,
    position    INT     NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE collection_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    anime_id      UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    position      INT  NOT NULL,
    note          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (collection_id, anime_id),
    UNIQUE (collection_id, position)
);

-- ============================================================
-- ТРИГГЕР: comments_count на anime
-- ============================================================
CREATE OR REPLACE FUNCTION update_anime_comments_count()
RETURNS TRIGGER AS $$
DECLARE
    target_anime_id UUID;
BEGIN
    IF TG_OP = 'INSERT' THEN
        target_anime_id := NEW.anime_id;
        IF target_anime_id IS NOT NULL THEN
            UPDATE anime SET comments_count = comments_count + 1 WHERE id = target_anime_id;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        target_anime_id := OLD.anime_id;
        IF target_anime_id IS NOT NULL THEN
            UPDATE anime SET comments_count = comments_count - 1 WHERE id = target_anime_id;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER comments_count_trigger
AFTER INSERT OR DELETE ON comments
FOR EACH ROW
EXECUTE FUNCTION update_anime_comments_count();

-- ============================================================
-- updated_at триггеры 
-- ============================================================
CREATE TRIGGER franchises_updated_at BEFORE UPDATE ON franchises
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER collections_updated_at BEFORE UPDATE ON collections
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();