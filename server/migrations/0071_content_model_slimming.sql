-- +goose Up

-- ===========================================================================
-- 内容模型精简（二次开发分叉，删除即真删，不做兼容）
--
-- 本批次包含四项改动，顺序即依赖顺序：
--   A. 删除「思考」
--   B. 删除「页面管理」+ 路由评论区解耦
--   C. 「文章」并入「手记」；「文章分类」并入「专栏」
--   D. 删除「导航菜单」，前台导航改硬编码
--
-- 说明：A 段中本应删除 page 的内置登记行（short_url='thinkings'）与
-- nav_menu 的「思考」入口，但 page / nav_menu 两张表会在 B、D 段整体删除，
-- 故此处不重复删除，避免冗余语句。
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- A. 删除「思考」
-- ---------------------------------------------------------------------------

-- A1. 触发器函数去掉 thinking 分支。
--     函数体内引用 thinking_metrics，表删除后分支若被命中会报错，必须一并重写。
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION sync_content_like_metrics()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.target_type = 'article' THEN
            INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (article_id)
            DO UPDATE SET likes = article_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'moment' THEN
            INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (moment_id)
            DO UPDATE SET likes = moment_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'page' THEN
            INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (page_id)
            DO UPDATE SET likes = page_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'album' THEN
            INSERT INTO album_metrics (album_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (album_id)
            DO UPDATE SET likes = album_metrics.likes + 1, updated_at = NOW();
        END IF;
        RETURN NEW;
    END IF;

    IF TG_OP = 'DELETE' THEN
        IF OLD.target_type = 'article' THEN
            UPDATE article_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE article_id = OLD.target_id;
        ELSIF OLD.target_type = 'moment' THEN
            UPDATE moment_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE moment_id = OLD.target_id;
        ELSIF OLD.target_type = 'page' THEN
            UPDATE page_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE page_id = OLD.target_id;
        ELSIF OLD.target_type = 'album' THEN
            UPDATE album_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE album_id = OLD.target_id;
        END IF;
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION adjust_comment_metrics_by_area(p_area_id BIGINT, p_delta INTEGER)
RETURNS void AS $$
DECLARE
    v_area_type VARCHAR(20);
    v_content_id BIGINT;
BEGIN
    IF p_area_id IS NULL OR p_delta = 0 THEN
        RETURN;
    END IF;

    SELECT area_type, content_id
    INTO v_area_type, v_content_id
    FROM comment_area
    WHERE id = p_area_id;

    IF v_content_id IS NULL THEN
        RETURN;
    END IF;

    IF v_area_type = 'article' THEN
        INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (article_id)
        DO UPDATE SET comments = GREATEST(article_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'moment' THEN
        INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (moment_id)
        DO UPDATE SET comments = GREATEST(moment_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'page' THEN
        INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (page_id)
        DO UPDATE SET comments = GREATEST(page_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'album' THEN
        INSERT INTO album_metrics (album_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (album_id)
        DO UPDATE SET comments = GREATEST(album_metrics.comments + p_delta, 0), updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- A2. 删除思考的评论区与评论（思考整体移除，其评论一并删除）
DELETE FROM comment
WHERE area_id IN (SELECT id FROM comment_area WHERE area_type = 'thinking');

DELETE FROM comment_area WHERE area_type = 'thinking';

-- A3. 站点配置去掉 thinking 推送类型（value 与 default_value 同步）
UPDATE sys_config
SET value         = '["article","moment"]',
    default_value = '["article","moment"]',
    updated_at    = NOW()
WHERE config_key = 'activitypub.publishTypes';

-- A4. 导出记录去掉思考计数字段
ALTER TABLE export_ops.export_record
    DROP COLUMN IF EXISTS thinkings_count;

-- A5. 删除思考表（thinking_metrics 先删，避免外键依赖）
DROP TABLE IF EXISTS thinking_metrics;
DROP TABLE IF EXISTS thinking;

-- ---------------------------------------------------------------------------
-- B. 删除「页面管理」
--
-- 说明：原方案此段还包含「路由评论区解耦」（新建 route_comment_area 表，把
-- 「路由 → 评论区」从 page 中剥离）。实测当前分叉中该前提不成立：
--   1. 前台挂载评论区的位置只有内容详情页（文章 / 手记 / 页面 / 相册），
--      areaId 来自内容自身的 comment_id 字段，从不经过内置路由；
--   2. 友链页、时间线页、标签页、统计页均无评论组件；
--   3. 全前台不存在「按内置路由 short_url 取 page 详情拿 commentId」的调用链；
--   4. page.comment_id 全为 NULL，comment_area 为空表。
-- 故 route_comment_area 建成即死代码，经确认跳过，不建。
-- 因此删除 page 不会破坏任何评论链路。
-- ---------------------------------------------------------------------------

-- B1. 触发器函数去掉 page 分支。
--     同 A1：函数体引用 page_metrics，表删除后分支若被命中会报错，必须一并重写。
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION sync_content_like_metrics()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.target_type = 'article' THEN
            INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (article_id)
            DO UPDATE SET likes = article_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'moment' THEN
            INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (moment_id)
            DO UPDATE SET likes = moment_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'album' THEN
            INSERT INTO album_metrics (album_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (album_id)
            DO UPDATE SET likes = album_metrics.likes + 1, updated_at = NOW();
        END IF;
        RETURN NEW;
    END IF;

    IF TG_OP = 'DELETE' THEN
        IF OLD.target_type = 'article' THEN
            UPDATE article_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE article_id = OLD.target_id;
        ELSIF OLD.target_type = 'moment' THEN
            UPDATE moment_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE moment_id = OLD.target_id;
        ELSIF OLD.target_type = 'album' THEN
            UPDATE album_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE album_id = OLD.target_id;
        END IF;
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION adjust_comment_metrics_by_area(p_area_id BIGINT, p_delta INTEGER)
RETURNS void AS $$
DECLARE
    v_area_type VARCHAR(20);
    v_content_id BIGINT;
BEGIN
    IF p_area_id IS NULL OR p_delta = 0 THEN
        RETURN;
    END IF;

    SELECT area_type, content_id
    INTO v_area_type, v_content_id
    FROM comment_area
    WHERE id = p_area_id;

    IF v_content_id IS NULL THEN
        RETURN;
    END IF;

    IF v_area_type = 'article' THEN
        INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (article_id)
        DO UPDATE SET comments = GREATEST(article_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'moment' THEN
        INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (moment_id)
        DO UPDATE SET comments = GREATEST(moment_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'album' THEN
        INSERT INTO album_metrics (album_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (album_id)
        DO UPDATE SET comments = GREATEST(album_metrics.comments + p_delta, 0), updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- B2. 删除页面的评论区与评论（页面整体移除，其评论一并删除）
DELETE FROM comment
WHERE area_id IN (SELECT id FROM comment_area WHERE area_type = 'page');

DELETE FROM comment_area WHERE area_type = 'page';

-- B3. 删除页面表（page_metrics 先删，避免外键依赖；idx_page_search_fts 随表删除）
DROP TABLE IF EXISTS page_metrics;
DROP TABLE IF EXISTS page;

-- B4. 导出记录去掉页面计数字段
ALTER TABLE export_ops.export_record
    DROP COLUMN IF EXISTS pages_count;

-- ---------------------------------------------------------------------------
-- C / D 段将在后续阶段追加到本文件
-- ---------------------------------------------------------------------------


-- +goose Down

-- 本迁移不可逆：思考数据已随表删除，无法还原。
-- Down 仅恢复结构，数据不回填（二次开发分叉，无兼容要求）。

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS thinking
(
    id         BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    content    TEXT        NOT NULL,
    author_id  BIGINT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    comment_id BIGINT,

    CONSTRAINT fk_thinking_author FOREIGN KEY (author_id) REFERENCES app_user (id) ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS thinking_metrics
(
    thinking_id BIGINT PRIMARY KEY
        REFERENCES thinking (id) ON DELETE CASCADE,
    views       BIGINT      NOT NULL DEFAULT 0,
    likes       INTEGER     NOT NULL DEFAULT 0,
    comments    INTEGER     NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ DEFAULT now()
);
-- +goose StatementEnd

ALTER TABLE export_ops.export_record
    ADD COLUMN IF NOT EXISTS thinkings_count BIGINT NOT NULL DEFAULT 0;

UPDATE sys_config
SET value         = '["article","moment","thinking"]',
    default_value = '["article","moment","thinking"]',
    updated_at    = NOW()
WHERE config_key = 'activitypub.publishTypes';

-- B 段 Down：恢复 page / page_metrics 结构（数据不回填）。
-- 同时把两个触发器函数还原为 0071 之前的完整版本（含 thinking 与 page 分支）——
-- 本文件的 Down 段按 A→B 顺序执行，此处位于最后，故最终状态即为迁移前状态。

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS page
(
    id                 BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    title              VARCHAR(255) NOT NULL,
    description        VARCHAR(255),
    ai_summary         TEXT,
    short_url          VARCHAR(255) NOT NULL,
    is_enabled         BOOLEAN      DEFAULT TRUE,
    is_builtin         BOOLEAN      DEFAULT FALSE,
    toc                JSONB        NOT NULL,
    content            TEXT         NOT NULL,
    comment_id         BIGINT,
    created_at         TIMESTAMPTZ  DEFAULT now(),
    updated_at         TIMESTAMPTZ  DEFAULT now(),
    deleted_at         TIMESTAMPTZ,
    content_hash       VARCHAR(32)  NOT NULL DEFAULT '',
    ext_info           JSONB,
    content_updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_page_short_url UNIQUE (short_url),

    FOREIGN KEY (comment_id) REFERENCES comment_area (id)
);
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_page_short_url ON page (short_url);

CREATE INDEX IF NOT EXISTS idx_page_search_fts ON page USING gin (((
    setweight(to_tsvector('simple'::regconfig, (COALESCE(title, ''::character varying))::text), 'A'::"char") ||
    setweight(to_tsvector('simple'::regconfig, (COALESCE(description, ''::character varying))::text), 'B'::"char")
) || setweight(to_tsvector('simple'::regconfig, COALESCE(content, ''::text)), 'C'::"char")));

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS page_metrics
(
    page_id    BIGINT PRIMARY KEY
        REFERENCES page (id) ON DELETE CASCADE,
    views      BIGINT      NOT NULL DEFAULT 0,
    likes      INTEGER     NOT NULL DEFAULT 0,
    comments   INTEGER     NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT now()
);
-- +goose StatementEnd

ALTER TABLE export_ops.export_record
    ADD COLUMN IF NOT EXISTS pages_count BIGINT NOT NULL DEFAULT 0;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION sync_content_like_metrics()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.target_type = 'article' THEN
            INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (article_id)
            DO UPDATE SET likes = article_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'moment' THEN
            INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (moment_id)
            DO UPDATE SET likes = moment_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'page' THEN
            INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (page_id)
            DO UPDATE SET likes = page_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'thinking' THEN
            INSERT INTO thinking_metrics (thinking_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (thinking_id)
            DO UPDATE SET likes = thinking_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'album' THEN
            INSERT INTO album_metrics (album_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (album_id)
            DO UPDATE SET likes = album_metrics.likes + 1, updated_at = NOW();
        END IF;
        RETURN NEW;
    END IF;

    IF TG_OP = 'DELETE' THEN
        IF OLD.target_type = 'article' THEN
            UPDATE article_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE article_id = OLD.target_id;
        ELSIF OLD.target_type = 'moment' THEN
            UPDATE moment_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE moment_id = OLD.target_id;
        ELSIF OLD.target_type = 'page' THEN
            UPDATE page_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE page_id = OLD.target_id;
        ELSIF OLD.target_type = 'thinking' THEN
            UPDATE thinking_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE thinking_id = OLD.target_id;
        ELSIF OLD.target_type = 'album' THEN
            UPDATE album_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE album_id = OLD.target_id;
        END IF;
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION adjust_comment_metrics_by_area(p_area_id BIGINT, p_delta INTEGER)
RETURNS void AS $$
DECLARE
    v_area_type VARCHAR(20);
    v_content_id BIGINT;
BEGIN
    IF p_area_id IS NULL OR p_delta = 0 THEN
        RETURN;
    END IF;

    SELECT area_type, content_id
    INTO v_area_type, v_content_id
    FROM comment_area
    WHERE id = p_area_id;

    IF v_content_id IS NULL THEN
        RETURN;
    END IF;

    IF v_area_type = 'article' THEN
        INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (article_id)
        DO UPDATE SET comments = GREATEST(article_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'moment' THEN
        INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (moment_id)
        DO UPDATE SET comments = GREATEST(moment_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'page' THEN
        INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (page_id)
        DO UPDATE SET comments = GREATEST(page_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'thinking' THEN
        INSERT INTO thinking_metrics (thinking_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (thinking_id)
        DO UPDATE SET comments = GREATEST(thinking_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'album' THEN
        INSERT INTO album_metrics (album_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (album_id)
        DO UPDATE SET comments = GREATEST(album_metrics.comments + p_delta, 0), updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
