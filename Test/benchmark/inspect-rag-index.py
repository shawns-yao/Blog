"""Observe real worker output; never split source text or generate embeddings."""
import hashlib
import json
import subprocess
import sys
import time
from pathlib import Path

import tiktoken

root = Path.cwd()
name = sys.argv[1]
if name not in {'open-rag-bench', 'beir-scifact'}:
    raise ValueError('Unknown frozen dataset')
dataset = root / 'Temp' / 'rag-benchmarks' / name
run = dataset / 'directed-runs' / (dataset / 'latest-directed-run.txt').read_text().strip()
documents = {item['id']: item for item in json.loads((dataset / 'processed' / 'documents.json').read_text(encoding='utf-8'))}


def observe(query):
    result = subprocess.run(['docker', 'exec', '-i', 'shawn-blog-postgres', 'sh', '-lc',
                             'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -X -v ON_ERROR_STOP=1 -A -t -q'],
                            input=query, encoding='utf-8', capture_output=True, timeout=45, check=True)
    return json.loads(result.stdout)


if '--wait' in sys.argv[2:]:
    deadline = time.monotonic() + 1500
    while observe(f"""SELECT count(*) FROM moment m JOIN rag_index_state s ON s.moment_id=m.id
        WHERE m.ext_info->>'ragBenchmark'='{name}' AND m.is_published=true AND m.deleted_at IS NULL
          AND s.status='ready' AND s.active_profile=s.desired_profile AND s.active_hash=s.source_hash;""") < len(documents):
        if time.monotonic() > deadline:
            raise TimeoutError('Actual project indexing did not finish')
        time.sleep(5)

query = f"""SELECT COALESCE(json_agg(x),'[]'::json) FROM (
    SELECT m.ext_info->>'corpusDocumentId' AS document_id,c.id,c.kind,c.start_at,c.end_at,
      c.content,c.context_header,cardinality(c.embedding) AS dimensions
    FROM rag_chunk c JOIN moment m ON m.id=c.moment_id JOIN rag_index_state s ON s.moment_id=m.id
    WHERE m.ext_info->>'ragBenchmark'='{name}' AND m.is_published=true AND m.deleted_at IS NULL
      AND s.status='ready' AND s.active_profile=s.desired_profile AND s.active_hash=s.source_hash
      AND c.profile=s.active_profile AND c.source_hash=s.active_hash
    ORDER BY m.id,c.seq
) x;"""
rows = observe(query)
encoder = tiktoken.get_encoding('cl100k_base')
metadata = json.loads((run / 'metadata.json').read_text(encoding='utf-8'))
maximum = int(metadata.get('temporarySettings', {}).get('rag.chunkMaxTokens', metadata['originalSettings']['rag.chunkMaxTokens']))
by_document = {}
for row in rows:
    by_document.setdefault(row['document_id'], []).append(row)
reports = []
for document_id, document in documents.items():
    content = document['content']
    coverage = bytearray(len(content))
    invalid = oversized = 0
    atomic = []
    token_counts = []
    for row in by_document.get(document_id, []):
        start, end = row['start_at'], row['end_at']
        if start < 0 or end > len(content) or end <= start or content[start:end] != row['content']:
            invalid += 1
            continue
        for index in range(start, end):
            coverage[index] = min(255, coverage[index] + 1)
        tokens = len(encoder.encode(row['context_header'] + '\n\n' + row['content'], disallowed_special=()))
        token_counts.append(tokens)
        limit = max(maximum, 4000) if row['kind'] in {'math', 'code', 'table', 'list'} else maximum
        oversized += tokens > limit
        if tokens > maximum:
            atomic.append({'kind': row['kind'], 'start': start, 'end': end, 'tokens': tokens})
    reports.append({'documentId': document_id, 'sourceSha256': hashlib.sha256(content.encode()).hexdigest(),
                    'chunks': len(token_counts), 'invalidPositions': invalid, 'overLimit': oversized,
                    'missingNonWhitespace': sum(not c.isspace() and coverage[i] == 0 for i, c in enumerate(content)),
                    'repeatedNonWhitespace': sum(not c.isspace() and coverage[i] > 1 for i, c in enumerate(content)),
                    'maximumTokens': max(token_counts, default=0), 'largeAtomicChunks': atomic})
report = {'testKind': '定向测试', 'entrypoint': 'Real project source queue and indexing worker; read-only output observation',
          'documentsExpected': len(documents), 'documentsIndexed': len(by_document), 'chunks': len(rows),
          'dimensions': sorted({row['dimensions'] for row in rows}), 'ordinaryMaximumTokens': maximum,
          'invalidPositions': sum(item['invalidPositions'] for item in reports),
          'overLimit': sum(item['overLimit'] for item in reports),
          'missingNonWhitespace': sum(item['missingNonWhitespace'] for item in reports),
          'repeatedNonWhitespace': sum(item['repeatedNonWhitespace'] for item in reports), 'documents': reports}
(run / 'index-integrity.json').write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding='utf-8')
print(json.dumps({key: value for key, value in report.items() if key != 'documents'}, ensure_ascii=False), flush=True)
if len(by_document) != len(documents) or report['invalidPositions'] or report['overLimit'] or report['missingNonWhitespace']:
    raise SystemExit(1)
