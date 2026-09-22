<script setup lang="ts">
import {
  NButton,
  NCard,
  NDataTable,
  NFormItem,
  NInput,
  NSelect,
  NPopconfirm,
  NSpace,
  useMessage,
} from 'naive-ui'
import { computed, h, onMounted, reactive, ref } from 'vue'

import { FormModal, ScrollContainer } from '@/components'
import { createColumn, deleteColumn, listColumns, updateColumn } from '@/services/taxonomy'
import { formatDate } from '@/utils/format'

import type { ColumnItem } from '@/services/taxonomy'
import type { DataTableColumns } from 'naive-ui'

defineOptions({
  name: 'MomentColumnManagement',
})

const message = useMessage()
const loading = ref(false)
const saving = ref(false)
const items = ref<ColumnItem[]>([])
const editVisible = ref(false)
const editingId = ref<number | null>(null)
const formModel = reactive({
  name: '',
  shortUrl: '',
  parentId: null as number | null,
})
const parentOptions = computed(() =>
  items.value
    .filter((item) => !item.parentId && item.id !== editingId.value)
    .map((item) => ({ label: item.name, value: item.id })),
)
const hasChildren = computed(() =>
  items.value.some((item) => item.parentId === editingId.value && editingId.value !== null),
)

const columns: DataTableColumns<ColumnItem> = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '分类名称', key: 'name', minWidth: 200 },
  {
    title: '层级',
    key: 'parentId',
    width: 170,
    render: (row) =>
      row.parentId
        ? `二级 · ${items.value.find((item) => item.id === row.parentId)?.name ?? ''}`
        : '一级分类',
  },
  { title: '短链接', key: 'shortUrl', minWidth: 180 },
  {
    title: '更新时间',
    key: 'updatedAt',
    width: 180,
    render: (row) => formatDate(row.updatedAt),
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row) =>
      h(NSpace, { size: 'small' }, () => [
        h(
          NButton,
          { size: 'small', tertiary: true, onClick: () => openEdit(row) },
          { default: () => '编辑' },
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => handleDelete(row) },
          {
            trigger: () =>
              h(
                NButton,
                { size: 'small', type: 'error', secondary: true },
                { default: () => '删除' },
              ),
            default: () => '确认归档该分类？有关联文章或二级分类时不可操作，文章不会删除。',
          },
        ),
      ]),
  },
]

const modalTitle = ref('新建专栏')

async function fetchData() {
  loading.value = true
  try {
    items.value = await listColumns()
  } catch (error: any) {
    message.error(error?.message || '获取专栏列表失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  modalTitle.value = '新建专栏'
  editingId.value = null
  formModel.name = ''
  formModel.shortUrl = ''
  formModel.parentId = null
  editVisible.value = true
}

function openEdit(row: ColumnItem) {
  modalTitle.value = '编辑专栏'
  editingId.value = row.id
  formModel.name = row.name
  formModel.shortUrl = row.shortUrl
  formModel.parentId = row.parentId
  editVisible.value = true
}

async function handleSubmit() {
  const name = formModel.name.trim()
  const shortUrl = formModel.shortUrl.trim()
  if (!name) {
    message.warning('请输入专栏名称')
    return
  }
  if (!shortUrl) {
    message.warning('请输入专栏短链接')
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await updateColumn(editingId.value, { name, shortUrl, parentId: formModel.parentId })
      message.success('专栏已更新')
    } else {
      await createColumn({ name, shortUrl, parentId: formModel.parentId })
      message.success('专栏已创建')
    }
    editVisible.value = false
    await fetchData()
  } catch (error: any) {
    message.error(error?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: ColumnItem) {
  try {
    await deleteColumn(row.id)
    message.success('删除成功')
    await fetchData()
  } catch (error: any) {
    message.error(error?.message || '删除失败')
  }
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <ScrollContainer
    wrapper-class="p-4"
    :scrollbar-props="{ trigger: 'none' }"
  >
    <NCard title="图书馆分类管理">
      <template #header-extra>
        <NButton
          type="primary"
          @click="openCreate"
          >新建分类</NButton
        >
      </template>

      <NDataTable
        :columns="columns"
        :data="items"
        :loading="loading"
        :row-key="(row: ColumnItem) => row.id"
      />
    </NCard>

    <FormModal
      v-model:show="editVisible"
      :title="modalTitle"
      :loading="saving"
      @confirm="handleSubmit"
    >
      <NFormItem label="上级分类">
        <NSelect
          v-model:value="formModel.parentId"
          :options="parentOptions"
          :disabled="hasChildren"
          clearable
          placeholder="无上级（一级分类）"
        />
      </NFormItem>
      <NFormItem label="分类名称">
        <NInput
          v-model:value="formModel.name"
          placeholder="请输入专栏名称"
        />
      </NFormItem>
      <NFormItem label="专栏短链">
        <NInput
          v-model:value="formModel.shortUrl"
          placeholder="例如 life-notes"
        />
      </NFormItem>
    </FormModal>
  </ScrollContainer>
</template>
