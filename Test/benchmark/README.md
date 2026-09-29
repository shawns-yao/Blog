# 公开数据集问答评测

本目录保存公开数据集的编排与评分工具。完整冻结子集运行标记为**公开权威数据测试**；仅选择原有失败样本复测时标记为**定向测试**，保存在独立的 `directed-runs/`，不得混入公开子集统计。数据、评判环境和运行结果保存在 `Temp/rag-benchmarks/`，不与项目单元测试混合。

## 数据集与首轮范围

| 集合 | 可追溯来源 | 首轮冻结范围 | 标准相关性单位 |
| --- | --- | --- | --- |
| Open RAG Benchmark | [项目](https://github.com/vectara/open-rag-bench)、[数据](https://huggingface.co/datasets/vectara/open_ragbench)；版本 `63f6b052ff83508b08e242db42263ee708815c26` | 24 道纯文本问题，抽取型和概括型各 12 道；对应的 24 篇完整论文及 24 篇官方干扰论文 | 原始 `doc_id:section_id` |
| BEIR SciFact | [项目](https://github.com/beir-cellar/beir)、[官方归档](https://public.ukp.informatik.tu-darmstadt.de/thakur/BEIR/datasets/scifact.zip)；归档 SHA-256 写入清单 | 32 个 test 查询，包含全部对应标准文档及补足到 100 篇的干扰文档 | 原始文档 ID |

按样本 ID 的 SHA-256 顺序确定首轮样本，先冻结，再运行。Open RAG 将官方解析 JSON 的文字 section 转成 Markdown，保留原始段落坐标；不测试 PDF 解析、OCR 或图像问答。SciFact 保留原始陈述，只添加“根据知识库资料核实并说明依据”的入口说明。标准答案和相关性标签仅供评分，不能写入项目语料或请求。

这些是缩小语料范围后的基线测量，不能与官方全量排行榜直接比较，也不能证明已选出最佳分块、重叠或 TopK 参数。BEIR 没有问答参考文本，不生成替代标准答案。

## 实际执行链路

经数据库准备授权，在日常数据库中批量保存带集合标识的临时手记；项目触发器与工作进程实际执行分块、嵌入和索引。问答仅调用 `/api/v2/public/ask`，继续使用项目配置的问题理解、改写、多查询、双路召回、RRF、重排序、证据预算及模型降级。脚本不调用内部 RAG 模块，不注入向量、候选或答案，不需要管理员账号。

两套集合必须串行执行。`contentKind=note` 隔离临时语料，启动前检查没有其他公开手记。结束时将该集合手记撤回为草稿，保留核心来源记录；撤回触发器清理派生分块。既有文章的正文指纹和发布状态必须与运行前一致，不清除运行聚合记录。

本地阶段记录需要同时配置服务端 `RAG_EVALUATION_TRACE_DIR=Temp/rag-evaluation` 并在请求中携带 `X-RAG-Evaluation: 1`；没有配置时默认关闭。记录实际候选顺序、上下文、配置指纹及返回值，只写本地文件，不改变公开响应。评测结束后移除本地开关并按已有启动方式重启后端。

## 运行

当前脚本面向本机 Windows 和既有 Docker 开发容器 `shawn-blog-server-dev`、`shawn-blog-postgres`。需要项目服务已启动且 RAG 已启用；启动服务和数据库准备须取得授权。密钥只读取忽略的 `server/.env`，不添加管理员环境变量。

```powershell
# 下载冻结版本并生成独立数据目录、Markdown 和来源清单。
node Test/benchmark/prepare-rag-datasets.mjs

# 一次执行一个集合，等待完成与撤回后再执行下一个。
node Test/benchmark/run-rag-benchmark.mjs beir-scifact
node Test/benchmark/run-rag-benchmark.mjs open-rag-bench --chunk-max=4000

# 修复验证：保留完整冻结语料，只运行指定的原始失败问题。
# 不传 --chunk-max 时使用项目日常配置；定向结果另存，公开基线不覆盖。
node Test/benchmark/run-rag-benchmark.mjs open-rag-bench --sample-ids=fe1c96e9-42de-45c9-9422-063b55c14ce5,e51bbbef-f955-4d23-86cd-6ff94e936279,55fd007f-2662-4393-8e90-ed8ca6b05a56

# 在该定向轮次进行中另开终端，观察真实工作进程的索引结果。
# 检查原文覆盖、位置、向量维度及普通／原子块上限，不替代项目分块。
Temp/rag-benchmarks/ragas-runtime/Scripts/python.exe Test/benchmark/inspect-rag-index.py open-rag-bench --wait

# 独立评分环境；完整依赖快照另存到 Temp。
python -m venv Temp/rag-benchmarks/ragas-runtime
Temp/rag-benchmarks/ragas-runtime/Scripts/python.exe -m pip install --upgrade pip
Temp/rag-benchmarks/ragas-runtime/Scripts/python.exe -m pip install ragas==0.4.3 instructor==1.17.0 langchain-community==0.3.31 python-dotenv
Temp/rag-benchmarks/ragas-runtime/Scripts/python.exe Test/benchmark/score-ragas.py beir-scifact
Temp/rag-benchmarks/ragas-runtime/Scripts/python.exe Test/benchmark/score-ragas.py open-rag-bench
Temp/rag-benchmarks/ragas-runtime/Scripts/python.exe Test/benchmark/score-ragas.py open-rag-bench --directed
```

Ragas 可以消费进行中的输出，也可以在项目测试完成后独立运行。评分使用 Gemini，避免默认用生成回答的 GPT 自评；回答相关性复用配置的 BGE 嵌入。Gemini 的实际请求至少间隔 12 秒，客户端和结构化适配器均不重试失败请求。评判模型与嵌入请求都独立于项目功能统计；多次评分分别存入运行目录的 `ragas/<UTC时间>/`，不覆盖此前失败记录。

Open RAG 首次入库在完整公式 1053 token 处超过日常 800 的硬上限；1600 轮次随后遇到 2919 token 的公式，两次记录分别保留。长论文轮次采用目标 500／硬上限 4000／重叠 60，父范围上限同步为 4000；这些调整发生在问答执行前，用于保证原文完整入库，不根据问答分数选择参数。4000 是当前项目允许的硬上限，不代表最佳块长。临时上限和父范围参数由脚本记录、恢复，并等待原有公开内容重新完成基线索引；这不是自动按内容类型选取参数。需要续评分时使用相同集合参数并添加 `--resume`，已有分数只在评分契约相同时复用。

## 指标契约与结果

每次运行先保存 `metadata.json`，另存 `source-map.json`、`index-snapshot.json`、逐样本 JSONL 和 CSV。来源清单包含官方版本、文件指纹、原始样本量、筛选规则及实际执行量；标准文档缺失或索引失败不能被静默排除。

按项目返回的真实排序，映射到原始文档／section，在首次出现的位置去重，测量 `Recall`、`HitRate`、`MRR`、二值 `nDCG`，截断为 1／3／6／10／20。向量与关键词指标使用首个查询的实际排序；融合、重排序及最终上下文使用项目实际合并顺序。短列表不补足标准文档，空列表和功能失败记零。`Recall@K` 的 K 是去重后的相关性单位，另外保留原始候选和上下文块数，二者不能混称。

Ragas 0.4.3 分别评判 `Faithfulness`、`AnswerAccuracy` 和 `AnswerRelevancy`。有据率使用实际发送给生成模型的全部上下文；无回答样本的准确性和相关性记零，有据率不适用。SciFact 没有问答标准答案，准确性不适用。准确性采用[官方带问题的双向评分](https://docs.ragas.io/en/stable/concepts/metrics/available_metrics/nvidia_metrics/)，保留原始标准答案，每个方向只请求一次。HTTP 异常与未定义值单独记录，不补造分数；每项列出实际评分分母，不能把有据率当作整体答对率。

Open RAG 的首个评分尝试使用 `FactualCorrectness`，发现它不读取问题，无法正确解释“Yes.”这类依赖问题的标准答案；旧评分与异常仍保留，最终准确性改用 `AnswerAccuracy`，另起评分目录，不重跑或修改项目问答。已完成的 SciFact 评分仅使用有据率和相关性，不受这项替换影响。

样本漏斗、外部失败、核心返回分别展示。网络、模型最终不可用和执行环境异常不混入核心功能分母；功能错误仍留在核心分母。外部失败超过 10% 时标注指标仅供参考。脚本每个样本只发送一次项目请求，不重试失败样本；项目自己的降级保持生效。首轮没有预设效果验收阈值，判定为 `INCONCLUSIVE`，不根据测量后得到的分数倒推通过线。

`evaluator-selftest.json` 的四项评分器检查属于独立**定向测试**，不混入公开数据集分数。临时原始数据、虚拟环境、逐样本内容及本地阶段记录不提交 Git。
