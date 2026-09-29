import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { mkdir, readFile, writeFile } from 'node:fs/promises';

// 定向测试：配置与临时语料是运行准备，索引由项目工作进程执行，问答只调用公开入口。
const root = new URL('../', import.meta.url);
const directory = new URL('Temp/rag-benchmarks/strategy-directed/', root);
await mkdir(directory, { recursive: true });
const output = new URL(`${new Date().toISOString().replaceAll(':', '-')}.json`, directory);
const quote = value => "'" + String(value).replaceAll("'", "''") + "'";
const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));
function sql(query) {
  const result = spawnSync('docker', ['exec', '-i', 'shawn-blog-postgres', 'sh', '-lc',
    'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -X -v ON_ERROR_STOP=1 -A -t -q'],
  { input: query, encoding: 'utf8', windowsHide: true, timeout: 30000, maxBuffer: 4 * 1024 * 1024 });
  if (result.error || result.status !== 0) throw new Error('database_preparation_failed');
  return result.stdout.trim() ? JSON.parse(result.stdout) : null;
}
const keys = ['adaptiveChunkingEnabled', 'adaptiveRetrievalEnabled', 'evidenceSelectionEnabled', 'evidenceDiversityWeight'];
const defaults = { adaptiveChunkingEnabled: true, adaptiveRetrievalEnabled: true, evidenceSelectionEnabled: true, evidenceDiversityWeight: 0.2 };
function configure(values) {
  const rows = Object.entries(values).map(([key, value]) => ({ key: `rag.${key}`, value: String(value), type: typeof value === 'boolean' ? 'bool' : 'string' }));
  sql(`INSERT INTO sys_config(config_key,value,is_sensitive,group_path,label,description,value_type,sort,meta)
    SELECT key,value,false,'rag',key,'RAG 策略运行配置',type,0,'{}'::jsonb
    FROM jsonb_to_recordset(${quote(JSON.stringify(rows))}::jsonb) AS p(key text,value text,type text)
    ON CONFLICT(config_key) DO UPDATE SET value=EXCLUDED.value,updated_at=now();`);
}
function snapshot() {
  return sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',id,'hash',content_hash,'published',is_published) ORDER BY id),'[]'::jsonb)
    FROM moment WHERE ext_info->>'ragBenchmark' IS NULL;`);
}
function states() {
  return sql(`SELECT jsonb_agg(jsonb_build_object('id',m.id,'status',s.status,'reason',s.last_error,'current',s.active_profile=s.desired_profile AND s.active_hash=s.source_hash))
    FROM moment m JOIN rag_index_state s ON s.moment_id=m.id WHERE m.is_published AND m.deleted_at IS NULL;`);
}
async function ready() {
  for (let i = 0; i < 100; i++) {
    const all = states();
    let availability;
    try {
      const response = await fetch('http://127.0.0.1:5173/api/v2/public/rag/status', { signal: AbortSignal.timeout(3000) });
      if (response.ok) availability = (await response.json()).data;
    } catch { /* 后端重启期间继续等待实际服务与工作进程就绪。 */ }
    if (availability?.available && availability.indexReady && all?.length && all.every(item => item.status === 'ready' && item.current)) return;
    if (all?.some(item => item.status === 'failed' && item.reason === 'oversized_atomic_block')) throw new Error('project_indexing_failed');
    await sleep(2000);
  }
  throw new Error('index_timeout');
}
const report = { testType: '定向测试', entry: 'POST http://127.0.0.1:5173/api/v2/public/ask',
  fixtureVersion: 2, startedAt: new Date().toISOString(), indexing: {}, samples: [], restoration: null,
  limitations: ['构造样本与公开评测分开，不证明最优参数；脚本不调用内部函数，不注入向量或候选'] };
const save = () => writeFile(output, JSON.stringify(report, null, 2));
const original = sql(`SELECT jsonb_object_agg(config_key,value) FROM sys_config WHERE config_key IN (${keys.map(key => quote(`rag.${key}`)).join(',')}) AND is_sensitive=false;`) ?? {};
const restore = Object.fromEntries(keys.map(key => [key, JSON.parse(original[`rag.${key}`] ?? String(defaults[key]))]));
const originalSources = snapshot();
const paragraph = 'Reliable retrieval finds relevant source evidence and preserves document boundaries for every answer. '.repeat(13);
const fixtures = [
  { id: 'short', title: '短文边界定向检查', content: [paragraph, paragraph, paragraph, paragraph].join('\n\n') },
  { id: 'forced', title: '长段落重叠定向检查', content: 'A continuous paragraph keeps the reference context close to its statement and preserves exact source ranges. '.repeat(200) },
  { id: 'exact', title: '分块错误码说明', content: '## 错误处理\n\n`ERR_RAG_CHUNK_007` 表示分块超过原子结构上限。处理方式是拆分完整结构单元，并重新发布文档。\n\n```go\nfunc CheckChunk() error {\n  return nil\n}\n```\n\n| 参数 | 含义 |\n| --- | --- |\n| chunkMaxTokens | 普通正文上限 |\n| parentMaxTokens | 父范围上限 |\n' },
  { id: 'atomic', title: '完整公式定向检查', content: `## 公式\n\n完整公式按一个原子结构保存。\n\n$$\n${'\\alpha_i + '.repeat(300)}\\beta\n$$\n` },
];
let imported = [];
try {
  const input = fixtures.map(doc => ({ ...doc, short_url: `rs-directed-v2-${doc.id}` }));
  imported = sql(`WITH inserted AS (
    INSERT INTO moment(title,summary,content,content_hash,author_id,toc,short_url,is_published,is_original,ext_info,content_updated_at)
    SELECT d.title,'',d.content,md5(d.content),(SELECT author_id FROM moment WHERE id=9),'[]'::jsonb,d.short_url,true,false,
      jsonb_build_object('contentKind','note','ragBenchmark','strategy-directed','corpusDocumentId',d.id),now()
    FROM jsonb_to_recordset(${quote(JSON.stringify(input))}::jsonb) AS d(id text,title text,content text,short_url text)
    ON CONFLICT(short_url) DO UPDATE SET is_published=true WHERE moment.ext_info->>'ragBenchmark'='strategy-directed' AND moment.content=EXCLUDED.content
    RETURNING id,ext_info->>'corpusDocumentId' AS document_id
  ) SELECT jsonb_agg(inserted) FROM inserted;`);
  assert.equal(imported.length, fixtures.length);
  for (const mode of ['baseline', 'adaptive']) {
    configure({ ...defaults, adaptiveChunkingEnabled: mode === 'adaptive' });
    await ready();
    const indexed = sql(`SELECT jsonb_agg(x ORDER BY x.id,x.seq) FROM (
      SELECT m.ext_info->>'corpusDocumentId' AS id,c.seq,c.kind,c.content,c.start_at AS start,c.end_at AS end
      FROM moment m JOIN rag_chunk c ON c.moment_id=m.id WHERE m.ext_info->>'ragBenchmark'='strategy-directed'
    ) x;`);
    report.indexing[mode] = indexed;
    for (const fixture of fixtures) {
      const source = Array.from(fixture.content), chunks = indexed.filter(chunk => chunk.id === fixture.id);
      const coverage = new Uint8Array(source.length);
      for (const chunk of chunks) {
        assert.equal(source.slice(chunk.start, chunk.end).join(''), chunk.content, '原文定位一致');
        coverage.fill(1, chunk.start, chunk.end);
      }
      assert(source.every((character, index) => coverage[index] || /\s/.test(character)), '正文覆盖完整');
    }
    await save();
  }
  const chunks = (mode, id) => report.indexing[mode].filter(chunk => chunk.id === id);
  report.indexChecks = {
    shortDocumentReduced: chunks('adaptive', 'short').length < chunks('baseline', 'short').length,
    forcedOverlapExists: chunks('adaptive', 'forced').some((chunk, i, all) => i > 0 && chunk.start < all[i-1].end),
    codePreserved: chunks('adaptive', 'exact').some(chunk => chunk.content.includes('func CheckChunk() error')),
    tablePreserved: chunks('adaptive', 'exact').some(chunk => chunk.content.includes('| chunkMaxTokens |')),
    formulaPreserved: chunks('adaptive', 'atomic').some(chunk => chunk.kind === 'math' && chunk.content.includes('\\beta\n$$')),
  };
  let previousAsk = 0, greetingAnswer;
  const cases = [
    { name: '问候', question: '你好', conversation: true },
    { name: '标题查找', question: '有 Go 相关的内容吗', title: true },
    { name: '事实覆盖', question: '什么是 RAG？', strategy: 'FACT', coverage: true },
    { name: '比较覆盖', question: '比较可搜索文字页面和扫描页面，两者的解析方式有什么区别？', strategy: 'COMPARE', rule: true },
    { name: '三步骤覆盖', question: '恢复双栏阅读顺序后，如何清理页眉页脚，再按标题和完整段落分块？', strategy: 'MULTI_HOP', rule: true, ids: [13, 14, 16] },
    { name: '精确标识符', question: '错误码 ERR_RAG_CHUNK_007 的定义在哪里？', kind: 'note', strategy: 'EXACT', keyword: true, ids: [imported.find(doc => doc.document_id === 'exact').id] },
    { name: '指代与条件保留', question: '它在 Java 17 和 Go 1.22 中有什么区别？', history: () => [{ role: 'user', content: '什么是 RAG？' }, { role: 'assistant', content: greetingAnswer }], protected: ['Java','17','Go','1.22'], allowNoEvidence: true },
    { name: '无依据', question: '根据本站资料，2031 年北极分公司的设备采购限额精确是多少元？', noEvidence: true },
  ];
  for (const item of cases) {
    await sleep(Math.max(0, 11000-(Date.now()-previousAsk))); previousAsk = Date.now();
    const sample = { name: item.name, sessionId: randomUUID(), checks: {}, externalFailure: null, projectFailure: null };
    report.samples.push(sample);
    const started = performance.now();
    try {
      const response = await fetch('http://127.0.0.1:5173/api/v2/public/ask', { method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-RAG-Evaluation': '1' },
        body: JSON.stringify({ question: item.question, contentKind: item.kind ?? 'article', sessionId: sample.sessionId, history: item.history?.() }),
        signal: AbortSignal.timeout(125000) });
      assert(response.ok, `http_${response.status}`);
      const payload = await response.json(); assert.equal(payload.code, 0);
      sample.answer = payload.data;
      sample.capture = JSON.parse(await readFile(new URL(`server/Temp/rag-evaluation/${sample.sessionId}.json`, root), 'utf8'));
      const answer = sample.answer, trace = answer.trace;
      if (answer.status === 'temporarily_unavailable') sample.externalFailure = sample.capture.run.Reason;
      else {
        sample.checks.status = item.allowNoEvidence ? ['answered','no_evidence'].includes(answer.status) : answer.status === (item.noEvidence ? 'no_evidence' : 'answered');
        if (item.conversation) sample.checks.conversation = answer.mode === 'conversation';
        if (item.title) sample.checks.titlePath = trace.vectorCandidates === 0 && trace.keywordCandidates > 0;
        if (item.strategy) sample.checks.strategy = trace.strategy === item.strategy;
        if (item.rule) sample.checks.ruleRoute = trace.understandingSource === 'rule';
        if (item.coverage) sample.checks.recallCapacity = trace.retrievalPolicy.vectorTopK === 20 && trace.retrievalPolicy.keywordTopK === 20 && trace.retrievalPolicy.rerankTopK === 40;
        if (item.keyword) sample.checks.skipsEmbedding = trace.retrievalPolicy.mode === 'keyword' && sample.capture.run.EmbeddingMs === null;
        if (item.ids) sample.checks.sources = item.ids.every(id => answer.citations.some(citation => citation.momentId === id));
        if (item.protected) sample.checks.conditions = item.protected.every(term => trace.query.toLowerCase().includes(term.toLowerCase()));
        sample.checks.budget = trace.contextTokens <= 6000;
        if (item.name === '事实覆盖') greetingAnswer = answer.answer;
      }
    } catch (error) {
      if (error.name === 'TimeoutError' || error.message === 'fetch failed' || error.code === 'ENOENT' || error instanceof SyntaxError)
        sample.externalFailure = 'transport_or_evaluation_environment';
      else sample.projectFailure = error.name + ':' + error.message;
    }
    sample.durationMs = Math.round(performance.now()-started);
    sample.passed = !sample.externalFailure && !sample.projectFailure && Object.values(sample.checks).every(Boolean);
    console.log(`${sample.name}: ${sample.passed ? 'PASS' : 'FAIL'} ${sample.durationMs}ms`);
    await save();
  }
} finally {
  sql(`UPDATE moment SET is_published=false WHERE ext_info->>'ragBenchmark'='strategy-directed' AND short_url LIKE 'rs-directed-%';`);
  configure(restore); await ready();
  assert.deepEqual(snapshot(), originalSources);
  report.restoration = { originalSourcesUnchanged: true, settingsRestored: true, temporaryPublished: sql(`SELECT count(*) FROM moment WHERE ext_info->>'ragBenchmark'='strategy-directed' AND is_published;`) };
  report.funnel = { raw: report.samples.length, excluded: 0, external: report.samples.filter(sample => sample.externalFailure).length,
    projectFailures: report.samples.filter(sample => sample.projectFailure).length,
    coreReturns: report.samples.filter(sample => !sample.externalFailure).length };
  report.finishedAt = new Date().toISOString(); await save(); console.log(output.pathname);
}
if (!Object.values(report.indexChecks ?? {}).every(Boolean) || report.samples.some(sample => !sample.passed)) process.exitCode = 1;
