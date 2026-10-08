const http = require('http')
const fs = require('fs')
const path = require('path')
const { chromium } = require('C:/Users/ztz/AppData/Local/npm-cache/_npx/e41f203b7505f1fb/node_modules/playwright')

const ROOT = path.resolve(__dirname, '../frontend/apps/mobile/dist-p0')
const OUTPUT = path.resolve(__dirname, 'p1-screens')
const PORT = 18130
fs.mkdirSync(OUTPUT, { recursive: true })

const product = {
  id: 'sku-p1',
  title: '多断点视觉验收占位宠物',
  mainImage: '', price: '3280.00', originalPrice: '3680.00', breedName: '占位品种',
  petGender: 1, favoriteCount: 12, sales: 34, status: 1,
  images: [], detailImages: [], videoUrl: '', videoCover: '', detailHtml: '',
  hasSku: 1, isFavorite: false, aiEnabled: false,
  petProfile: { gender: 1, genderText: '公', birthDate: '', ageText: '3 个月', vaccineDesc: '占位记录', dewormDesc: '占位记录', bodyType: '小型', coatColor: '占位色', personality: '温顺', healthDesc: '' },
  breed: { id: 'b1', categoryId: 'c1', name: '占位品种', cover: '' },
  category: { id: 'c1', name: '犬类' },
  skus: [
    { id: 'sku-1', specs: '标准方案', price: '3280.00' },
    { id: 'sku-2', specs: '完整方案', price: '3680.00' },
  ],
}

const mime = { '.html': 'text/html; charset=utf-8', '.js': 'text/javascript; charset=utf-8', '.css': 'text/css; charset=utf-8', '.png': 'image/png' }

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function json(res, data) {
  res.writeHead(200, { 'Content-Type': 'application/json; charset=utf-8' })
  res.end(JSON.stringify({ code: 0, msg: 'ok', data }))
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url, `http://127.0.0.1:${PORT}`)
  if (url.pathname.startsWith('/api/')) {
    if (url.pathname === '/api/products/sku-p1') return json(res, product)
    if (url.pathname === '/api/products/sku-p1/review-summary') return json(res, { avgRating: '4.8', total: 2, latest: [] })
    if (url.pathname === '/api/products/sku-p1/reviews') return json(res, { list: [], hasMore: false })
    if (url.pathname === '/api/products/sku-p1/related') return json(res, { total: 0, list: [] })
    if (url.pathname === '/api/categories') return json(res, [{ id: 'c1', name: '犬类', icon: '', breedCount: 1 }])
    if (url.pathname === '/api/products') {
      return json(res, { total: 4, list: Array.from({ length: 4 }, (_, index) => ({
        id: `p${index + 1}`, title: `占位宠物 ${index + 1}`, mainImage: '', price: `${3000 + index * 100}.00`,
        originalPrice: '3680.00', breedName: '占位品种', petGender: 1, favoriteCount: 0, sales: index, status: 1,
      })) })
    }
    return json(res, { list: [] })
  }

  let file = path.join(ROOT, decodeURIComponent(url.pathname))
  if (!file.startsWith(ROOT)) { res.writeHead(403); return res.end() }
  if (!fs.existsSync(file) || fs.statSync(file).isDirectory()) file = path.join(ROOT, 'index.html')
  res.writeHead(200, { 'Content-Type': mime[path.extname(file)] || 'application/octet-stream' })
  fs.createReadStream(file).pipe(res)
})

async function newPage(context) {
  const page = await context.newPage()
  page.setDefaultTimeout(12000)
  page.on('pageerror', error => console.error(`PAGEERROR ${error.message}`))
  return page
}

async function measure(page, selectors) {
  const output = {}
  for (const [name, selector] of Object.entries(selectors)) {
    output[name] = await page.locator(selector).first().evaluate(el => Math.round(el.getBoundingClientRect().width))
  }
  return output
}

async function captureBase(browser, width, height, suffix) {
  const context = await browser.newContext({ viewport: { width, height }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const page = await newPage(context)
  const base = `http://127.0.0.1:${PORT}`

  await page.goto(`${base}/pages/index/index`)
  await page.getByText('占位宠物 1').first().waitFor()
  await page.screenshot({ path: path.join(OUTPUT, `${suffix}-home.png`), fullPage: true })
  const homeWidth = await measure(page, { page: '.home', card: '.pet-card' })

  await page.goto(`${base}/pages/list/index`)
  await page.getByText('占位宠物 1').first().waitFor()
  await page.screenshot({ path: path.join(OUTPUT, `${suffix}-list.png`), fullPage: true })
  const listWidth = await measure(page, { page: '.lst', card: '.pet-card' })

  await page.goto(`${base}/pages/detail/index?id=sku-p1`)
  await page.getByText(product.title).first().waitFor()
  await page.screenshot({ path: path.join(OUTPUT, `${suffix}-detail.png`) })
  const detailWidth = await measure(page, { page: '.detail', card: '.head' })

  if (suffix === '390') {
    const actionbar = page.locator('.detail-actionbar')
    await actionbar.screenshot({ path: path.join(OUTPUT, '390-detail-actionbar.png') })
    const labels = await page.locator('.detail-actionbar .fav .fav-text').allTextContents()
    assert(labels.join(',') === '收藏,客服', '详情底栏语义文案重复')
    const favBox = await page.locator('.detail-actionbar .fav').first().boundingBox()
    console.log('ACTIONBAR FAV BOX', favBox)
    assert((favBox?.height || 0) >= 44, '详情底栏收藏触控高度不足')
  }

  await context.close()
  return { homeWidth, listWidth, detailWidth }
}

async function captureStates(browser) {
  const context = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const page = await newPage(context)
  const base = `http://127.0.0.1:${PORT}`

  await page.route('**/api/products?*', async route => {
    await new Promise(resolve => setTimeout(resolve, 700))
    try {
      await route.fulfill({ contentType: 'application/json', body: JSON.stringify({ code: 0, msg: 'ok', data: { total: 4, list: [] } }) })
    } catch {
      // A later handler may have taken over before the artificial delay ends.
    }
  })
  await page.goto(`${base}/pages/index/index`)
  await page.locator('.fb-grid').waitFor()
  await page.screenshot({ path: path.join(OUTPUT, 'state-home-skeleton.png') })
  await page.unroute('**/api/products?*')
  await page.route('**/api/products?*', route => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ code: 1, msg: 'Exception', data: null }) }))
  await page.reload()
  await page.getByText('商品加载失败').waitFor()
  await page.screenshot({ path: path.join(OUTPUT, 'state-home-error.png') })

  await page.unroute('**/api/products?*')

  await page.route('**/api/products?*', async route => {
    await new Promise(resolve => setTimeout(resolve, 500))
    try {
      await route.continue()
    } catch {
      // Ignore route handoff races after the skeleton is captured.
    }
  })

  await page.goto(`${base}/pages/list/index`)
  await page.locator('.fb-grid').waitFor()
  await page.screenshot({ path: path.join(OUTPUT, 'state-list-skeleton.png') })
  await page.unroute('**/api/products?*')

  await page.route('**/api/products/sku-p1', async route => {
    await new Promise(resolve => setTimeout(resolve, 500))
    try {
      await route.continue()
    } catch {
      // Ignore route handoff races after the skeleton is captured.
    }
  })
  await page.goto(`${base}/pages/detail/index?id=sku-p1`)
  await page.locator('.fb-detail').waitFor()
  await page.screenshot({ path: path.join(OUTPUT, 'state-detail-skeleton.png') })
  await page.unroute('**/api/products/sku-p1')
  await page.unroute('**/api/products/sku-p1')
  await page.route('**/api/products/sku-p1', route => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ code: 1, msg: 'Exception', data: null }) }))
  await page.goto(`${base}/pages/detail/index?id=sku-p1`)
  await page.getByText('商品详情加载失败').waitFor()
  await page.screenshot({ path: path.join(OUTPUT, 'state-detail-error.png') })
  await context.close()

  const listContext = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const listPage = await newPage(listContext)
  await listPage.route('**/api/products?*', route => route.fulfill({ contentType: 'application/json', body: JSON.stringify({ code: 1, msg: 'Exception', data: null }) }))
  await listPage.goto(`${base}/pages/list/index`)
  await listPage.getByText('商品加载失败').waitFor()
  await listPage.screenshot({ path: path.join(OUTPUT, 'state-list-error.png') })
  await listContext.close()
}

async function captureSku(browser) {
  const context = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const page = await newPage(context)
  await page.goto(`http://127.0.0.1:${PORT}/pages/detail/index?id=sku-p1`)
  await page.getByText(product.title).first().waitFor()
  await page.locator('.detail-actionbar .btn-buy').click()
  await page.locator('.sku-sheet-body').waitFor()
  const sheet = page.locator('.sku-sheet-body')
  await sheet.getByText('标准方案').first().click()
  await page.screenshot({ path: path.join(OUTPUT, 'state-sku-sheet.png') })
  const bodyBox = await page.locator('.sku-sheet-body').boundingBox()
  const confirmBox = await page.locator('.sku-sheet-confirm').boundingBox()
  const chipBox = await page.locator('.sku-chip').last().boundingBox()
  const dimensions = {
    width: Math.round(bodyBox?.width || 0),
    viewport: page.viewportSize().width,
    confirmHeight: Math.round(confirmBox?.height || 0),
    chipHeight: Math.round(chipBox?.height || 0),
  }
  console.log('SKU DIMENSIONS', dimensions)
  assert(dimensions.width >= dimensions.viewport - 1, 'SKU 弹层在 390px 视口未占满宽度')
  assert(dimensions.confirmHeight >= 44, 'SKU 确认按钮触控高度不足')
  assert(dimensions.chipHeight >= 44, 'SKU 选项触控高度不足')
  await page.getByText('关闭').click()
  await context.close()
  return dimensions
}

async function run() {
  await new Promise(resolve => server.listen(PORT, '127.0.0.1', resolve))
  const browser = await chromium.launch({
    executablePath: 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
    headless: true,
    args: ['--no-sandbox', '--disable-dev-shm-usage', '--disable-gpu'],
  })
  const evidence = {
    mobile390: await captureBase(browser, 390, 844, '390'),
    tablet768: await captureBase(browser, 768, 1024, '768'),
    desktop1440: await captureBase(browser, 1440, 900, '1440'),
  }
  evidence.states = await captureStates(browser)
  evidence.sku = await captureSku(browser)
  await browser.close()
  server.close()
  fs.writeFileSync(path.join(OUTPUT, 'evidence.json'), JSON.stringify(evidence, null, 2))
  console.log(JSON.stringify(evidence, null, 2))
}

run().catch(error => {
  console.error(error)
  process.exitCode = 1
})
