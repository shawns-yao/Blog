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
      { key: 'chunkSize', label: '分块大小（字符）', min: 200, max: 2000 },
      { key: 'chunkOverlap', label: '重叠大小（字符）', min: 0, max: 1999 },
    ],
  },
  {
    title: '召回与融合',
    fields: [
      { key: 'vectorTopK', label: '向量召回 TopK', min: 1, max: 100 },
      { key: 'keywordTopK', label: '关键词召回 TopK', min: 1, max: 100 },
      { key: 'minSimilarity', label: '向量相似度下限', min: 0, max: 1, step: 0.01 },
      { key: 'rrfK', label: 'RRF 平滑常数 K', min: 1, max: 200 },
      { key: 'rrfVectorWeight', label: 'RRF 向量权重', min: 0, max: 1, step: 0.05 },
      { key: 'rrfKeywordWeight', label: 'RRF 关键词权重', min: 0, max: 1, step: 0.05 },
      { key: 'rerankCandidateTopK', label: '融合候选 TopK', min: 1, max: 100 },
      { key: 'topK', label: '最终 TopK', min: 1, max: 20 },
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
  if (f.chunkSize != null && f.chunkOverlap != null && f.chunkOverlap >= f.chunkSize)
    return '重叠大小需要小于分块大小。'
  if (
    f.rrfVectorWeight != null &&
    f.rrfKeywordWeight != null &&
    Math.abs(f.rrfVectorWeight + f.rrfKeywordWeight - 1) > 0.000001
  )
    return 'RRF 两路权重之和需要等于 1。'
  if (f.rerankCandidateTopK != null && f.topK != null && f.rerankCandidateTopK < f.topK)
    return '融合候选 TopK 不能小于最终 TopK。'
  if (
    f.rerankCandidateTopK != null &&
    f.vectorTopK != null &&
    f.keywordTopK != null &&
    f.rerankCandidateTopK > f.vectorTopK + f.keywordTopK
  )
    return '融合候选数不能超过两路召回的总数。'
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
          按 Unicode 字符计数。修改分块大小、重叠大小或索引版本会触发重建。
        </p>
        <p
          v-if="group.title === '召回与融合'"
          class="mb-4 text-sm opacity-60"
        >
          两路召回经加权 RRF 融合，取候选进行重排序，最后去重并截取最终 TopK。
        </p>
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
              :disabled="group.title === '重排序' && !form.rerankEnabled"
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
