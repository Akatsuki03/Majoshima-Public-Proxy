/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { api } from '@/lib/api'

import type {
  ApiResponse,
  Ticket,
  TicketCloseReason,
  TicketListData,
  TicketMessage,
} from './types'

export async function getSelfTickets(params: {
  p?: number
  page_size?: number
}): Promise<ApiResponse<TicketListData>> {
  const search = new URLSearchParams()
  search.set('p', String(params.p ?? 1))
  search.set('page_size', String(params.page_size ?? 20))
  const res = await api.get(`/api/ticket/self?${search.toString()}`)
  return res.data
}

export async function createSelfTicket(data: {
  category: string
  title: string
  content: string
}): Promise<ApiResponse<Ticket>> {
  const res = await api.post('/api/ticket/self', data)
  return res.data
}

export async function getSelfTicket(
  id: number
): Promise<ApiResponse<Ticket>> {
  const res = await api.get(`/api/ticket/self/${id}`)
  return res.data
}

export async function replySelfTicket(
  id: number,
  content: string
): Promise<ApiResponse<TicketMessage>> {
  const res = await api.post(`/api/ticket/self/${id}/reply`, { content })
  return res.data
}

export async function getAdminTickets(params: {
  p?: number
  page_size?: number
  status?: string
  category?: string
  user_id?: number
}): Promise<ApiResponse<TicketListData>> {
  const search = new URLSearchParams()
  search.set('p', String(params.p ?? 1))
  search.set('page_size', String(params.page_size ?? 20))
  if (params.status) search.set('status', params.status)
  if (params.category) search.set('category', params.category)
  if (params.user_id) search.set('user_id', String(params.user_id))
  const res = await api.get(`/api/ticket/?${search.toString()}`)
  return res.data
}

export async function getAdminTicket(
  id: number
): Promise<ApiResponse<Ticket>> {
  const res = await api.get(`/api/ticket/${id}`)
  return res.data
}

export async function replyAdminTicket(
  id: number,
  content: string
): Promise<ApiResponse<TicketMessage>> {
  const res = await api.post(`/api/ticket/${id}/reply`, { content })
  return res.data
}

export async function closeAdminTicket(
  id: number,
  data: { close_reason: TicketCloseReason; close_message?: string }
): Promise<ApiResponse> {
  const res = await api.post(`/api/ticket/${id}/close`, data)
  return res.data
}
