<template>
  <section class="contracts-view">
    <h1>契约与完整性</h1>
    <p class="muted">前端消费工具（W05-W08 + W11 统一入口）只读展示。逻辑见 <code>src/lib/contracts.js</code>，不替代服务端签名/时间/身份语义校验。</p>

    <h2>完整性报告</h2>
    <div class="card-grid">
      <div class="card" :class="trustable ? 'ok' : 'warn'">
        <strong>{{ trustable ? 'counts 一致' : 'counts 不一致（不可信）' }}</strong>
        <span v-if="trustable">完整性：{{ completenessText }}（{{ verified }}/{{ inScope }}）</span>
        <span v-else class="error-text">{{ countErrorText }}</span>
        <span class="muted">outOfScope={{ outOfScope }} · total={{ total }}</span>
      </div>
    </div>

    <h2>许可 claims v2</h2>
    <div class="card-grid">
      <div v-for="g in grants" :key="g.grantId" class="card" :class="g.usable ? 'ok' : 'warn'">
        <strong>{{ g.featureKey }}</strong>
        <span>{{ g.state }} · {{ g.active ? '有效' : '无效' }} · {{ g.usable ? '可用' : '不可用' }}</span>
        <span v-if="g.missingDeps.length" class="error-text">缺依赖：{{ g.missingDeps.join(', ') }}</span>
      </div>
    </div>

    <h2>媒体收据/清单对账</h2>
    <div class="card-grid">
      <div class="card" :class="reconcileOk ? 'ok' : 'warn'">
        <strong>{{ reconcileOk ? '收据与清单一致' : '对账失败（不发假 ACK）' }}</strong>
        <span v-if="reconcileErrors.length" class="error-text">{{ reconcileErrors.join('；') }}</span>
      </div>
    </div>
  </section>
</template>
<script>
import { ref, computed } from 'vue'
import { renderReport, renderClaims, reconcile } from '@/lib/contracts'
export default {
  name: 'ContractsView',
  setup() {
    // 合成展示数据：真实运行时由服务端下发的 integrity-report/license-claims/manifest 填充
    const report = ref({
      counts: { total_discovered: 100, in_scope: 80, verified: 50, pending: 20, excluded: 8, source_missing: 2 },
    })
    const claims = ref({
      mode: 'v2',
      feature_grants: [
        { grant_id: 'g1', feature_key: 'archive.read', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' },
        { grant_id: 'g2', feature_key: 'archive.ingest', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' },
        { grant_id: 'g3', feature_key: 'media.upload.image', state: 'active', valid_from: '2026-01-01T00:00:00Z', valid_until: '2027-01-01T00:00:00Z' },
      ],
    })
    const catalog = ref({
      'archive.read': { dependsOn: [] },
      'archive.ingest': { dependsOn: [] },
      'media.upload.image': { dependsOn: ['archive.ingest'] },
    })
    const receipt = ref({
      receipt_id: 'r1', catalog_version: 1, tenant_id: 't1', deployment_id: 'd1',
      source: 'wechat-desktop', source_message_ref: 'src.1', seq: '1',
      committed_at: '2026-01-01T00:00:00Z', backup_set: 'bs1', object_ref: 'obj.1', media_kind: 'image',
    })
    const manifest = ref({
      manifest_id: 'mm-1', catalog_version: 1, tenant_id: 't1', deployment_id: 'd1',
      object_ref: 'obj.1', media_type: 'image',
      origin: { kind: 'original', source: 'wechat-desktop', source_message_ref: 'src.1' },
      bytes_ref: 's3://b/obj.1', sha256: 'a'.repeat(64), size_bytes: 10, seq: '1',
      created_at: '2026-01-01T00:00:00Z', state: 'stored',
    })

    const trustable = ref(false)
    const countErrorText = ref('')
    const verified = ref(0), inScope = ref(0), outOfScope = ref(0), total = ref(0), completenessText = ref('')
    try {
      const r = renderReport(report.value)
      trustable.value = true
      verified.value = r.verified; inScope.value = Math.round(r.verified / r.completeness); outOfScope.value = r.outOfScope; total.value = r.total
      completenessText.value = (r.completeness * 100).toFixed(1) + '%'
    } catch (e) { countErrorText.value = e.message }

    const claimsView = computed(() => renderClaims(claims.value, catalog.value))
    const grants = computed(() => claimsView.value.grants || [])

    const rec = reconcile(receipt.value, manifest.value)
    const reconcileOk = rec.ok
    const reconcileErrors = rec.errors

    return { trustable, countErrorText, verified, inScope, outOfScope, total, completenessText, grants, reconcileOk, reconcileErrors }
  },
}
</script>
<style scoped>
.contracts-view { padding: 1rem; max-width: 900px; margin: 0 auto; }
.card-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 0.5rem; }
.card { border: 1px solid #ddd; border-radius: 6px; padding: 0.5rem; display: flex; flex-direction: column; gap: 0.25rem; }
.card.ok { border-color: #2e7d32; }
.card.warn { border-color: #b26a00; }
.error-text { color: #c62828; }
.muted { color: #666; font-size: 0.85rem; }
code { background: #f0f0f0; padding: 0 0.25rem; border-radius: 3px; }
</style>