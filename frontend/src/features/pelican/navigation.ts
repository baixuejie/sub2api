import { h } from 'vue'

const GalleryIcon = {
  render: () =>
    h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': 1.5 }, [
      h('rect', { x: 3, y: 3, width: 18, height: 18, rx: 3 }),
      h('circle', { cx: 8, cy: 8, r: 1.5 }),
      h('path', { d: 'm3 17 5-5 4 4 4-6 5 7', 'stroke-linejoin': 'round' })
    ])
}
export function pelicanNavigation(t: (key: string) => string, admin = false) {
  return {
    path: admin ? '/admin/pelican' : '/pelican',
    label: t(admin ? 'pelican.adminNav' : 'pelican.nav'),
    icon: GalleryIcon
  }
}
