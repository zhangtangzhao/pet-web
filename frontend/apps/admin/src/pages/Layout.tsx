import { useState } from 'react'
import { Layout, Menu, Dropdown, message } from 'antd'
import {
  DashboardOutlined,
  ShoppingOutlined,
  TagsOutlined,
  RobotOutlined,
  GiftOutlined,
  ProfileOutlined,
  TeamOutlined,
  LogoutOutlined,
  KeyOutlined,
} from '@ant-design/icons'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { adminLogout } from '../api/admin'
import ChangePwdModal from '../components/ChangePwdModal'

const items = [
  { key: '/', icon: <DashboardOutlined />, label: '运营看板' },
  { key: '/products', icon: <ShoppingOutlined />, label: '宠物商品' },
  { key: '/catalog', icon: <TagsOutlined />, label: '分类/品种' },
  { key: '/knowledge', icon: <RobotOutlined />, label: 'AI 知识库' },
  { key: '/marketing', icon: <GiftOutlined />, label: '营销管理' },
  { key: '/orders', icon: <ProfileOutlined />, label: '订单管理' },
  { key: '/members', icon: <TeamOutlined />, label: '会员管理' },
]

export default function AdminLayout() {
  const nav = useNavigate()
  const { pathname } = useLocation()
  const [pwdOpen, setPwdOpen] = useState(false)

  const onLogout = async () => {
    await adminLogout()
    message.success('已退出')
    nav('/login')
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Layout.Sider theme="light" width={200} style={{ borderRight: '1px solid #f0f0f0' }}>
        <div
          style={{
            height: 56,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 700,
            fontSize: 16,
            color: '#ff7a45',
          }}
        >
          🐾 宠物运营后台
        </div>
        <Menu
          mode="inline"
          selectedKeys={[pathname]}
          items={items}
          onClick={({ key }) => nav(key)}
          style={{ borderInlineEnd: 'none' }}
        />
      </Layout.Sider>
      <Layout>
        <Layout.Header
          style={{
            background: '#fff',
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            paddingInline: 24,
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <Dropdown
            menu={{
              items: [
                { key: 'pwd', icon: <KeyOutlined />, label: '修改密码', onClick: () => setPwdOpen(true) },
                { type: 'divider' },
                { key: 'out', icon: <LogoutOutlined />, label: '退出登录', onClick: onLogout },
              ],
            }}
          >
            <a onClick={(e) => e.preventDefault()}>管理员 ▾</a>
          </Dropdown>
        </Layout.Header>
        <Layout.Content style={{ padding: 24, background: '#f5f5f5' }}>
          <Outlet />
        </Layout.Content>
      </Layout>
      <ChangePwdModal open={pwdOpen} onClose={() => setPwdOpen(false)} />
    </Layout>
  )
}
