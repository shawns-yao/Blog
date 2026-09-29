import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { createHash, randomUUID } from 'node:crypto';
import { readFile, writeFile, appendFile, mkdir } from 'node:fs/promises';
import { resolve, join } from 'node:path';

const name = process.argv[2];
assert(['open-rag-bench', 'beir-scifact'].includes(name), '指定单个测试集合');
const maxTokenArgument = process.argv.slice(3).find(value => value.startsWith('--chunk-max='));
const requestedMaxTokens = maxTokenArgument ? Number(maxTokenArgument.split('=')[1]) : null;
assert(requestedMaxTokens === null || Number.isInteger(requestedMaxTokens), '分块上限必须为整数');
const root = resolve('.'), dir = join(root, 'Temp/rag-benchmarks', name);
const runId = new Date().toISOString().replaceAll(':', '-');
const runDir = join(dir, 'runs', runId);
await mkdir(runDir, { recursive: true });
const dataset = JSON.parse(await readFile(join(dir, 'manifest.json'), 'utf8'));
const documents = JSON.parse(await readFile(join(dir, 'processed/documents.json'), 'utf8'));
const samples = JSON.parse(await readFile(join(dir, 'processed/evaluator-only.json'), 'utf8'));
const hash = value => createHash('sha256').update(value).digest('hex');
const quote = value => "'" + String(value).replaceAll("'", "''") + "'";
const save = async (path, value) => writeFile(path, JSON.stringify(value, null, 2));
const sleep = ms => new Promise(ok => setTimeout(ok, ms));
function sql(query) {
  const result = spawnSync('docker', ['exec', '-i', 'shawn-blog-postgres', 'sh', '-lc',
    'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -X -v ON_ERROR_STOP=1 -A -t -q'],
  { input: query, encoding: 'utf8', windowsHide: true, timeout: 45000, maxBuffer: 16 * 1024 * 1024 });
  if (result.error || result.status !== 0) throw new Error(`database_preparation_failed:${result.stderr?.slice(0, 500)}`);
  return result.stdout.trim() ? JSON.parse(result.stdout) : null;
}

// Scores are computed only after the public project endpoint has returned.
// Rank units are deduplicated original documents/sections, never rebuild-specific chunk IDs.
function rankMetrics(ids, gold, k) {
  const ranking = [...new Set(ids)].slice(0, k), expected = new Set(gold);
  const first = ranking.findIndex(id => expected.has(id));
  const found = ranking.filter(id => expected.has(id)).length;
  const dcg = ranking.reduce((sum, id, index) => sum + (expected.has(id) ? 1 / Math.log2(index + 2) : 0), 0);
  const ideal = Array.from({ length: Math.min(expected.size, k) }, (_, i) => 1 / Math.log2(i + 2)).reduce((a,b) => a+b, 0);
  return { hit: first >= 0 ? 1 : 0, recall: found / expected.size, mrr: first >= 0 ? 1 / (first + 1) : 0, ndcg: ideal ? dcg / ideal : 0 };
}
assert.deepEqual(rankMetrics(['x', 'a', 'a', 'b'], ['a', 'b'], 3).recall, 1);
assert.equal(rankMetrics([], ['a'], 3).mrr, 0);
assert.equal(rankMetrics(['x'], ['a'], 3).ndcg, 0);
assert.equal(rankMetrics(['a'], ['a'], 3).ndcg, 1);
await save(join(runDir, 'evaluator-selftest.json'), { type: '定向测试', checks: 4, passed: 4 });

let imported = [], originalSettings, temporarySettings = {};
const results = [];
const metadata = { objective: 'Establish a real-project retrieval/reranking/answer baseline on a frozen public dataset subset',
  testKind: '公开权威数据测试', testLevel: 'benchmark/end-to-end Markdown RAG',
  entrypoint: 'POST http://127.0.0.1:8080/api/v2/public/ask', dataset,
  goldVisibility: 'Evaluator only; project receives corpus text and query only',
  primaryMetric: 'Final context macro Recall@6 over original document/section IDs; denominator: core project returns',
  auxiliaryMetrics: 'HitRate, MRR and binary nDCG at 1/3/6/10/20 per actual phase; unanswerable/incorrect project outputs score zero',
  rankContract: 'Project ordering retained; document/section IDs deduplicated at first occurrence; no hit scores zero; empty/short ranks are not padded with relevant IDs',
  retryPolicy: 'One project request per sample, no runner retries; actual project provider fallback retained and separately logged',
  passCriteria: 'N/A: initial characterization without an invented acceptance threshold; decision INCONCLUSIVE',
  artifactDir: runDir, startedAt: new Date().toISOString(), status: 'running',
  evaluatorSha256: hash(await readFile(new URL(import.meta.url))),
  limitations: ['Frozen subset changes corpus difficulty; scores are not directly comparable with official full-corpus leaderboards',
    'JSON-to-Markdown is input format adaptation; PDF parsing/OCR are outside this run',
    'BEIR provides relevance labels, not question-answer reference texts'], results: [] };
await save(join(runDir, 'metadata.json'), metadata);
await writeFile(join(dir, 'latest-run.txt'), runId);

function documentSnapshot() {
  return sql(`SELECT COALESCE(jsonb_agg(x),'[]'::jsonb) FROM (
    SELECT m.id,m.short_url,m.content_hash,s.status,s.attempts,s.last_error,s.active_profile,s.desired_profile,
      s.active_hash=s.source_hash AS hash_current,count(c.id) AS chunks
    FROM moment m LEFT JOIN rag_index_state s ON s.moment_id=m.id LEFT JOIN rag_chunk c ON c.moment_id=m.id
    WHERE m.ext_info->>'ragBenchmark'=${quote(name)} GROUP BY m.id,s.moment_id
  ) x;`);
}
function unitIds(hits) {
  const units = [];
  for (const hit of hits) {
    const doc = imported.find(item => item.momentId === hit.momentId);
    if (!doc) throw new Error('corpus_isolation_failed');
    for (const section of doc.sections) {
      if (Math.max(section.start, hit.start) < Math.min(section.end, hit.end)) units.push(section.id);
    }
  }
  return [...new Set(units)];
}

try {
  const publicNotes = sql(`SELECT count(*) FROM moment WHERE is_published=true AND deleted_at IS NULL AND ext_info->>'contentKind'='note';`);
  assert.equal(publicNotes, 0, '临时测试集合必须与原有公开内容隔离');
  originalSettings = sql(`SELECT COALESCE(jsonb_object_agg(config_key,value),'{}'::jsonb) FROM sys_config WHERE config_key LIKE 'rag.%' AND is_sensitive=false;`);
  metadata.originalSettings = originalSettings;
  metadata.sourceSnapshot = sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',id,'hash',content_hash,'published',is_published) ORDER BY id),'[]'::jsonb) FROM moment WHERE ext_info->>'ragBenchmark' IS NULL;`);
  metadata.originalIndexProfiles = sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',m.id,'profile',s.active_profile)),'[]'::jsonb)
    FROM moment m JOIN rag_index_state s ON s.moment_id=m.id WHERE m.ext_info->>'ragBenchmark' IS NULL AND m.is_published=true AND m.deleted_at IS NULL;`);
  if (requestedMaxTokens !== null) {
    temporarySettings = { 'rag.chunkMaxTokens': String(requestedMaxTokens),
      'rag.parentMaxTokens': String(Math.max(requestedMaxTokens, Number(originalSettings['rag.parentMaxTokens']))) };
    sql(`UPDATE sys_config s SET value=v.value,updated_at=now() FROM jsonb_each_text(${quote(JSON.stringify(temporarySettings))}::jsonb) v WHERE s.config_key=v.key AND s.is_sensitive=false;`);
    metadata.temporarySettings = temporarySettings;
  }
  const input = documents.map(doc => ({ id: doc.id, title: Array.from(doc.title).slice(0, 255).join(''), content: doc.content,
    short_url: `rb-${name}-${doc.id}` }));
  const rows = sql(`WITH inserted AS (
    INSERT INTO moment(title,summary,content,content_hash,author_id,toc,short_url,is_published,is_original,ext_info,content_updated_at)
    SELECT d.title,'',d.content,md5(d.content),(SELECT author_id FROM moment WHERE id=9),'[]'::jsonb,d.short_url,true,false,
      jsonb_build_object('contentKind','note','ragBenchmark',${quote(name)},'corpusDocumentId',d.id),now()
    FROM jsonb_to_recordset(${quote(JSON.stringify(input))}::jsonb) AS d(id text,title text,content text,short_url text)
    ON CONFLICT(short_url) DO UPDATE SET is_published=true
      WHERE moment.ext_info->>'ragBenchmark'=${quote(name)} AND moment.content=EXCLUDED.content
    RETURNING id,ext_info->>'corpusDocumentId' AS document_id
  ) SELECT jsonb_agg(inserted) FROM inserted;`);
  imported = rows.map(row => ({ ...documents.find(doc => doc.id === row.document_id), momentId: row.id }));
  assert.equal(imported.length, documents.length, '全部冻结语料写入，不能漏掉困难文档');
  await save(join(runDir, 'source-map.json'), imported.map(({ content, ...document }) => document));
  metadata.importedDocuments = imported.length;
  await save(join(runDir, 'metadata.json'), metadata);
  const deadline = Date.now() + 1500000;
  let lastState = '';
  for (;;) {
    const docs = documentSnapshot();
    const counts = docs.reduce((sum, doc) => { sum[doc.status] = (sum[doc.status] ?? 0) + 1; return sum; }, {});
    const state = JSON.stringify(counts);
    if (lastState !== state) { console.log(`${name} 入库 ${state}`); lastState = state; }
    await save(join(runDir, 'index-snapshot.json'), docs);
    if (docs.length === documents.length && docs.every(doc => doc.status === 'ready' && doc.chunks > 0 && doc.hash_current && doc.active_profile === doc.desired_profile)) break;
    if (docs.some(doc => doc.status === 'failed' && (doc.attempts >= 5 || doc.last_error === 'oversized_atomic_block'))) throw new Error('project_indexing_failed');
    if (Date.now() >= deadline) throw new Error('project_indexing_timeout');
    await sleep(5000);
  }
  let previousAsk = 0;
  for (const [index, sample] of samples.entries()) {
    await sleep(Math.max(0, 11000 - (Date.now() - previousAsk)));
    previousAsk = Date.now();
    const sessionId = randomUUID(), question = name === 'beir-scifact'
      ? `根据知识库资料，请核实以下陈述并说明依据：${sample.question}` : sample.question;
    const result = { id: sample.id, question, sourceQuestion: sample.question, goldIds: sample.goldIds,
      reference: sample.reference, sessionId, externalFailure: null, projectFailure: null, stageMetrics: {} };
    const start = performance.now();
    try {
      const response = await fetch('http://127.0.0.1:8080/api/v2/public/ask', { method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-RAG-Evaluation': '1' },
        body: JSON.stringify({ question, contentKind: 'note', sessionId }), signal: AbortSignal.timeout(125000) });
      if (!response.ok) throw new Error(`transport_http_${response.status}`);
      const payload = await response.json();
      if (payload.code !== 0) { result.projectFailure = `project_response_${payload.code}`; }
      result.answer = payload.data;
      const capturePath = join(root, 'server/Temp/rag-evaluation', `${sessionId}.json`);
      const capture = JSON.parse(await readFile(capturePath, 'utf8'));
      result.capture = capture;
      for (const [stage, value] of Object.entries(capture.stages)) {
        const ranking = unitIds(value.lists[0] ?? []);
        result.stageMetrics[stage] = Object.fromEntries([1, 3, 6, 10, 20].map(k => [k, rankMetrics(ranking, sample.goldIds, k)]));
        result[`${stage}UnitIds`] = ranking;
      }
      if (result.answer.status === 'temporarily_unavailable') {
        const reason = capture.run.Reason ?? capture.run.reason;
        if (['embedding_unavailable','rerank_unavailable','generation_unavailable'].includes(reason)) result.externalFailure = reason;
        else result.projectFailure = reason || 'project_unavailable';
      }
    } catch (error) {
      if (String(error.message).includes('transport_') || error.name === 'TimeoutError' || error.message === 'fetch failed') result.externalFailure = 'transport_or_environment';
      else result.projectFailure = error.message;
    }
    result.durationMs = performance.now() - start;
    results.push(result);
    await appendFile(join(runDir, 'per-sample.jsonl'), JSON.stringify(result) + '\n');
    console.log(`${name} 问答 ${index + 1}/${samples.length} ${result.answer?.status ?? 'failed'} contextRecall@6=${result.stageMetrics.context?.[6]?.recall ?? 0} ${Math.round(result.durationMs)}ms`);
  }
  const core = results.filter(item => !item.externalFailure);
  const summary = { rawSamples: samples.length, excluded: 0, effectiveSamples: samples.length,
    externalFailures: results.length - core.length, projectFailures: results.filter(item => item.projectFailure).length,
    coreReturns: core.length, stages: {}, decision: 'INCONCLUSIVE',
    decisionReason: 'Initial baseline measurement; no predetermined effect acceptance threshold' };
  for (const stage of ['vector','keyword','fused','reranked','context']) {
    summary.stages[stage] = Object.fromEntries([1,3,6,10,20].map(k => [k,
      Object.fromEntries(['hit','recall','mrr','ndcg'].map(metric => [metric, core.length
        ? core.reduce((sum,item) => sum + (item.stageMetrics[stage]?.[k]?.[metric] ?? 0), 0) / core.length : null]))]));
  }
  const durations = results.map(item => item.durationMs).sort((a,b) => a-b);
  summary.p50Ms = durations[Math.ceil(durations.length * .5)-1];
  summary.p95Ms = durations[Math.ceil(durations.length * .95)-1];
  summary.answerStatus = results.reduce((sum,item) => { const key=item.answer?.status ?? 'failed'; sum[key]=(sum[key]??0)+1; return sum; }, {});
  summary.metricQualification = summary.externalFailures / samples.length > .1 ? '受外部因素影响，仅供参考' : 'Frozen subset baseline';
  await save(join(runDir, 'metrics.json'), summary);
  const csv = ['sample_id,status,external_failure,project_failure,fused_recall20,reranked_mrr10,context_recall6,duration_ms',
    ...results.map(item => [item.id,item.answer?.status??'',item.externalFailure??'',item.projectFailure??'',
      item.stageMetrics.fused?.[20]?.recall??0,item.stageMetrics.reranked?.[10]?.mrr??0,item.stageMetrics.context?.[6]?.recall??0,Math.round(item.durationMs)].join(','))].join('\n');
  await writeFile(join(runDir, 'summary.csv'), csv);
  metadata.status = 'completed'; metadata.summary = summary;
} catch (error) {
  metadata.status = 'inconclusive'; metadata.error = error.message;
  console.log(`${name} 阻塞 ${error.message}`);
  process.exitCode = 1;
} finally {
  if (imported.length) {
    sql(`UPDATE moment SET is_published=false WHERE ext_info->>'ragBenchmark'=${quote(name)} AND short_url LIKE ${quote(`rb-${name}-%`)};`);
  }
  if (Object.keys(temporarySettings).length) {
    const restore = Object.fromEntries(Object.keys(temporarySettings).map(key => [key, originalSettings[key]]));
    sql(`UPDATE sys_config s SET value=v.value,updated_at=now() FROM jsonb_each_text(${quote(JSON.stringify(restore))}::jsonb) v WHERE s.config_key=v.key AND s.is_sensitive=false;`);
    const restorationDeadline = Date.now() + 600000;
    for (;;) {
      const states = sql(`SELECT jsonb_agg(jsonb_build_object('id',m.id,'status',s.status,'profile',s.active_profile,'current',s.active_profile=s.desired_profile AND s.active_hash=s.source_hash))
        FROM moment m JOIN rag_index_state s ON s.moment_id=m.id WHERE m.ext_info->>'ragBenchmark' IS NULL AND m.is_published=true AND m.deleted_at IS NULL;`);
      if (states.length === metadata.originalIndexProfiles.length && states.every(state => state.status === 'ready' && state.current &&
        state.profile === metadata.originalIndexProfiles.find(original => original.id === state.id)?.profile)) break;
      if (Date.now() >= restorationDeadline) throw new Error('original_index_restoration_timeout');
      await sleep(5000);
    }
    metadata.configurationRestored = true;
  }
  metadata.finishedAt = new Date().toISOString();
  metadata.restoration = { corpusWithdrawn: imported.length > 0,
    sourceSnapshot: sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',id,'hash',content_hash,'published',is_published) ORDER BY id),'[]'::jsonb) FROM moment WHERE ext_info->>'ragBenchmark' IS NULL;`) };
  if (metadata.sourceSnapshot) assert.deepEqual(metadata.restoration.sourceSnapshot, metadata.sourceSnapshot, '现有内容与发布状态保持不变');
  await save(join(runDir, 'metadata.json'), metadata);
  console.log(`结果 ${runDir}`);
}
