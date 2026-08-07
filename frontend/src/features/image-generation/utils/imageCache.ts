const DB_NAME = 'sub2api-image-generation'
const STORE_NAME = 'images'
const DB_VERSION = 1
const MAX_ITEMS = 12
const MAX_BYTES = 100 * 1024 * 1024

export interface CachedImage {
  owner: string
  blob: Blob
  mimeType: string
  revisedPrompt?: string | null
  downloadName: string
  createdAt: number
}

interface StoredImage extends CachedImage {
  id: string
}

function openDatabase(): Promise<IDBDatabase | null> {
  if (typeof window === 'undefined' || !window.indexedDB) return Promise.resolve(null)
  return new Promise((resolve, reject) => {
    const request = window.indexedDB.open(DB_NAME, DB_VERSION)
    request.onupgradeneeded = () => {
      request.result.createObjectStore(STORE_NAME, { keyPath: 'id' })
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('image cache unavailable'))
  })
}

function transactionComplete(transaction: IDBTransaction): Promise<void> {
  return new Promise((resolve, reject) => {
    transaction.oncomplete = () => resolve()
    transaction.onerror = () => reject(transaction.error ?? new Error('image cache write failed'))
    transaction.onabort = () => reject(transaction.error ?? new Error('image cache write aborted'))
  })
}

export async function replaceCachedImages(owner: string, images: CachedImage[]): Promise<void> {
  if (!owner) return
  const database = await openDatabase()
  if (!database) return
  const selected: StoredImage[] = []
  let totalBytes = 0
  for (const image of images.slice(-MAX_ITEMS).reverse()) {
    const size = image.blob.size
    if (!size || totalBytes + size > MAX_BYTES) continue
    selected.push({ ...image, owner, id: `${owner}:${image.createdAt}:${selected.length}` })
    totalBytes += size
  }
  selected.reverse()
  const transaction = database.transaction(STORE_NAME, 'readwrite')
  const store = transaction.objectStore(STORE_NAME)
  const existing = store.getAll()
  existing.onsuccess = () => {
    for (const image of existing.result as StoredImage[]) {
      if (image.owner === owner) store.delete(image.id)
    }
    selected.forEach((image) => store.put(image))
  }
  await transactionComplete(transaction)
  database.close()
}

export async function loadCachedImages(owner: string): Promise<CachedImage[]> {
  if (!owner) return []
  const database = await openDatabase()
  if (!database) return []
  const transaction = database.transaction(STORE_NAME, 'readonly')
  const request = transaction.objectStore(STORE_NAME).getAll()
  const records = await new Promise<StoredImage[]>((resolve, reject) => {
    request.onsuccess = () => resolve((request.result as StoredImage[]).filter((image) => image.owner === owner))
    request.onerror = () => reject(request.error ?? new Error('image cache read failed'))
  })
  database.close()
  return records.sort((left, right) => left.createdAt - right.createdAt)
}

export async function clearCachedImages(owner: string): Promise<void> {
  if (!owner) return
  const database = await openDatabase()
  if (!database) return
  const transaction = database.transaction(STORE_NAME, 'readwrite')
  const store = transaction.objectStore(STORE_NAME)
  const request = store.getAll()
  request.onsuccess = () => {
    for (const image of request.result as StoredImage[]) {
      if (image.owner === owner) store.delete(image.id)
    }
  }
  await transactionComplete(transaction)
  database.close()
}
