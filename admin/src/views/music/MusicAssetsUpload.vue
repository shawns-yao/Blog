<script setup lang="ts">
import { useMutation } from '@tanstack/vue-query'
import { NButton, useMessage } from 'naive-ui'
import { ref } from 'vue'

import { uploadMusicAsset } from '@/services/music'

const props = defineProps<{ songId: string; songTitle: string }>()
const emit = defineEmits<{ saved: [] }>()
const message = useMessage()
const coverInput = ref<HTMLInputElement>()
const lyricsInput = ref<HTMLInputElement>()
const upload = useMutation({
  mutationFn: ({ kind, file }: { kind: 'cover' | 'lyrics'; file: File }) =>
    uploadMusicAsset(props.songId, kind, file),
  onSuccess: (result) => {
    message.success(
      result.kind === 'cover'
        ? '封面已保存，刷新音乐室后查看'
        : '歌词已保存，重新展开歌词页面后查看',
    )
    emit('saved')
  },
  onError: (error: Error) => message.error(error.message),
})
function selected(event: Event, kind: 'cover' | 'lyrics') {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || upload.isPending.value) return
  const allowed = kind === 'cover' ? /\.(jpe?g|png)$/i : /\.(lrc|txt)$/i
  const maximum = kind === 'cover' ? 10 * 1024 * 1024 : 1024 * 1024
  if (!allowed.test(file.name) || file.size === 0 || file.size > maximum) {
    message.error(
      kind === 'cover'
        ? '请选择不超过 10 MB 的 JPG 或 PNG 封面'
        : '请选择不超过 1 MB 的 UTF-8 LRC 或 TXT 歌词',
    )
    return
  }
  upload.mutate({ kind, file })
}
</script>

<template>
  <div class="flex gap-2 whitespace-nowrap">
    <input
      ref="coverInput"
      type="file"
      accept=".jpg,.jpeg,.png"
      hidden
      :aria-label="`上传 ${songTitle} 的封面`"
      @change="selected($event, 'cover')"
    />
    <input
      ref="lyricsInput"
      type="file"
      accept=".lrc,.txt"
      hidden
      :aria-label="`上传 ${songTitle} 的歌词`"
      @change="selected($event, 'lyrics')"
    />
    <NButton
      size="small"
      :disabled="upload.isPending.value"
      :loading="upload.isPending.value && upload.variables.value?.kind === 'cover'"
      :aria-label="`上传 ${songTitle} 的封面`"
      title="JPG 或 PNG，最大 10 MB"
      @click="coverInput?.click()"
      >封面</NButton
    >
    <NButton
      size="small"
      :disabled="upload.isPending.value"
      :loading="upload.isPending.value && upload.variables.value?.kind === 'lyrics'"
      :aria-label="`上传 ${songTitle} 的歌词`"
      title="UTF-8 LRC 或 TXT，最大 1 MB"
      @click="lyricsInput?.click()"
      >歌词</NButton
    >
  </div>
</template>
