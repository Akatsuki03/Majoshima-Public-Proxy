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
  penaltyEnabled: z.boolean(),
  penaltyAmount: z.coerce.number().min(0),
  penaltyMinInputTokens: z.coerce.number().int().min(0),
  maxInputTokens: z.coerce.number().int().min(0),
})

type Values = z.infer<typeof schema>

/**
 * Request guard settings: test-message penalty (input token threshold based)
 * and the global input token cap. Admin and root users are exempt from both.
 */
export function RequestGuardSection({
  defaultValues,
}: {
  defaultValues: {
    penaltyEnabled: boolean
    penaltyAmount: number
    penaltyMinInputTokens: number
    maxInputTokens: number
  }
}) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const form = useForm<Values>({
    resolver: zodResolver(schema) as unknown as Resolver<Values>,
    defaultValues: {
      penaltyEnabled: defaultValues.penaltyEnabled,
      penaltyAmount: defaultValues.penaltyAmount,
      penaltyMinInputTokens: defaultValues.penaltyMinInputTokens,
      maxInputTokens: defaultValues.maxInputTokens,
    },
  })

  const { isDirty, isSubmitting } = form.formState
  const penaltyEnabled = form.watch('penaltyEnabled')

  async function onSubmit(values: Values) {
    const updates: Array<{ key: string; value: string }> = []

    if (values.penaltyEnabled !== defaultValues.penaltyEnabled) {
      updates.push({
        key: 'test_penalty_setting.enabled',
        value: String(values.penaltyEnabled),
      })
    }
    if (values.penaltyAmount !== defaultValues.penaltyAmount) {
      updates.push({
        key: 'test_penalty_setting.amount',
        value: String(values.penaltyAmount),
      })
    }
    if (values.penaltyMinInputTokens !== defaultValues.penaltyMinInputTokens) {
      updates.push({
        key: 'test_penalty_setting.min_input_tokens',
        value: String(values.penaltyMinInputTokens),
      })
    }
    if (values.maxInputTokens !== defaultValues.maxInputTokens) {
      updates.push({
        key: 'input_limit_setting.max_input_tokens',
        value: String(values.maxInputTokens),
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
    <SettingsSection title={t('Request Guard')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending || isSubmitting}
            isSaveDisabled={!isDirty}
            saveLabel='Save request guard settings'
          />

          <FormField
            control={form.control}
            name='penaltyEnabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable test message penalty')}</FormLabel>
                  <FormDescription>
                    {t(
                      'Charge a penalty when the input token count is below the threshold. Admins are exempt'
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

          {penaltyEnabled && (
            <div className='grid gap-6 sm:grid-cols-2'>
              <FormField
                control={form.control}
                name='penaltyMinInputTokens'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Minimum input tokens')}</FormLabel>
                    <FormControl>
                      <Input type='number' min={0} {...field} />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'Requests with fewer input tokens are treated as test messages. 0 disables the check'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='penaltyAmount'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Penalty amount ($)')}</FormLabel>
                    <FormControl>
                      <Input type='number' min={0} step='0.01' {...field} />
                    </FormControl>
                    <FormDescription>
                      {t('Amount deducted when a test message is detected')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          )}

          <FormField
            control={form.control}
            name='maxInputTokens'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Global max input tokens')}</FormLabel>
                <FormControl>
                  <Input type='number' min={0} {...field} />
                </FormControl>
                <FormDescription>
                  {t(
                    'Reject requests whose input token count exceeds this cap. 0 disables the limit. Admins are exempt'
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
