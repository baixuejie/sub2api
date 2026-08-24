import { apiClient } from '@/api/client'
import { paymentAPI } from '@/api/payment'
import type { ModelPlazaResponse } from '../types/modelPlaza'

/** Extension 对核心公开端点的唯一适配入口。 */
export async function loadModelPlaza(options?: { signal?: AbortSignal }): Promise<ModelPlazaResponse> {
  const plazaRequest = apiClient.get<Omit<ModelPlazaResponse, 'balance_recharge_multiplier'>>('/model-plaza', {
    signal: options?.signal
  })
  // `/payment/config` is an authenticated user endpoint. Avoid calling it for
  // anonymous visitors because the shared 401 interceptor would otherwise
  // redirect a public model-plaza page to the login screen.
  const paymentRequest = localStorage.getItem('auth_token')
    ? paymentAPI.getConfig()
    : Promise.resolve<null>(null)
  const [plaza, payment] = await Promise.allSettled([plazaRequest, paymentRequest])
  if (plaza.status === 'rejected') throw plaza.reason
  const multiplier = payment.status === 'fulfilled' ? Number(payment.value?.data?.balance_recharge_multiplier) : 10
  return {
    ...plaza.value.data,
    balance_recharge_multiplier: Number.isFinite(multiplier) && multiplier > 0 ? multiplier : 10
  }
}
