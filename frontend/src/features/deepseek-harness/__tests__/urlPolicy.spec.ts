import { describe, expect, it } from 'vitest'
import {
  safeDownloadURL,
  safeHarnessURL,
  safeLaunchURI,
  selectHelperDownload
} from '../utils/urlPolicy'
import type { HarnessHelperDownloads } from '../types/deepseekHarness'

const downloads: HarnessHelperDownloads = {
  windows_amd64: 'https://downloads.example.com/windows-amd64.zip',
  windows_arm64: 'https://downloads.example.com/windows-arm64.zip',
  darwin_amd64: 'https://downloads.example.com/darwin-amd64.tar.gz',
  darwin_arm64: 'https://downloads.example.com/darwin-arm64.tar.gz',
  linux_amd64: 'https://downloads.example.com/linux-amd64.tar.gz',
  linux_arm64: 'https://downloads.example.com/linux-arm64.tar.gz',
  releases_page: 'https://downloads.example.com/releases'
}

describe('DeepSeek Harness URL policy', () => {
  it('accepts only the exact bootstrap protocol shape', () => {
    const valid = 'sub2api-harness://bootstrap?server=https%3A%2F%2Fapi.example.com&ticket=abc_123&operation_id=session-1'
    expect(safeLaunchURI(valid)).toContain('sub2api-harness://bootstrap')
    expect(safeLaunchURI(`${valid}&extension_id=hermes`)).toContain('extension_id=hermes')

    for (const unsafe of [
      'javascript:alert(1)',
      'sub2api-harness://other?server=https%3A%2F%2Fapi.example.com&ticket=abc&operation_id=one',
      'sub2api-harness://bootstrap:63118?server=https%3A%2F%2Fapi.example.com&ticket=abc&operation_id=one',
      'sub2api-harness://bootstrap?server=http%3A%2F%2Fevil.example.com&ticket=abc&operation_id=one',
      'sub2api-harness://bootstrap?server=https%3A%2F%2Fapi.example.com&ticket=a%2Fb&operation_id=one',
      'sub2api-harness://bootstrap?server=https%3A%2F%2Fapi.example.com&ticket=abc&operation_id=one&extra=value',
      'sub2api-harness://bootstrap?server=https%3A%2F%2Fapi.example.com&server=https%3A%2F%2Fevil.example.com&ticket=abc&operation_id=one',
      'sub2api-harness://bootstrap?server=https%3A%2F%2Fapi.example.com&ticket=abc&operation_id=one&extension_id=Hermes%2Fshell',
      'sub2api-harness://bootstrap?server=https%3A%2F%2Fapi.example.com%2Fv1&ticket=abc&operation_id=one'
    ]) {
      expect(safeLaunchURI(unsafe)).toBe('')
    }
  })

  it('allows HTTPS downloads and loopback Harness URLs only', () => {
    expect(safeDownloadURL('https://downloads.example.com/helper.zip')).toContain('https://')
    expect(safeDownloadURL('javascript:alert(1)')).toBe('')
    expect(safeDownloadURL('http://downloads.example.com/helper.zip')).toBe('')

    expect(safeHarnessURL('http://127.0.0.1:3080')).toBe('http://127.0.0.1:3080/')
    expect(safeHarnessURL('http://localhost:3080/path')).toBe('http://localhost:3080/path')
    expect(safeHarnessURL('https://example.com:3080')).toBe('')
    expect(safeHarnessURL('file:///tmp/config')).toBe('')
  })

  it('does not guess an artifact when browser architecture is ambiguous', () => {
    expect(selectHelperDownload(downloads, 'Win32', 'Windows NT 10.0; Win64; x64')).toContain('windows-amd64')
    expect(selectHelperDownload(downloads, 'Linux aarch64', 'Linux')).toContain('linux-arm64')
    expect(selectHelperDownload(downloads, 'MacIntel', 'Macintosh; Intel Mac OS X')).toBe('')
    expect(selectHelperDownload(downloads, 'Unknown', 'Unknown')).toBe('')
  })
})
