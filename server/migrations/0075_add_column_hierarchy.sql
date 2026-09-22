-- +goose Up
-- Core taxonomy data: existing categories remain roots, and article ownership is unchanged.
ALTER TABLE moment_column ADD COLUMN parent_id BIGINT;
ALTER TABLE moment_column
    ADD CONSTRAINT fk_moment_column_parent FOREIGN KEY (parent_id)
        REFERENCES moment_column(id) ON DELETE RESTRICT,
    ADD CONSTRAINT ck_moment_column_not_self CHECK (parent_id IS NULL OR parent_id <> id);
CREATE INDEX idx_moment_column_parent ON moment_column(parent_id) WHERE deleted_at IS NULL;

-- +goose Down
-- Refuse to discard hierarchy data when rolling back.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM moment_column WHERE parent_id IS NOT NULL) THEN
        RAISE EXCEPTION 'Reassign child categories before rolling back the hierarchy';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE moment_column DROP CONSTRAINT ck_moment_column_not_self;
ALTER TABLE moment_column DROP CONSTRAINT fk_moment_column_parent;
DROP INDEX idx_moment_column_parent;
ALTER TABLE moment_column DROP COLUMN parent_id;
