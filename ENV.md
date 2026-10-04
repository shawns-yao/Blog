# 环境配置

实际环境文件集中放在项目根目录，按用途分组，不在 `server`、`web`、`admin` 或 `deploy` 目录保存环境文件。

| 文件 | 用途 | 提交到公开仓库 |
| --- | --- | --- |
| `.env` | 数据库、认证、前后台、备案、音乐、存储及构建配置 | 否 |
| `.env.rag` | 聊天模型、向量、重排序及供应商密钥 | 否 |
| `.env.example` | 通用配置模板，不包含真实密钥 | 是 |
| `.env.rag.example` | RAG 配置模板，不包含真实密钥 | 是 |

每个文件使用注释和横线区分模块。首次本地配置从对应模板复制，填写实际值。现有环境迁移时保留原值，不能用模板或电脑配置覆盖服务器配置。

## 读取方式

- Go 从项目根目录加载 `.env` 和 `.env.rag`；已设置的进程环境变量优先。
- 前后台从根目录读取 `.env`，只将对应框架允许的 `PUBLIC_`／`VITE_` 变量暴露给浏览器。不得用这些前缀命名密钥。
- Compose 从根目录运行，明确传入 `--env-file .env`。生产 RAG 覆盖文件只向后端注入 `.env.rag`，并暂停依赖个人电脑的 GPT 通道。
- 两个私有文件均受版本管理忽略规则保护，并排除在 Docker 构建上下文之外。

网站账号存储在数据库 `app_user` 表中，密码保存为哈希，管理员使用 `is_admin` 标识；后台根据当前数据库标识校验权限。环境文件中的数据库、Navidrome 账号用于服务之间连接。

`AUTH_SECRET` 用于签名和校验 JWT 登录令牌，保留原服务器值。更换它会使已有 JWT 登录令牌失效。

## 现有服务器配置迁移

在服务器 `~/shawns-blog` 执行，将原有通用配置移动到根目录；根目录已存在时保留它：

```bash
cd ~/shawns-blog
if [ ! -f .env ]; then
  mv -- deploy/.env .env
fi
chmod 600 .env
nano .env.rag
chmod 600 .env.rag
```

将电脑根目录 `.env.rag` 中的公网聊天、供应商密钥、向量与重排序配置复制到服务器 `.env.rag`。供应商变量放在引用 `${变量名}` 的配置之前。个人电脑的 GPT 地址和密钥无需复制，服务器 `RAG_EVALUATION_TRACE_DIR` 留空。不要将电脑 `.env` 覆盖到服务器。

只检查配置，不启动容器：

```bash
docker compose --env-file .env \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.music.yml \
  -f deploy/docker-compose.rag.yml \
  config --quiet
```

不要公开私有文件内容或完整的 Compose 配置输出。首次切换部署目录时，先完成镜像构建和原有存储目录迁移，再创建应用容器；环境文件整理不迁移账号、文章、音乐或数据库卷。RAG 启动和验收见 [RAG 生产部署](deploy/RAG.md)。
