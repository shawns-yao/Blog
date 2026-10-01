import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';

export function testTarget(dataset, dedicated = false) {
  const targets = {
    'beir-scifact': { suffix: 'scifact', port: 18080 },
    'open-rag-bench': { suffix: 'open-rag', port: 18081 },
    'bagu-obsidian': { suffix: 'bagu', port: 18082 },
  };
  assert(Object.hasOwn(targets, dataset), '指定已有独立数据集合');
  const { suffix, port } = targets[dataset];
  const project = `grtblog-rag-test-${suffix}`;
  return dedicated ? { dedicated, project, port, databaseContainer: `${project}-postgres`,
    serverContainer: `${project}-server`, endpoint: `http://127.0.0.1:${port}/api/v2/public/ask` }
    : { dedicated, databaseContainer: 'shawn-blog-postgres', serverContainer: 'shawn-blog-server-dev',
      endpoint: 'http://127.0.0.1:8080/api/v2/public/ask' };
}

export function dockerOutput(args, input) {
  const result = spawnSync('docker', args, { input, encoding: 'utf8', windowsHide: true,
    timeout: 45000, maxBuffer: 32 * 1024 * 1024 });
  const detail = result.stderr?.split(/\r?\n/).filter(line => /^ERROR:|^pg_dump: error:/.test(line)).join('\n');
  if (result.error || result.status !== 0) {
    const error = new Error(`Docker 操作失败：${args[0]} ${args[1] ?? ''}${detail ? `；${detail}` : ''}`);
    error.code = 'RAG_TEST_DOCKER_FAILED';
    throw error;
  }
  return result.stdout;
}

export function sqlAt(container, query) {
  const output = dockerOutput(['exec', '-i', container, 'sh', '-lc',
    'exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -X -v ON_ERROR_STOP=1 -A -t -q'], query).trim();
  return output ? JSON.parse(output) : null;
}

export function verifyTestContainers(target) {
  assert(target.dedicated, '只检查专用测试容器');
  for (const container of [target.databaseContainer, target.serverContainer]) {
    const [info] = JSON.parse(dockerOutput(['inspect', container]));
    assert.equal(info.Config.Labels['com.docker.compose.project'], target.project, '测试项目隔离标识');
    assert.equal(info.Config.Labels['io.shawn-blog.rag-test'], 'true', '测试容器隔离标识');
    assert(info.State.Running, '先启动专用测试容器');
  }
}
