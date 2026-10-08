const http = require('http')
const fs = require('fs')
const path = require('path')
const { URL } = require('url')

const ROOT = path.resolve(__dirname, '../frontend/apps/mobile/dist-p0')
const API_TARGET = 'http://127.0.0.1:8888'
const PORT = 10087

const mime = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.webp': 'image/webp',
  '.gif': 'image/gif',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
}

function serveFile(req, res, pathname) {
  let filePath = path.join(ROOT, decodeURIComponent(pathname))
  if (!filePath.startsWith(ROOT)) {
    res.writeHead(403)
    return res.end()
  }
  if (!fs.existsSync(filePath) || fs.statSync(filePath).isDirectory()) {
    filePath = path.join(ROOT, 'index.html')
  }
  if (!fs.existsSync(filePath)) {
    res.writeHead(404)
    return res.end()
  }

  const ext = path.extname(filePath).toLowerCase()
  const cacheControl = ext === '.html' ? 'no-cache' : 'public, max-age=3600'
  res.writeHead(200, {
    'Content-Type': mime[ext] || 'application/octet-stream',
    'Cache-Control': cacheControl,
  })
  if (req.method === 'HEAD') return res.end()
  fs.createReadStream(filePath).pipe(res)
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url, `http://127.0.0.1:${PORT}`)

  if (url.pathname.startsWith('/api/')) {
    const headers = { ...req.headers, host: '127.0.0.1:8888' }
    delete headers['if-none-match']
    delete headers['if-modified-since']
    const proxyReq = http.request(`${API_TARGET}${req.url}`, {
      method: req.method,
      headers,
    }, proxyRes => {
      res.writeHead(proxyRes.statusCode || 502, proxyRes.headers)
      proxyRes.pipe(res)
    })
    proxyReq.on('error', () => {
      res.writeHead(502, { 'Content-Type': 'application/json; charset=utf-8' })
      res.end(JSON.stringify({ code: 1, msg: 'backend unavailable', data: null }))
    })
    req.pipe(proxyReq)
    return
  }

  serveFile(req, res, url.pathname)
})

server.listen(PORT, '0.0.0.0', () => {
  console.log(`${new Date().toISOString()} dist-p0 server ready: http://127.0.0.1:${PORT} root=${ROOT} pid=${process.pid}`)
})
