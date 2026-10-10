DROP INDEX IF EXISTS idx_collections_featured;
DROP INDEX IF EXISTS idx_collection_items_anime;
DROP INDEX IF EXISTS idx_collection_items_collection;
DROP INDEX IF EXISTS idx_franchise_items_anime;
DROP INDEX IF EXISTS idx_franchise_items_franchise;
DROP INDEX IF EXISTS idx_anime_title_en_trgm;
DROP INDEX IF EXISTS idx_anime_title_trgm;
DROP INDEX IF EXISTS idx_anime_country_type;
DROP INDEX IF EXISTS idx_anime_status_created;
DROP INDEX IF EXISTS idx_anime_type_rating;
DROP INDEX IF EXISTS idx_anime_franchise;
DROP INDEX IF EXISTS idx_anime_type;
DROP INDEX IF EXISTS idx_anime_country;
DROP INDEX IF EXISTS idx_anime_comments_count;
DROP INDEX IF EXISTS idx_anime_views;
DROP INDEX IF EXISTS idx_anime_updated_at;
DROP INDEX IF EXISTS idx_anime_created_at;
-- pg_trgm extension НЕ удаляем — может использоваться другими