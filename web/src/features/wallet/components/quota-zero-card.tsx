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
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { api } from '@/lib/api'

type QuotaZeroStatus = {
  enabled: boolean
  cooldown_days: number
  last_quota_zero_time: number
  cooldown_remaining_days: number
  can_zero: boolean
  quota: number
}

async function getQuotaZeroStatus(): Promise<{
  success: boolean
  message?: string
  data?: QuotaZeroStatus
}> {
  const res = await api.get('/api/user/quota_zero')
  return res.data
}

async function zeroQuota(): Promise<{ success: boolean; message?: string }> {
  const res = await api.post('/api/user/quota_zero')
  return res.data
}

export function QuotaZeroCard(props: { onZeroed?: () => void }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const statusQuery = useQuery({
    queryKey: ['quota-zero-status'],
    queryFn: getQuotaZeroStatus,
  })

  const mutation = useMutation({
    mutationFn: zeroQuota,
    onSuccess: async (res) => {
      if (!res.success) {
        toast.error(res.message || t('Failed to zero quota'))
        return
      }
      toast.success(t('Quota reset to zero'))
      await queryClient.invalidateQueries({ queryKey: ['quota-zero-status'] })
      props.onZeroed?.()
    },
    onError: () => toast.error(t('Failed to zero quota')),
  })

  const data = statusQuery.data?.data
  if (!data?.enabled || data.quota >= 0) {
    return null
  }

  return (
    <div className='rounded-lg border border-destructive/30 bg-destructive/5 p-4'>
      <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
        <div>
          <h3 className='text-sm font-medium'>{t('Negative Balance')}</h3>
          <p className='text-muted-foreground mt-1 text-sm'>
            {data.can_zero
              ? t(
                  'Your balance is negative. You can reset it to zero once per cooldown period.'
                )
              : t('Quota zero is on cooldown. {{days}} day(s) remaining.', {
                  days: data.cooldown_remaining_days,
                })}
          </p>
        </div>
        <Button
          variant='destructive'
          disabled={!data.can_zero || mutation.isPending}
          onClick={() => mutation.mutate()}
        >
          {t('Zero Quota')}
        </Button>
      </div>
    </div>
  )
}
