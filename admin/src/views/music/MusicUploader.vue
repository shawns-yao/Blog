<script setup lang="ts">
import { NAlert, NButton, NProgress } from 'naive-ui'
import { computed, onBeforeUnmount, ref } from 'vue'

import { uploadMusic } from '@/services/music'

const props = defineProps<{ disabled: boolean; maxBytes: number }>()
const emit = defineEmits<{ completed: []; busy: [value: boolean] }>()
type Task = {
  file: File
  percent: number
  status: 'pending' | 'uploading' | 'validating' | 'saved' | 'duplicate' | 'failed'
  error: string
}
const tasks = ref<Task[]>([])
const busy = ref(false)
const input = ref<HTMLInputElement>()
let controller: AbortController | null = null
const pending = computed(() =>
  tasks.value.some((task) => task.status === 'pending' || task.status === 'failed'),
)
const labels = {
  pending: '等待上传',
  uploading: '上传中',
  validating: '校验音频中',
  saved: '已保存，等待扫描',
  duplicate: '文件已存在',
  failed: '上传失败',
}

function select(event: Event) {
  const target = event.target as HTMLInputElement
  tasks.value = Array.from(target.files || []).map((file) => ({
    file,
    percent: 0,
    status: 'pending',
    error: '',
  }))
  target.value = ''
}

async function upload() {
  if (busy.value || props.disabled) return
  controller = new AbortController()
  busy.value = true
  emit('busy', true)
  let saved = 0
  try {
    for (const task of tasks.value) {
      if (task.status === 'saved' || task.status === 'duplicate') continue
      task.error = ''
      if (
        !/\.(mp3|flac|m4a)$/i.test(task.file.name) ||
        task.file.size > props.maxBytes ||
        !task.file.size
      ) {
        task.status = 'failed'
        task.error = '请选择未超过大小限制的 MP3、FLAC 或 M4A 文件'
        continue
      }
      task.status = 'uploading'
      task.percent = 0
      try {
        const result = await uploadMusic(
          task.file,
          (percent) => {
            task.percent = percent
            if (percent === 100) task.status = 'validating'
          },
          controller.signal,
        )
        task.percent = 100
        task.status = result.duplicate ? 'duplicate' : 'saved'
        saved++
      } catch (error) {
        if (controller.signal.aborted) break
        task.status = 'failed'
        task.error = error instanceof Error ? error.message : '上传失败'
      }
    }
    if (saved && !controller.signal.aborted) emit('completed')
  } finally {
    busy.value = false
    emit('busy', false)
  }
}

onBeforeUnmount(() => controller?.abort())
</script>

<template>
  <div class="space-y-5">
    <NAlert
      type="info"
      :show-icon="false"
      >支持 MP3、FLAC、M4A，单文件上限
      {{ Number((maxBytes / 1024 / 1024).toFixed(2)) }}
      MB。批量上传完成后统一扫描，扫描完成后请在曲库确认歌曲。</NAlert
    >
    <div class="flex flex-wrap gap-3">
      <input
        ref="input"
        type="file"
        multiple
        accept=".mp3,.flac,.m4a"
        class="hidden"
        aria-label="选择音乐文件"
        @change="select"
      />
      <NButton
        :disabled="disabled || busy"
        @click="input?.click()"
        >选择音乐文件</NButton
      >
      <NButton
        type="primary"
        :disabled="disabled || !pending"
        :loading="busy"
        @click="upload"
        >{{
          tasks.some((task) => task.status === 'failed') ? '重试未成功文件' : '上传音乐'
        }}</NButton
      >
    </div>
    <ul class="divide-y divide-gray-200 dark:divide-gray-700">
      <li
        v-for="(task, index) in tasks"
        :key="index"
        class="py-4"
      >
        <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
          <span class="text-sm break-all">{{ task.file.name }}</span
          ><span class="text-xs text-gray-500">{{ labels[task.status] }}</span>
        </div>
        <NProgress
          type="line"
          :percentage="task.percent"
          :status="
            task.status === 'failed'
              ? 'error'
              : task.status === 'saved' || task.status === 'duplicate'
                ? 'success'
                : 'default'
          "
          :show-indicator="true"
        />
        <p
          v-if="task.error"
          role="alert"
          class="mt-2 text-xs text-red-600"
        >
          {{ task.error }}
        </p>
      </li>
    </ul>
  </div>
</template>
