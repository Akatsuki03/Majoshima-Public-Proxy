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
  createSelfTicket,
  deleteSelfTicket,
  getSelfTicket,
  getSelfTickets,
  replySelfTicket,
} from './api'
import { TICKET_CATEGORIES, TICKET_CLOSE_REASONS, type Ticket, type TicketCategory } from './types'

function formatTime(ts: number) {
  if (!ts) return '-'
  return new Date(ts * 1000).toLocaleString()
}

function categoryLabel(t: (key: string) => string, category: string): string {
  const found = TICKET_CATEGORIES.find((item) => item.value === category)
  return found ? t(found.labelKey) : category
}

function closeReasonLabel(t: (key: string) => string, reason?: string): string {
  if (!reason) return ''
  const found = TICKET_CLOSE_REASONS.find((item) => item.value === reason)
  return found ? t(found.labelKey) : reason
}

export function TicketsPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [category, setCategory] = useState<TicketCategory>('bug')
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [reply, setReply] = useState('')
  const [deleteOpen, setDeleteOpen] = useState(false)

  const listQuery = useQuery({
    queryKey: ['self-tickets', page],
    queryFn: () => getSelfTickets({ p: page, page_size: 20 }),
  })

  const detailQuery = useQuery({
    queryKey: ['self-ticket', selectedId],
    queryFn: () => {
      if (selectedId == null) {
        return Promise.reject(new Error('no ticket selected'))
      }
      return getSelfTicket(selectedId)
    },
    enabled: selectedId != null,
  })

  const createMutation = useMutation({
    mutationFn: () =>
      createSelfTicket({ category, title, content }),
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Failed to create ticket'))
        return
      }
      toast.success(t('Ticket created'))
      setCreateOpen(false)
      setTitle('')
      setContent('')
      queryClient.invalidateQueries({ queryKey: ['self-tickets'] })
    },
    onError: () => toast.error(t('Failed to create ticket')),
  })

  const replyMutation = useMutation({
    mutationFn: () => {
      if (selectedId == null) {
        return Promise.reject(new Error('no ticket selected'))
      }
      return replySelfTicket(selectedId, reply)
    },
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Failed to send reply'))
        return
      }
      toast.success(t('Reply sent'))
      setReply('')
      queryClient.invalidateQueries({ queryKey: ['self-ticket', selectedId] })
      queryClient.invalidateQueries({ queryKey: ['self-tickets'] })
    },
    onError: () => toast.error(t('Failed to send reply')),
  })

  const deleteMutation = useMutation({
    mutationFn: () => {
      if (selectedId == null) {
        return Promise.reject(new Error('no ticket selected'))
      }
      return deleteSelfTicket(selectedId)
    },
    onSuccess: (res) => {
      if (!res.success) {
        toast.error(res.message || t('Failed to delete ticket'))
        return
      }
      toast.success(t('Ticket deleted'))
      setDeleteOpen(false)
      setSelectedId(null)
      queryClient.invalidateQueries({ queryKey: ['self-tickets'] })
    },
    onError: () => toast.error(t('Failed to delete ticket')),
  })

  const listData = listQuery.data?.data
  const tickets = listData?.items ?? []
  const total = listData?.total ?? 0
  const ticketDisabled = listData?.ticket_disabled === true
  const enabled = listData?.enabled !== false
  const selected = detailQuery.data?.data

  return (
    <>
      <SectionPageLayout fixedContent>
        <SectionPageLayout.Title>{t('Tickets')}</SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <Button
            type='button'
            onClick={() => setCreateOpen(true)}
            disabled={ticketDisabled || !enabled}
          >
            {t('New Ticket')}
          </Button>
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          {!enabled && (
            <p className='text-muted-foreground mb-4 text-sm'>
              {t('Ticket feature is currently disabled.')}
            </p>
          )}
          {ticketDisabled && (
            <p className='text-destructive mb-4 text-sm'>
              {t('Your ticket access has been disabled by an administrator.')}
            </p>
          )}

          <div className='grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)]'>
            <div className='space-y-2'>
              {tickets.length === 0 && !listQuery.isLoading && (
                <p className='text-muted-foreground text-sm'>
                  {t('No tickets yet.')}
                </p>
              )}
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
                  <span className='text-muted-foreground text-xs'>
                    {page}
                  </span>
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
                <TicketDetailPanel
                  ticket={selected}
                  reply={reply}
                  onReplyChange={setReply}
                  onSend={() => replyMutation.mutate()}
                  sending={replyMutation.isPending}
                  canReply={
                    selected.status === 'open' && !ticketDisabled && enabled
                  }
                  onDelete={() => setDeleteOpen(true)}
                />
              )}
            </div>
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('New Ticket')}</DialogTitle>
          </DialogHeader>
          <div className='space-y-3'>
            <div className='space-y-1.5'>
              <Label>{t('Category')}</Label>
              <Select
                value={category}
                onValueChange={(value) => {
                  if (value) setCategory(value as TicketCategory)
                }}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TICKET_CATEGORIES.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {t(item.labelKey)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className='space-y-1.5'>
              <Label>{t('Title')}</Label>
              <Input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                maxLength={200}
              />
            </div>
            <div className='space-y-1.5'>
              <Label>{t('Content')}</Label>
              <Textarea
                value={content}
                onChange={(e) => setContent(e.target.value)}
                rows={6}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => setCreateOpen(false)}
            >
              {t('Cancel')}
            </Button>
            <Button
              type='button'
              disabled={
                !title.trim() ||
                !content.trim() ||
                createMutation.isPending
              }
              onClick={() => createMutation.mutate()}
            >
              {t('Submit')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Delete Ticket')}</DialogTitle>
          </DialogHeader>
          <p className='text-muted-foreground text-sm'>
            {t(
              'This removes the ticket from your list. It still counts toward your daily ticket limit.'
            )}
          </p>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => setDeleteOpen(false)}
            >
              {t('Cancel')}
            </Button>
            <Button
              type='button'
              variant='destructive'
              disabled={deleteMutation.isPending}
              onClick={() => deleteMutation.mutate()}
            >
              {t('Delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

function TicketDetailPanel(props: {
  ticket: Ticket
  reply: string
  onReplyChange: (value: string) => void
  onSend: () => void
  sending: boolean
  canReply: boolean
  onDelete: () => void
}) {
  const { t } = useTranslation()
  const isClosed = props.ticket.status === 'closed'
  return (
    <div className='flex h-full flex-col gap-3'>
      <div>
        <div className='flex items-center justify-between gap-2'>
          <h3 className='font-semibold'>{props.ticket.title}</h3>
          <div className='flex shrink-0 items-center gap-2'>
            <Badge
              variant={
                props.ticket.status === 'open' ? 'default' : 'secondary'
              }
            >
              {props.ticket.status === 'open' ? t('Open') : t('Closed')}
            </Badge>
            <Button
              type='button'
              variant='destructive'
              size='sm'
              disabled={!isClosed}
              onClick={props.onDelete}
            >
              {t('Delete')}
            </Button>
          </div>
        </div>
        <p className='text-muted-foreground mt-1 text-xs'>
          {categoryLabel(t, props.ticket.category)} · #{props.ticket.id}
          {props.ticket.close_reason
            ? ` · ${closeReasonLabel(t, props.ticket.close_reason)}`
            : ''}
        </p>
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
              {message.is_admin ? t('Admin') : t('You')} ·{' '}
              {formatTime(message.created_at)}
            </div>
            <div className='whitespace-pre-wrap'>{message.content}</div>
          </div>
        ))}
      </div>
      {props.canReply && (
        <div className='space-y-2 border-t pt-3'>
          <Textarea
            value={props.reply}
            onChange={(e) => props.onReplyChange(e.target.value)}
            placeholder={t('Write a reply...')}
            rows={3}
          />
          <Button
            type='button'
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
