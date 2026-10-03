# 服务端分享卡片字体

`web.Dockerfile` 将本目录的字体与许可证复制到镜像，不在构建时下载 GitHub 发布压缩包。保留原来的字体家族、字重和斜体，网页字体仍由现有前端依赖提供。

## 来源与许可证

- Noto Serif SC：上游 [Serif2.003](https://github.com/notofonts/noto-cjk/tree/Serif2.003/Serif/SubsetOTF/SC)，保留 Regular 和 Bold 原始 OTF 文件；许可证见 `NotoSerifSC-LICENSE.txt`，版权信息保留在字体元数据中。文件已与该版本的 Git blob 标识核对。
- Google Sans Code：上游 [v6.001](https://github.com/googlefonts/googlesans-code/releases/tag/v6.001)，从发布包提取 `variable/` 中的两个原始 TTF 文件；许可证见 `GoogleSansCode-LICENSE.txt`。
- 两组字体均使用 SIL Open Font License 1.1，未转换、改名或修改字体内部数据。

## SHA-256

| 文件                              | SHA-256                                                            |
| --------------------------------- | ------------------------------------------------------------------ |
| `NotoSerifSC-Regular.otf`         | `e8f396decc1f0963a016a989c3d8852e863d1350996f573860a80767c83a1cd3` |
| `NotoSerifSC-Bold.otf`            | `24693d48bdb9152f0a06b02af625638a1097abd6de4010ebba027f6e82710527` |
| `GoogleSansCode[wght].ttf`        | `c9649573afcd966f61096b1e9c3c3115e7b54315ca2adf9839c6f8b17d5f8e1f` |
| `GoogleSansCode-Italic[wght].ttf` | `bdf116292f27aca16e2aefb5534042a25f244d3a9187a521ce97fc2117a4a844` |

## 定向构建检查

2026-10-03 已从项目 Dockerfile 构建 `og-fonts` 阶段，并在生成的 Alpine 镜像中通过 `fc-scan` 确认 Noto Serif SC Regular／Bold 和 Google Sans Code 可读取。完整 `renderer` 镜像、分享卡片实际渲染和服务器网络仍需按部署流程验证。
