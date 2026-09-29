import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { readFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { dockerOutput, sqlAt, testTarget } from './rag-test-target.mjs';

const [action, dataset] = process.argv.slice(2);
assert(['start', 'stop', 'status'].includes(action), '使用 start、stop 或 status');
const target = testTarget(dataset, true);
const root = resolve('.');
const serverEnv = await readFile(resolve(root, 'server/.env'), 'utf8');
const gptLine = serverEnv.split(/\r?\n/).find(line => /^RAG_CHAT_GPT_BASE_URL\s*=/.test(line));
const gptURL = new URL(gptLine ? gptLine.slice(gptLine.indexOf('=') + 1).trim().replace(/^(["'])(.*)\1$/, '$2')
  : 'http://127.0.0.1:8317/v1');
if (['127.0.0.1', 'localhost', '[::1]'].includes(gptURL.hostname)) gptURL.hostname = 'host.docker.internal';
const composeEnv = { ...process.env, RAG_TEST_PROJECT: target.project, RAG_TEST_PORT: String(target.port),
  RAG_TEST_GPT_BASE_URL: gptURL.toString().replace(/\/$/, '') };
const composeArgs = ['--env-file', 'deploy/.env', '-f', 'deploy/docker-compose.rag-test.yml', '-p', target.project];
const quote = value => "'" + String(value).replaceAll("'", "''") + "'";
const sleep = ms => new Promise(ok => setTimeout(ok, ms));

async function compose(args) {
  const child = spawn('docker-compose', [...composeArgs, ...args], { cwd: root, env: composeEnv,
    stdio: 'inherit', windowsHide: true });
  const code = await new Promise((ok, fail) => { child.on('error', fail); child.on('exit', ok); });
  assert.equal(code, 0, '测试环境 Compose 操作成功');
}

if (action === 'stop') {
  await compose(['stop']);
} else if (action === 'status') {
  await compose(['ps']);
} else {
  await compose(['up', '-d', '--wait', 'postgres', 'redis']);
  const initialized = sqlAt(target.databaseContainer, "SELECT to_json(to_regclass('public.moment') IS NOT NULL);");
  if (!initialized) {
    // Only structure and non-sensitive RAG settings are copied; no daily content or login identities.
    const schema = dockerOutput(['exec', 'shawn-blog-postgres', 'sh', '-lc',
      'exec pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --schema-only --no-owner --no-acl']);
    const settings = sqlAt('shawn-blog-postgres', `SELECT COALESCE(jsonb_agg(jsonb_build_object('config_key',config_key,'value',value,'value_type',value_type)),'[]'::jsonb)
      FROM sys_config WHERE config_key LIKE 'rag.%' AND is_sensitive=false;`);
    const history = sqlAt('shawn-blog-postgres', 'SELECT jsonb_agg(jsonb_build_object(\'version_id\',version_id,\'is_applied\',is_applied) ORDER BY id) FROM goose_db_version;');
    sqlAt(target.databaseContainer, `BEGIN;\n${schema}
      SET LOCAL search_path=public;
      INSERT INTO sys_config(config_key,value,is_sensitive,group_path,value_type)
        SELECT config_key,value,false,'rag',value_type FROM jsonb_to_recordset(${quote(JSON.stringify(settings))}::jsonb) AS s(config_key text,value text,value_type text);
      INSERT INTO goose_db_version(version_id,is_applied)
        SELECT version_id,is_applied FROM jsonb_to_recordset(${quote(JSON.stringify(history))}::jsonb) AS h(version_id bigint,is_applied boolean);
      INSERT INTO app_user(username,nickname,is_active,is_admin) VALUES('rag-benchmark-source','评测语料来源',false,false);
      INSERT INTO sys_config(config_key,value,is_sensitive,group_path,value_type)
        VALUES('test.rag.dataset',${quote(dataset)},false,'test','string');
      COMMIT;`);
    console.log('专用数据库已初始化：保留测试结构与检索配置，使用无登录权限的语料作者');
  } else {
    assert.equal(sqlAt(target.databaseContainer, "SELECT to_json(value) FROM sys_config WHERE config_key='test.rag.dataset';"), dataset,
      '复用同一集合的专用数据库');
    console.log('复用专用数据库与现有索引');
  }
  await compose(['up', '-d', '--force-recreate', '--no-deps', 'server']);
  const deadline = Date.now() + 180000;
  const statusURL = target.endpoint.replace(/\/ask$/, '/rag/status');
  for (;;) {
    try {
      const payload = await (await fetch(statusURL, { signal: AbortSignal.timeout(3000) })).json();
      if (payload.code === 0 && payload.data?.available) break;
    } catch { /* The actual backend may still be compiling. */ }
    assert(Date.now() < deadline, '专用后端启动超时，查看对应容器日志');
    await sleep(2000);
  }
  console.log(`专用测试环境已启动：${target.project}，问答入口 ${target.endpoint}`);
}
