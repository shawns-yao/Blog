import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { createHash, randomUUID } from 'node:crypto';
import { readFile, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

// 定向测试: database preparation is explicitly authorized. All indexing and
// answering run in the daily project, never in this script or an independent DB.
const root = new URL('../', import.meta.url);
const conversationChecks = process.argv.includes('--conversation');
const fixChecks = process.argv.includes('--verify-fixes');
const multihopChecks = process.argv.includes('--verify-multihop');
const plannedSamples = multihopChecks ? 1 : fixChecks ? 3 : conversationChecks ? 12 : 43;
const api = conversationChecks ? 'http://127.0.0.1:5173/api/v2' : 'http://127.0.0.1:8080/api/v2';
const output = new URL(multihopChecks ? 'Test/rag-md-multihop-results_20260929.json'
  : fixChecks ? 'Test/rag-md-fix-results_20260929.json'
  : conversationChecks ? 'Test/rag-md-conversation-results_20260929.json'
  : 'Test/rag-md-full-chain-results_20260929.json', root);
const counter = fileURLToPath(new URL('Temp/rag-reference-token-count.exe', root));
const manifest = JSON.parse(await readFile(new URL('Data/manifest/rag-document-ingestion-import_20260929.json', root), 'utf8'));
const ids = manifest.documents.map(d => d.momentId);
assert.deepEqual(ids, [9, 10, 11, 12, 13, 14, 15, 16]);
const report = {
  testType: '定向测试', entry: '/api/v2/public/ask', environment: '日常项目与数据库',
  transport: conversationChecks ? '前端开发代理' : '公开服务端接口',
  source: '八篇真实 Markdown 文章', startedAt: new Date().toISOString(), runStatus: 'running',
  limitations: ['本站定向样本，不属于公开权威评测；小样本不证明最优参数',
    'token 使用 cl100k_base 参考编码，不代表模型计费编码',
    '候选阶段记录实际项目 trace；答案检查含来源、原文位置与关键事实，人工复核另列'],
  phases: [], samples: [], setup: {}, restoration: null,
};
const delay = ms => new Promise(resolve => setTimeout(resolve, ms));
let lastAsk = 0;
let originalValues;
let baseline;
let initialProfile;
const valueTypes = key => ['multiQueryEnabled', 'rerankEnabled', 'rerankFallback', 'dynamicTopKEnabled'].includes(key) ? 'bool'
  : ['indexVersion', 'chatPriority', 'minSimilarity', 'rrfVectorWeight', 'rrfKeywordWeight', 'rerankThreshold', 'bm25K1', 'bm25B'].includes(key) ? 'string' : 'number';

function sql(query) {
  const result = spawnSync('docker', ['exec', '-i', 'shawn-blog-postgres', 'sh', '-lc',
    'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -X -v ON_ERROR_STOP=1 -A -t -q'],
  { input: query, encoding: 'utf8', timeout: 30000, maxBuffer: 8 * 1024 * 1024, windowsHide: true });
  if (result.error || result.status !== 0) throw new Error('database_execution_failed');
  const text = result.stdout.trim();
  return text ? JSON.parse(text) : null;
}
const quote = value => "'" + value.replaceAll("'", "''") + "'";
function saveSettings(values) {
  const rows = Object.entries(values).map(([key, value]) => ({
    key: `rag.${key}`, value: typeof value === 'string' ? value : JSON.stringify(value), type: valueTypes(key),
  }));
  sql(`INSERT INTO sys_config (config_key,value,is_sensitive,group_path,label,description,value_type,sort,meta)
    SELECT key,value,false,'rag',key,'RAG 运行参数',type,0,'{}'::jsonb
    FROM jsonb_to_recordset(${quote(JSON.stringify(rows))}::jsonb) AS p(key text,value text,type text)
    ON CONFLICT(config_key) DO UPDATE SET value=EXCLUDED.value,updated_at=now();`);
}
function documents() {
  return sql(`SELECT COALESCE(jsonb_agg(d ORDER BY d.id),'[]'::jsonb) FROM (
    SELECT m.id,m.title,m.content,m.content_hash,m.short_url,m.created_at,m.is_published,m.deleted_at,
      m.ext_info->>'contentKind' AS kind,s.status,s.attempts,s.last_error,s.indexed_at,s.index_duration_ms,
      s.source_hash,s.active_hash,s.active_profile,s.desired_profile,
      COALESCE((SELECT jsonb_agg(jsonb_build_object('id',c.id,'seq',c.seq,'content',c.content,
        'header',c.context_header,'kind',c.kind,'start',c.start_at,'end',c.end_at,
        'hash',c.source_hash,'profile',c.profile,'dimensions',cardinality(c.embedding),
        'norm',(SELECT sqrt(sum(v*v)) FROM unnest(c.embedding) v)) ORDER BY c.seq)
        FROM rag_chunk c WHERE c.moment_id=m.id),'[]'::jsonb) AS chunks
    FROM moment m LEFT JOIN rag_index_state s ON s.moment_id=m.id WHERE m.id IN (${ids.join(',')})
  ) d;`);
}
function tokenCounts(texts) {
  const result = spawnSync(counter, [], { input: JSON.stringify(texts), encoding: 'utf8', timeout: 30000,
    maxBuffer: 1024 * 1024, windowsHide: true });
  if (result.error || result.status !== 0) throw new Error('reference_measurement_failed');
  return JSON.parse(result.stdout);
}
async function request(path, body) {
  const response = await fetch(api + path, { method: body ? 'POST' : 'GET',
    headers: { 'Content-Type': 'application/json' }, body: body ? JSON.stringify(body) : undefined,
    signal: AbortSignal.timeout(125000) });
  if (!response.ok) throw new Error('http_environment_failed');
  const result = await response.json();
  if (result.code !== 0) throw new Error('project_request_failed');
  return result.data;
}
async function waitIndex(previousProfile, changed, expectedProfile = '') {
  const deadline = Date.now() + 240000;
  let lastSummary = '';
  while (Date.now() < deadline) {
    const docs = documents();
    const summary = docs.map(d => `${d.id}:${d.status}`).join(' ');
    if (summary !== lastSummary) { process.stdout.write(`索引 ${summary}\n`); lastSummary = summary; }
    if (docs.length === 8 && docs.every(d => d.status === 'ready' && d.active_profile === d.desired_profile
      && d.active_hash === d.source_hash && d.chunks.length > 0)
      && (!changed || docs.every(d => d.active_profile !== previousProfile))
      && (!expectedProfile || docs.every(d => d.active_profile === expectedProfile))
      && new Set(docs.map(d => d.active_profile)).size === 1) return docs;
    if (docs.some(d => d.status === 'failed' && d.attempts >= 5)) throw new Error('index_failed');
    await delay(5000);
  }
  throw new Error('index_timeout');
}
function metrics() {
  return sql(`SELECT jsonb_build_object('requests',COALESCE(sum(requests),0),
    'embeddingCalls',COALESCE(sum(embedding_calls),0),'retrievalCalls',COALESCE(sum(retrieval_calls),0),
    'rerankCalls',COALESCE(sum(rerank_calls),0),'rerankDegraded',COALESCE(sum(rerank_degraded),0),
    'generationCalls',COALESCE(sum(generation_calls),0)) FROM rag_query_metric;`);
}
function measure(docs, tuning) {
  const texts = docs.flatMap(d => d.chunks.map(c => c.header + '\n\n' + c.content));
  const counts = tokenCounts(texts);
  let position = 0;
  return docs.map(d => {
    const source = Array.from(d.content);
    const coverage = new Uint16Array(source.length);
    const lengths = [];
    for (const [index, c] of d.chunks.entries()) {
      assert.equal(c.seq, index, '分块序号连续');
      assert.equal(source.slice(c.start, c.end).join(''), c.content, '实际索引块对应原文');
      assert.equal(c.hash, d.active_hash, '实际块指纹一致');
      assert.equal(c.profile, d.active_profile, '实际块版本一致');
      assert.equal(c.dimensions, 1024, 'BGE 向量维度');
      assert(Math.abs(c.norm - 1) < 0.000001, '向量已归一化');
      const tokens = counts[position++];
      assert(tokens <= tuning.chunkMaxTokens, '实际子块超出上限');
      lengths.push(tokens);
      for (let offset = c.start; offset < c.end; offset++) coverage[offset]++;
    }
    let total = 0, missing = 0, duplicate = 0;
    source.forEach((ch, index) => {
      if (/\s/u.test(ch)) return;
      total++;
      if (!coverage[index]) missing++;
      duplicate += Math.max(0, coverage[index] - 1);
    });
    assert.equal(missing, 0, '实际索引没有遗漏原文');
    return { documentId: d.id, chunks: d.chunks.length, coverage: 1, duplicateRate: duplicate / total,
      minTokens: Math.min(...lengths), maxTokens: Math.max(...lengths), tokens: lengths,
      indexDurationMs: d.index_duration_ms, dimensions: 1024 };
  });
}

const facts = [
  { name: '检索必要性', question: 'RAG 为什么先检索文档，再让大模型回答？', ids: [9], concepts: [['依据', '证据', '资料'], ['训练', '记住', '知识', '记忆']] },
  { name: 'PDF 内容判断', question: '怎样判断 PDF 是可搜索的文字页面还是扫描页面？复制文字能作为唯一依据吗？', ids: [10], concepts: [['OCR'], ['文字层'], ['不能', '不应', '不够', '只能']] },
  { name: '图片与截图', question: '有正文的页面，里面的图片和截图应该怎样处理？', ids: [11], concepts: [['OCR'], ['视觉', '关系'], ['位置', '关联']] },
  { name: '表格与公式', question: '识别表格和公式时怎样保留结构与符号？', ids: [12], concepts: [['行列', '单元格'], ['LaTeX'], ['核对', '检查']] },
  { name: '双栏阅读顺序', question: '双栏页面、文字乱序和图片顺序怎么处理？', ids: [13], concepts: [['坐标', '区域'], ['顺序'], ['分块']] },
  { name: '页眉页脚', question: '文档的页眉页脚怎么处理？', ids: [14], concepts: [['保留'], ['噪声'], ['脚注', '条件', '例外']] },
  { name: '重复内容', question: '重复内容应该在文档处理的什么阶段处理？', ids: [15], concepts: [['结构', '来源'], ['噪声', '解析重复'], ['差异', '上下文', '近重复']] },
  { name: '结构分块', question: '文档最后怎样分块，标题和段落应该如何保留？', ids: [16], concepts: [['章节', '标题'], ['段落'], ['上下文', '条件']] },
];
const complex = [
  { name: '两类页面比较', question: '比较可搜索文字页面和扫描页面，两者的解析方式有什么区别？', ids: [10], strategy: 'COMPARE', concepts: [['OCR'], ['文字']] },
  { name: '跨文档多跳', question: '恢复双栏阅读顺序后，如何清理页眉页脚，再按标题和完整段落分块？', ids: [13,14,16], all: true, strategy: 'MULTI_HOP', concepts: [['顺序'], ['页眉', '页脚'], ['标题'], ['段落']] },
];
const probes = [facts[7], complex[1]];

async function ask(sample, tuning, phase, history = []) {
  await delay(Math.max(0, 11000 - (Date.now() - lastAsk)));
  lastAsk = Date.now();
  const started = performance.now();
  const result = { phase, name: sample.name, question: sample.question, expectedIds: sample.ids ?? [],
    passed: false, externalFailure: null, checks: {}, durationMs: 0 };
  report.samples.push(result);
  try {
    const beforeMetrics = metrics();
    const answer = await request('/public/ask', { question: sample.question, contentKind: 'article', sessionId: randomUUID(), history });
    const afterMetrics = metrics();
    result.metricDelta = Object.fromEntries(Object.keys(beforeMetrics).map(key => [key, afterMetrics[key] - beforeMetrics[key]]));
    Object.assign(result, { status: answer.status, mode: answer.mode, answer: answer.answer,
      reason: answer.reason, trace: answer.trace, sourceIds: answer.citations.map(c => c.momentId),
      evidenceSourceIds: answer.trace?.evidenceSourceIds, citations: answer.citations });
    if (answer.status === 'temporarily_unavailable') { result.externalFailure = 'model_or_environment'; return null; }
    const check = (name, condition) => { result.checks[name] = Boolean(condition); };
    if (sample.invalidScope) {
      check('invalidHistoryRejected', answer.status === 'invalid_scope' && answer.citations.length === 0);
      check('noModelDispatch', result.metricDelta.generationCalls === 0 && result.metricDelta.embeddingCalls === 0);
      result.passed = Object.values(result.checks).every(Boolean);
      return answer;
    }
    check('status', sample.noEvidence ? answer.status === 'no_evidence' : answer.status === 'answered');
    if (sample.mode) check('mode', answer.mode === sample.mode);
    if (sample.firstProvider) {
      check('configuredPriority', answer.trace.answerAttempts?.[0] === sample.firstProvider);
      result.preferredProvider = sample.firstProvider;
      result.preferredProviderAnswered = answer.trace.answerProvider === sample.firstProvider;
    }
    if (sample.intent) check('intent', answer.trace.intent === sample.intent);
    if (sample.strategy) check('strategy', answer.trace.strategy === sample.strategy);
    check('originalQuery', answer.trace.originalQuery === sample.question && answer.trace.queries.includes(sample.question));
    check('contextBudget', answer.trace.contextTokens <= tuning.contextMaxTokens);
    check('historyBudget', answer.trace.historyTokens <= tuning.historyMaxTokens);
    if (sample.noHistory) check('historyCleared', answer.trace.historyTokens === 0);
    if (sample.historyReduced) {
      result.inputHistoryTokens = tokenCounts([JSON.stringify(history)])[0];
      check('historyTrimmed', answer.trace.needHistory && answer.trace.historyTokens > 0 && answer.trace.historyTokens < result.inputHistoryTokens);
    }
    if (sample.multiQueryDisabled) check('singleQuery', !answer.trace.needMultiQuery && answer.trace.queries.length === 1);
    if (sample.needMultipleQueries) check('multipleQueries', answer.trace.needMultiQuery && answer.trace.queries.length > 1);
    if (sample.strategy === 'GLOBAL') check('sourceDiversity', new Set(result.sourceIds).size === result.sourceIds.length);
    check('finalTopK', answer.trace.evidenceCount <= tuning.topK);
    check('candidateTopK', answer.trace.fusedCandidates <= tuning.rerankCandidateTopK);
    const queryCount = answer.trace.queries.length;
    check('recallTopK', answer.trace.vectorCandidates <= tuning.vectorTopK * queryCount && answer.trace.keywordCandidates <= tuning.keywordTopK * queryCount);
    if (sample.ids) {
      const matches = sample.ids.map(id => result.sourceIds.includes(id));
      check('correctSources', sample.all ? matches.every(Boolean) : matches.some(Boolean));
      check('grounded', answer.mode === 'grounded');
      const sourceIds = [...new Set(result.sourceIds)];
      const sources = sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',id,'content',content,'published',is_published,'deleted',deleted_at)),'[]'::jsonb) FROM moment WHERE id IN (${sourceIds.length ? sourceIds.join(',') : 'NULL'});`);
      check('liveCitationRanges', answer.citations.length > 0 && answer.citations.every(c => {
        const source = sources.find(s => s.id === c.momentId);
        return source?.published && !source.deleted && Array.from(source.content).slice(c.start,c.end).join('') === c.content;
      }));
      check('facts', sample.concepts.every(alternatives => alternatives.some(word => (answer.answer || '').toLowerCase().includes(word.toLowerCase()))));
      check('retrievalExecuted', answer.trace.vectorCandidates + answer.trace.keywordCandidates > 0);
      if (sample.documentSearch) {
        check('titleSearch', answer.trace.intent === 'document_search');
        check('noModelDispatch', result.metricDelta.embeddingCalls === 0 && result.metricDelta.generationCalls === 0);
      } else {
        check('rerankStage', tuning.rerankEnabled ? result.metricDelta.rerankCalls > 0 : result.metricDelta.rerankCalls === 0);
        if (tuning.rerankEnabled) check('rerankAvailable', result.metricDelta.rerankDegraded === 0);
        check('gptFirst', answer.trace.answerAttempts?.[0] === 'gpt');
      }
    }
    if (sample.gpt) check('gptGeneration', answer.trace.answerProvider === 'gpt');
    result.passed = Object.values(result.checks).every(Boolean);
    return answer;
  } catch (error) {
    result.failure = error.name === 'AssertionError' ? 'assertion_failed' : error.message;
    result.externalFailure = error.name === 'AssertionError' ? null : 'network_or_execution';
    return null;
  } finally {
    result.durationMs = Math.round(performance.now() - started);
    process.stdout.write(`${phase} / ${sample.name}: ${result.externalFailure ? '外部失败' : result.passed ? '通过' : '未通过'} ${result.durationMs}ms\n`);
    await persist();
  }
}
async function persist() {
  const external = report.samples.filter(s => s.externalFailure);
  const core = report.samples.filter(s => !s.externalFailure);
  report.funnel = { original: plannedSamples, excluded: 0, effective: plannedSamples, executed: report.samples.length, pending: plannedSamples - report.samples.length };
  report.external = { failed: external.length, rate: report.samples.length ? external.length / report.samples.length : null };
  report.core = { returned: core.length, passed: core.filter(s => s.passed).length,
    passRate: core.length ? core.filter(s => s.passed).length / core.length : null,
    limitation: external.length > report.samples.length * 0.1 ? '受外部因素影响，仅供参考' : null };
  report.core.byPhase = Object.fromEntries([...new Set(report.samples.map(s => s.phase))].map(phaseName => {
    const phaseSamples = report.samples.filter(s => s.phase === phaseName);
    const phaseCore = phaseSamples.filter(s => !s.externalFailure);
    return [phaseName,{executed:phaseSamples.length,externalFailures:phaseSamples.length-phaseCore.length,
      returned:phaseCore.length,passed:phaseCore.filter(s => s.passed).length}];
  }));
  await writeFile(output, JSON.stringify(report, null, 2) + '\n');
}
async function phase(name, overrides, samples, previousProfile, reindex = false) {
  const tuning = { ...baseline, ...overrides };
  saveSettings(tuning);
  const docs = await waitIndex(previousProfile, reindex);
  const entry = { name, tuning, profile: docs[0].active_profile, chunkChecks: measure(docs, tuning) };
  report.phases.push(entry);
  process.stdout.write(`阶段 ${name}: ${entry.chunkChecks.reduce((n,d) => n+d.chunks,0)} 块\n`);
  await persist();
  for (const sample of samples) await ask(sample, tuning, name);
  return entry.profile;
}

try {
  assert((await request('/public/rag/status')).available, '问答未启用');
  const docs = documents();
  assert.equal(docs.length, 8);
  for (const doc of docs) {
    const source = manifest.documents.find(s => s.momentId === doc.id);
    assert.equal(doc.kind, 'article');
    assert.equal(doc.deleted_at, null);
    assert.equal(createHash('sha256').update(doc.content).digest('hex'), source.processedSha256, '数据库原文与整理副本一致');
  }
  originalValues = sql("SELECT COALESCE(jsonb_object_agg(config_key,value),'{}'::jsonb) FROM sys_config WHERE config_key LIKE 'rag.%' AND is_sensitive=false;");
  baseline = { chunkTargetTokens: 500, chunkMinTokens: 180, chunkMaxTokens: 800, chunkOverlapTokens: 60,
    parentMaxTokens: 1600, contextMaxTokens: 6000, historyMaxTokens: 3000, multiQueryEnabled: true,
    multiQueryMax: 3, bm25K1: 1.2, bm25B: 0.75, vectorTopK: 20, keywordTopK: 20, topK: 6,
    dynamicTopKEnabled: false, dynamicTopKMin: 2, dynamicTopKMax: 12,
    rrfK: 60, rrfVectorWeight: 0.7, rrfKeywordWeight: 0.3, rerankEnabled: true,
    rerankCandidateTopK: 40, rerankThreshold: 0, rerankFallback: true,
    chatPriority: JSON.stringify(['gpt','grok','gemini','opencode_go']) };
  // These are experiment inputs; the project loads and validates them itself.
  report.setup.documents = docs.map(d => ({id:d.id, previouslyPublished:d.is_published, contentHash:d.content_hash}));
  report.setup.originalSettings = originalValues;
  sql(`UPDATE moment SET is_published=true,updated_at=now() WHERE id IN (${ids.join(',')}) AND is_published=false;`);
  if (multihopChecks) {
    initialProfile = await phase('verify-multihop',{},[{...complex[1],needMultipleQueries:true}], '',false);
  } else if (fixChecks) {
    initialProfile = await phase('verify-fixes',{},complex.map(s => ({...s,needMultipleQueries:true})), '',false);
    await ask({name:'降低阈值后的无依据拒答',question:'按照站内文档，2031 年 2 月北极分公司的设备采购限额精确是多少元？',noEvidence:true},baseline,'verify-fixes');
  } else if (conversationChecks) {
    initialProfile = await phase('conversation-baseline', {}, [], '', false);
    const titleSources = sql(`SELECT jsonb_agg(jsonb_build_object('id',id,'title',title)) FROM moment
      WHERE is_published AND deleted_at IS NULL AND ext_info->>'contentKind'='article'
      AND title ~* '(^|[^a-z0-9_])(go|java)($|[^a-z0-9_])';`);
    for (const topic of ['Go','Java']) {
      const sourceIds = (titleSources || []).filter(d => new RegExp(`(^|[^a-z0-9_])${topic}($|[^a-z0-9_])`,'i').test(d.title)).map(d => d.id);
      assert(sourceIds.length > 0, `${topic} 已发布来源存在`);
      await ask({name:`${topic} 标题查找`,question:`有 ${topic} 相关的内容吗`,ids:sourceIds,
        concepts:[[topic]],documentSearch:true},baseline,'conversation-baseline');
    }
    await ask({name:'资料整体概括',question:'请概览本站现有文档讨论的文档入库主题，简要概括内容解析、清理和分块的要点，只总结检索到的资料。',
      ids:[10,13,14,15,16],strategy:'GLOBAL',concepts:[['解析'],['清理','清洗'],['分块']]},baseline,'conversation-baseline');
    const previous = JSON.parse(await readFile(new URL('Test/rag-md-full-chain-results_20260929.json',root),'utf8'));
    const previousAnswers = previous.samples.filter(s => s.status === 'answered' && s.mode === 'grounded' && s.name !== '追问前置事实')
      .sort((left,right) => Array.from(right.answer).length-Array.from(left.answer).length).slice(0,4);
    previousAnswers.push(previous.samples.find(s => s.name === '追问前置事实' && s.status === 'answered'));
    assert(previousAnswers.every(Boolean), '五轮真实回答已记录');
    const history = previousAnswers.flatMap(s => [{role:'user',content:s.question},{role:'assistant',content:s.answer}]);
    const followup = {name:'五轮历史裁剪与指代',question:'那它应该在分块前还是分块后处理？',ids:[14,16],
      strategy:'FOLLOW_UP',concepts:[['分块'],['前','之前','先']],historyReduced:true};
    await ask(followup,baseline,'history-five-rounds',history);
    for (const [name, tokens] of [['history-disabled',0],['history-small-budget',100]]) {
      const tuning = {...baseline,historyMaxTokens:tokens};
      saveSettings(tuning);
      await ask({name,question:followup.question,mode:'conversation',intent:'clarify',noHistory:true},tuning,name,history);
    }
    await phase('context-budget-800',{contextMaxTokens:800,topK:10},[facts[7]],initialProfile);
    await phase('multiquery-disabled',{multiQueryEnabled:false},[{...complex[0],multiQueryDisabled:true}],initialProfile);
    const extra = [{role:'user',content:'你好'},{role:'assistant',content:'你好！'}];
    await ask({name:'超出五轮历史拒绝',question:'它应该怎样处理？',invalidScope:true},baseline,'history-invalid',[...extra,...history]);
    for (const priority of [
      ['grok','gpt','gemini','opencode_go'],
      ['gemini','gpt','grok','opencode_go'],
      ['opencode_go','gpt','grok','gemini'],
    ]) {
      const tuning = {...baseline,chatPriority:JSON.stringify(priority)};
      saveSettings(tuning);
      await ask({name:`${priority[0]} 首选通道`,question:'你好',mode:'conversation',intent:'chat',firstProvider:priority[0]},tuning,'model-priority');
    }
  } else {
  initialProfile = await phase('baseline', {}, [...facts, ...complex], '', false);
  const before = await ask({name:'追问前置事实',question:'文档的页眉页脚怎么处理？',ids:[14],concepts:[['保留']]},baseline,'baseline');
  if (before?.answer) {
    await ask({name:'指代追问',question:'那它应该在分块前还是分块后处理？',ids:[14,16],strategy:'FOLLOW_UP',concepts:[['分块'],['前','之前','先']]},baseline,'baseline',
      [{role:'user',content:'文档的页眉页脚怎么处理？'},{role:'assistant',content:before.answer}]);
    await ask(facts[3],baseline,'topic-switch',
      [{role:'user',content:'文档的页眉页脚怎么处理？'},{role:'assistant',content:before.answer}]);
  }
  await ask({name:'自然问候',question:'你好',mode:'conversation',intent:'chat',gpt:true},baseline,'baseline');
  await ask({name:'无历史指代澄清',question:'它的下一步怎么做？',mode:'conversation',intent:'clarify'},baseline,'baseline');
  await ask({name:'未记载事实',question:'按照站内文档，2031 年 2 月北极分公司的设备采购限额精确是多少元？',noEvidence:true},baseline,'baseline');
  for (const [name, overrides] of [
    ['final-topk-3',{topK:3}], ['final-topk-10',{topK:10}],
    ['recall-topk-10',{vectorTopK:10,keywordTopK:10,rerankCandidateTopK:20}],
    ['recall-topk-30',{vectorTopK:30,keywordTopK:30,rerankCandidateTopK:60}],
    ['bm25-only',{rrfVectorWeight:0,rrfKeywordWeight:1,rerankEnabled:false}],
    ['vector-only',{rrfVectorWeight:1,rrfKeywordWeight:0,rerankEnabled:false}],
    ['rrf-without-rerank',{rerankEnabled:false}],
    ['rrf-k-20',{rrfK:20}], ['rrf-k-100',{rrfK:100}],
  ]) await phase(name,overrides,probes,initialProfile);
  await phase('rerank-threshold-0.2',{rerankThreshold:0.2,multiQueryEnabled:false},[complex[0]],initialProfile);
  let profile = initialProfile;
  for (const [name, overrides] of [
    ['child-target-350',{chunkTargetTokens:350}], ['child-target-700',{chunkTargetTokens:700}],
    ['overlap-0',{chunkOverlapTokens:0}], ['overlap-120',{chunkOverlapTokens:120}],
  ]) profile = await phase(name,overrides,probes,profile,true);
  }
  report.runStatus = 'completed';
} catch (error) {
  report.runStatus = 'failed';
  report.setupFailure = error.message;
  process.exitCode = 1;
} finally {
  if (originalValues && baseline) {
    try {
      const restored = Object.fromEntries(Object.entries(baseline).map(([key,value]) =>
        [key,originalValues[`rag.${key}`] ?? value]));
      saveSettings(restored);
      await delay(3000);
      const docs = await waitIndex('',false,initialProfile);
      report.restoration = { configurationRestored:true, indexReady:true, profile:docs[0].active_profile,
        documentsPublished:docs.every(d => d.is_published), originalContentPreserved:docs.every(d =>
          d.content_hash === report.setup.documents.find(s => s.id === d.id).contentHash) };
    } catch { report.restoration = { configurationRestored:false,indexReady:false }; process.exitCode = 1; }
  }
  report.completedAt = new Date().toISOString();
  await persist();
  if (report.samples.some(s => !s.passed)) process.exitCode = 1;
  process.stdout.write(`${report.testType}: ${report.runStatus}, 核心 ${report.core.passed}/${report.core.returned}, 外部失败 ${report.external.failed}\n`);
}
