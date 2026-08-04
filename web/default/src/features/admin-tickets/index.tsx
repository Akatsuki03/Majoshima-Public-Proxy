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
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { cn } from '@/lib/utils'

import {
  closeAdminTicket,
  getAdminTicket,
  getAdminTickets,
  replyAdminTicket,
} from '../tickets/api'
import {
  TICKET_CATEGORIES,
  TICKET_CLOSE_REASONS,
  type Ticket,
  type TicketCloseReason,
} from '../tickets/types'

function formatTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString()
}

function categoryLabel(t: (key: string) => string, category: string) {
  const found = TICKET_CATEGORIES.find((item) => item.value === category)
  return found ? t(found.labelKey) : category
}

export function AdminTicketsPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState<string>('open')
  const [category, setCategory] = useState<string>('all')
  const [userIdFilter, setUserIdFilter] = useState('')
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [reply, setReply] = useState('')
  const [closeOpen, setCloseOpen] = useState(false)
  const [closeReason, setCloseReason] =
    useState<TicketCloseReason>('resolved')
  const [closeMessage, setCloseMessage] = useState('')

  const listQuery = useQuery({
    queryKey: ['admin-tickets', page, status, category, userIdFilter],
    queryFn: () =>
      getAdminTickets({
        p: page,
        page_size: 20,
        status: status === 'all' ? undefined : status,
        category: category === 'all' ? undefined : category,
        user_id: userIdFilter ? Number(userIdFilter) : undefined,
      }),
  })

  const detailQuery = useQuery({
    queryKey: ['admin-ticket', selectedId],
    queryFn: () => {
      if (selectedId == null) {
        return Promise.reject(new Error('no ticket selected'))
      }
      return getAdminTicket(selectedId)
    },
    enabled: selectedId != null,
  })

  const replyMutation = useMutation({
    mutationFn: () => {
      if (selectedId == null) {
        return Promise.reject(new Error('no ticket selected'))
      }
      return replyAdminTicket(selectedId, reply)
    },
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Failed to send reply'))
        return
      }
      toast.success(t('Reply sent'))
      setReply('')
      queryClient.invalidateQueries({ queryKey: ['admin-ticket', selectedId] })
      queryClient.invalidateQueries({ queryKey: ['admin-tickets'] })
    },
  })

  const closeMutation = useMutation({
    mutationFn: () => {
      if (selectedId == null) {
        return Promise.reject(new Error('no ticket selected'))
      }
      return closeAdminTicket(selectedId, {
        close_reason: closeReason,
        close_message: closeMessage || undefined,
      })
    },
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Failed to close ticket'))
        return
      }
      toast.success(t('Ticket closed'))
      setCloseOpen(false)
      setCloseMessage('')
      queryClient.invalidateQueries({ queryKey: ['admin-ticket', selectedId] })
      queryClient.invalidateQueries({ queryKey: ['admin-tickets'] })
    },
  })

  const tickets = listQuery.data?.data?.items ?? []
  const total = listQuery.data?.data?.total ?? 0
  const selected = detailQuery.data?.data

  return (
    <SectionPageLayout fixedContent>
      <SectionPageLayout.Title>{t('Tickets')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mb-4 flex flex-wrap gap-2'>
          <Select
            value={status}
            onValueChange={(value) => {
              if (value) setStatus(value)
            }}
          >
            <SelectTrigger className='w-[140px]'>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='all'>{t('All statuses')}</SelectItem>
              <SelectItem value='open'>{t('Open')}</SelectItem>
              <SelectItem value='closed'>{t('Closed')}</SelectItem>
            </SelectContent>
          </Select>
          <Select
            value={category}
            onValueChange={(value) => {
              if (value) setCategory(value)
            }}
          >
            <SelectTrigger className='w-[180px]'>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='all'>{t('All categories')}</SelectItem>
              {TICKET_CATEGORIES.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {t(item.labelKey)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Input
            className='w-[140px]'
            placeholder={t('User ID')}
            value={userIdFilter}
            onChange={(e) => setUserIdFilter(e.target.value)}
          />
        </div>

        <div className='grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)]'>
          <div className='space-y-2'>
            {tickets.map((ticket) => (
              <button
                key={ticket.id}
                type='button'
                onClick={() => setSelectedId(ticket.id)}
                className={cn(
                  'hover:bg-muted/50 w-full rounded-lg border p-3 text-left transition-colors',
                  selectedId === ticket.id && 'border-primary bg-muted/40'
                )}
              >
                <div className='flex items-center justify-between gap-2'>
                  <span className='truncate font-medium'>{ticket.title}</span>
                  <Badge
                    variant={
                      ticket.status === 'open' ? 'default' : 'secondary'
                    }
                  >
                    {ticket.status === 'open' ? t('Open') : t('Closed')}
                  </Badge>
                </div>
                <div className='text-muted-foreground mt-1 text-xs'>
                  #{ticket.id} · {ticket.username || ticket.user_id} ·{' '}
                  {categoryLabel(t, ticket.category)} ·{' '}
                  {formatTime(ticket.created_at)}
                </div>
              </button>
            ))}
            {total > 20 && (
              <div className='flex items-center justify-between pt-2'>
                <Button
                  variant='outline'
                  size='sm'
                  disabled={page <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  {t('Previous')}
                </Button>
                <Button
                  variant='outline'
                  size='sm'
                  disabled={page * 20 >= total}
                  onClick={() => setPage((p) => p + 1)}
                >
                  {t('Next')}
                </Button>
              </div>
            )}
          </div>

          <div className='rounded-lg border p-4'>
            {!selected ? (
              <p className='text-muted-foreground text-sm'>
                {t('Select a ticket to view details.')}
              </p>
            ) : (
              <AdminTicketDetail
                ticket={selected}
                reply={reply}
                onReplyChange={setReply}
                onSend={() => replyMutation.mutate()}
                sending={replyMutation.isPending}
                onClose={() => setCloseOpen(true)}
              />
            )}
          </div>
        </div>
      </SectionPageLayout.Content>

      <Dialog open={closeOpen} onOpenChange={setCloseOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Close Ticket')}</DialogTitle>
          </DialogHeader>
          <div className='space-y-3'>
            <div className='space-y-1.5'>
              <Label>{t('Close reason')}</Label>
              <Select
                value={closeReason}
                onValueChange={(value) => {
                  if (value) setCloseReason(value as TicketCloseReason)
                }}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TICKET_CLOSE_REASONS.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {t(item.labelKey)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className='space-y-1.5'>
              <Label>
                {closeReason === 'custom'
                  ? t('Custom reply')
                  : t('Message (optional)')}
              </Label>
              <Textarea
                value={closeMessage}
                onChange={(e) => setCloseMessage(e.target.value)}
                rows={4}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setCloseOpen(false)}>
              {t('Cancel')}
            </Button>
            <Button
              disabled={
                closeMutation.isPending ||
                (closeReason === 'custom' && !closeMessage.trim())
              }
              onClick={() => closeMutation.mutate()}
            >
              {t('Close Ticket')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </SectionPageLayout>
  )
}

function AdminTicketDetail(props: {
  ticket: Ticket
  reply: string
  onReplyChange: (value: string) => void
  onSend: () => void
  sending: boolean
  onClose: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className='flex h-full flex-col gap-3'>
      <div className='flex items-start justify-between gap-2'>
        <div>
          <h3 className='font-semibold'>{props.ticket.title}</h3>
          <p className='text-muted-foreground mt-1 text-xs'>
            #{props.ticket.id} · {props.ticket.username || props.ticket.user_id}{' '}
            · {categoryLabel(t, props.ticket.category)}
          </p>
        </div>
        {props.ticket.status === 'open' && (
          <Button variant='outline' size='sm' onClick={props.onClose}>
            {t('Close Ticket')}
          </Button>
        )}
      </div>
      <div className='min-h-0 flex-1 space-y-3 overflow-y-auto'>
        {(props.ticket.messages ?? []).map((message) => (
          <div
            key={message.id}
            className={cn(
              'rounded-md border p-3 text-sm',
              message.is_admin && 'bg-muted/40'
            )}
          >
            <div className='text-muted-foreground mb-1 text-xs'>
              {message.is_admin
                ? message.username || t('Admin')
                : message.username || t('User')}{' '}
              · {formatTime(message.created_at)}
            </div>
            <div className='whitespace-pre-wrap'>{message.content}</div>
          </div>
        ))}
      </div>
      {props.ticket.status === 'open' && (
        <div className='space-y-2 border-t pt-3'>
          <Textarea
            value={props.reply}
            onChange={(e) => props.onReplyChange(e.target.value)}
            placeholder={t('Write a reply...')}
            rows={3}
          />
          <Button
            disabled={!props.reply.trim() || props.sending}
            onClick={props.onSend}
          >
            {t('Send Reply')}
          </Button>
        </div>
      )}
    </div>
  )
}
