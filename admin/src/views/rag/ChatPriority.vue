<script setup lang="ts">
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { NAlert, NButton, NTag } from 'naive-ui'
import { computed, ref, watch } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'

import { saveRagChatPriority } from '@/services/rag'

import type { RagChatChannel } from '@/services/rag'

const props = defineProps<{ channels: RagChatChannel[] }>()
const client = useQueryClient()
const draft = ref<RagChatChannel[]>([])
const saved = ref('[]')
const saveError = ref('')
const announcement = ref('')
const names: Record<string, string> = {
  gpt: 'GPT',
  grok: 'Grok',
  gemini: 'Gemini',
  opencode_go: 'OpenCode Go',
  deepseek: 'DeepSeek 官方',
}
const order = computed(() => draft.value.map((channel) => channel.name))
const dirty = computed(() => JSON.stringify(order.value) !== saved.value)
const fallback = computed(() => props.channels.find((channel) => channel.default))

function restore(channels = props.channels, clearAnnouncement = true) {
  draft.value = channels.filter((channel) => !channel.default).map((channel) => ({ ...channel }))
  saved.value = JSON.stringify(order.value)
  saveError.value = ''
  if (clearAnnouncement) announcement.value = ''
}

watch(
  () => props.channels,
  (channels) => {
    if (!dirty.value) restore(channels, false)
  },
  { immediate: true },
)

const mutation = useMutation({
  mutationFn: saveRagChatPriority,
  onSuccess: (value) => {
    restore(value.chatChannels)
    client.setQueryData(['rag', 'settings'], value)
    announcement.value = '优先级已保存，新请求将按此顺序调用。'
  },
  onError: (error: Error) => {
    saveError.value = error.message
  },
})

function move(index: number, direction: number) {
  const target = index + direction
  if (target < 0 || target >= draft.value.length || mutation.isPending.value) return
  const next = [...draft.value]
  const channel = next.splice(index, 1)[0]!
  next.splice(target, 0, channel)
  draft.value = next
  announcement.value = `${names[channel.name] || channel.name} 已移至第 ${target + 1} 优先级，待保存。`
}

function submit() {
  if (!dirty.value || mutation.isPending.value) return
  saveError.value = ''
  announcement.value = ''
  mutation.mutate([...order.value])
}
</script>

<template>
  <section
    class="border-b border-current/10 pb-6"
    aria-labelledby="rag-chat-priority-heading"
  >
    <h2
      id="rag-chat-priority-heading"
      class="mb-2 text-base font-medium"
    >
      语言模型优先级
    </h2>
    <p class="mb-4 text-sm opacity-60">
      拖动左侧手柄调整调用顺序；失败时依次降级，default 固定兜底。
    </p>
    <VueDraggable
      v-model="draft"
      tag="ol"
      handle=".rag-priority-handle"
      :animation="150"
      :disabled="mutation.isPending.value"
      :delay="150"
      :delay-on-touch-only="true"
      ghost-class="opacity-40"
      class="space-y-2"
      aria-label="可调整语言模型优先级"
    >
      <li
        v-for="(channel, index) in draft"
        :key="channel.name"
        :data-channel="channel.name"
        class="flex items-center gap-3 rounded border border-current/10 px-3 py-3"
      >
        <span
          class="rag-priority-handle shrink-0 cursor-grab touch-none p-1 opacity-60 active:cursor-grabbing"
          aria-hidden="true"
        >
          <span class="iconify size-5 ph--dots-six-vertical" />
        </span>
        <span
          class="w-5 shrink-0 text-center tabular-nums"
          aria-label="优先级"
          >{{ index + 1 }}</span
        >
        <div class="min-w-0 flex-1">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-medium">{{ names[channel.name] || channel.name }}</span>
            <NTag
              size="small"
              :bordered="false"
              :type="channel.configured ? 'success' : 'warning'"
            >
              {{ channel.configured ? '已配置' : '未配置' }}
            </NTag>
          </div>
          <p class="mt-1 text-sm break-words opacity-60">
            {{ channel.model
            }}<template v-if="channel.reasoningEffort"> · {{ channel.reasoningEffort }}</template>
          </p>
        </div>
        <div class="flex shrink-0 gap-1">
          <NButton
            quaternary
            circle
            size="small"
            :disabled="index === 0 || mutation.isPending.value"
            :aria-label="`上移 ${names[channel.name] || channel.name}`"
            @click="move(index, -1)"
          >
            <template #icon><span class="iconify ph--arrow-up" /></template>
          </NButton>
          <NButton
            quaternary
            circle
            size="small"
            :disabled="index === draft.length - 1 || mutation.isPending.value"
            :aria-label="`下移 ${names[channel.name] || channel.name}`"
            @click="move(index, 1)"
          >
            <template #icon><span class="iconify ph--arrow-down" /></template>
          </NButton>
        </div>
      </li>
    </VueDraggable>
    <div
      v-if="fallback"
      data-channel="default"
      class="mt-3 flex items-center gap-3 rounded bg-neutral-500/5 px-4 py-3"
    >
      <span
        class="iconify size-5 shrink-0 opacity-60 ph--lock-simple"
        aria-hidden="true"
      />
      <div class="min-w-0 flex-1">
        <div class="flex flex-wrap items-center gap-2">
          <NTag
            size="small"
            :bordered="false"
            >default</NTag
          >
          <span class="font-medium">{{ names[fallback.name] || fallback.name }}</span>
          <span class="text-sm opacity-60">{{ fallback.configured ? '已配置' : '未配置' }}</span>
        </div>
        <p class="mt-1 text-sm break-words opacity-60">{{ fallback.model }}</p>
      </div>
      <span class="shrink-0 text-sm opacity-60">固定兜底</span>
    </div>
    <NAlert
      v-if="saveError"
      type="error"
      class="mt-4"
      >{{ saveError }}</NAlert
    >
    <div class="mt-4 flex flex-wrap items-center gap-3">
      <NButton
        type="primary"
        :disabled="!dirty"
        :loading="mutation.isPending.value"
        @click="submit"
        >保存优先级</NButton
      >
      <NButton
        :disabled="!dirty || mutation.isPending.value"
        @click="restore()"
        >还原顺序</NButton
      >
      <span
        v-if="dirty"
        class="text-sm opacity-60"
        >有未保存的调整</span
      >
    </div>
    <p
      class="mt-2 text-sm opacity-60"
      role="status"
      aria-live="polite"
    >
      {{ announcement }}
    </p>
  </section>
</template>
