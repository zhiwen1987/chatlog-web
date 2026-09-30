<template>
  <div class="devices-page">
    <PageHeading eyebrow="DEVICES" title="设备" description="当前租户已登记的设备与最后心跳时间（presence）。"/>
    <section class="panel">
      <div class="source-section-heading"><div><h2>已登记设备</h2><p>{{ enterpriseMode ? 'Enterprise Server' : '当前模式不可用' }}</p></div><button class="btn" :disabled="busy || demoEnabled" @click="load"><UiIcon name="refresh" :size="16"/>刷新</button></div>
      <p v-if="demoEnabled" class="notice">当前是演示模式，设备在线状态需要 Enterprise 登录后查看。</p>
      <p v-if="error" class="error-text" role="alert">{{ error }}</p>
      <p v-else-if="busy" class="muted" role="status">正在读取设备…</p>
      <p v-else-if="!enterpriseMode" class="muted">设备在线状态仅在 Enterprise Server 模式可用。请前往「数据来源」切换到 Enterprise 并登录。</p>
      <div v-else-if="devices.length === 0" class="muted">还没有登记设备。设备端调用 <code>POST /api/v1/devices/register</code> 后会出现在这里。</div>
      <div v-else class="table-wrap">
        <table class="devices-table">
          <thead><tr><th>设备</th><th>平台</th><th>架构</th><th>客户端</th><th>状态</th><th>最后心跳</th></tr></thead>
          <tbody>
            <tr v-for="d in devices" :key="d.id">
              <td><strong>{{ d.device_name }}</strong><small class="mono">{{ shortId(d.id) }}</small></td>
              <td>{{ d.platform }}</td>
              <td>{{ d.architecture }}</td>
              <td>{{ d.client_version }}</td>
              <td><span class="pill" :class="statusClass(d.status)">{{ d.status }}</span></td>
              <td><span class="dot" :class="{offline:!online(d)}"/>{{ online(d) ? '在线' : '离线' }}<small class="muted">{{ fmtTime(d.last_seen_at) }}</small></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
<script>
import { ref, onMounted } from 'vue'
import PageHeading from '@/components/ui/PageHeading.vue'
import UiIcon from '@/components/ui/UiIcon.vue'
import enterprise from '@/api/enterprise'
import { demoEnabled } from '@/lib/demo'
const ONLINE_WINDOW_MS = 5 * 60 * 1000
export default {
  name: 'Devices',
  components: { PageHeading, UiIcon },
  setup () {
    const devices = ref([])
    const error = ref('')
    const busy = ref(false)
    const enterpriseMode = ref(enterprise.isEnterpriseMode())
    const shortId = id => (id || '').slice(0, 8)
    const online = d => {
      const t = d.last_seen_at ? Date.parse(d.last_seen_at) : NaN
      return Number.isFinite(t) && Date.now() - t < ONLINE_WINDOW_MS
    }
    const statusClass = s => s === 'active' ? 'pill-ok' : 'pill-warn'
    const fmtTime = t => t ? new Date(t).toLocaleString() : '从未'
    const load = async () => {
      error.value = ''
      busy.value = true
      try {
        const r = await enterprise.getDevices()
        devices.value = Array.isArray(r.data) ? r.data : []
        if (!Array.isArray(r.data)) error.value = '设备接口返回格式异常'
      } catch (e) {
        error.value = e.message
      } finally {
        busy.value = false
      }
    }
    onMounted(() => { if (enterpriseMode.value) load() })
    return { devices, error, busy, enterpriseMode, demoEnabled, shortId, online, statusClass, fmtTime, load }
  }
}
</script>
<style scoped>
.table-wrap { overflow-x: auto; }
.devices-table { width: 100%; border-collapse: collapse; }
.devices-table th, .devices-table td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border, #e5e7eb); vertical-align: middle; }
.devices-table th { font-size: 12px; text-transform: uppercase; letter-spacing: .04em; color: var(--muted, #6b7280); }
.devices-table td { font-size: 14px; }
.devices-table small { display: block; font-size: 12px; color: var(--muted, #6b7280); }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; background: #22c55e; margin-right: 6px; }
.dot.offline { background: #9ca3af; }
</style>