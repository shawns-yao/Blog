import assert from 'node:assert/strict';
import { readFile, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

// 定向测试 uses the eight imported articles; --smoke is a separate 冒烟测试.
const smoke = process.argv.includes('--smoke');
const token = process.env.RAG_VERIFY_ADMIN_TOKEN || '';
const api = 'http://127.0.0.1:8080/api/v2';
const root = new URL('../', import.meta.url);
const report = {
  testType: smoke ? '冒烟测试' : '定向测试',
  runStatus: 'running',
  source: smoke ? '日常项目入口；不使用八篇草稿作为检索证据' : '八篇文档入库系列 Markdown；日常项目入口',
  metricsPurpose: smoke ? '主链路可运行检查，不代表检索效果或问答准确率' : '本站内容定向检查，不属于公开权威评测',
  startedAt: new Date().toISOString(),
  samples: [],
};

async function request(path, options = {}) {
  const response = await fetch(api + path, {
    method: options.body ? 'POST' : 'GET',
    headers: { 'Content-Type': 'application/json', ...(options.auth ? { Authorization: `Bearer ${token}` } : {}) },
    body: options.body ? JSON.stringify(options.body) : undefined,
    signal: AbortSignal.timeout(125000),
  });
  if (response.status >= 500 || response.status === 408 || response.status === 429) {
    throw new Error('项目入口暂时不可用');
  }
  assert.equal(response.status, 200, `项目入口 HTTP ${response.status}`);
  const payload = await response.json();
  assert.equal(payload.code, 0, '项目入口返回错误');
  return payload.data;
}

function measureChunks(markdown, chunks) {
  const source = Array.from(markdown);
  const covered = new Uint16Array(source.length);
  for (const chunk of chunks) {
    assert.equal(source.slice(chunk.start, chunk.end).join(''), chunk.content, '子块必须对应原文范围');
    assert(chunk.parentStart <= chunk.start && chunk.parentEnd >= chunk.end, '父块必须包含子块');
    assert(chunk.headingPath.length && chunk.parentId && chunk.chunkId, '结构元数据完整');
    for (let i = chunk.start; i < chunk.end; i++) covered[i]++;
  }
  let characters = 0, missing = 0, duplicate = 0;
  source.forEach((ch, i) => {
    if (/\s/u.test(ch)) return;
    characters++;
    if (!covered[i]) missing++;
    duplicate += Math.max(0, covered[i] - 1);
  });
  assert.equal(missing, 0, '原文不能漏段');
  return { chunks: chunks.length, coverage: 1, duplicateRate: duplicate / characters,
    minTokens: Math.min(...chunks.map((c) => c.tokens)), maxTokens: Math.max(...chunks.map((c) => c.tokens)) };
}

const factual = [
  { name: '检索必要性', question: 'RAG 为什么先检索文档，再让大模型回答？', ids: [9] },
  { name: 'PDF 内容判断', question: '怎样判断 PDF 是可搜索的文字页面还是扫描页面？', ids: [10] },
  { name: '图片与截图', question: '有正文的页面，里面的图片和截图应该怎样处理？', ids: [11] },
  { name: '表格与公式', question: '识别表格和公式时怎样保留结构与符号？', ids: [12] },
  { name: '双栏阅读顺序', question: '双栏页面、文字乱序和图片顺序怎么处理？', ids: [13] },
  { name: '页眉页脚', question: '文档的页眉页脚怎么处理？', ids: [14] },
  { name: '重复内容', question: '重复内容应该在文档处理的什么阶段处理？', ids: [15] },
  { name: '结构分块', question: '文档最后怎样分块，标题和段落应该如何保留？', ids: [16] },
  { name: '两类页面比较', question: '比较可搜索文字页面和扫描页面，两者的解析方式有什么区别？', ids: [10], strategy: 'COMPARE' },
  { name: '连续处理步骤', question: '处理双栏页面的阅读顺序后，应该怎样清理页眉页脚，再进行分块？', ids: [13, 14, 16], strategy: 'MULTI_HOP', requireAll: true },
];

let settings;
let lastAsk = 0;
async function runSample(sample, history = []) {
  // The real public endpoint allows six requests per minute.
  const wait = Math.max(0, 11000 - (Date.now() - lastAsk));
  if (wait) await new Promise((resolve) => setTimeout(resolve, wait));
  lastAsk = Date.now();
  const started = performance.now();
  const result = { name: sample.name, externalFailure: false, passed: false };
  report.samples.push(result);
  try {
    const answer = await request('/public/ask', { body: { question: sample.question, sessionId: crypto.randomUUID(), history } });
    result.durationMs = Math.round(performance.now() - started);
    result.status = answer.status;
    result.reason = answer.reason;
    result.trace = answer.trace;
    result.sourceIds = answer.citations.map((c) => c.momentId);
    if (answer.status === 'temporarily_unavailable') {
      result.externalFailure = true;
      return null;
    }
    assert(sample.allowNoEvidence ? ['answered', 'no_evidence'].includes(answer.status) : answer.status === 'answered', '没有返回有效结果');
    if (sample.mode) assert.equal(answer.mode, sample.mode);
    if (sample.strategy) assert.equal(answer.trace.strategy, sample.strategy);
    if (sample.ids) {
      const matches = sample.ids.map((id) => result.sourceIds.includes(id));
      assert(sample.requireAll ? matches.every(Boolean) : matches.some(Boolean), '答案没有引用正确文章');
      assert.equal(answer.mode, 'grounded');
    }
    if (settings) {
      assert(answer.trace.contextTokens <= settings.tuning.contextMaxTokens, '证据预算超限');
      assert(answer.trace.historyTokens <= settings.tuning.historyMaxTokens, '历史预算超限');
      assert(answer.trace.queries.includes(sample.question), '原问题未参与召回');
    }
    if (sample.gpt) assert.equal(answer.trace.answerProvider, 'gpt', 'GPT 未成功生成；降级不算 GPT 可用');
    if (answer.status === 'answered') assert(answer.answer.length > 0);
    result.passed = true;
    return answer;
  } catch (error) {
    result.failure = error.name;
    if (error.name !== 'AssertionError') result.externalFailure = true;
    return null;
  } finally {
    result.durationMs ??= Math.round(performance.now() - started);
    process.stdout.write(`${sample.name}: ${result.externalFailure ? '外部失败' : result.passed ? '通过' : '未通过'}\n`);
  }
}

try {
  const status = await request('/public/rag/status');
  assert(status.available, '问答服务未启用');
  if (smoke) {
    await runSample({ name: 'GPT 问候', question: '你好', mode: 'conversation', gpt: true });
    await runSample({ name: '无历史指代澄清', question: '它的下一步怎么做？', mode: 'conversation' });
    await runSample({ name: '检索链路可运行', question: '请依据站内文章说明 RAG 为什么先检索再回答。', allowNoEvidence: true, strategy: 'FACT', gpt: true });
  } else {
    assert(token, '请在本地配置现有 RAG_VERIFY_ADMIN_TOKEN，测试不会创建账号或令牌');
    settings = await request('/admin/rag/settings', { auth: true });
    assert.equal(settings.chatChannels.find((c) => c.configured && !c.default)?.name, 'gpt', '测试要求 GPT 为首选');
    const manifest = JSON.parse(await readFile(new URL('Data/manifest/rag-document-ingestion-import_20260929.json', root), 'utf8'));
    const documents = await request('/admin/rag/documents?page=1&pageSize=100', { auth: true });
    for (const source of manifest.documents) {
      const doc = documents.items.find((d) => d.momentId === source.momentId);
      assert(doc?.published && doc.status === 'ready', '八篇文章必须已获准发布并完成当前索引；测试不会自动发布');
    }
    report.chunkChecks = [];
    for (const source of manifest.documents) {
      const markdown = await readFile(new URL(`Data/processed/rag-document-ingestion-20260929/${source.file}`, root), 'utf8');
      const chunks = await request('/admin/rag/preview', { auth: true, body: { title: source.title, markdown } });
      const measured = measureChunks(markdown, chunks);
      assert(measured.maxTokens <= settings.tuning.chunkMaxTokens, '子块超过上限');
      report.chunkChecks.push({ documentId: source.momentId, ...measured });
    }
    for (const sample of factual) await runSample(sample);
    const first = await runSample({ name: '追问前置事实', question: '文档的页眉页脚怎么处理？', ids: [14] });
    if (first) await runSample({ name: '指代追问', question: '那它应该在分块前还是分块后处理？', ids: [14, 16], strategy: 'FOLLOW_UP' },
      [{ role: 'user', content: '文档的页眉页脚怎么处理？' }, { role: 'assistant', content: first.answer }]);
    await runSample({ name: '切换明确主题', question: '怎样识别表格和公式？', ids: [12] }, first ?
      [{ role: 'user', content: '文档的页眉页脚怎么处理？' }, { role: 'assistant', content: first.answer }] : []);
  }
  report.runStatus = 'completed';
} catch (error) {
  report.runStatus = 'blocked';
  report.setupFailure = error.message.startsWith('请在本地') || error.message.startsWith('八篇文章') ? error.message : error.name;
  process.exitCode = 1;
} finally {
  const externalFailures = report.samples.filter((s) => s.externalFailure).length;
  const core = report.samples.filter((s) => !s.externalFailure);
  report.funnel = { original: report.samples.length, excluded: 0, effective: report.samples.length };
  report.external = { failed: externalFailures, rate: report.samples.length ? externalFailures / report.samples.length : null };
  report.core = { returned: core.length, passed: core.filter((s) => s.passed).length,
    passRate: core.length ? core.filter((s) => s.passed).length / core.length : null,
    limitation: externalFailures > report.samples.length * 0.1 ? '受外部因素影响，仅供参考' : null };
  report.completedAt = new Date().toISOString();
  const name = smoke ? 'rag-md-smoke-results_20260929.json' : 'rag-md-pipeline-results_20260929.json';
  await writeFile(fileURLToPath(new URL('Test/' + name, root)), JSON.stringify(report, null, 2) + '\n');
  if (report.samples.some((s) => !s.passed)) process.exitCode = 1;
  process.stdout.write(`${report.testType}: ${report.runStatus}，核心返回 ${core.length}，外部失败 ${externalFailures}\n`);
}
