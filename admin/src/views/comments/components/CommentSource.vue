<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query'
import { ChatbubbleEllipsesOutline } from '@vicons/ionicons5'
import { NSpin, NTag, NText, NIcon } from 'naive-ui'
import { computed } from 'vue'

import { getMoment } from '@/services/moments'

const props = defineProps<{
  type?: string
  id?: number
  initialTitle?: string
}>()

const isMoment = computed(() => props.type === 'moments' || props.type === 'moment')

const { data: moment, isLoading: isLoadingMoment } = useQuery({
  queryKey: ['moment', props.id],
  queryFn: () => getMoment(props.id!),
  enabled: computed(() => !!props.id && isMoment.value),
  staleTime: 1000 * 60 * 5,
})

const displayTitle = computed(() => {
  if (isMoment.value && moment.value) return moment.value.title || '手记'
  return props.initialTitle || '未知来源'
})

const typeTag = computed(() => {
  if (isMoment.value)
    return { type: 'warning' as const, icon: ChatbubbleEllipsesOutline, label: '手记' }
  return { type: 'default' as const, icon: undefined, label: props.type || '其他' }
})
</script>

<template>
  <div class="flex items-center gap-2">
    <n-tag
      :type="typeTag.type"
      size="small"
      :bordered="false"
      class="flex items-center"
    >
      {{ typeTag.label }}
      <template
        #icon
        v-if="typeTag.icon"
      >
        <n-icon :component="typeTag.icon" />
      </template>
    </n-tag>

    <n-spin
      v-if="isLoadingMoment"
      size="small"
    />

    <n-text
      v-else
      class="max-w-[200px] truncate text-sm"
      :title="displayTitle"
    >
      {{ displayTitle }}
    </n-text>
  </div>
</template>
