import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { constants } from 'node:fs';
import { copyFile, mkdir, readFile, readdir, writeFile } from 'node:fs/promises';
import { basename, dirname, join, relative, resolve } from 'node:path';
import { dockerOutput, sqlAt, testTarget, verifyTestContainers } from './benchmark/rag-test-target.mjs';

const [action, sourceArgument] = process.argv.slice(2);
assert(['prepare', 'schema', 'import', 'status', 'wait', 'verify', 'smoke'].includes(action), '使用 prepare、schema、import、status、wait、verify 或 smoke');
const target = testTarget('bagu-obsidian', true);
const batch = 'bagu-obsidian_20260930_v1';
const rawDir = resolve('Data/raw', batch);
const manifestPath = resolve('Data/manifest', `${batch}.json`);
const reportPath = resolve('Data/manifest', `${batch}_result.json`);
const folders = ['01-Java', '02-Spring框架', '03-MySQL', '04-Redis', '05-消息队列与搜索', '06-计算机基础', '07-AI与Agent', '08-RAG与MCP'];
const hash = value => createHash('sha256').update(value).digest('hex');
const quote = value => "'" + String(value).replaceAll("'", "''") + "'";
const sql = query => sqlAt(target.databaseContainer, query);
const save = async (path, value) => {
  await mkdir(dirname(path), { recursive: true });
  await writeFile(path, JSON.stringify(value, null, 2) + '\n', 'utf8');
};
const readJSON = async path => JSON.parse(await readFile(path, 'utf8'));

async function markdownFiles(directory) {
  const result = [];
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) result.push(...await markdownFiles(path));
    else if (entry.isFile() && /\.md$/i.test(entry.name)) result.push(path);
  }
  return result;
}

async function prepare() {
  assert(sourceArgument, 'prepare 需要原始八股目录');
  const sourceRoot = resolve(sourceArgument);
  const files = (await Promise.all(folders.map(folder => markdownFiles(join(sourceRoot, folder))))).flat().sort();
  assert.equal(files.length, 429, '确认范围为 01–08 的 429 篇，源文件数量变化时停止');
  let existing;
  try { existing = await readJSON(manifestPath); } catch (error) { if (error.code !== 'ENOENT') throw error; }
  const documents = [];
  const decoder = new TextDecoder('utf-8', { fatal: true, ignoreBOM: true });
  for (const path of files) {
    const bytes = await readFile(path);
    const content = decoder.decode(bytes);
    assert(Buffer.from(content, 'utf8').equals(bytes), '正文必须能按原始 UTF-8 字节无损保存');
    assert(content.trim(), '不跳过空文档，需明确处理');
    const sourcePath = relative(sourceRoot, path).replaceAll('\\', '/');
    const sha256 = hash(bytes);
    const rawPath = join(rawDir, 'documents', sourcePath);
    if (existing) {
      assert(existing.documents.some(doc => doc.sourcePath === sourcePath && doc.sha256 === sha256), '冻结批次的源文档已变化');
      assert.equal(hash(await readFile(rawPath)), sha256, '原始副本指纹不一致');
    } else {
      await mkdir(dirname(rawPath), { recursive: true });
      try { await copyFile(path, rawPath, constants.COPYFILE_EXCL); }
      catch (error) { if (error.code !== 'EEXIST') throw error; }
      assert.equal(hash(await readFile(rawPath)), sha256, '原始副本指纹不一致');
    }
    const heading = content.match(/^#\s+(.+)$/m)?.[1]?.trim();
    documents.push({ sourcePath, folder: sourcePath.split('/')[0], title: Array.from(heading || basename(path, '.md')).slice(0, 255).join(''),
      bytes: bytes.length, characters: Array.from(content).length, sha256,
      rawPath: relative(resolve('.'), rawPath).replaceAll('\\', '/'), shortUrl: `bagu-${hash(sourcePath).slice(0, 32)}` });
  }
  const manifest = existing ?? { batch, preparedAt: new Date().toISOString(), sourceRoot, folders,
    expectedDocuments: 429, sourceFormat: 'UTF-8 Markdown, byte-preserving snapshot', documents,
    target, lifecycle: { source: '独立知识库原始文档，保留来源副本', index: '项目工作进程生成的可重建派生向量',
      visibility: '仅绑定本机的独立 Docker 环境，不进入日常博客', deletion: '本次不删除、不撤回来源或既有索引' } };
  if (!existing) await save(manifestPath, manifest);
  console.log(JSON.stringify({ prepared: documents.length, folders: Object.fromEntries(folders.map(folder => [folder, documents.filter(doc => doc.folder === folder).length])),
    bytes: documents.reduce((total, doc) => total + doc.bytes, 0), manifest: manifestPath }, null, 2));
}

function schemaSnapshot() {
  return sql(`SELECT jsonb_build_object(
    'columns',(SELECT jsonb_agg(jsonb_build_object('table',table_name,'column',column_name,'type',data_type,'nullable',is_nullable,'default',column_default) ORDER BY table_name,ordinal_position)
      FROM information_schema.columns WHERE table_schema='public' AND table_name IN ('moment','app_user','rag_index_state','rag_chunk')),
    'constraints',(SELECT jsonb_agg(jsonb_build_object('table',conrelid::regclass::text,'name',conname,'definition',pg_get_constraintdef(oid)))
      FROM pg_constraint WHERE conrelid IN ('moment'::regclass,'app_user'::regclass,'rag_index_state'::regclass,'rag_chunk'::regclass)),
    'indexes',(SELECT jsonb_agg(jsonb_build_object('table',tablename,'definition',indexdef)) FROM pg_indexes
      WHERE schemaname='public' AND tablename IN ('moment','app_user','rag_index_state','rag_chunk')),
    'triggers',(SELECT jsonb_agg(jsonb_build_object('table',tgrelid::regclass::text,'definition',pg_get_triggerdef(oid))) FROM pg_trigger
      WHERE NOT tgisinternal AND tgrelid IN ('moment'::regclass,'app_user'::regclass,'rag_index_state'::regclass,'rag_chunk'::regclass)),
    'migrations',(SELECT jsonb_agg(jsonb_build_object('version',version_id,'applied',is_applied) ORDER BY id) FROM goose_db_version));`);
}

function dailySnapshot() {
  return sqlAt('shawn-blog-postgres', `SELECT COALESCE(jsonb_agg(jsonb_build_object('id',id,'hash',content_hash,'published',is_published,'deleted',deleted_at) ORDER BY id),'[]'::jsonb) FROM moment;`);
}

function indexSnapshot() {
  return sql(`SELECT COALESCE(jsonb_agg(d ORDER BY d.id),'[]'::jsonb) FROM (
    SELECT m.id,m.ext_info->>'sourcePath' AS source_path,s.status,s.attempts,s.last_error,s.active_profile,s.desired_profile,
      s.active_hash=s.source_hash AS hash_current,count(c.id) AS chunks,
      count(c.id) FILTER (WHERE cardinality(c.embedding)<>4096 OR array_position(c.embedding,NULL) IS NOT NULL
        OR c.embedding && ARRAY['NaN'::float8,'Infinity'::float8,'-Infinity'::float8]) AS invalid_vectors
    FROM moment m LEFT JOIN rag_index_state s ON s.moment_id=m.id LEFT JOIN rag_chunk c ON c.moment_id=m.id
    WHERE m.ext_info->>'ragImport'='bagu-obsidian' GROUP BY m.id,s.moment_id) d;`);
}

async function reportIndex() {
  const documents = indexSnapshot();
  const counts = {};
  for (const doc of documents) counts[doc.status ?? 'missing'] = (counts[doc.status ?? 'missing'] ?? 0) + 1;
  let report;
  try { report = await readJSON(reportPath); } catch (error) { if (error.code !== 'ENOENT') throw error; report = { batch, target }; }
  report.indexCheckedAt = new Date().toISOString();
  report.indexCounts = counts;
  report.indexSnapshot = documents;
  await save(reportPath, report);
  console.log(JSON.stringify({ at: report.indexCheckedAt, documents: documents.length, counts,
    chunks: documents.reduce((total, doc) => total + doc.chunks, 0),
    failures: documents.filter(doc => doc.status === 'failed').map(doc => ({ id: doc.id, attempts: doc.attempts, error: doc.last_error })) }));
  return documents;
}

if (action === 'prepare') {
  await prepare();
} else {
  verifyTestContainers(target);
  assert.equal(sql("SELECT to_json(value) FROM sys_config WHERE config_key='test.rag.dataset';"), 'bagu-obsidian', '只能写入已确认的独立八股数据库');
  if (action === 'schema') {
    const schema = schemaSnapshot();
    await save(resolve('Data/manifest', `${batch}_schema.json`), schema);
    console.log(JSON.stringify(schema, null, 2));
  } else {
    const manifest = await readJSON(manifestPath);
    assert.equal(manifest.documents.length, 429);
    if (action === 'import') {
      const input = await Promise.all(manifest.documents.map(async doc => {
        const bytes = await readFile(resolve(doc.rawPath));
        assert.equal(hash(bytes), doc.sha256, '冻结原文指纹不能变化');
        return { ...doc, content: bytes.toString('utf8') };
      }));
      let report;
      try { report = await readJSON(reportPath); } catch (error) { if (error.code !== 'ENOENT') throw error; }
      report ??= { batch, target, startedAt: new Date().toISOString(), dailyBefore: dailySnapshot() };
      await save(reportPath, report);
      const rows = sql(`BEGIN;
        CREATE TEMP TABLE bagu_input ON COMMIT DROP AS
          SELECT * FROM jsonb_to_recordset(${quote(JSON.stringify(input))}::jsonb)
          AS d("sourcePath" text,title text,content text,sha256 text,"shortUrl" text,folder text);
        INSERT INTO moment(title,summary,content,content_hash,author_id,toc,short_url,is_published,is_original,ext_info,content_updated_at)
          SELECT d.title,'',d.content,md5(d.content),(SELECT id FROM app_user WHERE username='rag-benchmark-source' AND is_active=false AND is_admin=false),
            '[]'::jsonb,d."shortUrl",true,false,jsonb_build_object('contentKind','note','ragImport','bagu-obsidian',
              'sourcePath',d."sourcePath",'sourceSha256',d.sha256,'sourceFolder',d.folder),now()
          FROM bagu_input d ON CONFLICT(short_url) DO NOTHING;
        DO $$ BEGIN IF EXISTS (SELECT 1 FROM bagu_input d LEFT JOIN moment m ON m.short_url=d."shortUrl"
          WHERE m.id IS NULL OR m.content IS DISTINCT FROM d.content OR m.title IS DISTINCT FROM d.title
            OR m.ext_info->>'ragImport' IS DISTINCT FROM 'bagu-obsidian' OR m.ext_info->>'sourcePath' IS DISTINCT FROM d."sourcePath"
            OR m.ext_info->>'sourceSha256' IS DISTINCT FROM d.sha256 OR m.is_published IS DISTINCT FROM true OR m.deleted_at IS NOT NULL)
          THEN RAISE EXCEPTION 'Frozen corpus conflict; no existing document is overwritten'; END IF; END $$;
        SELECT jsonb_agg(jsonb_build_object('sourcePath',d."sourcePath",'momentId',m.id) ORDER BY m.id)
          FROM bagu_input d JOIN moment m ON m.short_url=d."shortUrl";
        COMMIT;`);
      assert.equal(rows.length, 429, '所有文档必须写入');
      report.importedAt = new Date().toISOString();
      report.sourceMap = rows;
      await save(reportPath, report);
      console.log(`已批量写入 ${rows.length} 篇原文，向量由项目真实工作进程生成。`);
      await reportIndex();
    } else if (action === 'wait' || action === 'status') {
      const deadline = Date.now() + 180 * 60000;
      for (;;) {
        const docs = await reportIndex();
        if (action === 'status') break;
        assert.equal(docs.length, 429, '索引监控不得漏掉源文档');
        if (docs.every(doc => doc.status === 'ready' && doc.hash_current && doc.active_profile === doc.desired_profile && doc.chunks > 0 && doc.invalid_vectors === 0)) break;
        assert(!docs.some(doc => doc.status === 'failed' && doc.attempts >= 5), '存在耗尽重试的失败文档，保留来源及失败记录');
        assert(Date.now() < deadline, '索引等待超过三小时，保留当前进度');
        // 每次只聚合查询整个批次，避免按文档逐条查询。
        await new Promise(ok => setTimeout(ok, 30000));
      }
    } else if (action === 'smoke') {
      const report = await readJSON(reportPath);
      const samples = [
        { name: 'Java 参数传递', question: '依据知识库原文，Java 是值传递还是引用传递？为什么修改对象能影响调用方，而重新赋值形参不能？', title: 'Java 是值传递还是引用传递' },
        { name: 'RAG 与微调', question: '依据知识库原文，RAG 的工作原理是什么？与微调相比主要解决什么问题？', title: 'RAG 的工作原理' },
      ];
      const results = [];
      const status = await (await fetch(target.endpoint.replace(/\/ask$/, '/rag/status'), { signal: AbortSignal.timeout(5000) })).json();
      assert(status.code === 0 && status.data.available && status.data.indexReady, '独立知识库入口及索引必须可用');
      for (const sample of samples) {
        const expected = manifest.documents.find(doc => doc.title.includes(sample.title));
        assert(expected, '冒烟问题必须有真实来源');
        const sourceId = report.sourceMap.find(doc => doc.sourcePath === expected.sourcePath).momentId;
        const started = performance.now();
        const entry = { name: sample.name, question: sample.question, expectedSourceId: sourceId, passed: false, externalFailure: false };
        try {
          const response = await fetch(target.endpoint, { method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ question: sample.question, contentKind: 'note', sessionId: crypto.randomUUID(), history: [] }),
            signal: AbortSignal.timeout(125000) });
          if (response.status >= 500 || response.status === 408 || response.status === 429) { entry.externalFailure = true; throw new Error(`HTTP ${response.status}`); }
          assert.equal(response.status, 200);
          const payload = await response.json();
          assert.equal(payload.code, 0);
          entry.status = payload.data.status;
          entry.answer = payload.data.answer;
          entry.trace = payload.data.trace;
          entry.citedSourceIds = payload.data.citations.map(citation => citation.momentId);
          if (entry.status === 'temporarily_unavailable') { entry.externalFailure = true; throw new Error('项目最终模型不可用'); }
          assert.equal(entry.status, 'answered');
          assert.equal(payload.data.mode, 'grounded');
          assert(entry.answer.length && entry.citedSourceIds.includes(sourceId), '必须回答并引用原始问题文档');
          assert(entry.citedSourceIds.every(id => report.sourceMap.some(doc => doc.momentId === id)), '引用仅来自本次独立知识库');
          entry.passed = true;
        } catch (error) {
          entry.error = error.message;
          if (error.name === 'TimeoutError' || error.name === 'TypeError') entry.externalFailure = true;
        }
        entry.durationMs = Math.round(performance.now() - started);
        results.push(entry);
        console.log(JSON.stringify({ name: entry.name, passed: entry.passed, externalFailure: entry.externalFailure, durationMs: entry.durationMs }));
        if (performance.now() - started < 11000) await new Promise(ok => setTimeout(ok, 11000 - (performance.now() - started)));
      }
      const externalFailures = results.filter(entry => entry.externalFailure).length;
      const output = { testKind: '冒烟测试', testedAt: new Date().toISOString(), target,
        funnel: { raw: samples.length, excluded: 0, effective: samples.length, externalFailures, coreReturns: samples.length - externalFailures },
        passed: results.filter(entry => entry.passed).length, samples: results,
        scope: '项目公开入口实际检索、生成和引用检查，不代表全部文档问答准确率或公开集合验收' };
      await save(resolve('Test/rag-obsidian-smoke_20260930_v1.json'), output);
      assert(results.every(entry => entry.passed), '冒烟结果存在失败，保留真实记录');
    } else if (action === 'verify') {
      const rows = sql(`SELECT jsonb_agg(jsonb_build_object('id',m.id,'sourcePath',m.ext_info->>'sourcePath','content',m.content,
        'status',s.status,'hashCurrent',s.active_hash=s.source_hash,'profileCurrent',s.active_profile=s.desired_profile,
        'chunks',(SELECT jsonb_agg(jsonb_build_object('start',c.start_at,'end',c.end_at,'content',c.content,'dimensions',cardinality(c.embedding)) ORDER BY c.seq)
          FROM rag_chunk c WHERE c.moment_id=m.id)) ORDER BY m.id)
        FROM moment m JOIN rag_index_state s ON s.moment_id=m.id WHERE m.ext_info->>'ragImport'='bagu-obsidian';`);
      assert.equal(rows.length, 429);
      let chunks = 0, characters = 0;
      for (const doc of rows) {
        const source = manifest.documents.find(item => item.sourcePath === doc.sourcePath);
        assert(source, '来源必须在确认的冻结范围内');
        assert.equal(hash(Buffer.from(doc.content, 'utf8')), source.sha256, '数据库正文必须与原始文件完全一致');
        assert.equal(hash(await readFile(join(manifest.sourceRoot, source.sourcePath))), source.sha256, '原始 Obsidian 文件保持不变');
        assert.equal(doc.status, 'ready');
        assert(doc.hashCurrent && doc.profileCurrent && doc.chunks.length);
        const points = Array.from(doc.content);
        const covered = new Uint8Array(points.length);
        for (const chunk of doc.chunks) {
          assert.equal(chunk.dimensions, 4096);
          assert(chunk.start >= 0 && chunk.end > chunk.start && chunk.end <= points.length);
          assert.equal(points.slice(chunk.start, chunk.end).join(''), chunk.content, '分块位置必须对应真实原文');
          covered.fill(1, chunk.start, chunk.end);
          chunks++;
        }
        points.forEach((point, index) => { if (!/\s/u.test(point)) { assert(covered[index], `原文覆盖缺失：${source.sourcePath}`); characters++; } });
      }
      const snapshot = await reportIndex();
      assert(snapshot.every(doc => doc.invalid_vectors === 0), '所有向量必须有效');
      const report = await readJSON(reportPath);
      const dailyAfter = dailySnapshot();
      assert.deepEqual(dailyAfter, report.dailyBefore, '日常博客正文及发布状态保持不变');
      report.verification = { testKind: '定向测试', verifiedAt: new Date().toISOString(), documents: rows.length, chunks,
        nonWhitespaceCharacters: characters, sourceCoverage: 1, dimensions: 4096, sourceFilesUnchanged: true, dailyContentUnchanged: true,
        scope: '实际工作进程生成的索引与原文一致性，不代表整体问答准确性或公开集合验收' };
      const [container] = JSON.parse(dockerOutput(['inspect', target.serverContainer]));
      const environment = Object.fromEntries(container.Config.Env.map(item => {
        const separator = item.indexOf('=');
        return [item.slice(0, separator), item.slice(separator + 1)];
      }));
      report.embedding = { model: environment.RAG_EMBEDDING_MODEL, dimensions: Number(environment.RAG_EMBEDDING_DIMENSIONS) };
      report.effectiveTuning = sql(`SELECT jsonb_object_agg(config_key,value) FROM sys_config WHERE config_key IN (
        'rag.chunkTargetTokens','rag.chunkMinTokens','rag.chunkMaxTokens','rag.chunkOverlapTokens','rag.parentMaxTokens','rag.adaptiveChunkingEnabled');`);
      report.overlap = sql(`SELECT jsonb_build_object('overlappingChunks',count(*) FILTER (WHERE previous_end>start_at),
        'overlapCharacters',COALESCE(sum(greatest(previous_end-start_at,0)),0))
        FROM (SELECT start_at,lag(end_at) OVER (PARTITION BY moment_id ORDER BY seq) AS previous_end FROM rag_chunk) x;`);
      await save(reportPath, report);
      console.log(JSON.stringify(report.verification, null, 2));
    }
  }
}
