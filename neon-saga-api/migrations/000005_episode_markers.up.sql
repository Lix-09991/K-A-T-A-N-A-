CREATE TABLE IF NOT EXISTS episode_markers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    type       TEXT NOT NULL CHECK (type IN ('op', 'ed', 'skip', 'custom')),
    start_time INT NOT NULL CHECK (start_time >= 0),
    end_time   INT CHECK (end_time IS NULL OR end_time >= start_time),
    label      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_episode_markers_episode ON episode_markers(episode_id);
CREATE INDEX IF NOT EXISTS idx_episode_markers_type ON episode_markers(type);

DROP TRIGGER IF EXISTS episode_markers_updated_at ON episode_markers;
CREATE TRIGGER episode_markers_updated_at BEFORE UPDATE ON episode_markers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
