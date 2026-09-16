import { useEffect, useRef, useState } from 'react'
import { Button, Input, Text, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { post, setTokens } from '../../request'
import { LoginResp } from '../../types'
import './index.css'

export default function Login() {
  const { params } = useRouter()
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [countdown, setCountdown] = useState(0)
  const timer = useRef<ReturnType<typeof setInterval>>()

  useEffect(() => () => clearInterval(timer.current), [])

  const validPhone = /^1\d{10}$/.test(phone)

  const sendCode = async () => {
    if (!validPhone || countdown > 0) return
    try {
      await post('/auth/sms/send', { phone })
      Taro.showToast({ title: '验证码已发送', icon: 'success' })
      setCountdown(60)
      timer.current = setInterval(() => {
        setCountdown((c) => {
          if (c <= 1) clearInterval(timer.current)
          return c - 1
        })
      }, 1000)
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    }
  }

  const login = async () => {
    if (!validPhone || !code) {
      Taro.showToast({ title: '请填写手机号和验证码', icon: 'none' })
      return
    }
    try {
      // 分享带参归因：携带缓存的邀请码（后端仅对新注册生效）
      const inviteCode = (Taro.getStorageSync('pet_invite_code') as string) || ''
      const r = await post<LoginResp>('/auth/sms/login', { phone, code, inviteCode })
      setTokens(r.accessToken, r.refreshToken)
      Taro.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => {
        if (params.redirect) Taro.redirectTo({ url: decodeURIComponent(params.redirect) })
        else Taro.navigateBack()
      }, 600)
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    }
  }

  return (
    <View className='login'>
      <View className='login-title'>
        <Text className='login-hello'>你好，</Text>
        <Text className='login-hello'>欢迎来到宠物之家</Text>
      </View>
      <View className='login-form card'>
        <Input
          className='login-input'
          type='number'
          maxlength={11}
          placeholder='请输入手机号'
          value={phone}
          onInput={(e) => setPhone(e.detail.value)}
        />
        <View className='login-code-row'>
          <Input
            className='login-input'
            type='number'
            maxlength={6}
            placeholder='验证码'
            value={code}
            onInput={(e) => setCode(e.detail.value)}
          />
          <View className={`send ${validPhone && countdown === 0 ? 'send-on' : ''}`} onClick={sendCode}>
            <Text>{countdown > 0 ? `${countdown}s 后重发` : '获取验证码'}</Text>
          </View>
        </View>
        <Button className='btn-login' disabled={!validPhone || !code} onClick={login}>
          登 录
        </Button>
        <Text className='login-tip'>未注册的手机号将自动创建账号</Text>
      </View>
    </View>
  )
}
