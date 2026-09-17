import { useEffect, useState } from 'react'
import { Button, Card, Input, message } from 'antd'
import { client } from '../api/client'
export default function HomeConfig() {
  const [config, setConfig] = useState('[]')
  const [loading, setLoading] = useState(false)
  useEffect(() => { client.get('/admin/home-config').then((r: any) => setConfig(r?.config ?? '[]')).catch(() => {}) }, [])
  const save = async () => {
    setLoading(true)
    try { JSON.parse(config); await client.put('/admin/home-config', { config }); message.success('已保存') }
    catch (e: any) { message.error(e?.message || 'JSON 格式错误') } finally { setLoading(false) }
  }
  return (
    <Card title="首页装修配置（JSON 楼层配置）" extra={<Button type="primary" loading={loading} onClick={save}>保存</Button>}>
      <Input.TextArea rows={20} value={config} onChange={(e) => setConfig(e.target.value)} style={{ fontFamily: 'monospace' }} />
    </Card>
  )
}
