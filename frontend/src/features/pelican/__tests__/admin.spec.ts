import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import Admin from '../views/PelicanAdminView.vue'

const api = vi.hoisted(() => ({
  config: vi.fn(),
  list: vi.fn(),
  save: vi.fn(),
  clearKey: vi.fn(),
  detail: vi.fn(),
  source: vi.fn()
}))
vi.mock('../api', () => ({ pelicanAPI: api }))
vi.mock('vue-i18n', async () => {
  const { ref } = await import('vue')
  return { useI18n: () => ({ t: (key: string) => key, locale: ref('zh') }) }
})
vi.mock('@/api/admin/groups', () => ({
  getAll: async () => [
    { id: 1, name: 'A' },
    { id: 2, name: 'B' }
  ]
}))
vi.mock('@/components/layout/AppLayout.vue', () => ({
  default: { template: '<main><slot /></main>' }
}))
const config = {
  revision: 1,
  enabled: false,
  selected_group_ids: [],
  topic_mode: 'rotate',
  fixed_topic_id: 'pelican-ski',
  max_output_tokens: 16384,
  timeout_seconds: 300,
  retention_days: 30,
  next_run_at: null,
  key_configured: false,
  key_masked: '',
  encryption_ready: true,
  key_unavailable: false,
  model: 'gpt-6-astra',
  reasoning_effort: 'high'
}
let wrapper: VueWrapper | undefined
beforeEach(() => {
  vi.resetAllMocks()
  api.config.mockResolvedValue(config)
  api.list.mockResolvedValue({ items: [], next_cursor: '' })
  api.save.mockResolvedValue({ ...config, revision: 2, key_configured: true })
})
afterEach(() => wrapper?.unmount())

it('saves one key and several labels without any generation or editable model fields', async () => {
  wrapper = mount(Admin, { global: { stubs: { RouterLink: true, BaseDialog: true } } })
  await flushPromises()
  await wrapper.find('#pelican-key').setValue('designated-test-key')
  await wrapper.find('input[value="1"]').setValue(true)
  await wrapper.find('input[value="2"]').setValue(true)
  await wrapper.find('form').trigger('submit')
  await flushPromises()
  expect(api.save).toHaveBeenCalledTimes(1)
  expect(api.save.mock.calls[0][0]).toMatchObject({
    api_key: 'designated-test-key',
    selected_group_ids: [1, 2]
  })
  expect(api.save.mock.calls[0][0]).not.toHaveProperty('model')
  expect(api.save.mock.calls[0][0]).not.toHaveProperty('reasoning_effort')
  expect((wrapper.find('#pelican-key').element as HTMLInputElement).value).toBe('')
  expect(wrapper.text()).not.toContain('designated-test-key')
})
