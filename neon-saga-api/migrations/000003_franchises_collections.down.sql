DROP TRIGGER IF EXISTS collections_updated_at ON collections;
DROP TRIGGER IF EXISTS franchises_updated_at  ON franchises;
DROP TRIGGER IF EXISTS comments_count_trigger ON comments;
DROP FUNCTION IF EXISTS update_anime_comments_count();

DROP TABLE IF EXISTS collection_items;
DROP TABLE IF EXISTS collections;
DROP TABLE IF EXISTS franchise_items;
DROP TABLE IF EXISTS franchises;

ALTER TABLE anime
    DROP COLUMN IF EXISTS franchise_id,
    DROP COLUMN IF EXISTS comments_count,
    DROP COLUMN IF EXISTS views,
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS country;