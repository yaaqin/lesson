// Service worker MathQuest (PWA). Sengaja minimal: gak nge-cache halaman/API
// (data soal, nyawa, progres harus selalu fresh dari server). Satu-satunya
// kerjaan: kalau navigasi gagal karena offline, tampilin halaman offline
// sederhana alih-alih error bawaan browser.
const OFFLINE_HTML = `<!doctype html>
<html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>MathQuest — Offline</title>
<style>
  body{margin:0;min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:12px;
    font-family:system-ui,-apple-system,sans-serif;background:#fafafa;color:#18181b;text-align:center;padding:24px}
  @media (prefers-color-scheme:dark){body{background:#000;color:#fafafa}}
  button{border:0;border-radius:999px;padding:12px 24px;background:#2563eb;color:#fff;font-weight:600;font-size:14px}
  p{color:#71717a;font-size:14px;margin:0;max-width:320px}
</style></head>
<body><div style="font-size:48px">📡</div><h1 style="font-size:20px;margin:0">Kamu lagi offline</h1>
<p>MathQuest butuh koneksi internet buat ngambil soal & nyimpen progres. Cek koneksi kamu, lalu coba lagi.</p>
<button onclick="location.reload()">Coba lagi</button></body></html>`;

self.addEventListener("install", () => self.skipWaiting());
self.addEventListener("activate", (event) => event.waitUntil(self.clients.claim()));

self.addEventListener("fetch", (event) => {
  if (event.request.mode !== "navigate") return;
  event.respondWith(
    fetch(event.request).catch(
      () => new Response(OFFLINE_HTML, { headers: { "Content-Type": "text/html; charset=utf-8" } }),
    ),
  );
});
