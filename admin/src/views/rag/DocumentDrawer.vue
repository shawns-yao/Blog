<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { useWindowSize } from '@vueuse/core'
import {
  NAlert,
  NButton,
  NCode,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NPagination,
  NSpin,
  NTag,
} from 'naive-ui'
import { computed, ref, watch } from 'vue'

import { getRagChunks } from '@/services/rag'

import { statusLabels, statusType } from './format'

import type { RagDocument } from '@/services/rag'

const props = defineProps<{ show: boolean; document: RagDocument | null }>()
const emit = defineEmits<{ 'update:show': [value: boolean] }>()
const { width } = useWindowSize()
const page = ref(1)
const id = computed(() => props.document?.momentId ?? 0)
const enabled = computed(() => props.show && id.value > 0)
const chunks = useQuery({
  queryKey: ['rag', 'chunks', id, page],
  queryFn: ({ signal }) => getRagChunks(id.value, page.value, signal),
  enabled,
  retry: false,
  staleTime: 0,
})
watch(id, () => {
  page.value = 1
})
const kindLabels: Record<string, string> = {
  text: '文本',
  code: '代码',
  table: '表格',
  math: '公式',
}
</script>

<template>
  <NDrawer
    :show="show"
    :width="width < 768 ? '100%' : 680"
    @update:show="(value) => emit('update:show', value)"
  >
    <NDrawerContent
      :title="document?.title || '文档分块'"
      closable
    >
      <div
        v-if="document"
        class="mb-5 flex flex-wrap items-center gap-3"
      >
        <NTag
          :type="statusType(document.status)"
          :bordered="false"
          >{{ statusLabels[document.status] }}</NTag
        >
        <span class="text-sm opacity-60"
          >当前有效索引 · {{ chunks.data.value?.total ?? '—' }} 个分块</span
        >
      </div>
      <NAlert
        v-if="chunks.isError.value"
        type="error"
      >
        {{ chunks.error.value?.message }}
        <div class="mt-2">
          <NButton
            size="small"
            @click="() => chunks.refetch()"
            >重试</NButton
          >
        </div>
      </NAlert>
      <NSpin
        v-else-if="chunks.isPending.value"
        class="block! py-12"
      />
      <NEmpty
        v-else-if="!chunks.data.value?.items.length"
        description="当前没有可参与检索的分块"
        class="py-12"
      />
      <div
        v-else
        class="divide-y divide-current/10"
      >
        <article
          v-for="chunk in chunks.data.value.items"
          :key="chunk.seq"
          class="py-5 first:pt-0"
        >
          <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
            <h2 class="font-medium">分块 {{ chunk.seq + 1 }}</h2>
            <span class="text-xs opacity-60"
              >{{ kindLabels[chunk.kind] || chunk.kind }} · 字符范围 {{ chunk.start }}–{{
                chunk.end
              }}</span
            >
          </div>
          <p
            v-if="chunk.contextHeader"
            class="mb-3 text-sm whitespace-pre-wrap opacity-60"
          >
            {{ chunk.contextHeader }}
          </p>
          <NCode
            :code="chunk.content"
            language="text"
            word-wrap
          />
        </article>
      </div>
      <template #footer>
        <div class="flex w-full flex-wrap items-center justify-between gap-3">
          <NPagination
            v-model:page="page"
            :page-size="10"
            :item-count="chunks.data.value?.total ?? 0"
            :page-slot="5"
          />
          <NButton @click="emit('update:show', false)">关闭</NButton>
        </div>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
