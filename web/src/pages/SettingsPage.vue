<script setup lang="ts">
import { SaveIcon, ShieldAlertIcon } from '@lucide/vue'

import PageHeader from '@/components/PageHeader.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldLegend, FieldSet } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { NumberField, NumberFieldContent, NumberFieldDecrement, NumberFieldIncrement, NumberFieldInput } from '@/components/ui/number-field'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'

const { settings, busy, run } = useConsole()
</script>

<template>
  <PageHeader title="网络设置">
    <Button size="sm" :disabled="busy" @click="run(() => api.saveSettings(settings), '网络设置已保存')">
      <SaveIcon data-icon="inline-start" />保存设置
    </Button>
  </PageHeader>

  <div class="grid max-w-5xl items-start gap-4 @4xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
    <Card class="min-w-0 gap-4 py-4">
      <CardHeader class="px-4 sm:px-5">
        <CardTitle>代理入口</CardTitle>
        <CardDescription>保存后自动应用到 mihomo。</CardDescription>
      </CardHeader>
      <CardContent class="px-4 sm:px-5">
        <FieldGroup class="gap-4">
          <Field class="max-w-52 gap-2">
            <FieldLabel for="mixed-port">Mixed 代理端口</FieldLabel>
            <NumberField id="mixed-port" v-model="settings.mixedPort" :min="1" :max="65535" :format-options="{ useGrouping: false }">
              <NumberFieldContent>
                <NumberFieldDecrement />
                <NumberFieldInput aria-describedby="mixed-port-help" />
                <NumberFieldIncrement />
              </NumberFieldContent>
            </NumberField>
            <FieldDescription id="mixed-port-help">同时接受 HTTP 与 SOCKS5。</FieldDescription>
          </Field>

          <Separator />

          <FieldSet class="min-w-0">
            <FieldLegend>外部访问</FieldLegend>
            <FieldGroup class="gap-4 @md/field-group:grid @md/field-group:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
              <Field class="min-w-0 gap-2">
                <FieldLabel for="external-address">地址</FieldLabel>
                <Input id="external-address" v-model="settings.externalAddress" placeholder="203.0.113.10 或 proxy.example.com" aria-describedby="external-address-help" />
                <FieldDescription id="external-address-help">不包含协议与端口。</FieldDescription>
              </Field>
              <Field class="min-w-0 gap-2">
                <FieldLabel for="external-port">代理端口</FieldLabel>
                <NumberField id="external-port" v-model="settings.externalPort" :min="1" :max="65535" :format-options="{ useGrouping: false }">
                  <NumberFieldContent>
                    <NumberFieldDecrement />
                    <NumberFieldInput aria-describedby="external-port-help" />
                    <NumberFieldIncrement />
                  </NumberFieldContent>
                </NumberField>
                <FieldDescription id="external-port-help">须与 Docker 宿主机映射端口一致。</FieldDescription>
              </Field>
            </FieldGroup>
          </FieldSet>
        </FieldGroup>
      </CardContent>
    </Card>

    <div class="flex min-w-0 flex-col gap-4">
      <Card class="gap-4 py-4">
        <CardHeader class="px-4 sm:px-5"><CardTitle>局域网</CardTitle></CardHeader>
        <CardContent class="px-4 sm:px-5">
          <FieldGroup class="gap-4">
            <Field orientation="horizontal">
              <FieldLabel for="allow-lan">允许局域网连接</FieldLabel>
              <Switch id="allow-lan" v-model="settings.allowLan" />
            </Field>
            <Alert v-if="settings.allowLan" variant="destructive">
              <ShieldAlertIcon />
              <AlertTitle>代理端口对局域网开放</AlertTitle>
              <AlertDescription>当前未配置身份认证，请通过防火墙限制访问范围并仅在可信网络中使用。</AlertDescription>
            </Alert>
            <Field class="gap-2" :data-disabled="!settings.allowLan || undefined">
              <FieldLabel for="bind-address">内核绑定地址</FieldLabel>
              <Input id="bind-address" v-model="settings.bindAddress" :disabled="!settings.allowLan" aria-describedby="bind-address-help" />
              <FieldDescription id="bind-address-help">局域网开放时通常使用 *。</FieldDescription>
            </Field>
          </FieldGroup>
        </CardContent>
      </Card>

      <Card class="gap-4 py-4">
        <CardHeader class="px-4 sm:px-5"><CardTitle>管理界面</CardTitle></CardHeader>
        <CardContent class="px-4 sm:px-5">
          <FieldGroup>
            <Field class="gap-2">
              <FieldLabel for="listen">监听地址</FieldLabel>
              <Input id="listen" v-model="settings.listen" aria-describedby="listen-help" />
              <FieldDescription id="listen-help">修改后需重启 Docker Clash 服务。</FieldDescription>
            </Field>
          </FieldGroup>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
