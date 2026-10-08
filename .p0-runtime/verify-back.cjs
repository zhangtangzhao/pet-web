const http = require('http')
const fs = require('fs')
const path = require('path')
const { chromium } = require('C:/Users/ztz/AppData/Local/npm-cache/_npx/e41f203b7505f1fb/node_modules/playwright')

const ROOT = path.resolve(__dirname, '../frontend/apps/mobile/dist-p0')
const OUTPUT = path.resolve(__dirname, 'back-screens')
const PORT = 18099
fs.mkdirSync(OUTPUT, { recursive: true })

function product(id, title) {
  return {
    id, title, mainImage: '', price: '3280.00', originalPrice: '3680.00', breedName: '占位品种',
    petGender: 1, favoriteCount: 8, sales: 12, status: 1,
    images: [], detailImages: [], videoUrl: '', videoCover: '', detailHtml: '', hasSku: 0,
    isFavorite: false, aiEnabled: false,
    petProfile: { gender: 1, genderText: '公', birthDate: '', ageText: '3 个月', vaccineDesc: '', dewormDesc: '', bodyType: '', coatColor: '', personality: '', healthDesc: '' },
    breed: { id: 'b1', categoryId: 'c1', name: '占位品种', cover: '' },
    category: { id: 'c1', name: '犬类' },
    skus: [],
  }
}

const mime = { '.html': 'text/html; charset=utf-8', '.js': 'text/javascript; charset=utf-8', '.css': 'text/css; charset=utf-8' }

function json(res, data) {
  res.writeHead(200, { 'Content-Type': 'application/json; charset=utf-8' })
  res.end(JSON.stringify({ code: 0, msg: 'ok', data }))
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url, `http://127.0.0.1:${PORT}`)
  if (url.pathname.startsWith('/api/')) console.log(`API ${req.method} ${url.pathname}`)
  if (url.pathname.startsWith('/api/')) {
    if (url.pathname === '/api/products/p1') {
      const item = product('p1', '列表入口占位宠物')
      return json(res, item)
    }
    if (url.pathname === '/api/products/p2') {
      const item = product('p2', '搜索入口占位宠物')
      return json(res, item)
    }
    if (url.pathname === '/api/products/p1/related') {
      return json(res, { total: 1, list: [product('p2', '详情推荐占位宠物')] })
    }
    if (url.pathname === '/api/products') {
      return json(res, { total: 2, list: [product('p1', '列表入口占位宠物'), product('p2', '搜索入口占位宠物')] })
    }
    if (url.pathname === '/api/categories') return json(res, [{ id: 'c1', name: '犬类', icon: '', breedCount: 1 }])
    if (url.pathname === '/api/ai/recommend') {
      return json(res, { items: [product('p1', '推荐入口占位宠物')], source: 'rule' })
    }
    return json(res, { list: [] })
  }

  let file = path.join(ROOT, decodeURIComponent(url.pathname))
  if (!file.startsWith(ROOT)) { res.writeHead(403); return res.end() }
  if (!fs.existsSync(file) || fs.statSync(file).isDirectory()) file = path.join(ROOT, 'index.html')
  res.writeHead(200, { 'Content-Type': mime[path.extname(file)] || 'application/octet-stream' })
  fs.createReadStream(file).pipe(res)
})

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

async function checkBack(page, label) {
  const back = page.locator('.detail-back')
  await back.waitFor()
  await page.screenshot({ path: path.join(OUTPUT, `390-back-${label}.png`) })
  const box = await back.boundingBox()
  const text = await page.locator('.detail-back-text').textContent()
  assert(box && box.width >= 44 && box.height >= 44, `${label} 返回触控区不足：${box && `${box.width}x${box.height}`}`)
  assert(text && text.trim() === '返回', `${label} 返回语义缺失`)
  return { width: Math.round(box.width), height: Math.round(box.height), text: text.trim() }
}

async function run() {
  await new Promise(resolve => server.listen(PORT, '127.0.0.1', resolve))
  const browser = await chromium.launch({
    executablePath: 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
    headless: true,
    args: ['--no-sandbox', '--disable-dev-shm-usage', '--disable-gpu'],
  })
  const context = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const page = await context.newPage()
  const base = `http://127.0.0.1:${PORT}`
  const evidence = {}

  await page.goto(`${base}/pages/detail/index?id=p1`)
  await page.getByText('列表入口占位宠物').first().waitFor()
  evidence.direct = await checkBack(page, 'direct')
  await page.locator('.detail-back').click()
  await page.waitForURL(`${base}/pages/index/index`)
  evidence.directFallbackUrl = page.url()

  await page.goto(`${base}/pages/list/index`)
  await page.locator('.lst-card').first().click()
  await page.getByText('列表入口占位宠物').first().waitFor()
  evidence.list = await checkBack(page, 'list')
  await page.locator('.detail-back').click()
  await page.waitForURL(`${base}/pages/list/index`)
  evidence.listBackUrl = page.url()

  await page.goto(`${base}/pages/list/index?keyword=${encodeURIComponent('搜索')}`)
  await page.locator('.lst-card').first().click()
  await page.getByText('搜索入口占位宠物').first().waitFor()
  evidence.search = await checkBack(page, 'search')
  await page.locator('.detail-back').click()
  await page.waitForURL(/\/pages\/list\/index\?keyword=/)
  evidence.searchBackUrl = page.url()

  await page.goto(`${base}/pages/detail/index?id=p1`)
  await page.getByText('列表入口占位宠物').first().waitFor()
  await page.locator('.related-item').first().click({ force: true })
  await page.waitForResponse(response => response.url().includes('/api/products/p2'))
  await page.getByText('详情推荐占位宠物').first().waitFor({ state: 'attached' })
  evidence.recommend = await checkBack(page, 'recommend')

  await browser.close()
  server.close()
  fs.writeFileSync(path.join(OUTPUT, 'evidence.json'), JSON.stringify(evidence, null, 2))
  console.log(JSON.stringify(evidence, null, 2))
}

run().catch(error => {
  console.error(error)
  process.exitCode = 1
})
