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
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm, type Resolver } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { z } from 'zod'

import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const schema = z.object({
  ticketEnabled: z.boolean(),
  ticketDailyLimit: z.coerce.number().int().min(1),
  quotaZeroEnabled: z.boolean(),
  quotaZeroCooldownDays: z.coerce.number().int().min(0),
})

type Values = z.infer<typeof schema>

export function TicketQuotaSection({
  defaultValues,
}: {
  defaultValues: Values
}) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const form = useForm<Values>({
    resolver: zodResolver(schema) as unknown as Resolver<Values>,
    defaultValues,
  })

  useEffect(() => {
    form.reset(defaultValues)
  }, [defaultValues, form])

  const { isDirty, isSubmitting } = form.formState

  async function onSubmit(values: Values) {
    const updates: Array<{ key: string; value: string }> = []
    if (values.ticketEnabled !== defaultValues.ticketEnabled) {
      updates.push({
        key: 'ticket_setting.enabled',
        value: String(values.ticketEnabled),
      })
    }
    if (values.ticketDailyLimit !== defaultValues.ticketDailyLimit) {
      updates.push({
        key: 'ticket_setting.daily_limit',
        value: String(values.ticketDailyLimit),
      })
    }
    if (values.quotaZeroEnabled !== defaultValues.quotaZeroEnabled) {
      updates.push({
        key: 'quota_zero_setting.enabled',
        value: String(values.quotaZeroEnabled),
      })
    }
    if (
      values.quotaZeroCooldownDays !== defaultValues.quotaZeroCooldownDays
    ) {
      updates.push({
        key: 'quota_zero_setting.cooldown_days',
        value: String(values.quotaZeroCooldownDays),
      })
    }
    if (updates.length === 0) {
      toast.info(t('No changes to save'))
      return
    }
    for (const update of updates) {
      await updateOption.mutateAsync(update)
    }
    form.reset(values)
  }

  return (
    <SettingsSection title={t('Tickets & Quota Zero')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending || isSubmitting}
            isSaveDisabled={!isDirty}
            saveLabel='Save ticket and quota zero settings'
          />
          <FormField
            control={form.control}
            name='ticketEnabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable tickets')}</FormLabel>
                  <FormDescription>
                    {t('Allow users to submit support tickets.')}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                    disabled={updateOption.isPending || isSubmitting}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />
          <FormField
            control={form.control}
            name='ticketDailyLimit'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Daily new ticket limit')}</FormLabel>
                <FormControl>
                  <Input type='number' min={1} {...field} />
                </FormControl>
                <FormDescription>
                  {t(
                    'Maximum number of new tickets a user can open per day. Replies are unlimited.'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='quotaZeroEnabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable negative quota zero')}</FormLabel>
                  <FormDescription>
                    {t(
                      'Allow users with negative balance to reset quota to zero.'
                    )}
                  </FormDescription>
                </SettingsSwitchContent>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                    disabled={updateOption.isPending || isSubmitting}
                  />
                </FormControl>
              </SettingsSwitchItem>
            )}
          />
          <FormField
            control={form.control}
            name='quotaZeroCooldownDays'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Quota zero cooldown (days)')}</FormLabel>
                <FormControl>
                  <Input type='number' min={0} {...field} />
                </FormControl>
                <FormDescription>
                  {t(
                    'Days a user must wait before zeroing negative quota again.'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
