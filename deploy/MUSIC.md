# 音乐模块部署

前台入口为 `/music`，后台入口为 `/admin/music`。前台使用现有 Svelte 页面，后台使用现有 Vue 页面，Go 通过 Subsonic API 接入 Navidrome。博客身份认证继续使用现有用户系统；个人收藏与歌单新增 `0079_add_music_library.sql` 迁移，使用博客 PostgreSQL 持久保存。2026-10-03 前台类型检查、生产构建、定向静态检查、Go 编译与认证／路由 vet 已通过，浏览器验证范围见下文。2026-10-02 已在本地开发库 `grtblog` 单独执行 0079，Goose 记录版本 79 已应用，并核对三张表、六个索引及八项约束。历史迁移 74、78 未执行，本次不改变其状态；该结果不代表其他环境已迁移。

## 当前功能

- 后台一次多选 MP3、FLAC、M4A，显示每首歌曲的传输进度、后台处理状态及重复文件；可以同时选择同名 UTF-8 LRC，自动关联歌词。
- 上传接口保存完整原文件和任务记录后返回；后台单任务处理器使用 `ffprobe` 校验编码，使用 `ffmpeg` 生成完整 MP3 播放副本，再原子发布到曲库并扫描。原文件保留在扫描目录之外，后续播放复用完整副本。
- 状态依次为等待处理、生成播放缓存、扫描入库、已入库；失败显示具体阶段，重新上传相同文件可重试。任务记录保存到磁盘，服务重启后继续尚未完成的任务。后台页面轮询任务状态，关闭页面不会停止服务器处理；“已入库”表示处理成功且扫描已经结束，曲库结果仍以 Navidrome 为准。
- 前台提供歌曲搜索、分页、专辑浏览、播放队列、播放与暂停、上一首与下一首、进度及音量控制。
- 音乐室使用白色背景、红色交互与固定底部播放器，左侧提供首页、搜索、我的歌单、我的收藏和设置；排行榜入口隐藏，未实现排行统计。输入关键词后按回车搜索，移除搜索和刷新操作按钮，保留失败时的重试。全部歌曲、收藏、歌单列表与歌单详情每页 10 条；请求数量、页码和翻页步长共用 `MUSIC_PAGE_SIZE`。
- 设置保存密度、封面显示、音量和播放模式；点击“我的账号”直接进入设置，显示账号、昵称、邮箱和账号操作。游客点击账号入口打开白底红色登录表单。普通账号可以使用密码登录，后台接口继续要求管理员权限。
- 首页和搜索通过独立浏览接口展示全部曲库元数据，包括未公开歌曲。歌曲默认私有；游客只能播放管理员开启“公开试听”的歌曲。已有公开曲库、公开专辑接口仍只返回公开内容，歌词、封面和音频继续独立鉴权。播放白名单留空时允许所有已登录用户访问受限歌曲，填写白名单时保留指定用户限制。
- 点击底部播放器的封面、歌名或歌词按钮，展开白底红色的歌词页面；共用原播放器，收起后继续播放。同步歌词逐行高亮、跟随滚动并支持点击跳转，TXT 纯文本歌词静态展示。
- 后台曲库逐首补传 JPG／PNG 封面和 UTF-8 LRC／TXT 歌词；优先使用补传资源，缺少补传歌词时读取 Navidrome 返回的嵌入或外部歌词。未识别的歌词标签会显示“暂无歌词”，可补传 LRC／TXT。
- 收藏和歌单按博客账号保存，需登录操作。歌单支持创建、改名、删除、歌曲增删、相邻歌曲上移／下移和整单播放；排序支持跨分页边界。每个账号最多 100 个歌单，每个歌单最多 500 首。收藏或加入歌单不会开放私有播放权限。
- 播放器支持顺序播放、单曲循环、列表循环和随机播放；随机模式的上一首回到本次播放历史。播放模式、音量、显示设置和队列保存在当前浏览器域名下，刷新后恢复队列与当前选择并保持暂停；点击播放后重新检查私有音乐权限和会话。退出或切换账号清空队列、音频和对应音乐查询缓存。
- 播放器挂在全局布局，同一标签页内通过页面导航离开音乐室后继续播放。博客首页使用透明背景和浅色文字，音乐室仍为白底红色；底部按钮至少 44 像素，播放按钮 56 像素。下一首右侧提供红心，支持收藏／取消收藏并同步个人收藏列表，游客点击时要求登录。
- 首页书架恢复六本书，原“搜索”位置改为“音乐室”，保留其余书本的位置、尺寸与颜色；书灵继续提供搜索和问答入口。整单播放过滤不可播放歌曲。没有真正的视频 MV、音频标签编辑、个人上传或公开共享歌单。
- 收藏成功提示三秒后自动消失，错误信息继续保留；播放器和歌词页的进度条点击区域高 24 像素。音频实际时长不可用时使用曲库时长，暂停或恢复可播放状态时清除残留缓冲状态。

## 后台新增用户

更新代码并重新构建后，在 `/admin/audience/users` 的“本站用户管理”右上角点击“新增用户”，填写账号和密码，昵称与邮箱可以留空。密码至少八个字符，默认创建启用的普通用户；需要管理员时开启“管理员”。账号保存到博客数据库，密码使用现有 bcrypt 方式保存，普通用户不能调用后台创建接口。

新用户直接使用博客账号登录音乐室。希望所有登录用户可播放受限歌曲时，将项目根目录 `.env` 的 `MUSIC_ALLOWED_USER_IDS` 留空；需要限制特定用户时填写数据库用户编号，并重新创建后端容器。无论哪种方式，游客都只能播放公开试听歌曲。

## 批量上传、歌词关联与去重

在后台“音乐管理”的“批量上传”中点击“批量选择音乐与歌词”，一次选择多首音频和歌词，再点击“开始批量上传”。传输按队列逐首进行，已经传完的歌曲在服务器独立处理，不等待整批传输结束。关闭页面会中断尚未传完的文件，已经接收的任务继续处理。

自动关联采用完整文件名匹配，例如 `如愿.mp3` 对应 `如愿.lrc`，`如愿（现场版）.flac` 对应 `如愿（现场版）.lrc`。歌词必须为 UTF-8、最多 1 MiB，LRC 需要时间标签；多个同名候选或没有匹配音频的歌词会提示重新选择，不猜测所属版本。音频转换成 MP3 后，歌词也使用播放副本的文件名保存，Navidrome 按 `.lrc,.txt,embedded` 顺序读取。已经入库的歌曲可以在曲库点击该歌曲的“歌词”按钮单独补传 LRC 或 TXT，此处按歌曲编号直接关联。

去重按原始音频完整内容的 SHA-256 指纹进行，相同文件改名再上传仍只保留一个任务和一份歌曲；正在处理的重复上传复用同一任务。不同内容、编码、音质或版本分别保留，不按歌名合并，不做相似音频识别。重复音频仍可补充同名 LRC。歌词关联失败会单独显示，并支持只重试歌词，不要求重传成功的音频；原有曲库不会批量删除或改名。

批量歌词保存在 `originals/<内容指纹>/lyrics.lrc`，随后原子发布到对应曲库目录；重新上传同一音频的配对歌词会替换该配对歌词。曲库中单独补传的 `assets/<歌曲编号>/lyrics.json` 仍优先于配对歌词，更新已补传的歌词时继续使用该歌曲的“歌词”按钮。

歌词读取规则依据：[Navidrome 歌词来源配置](https://www.navidrome.org/docs/usage/configuration/options/)。

## 域名入口

目标博客域名为 `https://seecode.top/`，音乐入口为 `https://music.seecode.top/`，共用应用和业务数据库。音乐子域名的根路径跳转到 `/music/`，前端资源与 `/api/` 请求保留在音乐域名下。外层 Nginx 示例见 [seecode.conf.example](nginx/seecode.conf.example)，不修改项目托管的 `nginx.conf`。

部署时为两个域名配置 DNS 和 HTTPS 证书，按实际证书路径修改示例，项目 `.env` 设置 `NGINX_PORT=8088` 后再接入外层代理。示例尚未在服务器执行或通过 `nginx -t`，不是已经上线的配置；服务器现有站点、端口和证书需在实际部署时核对。

音乐子域名的“返回博客”在新标签页打开博客，保留原标签页播放；博客首页书架在新标签页打开音乐子域名。本地入口仍使用站内 `/music/` 导航。跨域或整页刷新会重建文档，不能保留原音频播放；账号数据共用，但当前登录令牌、队列和设置按域名保存，音乐室首次需要单独登录。若使用 OAuth，还需配置音乐域名对应的回调地址。

### 首页 ICP 备案展示

在服务器项目根目录的 `.env` 中配置备案编号和点击地址，Compose 将这两项传给前台 `renderer`：

```dotenv
SITE_ICP_NUMBER=冀ICP备2026004228号-2
SITE_ICP_URL=https://beian.miit.gov.cn/
```

首页底部显示备案编号，点击在新标签页打开备案网站；未配置编号时不显示。编号通过服务端布局读取运行时环境，只向页面返回这两项公开信息，不需要数据库迁移。此次首次加入展示逻辑，需要更新代码并重建 `renderer` 镜像。以后修改备案编号或链接，重新创建前台容器即可；已有首页静态快照还需按站点刷新流程更新。

```sh
docker compose --env-file ../.env -f docker-compose.yml -f docker-compose.music.yml \
  up -d --no-deps --no-build --force-recreate renderer
```

2026-10-03 已完成 ICP 展示的定向测试：使用实际生产构建和前台首页 HTTP 入口，配置编号时显示、未配置时隐藏；1280 像素桌面和 390 像素窄屏都没有横向溢出，链接点击区域高 44 像素，键盘焦点边框可见。类型检查零错误，保留既有时钟警告；生产构建、格式和 Compose 配置检查通过。此次仅运行独立前台预览，没有连接后端或数据库；真实手机、备案网站实际打开以及服务器更新版尚未验证。

## 目录与权限

```text
deploy/storage/music/originals/ # 新上传的原始音乐，需要备份
deploy/storage/music/jobs/      # 上传处理状态与重启恢复记录
deploy/storage/music/library/   # 新音乐的完整播放副本与旧曲库，Go 写入，Navidrome 只读
deploy/storage/music/staging/   # 上传临时文件，不进入 Navidrome 曲库
deploy/storage/music/assets/    # 按歌曲编号保存补传封面与规范化歌词
deploy/storage/music/public-catalog.json # 公开试听名单和歌曲元数据快照
deploy/storage/navidrome/       # Navidrome 自身数据和转码缓存
```

临时目录和曲库目录必须在同一文件系统，且临时目录必须位于曲库目录外。Compose 给 Go 挂载整个 `storage/music`，给 Navidrome 只读挂载 `storage/music/library`，保证完整文件原子发布。音乐不经过公开的 `/uploads` 路径。

只允许管理员上传、请求扫描及调整公开试听。管理员默认可以播放受限歌曲；`MUSIC_ALLOWED_USER_IDS` 留空时其他已登录用户也可播放，填写时仅允许指定博客用户编号。该名单修改后重新创建后端容器即可。公开试听不要求登录。

原始音乐属于需要长期保存的业务文件；播放副本属于可由原文件重建的派生文件；任务记录用于处理状态及中断恢复。尚未处理完成的原文件和任务记录不能清理。旧曲库不会自动迁移、改名或重新编码，因此既有歌曲编号、收藏和公开名单保持有效。Navidrome 的转码缓存上限默认设为 1 GB，保存在其数据卷中；该缓存可重新生成。

公开状态在曲库目录的父目录保存为 `public-catalog.json`，使用同目录临时文件原子替换；取消公开保留权限记录，不删除音乐。文件不存在时没有公开歌曲，损坏或无法读取时拒绝公开访问，禁止将整个曲库作为回退。公开元数据在开启开关时从 Navidrome 获取；外部修改音乐标签后重新开启开关以更新快照。当前按单个 Go 音乐服务写入该文件部署，多个独立 Go 实例不可共享写入。

补传资源保存为 `assets/<歌曲编号>/cover.jpg` 和 `lyrics.json`，属于需要保留的音乐附属资源，由 Go 单实例管理；同种资源再次上传会原子替换旧版本，没有版本历史，不修改原始音频。Navidrome 返回的歌曲路径是虚拟路径，因此不用于定位源文件。补传资源不进入公开 `/uploads`，歌词与封面都沿用歌曲公开名单和私有访问权限，补传操作本身不会公开歌曲。

封面单文件最多 10 MiB，限制为有效 JPG／PNG、长宽各不超过 4096 像素，规范化为最大 720 像素 JPEG。歌词单文件最多 1 MiB，仅接受 UTF-8，允许 BOM；LRC 支持 `[mm:ss.xxx]`、多时间标签及 `[offset:毫秒]`，按时间排序；TXT 无时间标签。补传后刷新音乐室获取新封面，重新展开歌词页面获取新歌词；没有重新上传音频或扫描的要求。

## 第一次部署

在服务器项目的 `deploy` 目录执行以下步骤。所有命令都需要由运维人员明确执行；本文不会自动运行命令。

1. 在项目根目录准备 `.env`、数据库和 HTTPS 反向代理，模板为根目录 `.env.example`，暂时保留 `MUSIC_ENABLED=false`。已有服务器配置按 [环境配置](../ENV.md) 移动并保留原值。真实用户名、密码不得提交到 Git；下列命令仍从 `deploy` 目录执行。
2. 创建目录并设置容器用户权限：

   ```sh
   mkdir -p storage/music/library storage/music/staging storage/navidrome
   chown -R 10001:10001 storage/music storage/navidrome
   ```

3. 启动 Navidrome：

   ```sh
   docker compose --env-file ../.env -f docker-compose.yml -f docker-compose.music.yml up -d navidrome
   ```

4. 从自己的电脑建立 SSH 隧道，访问 `http://127.0.0.1:4533` 完成 Navidrome 的首次管理员初始化：

   ```sh
   ssh -L 4533:127.0.0.1:4533 your-user@your-server
   ```

   创建供 Go 使用的独立集成账号，授予曲库读取、转码和扫描权限。`startScan` 需要管理员权限，因此当前集成账号须为 Navidrome 管理员；前台用户仅经 Go 访问这里实现的接口。不要填写个人音乐客户端账号，不开启公开分享。Navidrome 端口只绑定服务器回环地址。

5. 在项目根目录 `.env` 设置 `MUSIC_NAVIDROME_USER`、`MUSIC_NAVIDROME_PASSWORD`、`MUSIC_ALLOWED_USER_IDS`，再设置 `MUSIC_ENABLED=true`。
6. 构建并部署应用：

   ```sh
   docker compose --env-file ../.env -f docker-compose.yml -f docker-compose.music.yml up -d --build
   ```

   该命令沿用现有部署入口，其中包含博客数据库迁移；生产执行前按现有备份及迁移流程检查。新增 `0079_add_music_library.sql` 创建 `music_favorite`、`music_playlist`、`music_playlist_song`，不修改原始音乐与公开名单。未执行迁移时，个人功能会明确报错，不使用浏览器存储兜底。应用回退优先保留三张新表；执行迁移 Down 会删除个人收藏与歌单，必须先备份并另行确认，不能作为默认回退操作。

生产音乐播放需要 HTTPS。私有播放使用同源、`HttpOnly`、`Secure`、`SameSite=Strict` Cookie，路径仅为 `/api/v2/music`；私有曲库及会话接口使用现有 Bearer 登录，私有音频和封面每次请求检查账号及权限。退出登录清理音乐 Cookie。游客通过 `/api/v2/public/music/browse` 及其专辑接口查看全部歌曲的安全元数据；公开音频、歌词及封面接口仍只允许公开名单中的资源，返回 `no-store`。取消公开后新的媒体请求立即被拒绝，但歌曲信息仍可浏览。`/api/v2/music/access` 仅需登录并返回私有访问能力；个人收藏和歌单接口也只需登录，不受私有播放白名单限制。

## 源码构建与代理

服务器使用当前源码构建镜像。服务端分享卡片所需的 OTF／TTF 已随 `deploy/fonts` 保存，来源、版本和许可证见 [字体说明](fonts/README.md)；`web.Dockerfile` 通过 `og-fonts` 阶段复制本地文件，不再使用两个 GitHub 发布包的远程 `ADD`。前端 npm 依赖和系统软件包仍需网络下载。2026-10-03 用户提供的服务器日志已确认原版前台镜像完整构建并导出；新增 ICP 展示后需重建前台，分享卡片实际渲染尚未验证。

使用 Windows 电脑上的 HTTP／SOCKS5 混合代理时，可以建立 SSH 远程端口转发，将服务器回环端口转发到电脑上的代理端口。隧道窗口需保持运行；这不是服务器上的永久代理，也不需要将代理端口开放到公网。

```sh
# 在运行代理的电脑上执行；替换 SSH 用户与服务器地址。
ssh -N -o ExitOnForwardFailure=yes -o ServerAliveInterval=30 \
  -R 127.0.0.1:10808:127.0.0.1:10808 your-user@your-server
```

在 Linux 服务器构建前后台时，额外加载可选代理配置。该配置同时覆盖 `server` 和 `renderer`，使 npm、Go 模块和系统软件包下载经过代理。`network: host` 仅用于构建，使构建容器能访问 SSH 回环转发；代理通过构建参数传入，没有写入运行镜像的 `ENV`。

```sh
BUILD_HTTP_PROXY=http://127.0.0.1:10808 docker compose --env-file ../.env \
  -f docker-compose.yml -f docker-compose.music.yml \
  -f docker-compose.build-proxy.yml build server renderer
```

该配置只代理构建容器中的依赖下载，不配置 Docker 守护进程的镜像拉取代理，也不改变应用运行网络。仅支持 SOCKS5 的代理不能直接按上述 HTTP 地址使用；需确认代理同时提供 HTTP 接口。2026-10-03 已验证电脑 10808 的 HTTP 与 SOCKS5 请求成功，并通过 Compose 解析；用户提供的服务器输出显示经 SSH 转发访问 npm 返回 `{}`，随后原版前台镜像已完整构建。后端访问 `proxy.golang.org` 直连超时，现已将构建代理扩展到后端，服务器重试仍待验证。依据：[Docker 构建代理参数](https://docs.docker.com/build/building/variables/#proxy-arguments)。

运行阶段安装字体依赖时，若官方 Alpine 软件源下载失败，可通过 `ALPINE_MIRROR` 切换为同版本镜像源。这个参数只替换 `/etc/apk/repositories` 中的镜像根地址，保留基础镜像原有的 Alpine 版本分支；定义位于前端构建及生产依赖安装之后，以便复用已完成的构建缓存。

```sh
BUILD_HTTP_PROXY=http://127.0.0.1:10808 docker compose --env-file ../.env \
  -f docker-compose.yml -f docker-compose.music.yml \
  -f docker-compose.build-proxy.yml build \
  --build-arg ALPINE_MIRROR=https://mirrors.tuna.tsinghua.edu.cn/alpine renderer
```

2026-10-03 已完成字体依赖的定向测试：在当前 `node:22-alpine`（Alpine 3.24.2）基础镜像中执行项目 Dockerfile 的软件源替换及安装指令，直接访问与通过电脑 10808 HTTP 代理两种情况下，六个依赖包均安装成功，`fontconfig` 2.17.1 可运行。电脑经代理访问清华镜像的 v3.24 main／community 索引均返回 200；完整服务器镜像和分享卡片实际渲染仍待验证。软件源配置依据：[清华 Alpine 镜像使用说明](https://mirrors.tuna.tsinghua.edu.cn/help/alpine/)。

## 参数

| 参数                       | 默认值                  | 用途                                     |
| -------------------------- | ----------------------- | ---------------------------------------- |
| `MUSIC_ENABLED`            | `false`                 | 启用音乐模块                             |
| `MUSIC_NAVIDROME_URL`      | `http://navidrome:4533` | Go 可访问的 Navidrome 内网地址           |
| `MUSIC_NAVIDROME_USER`     | 空                      | 集成账号                                 |
| `MUSIC_NAVIDROME_PASSWORD` | 空                      | 集成账号密码，仅服务端使用               |
| `MUSIC_ALLOWED_USER_IDS`   | 空                      | 留空允许所有已登录用户；填写时按用户编号限制，用英文逗号分隔 |
| `MUSIC_MAX_UPLOAD_BYTES`   | `104857600`             | 单文件 100 MiB                           |
| `MUSIC_MAX_BIT_RATE`       | `192`                   | MP3 转码请求的码率上限，范围 64–320 kbps |
| `MUSIC_LIBRARY_DIR`        | `storage/music/library` | Go 曲库目录                              |
| `MUSIC_STAGING_DIR`        | `storage/music/staging` | Go 上传临时目录                          |
| `MUSIC_REQUEST_TIMEOUT`    | `15s`                   | 元数据及扫描请求超时                     |
| `MUSIC_FFPROBE_BINARY`     | `ffprobe`               | 音频校验程序路径                         |
| `MUSIC_FFMPEG_BINARY`      | `ffmpeg`                | 后台生成播放副本的程序路径               |
| `MUSIC_TRANSCODING_CACHE_SIZE` | `1GB`                | Navidrome 旧歌曲的转码缓存上限            |

服务器 Docker 镜像包含 `ffprobe` 和 `ffmpeg`；直接运行 Go 时，需要安装 FFmpeg 或设置程序的绝对路径。新上传音乐提前生成 MP3 播放副本，旧歌曲继续沿用 Navidrome 转码和缓存。配置依据：[Navidrome 缓存参数](https://www.navidrome.org/docs/usage/configuration/options/)。完整播放副本可以减少实时转码等待，但音乐传输仍受服务器出口带宽影响。

首台 4 核 8 GB、2 Mbps 服务器先使用默认 192 kbps、小曲库和少量用户。未来增加 30 Mbps 服务器时，如果音频仍经首台 Go 代理，出口带宽仍受首台限制；后续需要将音乐请求入口一并部署到高带宽服务器。

## 验证与备份

2026-10-04 本次修改仅做编译与静态检查，不启动博客项目、不连接数据库，也不运行迁移。Go API 编译和涉及模块的 vet、前后台类型检查与生产构建通过；前台保留既有时钟警告，后台保留既有构建分包警告。本次上传处理、批量歌词关联、重复文件复用、进度条交互、缓存生成和后台创建用户尚未经过更新代码的真实入口验证，不能作为功能或性能验收结果。

部署后执行定向测试：一次选择两首不同版本的音频和各自同名 LRC，确认后台完成后分别入库，歌词对应正确；同一音频原名和改名重复上传，确认没有新增歌曲；重复音频补充歌词，确认歌词可读取。检查处理失败重试、关闭页面后继续处理、重启后恢复任务，再从后台创建普通用户并检查音乐登录、后台接口拒绝。通过实际播放器检查暂停后不再显示残留缓冲、点击和拖动进度有效、收藏提示三秒后消失；首次与重复播放等待时间需要在服务器实测，2 Mbps 出口带宽仍会影响传输。

启用后执行定向测试：管理员上传一首拥有使用授权的音乐，扫描后确认默认私有，游客能浏览歌曲和专辑元数据，但不能获取私有音频、歌词或封面。允许账号可播放及拖动进度，普通非白名单账号登录后仍可播放公开歌曲。开启公开试听后检查公开搜索、专辑过滤、歌词、封面和 Range 音频请求，取消公开后确认媒体访问被拒绝但歌曲仍可浏览；重启检查公开状态和补传资源持久保存。再检查重复上传、伪造扩展名、超大文件、扫描失败和桌面／手机页面。歌词页面需实际验证展开／收起连续播放、跟随滚动、点击跳转、切换歌曲与键盘关闭。

个人功能需使用两个不同账号验证收藏、歌单隔离以及越权 ID 请求拒绝；检查重复收藏、重复添加歌曲、排序冲突、删除歌单不影响原始音乐、退出／换号清空播放与缓存。歌曲被 Navidrome 移除时保留个人记录并标记失效，仍可取消收藏或移除歌单项；上游故障不伪装成歌曲失效。

2026-10-03 在本地浏览器与实际 HTTP 入口完成 26 项最终定向检查，使用现有业务数据库和实际 Go／Navidrome，覆盖播放模式、刷新恢复、站内连续播放、两个普通账号登录、退出与换号清理、管理员接口拒绝、红心收藏／刷新／取消、回车搜索及 10+1 歌单分页。临时创建的 11 个空歌单已通过正常接口删除，测试收藏已取消；两个指定账号保留。桌面和 390 像素页面已检查，记录见 [music-playback_20261003_v1.json](../Test/music-playback_20261003_v1.json)。没有生产 HTTPS、真实手机或并发结论；歌曲与收藏的样本不足 11 首，其跨页、歌单详情跨页及整单播放尚未完整验收。

已有单元与仓储测试本轮未运行，不将其计入上述真实入口结果。仓储测试只允许明确指向可丢弃测试库的 `MUSIC_TEST_DATABASE_URL`；用户现有业务数据库不得用于创建隔离 schema 的仓储测试。歌词滚动／跳转、歌单排序／重命名、越权请求等仍需另行验收。当前本地音乐预览复用业务数据库，原音乐专用 PostgreSQL 和 Redis 已停止，没有删除其历史数据；本轮没有运行迁移或连接线上数据库。

以下为此次首页、收藏与歌单改动之前的历史记录，不代表本轮验证结果：2026-10-02 独立音乐环境的 46 项最终 HTTP 定向检查通过，覆盖补传、嵌入歌词、公开权限、音频一致性、Range 和重启；初次 MP3 合成歌词标签未被 Navidrome 识别，改用标准 FLAC 标签后通过，原始失败尝试保留在 `Test/music-lyrics_20261002_v1.json`。本地 Edge 扩展连接失败，歌词页面的桌面／手机实际布局和浏览器交互尚未验证；没有生产性能或并发结论。

个人收藏与歌单属于博客 PostgreSQL 数据，应确认数据库备份覆盖三张新增表。现有博客备份未纳入音乐文件及 Navidrome 数据。请单独备份 `storage/music/originals`、`storage/music/jobs`、`storage/music/library`、`storage/music/assets`、`storage/music/public-catalog.json` 和 `storage/navidrome`；备份 Navidrome 数据库时采用停机备份或其 SQLite 一致性备份方式。临时文件和可重新生成的转码缓存无须作为原始音乐保存。

官方依据：[Navidrome Docker 安装](https://www.navidrome.org/docs/installation/docker/)、[Navidrome Subsonic API](https://www.navidrome.org/docs/developers/subsonic-api/)、[OpenSubsonic API](https://opensubsonic.netlify.app/docs/endpoints/)、[Navidrome 0.64.2](https://github.com/navidrome/navidrome/releases/tag/v0.64.2)。

歌词与封面依据：[Navidrome 歌词来源配置](https://www.navidrome.org/docs/usage/configuration-options/)、[OpenSubsonic 歌曲歌词接口](https://opensubsonic.netlify.app/docs/endpoints/getlyricsbysongid/)、[结构化歌词与偏移](https://opensubsonic.netlify.app/docs/responses/structuredlyrics/)、[Navidrome 封面来源](https://www.navidrome.org/docs/usage/library/artwork/)。
