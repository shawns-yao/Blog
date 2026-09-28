<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import {
  NAlert,
  NButton,
  NDataTable,
  NInput,
  NPagination,
  NSelect,
  NSpace,
  NTag,
  useMessage,
} from 'naive-ui'
import { computed, h, ref, watch } from 'vue'

import { listRagDocuments, reindexRagDocument } from '@/services/rag'

import DocumentDrawer from './DocumentDrawer.vue'
import { dateTime, duration, failureLabel, statusLabels, statusType } from './format'

import type { RagDocument } from '@/services/rag'
import type { DataTableColumns } from 'naive-ui'

const props = defineProps<{ active: boolean }>()
const page = ref(1)
const pageSize = ref(20)
const searchText = ref('')
const search = ref('')
const status = ref<string | null>(null)
const contentKind = ref<string | null>(null)
const selected = ref<RagDocument | null>(null)
const showDrawer = ref(false)
const client = useQueryClient()
const message = useMessage()

const documents = useQuery({
  queryKey: ['rag', 'documents', page, pageSize, search, status, contentKind],
  queryFn: ({ signal }) =>
    listRagDocuments(
      {
        page: page.value,
        pageSize: pageSize.value,
        search: search.value || undefined,
        status: status.value || undefined,
        contentKind: contentKind.value || undefined,
      },
      signal,
    ),
  retry: false,
  enabled: computed(() => props.active),
  refetchInterval: 10000,
})

watch([pageSize, status, contentKind, search], () => {
  page.value = 1
})

const rebuild = useMutation({
  mutationFn: reindexRagDocument,
  onSuccess: async () => {
    message.success('已提交文档索引任务')
    await client.invalidateQueries({ queryKey: ['rag'] })
  },
  onError: (error: Error) => message.error(error.message),
})

const statusOptions = Object.entries(statusLabels).map(([value, label]) => ({ value, label }))
const kindOptions = [
  { label: '文章', value: 'article' },
  { label: '手记', value: 'note' },
]

const columns: DataTableColumns<RagDocument> = [
  { title: '文档', key: 'title', minWidth: 240, ellipsis: { tooltip: true } },
  {
    title: '类型',
    key: 'contentKind',
    width: 85,
    render: (row) =>
      row.contentKind === 'article' ? '文章' : row.contentKind === 'note' ? '手记' : '其他',
  },
  {
    title: '索引状态',
    key: 'status',
    width: 120,
    render: (row) =>
      h(
        NTag,
        { type: statusType(row.status), size: 'small', bordered: false },
        { default: () => statusLabels[row.status] },
      ),
  },
  { title: '有效分块', key: 'chunks', width: 90 },
  { title: '尝试次数', key: 'attempts', width: 90 },
  { title: '最近索引', key: 'indexedAt', width: 180, render: (row) => dateTime(row.indexedAt) },
  {
    title: '索引耗时',
    key: 'indexDurationMs',
    width: 105,
    render: (row) => duration(row.indexDurationMs),
  },
  {
    title: '失败原因',
    key: 'lastError',
    minWidth: 180,
    ellipsis: { tooltip: true },
    render: (row) => failureLabel(row.lastError),
  },
  {
    title: '操作',
    key: 'actions',
    width: 160,
    fixed: 'right',
    render: (row) =>
      h(
        NSpace,
        { size: 'small', wrap: false },
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                onClick: () => {
                  selected.value = row
                  showDrawer.value = true
                },
              },
              { default: () => '分块' },
            ),
            h(
              NButton,
              {
                size: 'small',
                disabled:
                  !row.published ||
                  !['article', 'note'].includes(row.contentKind) ||
                  rebuild.isPending.value,
                loading: rebuild.isPending.value && rebuild.variables.value === row.momentId,
                onClick: () => rebuild.mutate(row.momentId),
              },
              { default: () => (row.status === 'failed' ? '重试' : '重建') },
            ),
          ],
        },
      ),
  },
]

function submitSearch() {
  search.value = searchText.value.trim()
}
</script>

<template>
  <div class="space-y-4 pt-3">
    <form
      class="flex flex-wrap items-center gap-3"
      @submit.prevent="submitSearch"
    >
      <NInput
        v-model:value="searchText"
        placeholder="搜索文档标题"
        aria-label="文档标题"
        clearable
        :maxlength="100"
        class="w-full! sm:w-64!"
      />
      <NSelect
        v-model:value="contentKind"
        :options="kindOptions"
        placeholder="全部类型"
        aria-label="文档类型"
        clearable
        class="w-32!"
      />
      <NSelect
        v-model:value="status"
        :options="statusOptions"
        placeholder="全部状态"
        aria-label="索引状态"
        clearable
        class="w-36!"
      />
      <NButton attr-type="submit">搜索</NButton>
    </form>
    <NAlert
      v-if="documents.isError.value"
      type="error"
    >
      {{ documents.error.value?.message }}
      <div class="mt-2">
        <NButton
          size="small"
          @click="() => documents.refetch()"
          >重试</NButton
        >
      </div>
    </NAlert>
    <NDataTable
      remote
      :columns="columns"
      :data="documents.data.value?.items ?? []"
      :loading="documents.isPending.value"
      :bordered="false"
      :row-key="(row: RagDocument) => row.momentId"
      :scroll-x="1370"
    />
    <div class="flex flex-wrap items-center justify-between gap-3">
      <span class="text-sm opacity-60">{{
        documents.data.value ? `共 ${documents.data.value.total} 份文档` : '—'
      }}</span>
      <NPagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :item-count="documents.data.value?.total ?? 0"
        show-size-picker
        :page-sizes="[10, 20, 50]"
        :page-slot="5"
      />
    </div>
    <DocumentDrawer
      v-model:show="showDrawer"
      :document="selected"
    />
  </div>
</template>
