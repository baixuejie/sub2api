import type { RouteRecordRaw } from 'vue-router'

export const pelicanRoutes: RouteRecordRaw[] = [
  {
    path: '/pelican',
    name: 'PelicanGallery',
    component: () => import('./views/PelicanGalleryView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: false,
      title: 'Generation gallery',
      titleKey: 'pelican.title',
      descriptionKey: 'pelican.description'
    }
  },
  {
    path: '/admin/pelican',
    name: 'PelicanAdmin',
    component: () => import('./views/PelicanAdminView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Gallery settings',
      titleKey: 'pelican.adminTitle',
      descriptionKey: 'pelican.adminDescription'
    }
  }
]
