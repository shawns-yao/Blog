import assert from 'node:assert/strict';
import { createHash, randomUUID } from 'node:crypto';
import { readFile, writeFile, appendFile, mkdir } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { dockerOutput, sqlAt, testTarget, verifyTestContainers } from './rag-test-target.mjs';

const name = process.argv[2];
assert(['open-rag-bench', 'beir-scifact'].includes(name), '指定单个测试集合');
const maxTokenArgument = process.argv.slice(3).find(value => value.startsWith('--chunk-max='));
const requestedMaxTokens = maxTokenArgument ? Number(maxTokenArgument.split('=')[1]) : null;
assert(requestedMaxTokens === null || Number.isInteger(requestedMaxTokens), '分块上限必须为整数');
const sampleArgument = process.argv.slice(3).find(value => value.startsWith('--sample-ids='));
const selectedIds = sampleArgument ? sampleArgument.slice('--sample-ids='.length).split(',') : null;
assert(selectedIds === null || selectedIds.every(Boolean), '定向样本 ID 不能为空');
const testKind = selectedIds ? '定向测试' : '公开权威数据测试';
const indexTimeout = Number(process.argv.slice(3).find(value => value.startsWith('--index-timeout-minutes='))?.split('=')[1] ?? 25);
assert(Number.isInteger(indexTimeout) && indexTimeout > 0 && indexTimeout <= 120, '入库等待预算为 1–120 分钟');
const profileArgument = process.argv.slice(3).find(value => value.startsWith('--strategy-profile='));
const strategyProfile = profileArgument?.split('=')[1] ?? null;
assert(strategyProfile === null || ['baseline', 'adaptive'].includes(strategyProfile), '策略对照只能是 baseline 或 adaptive');
const retrievalProfile = process.argv.slice(3).find(value => value.startsWith('--retrieval-profile='))?.split('=')[1] ?? null;
assert(retrievalProfile === null || ['rrf', 'rerank', 'selection'].includes(retrievalProfile), '检索对照只能是 rrf、rerank 或 selection');
const environmentArgument = process.argv.slice(3).find(value => value.startsWith('--environment='));
const environment = environmentArgument?.split('=')[1] ?? 'docker';
assert(['daily', 'docker'].includes(environment), '环境只能是 daily 或 docker');
const target = testTarget(name, environment === 'docker');
if (target.dedicated) verifyTestContainers(target);
const providerFields = ['RAG_EMBEDDING_PROVIDER', 'RAG_EMBEDDING_MODEL', 'RAG_EMBEDDING_DIMENSIONS', 'RAG_EMBEDDING_SPACE_ID',
  'RAG_EMBEDDING_QUERY_INSTRUCTION', 'RAG_EMBEDDING_TIMEOUT', 'RAG_EMBEDDING_STAGE_TIMEOUT', 'RAG_EMBEDDING_INDEX_TIMEOUT',
  'RAG_EMBEDDING_FALLBACK_PROVIDER', 'RAG_EMBEDDING_FALLBACK_MODEL', 'RAG_RERANK_PROVIDER', 'RAG_RERANK_MODEL',
  'RAG_RERANK_TIMEOUT', 'RAG_RERANK_STAGE_TIMEOUT', 'RAG_RERANK_FALLBACK_PROVIDER', 'RAG_RERANK_FALLBACK_MODEL',
  'RAG_RERANK_LAST_RESORT_PROVIDER', 'RAG_RERANK_LAST_RESORT_MODEL'];
const runtimeEnv = JSON.parse(dockerOutput(['inspect', '--format', '{{json .Config.Env}}', target.serverContainer]));
const providerConfig = Object.fromEntries(runtimeEnv.map(line => [line.slice(0, line.indexOf('=')), line.slice(line.indexOf('=') + 1)])
  .filter(([key]) => providerFields.includes(key)));
const root = resolve('.'), dir = join(root, 'Temp/rag-benchmarks', name);
const runId = new Date().toISOString().replaceAll(':', '-');
const runDir = join(dir, selectedIds ? 'directed-runs' : 'runs', runId);
await mkdir(runDir, { recursive: true });
const dataset = JSON.parse(await readFile(join(dir, 'manifest.json'), 'utf8'));
const documents = JSON.parse(await readFile(join(dir, 'processed/documents.json'), 'utf8'));
const allSamples = JSON.parse(await readFile(join(dir, 'processed/evaluator-only.json'), 'utf8'));
const samples = selectedIds ? allSamples.filter(sample => selectedIds.includes(String(sample.id))) : allSamples;
assert(!selectedIds || samples.length === new Set(selectedIds).size, '所有定向样本必须来自已冻结的数据集');
const hash = value => createHash('sha256').update(value).digest('hex');
const quote = value => "'" + String(value).replaceAll("'", "''") + "'";
const save = async (path, value) => writeFile(path, JSON.stringify(value, null, 2));
const sleep = ms => new Promise(ok => setTimeout(ok, ms));
function sql(query) {
  return sqlAt(target.databaseContainer, query);
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
const metadata = { objective: selectedIds ? 'Verify identified failures using original queries and the unchanged frozen corpus'
  : 'Establish a real-project retrieval/reranking/answer baseline on a frozen public dataset subset',
  testKind, selectedIds, testLevel: 'benchmark/end-to-end Markdown RAG',
  strategyProfile, retrievalProfile, environment: target, providerConfig, indexTimeoutMinutes: indexTimeout,
  entrypoint: `POST ${target.endpoint}`, dataset,
  goldVisibility: 'Evaluator only; project receives corpus text and query only',
  primaryMetric: 'Final dispatched context macro recall over original document/section IDs; Recall@6 is reported separately; denominator: core project returns',
  auxiliaryMetrics: 'HitRate, MRR and binary nDCG at 1/3/6/10/20 per actual phase; unanswerable/incorrect project outputs score zero',
  rankContract: 'Project ordering retained; document/section IDs deduplicated at first occurrence; no hit scores zero; empty/short ranks are not padded with relevant IDs',
  retryPolicy: 'One project request per sample, no runner retries; actual project provider fallback retained and separately logged',
  passCriteria: selectedIds ? 'Directed defect evidence only; assess each specified failure separately, not a public benchmark acceptance score'
    : 'N/A: initial characterization without an invented acceptance threshold; decision INCONCLUSIVE',
  artifactDir: runDir, startedAt: new Date().toISOString(), status: 'running',
  evaluatorSha256: hash(await readFile(new URL(import.meta.url))),
  projectSourceSha256: Object.fromEntries(await Promise.all([
    'internal/app/rag/service.go', 'internal/app/rag/retrieval.go', 'internal/app/rag/evidence_selection.go',
    'internal/app/rag/context.go', 'internal/app/rag/topk.go', 'internal/domain/rag/entity.go',
    'internal/infra/ai/client.go', 'internal/infra/ai/openai.go', 'internal/infra/ai/rag_chat.go',
    'internal/app/rag/query_understand.go',
  ].map(async path => [path, hash(await readFile(join(root, 'server', path)))]))),
  limitations: ['Frozen subset changes corpus difficulty; scores are not directly comparable with official full-corpus leaderboards',
    'JSON-to-Markdown is input format adaptation; PDF parsing/OCR are outside this run',
    'BEIR provides relevance labels, not question-answer reference texts'], results: [] };
await save(join(runDir, 'metadata.json'), metadata);
await writeFile(join(dir, selectedIds ? 'latest-directed-run.txt' : 'latest-run.txt'), runId);

function documentSnapshot() {
  return sql(`SELECT COALESCE(jsonb_agg(x),'[]'::jsonb) FROM (
    SELECT m.id,m.short_url,m.content_hash,s.status,s.attempts,s.last_error,s.active_profile,s.desired_profile,s.indexed_at,
      s.active_hash=s.source_hash AS hash_current,count(c.id) AS chunks,max(c.id) AS latest_chunk_id
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
  const publicNotes = sql(`SELECT count(*) FROM moment WHERE is_published=true AND deleted_at IS NULL AND ext_info->>'contentKind'='note'
    ${target.dedicated ? `AND (ext_info->>'ragBenchmark' IS DISTINCT FROM ${quote(name)})` : ''};`);
  assert.equal(publicNotes, 0, '临时测试集合必须与原有公开内容隔离');
  if (target.dedicated) assert.equal(sql("SELECT to_json(value) FROM sys_config WHERE config_key='test.rag.dataset';"), name,
    '测试数据库必须属于当前集合');
  originalSettings = sql(`SELECT COALESCE(jsonb_object_agg(config_key,value),'{}'::jsonb) FROM sys_config WHERE config_key LIKE 'rag.%' AND is_sensitive=false;`);
  metadata.originalSettings = originalSettings;
  metadata.sourceSnapshot = sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',id,'hash',content_hash,'published',is_published) ORDER BY id),'[]'::jsonb) FROM moment WHERE ext_info->>'ragBenchmark' IS NULL;`);
  metadata.originalIndexProfiles = sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',m.id,'profile',s.active_profile)),'[]'::jsonb)
    FROM moment m JOIN rag_index_state s ON s.moment_id=m.id WHERE m.ext_info->>'ragBenchmark' IS NULL AND m.is_published=true AND m.deleted_at IS NULL;`);
  if (requestedMaxTokens !== null || strategyProfile !== null || retrievalProfile !== null) {
    if (requestedMaxTokens !== null) Object.assign(temporarySettings, { 'rag.chunkMaxTokens': String(requestedMaxTokens),
      'rag.parentMaxTokens': String(Math.max(requestedMaxTokens, Number(originalSettings['rag.parentMaxTokens']))) });
    if (strategyProfile !== null) {
      for (const key of ['adaptiveChunkingEnabled','adaptiveRetrievalEnabled','evidenceSelectionEnabled']) {
        assert(originalSettings[`rag.${key}`] !== undefined, '先在运行配置中登记新策略开关');
        temporarySettings[`rag.${key}`] = String(strategyProfile === 'adaptive');
      }
    }
    if (retrievalProfile !== null) {
      // Only query-time settings change; corpus, chunking and embedding identity remain intact.
      temporarySettings['rag.rerankEnabled'] = String(retrievalProfile !== 'rrf');
      temporarySettings['rag.evidenceSelectionEnabled'] = String(retrievalProfile === 'selection');
      for (const key of Object.keys(temporarySettings)) assert(originalSettings[key] !== undefined, '对照配置必须已经登记');
    }
    sql(`UPDATE sys_config s SET value=v.value,updated_at=now() FROM jsonb_each_text(${quote(JSON.stringify(temporarySettings))}::jsonb) v WHERE s.config_key=v.key AND s.is_sensitive=false;`);
    metadata.temporarySettings = temporarySettings;
  }
  const input = documents.map(doc => ({ id: doc.id, title: Array.from(doc.title).slice(0, 255).join(''), content: doc.content,
    short_url: `rb-${name}-${doc.id}` }));
  metadata.indexBefore = target.dedicated ? documentSnapshot() : [];
  const author = target.dedicated ? "(SELECT id FROM app_user WHERE username='rag-benchmark-source' AND is_active=false AND is_admin=false)"
    : '(SELECT author_id FROM moment WHERE id=9)';
  sql(`
    INSERT INTO moment(title,summary,content,content_hash,author_id,toc,short_url,is_published,is_original,ext_info,content_updated_at)
    SELECT d.title,'',d.content,md5(d.content),${author},'[]'::jsonb,d.short_url,true,false,
      jsonb_build_object('contentKind','note','ragBenchmark',${quote(name)},'corpusDocumentId',d.id),now()
    FROM jsonb_to_recordset(${quote(JSON.stringify(input))}::jsonb) AS d(id text,title text,content text,short_url text)
    ON CONFLICT(short_url) DO UPDATE SET is_published=true
      WHERE moment.ext_info->>'ragBenchmark'=${quote(name)} AND moment.content=EXCLUDED.content AND moment.is_published=false;`);
  const rows = sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',m.id,'document_id',d.id)),'[]'::jsonb)
    FROM moment m JOIN jsonb_to_recordset(${quote(JSON.stringify(input))}::jsonb) AS d(id text,title text,content text,short_url text)
      ON m.short_url=d.short_url AND m.content=d.content AND m.title=d.title
    WHERE m.ext_info->>'ragBenchmark'=${quote(name)} AND m.ext_info->>'corpusDocumentId'=d.id AND m.is_published=true AND m.deleted_at IS NULL;`);
  imported = rows.map(row => ({ ...documents.find(doc => doc.id === row.document_id), momentId: row.id }));
  assert.equal(imported.length, documents.length, '全部冻结语料写入，不能漏掉困难文档');
  await save(join(runDir, 'source-map.json'), imported.map(({ content, ...document }) => document));
  metadata.importedDocuments = imported.length;
  await save(join(runDir, 'metadata.json'), metadata);
  const deadline = Date.now() + indexTimeout * 60000;
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
  metadata.reusedReadyDocuments = documentSnapshot().filter(doc => metadata.indexBefore.some(before =>
    before.id === doc.id && before.status === 'ready' && before.hash_current && before.active_profile === doc.active_profile &&
    before.attempts === doc.attempts && before.chunks === doc.chunks && before.indexed_at === doc.indexed_at &&
    before.latest_chunk_id === doc.latest_chunk_id)).length;
  if (target.dedicated) console.log(`${name} 复用索引 ${metadata.reusedReadyDocuments}/${documents.length} 篇`);
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
      const response = await fetch(target.endpoint, { method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-RAG-Evaluation': '1' },
        body: JSON.stringify({ question, contentKind: 'note', sessionId }), signal: AbortSignal.timeout(125000) });
      if (!response.ok) throw new Error(`transport_http_${response.status}`);
      const payload = await response.json();
      if (payload.code !== 0) { result.projectFailure = `project_response_${payload.code}`; }
      result.answer = payload.data;
      const capture = JSON.parse(target.dedicated
        ? dockerOutput(['exec', target.serverContainer, 'cat', `/evaluation/${sessionId}.json`])
        : await readFile(join(root, 'server/Temp/rag-evaluation', `${sessionId}.json`), 'utf8'));
      result.capture = capture;
      for (const [stage, value] of Object.entries(capture.stages)) {
        const ranking = unitIds(value.lists[0] ?? []);
        result.stageMetrics[stage] = Object.fromEntries([1, 3, 6, 10, 20].map(k => [k, rankMetrics(ranking, sample.goldIds, k)]));
        result[`${stage}UnitIds`] = ranking;
      }
      result.finalContextRecall = rankMetrics(result.contextUnitIds ?? [], sample.goldIds, Number.MAX_SAFE_INTEGER).recall;
      if (result.answer.status === 'temporarily_unavailable') {
        const reason = capture.run.Reason ?? capture.run.reason;
        if (['embedding_unavailable','rerank_unavailable','generation_unavailable'].includes(reason)) result.externalFailure = reason;
        else result.projectFailure = reason || 'project_unavailable';
      }
    } catch (error) {
      if (String(error.message).includes('transport_') || error.name === 'TimeoutError' || error.message === 'fetch failed') result.externalFailure = 'transport_or_environment';
      else if (['ENOENT','RAG_TEST_DOCKER_FAILED'].includes(error.code) || error instanceof SyntaxError) result.externalFailure = 'evaluation_execution_failed';
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
    decisionReason: selectedIds ? 'Directed subset; no claim about complete benchmark quality'
      : 'Initial baseline measurement; no predetermined effect acceptance threshold' };
  for (const stage of ['vector','keyword','fused','reranked','context']) {
    summary.stages[stage] = Object.fromEntries([1,3,6,10,20].map(k => [k,
      Object.fromEntries(['hit','recall','mrr','ndcg'].map(metric => [metric, core.length
        ? core.reduce((sum,item) => sum + (item.stageMetrics[stage]?.[k]?.[metric] ?? 0), 0) / core.length : null]))]));
  }
  const durations = results.map(item => item.durationMs).sort((a,b) => a-b);
  summary.p50Ms = durations[Math.ceil(durations.length * .5)-1];
  summary.p95Ms = durations[Math.ceil(durations.length * .95)-1];
  const mean = values => values.length ? values.reduce((a,b) => a+b,0)/values.length : null;
  const percentile = (values, fraction) => values.length ? [...values].sort((a,b) => a-b)[Math.ceil(values.length*fraction)-1] : null;
  summary.finalContextRecall = mean(core.map(item => item.finalContextRecall ?? 0));
  summary.policy = {
    averageVectorTopK: mean(core.map(item => item.answer?.trace?.retrievalPolicy?.vectorTopK).filter(Number.isFinite)),
    averageKeywordTopK: mean(core.map(item => item.answer?.trace?.retrievalPolicy?.keywordTopK).filter(Number.isFinite)),
    averageRerankTopK: mean(core.map(item => item.answer?.trace?.retrievalPolicy?.rerankTopK).filter(Number.isFinite)),
    averageFinalK: mean(core.map(item => item.answer?.trace?.evidenceCount ?? 0)),
    p95FinalK: percentile(core.map(item => item.answer?.trace?.evidenceCount ?? 0), .95),
    averageContextTokens: mean(core.map(item => item.answer?.trace?.contextTokens ?? 0)),
  };
  summary.coreLatency = { p50Ms: percentile(core.map(item => item.durationMs), .5), p95Ms: percentile(core.map(item => item.durationMs), .95),
    averageUnderstandingMs: mean(core.map(item => item.answer?.trace?.understandingMs).filter(Number.isFinite)),
    averageEmbeddingMs: mean(core.map(item => item.capture?.run?.EmbeddingMs).filter(Number.isFinite)),
    averageRerankMs: mean(core.map(item => item.capture?.run?.RerankMs).filter(Number.isFinite)),
    averageGenerationMs: mean(core.map(item => item.capture?.run?.GenerationMs).filter(Number.isFinite)) };
  summary.answerStatus = results.reduce((sum,item) => { const key=item.answer?.status ?? 'failed'; sum[key]=(sum[key]??0)+1; return sum; }, {});
  const stageLoss = (before, after, k) => {
    const eligible = core.filter(item => item.stageMetrics[before]?.[k]?.hit === 1 && item.stageMetrics[after]?.[k]);
    const lost = eligible.filter(item => item.stageMetrics[after][k].hit === 0);
    return { k, relevantUnit: name === 'beir-scifact' ? 'document' : 'section', eligible: eligible.length,
      lost: lost.length, rate: eligible.length ? lost.length / eligible.length : null, sampleIds: lost.map(item => item.id) };
  };
  summary.evidenceLoss = Object.fromEntries([1, 6].map(k => [k, {
    rerank: stageLoss('fused', 'reranked', k), selection: stageLoss('reranked', 'selection', k),
    context: stageLoss('reranked', 'context', k),
  }]));
  summary.refusalsWithRelevantContext = core.filter(item => item.answer?.status === 'no_evidence' && item.finalContextRecall > 0).map(item => item.id);
  // A relevant document can lack the precise assertion; this count is not a false-refusal score.
  summary.metricQualification = summary.externalFailures / samples.length > .1 ? '受外部因素影响，仅供参考'
    : selectedIds ? 'Directed failures only; excluded from public benchmark aggregates' : 'Frozen subset baseline';
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
  if (imported.length && !target.dedicated) {
    sql(`UPDATE moment SET is_published=false WHERE ext_info->>'ragBenchmark'=${quote(name)} AND short_url LIKE ${quote(`rb-${name}-%`)};`);
  }
  if (Object.keys(temporarySettings).length && !target.dedicated) {
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
  metadata.configurationRetained = target.dedicated;
  metadata.finishedAt = new Date().toISOString();
  metadata.restoration = { corpusWithdrawn: imported.length > 0 && !target.dedicated, corpusRetained: target.dedicated,
    sourceSnapshot: sql(`SELECT COALESCE(jsonb_agg(jsonb_build_object('id',id,'hash',content_hash,'published',is_published) ORDER BY id),'[]'::jsonb) FROM moment WHERE ext_info->>'ragBenchmark' IS NULL;`) };
  if (metadata.sourceSnapshot) assert.deepEqual(metadata.restoration.sourceSnapshot, metadata.sourceSnapshot, '现有内容与发布状态保持不变');
  await save(join(runDir, 'metadata.json'), metadata);
  console.log(`结果 ${runDir}`);
}
