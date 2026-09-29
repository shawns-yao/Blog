import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { parseEnv } from 'node:util';

// Supplier protocol checks are prerequisites, not real-project acceptance scores.
const env = parseEnv(await readFile('server/.env', 'utf8'));
function value(key, seen = new Set()) {
  assert(!seen.has(key), '环境变量引用不能循环');
  const next = new Set(seen); next.add(key);
  return (env[key] ?? '').replace(/\$\{([A-Z0-9_]+)\}/g, (_, name) => value(name, next));
}
let input = [
  'Instruct: Given a question, retrieve passages that answer the question.\nQuery: Go 的 goroutine 和 Java 的线程有什么区别？',
  'Goroutines are lightweight units of execution scheduled by the Go runtime.',
  'Java 线程可以执行并发任务；虚拟线程和平台线程采用不同的运行机制。',
  '```go\nfunc main() { go work(); select {} }\n```\n这段代码启动一个 goroutine。',
  'Qwen embedding 主备需要使用相同版本、维度及输入处理。',
  'Markdown 标题、表格和公式需要保持原文结构和引用位置。',
];
let query = 'Go 的 goroutine 和 Java 的线程有什么区别？';
if (process.env.RAG_PROVIDER_PREFLIGHT_CAPTURE) {
  const rows = (await readFile(process.env.RAG_PROVIDER_PREFLIGHT_CAPTURE, 'utf8')).trim().split('\n').map(JSON.parse);
  const row = rows.find(item => item.capture?.stages?.fused?.lists?.[0]?.length);
  query = row.capture.answer?.trace?.query || row.question;
  input = ['query', ...row.capture.stages.fused.lists[0].map(item => item.contextHeader + '\n\n' + item.content)];
}
if (process.env.RAG_PROVIDER_PREFLIGHT_BATCH) {
  const batch = Number(process.env.RAG_PROVIDER_PREFLIGHT_BATCH);
  const seed = input;
  input = Array.from({length: batch}, (_, i) => seed[i % seed.length] + (process.env.RAG_PROVIDER_PREFLIGHT_LONG === '1' ? ' retrieval'.repeat(600) : ''));
}
const routes = [
  { name: 'hybgzs', url: 'https://ai.hybgzs.com/v1', key: value('HYBGZS_QWEN_API_KEY') },
  { name: 'tumuer', url: 'https://router.tumuer.me/v1', key: value('TUMUER_RAG_API_KEY') || value('RAG_EMBEDDING_FALLBACK_API_KEY') || value('RAG_EMBEDDING_API_KEY') },
];
const report = { testKind: '定向测试', scope: '供应商协议与向量兼容性核验，非项目功能验收',
  model: 'Qwen/Qwen3-Embedding-8B', dimensions: Number(process.env.RAG_PROVIDER_PREFLIGHT_EXPECTED_DIMENSIONS || 4096), inputCount: input.length,
  inputHashes: input.map(text => createHash('sha256').update(text).digest('hex')), providers: [] };
const vectors = [];
const requestedDimensions = Number(process.env.RAG_PROVIDER_PREFLIGHT_DIMENSIONS ?? 4096);
report.requestedDimensions = requestedDimensions;
report.createdAt = new Date().toISOString();
for (const route of routes) {
  assert(route.key, '本地环境必须配置供应商密钥');
  const result = { provider: route.name, embedding: {}, reranker: {} };
  for (const kind of process.env.RAG_PROVIDER_PREFLIGHT_RERANK_ONLY === '1' ? ['reranker']
    : process.env.RAG_PROVIDER_PREFLIGHT_EMBEDDING_ONLY === '1' ? ['embedding'] : ['embedding', 'reranker']) {
    const started = performance.now();
    try {
      const body = kind === 'embedding' ? { model: route.name === 'tumuer' ? (process.env.RAG_PROVIDER_PREFLIGHT_TUMUER_MODEL || report.model) : report.model, input, encoding_format: 'float' }
        : { model: 'Qwen/Qwen3-Reranker-8B', query,
          documents: input.slice(1), top_n: input.length - 1, return_documents: false };
      if (kind === 'embedding' && requestedDimensions) body.dimensions = requestedDimensions;
      result[kind].requestedModel = body.model;
      const headers = { 'Content-Type': 'application/json', Authorization: 'Bearer ' + route.key };
      if (process.env.RAG_PROVIDER_PREFLIGHT_USER_AGENT) headers['User-Agent'] = process.env.RAG_PROVIDER_PREFLIGHT_USER_AGENT;
      const response = await fetch(route.url + (kind === 'embedding' ? '/embeddings' : '/rerank'), {
        method: 'POST', headers,
        body: JSON.stringify(body), signal: AbortSignal.timeout(90000), redirect: 'error',
      });
      result[kind].httpStatus = response.status;
      result[kind].contentType = response.headers.get('content-type');
      if (!response.ok) {
        const raw = await response.text();
        result[kind].responseBytes = Buffer.byteLength(raw);
        result[kind].responseHash = createHash('sha256').update(raw).digest('hex');
        result[kind].cloudflarePage = /cloudflare|cf-ray/i.test(raw);
      }
      assert(response.ok, '供应商请求未成功');
      const payload = await response.json();
      if (kind === 'embedding') {
        result[kind].returnedModel = payload.model;
        result[kind].returnedCount = payload.data?.length;
        result[kind].returnedDimensions = payload.data?.map(item => item.embedding?.length);
        result[kind].returnedIndexes = payload.data?.map(item => item.index);
        assert.equal(payload.data.length, input.length);
        const items = [...payload.data].sort((a,b) => a.index-b.index);
        assert(items.every((item,i) => item.index === i && item.embedding.length === report.dimensions
          && item.embedding.every(Number.isFinite)), '向量协议或维度不正确');
        const normalized = items.map(item => {
          const norm = Math.hypot(...item.embedding); assert(norm > 0);
          return item.embedding.map(component => component / norm);
        });
        vectors.push({ provider: route.name, items: normalized });
        result[kind].dimensions = report.dimensions;
      } else {
        assert.equal(payload.results.length, input.length - 1);
        assert.equal(new Set(payload.results.map(item => item.index)).size, input.length - 1);
        assert(payload.results.every(item => item.index >= 0 && item.index < input.length - 1
          && Number.isFinite(item.relevance_score)), '重排协议不正确');
        result[kind].ranking = payload.results.map(item => item.index);
        result[kind].scores = payload.results.map(item => item.relevance_score);
      }
      result[kind].success = true;
    } catch (error) {
      result[kind].success = false;
      result[kind].error = error.name;
    }
    result[kind].durationMs = Math.round(performance.now() - started);
    console.log(JSON.stringify({ provider: route.name, kind, ...result[kind] }));
  }
  report.providers.push(result);
}
if (vectors.length === 2) {
  report.sameInputCosines = vectors[0].items.map((a,i) => a.reduce((sum,component,j) => sum + component*vectors[1].items[i][j],0));
  report.minimumSameInputCosine = Math.min(...report.sameInputCosines);
  report.maximumComponentDelta = Math.max(...vectors[0].items.map((a,i) => Math.max(...a.map((component,j) => Math.abs(component-vectors[1].items[i][j])))));
  // This screen catches protocol/space mismatches; it does not prove every input is equivalent.
  report.compatibilityScreenPassed = report.minimumSameInputCosine >= 0.999;
} else report.compatibilityScreenPassed = process.env.RAG_PROVIDER_PREFLIGHT_RERANK_ONLY === '1' ? null : false;
await mkdir('Temp', { recursive: true });
const artifact = 'Temp/rag-provider-preflight_' + report.createdAt.replace(/[:.]/g, '-') + '.json';
await writeFile(artifact, JSON.stringify(report, null, 2) + '\n');
console.log(JSON.stringify({ compatibilityScreenPassed: report.compatibilityScreenPassed,
  minimumSameInputCosine: report.minimumSameInputCosine, maximumComponentDelta: report.maximumComponentDelta, artifact }));
if (process.env.RAG_PROVIDER_PREFLIGHT_RERANK_ONLY === '1'
  ? report.providers.some(provider => !provider.reranker.success) : !report.compatibilityScreenPassed) process.exitCode = 1;
