import { apiClient } from '@/api/client'
import { paymentAPI } from '@/api/payment'
import type { ModelPlazaResponse } from '../types/modelPlaza'

/** Extension 对核心公开端点的唯一适配入口。 */
export async function loadModelPlaza(options?: { signal?: AbortSignal }): Promise<ModelPlazaResponse> {
  const plazaRequest = apiClient.get<Omit<ModelPlazaResponse, 'balance_recharge_multiplier'>>('/model-plaza', {
    signal: options?.signal
  })
  const paymentRequest = paymentAPI.getConfig()
  const [plaza, payment] = await Promise.allSettled([plazaRequest, paymentRequest])
  if (plaza.status === 'rejected') throw plaza.reason
  const multiplier = payment.status === 'fulfilled' ? Number(payment.value.data?.balance_recharge_multiplier) : 10
  return {
    ...plaza.value.data,
    balance_recharge_multiplier: Number.isFinite(multiplier) && multiplier > 0 ? multiplier : 10
  }
}
