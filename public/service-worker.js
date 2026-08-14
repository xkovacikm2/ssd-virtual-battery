const CACHE_NAME = 'ssd-virtual-battery-cache-v1';
const OFFLINE_URL = '/offline.html';
const PRECACHE_URLS = [
  '/',
  OFFLINE_URL,
  '/icon.png'
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME).then(cache => cache.addAll(PRECACHE_URLS))
  );
  self.skipWaiting();
});

self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys().then(keys =>
      Promise.all(keys.map(k => { if (k !== CACHE_NAME) return caches.delete(k); }))
    )
  );
  self.clients.claim();
});

self.addEventListener('fetch', event => {
  // navigation requests -> network-first with offline fallback
  if (event.request.mode === 'navigate') {
    event.respondWith(
      fetch(event.request).catch(() => caches.match(OFFLINE_URL))
    );
    return;
  }

  // other requests: try cache first, then network
  event.respondWith(
    caches.match(event.request).then(cached => {
      if (cached) return cached;
      return fetch(event.request).then(response => {
        return response;
      }).catch(() => {
        if (event.request.destination === 'document') {
          return caches.match(OFFLINE_URL);
        }
      });
    })
  );
});
