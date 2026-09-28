import assert from 'node:assert/strict';
import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

const entry = process.env.RAG_UI_TEST_URL || 'http://127.0.0.1:5173/';
const sessionId = crypto.randomUUID();
const history = [];
const samples = [];
let failure;

try {
  for (const scenario of [
    { question: '有go相关的内容吗', momentId: 2, topic: 'Go' },
    { question: '有和java相关的内容吗', momentId: 1, topic: 'Java' },
    { question: '有Python相关的内容吗', momentId: null, topic: 'Python' }
  ]) {
    const started = Date.now();
    const response = await fetch(new URL('/api/v2/public/ask', entry), {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question: scenario.question, sessionId, history }),
      signal: AbortSignal.timeout(95000)
    });
    const answer = (await response.json()).data;
    const sample = {
      question: scenario.question, durationMs: Date.now() - started, httpStatus: response.status,
      status: answer?.status, mode: answer?.mode, answer: answer?.answer,
      citationMomentIds: answer?.citations?.map((citation) => citation.momentId) ?? [],
      historyMessages: history.length, passed: false,
      externalFailure: response.status !== 200 ? 'network_or_environment' :
        answer?.status === 'temporarily_unavailable' ? 'model_or_environment' : null
    };
    samples.push(sample);
    assert.equal(response.status, 200);
    if (scenario.momentId !== null) {
      assert.equal(answer.status, 'answered', answer.reason);
      assert.equal(answer.mode, 'grounded');
      assert(answer.answer.toLowerCase().includes(scenario.topic.toLowerCase()));
      assert(answer.citations.some((citation) => citation.momentId === scenario.momentId));
      assert(answer.citations.every((citation) => citation.momentId === scenario.momentId), '切换话题不能沿用上一轮文档');
      history.push({ role: 'user', content: scenario.question }, { role: 'assistant', content: answer.answer });
    } else {
      assert.equal(answer.status, 'no_evidence');
      assert.equal(answer.citations.length, 0);
    }
    sample.passed = true;
  }
} catch (error) {
  failure = error.message;
} finally {
  const externalFailures = samples.filter((sample) => sample.externalFailure).length;
  const report = {
    testType: '定向测试', entry,
    sampleFunnel: { original: samples.length, excluded: 0, valid: samples.length },
    externalFailures: { total: externalFailures, rate: samples.length ? externalFailures / samples.length : 0 },
    coreFunction: { denominator: samples.length - externalFailures, passed: samples.filter((sample) => sample.passed).length },
    samples, failure: failure ?? null,
    limitation: '仅验证当前公开文章的文档查找与话题切换，不作为检索效果评测'
  };
  await writeFile(fileURLToPath(new URL('./rag-discovery-results_20260928.json', import.meta.url)), `${JSON.stringify(report, null, 2)}\n`, 'utf8');
  process.stdout.write(JSON.stringify(report, null, 2));
  if (failure) process.exitCode = 1;
}
