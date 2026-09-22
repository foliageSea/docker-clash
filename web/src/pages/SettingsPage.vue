<script setup lang="ts">
import { SaveIcon, ShieldAlertIcon } from '@lucide/vue'

import PageHeader from '@/components/PageHeader.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldContent, FieldDescription, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { NumberField, NumberFieldContent, NumberFieldDecrement, NumberFieldIncrement, NumberFieldInput } from '@/components/ui/number-field'
import { Switch } from '@/components/ui/switch'
import { useConsole } from '@/composables/use-console'
import { api } from '@/lib/api'

const { settings, busy, run } = useConsole()
</script>

<template>
  <PageHeader title="网络设置" description="配置管理服务监听与 Mixed 代理入口">
    <Button :disabled="busy" @click="run(() => api.saveSettings(settings), '网络设置已保存')"><SaveIcon data-icon="inline-start" />保存设置</Button>
  </PageHeader>

  <Card class="max-w-3xl">
    <CardHeader><CardTitle>入站网络</CardTitle><CardDescription>代理配置保存后会自动应用到 mihomo 内核。</CardDescription></CardHeader>
    <CardContent>
      <FieldGroup>
        <Field><FieldLabel for="listen">管理界面监听地址</FieldLabel><Input id="listen" v-model="settings.listen" /><FieldDescription>修改此项后需要重启 Docker Clash 服务。</FieldDescription></Field>
        <div class="grid gap-6 sm:grid-cols-2">
          <Field>
            <FieldLabel for="mixed-port">Mixed 代理端口</FieldLabel>
            <NumberField id="mixed-port" v-model="settings.mixedPort" :min="1" :max="65535" :format-options="{ useGrouping: false }">
              <NumberFieldContent><NumberFieldDecrement /><NumberFieldInput /><NumberFieldIncrement /></NumberFieldContent>
            </NumberField>
            <FieldDescription>同时接受 HTTP 与 SOCKS5。</FieldDescription>
          </Field>
          <Field><FieldLabel for="bind-address">内核绑定地址</FieldLabel><Input id="bind-address" v-model="settings.bindAddress" :disabled="!settings.allowLan" /><FieldDescription>局域网开放时通常使用 *。</FieldDescription></Field>
        </div>
        <div class="grid gap-6 sm:grid-cols-2">
          <Field><FieldLabel for="external-address">外部访问地址</FieldLabel><Input id="external-address" v-model="settings.externalAddress" placeholder="203.0.113.10 或 proxy.example.com" /><FieldDescription>不包含协议与端口。</FieldDescription></Field>
          <Field>
            <FieldLabel for="external-port">外部代理端口</FieldLabel>
            <NumberField id="external-port" v-model="settings.externalPort" :min="1" :max="65535" :format-options="{ useGrouping: false }">
              <NumberFieldContent><NumberFieldDecrement /><NumberFieldInput /><NumberFieldIncrement /></NumberFieldContent>
            </NumberField>
            <FieldDescription>须与 Docker 宿主机映射端口一致。</FieldDescription>
          </Field>
        </div>
        <Field orientation="horizontal" class="rounded-md border p-4">
          <FieldContent><FieldLabel for="allow-lan">允许局域网连接</FieldLabel><FieldDescription>允许可信局域网中的设备使用此代理入口。</FieldDescription></FieldContent>
          <Switch id="allow-lan" v-model="settings.allowLan" />
        </Field>
        <Alert v-if="settings.allowLan" variant="destructive"><ShieldAlertIcon /><AlertTitle>代理端口对局域网开放</AlertTitle><AlertDescription>当前未配置身份认证，请通过防火墙限制访问范围并仅在可信网络中使用。</AlertDescription></Alert>
      </FieldGroup>
    </CardContent>
  </Card>
</template>
