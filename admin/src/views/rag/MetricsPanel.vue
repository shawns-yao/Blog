<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { NAlert, NButton, NDataTable, NDescriptions, NDescriptionsItem, NSpin } from 'naive-ui'
import { computed } from 'vue'

import { getRagMetrics } from '@/services/rag'

import { duration, failureLabel } from './format'

import type { DataTableColumns } from 'naive-ui'

const props = defineProps<{ active: boolean }>()
const metrics = useQuery({
  queryKey: ['rag', 'metrics'],
  queryFn: getRagMetrics,
  retry: false,
  enabled: computed(() => props.active),
  refetchInterval: 30000,
})
const failureColumns: DataTableColumns<{ reason: string; count: number }> = [
  { title: '不可用原因', key: 'reason', render: (row) => failureLabel(row.reason) },
  { title: '请求次数', key: 'count', width: 120 },
]
</script>

<template>
  <div class="space-y-6 pt-3">
    <NAlert
      v-if="metrics.isError.value"
      type="error"
    >
      {{ metrics.error.value?.message }}
      <div class="mt-2">
        <NButton
          size="small"
          @click="() => metrics.refetch()"
          >重试</NButton
        >
      </div>
    </NAlert>
    <NSpin
      v-else-if="metrics.isPending.value"
      class="block! py-12"
    />
    <template v-else-if="metrics.data.value">
      <div>
        <h2 class="mb-4 text-base font-medium">近 {{ metrics.data.value.days }} 天请求结果</h2>
        <NAlert
          v-if="!metrics.data.value.requests"
          type="info"
          class="mb-4"
          >尚未记录问答请求。</NAlert
        >
        <NDescriptions
          label-placement="top"
          :column="2"
        >
          <NDescriptionsItem label="问答请求">{{ metrics.data.value.requests }}</NDescriptionsItem>
          <NDescriptionsItem label="返回回答">{{ metrics.data.value.answered }}</NDescriptionsItem>
          <NDescriptionsItem label="依据不足">{{
            metrics.data.value.noEvidence
          }}</NDescriptionsItem>
          <NDescriptionsItem label="服务不可用">{{
            metrics.data.value.unavailable
          }}</NDescriptionsItem>
          <NDescriptionsItem label="主通道失败">{{
            metrics.data.value.primaryFailures
          }}</NDescriptionsItem>
          <NDescriptionsItem label="兜底通道调用">{{
            metrics.data.value.fallbackRequests
          }}</NDescriptionsItem>
          <NDescriptionsItem label="重排序降级">{{
            metrics.data.value.rerankDegraded
          }}</NDescriptionsItem>
        </NDescriptions>
        <p class="mt-3 text-sm opacity-60">
          请求结果用于运行监控；回答准确率与检索效果需通过公开数据集单独验收。
        </p>
      </div>
      <div class="border-t border-current/10 pt-5">
        <h2 class="mb-4 text-base font-medium">平均耗时</h2>
        <NDescriptions
          label-placement="top"
          :column="2"
        >
          <NDescriptionsItem label="问答处理">{{
            duration(metrics.data.value.avgDurationMs)
          }}</NDescriptionsItem>
          <NDescriptionsItem label="问题嵌入">{{
            duration(metrics.data.value.avgEmbeddingMs)
          }}</NDescriptionsItem>
          <NDescriptionsItem label="检索">{{
            duration(metrics.data.value.avgRetrievalMs)
          }}</NDescriptionsItem>
          <NDescriptionsItem label="重排序">{{
            duration(metrics.data.value.avgRerankMs)
          }}</NDescriptionsItem>
          <NDescriptionsItem label="回答生成（含兜底）">{{
            duration(metrics.data.value.avgGenerationMs)
          }}</NDescriptionsItem>
        </NDescriptions>
        <p class="mt-3 text-sm opacity-60">
          每个阶段按实际执行次数统计，包含失败请求；未执行的阶段显示为 —。
          仅统计合法问答请求，耗时不含指标写入。
        </p>
      </div>
      <div class="border-t border-current/10 pt-5">
        <h2 class="mb-4 text-base font-medium">服务不可用原因</h2>
        <NDataTable
          :columns="failureColumns"
          :data="metrics.data.value.failures"
          :bordered="false"
        />
      </div>
    </template>
  </div>
</template>
