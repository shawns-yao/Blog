<script setup lang="ts">
import { NAlert, NButton, NProgress } from 'naive-ui'
import { computed, onBeforeUnmount, ref } from 'vue'

import { getMusicUploads, uploadCompanionLyrics, uploadMusic } from '@/services/music'

import type { MusicUpload } from '@/services/music'

const props = defineProps<{ disabled: boolean; maxBytes: number }>()
const emit = defineEmits<{ completed: []; busy: [value: boolean] }>()
type Task = {
  file: File
  percent: number
  status: 'pending' | 'uploading' | 'validating' | 'duplicate' | MusicUpload['state']
  error: string
  jobId?: string
  duplicate?: boolean
  lyrics?: File
  lyricsSaved?: boolean
  lyricsError: string
}
const tasks = ref<Task[]>([])
const busy = ref(false)
const input = ref<HTMLInputElement>()
let controller: AbortController | null = null
let pollTimer: ReturnType<typeof setTimeout> | undefined
let disposed = false
const pollError = ref('')
const selectionError = ref('')
const pending = computed(() =>
  tasks.value.some(
    (task) => task.status === 'pending' || task.status === 'failed' || task.lyricsError,
  ),
)
const labels = {
  pending: '等待上传',
  uploading: '上传中',
  validating: '服务器接收中',
  queued: '等待后台处理',
  processing: '正在准备播放缓存',
  scanning: '正在扫描入库',
  ready: '已入库',
  duplicate: '文件已存在',
  failed: '处理失败',
}

function startPolling() {
  if (disposed || pollTimer) return
  const active = tasks.value.filter(
    (task) => task.jobId && ['queued', 'processing', 'scanning'].includes(task.status),
  )
  if (!active.length) return
  pollTimer = setTimeout(async () => {
    pollTimer = undefined
    try {
      const results = await getMusicUploads(active.slice(0, 100).map((task) => task.jobId!))
      if (disposed) return
      pollError.value = ''
      let completed = false
      for (const result of results) {
        for (const task of tasks.value.filter((item) => item.jobId === result.id)) {
          completed ||= result.state === 'ready' && !['ready', 'duplicate'].includes(task.status)
          task.status = result.state === 'ready' && task.duplicate ? 'duplicate' : result.state
          task.error = result.error || ''
        }
      }
      if (completed) emit('completed')
    } catch (error) {
      if (!disposed)
        pollError.value = error instanceof Error ? error.message : '处理状态暂时无法读取'
    } finally {
      startPolling()
    }
  }, 3000)
}

function select(event: Event) {
  const target = event.target as HTMLInputElement
  const files = Array.from(target.files || [])
  target.value = ''
  selectionError.value = ''
  const audio = files.filter((file) => /\.(mp3|flac|m4a)$/i.test(file.name))
  const lyrics = files.filter((file) => /\.lrc$/i.test(file.name))
  const stem = (file: File) => file.name.replace(/\.[^.]+$/, '').normalize('NFC')
  if (!audio.length) {
    selectionError.value = '请同时选择音乐和同名 LRC；已入库歌曲可以在曲库中单独补传歌词'
    return
  }
  for (const lyric of lyrics) {
    const matches = audio.filter((file) => stem(file) === stem(lyric))
    if (matches.length > 1 || lyrics.filter((file) => stem(file) === stem(lyric)).length > 1) {
      selectionError.value = `“${lyric.name}”存在多个同名文件，请给不同版本使用不同文件名后重新选择`
      return
    }
  }
  const unmatched = lyrics.filter((lyric) => !audio.some((file) => stem(file) === stem(lyric)))
  if (unmatched.length) {
    selectionError.value = `未匹配歌词：${unmatched.map((file) => file.name).join('、')}。请使用与音频相同的文件名`
    return
  }
  const unsupported = files.filter((file) => !audio.includes(file) && !lyrics.includes(file))
  if (unsupported.length) {
    selectionError.value = '批量上传支持 MP3、FLAC、M4A 和同名 LRC；TXT 歌词请在曲库中单独补传'
    return
  }
  tasks.value = audio.map((file) => ({
    file,
    lyrics: lyrics.find((lyric) => stem(lyric) === stem(file)),
    percent: 0,
    status: 'pending',
    error: '',
    lyricsError: '',
  }))
}

async function saveLyrics(task: Task, signal: AbortSignal) {
  if (!task.lyrics || !task.jobId || task.lyricsSaved) return
  task.lyricsError = ''
  try {
    await uploadCompanionLyrics(task.jobId, task.lyrics, signal)
    task.lyricsSaved = true
  } catch (error) {
    if (!signal.aborted)
      task.lyricsError = error instanceof Error ? error.message : '歌词关联失败，请重试'
  }
}

async function upload() {
  if (busy.value || props.disabled) return
  controller = new AbortController()
  busy.value = true
  emit('busy', true)
  let saved = 0
  try {
    for (const task of tasks.value) {
      if (!['pending', 'failed'].includes(task.status)) {
        if (task.lyricsError) await saveLyrics(task, controller.signal)
        continue
      }
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
      if (task.lyrics && (!task.lyrics.size || task.lyrics.size > 1024 * 1024)) {
        task.status = 'failed'
        task.error = '配对的 LRC 歌词必须非空，且不超过 1 MB'
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
        task.jobId = result.id
        task.duplicate = result.duplicate
        task.status = result.duplicate && result.state === 'ready' ? 'duplicate' : result.state
        task.error = result.error || ''
        startPolling()
        await saveLyrics(task, controller.signal)
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

onBeforeUnmount(() => {
  disposed = true
  controller?.abort()
  clearTimeout(pollTimer)
})
</script>

<template>
  <div class="space-y-5">
    <NAlert
      type="info"
      :show-icon="false"
      >支持 MP3、FLAC、M4A，单文件上限
      {{ Number((maxBytes / 1024 / 1024).toFixed(2)) }}
      MB。可一次多选，同名 LRC
      自动关联；相同内容不重复入库，不同版本分别保留。上传后在后台处理。</NAlert
    >
    <div class="flex flex-wrap gap-3">
      <input
        ref="input"
        type="file"
        multiple
        accept=".mp3,.flac,.m4a,.lrc"
        class="hidden"
        aria-label="批量选择音乐和歌词文件"
        @change="select"
      />
      <NButton
        :disabled="disabled || busy"
        @click="input?.click()"
        >批量选择音乐与歌词</NButton
      >
      <NButton
        type="primary"
        :disabled="disabled || !pending"
        :loading="busy"
        @click="upload"
        >{{
          tasks.some((task) => task.status === 'failed' || task.lyricsError)
            ? '重试未成功文件'
            : '开始批量上传'
        }}</NButton
      >
    </div>
    <NAlert
      v-if="selectionError"
      type="warning"
      >{{ selectionError }}</NAlert
    >
    <NAlert
      v-if="pollError"
      type="warning"
      >{{ pollError }}</NAlert
    >
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
        <p
          v-if="task.lyrics"
          class="mb-2 text-xs text-gray-500"
        >
          歌词：{{ task.lyrics.name }}（{{ task.lyricsSaved ? '已关联' : '随音乐上传' }}）
        </p>
        <NProgress
          type="line"
          :percentage="task.percent"
          :status="
            task.status === 'failed'
              ? 'error'
              : task.status === 'ready' || task.status === 'duplicate'
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
        <p
          v-if="task.lyricsError"
          role="alert"
          class="mt-2 text-xs text-red-600"
        >
          {{ task.lyricsError }}
        </p>
      </li>
    </ul>
  </div>
</template>
