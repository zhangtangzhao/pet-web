import { Navigate, Route, Routes } from 'react-router-dom'
import AdminLayout from './pages/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Products from './pages/Products'
import Catalog from './pages/Catalog'
import Orders from './pages/Orders'
import Members from './pages/Members'
import Support from './pages/Support'
import Knowledge from './pages/Knowledge'
import Marketing from './pages/Marketing'
import Reviews from './pages/Reviews'
import AfterSale from './pages/AfterSale'
import Banners from './pages/Banners'
import ShipMethods from './pages/ShipMethods'
import Finance from './pages/Finance'
import Suppliers from './pages/Suppliers'
import FlashSales from './pages/FlashSales'
import Groups from './pages/Groups'
import Stores from './pages/Stores'
import PointsShop from './pages/PointsShop'
import CalendarPage from './pages/Calendar'
import BigScreen from './pages/BigScreen'
import Bargain from './pages/Bargain'
import Auction from './pages/Auction'
import Bookings from './pages/Bookings'
import InsuranceApplies from './pages/InsuranceApplies'
import Report from './pages/Report'
import Ops from './pages/Ops'

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<AdminLayout />}>
        <Route index element={<Dashboard />} />
        <Route path="products" element={<Products />} />
        <Route path="catalog" element={<Catalog />} />
        <Route path="knowledge" element={<Knowledge />} />
        <Route path="marketing" element={<Marketing />} />
        <Route path="flash-sales" element={<FlashSales />} />
        <Route path="group-buys" element={<Groups />} />
        <Route path="stores" element={<Stores />} />
        <Route path="points-shop" element={<PointsShop />} />
        <Route path="activity-calendar" element={<CalendarPage />} />
        <Route path="bigscreen" element={<BigScreen />} />
        <Route path="bargain" element={<Bargain />} />
        <Route path="auction" element={<Auction />} />
        <Route path="bookings" element={<Bookings />} />
        <Route path="insurance-applies" element={<InsuranceApplies />} />
        <Route path="banner" element={<Banners />} />
        <Route path="shipping" element={<ShipMethods />} />
        <Route path="orders" element={<Orders />} />
        <Route path="finance" element={<Finance />} />
        <Route path="suppliers" element={<Suppliers />} />
        <Route path="reviews" element={<Reviews />} />
        <Route path="aftersale" element={<AfterSale />} />
        <Route path="members" element={<Members />} />
        <Route path="report" element={<Report />} />
        <Route path="ops" element={<Ops />} />
        <Route path="support" element={<Support />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}


