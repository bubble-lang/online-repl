const CACHE_NAME = 'bubble-repl-cache';
const urlsToCache = [
  '/',
  '/index.html',
  'https://cdn.jsdelivr.net/gh/bubble-lang/online-repl@main/main.wasm',
  'https://cdn.jsdelivr.net/gh/bubble-lang/online-repl@main/wasm_exec.js',
  // Add more files to cache as needed
];

self.addEventListener('install', function(event) {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(function(cache) {
        return cache.addAll(urlsToCache);
      })
  );
});

self.addEventListener('fetch', function(event) {
  event.respondWith(
    caches.match(event.request)
      .then(function(response) {
        if (response) {
          return response;
        }
        return fetch(event.request);
      })
  );
});