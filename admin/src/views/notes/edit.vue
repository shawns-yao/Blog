<script setup lang="ts">
import { PreviewLink20Regular } from '@vicons/fluent'
import { PaperPlaneOutline, SaveOutline } from '@vicons/ionicons5'
import {
  NAlert,
  NButton,
  NButtonGroup,
  NCard,
  NDivider,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NPopconfirm,
  NPopover,
  NSelect,
  NSkeleton,
  NSwitch,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, onMounted, onUnmounted, ref, watch, toRef } from 'vue'

import ImageInput from '@/components/image-picker/ImageInput.vue'
import MarkdownEditor from '@/components/markdown-editor/MarkdownEditor.vue'
import MarkdownPreview from '@/components/markdown-editor/MarkdownPreview.vue'
import { parseFederationSignals } from '@/composables/markdown-editor/utils/federation-signals'
import { publishFederationActivityPub } from '@/services/federation-admin'
import { getMomentFederationInteractions, resetMomentFederationSignals } from '@/services/moments'
import AiSummaryAssist from '@/views/shared/content-editor/components/AiSummaryAssist.vue'
import ContentTagSelect from '@/views/shared/content-editor/components/ContentTagSelect.vue'
import EditorStatsOverlay from '@/views/shared/content-editor/components/EditorStatsOverlay.vue'
import {
  useAiSummaryGeneration,
  useAiTitleGeneration,
} from '@/views/shared/content-editor/composables/use-ai-tools'
import { useEditorStats } from '@/views/shared/content-editor/composables/use-editor-stats'
import { usePreviewFrame } from '@/views/shared/content-editor/composables/use-preview-frame'

import MomentAtmosphereFields from './components/MomentAtmosphereFields.vue'
import { useMomentForm } from './composables/use-moment-form'
import { useMomentTaxonomySelect } from './composables/use-moment-taxonomy-select'

import type { MomentDetail } from '@/services/moments'
import type { FederationOutboundInteractionResp } from '@/types/federation'

type FederationSignalType = 'mention' | 'citation'

interface FederationSignalRow {
  key: string
  type: FederationSignalType
  instance: string
  target: string
  marker: string
  inContent: boolean
  deliveredAt: string | null
  outbound: FederationOutboundInteractionResp | null
}

defineOptions({ name: 'NoteEdit' })

const message = useMessage()

const { form, saving, imageProcessing, extInfo, baseExtInfo, isCreating, fetch, save } =
  useMomentForm()

const {
  columnOptions,
  topicOptions,
  topicsLoading,
  topicCreating,
  selectedTopics,
  newColumnModal,
  setInitialTopics,
  createAndSelectTopic,
  createNewColumn,
} = useMomentTaxonomySelect(toRef(form, 'topicIds'), toRef(form, 'columnId'), message)

const { cursorPos, selectionStats, statsIdle, markActivity, handleCursorChange, getStats } =
  useEditorStats()

const showMeta = ref(false)
const loadedMoment = ref<MomentDetail | null>(null)
const apPublishing = ref(false)
const federationInteractionsLoading = ref(false)
const federationInteractionsError = ref('')
const federationOutbounds = ref<FederationOutboundInteractionResp[]>([])
const resetAllFederationLoading = ref(false)
const resetSignalLoadingKeys = ref<Record<string, boolean>>({})

const isYearSummary = computed({
  get: () => form.yearSummary !== null,
  set: (value: boolean) => {
    form.yearSummary = value ? (form.yearSummary ?? new Date().getFullYear()) : null
  },
})

const { loading: aiGenerating, generate: handleAIGenerate } = useAiTitleGeneration({
  getContent: () => form.content,
  applyResult: (result) => {
    form.title = result.title
    form.shortUrl = result.shortUrl
  },
  message,
})

const {
  loading: aiSummaryLoading,
  result: aiSummaryResult,
  done: aiSummaryDone,
  generate: handleAISummary,
  adopt: adoptAISummary,
  dismiss: dismissAISummary,
} = useAiSummaryGeneration({
  getContent: () => form.content,
  adoptSummary: (summary) => {
    form.aiSummary = summary
  },
  message,
})

const stats = computed(() => getStats(form.content))
const actionLabel = computed(() => {
  if (!form.isPublished) return '保存'
  return isCreating.value ? '发布' : '发布新版本'
})
const actionIcon = computed(() => (form.isPublished ? PaperPlaneOutline : SaveOutline))
const {
  showPreview,
  previewMode,
  previewFrameRef,
  previewUrl,
  fetchWebsiteInfo,
  schedulePreviewPayload,
  handlePreviewMessage,
  handlePreviewFrameLoad,
  resetPreviewReady,
  cleanup,
} = usePreviewFrame({
  previewPath: '/internal/preview/moment',
  readyType: 'grtblog-preview:ready',
  postType: 'grtblog-preview:moment',
  buildPayload: buildPreviewPayload,
  message,
})

function formatDateTime(value?: string | null) {
  if (!value) return '-'
  const timestamp = Date.parse(value)
  if (Number.isNaN(timestamp)) return '-'
  return new Date(timestamp).toLocaleString()
}

const apStatusText = computed(() => {
  if (isCreating.value) return '未创建'
  return loadedMoment.value?.activityPubObjectId ? '已发布' : '未发布'
})

const apLastPublishedAtText = computed(() =>
  formatDateTime(loadedMoment.value?.activityPubLastPublishedAt ?? null),
)

const canRepublishToActivityPub = computed(
  () => !isCreating.value && !!loadedMoment.value?.id && form.isPublished,
)

async function handleRepublishActivityPub() {
  if (!loadedMoment.value?.id) return
  if (!form.isPublished) {
    message.warning('请先设为发布并保存，再手动补发')
    return
  }
  apPublishing.value = true
  try {
    const resp = await publishFederationActivityPub({
      source_type: 'moment',
      source_id: loadedMoment.value.id,
    })
    loadedMoment.value.activityPubObjectId =
      resp.object_id || loadedMoment.value.activityPubObjectId
    loadedMoment.value.activityPubLastPublishedAt = resp.published_at
    message.success(`补发完成：成功 ${resp.success_count}，失败 ${resp.failure_count}`)
  } catch (err) {
    message.error(err instanceof Error ? err.message : '补发失败')
  } finally {
    apPublishing.value = false
  }
}

function readFederationRegistry(value: unknown) {
  const root = value && typeof value === 'object' ? (value as Record<string, unknown>) : {}
  const rawRegistry = root._federation_delivery_registry
  const registry =
    rawRegistry && typeof rawRegistry === 'object' ? (rawRegistry as Record<string, unknown>) : {}
  const mentionRegistry =
    registry.mentions && typeof registry.mentions === 'object'
      ? (registry.mentions as Record<string, unknown>)
      : {}
  const citationRegistry =
    registry.citations && typeof registry.citations === 'object'
      ? (registry.citations as Record<string, unknown>)
      : {}

  const mentions: Record<string, string> = {}
  const citations: Record<string, string> = {}

  for (const [key, rawValue] of Object.entries(mentionRegistry)) {
    const normalizedKey = key.trim()
    if (!normalizedKey) continue
    mentions[normalizedKey] = typeof rawValue === 'string' ? rawValue : ''
  }
  for (const [key, rawValue] of Object.entries(citationRegistry)) {
    const normalizedKey = key.trim()
    if (!normalizedKey) continue
    citations[normalizedKey] = typeof rawValue === 'string' ? rawValue : ''
  }

  return { mentions, citations }
}

const federationOutboundBySignalKey = computed(() => {
  const map = new Map<string, FederationOutboundInteractionResp>()
  for (const item of federationOutbounds.value) {
    const key = item.signal_key?.trim()
    if (!key) continue
    const prev = map.get(key)
    if (!prev) {
      map.set(key, item)
      continue
    }
    const prevTs = Date.parse(prev.updated_at || prev.created_at || '')
    const curTs = Date.parse(item.updated_at || item.created_at || '')
    if ((Number.isFinite(curTs) ? curTs : 0) >= (Number.isFinite(prevTs) ? prevTs : 0)) {
      map.set(key, item)
    }
  }
  return map
})

const federationSignalRows = computed<FederationSignalRow[]>(() => {
  const rows = new Map<string, FederationSignalRow>()
  const parsedSignals = parseFederationSignals(form.content || '')
  const registry = readFederationRegistry(extInfo.value ?? baseExtInfo.value ?? null)

  const ensureMentionRow = (key: string) => {
    const splitAt = key.indexOf('@')
    if (splitAt <= 0 || splitAt >= key.length - 1) return
    const target = key.slice(0, splitAt)
    const instance = key.slice(splitAt + 1)
    if (!target || !instance) return
    if (!rows.has(key)) {
      rows.set(key, {
        key,
        type: 'mention',
        instance,
        target,
        marker: `<@${target}@${instance}>`,
        inContent: false,
        deliveredAt: null,
        outbound: null,
      })
    }
  }

  const ensureCitationRow = (key: string) => {
    const splitAt = key.indexOf('|')
    if (splitAt <= 0 || splitAt >= key.length - 1) return
    const instance = key.slice(0, splitAt)
    const target = key.slice(splitAt + 1)
    if (!instance || !target) return
    if (!rows.has(key)) {
      rows.set(key, {
        key,
        type: 'citation',
        instance,
        target,
        marker: `<cite:${instance}|${target}>`,
        inContent: false,
        deliveredAt: null,
        outbound: null,
      })
    }
  }

  parsedSignals.mentions.forEach((item) => {
    ensureMentionRow(item.key)
    const row = rows.get(item.key)
    if (row) row.inContent = true
  })
  parsedSignals.citations.forEach((item) => {
    ensureCitationRow(item.key)
    const row = rows.get(item.key)
    if (row) row.inContent = true
  })

  Object.entries(registry.mentions).forEach(([key, deliveredAt]) => {
    ensureMentionRow(key)
    const row = rows.get(key)
    if (row) row.deliveredAt = deliveredAt || null
  })
  Object.entries(registry.citations).forEach(([key, deliveredAt]) => {
    ensureCitationRow(key)
    const row = rows.get(key)
    if (row) row.deliveredAt = deliveredAt || null
  })

  federationOutboundBySignalKey.value.forEach((outbound, key) => {
    if (outbound.type === 'mention') {
      ensureMentionRow(key)
    } else if (outbound.type === 'citation') {
      ensureCitationRow(key)
    }
    const row = rows.get(key)
    if (row) row.outbound = outbound
  })

  return Array.from(rows.values()).sort((a, b) => {
    if (a.type !== b.type) return a.type === 'mention' ? -1 : 1
    return a.key.localeCompare(b.key)
  })
})

function outboundStatusTagType(status?: string | null) {
  const normalized = (status || '').trim().toLowerCase()
  if (normalized === 'approved' || normalized === 'accepted') return 'success'
  if (normalized === 'queued' || normalized === 'sending') return 'warning'
  if (
    normalized === 'rejected' ||
    normalized === 'failed' ||
    normalized === 'timeout' ||
    normalized === 'dead'
  ) {
    return 'error'
  }
  return 'default'
}

function outboundStatusText(status?: string | null) {
  const normalized = (status || '').trim().toLowerCase()
  switch (normalized) {
    case 'queued':
      return '排队中'
    case 'sending':
      return '投递中'
    case 'accepted':
      return '已接收'
    case 'approved':
      return '已通过'
    case 'rejected':
      return '已拒绝'
    case 'failed':
      return '失败'
    case 'timeout':
      return '超时'
    case 'dead':
      return '终止'
    default:
      return normalized || '未知'
  }
}

function signalStatusText(row: FederationSignalRow) {
  if (row.outbound) return `队列：${outboundStatusText(row.outbound.status)}`
  if (row.deliveredAt) return '已记录（未发现出站记录）'
  if (row.inContent) return '正文存在，待触发'
  return '已脱离正文'
}

async function fetchFederationInteractions(momentID?: number) {
  const id = momentID ?? loadedMoment.value?.id
  if (!id) {
    federationOutbounds.value = []
    federationInteractionsError.value = ''
    return
  }
  federationInteractionsLoading.value = true
  federationInteractionsError.value = ''
  try {
    const data = await getMomentFederationInteractions(id)
    federationOutbounds.value = data.outbound ?? []
  } catch (err) {
    federationOutbounds.value = []
    federationInteractionsError.value = err instanceof Error ? err.message : '加载联合状态失败'
  } finally {
    federationInteractionsLoading.value = false
  }
}

function setSignalResetLoading(key: string, loading: boolean) {
  const next = { ...resetSignalLoadingKeys.value }
  if (loading) next[key] = true
  else delete next[key]
  resetSignalLoadingKeys.value = next
}

function applyFederationResetResult(extInfoValue: MomentDetail['extInfo']) {
  const cloned = extInfoValue ? JSON.parse(JSON.stringify(extInfoValue)) : null
  baseExtInfo.value = cloned
  if (loadedMoment.value) {
    loadedMoment.value.extInfo = cloned
  }
}

async function handleResetFederationSignal(row: FederationSignalRow) {
  if (!loadedMoment.value?.id) return
  setSignalResetLoading(row.key, true)
  try {
    const payload = row.type === 'mention' ? { mentions: [row.key] } : { citations: [row.key] }
    const result = await resetMomentFederationSignals(loadedMoment.value.id, payload)
    applyFederationResetResult(result.extInfo ?? null)
    await fetchFederationInteractions(loadedMoment.value.id)
    if (result.retriggered) {
      message.success('已重置并重新触发该联合条目')
    } else {
      message.success('已重置该联合条目（当前未触发出站）')
    }
  } catch (err) {
    message.error(err instanceof Error ? err.message : '重置失败')
  } finally {
    setSignalResetLoading(row.key, false)
  }
}

async function handleResetAllFederationSignals() {
  if (!loadedMoment.value?.id) return
  resetAllFederationLoading.value = true
  try {
    const result = await resetMomentFederationSignals(loadedMoment.value.id)
    applyFederationResetResult(result.extInfo ?? null)
    await fetchFederationInteractions(loadedMoment.value.id)
    if (result.retriggered) {
      message.success('已重置全部联合条目并重新触发')
    } else {
      message.success('已重置全部联合条目（当前未触发出站）')
    }
  } catch (err) {
    message.error(err instanceof Error ? err.message : '重置失败')
  } finally {
    resetAllFederationLoading.value = false
  }
}

function buildPreviewPayload() {
  const nowIso = new Date().toISOString()
  const selectedColumn = columnOptions.value.find((option) => option.value === form.columnId)
  return {
    id: loadedMoment.value?.id ?? 0,
    title: form.title,
    summary: form.summary,
    aiSummary: form.aiSummary ?? null,
    content: form.content,
    contentHash: loadedMoment.value?.contentHash ?? '',
    shortUrl: form.shortUrl,
    cover: form.cover || null,
    extInfo: extInfo.value,
    columnId: form.columnId,
    columnName: selectedColumn?.label ? String(selectedColumn.label) : undefined,
    commentAreaId: null,
    toc: undefined,
    topics: selectedTopics.value,
    metrics: loadedMoment.value ? { views: 0, likes: 0, comments: 0 } : undefined,
    isPublished: form.isPublished,
    isTop: form.isTop,
    isOriginal: form.isOriginal,
    createdAt: loadedMoment.value?.createdAt ?? nowIso,
    updatedAt: nowIso,
    authorId: loadedMoment.value?.authorId ?? 0,
  }
}
onMounted(async () => {
  window.addEventListener('message', handlePreviewMessage)

  const [data] = await Promise.all([fetch(), fetchWebsiteInfo()])
  loadedMoment.value = data as MomentDetail | null
  await fetchFederationInteractions(data?.id)
  if (data?.topics) {
    setInitialTopics(data.topics)
  }
})

onUnmounted(() => {
  window.removeEventListener('message', handlePreviewMessage)
  cleanup()
})

watch(
  () => [
    form.title,
    form.summary,
    form.aiSummary,
    form.content,
    form.cover,
    form.shortUrl,
    form.columnId,
    form.topicIds,
    form.isPublished,
    form.isTop,
    form.isOriginal,
    form.allowComment,
    form.weather,
    form.mood,
  ],
  () => {
    schedulePreviewPayload()
  },
  { deep: true },
)

watch([showPreview, previewMode, previewUrl], () => {
  schedulePreviewPayload()
})

watch(previewUrl, () => {
  resetPreviewReady()
})
</script>

<template>
  <div class="flex h-full min-h-0 flex-col">
    <header
      class="z-10 flex shrink-0 flex-col gap-3 px-10 py-8 backdrop-blur sm:h-24 sm:flex-row sm:items-center sm:justify-between sm:py-0"
    >
      <div class="flex w-full items-center gap-4 sm:flex-1">
        <NInput
          v-model:value="form.title"
          placeholder="在这里开始你的记录..."
          :bordered="false"
          class="flex-1 text-xl! leading-tight font-bold sm:text-2xl!"
          style="--n-caret-color: var(--primary-color); background-color: transparent"
        />
      </div>

      <div class="flex w-full flex-wrap items-center gap-3 sm:w-auto sm:flex-nowrap sm:gap-4">
        <div class="flex items-baseline gap-1">
          <div class="iconify self-center ph--link-simple" />
          <span class="text-xs leading-none">/moments/</span>
          <input
            v-model="form.shortUrl"
            placeholder="请填写短链接"
            class="w-24 border-b border-current/30 p-0 pb-0.5 text-[11px] leading-none focus:border-primary focus:outline-none sm:w-32"
          />
        </div>

        <NButton
          quaternary
          size="small"
          :loading="aiGenerating"
          :disabled="!form.content?.trim()"
          @click="handleAIGenerate"
        >
          <template #icon><div class="iconify ph--robot" /></template>
          AI
        </NButton>

        <NButtonGroup>
          <NButton
            :type="!form.isPublished ? 'primary' : 'default'"
            :ghost="form.isPublished"
            @click="form.isPublished = false"
          >
            草稿
          </NButton>
          <NButton
            :type="form.isPublished ? 'primary' : 'default'"
            :ghost="!form.isPublished"
            @click="form.isPublished = true"
          >
            发布
          </NButton>
        </NButtonGroup>

        <div class="flex items-center gap-2">
          <span
            v-if="imageProcessing"
            class="text-xs text-amber-600"
          >
            正在处理图片…
          </span>
          <NButton
            quaternary
            circle
            size="small"
            @click="showMeta = true"
          >
            <template #icon><div class="iconify text-xl ph--sliders-horizontal" /></template>
          </NButton>

          <NButton
            quaternary
            circle
            size="small"
            :type="showPreview ? 'primary' : 'default'"
            @click="showPreview = !showPreview"
          >
            <template #icon><PreviewLink20Regular /></template>
          </NButton>

          <NButton
            type="primary"
            size="medium"
            :loading="saving"
            :disabled="saving || imageProcessing"
            @click="save"
            class="px-5 font-medium shadow-sm active:scale-95"
          >
            <template #icon><component :is="actionIcon" /></template>
            {{ actionLabel }}
          </NButton>
        </div>
      </div>
    </header>

    <main class="flex min-h-0 flex-1 overflow-hidden">
      <div
        class="editor-container grid h-full min-h-0 w-full"
        :class="showPreview ? 'grid-cols-1 lg:grid-cols-2' : 'grid-cols-1'"
      >
        <div
          class="pane editor-pane relative h-full overflow-auto"
          @scroll="markActivity"
          @wheel="markActivity"
        >
          <MarkdownEditor
            v-model="form.content"
            class="h-full min-h-full"
            @cursor-change="handleCursorChange"
          />

          <EditorStatsOverlay
            :idle="statsIdle"
            :cursor-line="cursorPos.line"
            :cursor-column="cursorPos.column"
            :reading-minutes="stats.readingMinutes"
            :char-count="stats.charCount"
            :chinese-char-count="stats.chineseCharCount"
            :word-count="stats.wordCount"
            :total-char-count="stats.totalCharCount"
            :paragraph-count="stats.paragraphCount"
            :selection-total="selectionStats.total"
            :selection-chars="selectionStats.chars"
          />
        </div>

        <div
          v-if="showPreview"
          class="pane preview-pane relative h-full overflow-auto"
          @scroll="markActivity"
        >
          <div class="absolute top-3 right-3 z-10">
            <NPopover
              trigger="click"
              placement="bottom-end"
            >
              <template #trigger>
                <NButton
                  tertiary
                  type="primary"
                  circle
                  size="small"
                  class="shadow-sm"
                >
                  <template #icon><div class="iconify text-lg ph--dots-three-vertical" /></template>
                </NButton>
              </template>
              <div class="flex flex-col gap-1 p-1">
                <NButton
                  :type="previewMode === 'markdown' ? 'primary' : 'default'"
                  quaternary
                  size="small"
                  class="w-full justify-start px-2"
                  @click="previewMode = 'markdown'"
                  >Markdown 预览</NButton
                >
                <NButton
                  :type="previewMode === 'page' ? 'primary' : 'default'"
                  quaternary
                  size="small"
                  class="w-full justify-start px-2"
                  @click="previewMode = 'page'"
                  >网页预览</NButton
                >
              </div>
            </NPopover>
          </div>

          <MarkdownPreview
            v-if="previewMode === 'markdown'"
            :source="form.content"
            class="p-4 sm:p-8"
          />
          <div
            v-else
            class="h-full w-full"
          >
            <iframe
              v-if="previewUrl"
              :src="previewUrl"
              ref="previewFrameRef"
              class="h-full w-full border-0"
              @load="handlePreviewFrameLoad"
            />
            <div
              v-else
              class="flex h-full items-center justify-center text-sm opacity-60"
            >
              请先在站点设置中配置 public_url
            </div>
          </div>
        </div>
      </div>
    </main>

    <NDrawer
      v-model:show="showMeta"
      placement="right"
      width="400"
    >
      <NDrawerContent
        title="手记设置"
        :native-scrollbar="false"
        closable
        header-style="padding: 24px;"
        body-style="padding: 24px;"
      >
        <div class="flex flex-col gap-6">
          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--tag" />
              <span>分区与话题</span>
            </div>
            <NForm
              label-placement="top"
              label-width="auto"
              class="space-y-4"
            >
              <NFormItem
                label="分区"
                :show-feedback="false"
              >
                <div class="flex w-full items-center gap-2">
                  <NSelect
                    v-model:value="form.columnId"
                    :options="columnOptions"
                    placeholder="选择分区"
                    clearable
                    filterable
                    class="flex-1"
                  />
                  <NButton
                    quaternary
                    size="small"
                    @click="newColumnModal.show = true"
                    >新建</NButton
                  >
                </div>
              </NFormItem>
              <NFormItem
                label="话题"
                :show-feedback="false"
              >
                <ContentTagSelect
                  v-model="form.topicIds"
                  :options="topicOptions"
                  :loading="topicsLoading"
                  :creating="topicCreating"
                  noun="话题"
                  class="w-full"
                  @create="createAndSelectTopic"
                />
              </NFormItem>
            </NForm>
          </div>

          <NDivider style="margin: 0" />

          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--cloud-sun" />
              <span>此刻</span>
            </div>
            <MomentAtmosphereFields
              v-model:weather="form.weather"
              v-model:mood="form.mood"
            />
          </div>

          <NDivider style="margin: 0" />

          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--note" />
              <span>元信息</span>
            </div>
            <NForm
              label-placement="top"
              label-width="auto"
              class="space-y-4"
            >
              <NFormItem :show-feedback="false">
                <template #label>
                  <span>摘要</span>
                  <span class="ml-1 text-xs opacity-50">用于外显描述、OG 信息、SEO</span>
                </template>
                <NInput
                  v-model:value="form.summary"
                  type="textarea"
                  placeholder="外显摘要，用于列表卡片、网页描述、社交分享..."
                  :autosize="{ minRows: 2, maxRows: 4 }"
                />
              </NFormItem>
              <AiSummaryAssist
                :model-value="form.aiSummary"
                :loading="aiSummaryLoading"
                :result="aiSummaryResult"
                :done="aiSummaryDone"
                placeholder="AI 生成的内容导读，展示在正文之前..."
                :disabled="!form.content?.trim()"
                @update:model-value="form.aiSummary = $event"
                @generate="handleAISummary"
                @adopt="adoptAISummary"
                @dismiss="dismissAISummary"
              />
              <NFormItem
                label="封面"
                :show-feedback="false"
              >
                <ImageInput v-model:value="form.cover" />
              </NFormItem>
            </NForm>
          </div>

          <NDivider style="margin: 0" />

          <div class="space-y-4">
            <div class="flex items-center gap-2 text-sm font-medium">
              <div class="iconify ph--toggle-left" />
              <span>属性</span>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="flex items-center justify-between rounded-lg px-4 py-3">
                <span class="text-sm">置顶</span
                ><NSwitch
                  v-model:value="form.isTop"
                  size="small"
                />
              </div>
              <div class="flex items-center justify-between rounded-lg px-4 py-3">
                <span class="text-sm">允许评论</span
                ><NSwitch
                  v-model:value="form.allowComment"
                  size="small"
                />
              </div>
              <div class="flex items-center justify-between rounded-lg px-4 py-3">
                <span class="text-sm">原创</span
                ><NSwitch
                  v-model:value="form.isOriginal"
                  size="small"
                />
              </div>
              <div class="col-span-2 rounded-lg px-4 py-3">
                <div class="flex items-start justify-between gap-4">
                  <div class="min-w-0 space-y-1">
                    <div class="text-sm">ActivityPub：{{ apStatusText }}</div>
                    <div class="text-xs opacity-70">最近发布：{{ apLastPublishedAtText }}</div>
                    <div
                      v-if="loadedMoment?.activityPubObjectId"
                      class="text-xs break-all opacity-70"
                    >
                      {{ loadedMoment.activityPubObjectId }}
                    </div>
                  </div>
                  <NButton
                    size="small"
                    secondary
                    :loading="apPublishing"
                    :disabled="!canRepublishToActivityPub || apPublishing"
                    @click="handleRepublishActivityPub"
                  >
                    手动补发
                  </NButton>
                </div>
              </div>
              <div class="col-span-2 rounded-lg px-4 py-3">
                <div class="flex items-center justify-between gap-3">
                  <span class="text-sm">这是年终总结</span>
                  <NSwitch
                    v-model:value="isYearSummary"
                    size="small"
                  />
                </div>
                <div
                  v-if="isYearSummary"
                  class="mt-3"
                >
                  <NInputNumber
                    v-model:value="form.yearSummary"
                    :min="1900"
                    :max="3000"
                    :precision="0"
                    class="w-full"
                    placeholder="输入年份，例如 2024"
                  />
                </div>
              </div>
            </div>
          </div>

          <NDivider style="margin: 0" />

          <div class="space-y-4">
            <div class="flex items-center justify-between gap-3">
              <div class="flex items-center gap-2 text-sm font-medium">
                <div class="iconify ph--circles-three" />
                <span>联合条目状态</span>
              </div>
              <NPopconfirm
                trigger="click"
                @positive-click="handleResetAllFederationSignals"
              >
                <template #trigger>
                  <NButton
                    size="small"
                    secondary
                    :loading="resetAllFederationLoading"
                    :disabled="
                      isCreating || federationSignalRows.length === 0 || resetAllFederationLoading
                    "
                  >
                    全部重置
                  </NButton>
                </template>
                将重置全部联合条目状态，并尝试重新触发出站。
              </NPopconfirm>
            </div>

            <NAlert
              v-if="isCreating"
              type="info"
              :show-icon="false"
            >
              新建手记保存后可查看联合条目状态。
            </NAlert>

            <NAlert
              v-else-if="federationInteractionsError"
              type="warning"
              :show-icon="false"
            >
              {{ federationInteractionsError }}
            </NAlert>

            <div
              v-else-if="federationInteractionsLoading"
              class="space-y-3"
            >
              <NSkeleton
                text
                :repeat="2"
              />
              <NSkeleton
                text
                :repeat="2"
              />
            </div>

            <NEmpty
              v-else-if="federationSignalRows.length === 0"
              size="small"
              description="当前未识别到联合条目"
            />

            <div
              v-else
              class="space-y-3"
            >
              <div
                v-for="row in federationSignalRows"
                :key="row.key"
                class="rounded-lg border border-current/10 p-3"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0 space-y-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <NTag
                        size="small"
                        :bordered="false"
                        :type="row.type === 'mention' ? 'info' : 'warning'"
                      >
                        {{ row.type === 'mention' ? '提及' : '引用' }}
                      </NTag>
                      <NTag
                        size="small"
                        :bordered="false"
                        :type="outboundStatusTagType(row.outbound?.status)"
                      >
                        {{ signalStatusText(row) }}
                      </NTag>
                      <NTag
                        v-if="!row.inContent"
                        size="small"
                        :bordered="false"
                        type="default"
                      >
                        已不在正文
                      </NTag>
                    </div>
                    <div class="font-mono text-xs break-all opacity-80">
                      {{ row.marker }}
                    </div>
                    <div class="text-xs opacity-70">
                      目标：{{ row.instance }} / {{ row.target }}
                    </div>
                    <div
                      v-if="row.deliveredAt"
                      class="text-xs opacity-70"
                    >
                      已记录时间：{{ formatDateTime(row.deliveredAt) }}
                    </div>
                    <div
                      v-if="row.outbound?.updated_at"
                      class="text-xs opacity-70"
                    >
                      队列更新时间：{{ formatDateTime(row.outbound.updated_at) }}
                    </div>
                  </div>

                  <NPopconfirm
                    trigger="click"
                    @positive-click="handleResetFederationSignal(row)"
                  >
                    <template #trigger>
                      <NButton
                        size="tiny"
                        secondary
                        :loading="!!resetSignalLoadingKeys[row.key]"
                        :disabled="!!resetSignalLoadingKeys[row.key]"
                      >
                        重置
                      </NButton>
                    </template>
                    将重置该条目状态，并尝试重新触发一次出站。
                  </NPopconfirm>
                </div>
              </div>
            </div>
          </div>
        </div>
      </NDrawerContent>
    </NDrawer>

    <NModal
      v-model:show="newColumnModal.show"
      style="width: 420px; max-width: 90vw"
    >
      <NCard
        title="新建分区"
        size="small"
      >
        <NForm
          label-placement="top"
          label-width="auto"
          class="space-y-3"
        >
          <NFormItem
            label="名称"
            :show-feedback="false"
          >
            <NInput
              v-model:value="newColumnModal.name"
              placeholder="例如：日常"
            />
          </NFormItem>
          <NFormItem
            label="短链接"
            :show-feedback="false"
          >
            <NInput
              v-model:value="newColumnModal.slug"
              placeholder="例如：daily"
            />
          </NFormItem>
        </NForm>
        <div class="mt-4 flex justify-end gap-2">
          <NButton
            quaternary
            @click="newColumnModal.show = false"
            >取消</NButton
          >
          <NButton
            type="primary"
            :loading="newColumnModal.loading"
            @click="createNewColumn"
            >创建并选择</NButton
          >
        </div>
      </NCard>
    </NModal>
  </div>
</template>

<style scoped>
.editor-container {
  /* Clean grid layout handled by Tailwind classes */
}

.pane::-webkit-scrollbar,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar) {
  width: 5px;
  height: 5px;
}

.pane::-webkit-scrollbar-track,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar-track),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar-track) {
  background: transparent;
}

:global(.dark) .pane::-webkit-scrollbar-thumb,
:global(.dark) .editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb),
:global(.dark) .preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb) {
  background-color: #374151;
}

.pane::-webkit-scrollbar-thumb:hover,
.editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb:hover),
.preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb:hover) {
  background-color: #d1d5db;
}

:global(.dark) .pane::-webkit-scrollbar-thumb:hover,
:global(.dark) .editor-pane :deep(.cm-scroller::-webkit-scrollbar-thumb:hover),
:global(.dark) .preview-pane :deep(.markdown-preview::-webkit-scrollbar-thumb:hover) {
  background-color: #4b5563;
}

.editor-pane :deep(.cm-editor) {
  height: 100% !important;
  font-family: inherit;
}

.editor-pane :deep(.cm-scroller) {
  padding-bottom: 50vh;
  font-family: 'JetBrains Mono', monospace;
  line-height: 1.6;
}

.preview-pane :deep(.markdown-preview) {
  padding-bottom: 50vh;
}
</style>
