import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const entry = 'http://127.0.0.1:18091/api/v2';
const database = process.env.RAG_COMPARISON_DATABASE || 'rag_test_embedding20260929';
assert(/^rag_test_[a-z0-9]+$/.test(database), '仅允许独立测试库');
assert(process.env.HYBGZS_QWEN_API_KEY, '候选模型需要本地密钥');
const account = `rag-embedding-${crypto.randomUUID().slice(0, 8)}`;
const password = crypto.randomUUID();
const sessionId = crypto.randomUUID();
const redisPrefix = `rag-embedding-test:${crypto.randomUUID()}:`;
const records = [];
let server;
let token = '';
let lastAsk = 0;

const documents = [
  { title: 'Go 服务目录规范', shortUrl: 'embedding-go', content: '# 目录职责\n\nHTTP 入口放在 cmd/api；业务流程放在 internal/app；数据库适配放在 internal/infra。业务流程不能依赖 HTTP 请求对象。\n' },
  { title: 'Java 日期处理记录', shortUrl: 'embedding-java', content: '# 时间转换\n\n跨时区时间统一用 Instant 保存。展示给用户时用 ZoneId 转换为本地时间，不能保存没有时区的字符串作为唯一时间。\n' },
  { title: '星舟项目运行手册', shortUrl: 'embedding-inspection', content: '# 电池巡检\n\n电池每 14 天巡检一次；电量低于 30% 时补充充电。日志保留 90 天。\n' },
];
const questions = [
  { question: '根据本站《Go 服务目录规范》，HTTP 入口、业务流程和数据库适配分别放在哪些目录？', document: 0, words: ['cmd/api', 'internal/app', 'internal/infra'] },
  { question: '根据本站《Java 日期处理记录》，跨时区时间保存用哪个类型？显示时用哪个对象转换？', document: 1, words: ['Instant', 'ZoneId'] },
  { question: '根据本站《星舟项目运行手册》，设备低电量需要补充能量时，电量阈值与检查间隔是什么？', document: 2, words: ['30', '14'] },
  { question: '根据本站《Go 服务目录规范》，作者使用的生产服务器有多少台？', noEvidence: true },
];
const profiles = [
  { name: 'BGE 基线', embedding: process.env.RAG_EMBEDDING_MODEL || 'BAAI/bge-m3', reranker: process.env.RAG_RERANK_MODEL || 'Pro/BAAI/bge-reranker-v2-m3' },
  { name: 'Qwen 向量，BGE 重排序', embedding: 'Qwen/Qwen3-Embedding-8B', reranker: process.env.RAG_RERANK_MODEL || 'Pro/BAAI/bge-reranker-v2-m3', qwenEmbedding: true },
  { name: 'BGE 向量，Qwen 重排序', embedding: process.env.RAG_EMBEDDING_MODEL || 'BAAI/bge-m3', reranker: 'Qwen/Qwen3-Reranker-8B', qwenReranker: true },
];

async function request(path, { method = 'GET', body, auth = true } = {}) {
  const response = await fetch(`${entry}${path}`, {
    method, headers: { 'Content-Type': 'application/json', ...(auth && token ? { Authorization: `Bearer ${token}` } : {}) },
    body: body === undefined ? undefined : JSON.stringify(body), signal: AbortSignal.timeout(95000),
  });
  const payload = await response.json();
  assert.equal(response.status, 200, `项目入口 HTTP ${response.status}`);
  assert.equal(payload.code, 0, '项目入口返回错误');
  return payload.data;
}

async function stop() {
  if (!server || server.exitCode !== null) return;
  const exited = new Promise((resolve) => server.once('exit', resolve));
  server.kill();
  await exited;
  server = undefined;
}

async function start(profile) {
  await stop();
  const env = { ...process.env, APP_ENV: 'test', APP_PORT: '18091', DB_DRIVER: 'postgres',
    DB_DSN: `postgres://postgres@127.0.0.1:15439/${database}?sslmode=disable`,
    REDIS_ADDR: '127.0.0.1:16389', REDIS_PASSWORD: '', REDIS_DB: '0', REDIS_PREFIX: redisPrefix,
    AUTH_SECRET: crypto.randomUUID() + crypto.randomUUID(), RAG_ENABLED: 'true',
    RAG_HISTORY_MAX_ROUNDS: '5', RAG_HISTORY_MAX_CHARACTERS: '20000',
    HTMLSNAPSHOT_BASE_URL: 'http://127.0.0.1:1', GEOIP_DB_URL: 'http://127.0.0.1:1', GEOIP_ASN_URL: 'http://127.0.0.1:1',
    APP_UPDATE_CHECK_ENABLED: 'false', RAG_EMBEDDING_DIMENSIONS: '',
  };
  if (profile.qwenEmbedding) Object.assign(env, { RAG_EMBEDDING_BASE_URL: 'https://ai.hybgzs.com/v1',
    RAG_EMBEDDING_MODEL: profile.embedding, RAG_EMBEDDING_API_KEY: process.env.HYBGZS_QWEN_API_KEY });
  if (profile.qwenReranker) Object.assign(env, { RAG_RERANK_BASE_URL: 'https://ai.hybgzs.com/v1',
    RAG_RERANK_MODEL: profile.reranker, RAG_RERANK_API_KEY: process.env.HYBGZS_QWEN_API_KEY });
  for (const key of Object.keys(env)) if (/^(MEDIA_R2_|S3_API_|LEOSTUDIO_R2_|Access_Key_ID$|Secret_Access_Key$)/.test(key)) delete env[key];
  server = spawn(fileURLToPath(new URL('../Temp/rag-test-api-next.exe', import.meta.url)), [], {
    cwd: fileURLToPath(new URL('../Temp/rag-runtime', import.meta.url)), env, windowsHide: true, stdio: 'ignore',
  });
  server.on('error', () => {});
  const until = Date.now() + 30000;
  while (Date.now() < until) {
    assert(server.exitCode === null, '独立 API 启动失败');
    try { await request('/public/rag/status', { auth: false }); return; } catch { /* Wait for the real API. */ }
    await new Promise((resolve) => setTimeout(resolve, 500));
  }
  throw new Error('独立 API 启动超时');
}

async function indexSnapshot(record) {
  const until = Date.now() + 240000;
  while (Date.now() < until) {
    const stats = await request('/admin/rag/index');
    const list = await request('/admin/rag/documents?pageSize=100');
    if (stats.ready === documents.length && list.items.every((item) => item.status === 'ready')) {
      record.index = { readyDocuments: stats.ready, chunks: stats.chunks, dimensions: stats.embeddingDimension,
        documentIndexDurationsMs: list.items.map((item) => item.indexDurationMs) };
      return;
    }
    if (list.items.some((item) => item.status === 'failed' && item.attempts >= 5)) throw new Error('候选模型索引失败');
    await new Promise((resolve) => setTimeout(resolve, 2000));
  }
  throw new Error('候选模型索引等待超时');
}

try {
  for (const [profileIndex, profile] of profiles.entries()) {
    const record = { profile: profile.name, embeddingModel: profile.embedding, rerankerModel: profile.reranker, queries: [] };
    records.push(record);
    await start(profile);
    token = '';
    if (profileIndex === 0) await request('/auth/register', { method: 'POST', auth: false,
      body: { username: account, nickname: '向量链路对比', email: `${account}@example.invalid`, password } });
    token = (await request('/auth/login', { method: 'POST', auth: false, body: { credential: account, password } })).token;
    const settings = await request('/admin/rag/settings');
    assert.equal(settings.embeddingModel, profile.embedding);
    assert.equal(settings.rerankModel, profile.reranker);
    if (profileIndex === 0) for (const doc of documents) await request('/moments/', { method: 'POST', body: {
      ...doc, isPublished: true, extInfo: { contentKind: 'article' },
    } });
    const indexStarted = Date.now();
    try { await indexSnapshot(record); record.indexWaitMs = Date.now() - indexStarted; }
    catch { record.indexFailure = 'model_or_environment'; continue; }
    for (const probe of questions) {
      const waitMs = 15500 - (Date.now() - lastAsk);
      if (waitMs > 0) await new Promise((resolve) => setTimeout(resolve, waitMs));
      const sample = { question: probe.question, passed: false, externalFailure: null };
      record.queries.push(sample);
      const started = Date.now();
      lastAsk = started;
      try {
        const answer = await request('/public/ask', { method: 'POST', auth: false, body: { question: probe.question, sessionId } });
        Object.assign(sample, { status: answer.status, answer: answer.answer, trace: answer.trace,
          citationTitles: answer.citations.map((citation) => citation.title) });
        if (answer.status === 'temporarily_unavailable' || answer.trace?.embeddingDegraded) sample.externalFailure = 'model_or_environment';
        assert.equal(answer.trace?.intent, 'knowledge_query');
        if (probe.noEvidence) assert.equal(answer.status, 'no_evidence');
        else {
          assert.equal(answer.status, 'answered');
          assert.equal(answer.mode, 'grounded');
          assert(sample.citationTitles.includes(documents[probe.document].title));
          for (const word of probe.words) assert(answer.answer.includes(word));
        }
        sample.passed = true;
      } catch (error) {
        sample.failure = error.message;
        if (!('status' in sample)) sample.externalFailure = 'network_or_environment';
      }
      sample.durationMs = Date.now() - started;
      process.stdout.write(`${profile.name}: ${sample.passed ? '回答检查通过' : '回答检查未通过'}${sample.externalFailure ? '，存在外部失败' : ''}\n`);
    }
  }
} finally {
  await stop();
  for (const record of records) {
    const valid = record.queries.length;
    const external = record.queries.filter((sample) => sample.externalFailure).length;
    record.sampleFunnel = { original: questions.length, excluded: questions.length - valid, valid };
    record.externalFailures = { total: external, rate: valid ? external / valid : 0, indexing: record.indexFailure ? 1 : 0 };
    record.coreFunction = { denominator: valid - external,
      passed: record.queries.filter((sample) => !sample.externalFailure && sample.passed).length,
      qualification: external / valid > 0.1 ? '受外部因素影响，仅供参考' : null };
    const durations = record.queries.map((sample) => sample.durationMs).sort((a, b) => a - b);
    record.endToEndLatency = durations.length ? { meanMs: durations.reduce((a, b) => a + b, 0) / durations.length,
      minMs: durations[0], maxMs: durations.at(-1), includesFallbacks: true } : null;
  }
  const report = { testType: '定向测试', entry,
    corpus: { origin: '本项目合成的 Go、Java 与数值记录', version: '20260929', documents: documents.length, questions: questions.length },
    profiles: records,
    limitation: '通过项目注册、内容、索引、配置和问答入口运行。同一原文、分块与检索参数，每次更换向量模型后重建独立索引；日常配置不变。沿用 15 秒查询向量预算，嵌入降级单独列为外部失败，不计入向量效果指标。总延迟含意图识别、生成、重排序与降级；模型改写会变化，不是纯向量性能。小样本不能证明模型优劣、最优阈值或公开基准效果。',
  };
  await writeFile(fileURLToPath(new URL('./rag-embedding-comparison-results_20260929.json', import.meta.url)), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
  process.stdout.write(JSON.stringify(report, null, 2));
}
