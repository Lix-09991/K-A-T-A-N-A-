-- ============================================================
-- SCALAR — для сортировок
-- ============================================================
CREATE INDEX idx_anime_created_at     ON anime(created_at DESC);
CREATE INDEX idx_anime_updated_at     ON anime(updated_at DESC);
CREATE INDEX idx_anime_views          ON anime(views DESC);
CREATE INDEX idx_anime_comments_count ON anime(comments_count DESC);
-- rating / year / status уже есть в миграции #1

-- ============================================================
-- SCALAR — для фильтров
-- ============================================================
CREATE INDEX idx_anime_country ON anime(country);
CREATE INDEX idx_anime_type    ON anime(type);
CREATE INDEX idx_anime_franchise ON anime(franchise_id) WHERE franchise_id IS NOT NULL;

-- ============================================================
-- COMPOSITE — частые комбо
-- ============================================================
CREATE INDEX idx_anime_type_rating    ON anime(type, rating DESC);
CREATE INDEX idx_anime_status_created ON anime(status, created_at DESC);
CREATE INDEX idx_anime_country_type   ON anime(country, type);

-- ============================================================
-- TRIGRAM — быстрый ILIKE (в т.ч. с опечатками)
-- ============================================================
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_anime_title_trgm    ON anime USING GIN(title gin_trgm_ops);
CREATE INDEX idx_anime_title_en_trgm ON anime USING GIN(title_en gin_trgm_ops);

-- ============================================================
-- FRANCHISE ITEMS
-- ============================================================
CREATE INDEX idx_franchise_items_franchise ON franchise_items(franchise_id, position);
CREATE INDEX idx_franchise_items_anime     ON franchise_items(anime_id);

-- ============================================================
-- COLLECTIONS
-- ============================================================
CREATE INDEX idx_collection_items_collection ON collection_items(collection_id, position);
CREATE INDEX idx_collection_items_anime      ON collection_items(anime_id);
CREATE INDEX idx_collections_featured        ON collections(is_featured, position)
    WHERE is_featured = true;