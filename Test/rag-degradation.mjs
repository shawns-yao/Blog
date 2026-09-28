import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { writeFile } from "node:fs/promises";

const base = process.env.RAG_TEST_API || "http://127.0.0.1:18089/api/v2";
assert.equal(new URL(base).hostname, "127.0.0.1");
const report = {
  testType: "定向测试",
  dataType: "合成文档",
  cases: [],
  externalFailures: [],
  fault:
    "主通道与重排序地址均为本机关闭端口，真实连接失败；不替换核心模块或上游响应",
};
let token = "";
async function api(path, method = "GET", body) {
  const response = await fetch(`${base}${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    ...(body ? { body: JSON.stringify(body) } : {}),
    signal: AbortSignal.timeout(95000),
  });
  const value = await response.json();
  assert.equal(value.code, 0, value.msg);
  return value.data;
}
try {
  const account = `rag-degrade-${randomUUID().slice(0, 8)}`;
  const password = randomUUID();
  await api("/auth/register", "POST", {
    username: account,
    nickname: "降级验证",
    email: `${account}@example.invalid`,
    password,
  });
  token = (await api("/auth/login", "POST", { credential: account, password }))
    .token;
  const settings = await api("/admin/rag/settings");
  assert(
    settings.enabled &&
      settings.tuning.rerankEnabled &&
      settings.tuning.rerankFallback,
  );
  const source = await api("/moments/", "POST", {
    title: "云帆项目检查规范",
    content:
      "# 云帆项目检查规范\n\n云帆项目每21天检查一次备用电池。检查记录保存60天。",
    shortUrl: "rag-test-degradation",
    isPublished: true,
    extInfo: { contentKind: "note" },
  });
  const deadline = Date.now() + 150000;
  let ready = false;
  do {
    const list = await api("/admin/rag/documents");
    ready = list.items.some(
      (item) => item.momentId === source.id && item.status === "ready",
    );
    if (ready) break;
    await new Promise((resolve) => setTimeout(resolve, 2000));
  } while (Date.now() < deadline);
  assert(ready, "索引尚未就绪");
  const question = {
    question: "云帆项目多久检查一次备用电池？",
    sessionId: randomUUID(),
  };
  const answer = await api("/public/ask", "POST", question);
  assert.equal(answer.status, "answered");
  assert(answer.answer.includes("21"));
  let metrics = await api("/admin/rag/metrics");
  assert.equal(metrics.rerankDegraded, 1);
  assert.equal(metrics.fallbackRequests, 1);
  report.cases.push({
    name: "重排序失败回退至 RRF，官方兜底回答",
    passed: true,
  });
  process.stdout.write("重排序失败回退至 RRF: passed\n");
  await api("/admin/rag/settings", "PUT", {
    ...settings.tuning,
    rerankFallback: false,
  });
  const unavailable = await api("/public/ask", "POST", {
    ...question,
    sessionId: randomUUID(),
  });
  assert.equal(unavailable.status, "temporarily_unavailable");
  assert.equal(unavailable.citations.length, 0);
  metrics = await api("/admin/rag/metrics");
  assert.equal(metrics.requests, 2);
  assert.equal(metrics.unavailable, 1);
  assert.equal(metrics.fallbackRequests, 1);
  report.cases.push({
    name: "禁止重排序失败回退时暂停回答，无额外生成",
    passed: true,
  });
  process.stdout.write("禁止回退时暂停回答: passed\n");
} catch (error) {
  report.cases.push({
    name: "降级链路验证",
    passed: false,
    error: error.message,
  });
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
} finally {
  report.funnel = {
    original: 2,
    excluded: 0,
    effective: 2,
    externalFailures: report.externalFailures.length,
    coreReturned: report.cases.filter((item) => item.passed).length,
  };
  await writeFile(
    new URL("./rag-degradation_result_20260928.json", import.meta.url),
    JSON.stringify(report, null, 2) + "\n",
  );
}
