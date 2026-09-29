import assert from 'node:assert/strict';
import { randomUUID } from 'node:crypto';
import { mkdir, readFile, writeFile } from 'node:fs/promises';

// 定向测试：直接调用日常前端代理入口，不准备数据库，不调用项目内部模块。
const root = new URL('../', import.meta.url);
const directory = new URL('Temp/rag-benchmarks/query-routing/', root);
await mkdir(directory, { recursive: true });
const output = new URL(`${new Date().toISOString().replaceAll(':', '-')}.json`, directory);
const samples = [
  { question: '请核实：RAG 会先检索外部知识，再让语言模型根据这些证据生成回答。', rule: true },
  { question: 'Verify whether Retrieval-Augmented Generation (RAG) retrieves external knowledge before generating an answer.', rule: true, ordinaryTerm: 'answer' },
  { question: 'Does retrieval-augmented generation retrieve external knowledge before generating an answer?', rule: true },
  { question: '这个是否正确？请核实一下。', clarify: true },
];
const report = { testType: '定向测试', entry: 'POST http://127.0.0.1:5173/api/v2/public/ask', samples: [] };
for (const item of samples) {
  const sessionId = randomUUID(), started = performance.now();
  const sample = { question: item.question, sessionId, checks: {}, externalFailure: null, projectFailure: null };
  try {
    const response = await fetch('http://127.0.0.1:5173/api/v2/public/ask', { method: 'POST',
      headers: { 'Content-Type': 'application/json', 'X-RAG-Evaluation': '1' },
      body: JSON.stringify({ question: item.question, contentKind: 'article', sessionId }), signal: AbortSignal.timeout(125000) });
    assert(response.ok, `http_${response.status}`);
    const payload = await response.json(); assert.equal(payload.code, 0);
    sample.answer = payload.data;
    sample.capture = JSON.parse(await readFile(new URL(`server/Temp/rag-evaluation/${sessionId}.json`, root), 'utf8'));
    if (sample.answer.status === 'temporarily_unavailable') sample.externalFailure = sample.capture.run.Reason;
    else if (item.rule) {
      sample.checks.answered = sample.answer.status === 'answered' && sample.answer.citations.length > 0;
      sample.checks.rule = sample.answer.trace.understandingSource === 'rule' && sample.answer.trace.understandingMs === 0;
      sample.checks.strategy = sample.answer.trace.strategy === 'FACT';
      const policy = sample.answer.trace.retrievalPolicy;
      sample.checks.recallCapacity = policy.version === 'query-policy-v2' && policy.vectorTopK === 20 && policy.keywordTopK === 20 && policy.rerankTopK === 40;
      sample.checks.budget = sample.answer.trace.contextTokens <= 6000;
      if (item.ordinaryTerm) sample.checks.ordinaryTermNotProtected = !sample.answer.trace.protectedTerms.includes(item.ordinaryTerm);
    } else sample.checks.clarifiesMissingReferent = sample.answer.status === 'answered' &&
      sample.answer.mode === 'conversation' && sample.answer.trace.intent === 'clarify' && sample.answer.citations.length === 0;
  } catch (error) {
    if (error.name === 'TimeoutError' || error.message === 'fetch failed' || error.code === 'ENOENT' || error instanceof SyntaxError)
      sample.externalFailure = 'transport_or_evaluation_environment';
    else sample.projectFailure = `${error.name}:${error.message}`;
  }
  sample.durationMs = Math.round(performance.now()-started);
  sample.passed = !sample.externalFailure && !sample.projectFailure && Object.values(sample.checks).every(Boolean);
  report.samples.push(sample);
  await writeFile(output, JSON.stringify(report, null, 2));
  console.log(`${sample.passed ? 'PASS' : 'FAIL'} ${sample.durationMs}ms`);
  await new Promise(resolve => setTimeout(resolve, 11000));
}
report.funnel = { raw: samples.length, excluded: 0, external: report.samples.filter(sample => sample.externalFailure).length,
  projectFailures: report.samples.filter(sample => sample.projectFailure).length,
  coreReturns: report.samples.filter(sample => !sample.externalFailure).length };
await writeFile(output, JSON.stringify(report, null, 2));
console.log(output.pathname);
if (report.samples.some(sample => !sample.passed)) process.exitCode = 1;
