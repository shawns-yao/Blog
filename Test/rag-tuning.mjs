import assert from 'node:assert/strict';
import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const apiBase = process.env.RAG_TEST_API || 'http://127.0.0.1:18089/api/v2';
const target = new URL(apiBase);
assert(['127.0.0.1', 'localhost'].includes(target.hostname) && target.port === '18089', '只能在独立测试服务运行');
const account = `rag-tuning-${crypto.randomUUID().slice(0, 8)}`;
const password = crypto.randomUUID();
const sessionId = crypto.randomUUID();
const checks = [];
const previews = [];
const queries = [];
let token = '';
let initialTuning;
let lastAsk = 0;
let failure;

async function request(path, { method = 'GET', body, auth = true, allowError = false } = {}) {
  const response = await fetch(`${apiBase}${path}`, {
    method, headers: { 'Content-Type': 'application/json', ...(auth && token ? { Authorization: `Bearer ${token}` } : {}) },
    body: body === undefined ? undefined : JSON.stringify(body), signal: AbortSignal.timeout(95000)
  });
  const payload = await response.json();
  if (allowError) return payload;
  assert.equal(response.status, 200, `入口 HTTP ${response.status}`);
  assert.equal(payload.code, 0, payload.message);
  return payload.data;
}

async function check(name, action) {
  const item = { name, passed: false };
  checks.push(item);
  try { Object.assign(item, await action()); item.passed = true; }
  catch (error) { item.failure = error.message; throw error; }
  process.stdout.write(`${name}: 通过\n`);
}

const paragraph = '巡检时先观察设备温度、记录电量和现场环境，再核对最近一次记录。每次检查需要保存检查日期、设备编号和处理结果。';
const articleBody = `# 星舟项目运行手册

## 电池巡检
电池每 14 天巡检一次；电量低于 30% 时补充充电。日志保留 90 天。

${Array.from({ length: 24 }, (_, i) => `第 ${i + 1} 项记录：${paragraph}`).join('\n\n')}

## 备件
| 部件 | 数量 |
| --- | --- |
| 备用电池 | 2 |
| 温度探头 | 4 |

## 命令

~~~sh
echo inspection-ready
echo keep-log-90-days
~~~
`;
const noteBody = '# 星舟故障恢复\n\n温度报警时，先断开充电，等待 10 分钟，再检查温度。\n';

function measure(markdown, chunks) {
  const source = Array.from(markdown);
  const coverage = new Uint16Array(source.length);
  const lengths = [];
  for (const chunk of chunks) {
    assert.equal(source.slice(chunk.start, chunk.end).join(''), chunk.content, '偏移必须对应真实原文');
    lengths.push(Array.from(chunk.content).length);
    for (let i = chunk.start; i < chunk.end; i++) coverage[i]++;
  }
  let nonWhitespace = 0, covered = 0, duplicated = 0;
  source.forEach((character, i) => {
    if (/\s/u.test(character)) return;
    nonWhitespace++;
    if (coverage[i] > 0) covered++;
    if (coverage[i] > 1) duplicated += coverage[i] - 1;
  });
  return { chunks: chunks.length, minCharacters: Math.min(...lengths), maxCharacters: Math.max(...lengths),
    meanCharacters: lengths.reduce((a, b) => a + b, 0) / lengths.length,
    nonWhitespaceCharacters: nonWhitespace, coveredCharacters: covered,
    coverageRate: covered / nonWhitespace, duplicatedCharacters: duplicated, duplicateRate: duplicated / nonWhitespace };
}

async function waitReady(ids) {
  const until = Date.now() + 150000;
  while (Date.now() < until) {
    const list = await request('/admin/rag/documents?pageSize=100');
    const items = list.items.filter((item) => ids.includes(item.momentId));
    if (items.length === ids.length && items.every((item) => item.status === 'ready')) return;
    if (items.some((item) => item.status === 'failed')) throw new Error('独立语料索引失败，检查模型或环境');
    await new Promise((resolve) => setTimeout(resolve, 2000));
  }
  throw new Error('独立语料索引超时');
}

try {
  await check('真实注册登录与配置读取', async () => {
    await request('/auth/register', { method: 'POST', auth: false,
      body: { username: account, nickname: 'RAG 定向验证', email: `${account}@example.invalid`, password } });
    token = (await request('/auth/login', { method: 'POST', auth: false, body: { credential: account, password } })).token;
    assert(token);
    const settings = await request('/admin/rag/settings');
    assert(settings.enabled && settings.embeddingConfigured && settings.fallbackConfigured);
    initialTuning = settings.tuning;
  });
  await check('不同大小与重叠的真实分块预览', async () => {
    for (const [chunkSize, chunkOverlap] of [[600, 60], [1200, 120], [1800, 180], [1200, 0], [1200, 240]]) {
      const chunks = await request('/admin/rag/preview', { method: 'POST',
        body: { title: '星舟项目运行手册', markdown: articleBody, chunkSize, chunkOverlap } });
      const metrics = measure(articleBody, chunks);
      assert.equal(metrics.coverageRate, 1, '非空白原文不能遗漏');
      assert(chunks.some((chunk) => chunk.content.includes('echo inspection-ready\necho keep-log-90-days')));
      assert(chunks.some((chunk) => chunk.content.includes('| 温度探头 | 4 |')));
      previews.push({ chunkSize, chunkOverlap, ...metrics });
    }
    assert(new Set(previews.map((preview) => preview.chunks)).size > 1, '参数必须影响实际分块');
    assert.deepEqual((await request('/admin/rag/settings')).tuning, initialTuning, '预览不能保存配置');
    const invalid = await request('/admin/rag/preview', { method: 'POST', allowError: true,
      body: { title: '校验', markdown: noteBody, chunkSize: 600, chunkOverlap: 600 } });
    assert.notEqual(invalid.code, 0);
  });
  await check('可配置历史窗口与空索引交流', async () => {
    const status = await request('/public/rag/status', { auth: false });
    assert.equal(status.available, true);
    assert.equal(status.indexReady, false);
    assert.deepEqual(status.history, { maxRounds: 2, maxCharacters: 3000 });
    const history = Array.from({ length: 3 }, () => [{ role: 'user', content: '问题' }, { role: 'assistant', content: '回答' }]).flat();
    const rejected = await request('/public/ask', { method: 'POST', auth: false, body: { question: '你好', sessionId, history } });
    assert.equal(rejected.status, 'invalid_scope');
    lastAsk = Date.now();
    await new Promise((resolve) => setTimeout(resolve, 15500));
    const greeting = await request('/public/ask', { method: 'POST', auth: false, body: { question: '你好', sessionId } });
    lastAsk = Date.now();
    assert.equal(greeting.status, 'answered', greeting.reason);
    assert.equal(greeting.mode, 'conversation');
    assert.equal(greeting.trace.evidenceCount, 0);
  });
  let article, note;
  await check('独立语料通过项目入口入库', async () => {
    article = await request('/moments/', { method: 'POST', body: { title: '星舟项目运行手册', content: articleBody,
      summary: '电池巡检规范', shortUrl: 'rag-tuning-inspection', isPublished: true, extInfo: { contentKind: 'article' } } });
    note = await request('/moments/', { method: 'POST', body: { title: '星舟故障恢复', content: noteBody,
      shortUrl: 'rag-tuning-recovery', isPublished: true, extInfo: { contentKind: 'note' } } });
    await waitReady([article.id, note.id]);
    const stats = await request('/admin/rag/index');
    assert.equal(stats.ready, 2);
    return { readyDocuments: stats.ready, indexedChunks: stats.chunks };
  });
  const base = { ...initialTuning, vectorTopK: 8, keywordTopK: 8, topK: 4, rerankCandidateTopK: 16 };
  const profiles = [
    { name: '关键词权重', rrfVectorWeight: 0, rrfKeywordWeight: 1, rerankEnabled: false },
    { name: '向量权重', rrfVectorWeight: 1, rrfKeywordWeight: 0, rerankEnabled: false },
    { name: 'RRF 融合', rrfVectorWeight: 0.7, rrfKeywordWeight: 0.3, rerankEnabled: false },
    { name: 'RRF 加重排序', rrfVectorWeight: 0.7, rrfKeywordWeight: 0.3, rerankEnabled: true }
  ];
  const probes = [
    { question: '根据《星舟项目运行手册》，电池巡检周期和补充充电阈值是多少？', momentId: article.id, words: ['14', '30'] },
    { question: '《星舟故障恢复》中断开充电后要等待多久？', momentId: note.id, words: ['10'] },
    { question: '《星舟项目运行手册》里电池的保修期是多少年？', noEvidence: true }
  ];
  for (const profile of profiles) {
    const { name, ...overrides } = profile;
    await request('/admin/rag/settings', { method: 'PUT', body: { ...base, ...overrides } });
    for (const probe of probes) {
      const waitMs = 15500 - (Date.now() - lastAsk);
      if (waitMs > 0) await new Promise((resolve) => setTimeout(resolve, waitMs));
      const sample = { profile: name, question: probe.question, passed: false, externalFailure: null };
      queries.push(sample);
      const started = Date.now();
      lastAsk = started;
      try {
        const answer = await request('/public/ask', { method: 'POST', auth: false, body: { question: probe.question, sessionId } });
        Object.assign(sample, { status: answer.status, mode: answer.mode, answer: answer.answer, trace: answer.trace,
          citationMomentIds: answer.citations.map((citation) => citation.momentId) });
        if (answer.status === 'temporarily_unavailable') sample.externalFailure = 'model_or_environment';
        if (probe.noEvidence) assert.equal(answer.status, 'no_evidence', answer.reason);
        else {
          assert.equal(answer.status, 'answered', answer.reason);
          assert.equal(answer.mode, 'grounded');
          assert(answer.citations.some((citation) => citation.momentId === probe.momentId));
          for (const word of probe.words) assert(answer.answer.includes(word));
        }
        assert.equal(answer.trace.intent, 'knowledge_query');
        sample.passed = true;
      } catch (error) {
        sample.failure = error.message;
        if (!('status' in sample)) sample.externalFailure = 'network_or_environment';
      }
      sample.durationMs = Date.now() - started;
      process.stdout.write(`${name}: ${sample.passed ? '通过' : '失败'}\n`);
    }
  }
} catch (error) { failure = error.message; }
finally {
  if (initialTuning) {
    try { await request('/admin/rag/settings', { method: 'PUT', body: initialTuning }); }
    catch { failure ||= '独立测试配置恢复失败'; }
  }
  const external = queries.filter((sample) => sample.externalFailure).length;
  const valid = queries.length;
  const report = {
    testType: '定向测试', entry: apiBase, checks, previews,
    sampleFunnel: { original: 12, excluded: 12 - valid, valid },
    externalFailures: { total: external, rate: valid ? external / valid : 0 },
    coreFunction: { denominator: valid - external, passed: queries.filter((sample) => sample.passed).length,
      qualification: valid && external / valid > 0.1 ? '受外部因素影响，仅供参考' : null },
    queries, failure: failure ?? null,
    limitation: '合成语料的定向参数检查；预览没有重建索引，各检索阶段使用同一内容和同一索引。单路权重不关闭另一通道的计算，耗时不是独立单路性能。不能据此选定最优块长、权重或重排序阈值。未进行公开权威数据测试。'
  };
  await writeFile(fileURLToPath(new URL('./rag-tuning-results_20260929.json', import.meta.url)), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
  process.stdout.write(JSON.stringify(report, null, 2));
  if (failure || checks.some((item) => !item.passed) || queries.some((item) => !item.passed)) process.exitCode = 1;
}
