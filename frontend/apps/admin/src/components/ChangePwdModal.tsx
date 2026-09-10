import { Form, Input, message, Modal } from 'antd'
import { adminChangePassword } from '../api/admin'

export default function ChangePwdModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [form] = Form.useForm()

  const onOk = async () => {
    const v = await form.validateFields()
    try {
      await adminChangePassword(v.oldPassword, v.newPassword)
      message.success('密码已更新')
      form.resetFields()
      onClose()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  return (
    <Modal title="修改密码" open={open} onOk={onOk} onCancel={onClose} destroyOnClose>
      <Form form={form} layout="vertical">
        <Form.Item name="oldPassword" label="原密码" rules={[{ required: true, message: '请输入原密码' }]}>
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="newPassword"
          label="新密码（至少 8 位）"
          rules={[
            { required: true, message: '请输入新密码' },
            { min: 8, message: '至少 8 位' },
          ]}
        >
          <Input.Password />
        </Form.Item>
      </Form>
    </Modal>
  )
}
