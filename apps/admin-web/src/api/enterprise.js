// Enterprise API Adapter（Phase 3）
// 连接 Chatlog Enterprise Server（PostgreSQL），保留 legacy-chatlog-http 兼容。
// 切换：localStorage['chatlog-enterprise-mode']==='enterprise'
import { parseChatLogs, normalizeContact, normalizeRoom, normalizeSession, safeUrl, csvString } from '@/lib/data'

const MODE_KEY = 'chatlog-enterprise-mode'
const TOKEN_KEY = 'chatlog-enterprise-token'
const API_KEY = 'chatlog-enterprise-api-base'

export function isEnterpriseMode () {
  try { return localStorage.getItem(MODE_KEY) === 'enterprise' } catch (_) { return false }
}

export function setEnterpriseMode (on) {
  try { localStorage.setItem(MODE_KEY, on ? 'enterprise' : '') } catch (_) { /* storage 不可用时忽略 */ }
}

export function getEnterpriseBase () {
  let saved = null
  try { saved = localStorage.getItem(API_KEY) } catch (_) { /* storage 不可用时忽略 */ }
  return saved || 'http://127.0.0.1:8080'
}

export function setEnterpriseBase (value) {
  const base = String(value || '').trim().replace(/\/+$/, '')
  if (base && !safeUrl(base)) throw new Error('请输入有效的 HTTP 或 HTTPS 服务地址。')
  localStorage.setItem(API_KEY, base)
}

export function getEnterpriseToken () {
  try { return localStorage.getItem(TOKEN_KEY) || '' } catch (_) { return '' }
}

export function setEnterpriseToken (token) {
  localStorage.setItem(TOKEN_KEY, token)
}

async function request (path, params = {}, method = 'GET') {
  const query = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) if (v !== '' && v != null) query.set(k, String(v))
  const token = getEnterpriseToken()
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), 12000)
  try {
    const response = await fetch(`${getEnterpriseBase()}${path}${query.size ? '?' + query : ''}`, {
      method,
      signal: controller.signal,
      credentials: 'omit',
      cache: 'no-store',
      headers: {
        Accept: 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      }
    })
    if (response.status === 401 || response.status === 403) throw new Error('Enterprise Server 拒绝访问，请检查登录状态与权限。')
    if (!response.ok) throw new Error(`Enterprise Server 返回 HTTP ${response.status}`)
    const raw = await response.text()
    if (response.headers.get('content-type')?.includes('text/html') || /^\s*<!doctype html/i.test(raw)) throw new Error('收到网页而不是数据，请检查服务地址。')
    let data = raw
    try { data = JSON.parse(raw) } catch (_) { /* Text 是合法 legacy 格式 */ }
    return { data, headers: Object.fromEntries(response.headers.entries()), status: response.status }
  } catch (error) {
    if (error.name === 'AbortError') throw new Error('连接超时，请确认 Enterprise Server 已启动。')
    throw error
  } finally { clearTimeout(timeout) }
}

// 解包 server 的 {data:[...], total} 响应
function unwrap (result, kind, normalize) {
  const payload = result.data && typeof result.data === 'object' && Array.isArray(result.data.data) ? result.data : result.data
  const rows = Array.isArray(payload) ? payload : (payload?.data || [])
  return {
    ...result,
    data: rows.map(normalize),
    total: payload?.total ?? rows.length
  }
}

export default {
  isEnterpriseMode,
  setEnterpriseMode,
  getEnterpriseBase,
  setEnterpriseBase,
  getEnterpriseToken,
  setEnterpriseToken,
  getContacts: () => request('/api/v1/contacts').then(r => unwrap(r, 'contacts', normalizeContact)),
  getChatrooms: () => request('/api/v1/conversations').then(r => unwrap(r, 'chatrooms', normalizeRoom)),
  getSessions: () => request('/api/v1/conversations').then(r => unwrap(r, 'sessions', normalizeSession)),
  async getChatLogs (params = {}) {
    const result = await request('/api/v1/messages', params)
    const unwrapped = unwrap(result, 'chatlog', normalizeMessageRow)
    return { ...unwrapped, data: parseChatLogs(unwrapped.data) }
  },
  getChatLogsRaw: (params = {}) => request('/api/v1/messages', params),
  exportChatLogs: async (params = {}) => {
    const result = await request('/api/v1/messages', { ...params, page_size: 1000 })
    const rows = Array.isArray(result.data) ? result.data : (result.data?.data || [])
    return { ...result, data: csvString(rows.map(normalizeMessageRow), [
      { key: 'id', label: 'ID' }, { key: 'conversation_id', label: '会话 ID' },
      { key: 'direction', label: '方向' }, { key: 'message_type', label: '类型' },
      { key: 'content_text', label: '内容' }, { key: 'sent_at', label: '时间' }
    ]) }
  },
  getImageUrl: id => `${getEnterpriseBase()}/media/${encodeURIComponent(id)}`,
  getVideoUrl: id => `${getEnterpriseBase()}/media/${encodeURIComponent(id)}`,
  getVoiceUrl: id => `${getEnterpriseBase()}/media/${encodeURIComponent(id)}`,
  getFileUrl: id => `${getEnterpriseBase()}/media/${encodeURIComponent(id)}`,
  getDataUrl: path => `${getEnterpriseBase()}/data/${String(path).split('/').map(encodeURIComponent).join('/')}`,
  getLocalAnalysis: () => Promise.reject(new Error('Enterprise 模式不支持本地分析')),
  getLocalMedia: () => Promise.reject(new Error('Enterprise 模式不支持本地媒体'))
}

function normalizeMessageRow (item) {
  return {
    ...item,
    id: item.id,
    content: item.content_text,
    sender: item.direction,
    time: item.sent_at
  }
}