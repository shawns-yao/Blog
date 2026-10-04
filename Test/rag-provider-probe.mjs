import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const results = [];
const selectedNames = process.env.RAG_PROVIDER_CASES?.split(',');
let lastRequest = 0;
async function probe(name, model, path, key, body, inspect, native = false) {
  if (selectedNames && !selectedNames.includes(name)) return;
  const item = { name, model, protocol: native ? 'gemini' : 'openai', passed: false };
  results.push(item);
  if (!key) { item.failure = 'credential_not_configured'; return; }
  const waitMs = 6000 - (Date.now() - lastRequest);
  if (waitMs > 0) await new Promise((resolve) => setTimeout(resolve, waitMs));
  lastRequest = Date.now();
  try {
    const response = await fetch(`https://ai.hybgzs.com${path}`, {
      method: 'POST', redirect: 'manual', signal: AbortSignal.timeout(35000),
      headers: { 'Content-Type': 'application/json', 'User-Agent': 'shawns-blog-rag/1.0',
        ...(native ? { 'x-goog-api-key': key } : { Authorization: `Bearer ${key}` }) },
      body: JSON.stringify(body)
    });
    item.httpStatus = response.status;
    if (response.status !== 200) {
      item.failure = 'provider_http_error';
      try {
        const error = await response.json();
        const message = String(error.error?.message || error.msg || '');
        if (/no available channel|无可用渠道|无渠道/i.test(message)) item.failure = 'provider_channel_unavailable';
        else if (/quota|额度|余额/i.test(message)) item.failure = 'provider_quota_unavailable';
        else if (/unauthorized|invalid.*key|认证|鉴权/i.test(message)) item.failure = 'provider_authentication_error';
      } catch { /* Do not expose provider error bodies. */ }
      return;
    }
    const data = await response.json();
    Object.assign(item, inspect(data));
  } catch { item.failure = 'network_or_protocol_error'; }
  finally { item.durationMs = Date.now() - lastRequest; process.stdout.write(`${name}: ${item.passed ? '通过' : '未通过'}\n`); }
}

await probe('Grok 对话协议', 'grok-4.7', '/v1/chat/completions', process.env.RAG_CHAT_GROK_API_KEY,
  { model: 'grok-4.7', messages: [{ role: 'user', content: '只回答一个字：好' }], max_tokens: 64, temperature: 0, stream: false },
  (data) => ({ passed: typeof data.choices?.[0]?.message?.content === 'string' && data.choices[0].message.content.trim().length > 0,
    returnedModel: data.model }));
await probe('Gemini 原生对话协议', 'gemini-3.1-flash-lite-preview', '/gemini/v1beta/models/gemini-3.1-flash-lite-preview:generateContent',
  process.env.RAG_CHAT_GEMINI_API_KEY,
  { contents: [{ role: 'user', parts: [{ text: '只回答一个字：好' }] }], generationConfig: { temperature: 0, maxOutputTokens: 64 } },
  (data) => ({ passed: data.candidates?.[0]?.content?.parts?.some((part) => typeof part.text === 'string' && part.text.trim().length > 0) === true }), true);
await probe('Gemini OpenAI 兼容协议', 'gemini-3.1-flash-lite-preview', '/v1/chat/completions', process.env.RAG_CHAT_GEMINI_API_KEY,
  { model: 'gemini-3.1-flash-lite-preview', messages: [{ role: 'user', content: '只回答一个字：好' }], max_tokens: 64, temperature: 0, stream: false },
  (data) => ({ passed: typeof data.choices?.[0]?.message?.content === 'string' && data.choices[0].message.content.trim().length > 0,
    returnedModel: data.model }));
for (const dimensions of [undefined, 1024, 2048]) {
  await probe(`Qwen 向量${dimensions ? ` ${dimensions} 维` : '原生维度'}`, 'Qwen/Qwen3-Embedding-8B', '/v1/embeddings', process.env.HYBGZS_QWEN_API_KEY,
    { model: 'Qwen/Qwen3-Embedding-8B', input: ['Go 项目的目录与职责', 'Java 项目的结构'], ...(dimensions ? { dimensions } : {}) },
    (data) => {
      const vectors = data.data?.map((item) => item.embedding) ?? [];
      const actual = vectors[0]?.length ?? 0;
      return { passed: vectors.length === 2 && actual > 0 && vectors.every((vector) => vector.length === actual && vector.every(Number.isFinite)) &&
        (!dimensions || actual === dimensions), requestedDimensions: dimensions ?? null, actualDimensions: actual };
    });
}
await probe('Qwen 重排序协议', 'Qwen/Qwen3-Reranker-8B', '/v1/rerank', process.env.HYBGZS_QWEN_API_KEY,
  { model: 'Qwen/Qwen3-Reranker-8B', query: 'Go 项目的目录与职责',
    documents: ['Go 项目分为入口、业务与基础设施目录。', 'Java 的集合类包括列表和映射。'], top_n: 2, return_documents: false },
  (data) => ({ passed: Array.isArray(data.results) && data.results.length === 2 && data.results.every((item) =>
    Number.isInteger(item.index) && Number.isFinite(item.relevance_score)),
    rankedIndices: data.results?.map((item) => item.index), scores: data.results?.map((item) => item.relevance_score) }));

const report = { testType: '定向测试', scope: '外部供应商协议核验，不计入项目功能指标', results,
  limitation: '仅核验指定模型的真实请求格式、输出维度和一次调用可用性，不证明检索质量、价格、配额或长期可用性。未改动在用向量模型或重建索引。' };
await writeFile(fileURLToPath(new URL(process.env.RAG_PROVIDER_REPORT_FILE || './rag-provider-probe-results_20260929.json', import.meta.url)), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
process.stdout.write(JSON.stringify(report, null, 2));
if (results.some((item) => !item.passed)) process.exitCode = 1;
