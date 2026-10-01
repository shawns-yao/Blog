-- +goose Up
-- 留言区为长期核心记录；0 是站级占位标识，不引用文章或手记。
INSERT INTO comment_area (area_name, area_type, content_id, is_closed)
VALUES ('留言板', 'guestbook', 0, FALSE)
ON CONFLICT (area_type, content_id) DO NOTHING;

-- +goose Down
-- 回退仅关闭入口，保留留言及其审核、回复历史，不级联删除。
UPDATE comment_area
SET is_closed = TRUE, updated_at = now()
WHERE area_type = 'guestbook' AND content_id = 0;
