# RAG 生产部署

RAG 复用现有 Go 服务、PostgreSQL 和 Redis。生产配置通过 `docker-compose.rag.yml` 将根目录私有 `.env.rag` 注入后端容器；启用此覆盖文件时开启 RAG，并暂停依赖个人电脑的 GPT 通道。其余公网通道沿用环境文件中的开关、模型和密钥。

## 准备私有配置

从本地根目录 `.env.rag` 复制公网聊天、向量和重排序配置到服务器根目录 `.env.rag`。供应商变量 `HYBGZS_QWEN_API_KEY`、`TUMUER_RAG_API_KEY` 应放在使用 `${变量名}` 的配置之前。GPT 通道的密钥与地址无需复制；服务器 `RAG_EVALUATION_TRACE_DIR` 应留空。

也可以从根目录 `.env.rag.example` 创建文件，再填写真实模型密钥。数据库、Redis、认证和音乐配置使用服务器原有值，原 `deploy/.env` 移到根目录 `.env`，不从电脑覆盖。文件位置和迁移步骤见 [环境配置](../ENV.md)。

```bash
cd ~/shawns-blog
nano .env.rag
chmod 600 .env.rag

docker compose --env-file .env \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.music.yml \
  -f deploy/docker-compose.rag.yml \
  config --quiet
```

`.env.rag` 受现有忽略规则保护，并从 Docker 构建上下文排除。它只用于后端运行阶段，不传给前台、管理后台或构建参数。不要把其内容或完整的 `docker compose config` 输出粘贴到公开记录。

## 启动与更新

首次部署从已有目录切换源码时，先完成镜像构建和存储目录迁移，再用三个覆盖文件启动。RAG 没有独立应用容器，也不需要启动本地评测环境。后续仅调整 RAG 密钥或模型时，无需重建镜像，重建后端容器即可：

```bash
cd ~/shawns-blog
docker compose --env-file .env \
  -f deploy/docker-compose.yml \
  -f deploy/docker-compose.music.yml \
  -f deploy/docker-compose.rag.yml \
  up -d --no-deps --force-recreate --no-build server
```

此命令执行现有后端启动入口；启动入口会检查项目数据库迁移。应用前应确认使用目标服务器的配置与数据库，并具备服务启动授权。后续 `up` 命令继续携带 RAG 覆盖文件，避免重新创建后端时遗漏模型配置。

## 服务器验收

先调用 `/api/v2/public/rag/status` 确认模型配置可用，再通过后台发布少量文档，等待实际工作进程生成索引，最后从书灵问答入口验证回答和引用。空库的“没有资料”回答不能证明文档检索成功。

首次上线执行冒烟测试：状态接口、少量真实文档的索引、问答及引用。公开权威数据测试脚本保留在 `Test/benchmark/`；它们需要独立评测环境，不能直接对线上业务数据库运行。服务器尚未执行的测试不能描述为通过。
