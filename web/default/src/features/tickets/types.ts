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
export type TicketCategory =
  | 'special_usage'
  | 'tool_call'
  | 'bug'
  | 'other'

export type TicketStatus = 'open' | 'closed'

export type TicketCloseReason =
  | 'resolved'
  | 'unresolved'
  | 'invalid'
  | 'custom'

export type TicketMessage = {
  id: number
  ticket_id: number
  user_id: number
  is_admin: boolean
  content: string
  created_at: number
  username?: string
}

export type Ticket = {
  id: number
  user_id: number
  category: TicketCategory
  title: string
  status: TicketStatus
  close_reason?: string
  created_at: number
  updated_at: number
  closed_at?: number
  username?: string
  messages?: TicketMessage[]
}

export type TicketListData = {
  items: Ticket[]
  total: number
  page: number
  page_size: number
  ticket_disabled?: boolean
  enabled?: boolean
  daily_limit?: number
}

export type ApiResponse<T = unknown> = {
  success: boolean
  message?: string
  data?: T
}

export const TICKET_CATEGORIES: {
  value: TicketCategory
  labelKey: string
}[] = [
  { value: 'special_usage', labelKey: 'Special usage declaration' },
  { value: 'tool_call', labelKey: 'Tool call declaration' },
  { value: 'bug', labelKey: 'Bug report' },
  { value: 'other', labelKey: 'Other' },
]

export const TICKET_CLOSE_REASONS: {
  value: TicketCloseReason
  labelKey: string
}[] = [
  { value: 'resolved', labelKey: 'Resolved' },
  { value: 'unresolved', labelKey: 'Unresolved' },
  { value: 'invalid', labelKey: 'Invalid ticket' },
  { value: 'custom', labelKey: 'Custom reply' },
]
