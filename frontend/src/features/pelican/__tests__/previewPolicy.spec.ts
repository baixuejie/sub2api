import { describe, expect, it } from 'vitest'
import { buildPreviewHTML } from '../previewPolicy'

describe('isolated SVG preview', () => {
  it('preserves gradients and animation while replacing model policy', () => {
    const result = buildPreviewHTML(
      `<html><head><meta http-equiv="Content-Security-Policy" content="default-src *"><style>@keyframes sway { to { transform: translateY(4px) } } .bird { animation: sway 2s infinite }</style></head><body><svg viewBox="0 0 960 720"><defs><linearGradient id="sky"><stop offset="0" stop-color="blue" /></linearGradient></defs><rect fill="url(#sky)" width="960" height="720"/><g class="bird"><circle r="20"/><animateTransform attributeName="transform" type="translate" values="0 0;0 5" dur="2s" repeatCount="indefinite"/></g></svg></body></html>`
    )
    expect(result).toContain('@keyframes sway')
    expect(result).toContain('animateTransform')
    expect(result).toContain('url(#sky)')
    const doc = new DOMParser().parseFromString(result!, 'text/html')
    expect(doc.querySelectorAll('meta[http-equiv="Content-Security-Policy"]')).toHaveLength(1)
    expect(doc.querySelector('meta')?.getAttribute('content')).toContain("script-src 'none'")
    expect(doc.querySelector('meta')?.getAttribute('content')).toContain("connect-src 'none'")
  })

  it('removes active elements, external references, and URL-changing animations', () => {
    const result = buildPreviewHTML(
      `<svg onload="alert(1)"><script>alert(1)</script><foreignObject><iframe src="https://example.com"/></foreignObject><use href="https://example.com/a.svg#x"/><animate attributeName="href" values="javascript:alert(1)"/></svg><img src="https://example.com/track"><form action="https://example.com"><input/></form>`
    )
    const doc = new DOMParser().parseFromString(result!, 'text/html')
    expect(doc.querySelector('script, foreignObject, iframe, img, form, input, animate')).toBeNull()
    expect(doc.querySelector('[onload]')).toBeNull()
    expect(doc.querySelector('[href]')).toBeNull()
  })

  it('does not offer a preview for empty or oversized output', () => {
    expect(buildPreviewHTML('')).toBeNull()
    expect(buildPreviewHTML('<html><body>no SVG</body></html>')).toBeNull()
    expect(buildPreviewHTML('x'.repeat(1048577))).toBeNull()
  })
})
