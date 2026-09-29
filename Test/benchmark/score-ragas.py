import asyncio
import contextvars
import hashlib
import importlib.metadata
import json
import logging
import math
import os
import sys
from datetime import datetime, timezone
from pathlib import Path

os.environ['RAGAS_DO_NOT_TRACK'] = 'true'
from dotenv import dotenv_values
from openai import AsyncOpenAI, DefaultAsyncHttpxClient
from ragas.llms import llm_factory
from ragas.embeddings.base import embedding_factory
from ragas.metrics.collections import Faithfulness, AnswerAccuracy, AnswerRelevancy

logging.getLogger('instructor').setLevel(logging.CRITICAL)
logging.getLogger('ragas').setLevel(logging.CRITICAL)


async def main():
    root = Path.cwd()
    dataset_dir = root / 'Temp' / 'rag-benchmarks' / sys.argv[1]
    run_dir = dataset_dir / 'runs' / (dataset_dir / 'latest-run.txt').read_text().strip()
    resume = '--resume' in sys.argv[2:]
    score_name = (run_dir / 'latest-ragas.txt').read_text(encoding='utf-8').strip() if resume else datetime.now(timezone.utc).strftime('%Y%m%dT%H%M%SZ')
    score_dir = run_dir / 'ragas' / score_name
    score_dir.mkdir(parents=True, exist_ok=resume)
    (run_dir / 'latest-ragas.txt').write_text(score_dir.name, encoding='utf-8')
    env = dotenv_values(root / 'server' / '.env')
    judge_model = env.get('RAG_CHAT_GEMINI_MODEL', 'gemini-3.1-flash-lite-preview')
    judge_lock, last_dispatch = asyncio.Lock(), 0.0
    failed_requests = contextvars.ContextVar('failed_judge_requests', default=None)

    async def pace_judge_request(request):
        nonlocal last_dispatch
        async with judge_lock:
            now = asyncio.get_running_loop().time()
            await asyncio.sleep(max(0, 12 - (now - last_dispatch)))
            last_dispatch = asyncio.get_running_loop().time()

    async def observe_judge_response(response):
        failures = failed_requests.get()
        if response.is_error and failures is not None:
            failures.append(response.status_code)

    judge = AsyncOpenAI(api_key=env['RAG_CHAT_GEMINI_API_KEY'],
                        base_url=env['RAG_CHAT_GEMINI_BASE_URL'], timeout=40, max_retries=0,
                        http_client=DefaultAsyncHttpxClient(event_hooks={'request': [pace_judge_request], 'response': [observe_judge_response]}))
    embed_client = AsyncOpenAI(api_key=env['RAG_EMBEDDING_API_KEY'],
                               base_url=env['RAG_EMBEDDING_BASE_URL'], timeout=30, max_retries=0)
    llm = llm_factory(judge_model, client=judge, temperature=0, max_tokens=8192, max_retries=1)
    embeddings = embedding_factory('openai', model=env['RAG_EMBEDDING_MODEL'], client=embed_client)
    faithfulness = Faithfulness(llm=llm)
    accuracy = AnswerAccuracy(llm=llm, max_retries=1)
    relevancy = AnswerRelevancy(llm=llm, embeddings=embeddings)
    contract = {
        'testKind': '公开权威数据测试 / 对项目真实输出的独立评分',
        'ragasVersion': importlib.metadata.version('ragas'),
        'judgeModel': judge_model, 'judgeProvider': 'Gemini', 'judgeTemperature': 0,
        'judgeMaxTokens': 8192, 'judgeTopP': 0.1, 'instructorAttempts': 1,
        'judgeMinimumRequestIntervalSeconds': 12, 'metricDeadlineSeconds': 240,
        'metrics': ['faithfulness', 'answer_accuracy', 'answer_relevancy'],
        'answerAccuracyAttemptsPerJudge': 1,
        'answerAccuracyPolicy': 'Official question-aware dual-perspective metric; no custom rubric or modified reference answers',
        'artifactDir': str(score_dir),
        'embeddingModel': env['RAG_EMBEDDING_MODEL'], 'retryPolicy': 'no runner retries',
        'contextContract': 'All passages actually dispatched by the project; not just cited passages',
        'faithfulnessDenominator': 'Answered samples with contexts and successful judge calls; refusals are N/A',
        'accuracyDenominator': 'Core samples with official reference answers; unanswered samples score zero; judge failures separate',
        'relevancyDenominator': 'Core samples with successful judge calls; unanswered samples score zero',
        'referencePolicy': 'BEIR has no QA reference answers, so answer accuracy is N/A',
        'scriptSha256': hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
        'startedAt': datetime.now(timezone.utc).isoformat(),
    }
    if resume:
        previous = json.loads((score_dir / 'ragas-contract.json').read_text(encoding='utf-8'))
        for key in ['ragasVersion', 'judgeModel', 'judgeTemperature', 'judgeMaxTokens', 'judgeTopP', 'instructorAttempts', 'judgeMinimumRequestIntervalSeconds', 'embeddingModel', 'metrics', 'answerAccuracyAttemptsPerJudge']:
            if contract[key] != previous.get(key):
                raise ValueError('Scoring contract changed; start a separate scoring attempt')
        (score_dir / 'resume-contract.json').write_text(json.dumps(contract, ensure_ascii=False, indent=2), encoding='utf-8')
    else:
        (score_dir / 'ragas-contract.json').write_text(json.dumps(contract, ensure_ascii=False, indent=2), encoding='utf-8')
    result_file = score_dir / 'ragas-per-sample.jsonl'
    results = [json.loads(line) for line in result_file.read_text(encoding='utf-8').splitlines()] if resume and result_file.exists() else []
    processed = {result['id'] for result in results}
    while True:
        file = run_dir / 'per-sample.jsonl'
        samples = [json.loads(line) for line in file.read_text(encoding='utf-8').splitlines()] if file.exists() else []
        for sample in samples:
            if sample['id'] in processed:
                continue
            processed.add(sample['id'])
            result = {'id': sample['id'], 'scores': {}, 'judgeFailures': {}, 'skipped': {}}
            if sample['externalFailure']:
                result['skipped']['all'] = 'project external failure'
            else:
                answer = sample.get('answer') or {}
                response = answer.get('answer') or ''
                contexts = [f"{item['title']}\n{item['section']}\n{item['content']}"
                            for item in (sample.get('capture', {}).get('contexts') or [])]
                tasks = {}
                if answer.get('status') == 'answered' and response:
                    if contexts:
                        tasks['faithfulness'] = lambda: faithfulness.ascore(user_input=sample['question'], response=response, retrieved_contexts=contexts)
                    else:
                        result['skipped']['faithfulness'] = 'no retrieved contexts'
                    tasks['answer_relevancy'] = lambda: relevancy.ascore(user_input=sample['question'], response=response)
                    if sample.get('reference'):
                        tasks['answer_accuracy'] = lambda: accuracy.ascore(user_input=sample['question'], response=response, reference=sample['reference'])
                else:
                    result['scores']['answer_relevancy'] = 0
                    result['skipped']['faithfulness'] = 'unanswered sample'
                    if sample.get('reference'):
                        result['scores']['answer_accuracy'] = 0
                if not sample.get('reference'):
                    result['skipped']['answer_accuracy'] = 'dataset has no reference answer'
                async def score_metric(metric, invoke):
                    failures = []
                    token = failed_requests.set(failures)
                    try:
                        score = await asyncio.wait_for(invoke(), timeout=240)
                        if failures:
                            return metric, None, 'JudgeHTTPStatus:' + ','.join(map(str, failures))
                        value = float(score.value)
                        return metric, value if math.isfinite(value) else None, None if math.isfinite(value) else 'UndefinedMetric'
                    except Exception as error:
                        return metric, None, type(error).__name__
                    finally:
                        failed_requests.reset(token)
                for metric, value, error in await asyncio.gather(*(score_metric(metric, invoke) for metric, invoke in tasks.items())):
                    if error:
                        result['judgeFailures'][metric] = error
                    else:
                        result['scores'][metric] = value
            results.append(result)
            with (score_dir / 'ragas-per-sample.jsonl').open('a', encoding='utf-8') as output:
                output.write(json.dumps(result, ensure_ascii=False) + '\n')
            print(f"Ragas {sys.argv[1]} {len(results)} {result['scores']} failures={result['judgeFailures']}", flush=True)
        metadata = json.loads((run_dir / 'metadata.json').read_text(encoding='utf-8'))
        if metadata['status'] != 'running':
            final_samples = [json.loads(line) for line in file.read_text(encoding='utf-8').splitlines()] if file.exists() else []
            if all(sample['id'] in processed for sample in final_samples):
                break
        await asyncio.sleep(5)
    summary = {'samplesSeen': len(results), 'projectSamplesAvailable': len(final_samples), 'coverageComplete': len(processed) == len(final_samples),
               'metrics': {}, 'judgeModel': judge_model, 'ragasVersion': contract['ragasVersion']}
    for metric in contract['metrics']:
        values = [result['scores'][metric] for result in results if metric in result['scores']]
        failures = sum(metric in result['judgeFailures'] for result in results)
        summary['metrics'][metric] = {'scoredSamples': len(values), 'judgeFailures': failures,
                                      'mean': sum(values) / len(values) if values else None,
                                      'qualification': '受评判模型或环境影响，仅供参考' if failures / max(1, len(values) + failures) > .1 else 'conditional metric; see contract'}
    summary['artifactDir'] = str(score_dir)
    (score_dir / 'ragas-metrics.json').write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding='utf-8')
    (run_dir / 'ragas-metrics.json').write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding='utf-8')
    await judge.close()
    await embed_client.close()


asyncio.run(main())
