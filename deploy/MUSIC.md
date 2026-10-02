# 音乐模块部署

前台入口为 `/music`，后台入口为 `/admin/music`。前台使用现有 Svelte 页面，后台使用现有 Vue 页面，Go 通过 Subsonic API 接入 Navidrome。博客身份认证继续使用现有用户系统，无须新增博客数据库表或迁移。

## 当前功能

- 后台批量上传 MP3、FLAC、M4A，显示传输进度、校验结果及重复文件。
- Go 使用 `ffprobe` 校验音频编码，在专用临时目录保存完整文件，再原子移动到曲库目录。
- 批量上传后统一请求扫描；后台展示扫描状态和 Navidrome 实际返回的曲库。“已保存”仅表示文件保存成功，入库结果以曲库为准。
- 前台提供歌曲搜索、分页、专辑浏览、播放队列、播放与暂停、上一首与下一首、进度及音量控制。
- 音乐室使用白色背景、红色交互与固定底部播放器，左侧提供搜索、歌单、排行榜、我的收藏和设置；歌单、排行榜及收藏目前预留，设置支持页面内的密度、封面与音量调整。
- 歌曲默认私有。管理员在后台曲库逐首开启“公开试听”，游客可搜索和播放这些歌曲；公开专辑仅包含已公开的歌曲，封面及音频也按公开名单检查。
- 点击底部播放器的封面、歌名或歌词按钮，展开白底红色的歌词页面；共用原播放器，收起后继续播放。同步歌词逐行高亮、跟随滚动并支持点击跳转，TXT 纯文本歌词静态展示。
- 后台曲库逐首补传 JPG／PNG 封面和 UTF-8 LRC／TXT 歌词；优先使用补传资源，缺少补传歌词时读取 Navidrome 返回的嵌入或外部歌词。未识别的歌词标签会显示“暂无歌词”，可补传 LRC／TXT。
- 队列仅保存在当前音乐页面，离开页面停止播放。第一版不提供个人收藏、持久歌单、真正的视频 MV 或音频标签编辑。

## 目录与权限

```text
deploy/storage/music/library/   # 原始音乐，Go 写入，Navidrome 只读
deploy/storage/music/staging/   # 上传临时文件，不进入 Navidrome 曲库
deploy/storage/music/assets/    # 按歌曲编号保存补传封面与规范化歌词
deploy/storage/music/public-catalog.json # 公开试听名单和歌曲元数据快照
deploy/storage/navidrome/       # Navidrome 自身数据和转码缓存
```

临时目录和曲库目录必须在同一文件系统，且临时目录必须位于曲库目录外。Compose 给 Go 挂载整个 `storage/music`，给 Navidrome 只读挂载 `storage/music/library`，保证完整文件原子发布。音乐不经过公开的 `/uploads` 路径。

只允许管理员上传、请求扫描及调整公开试听。私有曲库继续使用现有权限：管理员默认可以播放；其他用户须将博客用户编号加入 `MUSIC_ALLOWED_USER_IDS`，留空时仅管理员可以访问私有音乐。这个名单配置在 Go 环境变量中，修改后重建 Go 容器。公开试听不要求登录，也不会授予私有曲库访问权限。

公开状态在曲库目录的父目录保存为 `public-catalog.json`，使用同目录临时文件原子替换；取消公开保留权限记录，不删除音乐。文件不存在时没有公开歌曲，损坏或无法读取时拒绝公开访问，禁止将整个曲库作为回退。公开元数据在开启开关时从 Navidrome 获取；外部修改音乐标签后重新开启开关以更新快照。当前按单个 Go 音乐服务写入该文件部署，多个独立 Go 实例不可共享写入。

补传资源保存为 `assets/<歌曲编号>/cover.jpg` 和 `lyrics.json`，属于需要保留的音乐附属资源，由 Go 单实例管理；同种资源再次上传会原子替换旧版本，没有版本历史，不修改原始音频。Navidrome 返回的歌曲路径是虚拟路径，因此不用于定位源文件。补传资源不进入公开 `/uploads`，歌词与封面都沿用歌曲公开名单和私有访问权限，补传操作本身不会公开歌曲。

封面单文件最多 10 MiB，限制为有效 JPG／PNG、长宽各不超过 4096 像素，规范化为最大 720 像素 JPEG。歌词单文件最多 1 MiB，仅接受 UTF-8，允许 BOM；LRC 支持 `[mm:ss.xxx]`、多时间标签及 `[offset:毫秒]`，按时间排序；TXT 无时间标签。补传后刷新音乐室获取新封面，重新展开歌词页面获取新歌词；没有重新上传音频或扫描的要求。

## 第一次部署

在服务器项目的 `deploy` 目录执行以下步骤。所有命令都需要由运维人员明确执行；本文不会自动运行命令。

1. 按现有部署方式准备基础 `.env`、数据库和 HTTPS 反向代理。将 `music.env.example` 的设置合并到本地 `.env`，暂时保留 `MUSIC_ENABLED=false`。真实用户名、密码不得提交到 Git。
2. 创建目录并设置容器用户权限：

   ```sh
   mkdir -p storage/music/library storage/music/staging storage/navidrome
   chown -R 10001:10001 storage/music storage/navidrome
   ```

3. 启动 Navidrome：

   ```sh
   docker compose -f docker-compose.yml -f docker-compose.music.yml up -d navidrome
   ```

4. 从自己的电脑建立 SSH 隧道，访问 `http://127.0.0.1:4533` 完成 Navidrome 的首次管理员初始化：

   ```sh
   ssh -L 4533:127.0.0.1:4533 your-user@your-server
   ```

   创建供 Go 使用的独立集成账号，授予曲库读取、转码和扫描权限。`startScan` 需要管理员权限，因此当前集成账号须为 Navidrome 管理员；前台用户仅经 Go 访问这里实现的接口。不要填写个人音乐客户端账号，不开启公开分享。Navidrome 端口只绑定服务器回环地址。

5. 在本地 `.env` 设置 `MUSIC_NAVIDROME_USER`、`MUSIC_NAVIDROME_PASSWORD`、`MUSIC_ALLOWED_USER_IDS`，再设置 `MUSIC_ENABLED=true`。
6. 构建并部署应用：

   ```sh
   docker compose -f docker-compose.yml -f docker-compose.music.yml up -d --build
   ```

   该命令沿用现有部署入口，其中包含博客数据库迁移；生产执行前按现有备份及迁移流程检查。音乐模块自身未新增 SQL 迁移。

生产音乐播放需要 HTTPS。私有播放使用同源、`HttpOnly`、`Secure`、`SameSite=Strict` Cookie，路径仅为 `/api/v2/music`；私有曲库及会话接口使用现有 Bearer 登录，私有音频和封面每次请求检查账号及权限。退出登录清理音乐 Cookie。游客使用 `/api/v2/public/music` 下的独立接口，服务端只允许公开名单中的歌曲及对应封面；返回 `no-store`，取消公开后新的请求立即被拒绝。

## 参数

| 参数                       | 默认值                  | 用途                                     |
| -------------------------- | ----------------------- | ---------------------------------------- |
| `MUSIC_ENABLED`            | `false`                 | 启用音乐模块                             |
| `MUSIC_NAVIDROME_URL`      | `http://navidrome:4533` | Go 可访问的 Navidrome 内网地址           |
| `MUSIC_NAVIDROME_USER`     | 空                      | 集成账号                                 |
| `MUSIC_NAVIDROME_PASSWORD` | 空                      | 集成账号密码，仅服务端使用               |
| `MUSIC_ALLOWED_USER_IDS`   | 空                      | 非管理员的博客用户编号，用英文逗号分隔   |
| `MUSIC_MAX_UPLOAD_BYTES`   | `104857600`             | 单文件 100 MiB                           |
| `MUSIC_MAX_BIT_RATE`       | `192`                   | MP3 转码请求的码率上限，范围 64–320 kbps |
| `MUSIC_LIBRARY_DIR`        | `storage/music/library` | Go 曲库目录                              |
| `MUSIC_STAGING_DIR`        | `storage/music/staging` | Go 上传临时目录                          |
| `MUSIC_REQUEST_TIMEOUT`    | `15s`                   | 元数据及扫描请求超时                     |
| `MUSIC_FFPROBE_BINARY`     | `ffprobe`               | 音频校验程序路径                         |

服务器 Docker 镜像包含 `ffprobe`；直接运行 Go 时，需要在本机安装 FFmpeg 或设置程序的绝对路径。原始文件保持不变，播放时向 Navidrome 请求 MP3 转码。Navidrome 需要保留可用的 MP3 转码配置，实际音质、缓存和拖动进度须通过真实歌曲验证。

首台 4 核 8 GB、2 Mbps 服务器先使用默认 192 kbps、小曲库和少量用户。未来增加 30 Mbps 服务器时，如果音频仍经首台 Go 代理，出口带宽仍受首台限制；后续需要将音乐请求入口一并部署到高带宽服务器。

## 验证与备份

启用后执行定向测试：管理员上传一首拥有使用授权的音乐，扫描后确认默认私有，允许账号可以搜索、播放及拖动进度，游客不能获取该歌曲、专辑、歌词或封面。开启公开试听后检查匿名搜索、专辑过滤、歌词、封面和 Range 音频请求，取消公开后再次确认访问被拒绝，重启检查公开状态和补传资源持久保存。再检查重复上传、伪造扩展名、超大文件、扫描失败和桌面／手机页面。歌词页面需实际验证展开／收起连续播放、跟随滚动、点击跳转、切换歌曲与键盘关闭。

2026-10-02 独立音乐环境的 46 项最终 HTTP 定向检查通过，覆盖补传、嵌入歌词、公开权限、音频一致性、Range 和重启；初次 MP3 合成歌词标签未被 Navidrome 识别，改用标准 FLAC 标签后通过，原始失败尝试保留在 `Test/music-lyrics_20261002_v1.json`。本地 Edge 扩展连接失败，歌词页面的桌面／手机实际布局和浏览器交互尚未验证；没有生产性能或并发结论。

现有博客备份未纳入音乐及 Navidrome 数据。请单独备份 `storage/music/library`、`storage/music/assets`、`storage/music/public-catalog.json` 和 `storage/navidrome`；备份 Navidrome 数据库时采用停机备份或其 SQLite 一致性备份方式。临时文件和可重新生成的转码缓存无须作为原始音乐保存。

官方依据：[Navidrome Docker 安装](https://www.navidrome.org/docs/installation/docker/)、[Navidrome Subsonic API](https://www.navidrome.org/docs/developers/subsonic-api/)、[OpenSubsonic API](https://opensubsonic.netlify.app/docs/endpoints/)、[Navidrome 0.64.2](https://github.com/navidrome/navidrome/releases/tag/v0.64.2)。

歌词与封面依据：[Navidrome 歌词来源配置](https://www.navidrome.org/docs/usage/configuration-options/)、[OpenSubsonic 歌曲歌词接口](https://opensubsonic.netlify.app/docs/endpoints/getlyricsbysongid/)、[结构化歌词与偏移](https://opensubsonic.netlify.app/docs/responses/structuredlyrics/)、[Navidrome 封面来源](https://www.navidrome.org/docs/usage/library/artwork/)。
