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
import { useQuery } from '@tanstack/react-query'
import { useEffect, useMemo } from 'react'
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
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { getGroups } from '@/features/users/api'

import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const schema = z.object({
  enabled: z.boolean(),
  patternsText: z.string(),
  exemptUserGroupsText: z.string(),
  exemptBillingGroupsText: z.string(),
})

type Values = z.infer<typeof schema>

function splitLines(value: string): string[] {
  return value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
}

function arraysEqual(a: string[], b: string[]) {
  return JSON.stringify(a) === JSON.stringify(b)
}

export function UABlacklistSection({
  defaultValues,
}: {
  defaultValues: {
    enabled: boolean
    patterns: string[]
    exemptUserGroups: string[]
    exemptBillingGroups: string[]
  }
}) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const formDefaults = useMemo(
    () => ({
      enabled: defaultValues.enabled,
      patternsText: defaultValues.patterns.join('\n'),
      exemptUserGroupsText: defaultValues.exemptUserGroups.join('\n'),
      exemptBillingGroupsText: defaultValues.exemptBillingGroups.join('\n'),
    }),
    [defaultValues]
  )

  const form = useForm<Values>({
    resolver: zodResolver(schema) as unknown as Resolver<Values>,
    defaultValues: formDefaults,
  })

  const { data: groupsData } = useQuery({
    queryKey: ['groups'],
    queryFn: getGroups,
    staleTime: 5 * 60 * 1000,
  })
  const groups = groupsData?.data ?? []

  useEffect(() => {
    form.reset(formDefaults)
  }, [formDefaults, form])

  const { isDirty, isSubmitting } = form.formState

  async function onSubmit(values: Values) {
    const patterns = splitLines(values.patternsText)
    const exemptUserGroups = splitLines(values.exemptUserGroupsText)
    const exemptBillingGroups = splitLines(values.exemptBillingGroupsText)
    const updates: Array<{ key: string; value: string }> = []

    if (values.enabled !== defaultValues.enabled) {
      updates.push({
        key: 'ua_blacklist_setting.enabled',
        value: String(values.enabled),
      })
    }
    if (!arraysEqual(patterns, defaultValues.patterns)) {
      updates.push({
        key: 'ua_blacklist_setting.patterns',
        value: JSON.stringify(patterns),
      })
    }
    if (!arraysEqual(exemptUserGroups, defaultValues.exemptUserGroups)) {
      updates.push({
        key: 'ua_blacklist_setting.exempt_user_groups',
        value: JSON.stringify(exemptUserGroups),
      })
    }
    if (
      !arraysEqual(exemptBillingGroups, defaultValues.exemptBillingGroups)
    ) {
      updates.push({
        key: 'ua_blacklist_setting.exempt_billing_groups',
        value: JSON.stringify(exemptBillingGroups),
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
    <SettingsSection title={t('UA Blacklist')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending || isSubmitting}
            isSaveDisabled={!isDirty}
            saveLabel='Save UA blacklist settings'
          />
          <FormField
            control={form.control}
            name='enabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable UA blacklist')}</FormLabel>
                  <FormDescription>
                    {t(
                      'Matching requests disable the account and record a ban reason.'
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
            name='patternsText'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('UA patterns (one per line)')}</FormLabel>
                <FormControl>
                  <Textarea rows={5} {...field} />
                </FormControl>
                <FormDescription>
                  {t(
                    'Case-insensitive substring match. Examples: claude-cli, codex-cli'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='exemptUserGroupsText'
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  {t('Exempt user groups (one per line)')}
                </FormLabel>
                <FormControl>
                  <Textarea rows={4} {...field} />
                </FormControl>
                <FormDescription>
                  {t(
                    'These user groups never trigger UA bans, for any billing group.'
                  )}
                  {groups.length > 0
                    ? ` ${t('Available groups')}: ${groups.join(', ')}`
                    : ''}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name='exemptBillingGroupsText'
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  {t('Exempt billing groups (one per line)')}
                </FormLabel>
                <FormControl>
                  <Textarea rows={4} {...field} />
                </FormControl>
                <FormDescription>
                  {t(
                    'Non-exempt users calling through these billing groups will not be banned.'
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
