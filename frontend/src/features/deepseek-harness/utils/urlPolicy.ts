import type { HarnessHelperDownloads } from '../types/deepseekHarness'

const opaqueTokenPattern = /^[A-Za-z0-9_-]+$/

export function safeLaunchURI(raw: string | undefined): string {
  if (!raw) return ''
  try {
    const parsed = new URL(raw)
    const server = parsed.searchParams.get('server') || ''
    const ticket = parsed.searchParams.get('ticket') || ''
    const operationId = parsed.searchParams.get('operation_id') || ''
    const queryKeys = Array.from(parsed.searchParams.keys())
    if (
      parsed.protocol !== 'sub2api-harness:' ||
      parsed.hostname !== 'bootstrap' ||
      (parsed.pathname !== '' && parsed.pathname !== '/') ||
      parsed.username ||
      parsed.password ||
      parsed.hash ||
      queryKeys.length !== 3 ||
      queryKeys.some((key) => !['server', 'ticket', 'operation_id'].includes(key)) ||
      !opaqueTokenPattern.test(ticket) ||
      !opaqueTokenPattern.test(operationId) ||
      !safeServerURL(server)
    ) {
      return ''
    }
    return parsed.toString()
  } catch {
    return ''
  }
}

export function safeDownloadURL(raw: string | undefined): string {
  if (!raw) return ''
  try {
    const parsed = new URL(raw)
    if (parsed.protocol !== 'https:' || parsed.username || parsed.password) return ''
    return parsed.toString()
  } catch {
    return ''
  }
}

export function safeHarnessURL(raw: string | undefined): string {
  if (!raw) return ''
  try {
    const parsed = new URL(raw)
    if (
      parsed.protocol !== 'http:' ||
      parsed.username ||
      parsed.password ||
      !parsed.port ||
      !isLoopbackHost(parsed.hostname)
    ) {
      return ''
    }
    return parsed.toString()
  } catch {
    return ''
  }
}

export function selectHelperDownload(
  downloads: HarnessHelperDownloads,
  platform: string,
  userAgent: string,
  architecture = ''
): string {
  const system = `${platform} ${userAgent}`.toLowerCase()
  const arch = architecture.toLowerCase()
  const isArm = /arm64|aarch64/.test(`${arch} ${system}`)
  const isX64 = /x86_64|amd64|win64|x64/.test(`${arch} ${system}`)

  if (system.includes('win')) {
    if (isArm) return safeDownloadURL(downloads.windows_arm64)
    if (isX64) return safeDownloadURL(downloads.windows_amd64)
    return ''
  }
  if (system.includes('mac') || system.includes('darwin')) {
    if (arch.includes('arm') || system.includes('arm64')) return safeDownloadURL(downloads.darwin_arm64)
    if (arch.includes('x86') || arch.includes('amd64')) return safeDownloadURL(downloads.darwin_amd64)
    return ''
  }
  if (system.includes('linux')) {
    if (isArm) return safeDownloadURL(downloads.linux_arm64)
    if (isX64) return safeDownloadURL(downloads.linux_amd64)
  }
  return ''
}

function safeServerURL(raw: string): boolean {
  try {
    const parsed = new URL(raw)
    if (
      parsed.username ||
      parsed.password ||
      parsed.search ||
      parsed.hash ||
      (parsed.pathname !== '' && parsed.pathname !== '/')
    ) {
      return false
    }
    if (parsed.protocol === 'https:') return true
    return parsed.protocol === 'http:' && isLoopbackHost(parsed.hostname)
  } catch {
    return false
  }
}

function isLoopbackHost(hostname: string): boolean {
  const host = hostname.replace(/^\[|\]$/g, '').toLowerCase()
  if (host === 'localhost' || host === '::1') return true
  const octets = host.split('.').map(Number)
  return (
    octets.length === 4 &&
    octets.every((octet) => Number.isInteger(octet) && octet >= 0 && octet <= 255) &&
    octets[0] === 127
  )
}
