import { createHash } from 'node:crypto';
import { spawn } from 'node:child_process';
import { readFile, writeFile, mkdir, access } from 'node:fs/promises';
import { resolve, join } from 'node:path';

const base = resolve('Temp/rag-benchmarks');
const hash = value => createHash('sha256').update(value).digest('hex');
const ordered = entries => entries.sort((a, b) => hash(String(a[0])).localeCompare(hash(String(b[0]))));
const json = async path => JSON.parse(await readFile(path, 'utf8'));
const save = async (path, value) => writeFile(path, JSON.stringify(value, null, 2));
const lines = async path => (await readFile(path, 'utf8')).trim().split(/\r?\n/).map(JSON.parse);
const openRelease = '63f6b052ff83508b08e242db42263ee708815c26';
const beirArchiveHash = '536e14446a0ba56ed1398ab1055f39fe852686ecad24a6306c80c490fa8e0165';

async function download(url, path) {
  await new Promise((ok, fail) => {
    const child = spawn('curl.exe', ['--fail', '-L', '--http1.1', '--connect-timeout', '10', '--max-time', '90',
      '--retry', '2', '--retry-all-errors', '-sS', url, '-o', path], { windowsHide: true, stdio: ['ignore', 'ignore', 'pipe'] });
    let errors = '';
    child.stderr.on('data', data => { errors += data; });
    child.on('error', fail);
    child.on('exit', code => code === 0 ? ok() : fail(new Error(`dataset_download_failed:${code}:${errors.slice(0, 200)}`)));
  });
}

async function ensureSources() {
  const open = join(base, 'open-rag-bench');
  await mkdir(join(open, 'raw'), { recursive: true });
  for (const [path, url] of [
    [join(open, 'repository.json'), `https://huggingface.co/api/datasets/vectara/open_ragbench/revision/${openRelease}`],
    ...['queries', 'qrels', 'answers'].map(name => [join(open, `raw/${name}.json`),
      `https://huggingface.co/datasets/vectara/open_ragbench/raw/${openRelease}/pdf/arxiv/${name}.json`]),
  ]) {
    try { await json(path); } catch { await download(url, path); }
  }
  if ((await json(join(open, 'repository.json'))).sha !== openRelease) throw new Error('dataset_release_mismatch');
  const beir = join(base, 'beir-scifact/raw'), archive = join(beir, 'scifact.zip');
  await mkdir(beir, { recursive: true });
  try { await access(archive); } catch {
    await download('https://public.ukp.informatik.tu-darmstadt.de/thakur/BEIR/datasets/scifact.zip', archive);
  }
  if (hash(await readFile(archive)) !== beirArchiveHash) throw new Error('dataset_archive_hash_mismatch');
  try { await access(join(beir, 'scifact/corpus.jsonl')); } catch {
    const quote = value => "'" + value.replaceAll("'", "''") + "'";
    await new Promise((ok, fail) => {
      const child = spawn('powershell.exe', ['-NoProfile', '-NonInteractive', '-Command',
        `Expand-Archive -LiteralPath ${quote(archive)} -DestinationPath ${quote(beir)}`],
      { windowsHide: true, stdio: 'ignore' });
      child.on('error', fail);
      child.on('exit', code => code === 0 ? ok() : fail(new Error('dataset_archive_extract_failed')));
    });
  }
}

async function prepareOpen() {
  const dir = join(base, 'open-rag-bench');
  const repository = await json(join(dir, 'repository.json'));
  const queries = await json(join(dir, 'raw/queries.json'));
  const qrels = await json(join(dir, 'raw/qrels.json'));
  const answers = await json(join(dir, 'raw/answers.json'));
  const pool = ordered(Object.entries(queries).filter(([, value]) => value.source === 'text'));
  const selected = [], selectedDocs = new Set();
  for (const type of ['extractive', 'abstractive']) {
    let count = 0;
    for (const [id, item] of pool) {
      if (item.type !== type || selectedDocs.has(qrels[id].doc_id)) continue;
      selected.push({ id, question: item.query, type, gold: qrels[id], reference: answers[id] });
      selectedDocs.add(qrels[id].doc_id);
      if (++count === 12) break;
    }
  }
  const allGoldDocs = new Set(Object.values(qrels).map(value => value.doc_id));
  const corpusFiles = repository.siblings.filter(file => file.rfilename.includes('/corpus/'));
  const negatives = ordered(corpusFiles.map(file => [file.rfilename.split('/').pop().replace('.json', ''), file]))
    .filter(([id]) => !allGoldDocs.has(id)).slice(0, 24).map(([id]) => id);
  const docIds = [...selectedDocs, ...negatives].sort();
  await mkdir(join(dir, 'raw/corpus'), { recursive: true });
  await mkdir(join(dir, 'processed/markdown'), { recursive: true });
  let next = 0;
  await Promise.all(Array.from({ length: 4 }, async () => {
    while (next < docIds.length) {
      const id = docIds[next++];
      const path = join(dir, `raw/corpus/${id}.json`);
      try { await json(path); } catch {
        await download(`https://huggingface.co/datasets/vectara/open_ragbench/raw/${repository.sha}/pdf/arxiv/corpus/${id}.json`, path);
      }
      console.log(`Open RAG 下载 ${next}/${docIds.length}`);
    }
  }));
  const documents = [];
  for (const id of docIds) {
    const rawPath = join(dir, `raw/corpus/${id}.json`);
    const source = await json(rawPath);
    let markdown = '', sections = [];
    for (const [index, section] of source.sections.entries()) {
      const heading = `## Section ${index}\n\n`;
      const start = Array.from(markdown).length + Array.from(heading).length;
      markdown += heading + section.text;
      sections.push({ id: `${id}:${index}`, index, start, end: Array.from(markdown).length });
      markdown += '\n\n';
    }
    await writeFile(join(dir, `processed/markdown/${id}.md`), markdown);
    documents.push({ id, title: source.title, content: markdown, sections,
      rawSha256: hash(await readFile(rawPath)), markdownSha256: hash(markdown), negative: negatives.includes(id) });
  }
  const samples = selected.map(sample => ({ ...sample, goldIds: [`${sample.gold.doc_id}:${sample.gold.section_id}`] }));
  const manifest = { name: 'Open RAG Benchmark', source: 'https://github.com/vectara/open-rag-bench',
    release: repository.sha, sourceDataset: 'https://huggingface.co/datasets/vectara/open_ragbench',
    originalQuestions: Object.keys(queries).length, modalityExcluded: Object.keys(queries).length - pool.length,
    eligibleQuestions: pool.length, executedQuestions: samples.length, corpusDocuments: documents.length,
    negativeDocuments: negatives.length, sampling: 'SHA-256 order; 12 extractive and 12 abstractive; distinct positive documents; 24 official negative documents',
    metadataSha256: Object.fromEntries(await Promise.all(['queries', 'qrels', 'answers'].map(async name =>
      [name, hash(await readFile(join(dir, `raw/${name}.json`)))]))),
    scope: 'Frozen Markdown text subset, not the full multimodal benchmark', preparedAt: new Date().toISOString() };
  await save(join(dir, 'manifest.json'), manifest);
  await save(join(dir, 'processed/documents.json'), documents);
  await save(join(dir, 'processed/evaluator-only.json'), samples);
  console.log(JSON.stringify(manifest));
}

async function prepareBeir() {
  const dir = join(base, 'beir-scifact'), raw = join(dir, 'raw/scifact');
  const corpus = await lines(join(raw, 'corpus.jsonl'));
  const queries = await lines(join(raw, 'queries.jsonl'));
  const qrels = new Map();
  for (const line of (await readFile(join(raw, 'qrels/test.tsv'), 'utf8')).trim().split(/\r?\n/).slice(1)) {
    const [qid, id, score] = line.split('\t');
    if (!qrels.has(qid)) qrels.set(qid, []);
    if (Number(score) > 0) qrels.get(qid).push(id);
  }
  const queryMap = new Map(queries.map(item => [item._id, item]));
  const selected = ordered([...qrels.entries()]).slice(0, 32);
  const goldDocs = new Set(selected.flatMap(([, ids]) => ids));
  const otherDocs = ordered(corpus.map(item => [item._id, item])).filter(([id]) => !goldDocs.has(id));
  const docIds = new Set([...goldDocs, ...otherDocs.slice(0, 100 - goldDocs.size).map(([id]) => id)]);
  await mkdir(join(dir, 'processed/markdown'), { recursive: true });
  const documents = corpus.filter(item => docIds.has(item._id)).map(item => ({
    id: item._id, title: item.title, content: item.text, markdownSha256: hash(item.text),
    sections: [{ id: item._id, start: 0, end: Array.from(item.text).length }], negative: !goldDocs.has(item._id),
  }));
  for (const document of documents) await writeFile(join(dir, `processed/markdown/${document.id}.md`), document.content);
  const samples = selected.map(([id, goldIds]) => ({ id, question: queryMap.get(id).text, goldIds,
    reference: null, type: 'scientific_claim', gold: { doc_ids: goldIds } }));
  const manifest = { name: 'BEIR SciFact', source: 'https://github.com/beir-cellar/beir',
    release: 'official scifact.zip / test split', archiveSha256: hash(await readFile(join(dir, 'raw/scifact.zip'))),
    originalCorpusDocuments: corpus.length, originalQuestions: qrels.size, executedQuestions: samples.length,
    corpusDocuments: documents.length, sampling: 'SHA-256 order: 32 test queries; all associated positive documents plus corpus distractors up to 100 documents',
    scope: 'Frozen corpus/query subset, not the official full-corpus BEIR score; no generated gold answers', preparedAt: new Date().toISOString() };
  await save(join(dir, 'manifest.json'), manifest);
  await save(join(dir, 'processed/documents.json'), documents);
  await save(join(dir, 'processed/evaluator-only.json'), samples);
  console.log(JSON.stringify(manifest));
}

await ensureSources();
await prepareBeir();
await prepareOpen();
