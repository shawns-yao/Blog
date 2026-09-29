import { spawnSync } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { mkdir, writeFile } from 'node:fs/promises';

// 定向测试：运行配置准备经授权；问答只经过项目公开入口。
const root = new URL('../', import.meta.url);
const api = 'http://127.0.0.1:8080/api/v2';
const directory = new URL('Temp/rag-benchmarks/dynamic-topk/', root);
await mkdir(directory, { recursive: true });
const output = new URL(`${new Date().toISOString().replaceAll(':', '-')}.json`, directory);
const keys = ['dynamicTopKEnabled', 'dynamicTopKMin', 'dynamicTopKMax', 'topK', 'contextMaxTokens', 'rerankEnabled'];
const defaults = { dynamicTopKEnabled: true, dynamicTopKMin: 2, dynamicTopKMax: 12, topK: 6, contextMaxTokens: 6000, rerankEnabled: true };
const report = {
  testType: '定向测试', entry: '/api/v2/public/ask', startedAt: new Date().toISOString(),
  source: '日常数据库中已发布的文章，包括原有八篇文档',
  limitations: ['不是公开权威数据集成绩，不证明最优参数；外部模型失败单独统计',
    '脚本不调用内部模块，不注入候选、向量或答案；运行参数准备与问答执行分开'],
  samples: [], checks: {}, restoration: null,
};
const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));
const quote = text => "'" + text.replaceAll("'", "''") + "'";
function sql(query) {
  const r = spawnSync('docker', ['exec', '-i', 'shawn-blog-postgres', 'sh', '-lc',
    'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -X -v ON_ERROR_STOP=1 -A -t -q'],
  { input: query, encoding: 'utf8', timeout: 30000, windowsHide: true });
  if (r.error || r.status !== 0) throw new Error('database_execution_failed');
  return r.stdout.trim() ? JSON.parse(r.stdout.trim()) : null;
}
function configure(values) {
  const rows = Object.entries(values).map(([key, value]) => ({ key: `rag.${key}`,
    value: typeof value === 'string' ? value : JSON.stringify(value),
    type: key.endsWith('Enabled') ? 'bool' : 'number' }));
  sql(`INSERT INTO sys_config(config_key,value,is_sensitive,group_path,label,description,value_type,sort,meta)
    SELECT key,value,false,'rag',key,'RAG 运行参数',type,0,'{}'::jsonb
    FROM jsonb_to_recordset(${quote(JSON.stringify(rows))}::jsonb) AS p(key text,value text,type text)
    ON CONFLICT(config_key) DO UPDATE SET value=EXCLUDED.value,updated_at=now();`);
}
function sources() {
  return sql(`SELECT jsonb_agg(d ORDER BY d.id) FROM (
    SELECT m.id,m.content,m.content_hash,m.is_published,s.active_profile,s.status
    FROM moment m JOIN rag_index_state s ON s.moment_id=m.id
    WHERE m.is_published AND m.deleted_at IS NULL AND m.ext_info->>'contentKind'='article'
  ) d;`);
}
async function waitReady() {
  for (let i = 0; i < 45; i++) {
    try {
      const r = await fetch(`${api}/public/rag/status`, { signal: AbortSignal.timeout(3000) });
      const data = (await r.json()).data;
      if (r.ok && data?.available && data.indexReady) return;
    } catch {}
    await sleep(2000);
  }
  throw new Error('project_not_ready');
}
let lastAsk = 0;
let baseline;
const original = sql(`SELECT COALESCE(jsonb_object_agg(config_key,value),'{}'::jsonb)
  FROM sys_config WHERE config_key IN (${keys.map(k => quote(`rag.${k}`)).join(',')}) AND is_sensitive=false;`);
const restore = Object.fromEntries(keys.map(key => [key, original[`rag.${key}`] ?? defaults[key]]));
const initialSources = sources();
const stableSources = docs => docs.map(({ id, content_hash, is_published, active_profile, status }) =>
  ({ id, content_hash, is_published, active_profile, status }));
const sourceById = new Map(initialSources.map(source => [source.id, Array.from(source.content)]));
const fact = { name: '事实问答', question: '什么是 RAG？', strategy: 'FACT', ids: [9] };
const multi = { name: '三步骤覆盖', question: '恢复双栏阅读顺序后，如何清理页眉页脚，再按标题和完整段落分块？',
  strategy: 'MULTI_HOP', ids: [13, 14, 16] };
const scoreGapOnly = process.argv.includes('--score-gap-only');
const factOnly = process.argv.includes('--fact-only');
const samples = scoreGapOnly ? [
  { name: '固定清晰排名基线', question: 'RAG 为什么先检索文档，再让大模型回答？', strategy: 'FACT', ids: [9], settings: { dynamicTopKEnabled: false } },
  { name: '动态清晰排名', question: 'RAG 为什么先检索文档，再让大模型回答？', strategy: 'FACT', ids: [9] },
] : factOnly ? [
  { ...fact, name: '固定事实基线', settings: { dynamicTopKEnabled: false } },
  { ...fact, name: '动态事实问答' },
  multi,
] : [
  { ...fact, name: '固定事实基线', settings: { dynamicTopKEnabled: false } },
  { ...fact, name: '动态事实问答' },
  { name: '动态比较', question: '比较可搜索文字页面和扫描页面，两者的解析方式有什么区别？', strategy: 'COMPARE', ids: [10] },
  multi,
  { name: '动态概览', question: '请概览本站现有文档讨论的文档入库主题，概括内容解析、清理和分块的要点，只总结检索到的资料。', strategy: 'GLOBAL', ids: [9, 10, 13, 14, 15, 16], any: true },
  { ...multi, name: '动态最大值约束', settings: { dynamicTopKMax: 4 } },
  { ...multi, name: '小证据预算', settings: { contextMaxTokens: 800 } },
  { ...fact, name: '关闭重排的保守选择', settings: { rerankEnabled: false } },
  { name: '无依据拒答', question: '按照站内文档，2031 年 2 月北极分公司的设备采购限额精确是多少元？', noEvidence: true },
];
try {
  await waitReady();
  baseline = Object.fromEntries(keys.map(key => [key, JSON.parse(typeof restore[key] === 'string' ? restore[key] : JSON.stringify(restore[key]))]));
  for (const sample of samples) {
    const tuning = { ...baseline, ...defaults, ...sample.settings };
    configure(tuning);
    await sleep(Math.max(0, 11000 - (Date.now() - lastAsk)));
    lastAsk = Date.now();
    const result = { name: sample.name, settings: tuning, question: sample.question, sessionId: randomUUID(), checks: {}, passed: false, externalFailure: null, projectFailure: null };
    report.samples.push(result);
    const started = performance.now();
    try {
      const response = await fetch(`${api}/public/ask`, { method: 'POST', headers: { 'Content-Type': 'application/json', 'X-RAG-Evaluation': '1' },
        body: JSON.stringify({ question: sample.question, contentKind: 'article', sessionId: result.sessionId }),
        signal: AbortSignal.timeout(125000) });
      if (!response.ok) {
        result.projectFailure = `http_${response.status}`;
        throw new Error('project_request_failed');
      }
      const payload = await response.json();
      if (payload.code !== 0) {
        result.projectFailure = 'business_error';
        throw new Error('project_request_failed');
      }
      const answer = payload.data;
      Object.assign(result, { status: answer.status, mode: answer.mode, answer: answer.answer, reason: answer.reason,
        trace: answer.trace, citations: answer.citations });
      if (answer.status === 'temporarily_unavailable') {
        result.externalFailure = 'model_or_environment';
      } else {
        const t = answer.trace;
        result.checks.status = answer.status === (sample.noEvidence ? 'no_evidence' : 'answered');
        if (sample.strategy) result.checks.strategy = t.strategy === sample.strategy;
        result.checks.mode = sample.noEvidence || answer.mode === 'grounded';
        result.checks.dynamicMode = t.dynamicTopK === tuning.dynamicTopKEnabled;
        result.checks.boundedTarget = t.topKTarget <= (tuning.dynamicTopKEnabled ? tuning.dynamicTopKMax : tuning.topK);
        result.checks.boundedCount = t.evidenceCount <= t.topKTarget;
        result.checks.budget = t.contextTokens <= tuning.contextMaxTokens;
        result.checks.validCitations = answer.citations.every(c => sourceById.has(c.momentId) &&
          sourceById.get(c.momentId).slice(c.start, c.end).join('') === c.content);
        if (sample.ids) {
          const found = sample.ids.map(id => answer.citations.some(c => c.momentId === id));
          result.checks.expectedSources = sample.any ? found.some(Boolean) : found.every(Boolean);
          if (sample.name === '小证据预算') delete result.checks.expectedSources;
        }
        if (sample.name === '三步骤覆盖') result.checks.increasedTarget = t.topKTarget > tuning.topK;
        if (sample.name === '关闭重排的保守选择') result.checks.noScoreGap = !t.topKReason.includes('score_gap');
        result.passed = Object.values(result.checks).every(Boolean);
      }
    } catch (error) {
      if (!result.projectFailure) result.externalFailure = error.cause?.code ?? error.message;
    }
    result.durationMs = Math.round(performance.now() - started);
    console.log(`${result.name}: ${result.externalFailure ?? (result.passed ? 'PASS' : 'FAIL')} count=${result.trace?.evidenceCount ?? '-'} target=${result.trace?.topKTarget ?? '-'} ${result.trace?.topKReason ?? ''}`);
    await writeFile(output, JSON.stringify(report, null, 2));
  }
  const fixed = report.samples[0], dynamic = report.samples[1];
  report.observations = { factReduced: !fixed.externalFailure && !dynamic.externalFailure && dynamic.trace.evidenceCount < fixed.trace.evidenceCount };
  if (scoreGapOnly) report.checks.clearScoreGapReduced = report.observations.factReduced && dynamic.trace.topKReason.includes('score_gap');
  report.checks.dynamicCountsDiffer = new Set(report.samples.filter(s => s.trace?.dynamicTopK).map(s => s.trace.evidenceCount)).size > 1;
  if (scoreGapOnly) delete report.checks.dynamicCountsDiffer;
} finally {
  configure(restore);
  await waitReady();
  const finalSources = stableSources(sources());
  report.restoration = { articlesUnchanged: JSON.stringify(finalSources) === JSON.stringify(stableSources(initialSources)),
    defaultsForNewKeys: defaults.dynamicTopKEnabled, articleCount: finalSources.length };
  report.finishedAt = new Date().toISOString();
  const external = report.samples.filter(s => s.externalFailure).length;
  report.funnel = { original: samples.length, excluded: 0, effective: samples.length,
    attempted: report.samples.length, externalFailures: external, projectFailures: report.samples.filter(s => s.projectFailure).length,
    coreReturns: report.samples.length - external,
    corePassed: report.samples.filter(s => s.passed).length };
  report.qualification = external / samples.length > 0.1 ? '受外部因素影响，仅供参考' : '定向验证，不代表公开效果';
  await writeFile(output, JSON.stringify(report, null, 2));
  console.log(JSON.stringify({ output: output.pathname, funnel: report.funnel, checks: report.checks, restoration: report.restoration }));
}
if (report.samples.some(s => !s.externalFailure && !s.passed) || Object.values(report.checks).some(v => !v) || !report.restoration.articlesUnchanged) process.exitCode = 1;
