import { Navigate, Route, Routes } from 'react-router-dom'
import AdminLayout from './pages/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Products from './pages/Products'
import Catalog from './pages/Catalog'
import Orders from './pages/Orders'
import Members from './pages/Members'
import Knowledge from './pages/Knowledge'
import Marketing from './pages/Marketing'

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
        <Route path="orders" element={<Orders />} />
        <Route path="members" element={<Members />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
