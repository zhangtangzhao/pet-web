const http = require('http')
const fs = require('fs')
const path = require('path')
const { chromium } = require('C:/Users/ztz/AppData/Local/npm-cache/_npx/e41f203b7505f1fb/node_modules/playwright')

const ROOT = path.resolve(__dirname, '../frontend/apps/mobile/dist-p0')
const OUTPUT = path.resolve(__dirname, 'screens')
const PORT = 18131
const apiCalls = []
fs.mkdirSync(OUTPUT, { recursive: true })

const product = {
  id: 'sku-p1',
  title: '运行态验收占位宠物',
  mainImage: '',
  price: '3280.00',
  originalPrice: '3680.00',
  breedName: '占位品种',
  petGender: 1,
  favoriteCount: 0,
  sales: 0,
  status: 1,
  images: [],
  detailImages: [],
  videoUrl: '',
  videoCover: '',
  detailHtml: '',
  hasSku: 1,
  isFavorite: false,
  aiEnabled: false,
  petProfile: {
    gender: 1,
    genderText: '公',
    birthDate: '',
    ageText: '3 个月',
    vaccineDesc: '占位记录',
    dewormDesc: '占位记录',
    bodyType: '小型',
    coatColor: '占位色',
    personality: '温顺',
    healthDesc: '',
  },
  breed: { id: 'breed-1', categoryId: 'cat-1', name: '占位品种', cover: '' },
  category: { id: 'cat-1', name: '犬类' },
  skus: [
    { id: 'sku-1', specs: '标准方案', price: '3280.00' },
    { id: 'sku-2', specs: '完整方案', price: '3680.00' },
  ],
}

function reviews(cursor) {
  if (!cursor) {
    return {
      list: Array.from({ length: 10 }, (_, index) => review(`p1-r${index + 1}`)),
      hasMore: true,
    }
  }
  const start = Number(cursor.match(/(\d+)$/)?.[1] || '0')
  return {
    list: Array.from({ length: 10 }, (_, index) => review(`p1-r${start + index + 1}`)),
    hasMore: false,
  }
}

function review(id) {
  return {
    id,
    orderNo: `order-${id}`,
    memberId: 'member-1',
    nickname: `用户 ${id}`,
    avatar: '',
    productId: 'sku-p1',
    productTitle: product.title,
    rating: 5,
    content: `评价内容 ${id}`,
    images: [],
    status: 1,
    createdAt: '2026-09-21T10:00:00Z',
  }
}

function send(res, data) {
  const body = JSON.stringify({ code: 0, msg: 'ok', data })
  res.writeHead(200, { 'Content-Type': 'application/json; charset=utf-8' })
  res.end(body)
}

const mime = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.png': 'image/png',
  '.svg': 'image/svg+xml',
}

const server = http.createServer((req, res) => {
  const url = new URL(req.url, `http://127.0.0.1:${PORT}`)
  const chunks = []
  req.on('data', chunk => chunks.push(chunk))
  req.on('end', () => {
    const body = Buffer.concat(chunks).toString('utf8')
    apiCalls.push({ method: req.method, path: url.pathname + url.search, body: body || undefined })
    console.log(`API ${req.method} ${url.pathname}${url.search}${body ? ` ${body}` : ''}`)

    if (url.pathname === '/api/products/sku-p1') return send(res, product)
    if (url.pathname === '/api/products/sku-p1/review-summary') {
      return send(res, { avgRating: '4.8', total: 21, latest: [review('p1-r0')] })
    }
    if (url.pathname === '/api/products/sku-p1/reviews') {
      return send(res, reviews(url.searchParams.get('cursor') || ''))
    }
    if (url.pathname === '/api/auth/sms/send') return send(res, { ok: true })
    if (url.pathname === '/api/auth/sms/login') {
      return send(res, {
        accessToken: 'verify-access',
        refreshToken: 'verify-refresh',
        expiresIn: 3600,
        member: { id: 'member-1', nickname: '验收用户', avatar: '', phone: '13800000000', isNew: false },
      })
    }
    if (url.pathname === '/api/products/sku-p1/related') return send(res, { total: 0, list: [] })
    if (url.pathname === '/api/invite') return send(res, { inviteCode: 'verify' })
    if (url.pathname === '/api/services') return send(res, [])
    if (url.pathname === '/api/ship/methods') {
      return send(res, [{ id: 'ship-1', name: '配送', kind: 2, description: '验收配送', fee: '0.00' }])
    }
    if (url.pathname === '/api/addresses') {
      return send(res, [{ id: 'addr-1', name: '验收人', phone: '13800000000', address: '占位地址', isDefault: 1 }])
    }
    if (url.pathname === '/api/flash-sales' || url.pathname === '/api/stores') return send(res, [])
    if (url.pathname === '/api/points') return send(res, { points: 0 })
    if (url.pathname === '/api/trade-config') return send(res, { depositPercent: 0, depositHoldDays: 0 })
    if (url.pathname === '/api/coupons/usable') return send(res, [])
    if (url.pathname === '/api/cart' && req.method === 'POST') return send(res, { id: 'cart-1' })
    if (url.pathname === '/api/orders' && req.method === 'POST') {
      return send(res, {
        orderNo: 'order-runtime-1',
        paymentNo: 'pay-1',
        payAmount: '3280.00',
        expireAt: '2026-09-21T12:00:00Z',
        payParams: null,
        h5PayUrl: '',
      })
    }
    if (url.pathname === '/api/orders/order-runtime-1') {
      return send(res, {
        orderNo: 'order-runtime-1',
        status: 10,
        statusText: '待支付',
        totalAmount: '3280.00',
        discountAmount: '0.00',
        serviceFee: '0.00',
        shipFee: '0.00',
        payAmount: '3280.00',
        couponInfo: '',
        serviceItems: '',
        contactName: '验收人',
        contactPhone: '13800000000',
        remark: '',
        shipMethod: '配送',
        shipAddress: '占位地址',
        shipStatus: 0,
        shipNo: '',
        expireAt: '2026-09-21T12:00:00Z',
        createdAt: '2026-09-21T10:00:00Z',
        items: [{ productId: 'sku-p1', productTitle: product.title, productImage: '', breedName: '占位品种', price: '3280.00', quantity: 1, skuSpecs: '标准方案' }],
      })
    }
    if (url.pathname.startsWith('/api/')) return send(res, { list: [] })

    let filePath = path.join(ROOT, decodeURIComponent(url.pathname))
    if (!filePath.startsWith(ROOT)) {
      res.writeHead(403)
      return res.end()
    }
    if (!fs.existsSync(filePath) || fs.statSync(filePath).isDirectory()) {
      filePath = path.join(ROOT, 'index.html')
    }
    res.writeHead(200, { 'Content-Type': mime[path.extname(filePath)] || 'application/octet-stream' })
    fs.createReadStream(filePath).pipe(res)
  })
})

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

async function newPage(context) {
  const page = await context.newPage()
  page.setDefaultTimeout(10000)
  page.on('console', message => console.log(`CONSOLE ${message.type()}: ${message.text()}`))
  page.on('pageerror', error => console.log(`PAGEERROR ${error.stack || error.message}`))
  await page.addInitScript(() => {
    localStorage.setItem('pet_access', 'verify-access')
    localStorage.setItem('pet_refresh', 'verify-refresh')
  })
  return page
}

async function chooseSku(page) {
  await page.getByText('选择规格').first().waitFor()
  const sheet = page.locator('.sku-sheet-body')
  await sheet.getByText('标准方案').first().click()
  await page.screenshot({ path: path.join(OUTPUT, '01-sku-selected.png') })
  await sheet.getByText('完成选择').click()
}

async function loginIfRedirected(page) {
  const loginText = page.getByText('欢迎来到宠物之家')
  try {
    await loginText.waitFor({ timeout: 1500 })
  } catch {
    return false
  }
  await page.locator('input[placeholder="请输入手机号"]').fill('13800000000')
  await page.locator('input[placeholder="验证码"]').fill('123456')
  await page.getByText('获取验证码').click()
  await page.getByText('登 录').click()
  return true
}

async function run() {
  await new Promise(resolve => server.listen(PORT, '127.0.0.1', resolve))
  const browser = await chromium.launch({
    executablePath: 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
    headless: true,
    args: ['--no-sandbox', '--disable-dev-shm-usage', '--disable-gpu'],
  })
  const results = {}

  const context1 = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const buyPage = await newPage(context1)
  const buyUrl = `http://127.0.0.1:${PORT}/pages/detail/index?id=sku-p1`
  await buyPage.goto(buyUrl)
  await buyPage.getByText(product.title).first().waitFor()
  await buyPage.locator('.detail-actionbar .btn-buy').getByText('立即购买').click({ force: true })
  await buyPage.waitForTimeout(1000)
  await buyPage.screenshot({ path: path.join(OUTPUT, 'debug-after-buy-click.png'), fullPage: true })
  console.log('SKU BODY COUNT', await buyPage.locator('.sku-sheet-body').count())
  console.log('ACTIONBAR HTML', await buyPage.locator('.detail-actionbar').innerHTML())
  await chooseSku(buyPage)
  await buyPage.waitForURL(/\/pages\/checkout\/index\?id=sku-p1&skuId=sku-1/)
  await loginIfRedirected(buyPage)
  await buyPage.getByText('联系人信息').waitFor()
  results.skuBuy = {
    finalUrl: buyPage.url(),
    assertion: '立即购买 -> 选择标准方案 -> 确认后自动进入 checkout，URL 含 id=sku-p1&skuId=sku-1',
  }
  await buyPage.screenshot({ path: path.join(OUTPUT, '02-sku-buy-checkout.png'), fullPage: true })
  await context1.close()

  const context2 = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const cartPage = await newPage(context2)
  await cartPage.goto(`http://127.0.0.1:${PORT}/pages/detail/index?id=sku-p1`)
  await cartPage.getByText(product.title).first().waitFor()
  await cartPage.locator('.detail-actionbar .btn-cart').click({ force: true })
  await cartPage.getByText('选择规格').first().waitFor()
  const cartSheet = cartPage.locator('.sku-sheet-body')
  await cartSheet.getByText('完整方案').first().click()
  await cartSheet.getByText('完成选择').click()
  await cartPage.getByText('已加入购物车').waitFor()
  const cartCall = apiCalls.find(item => item.method === 'POST' && item.path === '/api/cart')
  assert(cartCall && JSON.parse(cartCall.body).productId === 'sku-p1' && JSON.parse(cartCall.body).skuId === 'sku-2', '加购请求缺少正确商品/SKU ID')
  results.skuCart = { finalUrl: cartPage.url(), request: cartCall, assertion: '加入购物车 -> 选择完整方案 -> 确认后自动 POST /api/cart 并显示成功提示' }
  await cartPage.screenshot({ path: path.join(OUTPUT, '03-sku-cart-toast.png') })
  await context2.close()

  const context3 = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const checkoutPage = await newPage(context3)
  await checkoutPage.goto(`http://127.0.0.1:${PORT}/pages/checkout/index?id=sku-p1&skuId=sku-1`)
  await loginIfRedirected(checkoutPage)
  await checkoutPage.getByText('联系人信息').waitFor()
  await checkoutPage.locator('input[placeholder="联系人姓名"]').fill('验收人')
  await checkoutPage.locator('input[placeholder="联系手机号"]').fill('13800000000')
  await checkoutPage.getByText('提交订单').click()
  await checkoutPage.getByText('订单已创建').waitFor()
  await checkoutPage.screenshot({ path: path.join(OUTPUT, '04-order-created-modal.png') })
  const beforeUrl = checkoutPage.url()
  await checkoutPage.getByText('查看订单').last().click()
  await checkoutPage.waitForURL(/\/pages\/order-detail\/index\?orderNo=order-runtime-1/)
  await checkoutPage.getByText('待支付').waitFor()
  results.orderRedirect = {
    beforeUrl,
    finalUrl: checkoutPage.url(),
    assertion: '单确认订单弹窗 -> 查看订单后 URL 变为订单详情，且不再显示 checkout 联系人表单',
  }
  assert(!checkoutPage.url().includes('/pages/checkout/index'), '订单创建后仍停留在 checkout')
  await checkoutPage.waitForTimeout(1000)
  assert(await checkoutPage.getByText('联系人信息').count() === 0, '订单详情页仍渲染 checkout 联系人表单')
  await checkoutPage.screenshot({ path: path.join(OUTPUT, '05-order-detail.png') })
  await context3.close()

  const context4 = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, locale: 'zh-CN' })
  const reviewPage = await newPage(context4)
  await reviewPage.goto(`http://127.0.0.1:${PORT}/pages/detail/index?id=sku-p1`)
  await reviewPage.getByText(product.title).first().waitFor()
  await reviewPage.getByText('全部评价').click()
  await reviewPage.getByText('评价内容 p1-r10').waitFor({ state: 'attached' })
  await reviewPage.getByText('加载更多').click({ force: true })
  await reviewPage.getByText('评价内容 p1-r20').waitFor({ state: 'attached' })
  const reviewCalls = apiCalls.filter(item => item.method === 'GET' && item.path.startsWith('/api/products/sku-p1/reviews?'))
  assert(reviewCalls.length >= 2, '评价加载更多请求不足')
  assert(reviewCalls[0].path === '/api/products/sku-p1/reviews?limit=10', '评价首页接口路径不正确')
  assert(reviewCalls[1].path === '/api/products/sku-p1/reviews?limit=10&cursor=p1-r10', '评价加载更多接口路径不正确')
  results.reviewPagination = {
    requests: reviewCalls.slice(0, 2).map(item => item.path),
    assertion: '首页评价请求 product=sku-p1；加载更多继续使用 product=sku-p1 和 cursor=p1-r10，页面出现 p1-r20',
  }
  await reviewPage.screenshot({ path: path.join(OUTPUT, '06-review-load-more.png'), fullPage: true })
  await context4.close()

  await browser.close()
  server.close()
  fs.writeFileSync(path.join(OUTPUT, 'evidence.json'), JSON.stringify(results, null, 2))
  console.log('EVIDENCE\n' + JSON.stringify(results, null, 2))
}

run().catch(async error => {
  console.error(error)
  fs.writeFileSync(path.join(OUTPUT, 'evidence-error.json'), `${new Date().toISOString()} ${error.stack || error.message}`)
  process.exitCode = 1
  setTimeout(() => process.exit(process.exitCode), 100)
})
