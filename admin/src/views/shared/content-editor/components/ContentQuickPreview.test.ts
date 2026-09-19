import { flushPromises, shallowMount } from '@vue/test-utils'
import { NTooltip } from 'naive-ui'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import MarkdownPreview from '@/components/markdown-editor/MarkdownPreview.vue'
import { getMoment } from '@/services/moments'

import ContentQuickPreview from './ContentQuickPreview.vue'

vi.mock('@/services/moments', () => ({ getMoment: vi.fn() }))

describe('ContentQuickPreview', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads moment content only when the tooltip is opened', async () => {
    vi.mocked(getMoment).mockResolvedValue({ content: '# 手记预览' } as never)
    const wrapper = shallowMount(ContentQuickPreview, {
      props: { contentType: 'moment', contentId: 12 },
      global: { renderStubDefaultSlot: true },
    })

    expect(getMoment).not.toHaveBeenCalled()
    wrapper.findComponent(NTooltip).vm.$emit('update:show', true)
    await flushPromises()

    expect(getMoment).toHaveBeenCalledWith(12)
    expect(wrapper.findComponent(MarkdownPreview).props('source')).toBe('# 手记预览')
  })
})
