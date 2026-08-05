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
import { Input } from '@/components/ui/input'
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
  allowedGroupsText: z.string(),
  targetGroup: z.string(),
  promoteOnResolve: z.boolean(),
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

export function ToolCallGuardSection({
  defaultValues,
}: {
  defaultValues: {
    enabled: boolean
    allowedGroups: string[]
    targetGroup: string
    promoteOnResolve: boolean
  }
}) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const formDefaults = useMemo(
    () => ({
      enabled: defaultValues.enabled,
      allowedGroupsText: defaultValues.allowedGroups.join('\n'),
      targetGroup: defaultValues.targetGroup,
      promoteOnResolve: defaultValues.promoteOnResolve,
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
    const allowedGroups = splitLines(values.allowedGroupsText)
    const targetGroup = values.targetGroup.trim()
    const updates: Array<{ key: string; value: string }> = []

    if (values.enabled !== defaultValues.enabled) {
      updates.push({
        key: 'tool_call_guard_setting.enabled',
        value: String(values.enabled),
      })
    }
    if (!arraysEqual(allowedGroups, defaultValues.allowedGroups)) {
      updates.push({
        key: 'tool_call_guard_setting.allowed_groups',
        value: JSON.stringify(allowedGroups),
      })
    }
    if (targetGroup !== defaultValues.targetGroup) {
      updates.push({
        key: 'tool_call_guard_setting.target_group',
        value: targetGroup,
      })
    }
    if (values.promoteOnResolve !== defaultValues.promoteOnResolve) {
      updates.push({
        key: 'tool_call_guard_setting.promote_on_resolve',
        value: String(values.promoteOnResolve),
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
    <SettingsSection title={t('Tool Call Guard')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)} autoComplete='off'>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending || isSubmitting}
            isSaveDisabled={!isDirty}
            saveLabel='Save tool call guard settings'
          />
          <FormField
            control={form.control}
            name='enabled'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>{t('Enable tool call guard')}</FormLabel>
                  <FormDescription>
                    {t(
                      'Requests containing tool calls are rejected unless the user group or billing group is in the allowed list. The admin group is always allowed.'
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
            name='allowedGroupsText'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Allowed groups (one per line)')}</FormLabel>
                <FormControl>
                  <Textarea rows={4} {...field} />
                </FormControl>
                <FormDescription>
                  {t(
                    'Tool calls pass when the user group or billing group matches any entry.'
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
            name='promoteOnResolve'
            render={({ field }) => (
              <SettingsSwitchItem>
                <SettingsSwitchContent>
                  <FormLabel>
                    {t('Move user to target group on ticket resolve')}
                  </FormLabel>
                  <FormDescription>
                    {t(
                      'When an admin resolves a tool call ticket, the user is moved into the target group automatically.'
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
            name='targetGroup'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Target group')}</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormDescription>
                  {t(
                    'The group must exist in group ratio settings, otherwise promotion is skipped.'
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
