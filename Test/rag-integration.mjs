import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { createRequire } from "node:module";
import { fileURLToPath } from "node:url";
import { writeFile } from "node:fs/promises";

const require = createRequire(new URL("../web/package.json", import.meta.url));
const { chromium } = require("playwright");
const apiBase = process.env.RAG_TEST_API || "http://127.0.0.1:18089/api/v2";
const adminBase = process.env.RAG_TEST_ADMIN || "http://127.0.0.1:5879";
const webBase = process.env.RAG_TEST_WEB || "http://127.0.0.1:5179";
assert(["127.0.0.1", "localhost"].includes(new URL(apiBase).hostname));
const report = {
  testType: "定向测试",
  dataType: "合成文档",
  cases: [],
  externalFailures: [],
  limitations: [
    "不是公开权威数据测试，不提供检索或问答准确率",
    "主通道使用真实连接失败配置；只验证官方兜底，不调用 OpenCode Go",
  ],
};
let token = "";
const account = `rag-test-${randomUUID().slice(0, 8)}`;
const password = randomUUID();

async function request(
  path,
  { method = "GET", body, auth = true, allowError = false } = {},
) {
  const response = await fetch(`${apiBase}${path}`, {
    method,
    headers: {
      ...(body ? { "Content-Type": "application/json" } : {}),
      ...(auth && token ? { Authorization: `Bearer ${token}` } : {}),
    },
    ...(body ? { body: JSON.stringify(body) } : {}),
    signal: AbortSignal.timeout(95000),
  });
  const value = await response.json();
  if (allowError)
    return { httpStatus: response.status, code: value.code, data: value.data };
  assert.equal(value.code, 0, `${method} ${path}: ${value.msg}`);
  return value.data;
}

async function test(name, work) {
  const start = performance.now();
  try {
    const result = await work();
    report.cases.push({
      name,
      passed: true,
      durationMs: Math.round(performance.now() - start),
      ...result,
    });
    process.stdout.write(`${name}: passed\n`);
  } catch (error) {
    report.cases.push({
      name,
      passed: false,
      durationMs: Math.round(performance.now() - start),
      error: error.message,
    });
    throw error;
  }
}

async function waitReady(ids) {
  const deadline = Date.now() + 150000;
  let list;
  do {
    list = await request("/admin/rag/documents?pageSize=100");
    if (
      ids.every((id) =>
        list.items.some(
          (item) => item.momentId === id && item.status === "ready",
        ),
      )
    )
      return list;
    const failed = list.items.find(
      (item) => ids.includes(item.momentId) && item.status === "failed",
    );
    if (failed && failed.attempts >= 5) {
      if (failed.lastError === "embedding_unavailable")
        report.externalFailures.push({
          stage: "index",
          reason: failed.lastError,
        });
      throw new Error(`索引重试耗尽: ${failed.lastError}`);
    }
    await new Promise((resolve) => setTimeout(resolve, 2000));
  } while (Date.now() < deadline);
  throw new Error("索引等待超时");
}

async function ask(question) {
  const answer = await request("/public/ask", {
    method: "POST",
    auth: false,
    body: { question, sessionId: randomUUID() },
  });
  if (answer.status === "temporarily_unavailable") {
    report.externalFailures.push({ stage: "question", reason: answer.reason });
    throw new Error(`问答不可用: ${answer.reason}`);
  }
  return answer;
}

const articleBody =
  "# 星舟项目运行手册\n\n## 电池巡检\n\n星舟项目的电池巡检周期是每14天一次。剩余电量低于30%时，必须补充充电。巡检记录保留90天。\n\n## 部件清单\n\n| 部件 | 数量 |\n| --- | --- |\n| 备用电池 | 2 |\n| 温度探头 | 4 |\n\n## 记录命令\n\n```sh\necho inspection-ready\n```\n";
const noteBody =
  "# 星舟故障恢复\n\n## 恢复顺序\n\n星舟项目出现故障时，先断开充电线路，再等待10分钟，最后检查温度探头。恢复时由值班人员填写检修记录。";
let article;
let note;
let draft;
let savedTuning;
let browser;
let activePage;

try {
  await test("管理员认证与非敏感配置", async () => {
    const denied = await request("/admin/rag/settings", {
      auth: false,
      allowError: true,
    });
    assert.notEqual(denied.code, 0);
    await request("/auth/register", {
      method: "POST",
      auth: false,
      body: {
        username: account,
        nickname: "RAG 验证",
        email: `${account}@example.invalid`,
        password,
      },
    });
    const login = await request("/auth/login", {
      method: "POST",
      auth: false,
      body: { credential: account, password },
    });
    token = login.token;
    assert(token);
    const settings = await request("/admin/rag/settings");
    assert(
      settings.enabled &&
        settings.embeddingConfigured &&
        settings.fallbackConfigured,
    );
    assert(!/apiKey|https?:\/\//i.test(JSON.stringify(settings)));
    savedTuning = settings.tuning;
  });

  await test("参数拒绝与批量保存", async () => {
    const invalid = await request("/admin/rag/settings", {
      method: "PUT",
      allowError: true,
      body: { ...savedTuning, chunkOverlap: savedTuning.chunkSize },
    });
    assert.notEqual(invalid.code, 0);
    savedTuning = {
      ...savedTuning,
      chunkSize: 360,
      chunkOverlap: 40,
      vectorTopK: 8,
      keywordTopK: 8,
      topK: 4,
      rerankCandidateTopK: 16,
    };
    await request("/admin/rag/settings", { method: "PUT", body: savedTuning });
    assert.deepEqual(
      (await request("/admin/rag/settings")).tuning,
      savedTuning,
    );
  });

  await test("真实分块预览与原文偏移", async () => {
    const chunks = await request("/admin/rag/preview", {
      method: "POST",
      body: { title: "星舟项目运行手册", markdown: articleBody },
    });
    assert(chunks.length > 0);
    const runes = Array.from(articleBody);
    for (const chunk of chunks)
      assert.equal(runes.slice(chunk.start, chunk.end).join(""), chunk.content);
    assert(
      chunks.some((chunk) => chunk.content.includes("echo inspection-ready")),
    );
    assert(chunks.some((chunk) => chunk.content.includes("| 温度探头 | 4 |")));
    return { chunks: chunks.length };
  });

  await test("公开入库与草稿排除", async () => {
    article = await request("/moments/", {
      method: "POST",
      body: {
        title: "星舟项目运行手册",
        content: articleBody,
        summary: "电池巡检规范",
        shortUrl: "rag-test-inspection",
        isPublished: true,
        extInfo: { contentKind: "article" },
      },
    });
    note = await request("/moments/", {
      method: "POST",
      body: {
        title: "星舟故障恢复",
        content: noteBody,
        shortUrl: "rag-test-recovery",
        isPublished: true,
        extInfo: { contentKind: "note" },
      },
    });
    draft = await request("/moments/", {
      method: "POST",
      body: {
        title: "尚未发布的演练记录",
        content: "内部演练口令：未发布合成文档不得参与问答。",
        shortUrl: "rag-test-draft",
        isPublished: false,
        extInfo: { contentKind: "note" },
      },
    });
    const list = await waitReady([article.id, note.id]);
    assert.equal(
      list.items.find((item) => item.momentId === draft.id).status,
      "excluded",
    );
    const chunks = await request(`/admin/rag/documents/${article.id}/chunks`);
    assert(
      chunks.total > 0 &&
        chunks.items.length > 0 &&
        chunks.items.every((chunk) => !("embedding" in chunk)),
    );
    assert.equal(
      (await request(`/admin/rag/documents/${draft.id}/chunks`)).total,
      0,
    );
    const stats = await request("/admin/rag/index");
    assert.equal(stats.ready, 2);
    assert.equal(stats.embeddingDimension, 1024);
    return {
      readyDocuments: stats.ready,
      indexedChunks: stats.chunks,
      vectorDimension: stats.embeddingDimension,
    };
  });

  await test("真实混合检索重排序与官方兜底", async () => {
    const answer = await ask(
      "星舟项目的电池巡检周期是多少天，低于什么电量需要补充充电？",
    );
    assert.equal(answer.status, "answered");
    assert(answer.answer.includes("14") && answer.answer.includes("30"));
    assert(
      answer.citations.length > 0 &&
        answer.citations.length <= savedTuning.topK,
    );
    assert(
      answer.citations.some((citation) => citation.momentId === article.id),
    );
    assert(
      answer.citations.every((citation) => citation.momentId !== draft.id),
    );
    const metrics = await request("/admin/rag/metrics");
    assert.equal(metrics.answered, 1);
    assert.equal(metrics.primaryFailures, 1);
    assert.equal(metrics.fallbackRequests, 1);
    assert.equal(metrics.rerankDegraded, 0);
    assert(metrics.avgRerankMs !== null && metrics.avgGenerationMs !== null);
    return {
      citations: answer.citations.length,
      rerankDegraded: metrics.rerankDegraded,
      fallbackRequests: metrics.fallbackRequests,
    };
  });

  await test("无依据拒答", async () => {
    const answer = await ask("光谱海马项目在2026年的售票收入是多少？");
    assert.equal(answer.status, "no_evidence");
    assert.equal(answer.citations.length, 0);
  });

  await test("关键词权重与最终 TopK 生效", async () => {
    const tuning = {
      ...savedTuning,
      rrfVectorWeight: 0,
      rrfKeywordWeight: 1,
      topK: 1,
      rerankEnabled: false,
    };
    await request("/admin/rag/settings", { method: "PUT", body: tuning });
    const answer = await ask("星舟项目电池巡检周期是多少天？");
    assert.equal(answer.status, "answered");
    assert.equal(answer.citations.length, 1);
    assert(answer.answer.includes("14"));
    await request("/admin/rag/settings", { method: "PUT", body: savedTuning });
  });

  await test("单篇重建与撤回公开边界", async () => {
    await request(`/admin/rag/documents/${note.id}/reindex`, {
      method: "POST",
    });
    await waitReady([note.id]);
    await request("/admin/moments/published", {
      method: "PUT",
      body: { ids: [note.id], isPublished: false },
    });
    assert.equal(
      (await request(`/admin/rag/documents/${note.id}/chunks`)).total,
      0,
    );
    const denied = await request(`/admin/rag/documents/${note.id}/reindex`, {
      method: "POST",
      allowError: true,
    });
    assert.notEqual(denied.code, 0);
    await request("/admin/moments/published", {
      method: "PUT",
      body: { ids: [note.id], isPublished: true },
    });
    await waitReady([note.id]);
  });

  browser = await chromium.launch({ channel: "msedge", headless: true });
  await test("后台真实登录及桌面移动端页面", async () => {
    const context = await browser.newContext({
      viewport: { width: 1440, height: 900 },
      reducedMotion: "reduce",
    });
    const page = await context.newPage();
    activePage = page;
    const errors = [];
    page.on("pageerror", (error) => errors.push(error.message));
    await page.goto(`${adminBase}/sign-in?r=/rag`, {
      waitUntil: "domcontentloaded",
    });
    await page.getByPlaceholder("请输入账号或邮箱").fill(account);
    await page.getByPlaceholder("请输入密码", { exact: true }).fill(password);
    await page.getByRole("button", { name: /^登\s*录$/ }).click();
    await page
      .getByRole("heading", { name: /^RAG 知识库/ })
      .waitFor({ timeout: 60000 });
    await page
      .getByRole("cell", { name: "星舟项目运行手册", exact: true })
      .waitFor();
    await page.screenshot({
      path: fileURLToPath(
        new URL(
          "../Image/figures/rag-admin-desktop_20260928.png",
          import.meta.url,
        ),
      ),
      fullPage: true,
      animations: "disabled",
    });
    await page.getByText("检索配置", { exact: true }).click();
    const chunkInput = page
      .locator(".n-form-item")
      .filter({ hasText: "分块大小（字符）" })
      .locator("input");
    await chunkInput.fill("400");
    await page.getByText("运行指标", { exact: true }).click();
    await page.getByText("检索配置", { exact: true }).click();
    assert.equal(await chunkInput.inputValue(), "400");
    await page.getByRole("button", { name: "还原", exact: true }).click();
    await page.setViewportSize({ width: 390, height: 844 });
    await page.screenshot({
      path: fileURLToPath(
        new URL(
          "../Image/figures/rag-admin-settings-mobile_20260928.png",
          import.meta.url,
        ),
      ),
      fullPage: true,
      animations: "disabled",
    });
    assert.equal(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
      true,
    );
    await page.getByText("索引文档", { exact: true }).click();
    await page
      .getByRole("row")
      .filter({ hasText: "星舟项目运行手册" })
      .getByRole("button", { name: "分块", exact: true })
      .click();
    await page.getByRole("heading", { name: "分块 1", exact: true }).waitFor();
    await page.screenshot({
      path: fileURLToPath(
        new URL(
          "../Image/figures/rag-admin-drawer-mobile_20260928.png",
          import.meta.url,
        ),
      ),
      fullPage: true,
      animations: "disabled",
    });
    assert.equal(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
      true,
    );
    assert.deepEqual(errors, []);
    await context.close();
  });

  await test("前台真实问答及引用", async () => {
    const context = await browser.newContext({
      viewport: { width: 1440, height: 900 },
      reducedMotion: "reduce",
    });
    const page = await context.newPage();
    activePage = page;
    await page.goto(webBase, { waitUntil: "domcontentloaded", timeout: 95000 });
    await page
      .getByRole("button", { name: "打开站内问答", exact: true })
      .click();
    const dialog = page.getByRole("dialog", { name: "站内问答", exact: true });
    await dialog
      .getByRole("textbox", { name: "你的问题", exact: true })
      .fill("星舟项目的电池巡检周期是多少天？");
    await dialog.getByRole("button", { name: "发送问题", exact: true }).click();
    await dialog
      .getByRole("heading", { name: "原文依据", exact: true })
      .waitFor({ timeout: 95000 });
    const citation = dialog
      .getByRole("link")
      .filter({ hasText: "星舟项目运行手册" })
      .first();
    assert.match(
      await citation.getAttribute("href"),
      /\/moments\/\d{4}\/\d{2}\/\d{2}\/rag-test-inspection$/,
    );
    await page.screenshot({
      path: fileURLToPath(
        new URL(
          "../Image/figures/rag-answer-desktop_20260928.png",
          import.meta.url,
        ),
      ),
      fullPage: true,
    });
    await page.setViewportSize({ width: 390, height: 844 });
    await page.screenshot({
      path: fileURLToPath(
        new URL(
          "../Image/figures/rag-answer-mobile_20260928.png",
          import.meta.url,
        ),
      ),
      fullPage: true,
    });
    assert.equal(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
      true,
    );
    await page.keyboard.press("Escape");
    await dialog.waitFor({ state: "hidden" });
    await context.close();
  });
} catch (error) {
  if (activePage && !activePage.isClosed()) {
    await activePage.screenshot({
      path: fileURLToPath(
        new URL(
          "../Image/draft/rag-test-failure_20260928.png",
          import.meta.url,
        ),
      ),
      fullPage: true,
    });
    process.stderr.write(
      `页面地址：${activePage.url()}\n页面文本：${(await activePage.locator("body").innerText()).slice(0, 1800)}\n`,
    );
  }
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
} finally {
  await browser?.close();
  report.funnel = {
    original: report.cases.length,
    excluded: 0,
    effective: report.cases.length,
    externalFailures: report.externalFailures.length,
    coreReturned: report.cases.length - report.externalFailures.length,
  };
  await writeFile(
    fileURLToPath(
      new URL("./rag-integration_result_20260928.json", import.meta.url),
    ),
    JSON.stringify(report, null, 2) + "\n",
  );
}
