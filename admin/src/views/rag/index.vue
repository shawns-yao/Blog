<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import {
  NAlert,
  NButton,
  NCard,
  NPopconfirm,
  NSkeleton,
  NTabPane,
  NTabs,
  NTag,
  useMessage,
} from 'naive-ui'
import { ref } from 'vue'

import { PageHeader, ScrollContainer } from '@/components'
import { getRagIndex, getRagSettings, reindexRag } from '@/services/rag'

import IndexDocuments from './IndexDocuments.vue'
import MetricsPanel from './MetricsPanel.vue'
import SettingsPanel from './SettingsPanel.vue'

defineOptions({ name: 'RagManagement' })

const tab = ref('documents')
const client = useQueryClient()
const message = useMessage()
const settings = useQuery({ queryKey: ['rag', 'settings'], queryFn: getRagSettings, retry: false })
const index = useQuery({
  queryKey: ['rag', 'index'],
  queryFn: getRagIndex,
  retry: false,
  refetchInterval: 10000,
})
const reindex = useMutation({
  mutationFn: reindexRag,
  onSuccess: async () => {
    message.success('已提交重建任务')
    await client.invalidateQueries({ queryKey: ['rag'] })
  },
  onError: (error: Error) => message.error(error.message),
})

const counts = [
  { key: 'ready', label: '已索引' },
  { key: 'chunks', label: '有效分块' },
  { key: 'pending', label: '等待索引' },
  { key: 'running', label: '索引中' },
  { key: 'outdated', label: '索引过期' },
  { key: 'unindexed', label: '尚未入库' },
  { key: 'failed', label: '索引失败' },
  { key: 'excluded', label: '不参与检索' },
] as const

function refresh() {
  void client.invalidateQueries({ queryKey: ['rag'] })
}
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <NCard :bordered="false">
      <PageHeader
        title="RAG 知识库"
        icon="ph--database"
      >
        <template #badge>
          <NTag
            v-if="settings.data.value"
            :type="settings.data.value.enabled ? 'success' : 'default'"
            :bordered="false"
          >
            {{ settings.data.value.enabled ? '已启用' : '未启用' }}
          </NTag>
        </template>
        <template #actions>
          <NButton @click="refresh">刷新</NButton>
          <NPopconfirm @positive-click="() => reindex.mutate()">
            <template #trigger>
              <NButton
                :disabled="!settings.data.value"
                :loading="reindex.isPending.value"
                >重建全部索引</NButton
              >
            </template>
            重建全部公开文章与手记的分块及向量，会产生模型请求费用。是否继续？
          </NPopconfirm>
        </template>
      </PageHeader>
      <NAlert
        v-if="settings.isError.value || index.isError.value"
        type="error"
        class="mt-4"
      >
        {{ settings.error.value?.message || index.error.value?.message }}
      </NAlert>
      <NAlert
        v-else-if="settings.data.value && !settings.data.value.enabled"
        type="info"
        class="mt-4"
      >
        问答暂未启用，配置与索引记录仍可查看。重建任务会在启用后处理。
      </NAlert>
      <dl
        class="mt-6 grid grid-cols-2 gap-x-6 gap-y-5 sm:grid-cols-4 xl:grid-cols-8"
        aria-label="索引状态"
      >
        <div
          v-for="item in counts"
          :key="item.key"
        >
          <dt class="text-sm opacity-60">{{ item.label }}</dt>
          <dd class="mt-1 text-2xl font-medium tabular-nums">
            <NSkeleton
              v-if="index.isPending.value"
              text
              :width="48"
            />
            <template v-else>{{ index.data.value?.[item.key] ?? '—' }}</template>
          </dd>
        </div>
      </dl>
    </NCard>
    <NCard :bordered="false">
      <NTabs
        v-model:value="tab"
        type="line"
        animated
      >
        <NTabPane
          name="documents"
          tab="索引文档"
          display-directive="show:lazy"
        >
          <IndexDocuments :active="tab === 'documents'" />
        </NTabPane>
        <NTabPane
          name="settings"
          tab="检索配置"
          display-directive="show:lazy"
        >
          <SettingsPanel :settings="settings.data.value" />
        </NTabPane>
        <NTabPane
          name="metrics"
          tab="运行指标"
          display-directive="show:lazy"
        >
          <MetricsPanel :active="tab === 'metrics'" />
        </NTabPane>
      </NTabs>
    </NCard>
  </ScrollContainer>
</template>
