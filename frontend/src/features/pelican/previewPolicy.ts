import DOMPurify from 'dompurify'

export const PREVIEW_POLICY_VERSION = 1
const csp =
  "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; connect-src 'none'; img-src 'none'; font-src 'none'; media-src 'none'; object-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'"
const animatedAttributes = new Set(
  'transform d x y x1 x2 y1 y2 cx cy r rx ry width height dx dy rotate points opacity fill fill-opacity stroke stroke-width stroke-opacity stroke-dashoffset stroke-dasharray offset stop-color stop-opacity'.split(
    ' '
  )
)

// This function only prepares inert HTML. The iframe supplies a separate origin
// and never grants script permission; model HTML is never mounted in the app DOM.
export function buildPreviewHTML(raw: string): string | null {
  if (!raw || raw.length > 1048576) return null
  const clean = DOMPurify.sanitize(raw, {
    WHOLE_DOCUMENT: true,
    USE_PROFILES: { html: true, svg: true, svgFilters: true },
    ADD_TAGS: ['style', 'animate', 'animateTransform', 'animateMotion', 'mpath'],
    ADD_ATTR: [
      'attributeName',
      'attributeType',
      'from',
      'to',
      'by',
      'begin',
      'end',
      'dur',
      'repeatCount',
      'repeatDur',
      'values',
      'keyTimes',
      'keySplines',
      'keyPoints',
      'calcMode',
      'additive',
      'accumulate'
    ],
    FORBID_TAGS: [
      'script',
      'foreignObject',
      'iframe',
      'object',
      'embed',
      'form',
      'input',
      'button',
      'a',
      'link',
      'base',
      'meta',
      'audio',
      'video',
      'img',
      'image',
      'feImage'
    ],
    FORBID_ATTR: ['src', 'srcset', 'action', 'formaction', 'target']
  })
  const doc = new DOMParser().parseFromString(clean, 'text/html')
  if (!doc.querySelector('svg')) return null
  for (const element of doc.querySelectorAll('*')) {
    for (const attribute of [...element.attributes]) {
      if (
        attribute.localName.toLowerCase() === 'href' &&
        !/^#[\p{L}\p{N}_.-]+$/u.test(attribute.value)
      )
        element.removeAttributeNode(attribute)
    }
    const attributeName = element.getAttribute('attributeName')
    if (attributeName && !animatedAttributes.has(attributeName.toLowerCase())) element.remove()
  }
  const policy = doc.createElement('meta')
  policy.setAttribute('http-equiv', 'Content-Security-Policy')
  policy.setAttribute('content', csp)
  doc.head.prepend(policy)
  return '<!DOCTYPE html>' + doc.documentElement.outerHTML
}
