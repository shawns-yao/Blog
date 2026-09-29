import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { sqlAt, dockerOutput, testTarget, verifyTestContainers } from './benchmark/rag-test-target.mjs';

const dataset = 'beir-scifact';
const target = testTarget(dataset, true);
verifyTestContainers(target);
const samples = JSON.parse(await readFile(`Temp/rag-benchmarks/${dataset}/processed/evaluator-only.json`, 'utf8'));
const sample = samples[2];
const question = `根据知识库资料，请核实以下陈述并说明依据：${sample.question}；请简短回答。`;
const startedAt = new Date().toISOString();
const report = { testKind: '定向测试', scope: '真实项目入口的供应商故障切换与索引复用，不并入公开集合成绩',
  entrypoint: target.endpoint, dataset, sampleId: sample.id, startedAt, scenarios: [] };
function snapshot() {
  return sqlAt(target.databaseContainer, `SELECT jsonb_agg(x ORDER BY x.id) FROM (
    SELECT s.moment_id AS id,s.status,s.attempts,s.active_profile,s.desired_profile,s.indexed_at,
      s.source_hash,s.active_hash,count(c.id) AS chunks,max(c.id) AS latest_chunk_id
    FROM rag_index_state s LEFT JOIN rag_chunk c ON c.moment_id=s.moment_id GROUP BY s.moment_id
  ) x;`);
}
const before = snapshot();
assert.deepEqual(sqlAt(target.databaseContainer, 'SELECT jsonb_agg(DISTINCT cardinality(embedding)) FROM rag_chunk;'), [4096],
  '故障检查必须复用 Qwen 的固定 4096 维索引');
assert(before.length === 100 && before.every(item => item.status === 'ready' && item.active_profile === item.desired_profile),
  '先完成同空间 Qwen 实际入库');
async function start(faults = []) {
  const child = spawn(process.execPath, ['Test/benchmark/rag-test-environment.mjs', 'start', dataset, '--models=qwen',
    ...faults.map(value => '--fault=' + value)], { stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true });
  // Compose diagnostics are not needed for scoring and may contain local configuration.
  child.stdout.resume(); child.stderr.resume();
  const code = await new Promise((ok, fail) => { child.on('error', fail); child.on('exit', ok); });
  assert.equal(code, 0, '专用项目启动成功');
}
async function ask() {
  const sessionId = randomUUID(), started = performance.now();
  const payload = await (await fetch(target.endpoint, { method: 'POST', redirect: 'error',
    headers: { 'Content-Type': 'application/json', 'X-RAG-Evaluation': '1' },
    body: JSON.stringify({question, contentKind: 'note', sessionId}), signal: AbortSignal.timeout(125000) })).json();
  assert.equal(payload.code, 0, '项目正常返回');
  const capture = JSON.parse(dockerOutput(['exec', target.serverContainer, 'cat', `/evaluation/${sessionId}.json`]));
  return { answer: payload.data, durationMs: Math.round(performance.now() - started), run: capture.run };
}
const scenarios = [
  {name: 'primary', faults: [], verify: ({answer}) => {
    assert.equal(answer.trace.embeddingAttempts[0], 'hybgzs');
    assert.equal(answer.trace.rerankProvider, 'tumuer_bge');
    assert(answer.trace.vectorCandidates > 0);
  }},
  {name: 'primary_failure', faults: ['embedding-primary', 'rerank-primary'], verify: ({answer}) => {
    assert.equal(answer.trace.embeddingProvider, 'tumuer');
    assert.equal(answer.trace.embeddingFallbackUsed, true);
    assert.equal(answer.trace.embeddingDegraded, false);
    assert(answer.trace.rerankAttempts.includes('tumuer_qwen'));
    assert(answer.trace.embeddingFailures.some(item => item.provider === 'hybgzs' && item.reason === 'network_error'));
  }},
  {name: 'cooldown', faults: null, verify: ({answer}) => {
    assert.equal(answer.trace.embeddingProvider, 'tumuer');
    assert(!answer.trace.embeddingAttempts.includes('hybgzs'));
    assert(answer.trace.embeddingFailures.some(item => item.provider === 'hybgzs' && item.reason === 'cooldown'));
  }},
  {name: 'all_rerank_failure', faults: ['rerank-all'], verify: ({answer,run}) => {
    assert.equal(answer.trace.rerankProvider, undefined);
    assert.equal(run.RerankDegraded, true);
    assert(answer.trace.rerankedCandidates > 0);
  }},
  {name: 'all_vector_and_rerank_failure', faults: ['embedding-all', 'rerank-all'], verify: ({answer,run}) => {
    assert.equal(answer.trace.embeddingDegraded, true);
    assert.equal(answer.trace.vectorCandidates, 0);
    assert(answer.trace.keywordCandidates > 0 && answer.trace.evidenceCount > 0);
    assert.equal(run.RerankDegraded, true);
  }},
];
try {
  for (const scenario of scenarios) {
    report.currentScenario = scenario.name;
    if (scenario.faults !== null) await start(scenario.faults);
    const result = await ask();
    const observed = {scenario: scenario.name, faults: scenario.faults, routingAssertionsPassed: false,
      ...result, externalGenerationFailure: result.answer.status === 'temporarily_unavailable'};
    report.scenarios.push(observed);
    scenario.verify(result);
    assert.deepEqual(snapshot(), before, '换站、重排及故障配置不重建同空间向量');
    observed.routingAssertionsPassed = true;
    console.log(JSON.stringify({scenario: scenario.name, status: result.answer.status, durationMs: result.durationMs,
      embeddingProvider: result.answer.trace.embeddingProvider, rerankProvider: result.answer.trace.rerankProvider}));
    if (scenario.name === 'primary_failure') await new Promise(ok => setTimeout(ok, 11000));
  }
  report.routingAssertionsPassed = true;
} catch (error) {
  report.routingAssertionsPassed = false;
  report.error = error.name;
  process.exitCode = 1;
  console.log(JSON.stringify({failed: true, scenario: report.currentScenario, error: error.name}));
} finally {
  await start();
  report.indexReused = JSON.stringify(snapshot()) === JSON.stringify(before);
  assert(report.indexReused, '恢复正常主渠道后仍复用索引');
  report.finishedAt = new Date().toISOString();
  await mkdir('Temp/rag-benchmarks/provider-failover', {recursive: true});
  const path = 'Temp/rag-benchmarks/provider-failover/' + startedAt.replace(/[:.]/g, '-') + '.json';
  await writeFile(path, JSON.stringify(report,null,2) + '\n');
  console.log(JSON.stringify({artifact: path, routingAssertionsPassed: report.routingAssertionsPassed, indexReused: report.indexReused}));
}
