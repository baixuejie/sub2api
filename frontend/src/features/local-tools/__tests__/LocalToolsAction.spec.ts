import { afterEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import type { ApiKey } from '@/types'
import LocalToolsAction from '../components/LocalToolsAction.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ locale: { value: 'zh-CN' }, t: (key: string) => key })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError: vi.fn(), showSuccess: vi.fn() })
}))

vi.mock('@/api/client', () => ({
  apiClient: { get: vi.fn(), post: vi.fn() }
}))

const openedHarness = vi.fn()
const apiKey = {
  id: 12,
  key: 'sk-local-tools-test',
  status: 'active',
  group: null
} as ApiKey

const mountAction = (props: Record<string, unknown> = {}) =>
  mount(LocalToolsAction, {
    attachTo: document.body,
    props: {
      apiKey,
      publicSettings: { site_name: 'Sub2API Test', api_base_url: 'https://api.example.com' },
      showCcSwitch: true,
      deepSeekHarnessEnabled: true,
      ...props
    },
    global: {
      stubs: {
        Icon: { template: '<span />' },
        DeepSeekHarnessAction: {
          props: ['status'],
          methods: { openDialog: openedHarness },
          template: '<div><slot name="trigger" :open="openDialog" :disabled="status !== \'active\'" /></div>'
        }
      }
    }
  })

function bodyButton(label: string): HTMLButtonElement {
  const button = [...document.body.querySelectorAll('button')].find((item) =>
    item.textContent?.includes(label)
  )
  if (!button) throw new Error(`button not found: ${label}`)
  return button
}

afterEach(() => {
  document.body.innerHTML = ''
  openedHarness.mockReset()
})

describe('LocalToolsAction', () => {
  it('uses one dropdown entry for CC Switch and DeepSeek Harness', async () => {
    const open = vi.spyOn(window, 'open').mockImplementation(() => null)
    const wrapper = mountAction()

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(document.body.textContent).toContain('CC Switch')
    expect(document.body.textContent).toContain('DeepSeek Harness')

    bodyButton('CC Switch').click()
    await flushPromises()
    expect(open).toHaveBeenCalledWith(expect.stringMatching(/^ccswitch:\/\/v1\/import\?/), '_self')
    wrapper.unmount()
    open.mockRestore()
  })

  it('opens the Harness action through the shared tool menu', async () => {
    const wrapper = mountAction()

    await wrapper.get('button').trigger('click')
    await flushPromises()
    bodyButton('DeepSeek Harness').click()
    await flushPromises()

    expect(openedHarness).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  it('keeps the Antigravity CC Switch client selection inside the extension', async () => {
    const open = vi.spyOn(window, 'open').mockImplementation(() => null)
    const wrapper = mountAction({
      apiKey: {
        ...apiKey,
        group: { platform: 'antigravity' }
      } as ApiKey
    })

    await wrapper.get('button').trigger('click')
    await flushPromises()
    bodyButton('CC Switch').click()
    await flushPromises()

    expect(document.body.textContent).toContain('keys.ccsClientSelect.title')
    bodyButton('keys.ccsClientSelect.claudeCode').click()
    await flushPromises()
    expect(open).toHaveBeenCalledWith(expect.stringContaining('app=claude'), '_self')

    wrapper.unmount()
    open.mockRestore()
  })

  it('keeps the trigger disabled when the only tool cannot use the key', () => {
    const wrapper = mountAction({
      showCcSwitch: false,
      apiKey: { ...apiKey, status: 'inactive' }
    })

    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })
})
