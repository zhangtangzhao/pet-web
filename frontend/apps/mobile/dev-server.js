const http = require('http');
const fs = require('fs');
const path = require('path');
const { URL } = require('url');

const DIST = path.join(__dirname, 'dist');
const API_TARGET = 'http://127.0.0.1:8888';
const PORT = 10086;

const server = http.createServer((req, res) => {
  const url = new URL(req.url, `http://localhost:${PORT}`);

  // Proxy /api → backend
  if (url.pathname.startsWith('/api/')) {
    const proxyReq = http.request(API_TARGET + req.url, {
      method: req.method,
      headers: { ...req.headers, host: '127.0.0.1:8888' },
    }, (proxyRes) => {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.pipe(res);
    });
    proxyReq.on('error', () => { res.writeHead(502); res.end('proxy error'); });
    req.pipe(proxyReq);
    return;
  }

  // Static files with SPA fallback
  let filePath = path.join(DIST, url.pathname);
  if (!fs.existsSync(filePath) || fs.statSync(filePath).isDirectory()) {
    filePath = path.join(DIST, 'index.html');
  }
  if (fs.existsSync(filePath)) {
    const ext = path.extname(filePath);
    const mimes = { '.html': 'text/html', '.js': 'application/javascript', '.css': 'text/css', '.png': 'image/png', '.jpg': 'image/jpeg', '.svg': 'image/svg+xml', '.json': 'application/json' };
    res.writeHead(200, { 'Content-Type': mimes[ext] || 'application/octet-stream' });
    fs.createReadStream(filePath).pipe(res);
  } else {
    res.writeHead(404); res.end('Not found');
  }
});

server.listen(PORT, () => console.log(`Mobile H5 dev server → http://localhost:${PORT}`));
