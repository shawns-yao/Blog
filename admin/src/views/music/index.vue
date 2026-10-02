<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { NAlert, NButton, NCard, NInput, NSwitch, NTag, useMessage } from 'naive-ui'
import { computed, ref, watch } from 'vue'

import { PageHeader, ScrollContainer } from '@/components'
import { getMusicCatalog, getMusicStatus, scanMusic, setMusicVisibility } from '@/services/music'

import MusicAssetsUpload from './MusicAssetsUpload.vue'
import MusicUploader from './MusicUploader.vue'

defineOptions({ name: 'MusicManagement' })
const client = useQueryClient()
const message = useMessage()
const busy = ref(false)
const input = ref('')
const search = ref('')
const offset = ref(0)
const status = useQuery({
  queryKey: ['music-admin', 'status'],
  queryFn: getMusicStatus,
  retry: false,
  refetchOnWindowFocus: true,
  refetchInterval: (query) => (query.state.data?.scan?.scanning ? 3000 : false),
})
const available = computed(() => !!status.data.value?.available)
const catalog = useQuery({
  queryKey: ['music-admin', 'catalog', search, offset],
  queryFn: () => getMusicCatalog(search.value, offset.value),
  enabled: available,
  retry: false,
})
const scan = useMutation({
  mutationFn: scanMusic,
  onSuccess: async () => {
    message.success('已请求扫描，请在曲库确认入库结果')
    await status.refetch()
    await catalog.refetch()
  },
  onError: (error: Error) => message.error(error.message),
})
const visibility = useMutation({
  mutationFn: ({ id, public: isPublic }: { id: string; public: boolean }) =>
    setMusicVisibility(id, isPublic),
  onSuccess: async (song) => {
    message.success(song.public ? '已设为公开试听' : '已取消公开试听')
    await catalog.refetch()
  },
  onError: (error: Error) => message.error(error.message),
})
watch(
  () => status.data.value?.scan?.scanning,
  (scanning, wasScanning) => {
    if (wasScanning && !scanning) void catalog.refetch()
  },
)
function refresh() {
  void client.invalidateQueries({ queryKey: ['music-admin'] })
}
function submitSearch() {
  search.value = input.value.trim()
  offset.value = 0
}
</script>

<template>
  <ScrollContainer wrapper-class="flex flex-col gap-y-4">
    <NCard :bordered="false">
      <PageHeader
        title="音乐管理"
        icon="ph--music-notes"
        description="上传音乐，扫描入库，在前台音乐室播放。"
      >
        <template #badge
          ><NTag
            :type="available ? 'success' : 'warning'"
            :bordered="false"
            >{{ available ? '服务已连接' : '服务未连接' }}</NTag
          ></template
        >
        <template #actions
          ><NButton @click="refresh">刷新</NButton
          ><NButton
            :disabled="!available || busy || !!status.data.value?.scan?.scanning"
            :loading="scan.isPending.value"
            @click="scan.mutate()"
            >扫描曲库</NButton
          ></template
        >
      </PageHeader>
      <NAlert
        v-if="status.error.value || status.data.value?.message"
        type="warning"
        class="mt-4"
        >{{ status.error.value?.message || status.data.value?.message }}</NAlert
      >
      <p
        v-if="status.data.value?.scan"
        role="status"
        class="mt-4 text-sm text-gray-500"
      >
        {{ status.data.value.scan.scanning ? '正在扫描音乐文件…' : '扫描空闲' }} · 已扫描
        {{ status.data.value.scan.count }} 个文件 · 播放上限 {{ status.data.value.maxBitRate }} kbps
      </p>
      <p
        v-if="status.data.value && !status.data.value.configured"
        class="mt-4 text-sm text-gray-500"
      >
        请按 deploy/MUSIC.md 配置 Navidrome 和音乐专用目录，再启用音乐服务。
      </p>
    </NCard>
    <NCard
      title="批量上传"
      :bordered="false"
      ><MusicUploader
        :disabled="!available || scan.isPending.value || !!status.data.value?.scan?.scanning"
        :max-bytes="status.data.value?.maxUploadBytes || 104857600"
        @busy="(value) => (busy = value)"
        @completed="scan.mutate()"
    /></NCard>
    <NCard
      title="已入库曲库"
      :bordered="false"
    >
      <form
        class="mb-4 flex gap-3"
        @submit.prevent="submitSearch"
      >
        <NInput
          v-model:value="input"
          placeholder="搜索歌曲、艺术家或专辑"
          :maxlength="100"
          aria-label="搜索曲库"
        /><NButton
          attr-type="submit"
          :disabled="!available"
          >搜索</NButton
        >
      </form>
      <p
        v-if="catalog.error.value"
        role="alert"
        class="py-4 text-red-600"
      >
        {{ catalog.error.value.message }}
      </p>
      <p
        v-else-if="catalog.isFetching.value"
        role="status"
        class="py-4 text-gray-500"
      >
        正在加载曲库…
      </p>
      <div
        v-else-if="catalog.data.value?.songs.length"
        class="overflow-x-auto"
      >
        <table class="w-full text-left text-sm">
          <thead>
            <tr class="border-b border-gray-200 text-gray-500 dark:border-gray-700">
              <th class="py-3 font-normal">歌曲</th>
              <th class="py-3 font-normal">艺术家</th>
              <th class="py-3 font-normal">专辑</th>
              <th class="py-3 font-normal">公开试听</th>
              <th class="py-3 font-normal">封面与歌词</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="song in catalog.data.value.songs"
              :key="song.id"
              class="border-b border-gray-100 dark:border-gray-800"
            >
              <td class="py-3 pr-4">{{ song.title }}</td>
              <td class="py-3 pr-4">{{ song.artist || '未知艺术家' }}</td>
              <td class="py-3">{{ song.album || '未归属专辑' }}</td>
              <td class="py-3">
                <NSwitch
                  :value="song.public"
                  :disabled="visibility.isPending.value"
                  :loading="
                    visibility.isPending.value && visibility.variables.value?.id === song.id
                  "
                  :aria-label="`公开试听：${song.title}`"
                  @update:value="
                    (value: boolean) => visibility.mutate({ id: song.id, public: value })
                  "
                />
              </td>
              <td class="py-3 pl-4">
                <MusicAssetsUpload
                  :song-id="song.id"
                  :song-title="song.title"
                  @saved="catalog.refetch()"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p
        v-else
        class="py-8 text-gray-500"
      >
        {{ search ? '没有找到匹配歌曲' : '曲库还没有歌曲' }}
      </p>
      <div
        v-if="offset > 0 || catalog.data.value?.hasMore"
        class="mt-4 flex justify-between"
      >
        <NButton
          :disabled="offset === 0 || catalog.isFetching.value"
          @click="offset = Math.max(0, offset - 30)"
          >上一页</NButton
        ><NButton
          :disabled="!catalog.data.value?.hasMore || catalog.isFetching.value"
          @click="offset += 30"
          >下一页</NButton
        >
      </div>
    </NCard>
  </ScrollContainer>
</template>
