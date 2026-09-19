-- +goose Up

-- ===========================================================================
-- C. 「文章」并入手记
--
-- 目标：上游「一个实体按篇幅切三份」的模型收敛为「一个长文实体」。
--   保留：moment（手记，唯一长文实体）、moment_column（对外显示「专栏」）、
--         moment_topic（手记—标签关联）、tag
--   删除：article / article_metrics / article_category / article_tag
--
-- 字段口径（见方案 §2.1）：
--   moment 增加 cover（从 article 并入）
--   删除 article.lead_in（导语）、moment.img（配图）
--   —— 图片与引用直接写在正文 Markdown 里，不再单列字段
--
-- 当前两表数据均为 0 行，数据搬迁为空操作；语句仍按「保留 article.id
-- 作为 moment.id」的正确写法保留，以便将来在有数据的库上执行时引用不失效。
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- C1. moment 列调整：加 cover、删 img
--     必须先于数据搬迁——搬迁语句要写入 cover 列。
-- ---------------------------------------------------------------------------
ALTER TABLE moment ADD COLUMN IF NOT EXISTS cover VARCHAR(255);
ALTER TABLE moment DROP COLUMN IF EXISTS img;

-- ---------------------------------------------------------------------------
-- C2. 数据搬迁（必须先于外键改指，否则新外键校验不过）
-- ---------------------------------------------------------------------------

-- 保留 id：content_like.target_id / comment_area.content_id / 联邦表引用
-- 都按原 article.id 记录，重排 id 会让这些引用全部失配。
-- 若 moment 已存在同 id 行会触发唯一键冲突——当前为空表，无此风险。
-- +goose StatementBegin
INSERT INTO moment (
    id, title, summary, ai_summary, content, author_id, toc,
    cover, column_id, comment_id, short_url, is_published,
    created_at, updated_at, deleted_at,
    is_top, is_hot, is_original, content_hash, ext_info,
    activitypub_object_id, activitypub_last_published_at, content_updated_at
)
SELECT
    id, title, summary, ai_summary, content, author_id, toc,
    cover, category_id, comment_id, short_url, is_published,
    created_at, updated_at, deleted_at,
    is_top, is_hot, is_original, content_hash, ext_info,
    activitypub_object_id, activitypub_last_published_at, content_updated_at
FROM article;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
SELECT article_id, views, likes, comments, updated_at FROM article_metrics
ON CONFLICT (moment_id) DO NOTHING;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO moment_topic (moment_id, tag_id)
SELECT article_id, tag_id FROM article_tag
ON CONFLICT (moment_id, tag_id) DO NOTHING;
-- +goose StatementEnd

-- 显式写入 id 后需要把 identity 序列推过当前最大值
-- +goose StatementBegin
DO $$
DECLARE
    v_max BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) INTO v_max FROM moment;
    IF v_max > 0 THEN
        PERFORM setval(pg_get_serial_sequence('moment', 'id'), v_max, true);
    END IF;
END $$;
-- +goose StatementEnd

-- ---------------------------------------------------------------------------
-- C3. 联邦表引用从 article 改指 moment
-- ---------------------------------------------------------------------------
ALTER TABLE federated_citation RENAME COLUMN target_article_id TO target_moment_id;
ALTER TABLE federated_citation DROP CONSTRAINT IF EXISTS fk_federated_citation_article;
ALTER TABLE federated_citation
    ADD CONSTRAINT fk_federated_citation_moment FOREIGN KEY (target_moment_id) REFERENCES moment (id);
DROP INDEX IF EXISTS idx_federated_citation_target_status;
CREATE INDEX IF NOT EXISTS idx_federated_citation_target_status
    ON federated_citation (target_moment_id, status);

ALTER TABLE federation_outbound_delivery RENAME COLUMN source_article_id TO source_moment_id;
ALTER TABLE federation_outbound_delivery DROP CONSTRAINT IF EXISTS fk_federation_outbound_article;
ALTER TABLE federation_outbound_delivery
    ADD CONSTRAINT fk_federation_outbound_moment FOREIGN KEY (source_moment_id) REFERENCES moment (id) ON DELETE SET NULL;

-- ---------------------------------------------------------------------------
-- C4. 多态类型收窄：'article' → 'moment'
--     content_like / comment_area / analytics 三处均无 CHECK 约束，
--     类型合法性只由 Go 侧枚举保证，故此处只需搬数据。
-- ---------------------------------------------------------------------------
UPDATE content_like SET target_type = 'moment' WHERE target_type = 'article';
UPDATE comment_area SET area_type = 'moment' WHERE area_type = 'article';
UPDATE analytics_content_hourly SET content_type = 'moment' WHERE content_type = 'article';
UPDATE analytics_visitor_view SET content_type = 'moment' WHERE content_type = 'article';

-- ---------------------------------------------------------------------------
-- C5. 配置项随实体更名：发布类型只剩手记，热门阈值键改挂 moment
-- ---------------------------------------------------------------------------
UPDATE sys_config
SET value         = '["moment"]',
    default_value = '["moment"]',
    updated_at    = NOW()
WHERE config_key = 'activitypub.publishTypes';

UPDATE sys_config SET config_key = 'moment.hot.views'    WHERE config_key = 'article.hot.views';
UPDATE sys_config SET config_key = 'moment.hot.likes'    WHERE config_key = 'article.hot.likes';
UPDATE sys_config SET config_key = 'moment.hot.comments' WHERE config_key = 'article.hot.comments';

-- ---------------------------------------------------------------------------
-- C6. 导出记录去掉文章计数字段
-- ---------------------------------------------------------------------------
ALTER TABLE export_ops.export_record
    DROP COLUMN IF EXISTS article_count;

-- ---------------------------------------------------------------------------
-- C6.5 触发器函数去掉 article 分支
--      sync_content_like_metrics / adjust_comment_metrics_by_area 的函数体内
--      仍有 INSERT INTO article_metrics 分支。plpgsql 函数体不做依赖登记，
--      C7 删表时不会报错，但该分支一旦命中即抛 relation does not exist。
--      C4 已把存量多态类型全部收窄为 moment，Go 侧枚举也已去掉 article，
--      分支不可达；此处直接移除，避免留下指向已删表的死引用。
-- ---------------------------------------------------------------------------
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.sync_content_like_metrics()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.target_type = 'moment' THEN
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
        IF OLD.target_type = 'moment' THEN
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
$function$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.adjust_comment_metrics_by_area(p_area_id bigint, p_delta integer)
 RETURNS void
 LANGUAGE plpgsql
AS $function$
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

    IF v_area_type = 'moment' THEN
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
$function$;
-- +goose StatementEnd

-- ---------------------------------------------------------------------------
-- C7. 删表（引用方先删：article_metrics / article_tag → article → article_category）
--     idx_article_search_fts / idx_article_short_url / uq_article_* 随表删除。
-- ---------------------------------------------------------------------------
DROP TABLE IF EXISTS article_metrics;
DROP TABLE IF EXISTS article_tag;
DROP TABLE IF EXISTS article;
DROP TABLE IF EXISTS article_category;

-- ---------------------------------------------------------------------------
-- D 段（删除「导航菜单」）将在后续阶段作为独立迁移追加
-- ---------------------------------------------------------------------------


-- +goose Down

-- 本迁移不可逆：article 数据已并入 moment，无法拆分还原（合并后无法区分
-- 哪条 moment 原本是 article）。Down 仅恢复结构，数据不回填。
-- 同理，C4 的多态类型收窄无法回滚——回滚会把原本就是 moment 的记录也标成
-- article，故此处不动，仅以注释说明。

-- C6 / C5 回滚
ALTER TABLE export_ops.export_record
    ADD COLUMN IF NOT EXISTS article_count BIGINT NOT NULL DEFAULT 0;

UPDATE sys_config
SET value         = '["article","moment"]',
    default_value = '["article","moment"]',
    updated_at    = NOW()
WHERE config_key = 'activitypub.publishTypes';

UPDATE sys_config SET config_key = 'article.hot.views'    WHERE config_key = 'moment.hot.views';
UPDATE sys_config SET config_key = 'article.hot.likes'    WHERE config_key = 'moment.hot.likes';
UPDATE sys_config SET config_key = 'article.hot.comments' WHERE config_key = 'moment.hot.comments';

-- C3 回滚：联邦表引用改回 article
ALTER TABLE federated_citation DROP CONSTRAINT IF EXISTS fk_federated_citation_moment;
DROP INDEX IF EXISTS idx_federated_citation_target_status;
ALTER TABLE federated_citation RENAME COLUMN target_moment_id TO target_article_id;

ALTER TABLE federation_outbound_delivery DROP CONSTRAINT IF EXISTS fk_federation_outbound_moment;
ALTER TABLE federation_outbound_delivery RENAME COLUMN source_moment_id TO source_article_id;

-- C1 回滚：恢复 moment.img，去掉 cover
ALTER TABLE moment ADD COLUMN IF NOT EXISTS img TEXT;
ALTER TABLE moment DROP COLUMN IF EXISTS cover;

-- C7 回滚：恢复四张表结构（顺序与删除相反）
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS article_category
(
    id         BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    name       VARCHAR(45)  NOT NULL,
    short_url  VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT uq_article_category_name UNIQUE (name),
    CONSTRAINT uq_article_category_short_url UNIQUE (short_url)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS article
(
    id                            BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    title                         VARCHAR(255) NOT NULL,
    summary                       TEXT         NOT NULL,
    ai_summary                    TEXT,
    lead_in                       TEXT,
    toc                           JSONB        NOT NULL,
    content                       TEXT         NOT NULL,
    author_id                     BIGINT       NOT NULL,
    cover                         VARCHAR(255),
    category_id                   BIGINT,
    comment_id                    BIGINT,
    short_url                     VARCHAR(255) NOT NULL,
    is_published                  BOOLEAN      DEFAULT FALSE,
    created_at                    TIMESTAMPTZ  DEFAULT now(),
    updated_at                    TIMESTAMPTZ  DEFAULT now(),
    deleted_at                    TIMESTAMPTZ,
    is_top                        BOOLEAN      DEFAULT FALSE,
    is_hot                        BOOLEAN      DEFAULT FALSE,
    is_original                   BOOLEAN      DEFAULT TRUE,
    content_hash                  VARCHAR(32)  NOT NULL DEFAULT '',
    ext_info                      JSONB,
    activitypub_object_id         VARCHAR(500),
    activitypub_last_published_at TIMESTAMPTZ,
    content_updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT uq_article_short_url UNIQUE (short_url),
    FOREIGN KEY (author_id) REFERENCES app_user (id),
    FOREIGN KEY (category_id) REFERENCES article_category (id),
    FOREIGN KEY (comment_id) REFERENCES comment_area (id)
);
-- +goose StatementEnd

CREATE UNIQUE INDEX IF NOT EXISTS uq_article_activitypub_object_id
    ON article (activitypub_object_id) WHERE activitypub_object_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_article_published_created_at ON article (is_published, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_short_url ON article (short_url);
CREATE INDEX IF NOT EXISTS idx_article_search_fts ON article USING gin (((
    setweight(to_tsvector('simple'::regconfig, (COALESCE(title, ''::character varying))::text), 'A'::"char") ||
    setweight(to_tsvector('simple'::regconfig, COALESCE(summary, ''::text)), 'B'::"char")
) || setweight(to_tsvector('simple'::regconfig, COALESCE(content, ''::text)), 'C'::"char")));

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS article_metrics
(
    article_id BIGINT PRIMARY KEY
        REFERENCES article (id) ON DELETE CASCADE,
    views      BIGINT      NOT NULL DEFAULT 0,
    likes      INTEGER     NOT NULL DEFAULT 0,
    comments   INTEGER     NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS article_tag
(
    id         BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    article_id BIGINT NOT NULL REFERENCES article (id) ON DELETE CASCADE,
    tag_id     BIGINT NOT NULL REFERENCES tag (id) ON DELETE CASCADE,

    CONSTRAINT uq_article_tag UNIQUE (article_id, tag_id)
);
-- +goose StatementEnd

-- C6.5 回滚：恢复带 article 分支的函数定义
--     必须在 article_metrics 重建之后执行。
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.sync_content_like_metrics()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
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
$function$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.adjust_comment_metrics_by_area(p_area_id bigint, p_delta integer)
 RETURNS void
 LANGUAGE plpgsql
AS $function$
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
$function$;
-- +goose StatementEnd

-- C3 收尾：article 已恢复，把联邦表外键重新指回它
ALTER TABLE federated_citation
    ADD CONSTRAINT fk_federated_citation_article FOREIGN KEY (target_article_id) REFERENCES article (id);
CREATE INDEX IF NOT EXISTS idx_federated_citation_target_status
    ON federated_citation (target_article_id, status);

ALTER TABLE federation_outbound_delivery
    ADD CONSTRAINT fk_federation_outbound_article FOREIGN KEY (source_article_id) REFERENCES article (id) ON DELETE SET NULL;
