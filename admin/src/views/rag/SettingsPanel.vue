<script setup lang="ts">
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import {
  NAlert,
  NButton,
  NDescriptions,
  NDescriptionsItem,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSwitch,
} from 'naive-ui'
import { computed, ref, watch } from 'vue'

import { useInjection } from '@/composables'
import { mediaQueryInjectionKey } from '@/injection'
import { saveRagSettings } from '@/services/rag'

import ChatPriority from './ChatPriority.vue'

import type { RagSettings, RagTuning } from '@/services/rag'

const props = defineProps<{ settings?: RagSettings }>()
const { isMaxSm } = useInjection(mediaQueryInjectionKey)
type Draft = { [K in keyof RagTuning]: RagTuning[K] extends number ? number | null : RagTuning[K] }
type NumberKey = {
  [K in keyof RagTuning]: RagTuning[K] extends number ? K : never
}[keyof RagTuning]
interface Field {
  key: NumberKey
  label: string
  min: number
  max: number
  step?: number
}
const form = ref<Draft | null>(null)
const saved = ref('')
const saveError = ref('')
const client = useQueryClient()
const dirty = computed(() => form.value !== null && JSON.stringify(form.value) !== saved.value)

watch(
  () => props.settings,
  (value) => {
    if (value && !dirty.value) {
      form.value = { ...value.tuning }
      saved.value = JSON.stringify(form.value)
    }
  },
  { immediate: true },
)

const groups: { title: string; fields: Field[] }[] = [
  {
    title: '分块',
    fields: [
      { key: 'chunkTargetTokens', label: '子块目标（token）', min: 100, max: 4000 },
      { key: 'chunkMinTokens', label: '短块合并阈值（token）', min: 1, max: 4000 },
      { key: 'chunkMaxTokens', label: '子块上限（token）', min: 100, max: 4000 },
      { key: 'chunkOverlapTokens', label: '重叠上限（token）', min: 0, max: 3999 },
      { key: 'parentMaxTokens', label: '父块上限（token）', min: 100, max: 8000 },
    ],
  },
  {
    title: '召回与融合',
    fields: [
      { key: 'vectorTopK', label: '向量召回 TopK', min: 1, max: 100 },
      { key: 'keywordTopK', label: 'BM25 召回 TopK', min: 1, max: 100 },
      { key: 'bm25K1', label: 'BM25 K1', min: 0.01, max: 3, step: 0.05 },
      { key: 'bm25B', label: 'BM25 B', min: 0, max: 1, step: 0.05 },
      { key: 'minSimilarity', label: '向量相似度下限', min: 0, max: 1, step: 0.01 },
      { key: 'rrfK', label: 'RRF 平滑常数 K', min: 1, max: 200 },
      { key: 'rrfVectorWeight', label: 'RRF 向量权重', min: 0, max: 1, step: 0.05 },
      { key: 'rrfKeywordWeight', label: 'RRF 关键词权重', min: 0, max: 1, step: 0.05 },
      { key: 'rerankCandidateTopK', label: '融合候选 TopK', min: 1, max: 100 },
      { key: 'topK', label: '基准 TopK', min: 1, max: 20 },
    ],
  },
  {
    title: '证据数量',
    fields: [
      { key: 'dynamicTopKMin', label: '动态 TopK 下限', min: 1, max: 20 },
      { key: 'dynamicTopKMax', label: '动态 TopK 上限', min: 1, max: 20 },
      { key: 'evidenceDiversityWeight', label: '证据多样性权重', min: 0, max: 1, step: 0.05 },
    ],
  },
  {
    title: '问题理解与上下文',
    fields: [
      { key: 'multiQueryMax', label: '子查询上限', min: 1, max: 3 },
      { key: 'contextMaxTokens', label: '证据预算（token）', min: 100, max: 16000 },
      { key: 'historyMaxTokens', label: '历史预算（token）', min: 0, max: 8000 },
    ],
  },
  {
    title: '重排序',
    fields: [{ key: 'rerankThreshold', label: '重排序分数下限', min: -10, max: 10, step: 0.01 }],
  },
]

const validation = computed(() => {
  const f = form.value
  if (!f) return ''
  if (Object.values(f).some((value) => value === null)) return '请填写所有数值参数。'
  if (
    f.chunkMinTokens! > f.chunkTargetTokens! ||
    f.chunkTargetTokens! > f.chunkMaxTokens! ||
    f.chunkOverlapTokens! >= f.chunkTargetTokens!
  )
    return '子块需满足短块合并阈值 ≤ 目标 ≤ 上限，重叠小于目标。'
  if (f.parentMaxTokens! < f.chunkMaxTokens! || f.contextMaxTokens! < f.chunkMaxTokens!)
    return '父块上限和证据预算不能小于子块上限。'
  if (
    f.rrfVectorWeight != null &&
    f.rrfKeywordWeight != null &&
    Math.abs(f.rrfVectorWeight + f.rrfKeywordWeight - 1) > 0.000001
  )
    return 'RRF 两路权重之和需要等于 1。'
  if (f.rerankCandidateTopK != null && f.topK != null && f.rerankCandidateTopK < f.topK)
    return '融合候选 TopK 不能小于最终 TopK。'
  if (f.dynamicTopKMin! > f.dynamicTopKMax!) return '动态 TopK 下限不能超过上限。'
  if (f.dynamicTopKEnabled && f.dynamicTopKMax! > f.rerankCandidateTopK!)
    return '动态 TopK 上限不能超过融合候选数。'
  if (
    f.rerankCandidateTopK != null &&
    f.vectorTopK != null &&
    f.keywordTopK != null &&
    f.rerankCandidateTopK > (f.vectorTopK + f.keywordTopK) * (1 + f.multiQueryMax!)
  )
    return '融合候选数不能超过所有查询的两路召回总数。'
  if (!f.indexVersion.trim()) return '请填写索引版本。'
  return ''
})

const mutation = useMutation({
  mutationFn: saveRagSettings,
  onSuccess: async (value) => {
    form.value = { ...value.tuning }
    saved.value = JSON.stringify(form.value)
    saveError.value = ''
    client.setQueryData(['rag', 'settings'], value)
    await client.invalidateQueries({ queryKey: ['rag'] })
  },
  onError: (error: Error) => {
    saveError.value = error.message
  },
})

function setNumber(key: NumberKey, value: number | null) {
  if (form.value) form.value[key] = value
}
function submit() {
  if (!form.value || validation.value || !dirty.value || mutation.isPending.value) return
  saveError.value = ''
  mutation.mutate({ ...form.value } as RagTuning)
}
function reset() {
  if (props.settings) {
    form.value = { ...props.settings.tuning }
    saved.value = JSON.stringify(form.value)
    saveError.value = ''
  }
}
</script>

<template>
  <div
    v-if="settings && form"
    class="space-y-6 pt-3"
  >
    <ChatPriority :channels="settings.chatChannels" />
    <NDescriptions
      label-placement="top"
      :column="isMaxSm ? 1 : 2"
      class="break-words"
    >
      <NDescriptionsItem label="向量模型"
        >{{ settings.embeddingModel }} ·
        {{ settings.embeddingConfigured ? '已配置' : '未配置' }}</NDescriptionsItem
      >
      <NDescriptionsItem label="重排序模型"
        >{{ settings.rerankModel }} ·
        {{ settings.rerankConfigured ? '已配置' : '未配置' }}</NDescriptionsItem
      >
    </NDescriptions>
    <NAlert
      v-if="form.rerankEnabled && !settings.rerankConfigured"
      type="warning"
    >
      重排序模型尚未配置。{{
        form.rerankFallback
          ? '问答将使用 RRF 排名，并记录重排序降级。'
          : '问答会暂停，直到重排序配置可用。'
      }}
    </NAlert>
    <NForm
      label-placement="top"
      @submit.prevent="submit"
    >
      <section
        v-for="group in groups"
        :key="group.title"
        class="mb-6 border-b border-current/10 pb-3"
      >
        <h2 class="mb-4 text-base font-medium">{{ group.title }}</h2>
        <p
          v-if="group.title === '分块'"
          class="mb-4 text-sm opacity-60"
        >
          参考编码
          {{ settings.tokenEncoding }}。保留段落与结构边界；修改分块参数或索引版本会重建子块索引。
        </p>
        <div
          v-if="group.title === '分块'"
          class="mb-4 flex items-center gap-3"
        >
          <span>按文档结构调整分块</span>
          <NSwitch
            v-model:value="form.adaptiveChunkingEnabled"
            aria-label="按文档结构调整分块"
          />
        </div>
        <div
          v-if="group.title === '召回与融合'"
          class="mb-4 flex items-center gap-3"
        >
          <span>按问题调整召回与重排</span>
          <NSwitch
            v-model:value="form.adaptiveRetrievalEnabled"
            aria-label="按问题调整召回与重排"
          />
        </div>
        <div
          v-if="group.title === '证据数量'"
          class="mb-4 flex items-center gap-3"
        >
          <span>动态 TopK</span>
          <NSwitch
            v-model:value="form.dynamicTopKEnabled"
            aria-label="启用动态 TopK"
          />
        </div>
        <div
          v-if="group.title === '证据数量'"
          class="mb-4 flex items-center gap-3"
        >
          <span>按覆盖与重复信息选择证据</span>
          <NSwitch
            v-model:value="form.evidenceSelectionEnabled"
            aria-label="按覆盖与重复信息选择证据"
          />
        </div>
        <p
          v-if="group.title === '证据数量'"
          class="mb-4 text-sm opacity-60"
        >
          启用后按问题类型、证据排名和上下文预算调整数量；下限不保证补足无关证据。关闭后使用基准
          TopK 作为固定上限。
        </p>
        <p
          v-if="group.title === '召回与融合'"
          class="mb-4 text-sm opacity-60"
        >
          启用策略后，召回与融合参数作为上限；实际数量按问题调整。原问题与子查询分别召回，再经 RRF
          和重排序选择证据。
        </p>
        <div
          v-if="group.title === '问题理解与上下文'"
          class="mb-4 flex items-center gap-3"
        >
          <span>比较与多步骤问题启用子查询</span>
          <NSwitch
            v-model:value="form.multiQueryEnabled"
            aria-label="启用子查询"
          />
        </div>
        <div
          v-if="group.title === '重排序'"
          class="mb-4 flex flex-wrap gap-x-8 gap-y-4"
        >
          <div class="flex items-center gap-3">
            <span>启用重排序</span
            ><NSwitch
              v-model:value="form.rerankEnabled"
              aria-label="启用重排序"
            />
          </div>
          <div class="flex items-center gap-3">
            <span>失败时回退到 RRF</span
            ><NSwitch
              v-model:value="form.rerankFallback"
              :disabled="!form.rerankEnabled"
              aria-label="重排序失败时回退到 RRF"
            />
          </div>
        </div>
        <div class="grid grid-cols-1 gap-x-6 sm:grid-cols-2 xl:grid-cols-3">
          <NFormItem
            v-for="field in group.fields"
            :key="field.key"
            :label="field.label"
          >
            <NInputNumber
              :value="form[field.key]"
              :min="field.min"
              :max="field.max"
              :step="field.step ?? 1"
              :precision="field.step ? 2 : 0"
              :aria-label="field.label"
              class="w-full"
              :disabled="
                (group.title === '重排序' && !form.rerankEnabled) ||
                (field.key.startsWith('dynamicTopK') && !form.dynamicTopKEnabled) ||
                (field.key === 'evidenceDiversityWeight' && !form.evidenceSelectionEnabled)
              "
              @update:value="(value) => setNumber(field.key, value)"
            />
          </NFormItem>
          <NFormItem
            v-if="group.title === '分块'"
            label="索引版本"
          >
            <NInput
              v-model:value="form.indexVersion"
              :maxlength="64"
              aria-label="索引版本"
            />
          </NFormItem>
        </div>
      </section>
      <NAlert
        v-if="validation || saveError"
        type="error"
        class="mb-4"
        >{{ validation || saveError }}</NAlert
      >
      <div class="flex flex-wrap items-center gap-3">
        <NButton
          type="primary"
          attr-type="submit"
          :loading="mutation.isPending.value"
          :disabled="!dirty || !!validation"
          >保存配置</NButton
        >
        <NButton
          :disabled="!dirty || mutation.isPending.value"
          @click="reset"
          >还原</NButton
        >
        <span
          v-if="mutation.isSuccess.value && !dirty"
          class="text-sm opacity-60"
          role="status"
          >配置已保存</span
        >
      </div>
    </NForm>
  </div>
</template>
