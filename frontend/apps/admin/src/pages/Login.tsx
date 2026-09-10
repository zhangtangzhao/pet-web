import { useEffect, useState } from 'react'
import { Button, Card, Form, Input, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { adminLogin, getCaptcha } from '../api/admin'
import { ACCESS_KEY } from '../api/client'

export default function Login() {
  const nav = useNavigate()
  const [loading, setLoading] = useState(false)
  const [captcha, setCaptcha] = useState<{ captchaId: string; imageB64: string }>()

  const loadCaptcha = () => getCaptcha().then(setCaptcha).catch(() => {})
  useEffect(() => {
    loadCaptcha()
  }, [])

  const onFinish = async (v: { username: string; password: string; captchaCode?: string }) => {
    setLoading(true)
    try {
      await adminLogin(v.username, v.password, captcha?.captchaId, v.captchaCode)
      message.success('登录成功')
      nav('/')
    } catch (e: any) {
      message.error(e.message)
      loadCaptcha()
    } finally {
      setLoading(false)
    }
  }

  if (localStorage.getItem(ACCESS_KEY)) {
    nav('/')
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #fff4ee 0%, #ffe6d9 100%)',
      }}
    >
      <Card title="宠物交易平台 · 运营后台" style={{ width: 380 }} headStyle={{ textAlign: 'center', fontSize: 18 }}>
        <Form onFinish={onFinish} size="large">
          <Form.Item name="username" rules={[{ required: true, message: '请输入账号' }]}>
            <Input placeholder="账号" autoComplete="username" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="密码" autoComplete="current-password" />
          </Form.Item>
          <Form.Item>
            <div style={{ display: 'flex', gap: 8 }}>
              <Form.Item name="captchaCode" noStyle>
                <Input placeholder="图形验证码（可选）" style={{ flex: 1 }} />
              </Form.Item>
              {captcha && (
                <img
                  src={captcha.imageB64}
                  alt="验证码"
                  title="点击刷新"
                  onClick={loadCaptcha}
                  style={{ height: 40, cursor: 'pointer', borderRadius: 6 }}
                />
              )}
            </div>
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading}>
            登 录
          </Button>
        </Form>
      </Card>
    </div>
  )
}
