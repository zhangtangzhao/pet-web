const fs = require('fs')
const path = require('path')
const crypto = require('crypto')
const { chromium } = require('playwright')

const BASE = 'http://127.0.0.1:10087'
const OUTPUT = path.resolve(__dirname, 'user-access')
const PRODUCT_ID = '3005'
fs.mkdirSync(OUTPUT, { recursive: true })

function assert(condition, message) {
  if (!condition) throw new Error(message)
}

function sha12(buffer) {
  return crypto.createHash('sha256').update(buffer).digest('hex').slice(0, 12)
}

async function newPage(browser, viewport) {
  const context = await browser.newContext({
    viewport,
    deviceScaleFactor: 2,
    locale: 'zh-CN',
    serviceWorkers: 'block',
  })
  const page = await context.newPage()
  page.setDefaultTimeout(15000)
  const cdp = await context.newCDPSession(page)
  await cdp.send('Network.setCacheDisabled', { cacheDisabled: true })
  await cdp.send('Network.setBypassServiceWorker', { bypass: true })
  return { context, page }
}

async function hardLoad(page, url) {
  await page.goto(url, { waitUntil: 'domcontentloaded' })
  const response = await page.reload({ waitUntil: 'domcontentloaded' })
  const body = await response.body()
  return {
    url: response.url(),
    status: response.status(),
    cacheControl: response.headers()['cache-control'] || '',
    fromServiceWorker: response.fromServiceWorker(),
    sha12: sha12(body),
  }
}

function listenForAssets(page) {
  const wanted = new Set(['/css/app.css', '/css/120.css', '/js/app.js', '/chunk/120.js'])
  const promises = []
  page.on('response', response => {
    let pathname = ''
    try {
      pathname = new URL(response.url()).pathname
    } catch {
      return
    }
    if (!wanted.has(pathname)) return
    promises.push(Promise.resolve({
      path: pathname,
      status: response.status(),
      fromServiceWorker: response.fromServiceWorker(),
      cacheControl: response.headers()['cache-control'] || '',
    }))
  })
  return async () => await Promise.all(promises)
}

async function fetchAssetFingerprints(context) {
  const paths = ['/css/app.css', '/css/120.css', '/js/app.js', '/chunk/120.js']
  return await Promise.all(paths.map(async pathname => {
    const response = await context.request.get(`${BASE}${pathname}`, {
      headers: { 'cache-control': 'no-cache, max-age=0', pragma: 'no-cache' },
    })
    return {
      path: pathname,
      status: response.status(),
      cacheControl: response.headers()['cache-control'] || '',
      sha12: sha12(await response.body()),
    }
  }))
}

async function checkBack(page, label) {
  await page.locator('.detail-navbar').waitFor()
  const back = page.locator('.detail-back')
  await back.waitFor()
  await page.screenshot({ path: path.join(OUTPUT, `390-detail-back-${label}.png`) })
  const box = await back.boundingBox()
  const text = (await page.locator('.detail-back-text').textContent()).trim()
  assert(box && box.width >= 44 && box.height >= 44, `${label} 返回触控区不足 44px`)
  assert(text === '返回', `${label} 返回文案缺失`)
  return { width: Math.round(box.width), height: Math.round(box.height), text }
}

async function run() {
  const browser = await chromium.launch({
    executablePath: 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
    headless: true,
    args: ['--no-sandbox', '--disable-dev-shm-usage', '--disable-gpu'],
  })
  const evidence = { server: 'dist-p0 via http://127.0.0.1:10087', urls: {}, hardRefresh: {}, assets: [], navigation: {} }

  const mobile = await newPage(browser, { width: 390, height: 844 })
  const collectMobileAssets = listenForAssets(mobile.page)
  evidence.urls.mobileHome = `${BASE}/pages/index/index`
  evidence.hardRefresh.mobileHome = await hardLoad(mobile.page, evidence.urls.mobileHome)
  await mobile.page.locator('.home').waitFor()
  await mobile.page.locator('.pet-card').first().waitFor()
  await mobile.page.screenshot({ path: path.join(OUTPUT, '390-home.png'), fullPage: true })

  evidence.urls.mobileList = `${BASE}/pages/list/index`
  evidence.hardRefresh.mobileList = await hardLoad(mobile.page, evidence.urls.mobileList)
  await mobile.page.locator('.pet-card').first().waitFor()
  await mobile.page.screenshot({ path: path.join(OUTPUT, '390-list.png'), fullPage: true })

  evidence.urls.mobileDetail = `${BASE}/pages/detail/index?id=${PRODUCT_ID}`
  evidence.hardRefresh.mobileDetail = await hardLoad(mobile.page, evidence.urls.mobileDetail)
  await mobile.page.waitForLoadState('networkidle')
  evidence.assets = await collectMobileAssets()
  evidence.assetFingerprints = await fetchAssetFingerprints(mobile.context)
  evidence.ui = await checkBack(mobile.page, 'direct-real')
  evidence.ui.brandVivid = (await mobile.page.evaluate(() => getComputedStyle(document.body).getPropertyValue('--brand-vivid').trim())) || ''
  assert(/^#F4623A$/i.test(evidence.ui.brandVivid), '新 Token --brand-vivid 未生效')
  await mobile.page.screenshot({ path: path.join(OUTPUT, '390-detail.png'), fullPage: true })
  await mobile.page.locator('.detail-back').click()
  await mobile.page.waitForURL(`${BASE}/pages/index/index`)
  await mobile.page.locator('.home').waitFor()
  evidence.navigation.directFallback = mobile.page.url()
  await mobile.page.screenshot({ path: path.join(OUTPUT, '390-after-direct-back.png') })
  await mobile.context.close()

  const listRun = await newPage(browser, { width: 390, height: 844 })
  await listRun.page.goto(`${BASE}/pages/list/index`, { waitUntil: 'domcontentloaded' })
  await listRun.page.locator('.pet-card').first().waitFor()
  await listRun.page.locator('.pet-card').first().click()
  await listRun.page.locator('.detail-back').waitFor()
  evidence.navigation.listBack = await checkBack(listRun.page, 'list-real')
  await listRun.page.locator('.detail-back').click()
  await listRun.page.waitForURL(`${BASE}/pages/list/index`)
  evidence.navigation.listBackUrl = listRun.page.url()
  await listRun.context.close()

  const searchRun = await newPage(browser, { width: 390, height: 844 })
  await searchRun.page.goto(`${BASE}/pages/list/index`, { waitUntil: 'domcontentloaded' })
  await searchRun.page.locator('.pet-card').first().waitFor()
  await searchRun.page.waitForTimeout(3000)
  await searchRun.page.locator('.lst-input input').fill('柯基')
  await searchRun.page.locator('.lst-input input').press('Enter')
  await searchRun.page.locator('.lst-cat').first().click({ force: true })
  await searchRun.page.waitForFunction(() => {
    const cards = document.querySelectorAll('.pet-card').length
    const count = Array.from(document.querySelectorAll('.lst-count')).map(node => node.textContent || '').join('')
    return cards === 1 && count.includes('共 1 只')
  })
  await searchRun.page.waitForTimeout(1000)
  await searchRun.page.getByText('柯基妹妹 双色萌爆').click({ force: true })
  await searchRun.page.locator('.detail-back').waitFor()
  evidence.navigation.searchBack = await checkBack(searchRun.page, 'search-real')
  await searchRun.page.locator('.detail-back-text').click({ force: true })
  await searchRun.page.waitForTimeout(1000)
  await searchRun.page.locator('.lst').waitFor()
  evidence.navigation.searchBackUrl = searchRun.page.url()
  assert(new URL(searchRun.page.url()).pathname === '/pages/list/index', '搜索返回路径不正确')
  evidence.navigation.searchKeywordPreserved = await searchRun.page.locator('.lst-input input').inputValue()
  assert(evidence.navigation.searchKeywordPreserved === '柯基', '搜索返回丢失关键字')
  evidence.navigation.searchBackCardCount = await searchRun.page.locator('.pet-card').count()
  assert(evidence.navigation.searchBackCardCount > 0, '搜索返回后结果卡片丢失')
  await searchRun.page.screenshot({ path: path.join(OUTPUT, '390-back-search-list.png') })
  await searchRun.context.close()

  const relatedRun = await newPage(browser, { width: 390, height: 844 })
  await relatedRun.page.goto(`${BASE}/pages/detail/index?id=${PRODUCT_ID}`, { waitUntil: 'domcontentloaded' })
  await relatedRun.page.locator('.detail-back').waitFor()
  await relatedRun.page.locator('.related-item').first().waitFor()
  await relatedRun.page.locator('.related-item').first().click()
  await relatedRun.page.waitForURL(/\/pages\/detail\/index\?id=\d+/)
  evidence.navigation.recommendBack = await checkBack(relatedRun.page, 'recommend-real')
  evidence.navigation.recommendUrl = relatedRun.page.url()
  await relatedRun.context.close()

  const tablet = await newPage(browser, { width: 768, height: 1024 })
  await tablet.page.goto(`${BASE}/pages/detail/index?id=${PRODUCT_ID}`, { waitUntil: 'domcontentloaded' })
  await tablet.page.locator('.detail-back').waitFor()
  await tablet.page.screenshot({ path: path.join(OUTPUT, '768-detail.png'), fullPage: true })
  await tablet.context.close()

  const desktop = await newPage(browser, { width: 1440, height: 900 })
  await desktop.page.goto(`${BASE}/pages/index/index`, { waitUntil: 'domcontentloaded' })
  await desktop.page.locator('.pet-card').first().waitFor()
  await desktop.page.screenshot({ path: path.join(OUTPUT, '1440-home.png'), fullPage: true })
  await desktop.page.goto(`${BASE}/pages/detail/index?id=${PRODUCT_ID}`, { waitUntil: 'domcontentloaded' })
  await desktop.page.locator('.detail-back').waitFor()
  await desktop.page.screenshot({ path: path.join(OUTPUT, '1440-detail.png'), fullPage: true })
  await desktop.context.close()

  await browser.close()
  evidence.checkedAt = new Date().toISOString()
  fs.writeFileSync(path.join(OUTPUT, 'evidence.json'), JSON.stringify(evidence, null, 2))
  console.log(JSON.stringify(evidence, null, 2))
}

run().catch(error => {
  console.error(error)
  fs.writeFileSync(path.join(OUTPUT, 'evidence-error.json'), `${new Date().toISOString()} ${error.stack || error.message}`)
  process.exitCode = 1
})
