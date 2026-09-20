<script setup lang="ts">
import {
  NAlert,
  NButton,
  NDataTable,
  NDrawer,
  NDrawerContent,
  NEmpty,
  NImage,
  NPagination,
  NSpin,
} from 'naive-ui'
import { onMounted, ref } from 'vue'
import { getAlbum, listAlbums, type AlbumDetail, type AlbumListItem } from '@/services/albums'

defineOptions({ name: 'AlbumArchive' })
const items = ref<AlbumListItem[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const error = ref('')
const show = ref(false)
const detail = ref<AlbumDetail | null>(null)
const detailLoading = ref(false)
let requestId = 0

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await listAlbums({ page: page.value, pageSize: 20 })
    items.value = result.items
    total.value = result.total
  } catch (err) {
    error.value = err instanceof Error ? err.message : '相册加载失败'
  } finally {
    loading.value = false
  }
}
async function inspect(id: number) {
  const request = ++requestId
  show.value = true
  detail.value = null
  detailLoading.value = true
  error.value = ''
  try {
    const result = await getAlbum(id)
    if (request === requestId) detail.value = result
  } catch (err) {
    error.value = err instanceof Error ? err.message : '相册加载失败'
  } finally {
    if (request === requestId) detailLoading.value = false
  }
}
onMounted(load)
</script>

<template>
  <div class="space-y-4 p-6">
    <h1 class="text-lg font-medium">历史相册</h1>
    <NAlert
      type="info"
      :bordered="false"
      >旧相册仅供查阅，照片和原始信息保留。新的图片内容请通过手记发布。</NAlert
    >
    <NAlert
      v-if="error"
      type="error"
      >{{ error }}</NAlert
    >
    <NDataTable
      :loading="loading"
      :data="items"
      :row-key="(row: AlbumListItem) => row.id"
      :columns="[
        { title: '名称', key: 'title' },
        { title: '照片', key: 'photoCount' },
        { title: '创建时间', key: 'createdAt' },
      ]"
      :row-props="
        (row: AlbumListItem) => ({
          onClick: () => inspect(row.id),
          tabindex: 0,
          onKeydown: (event: KeyboardEvent) => {
            if (event.key === 'Enter') inspect(row.id)
          },
          style: 'cursor: pointer',
        })
      "
    />
    <NPagination
      v-model:page="page"
      :item-count="total"
      :page-size="20"
      @update:page="load"
    />
    <NButton @click="load">刷新</NButton>
    <NDrawer
      v-model:show="show"
      width="min(720px, 100vw)"
    >
      <NDrawerContent
        :title="detail?.title || '相册详情'"
        closable
      >
        <NSpin :show="detailLoading">
          <template v-if="detail">
            <p class="mb-4">{{ detail.description }}</p>
            <div class="grid grid-cols-2 gap-4">
              <figure
                v-for="photo in detail.photos"
                :key="photo.id"
              >
                <NImage
                  :src="photo.url"
                  :alt="photo.caption || '历史照片'"
                />
                <figcaption class="mt-2 text-sm">
                  {{ photo.caption || photo.description }}
                </figcaption>
              </figure>
            </div>
            <NEmpty
              v-if="!detail.photos.length"
              description="相册暂无照片"
            />
          </template>
        </NSpin>
      </NDrawerContent>
    </NDrawer>
  </div>
</template>
