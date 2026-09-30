import http from './http'
import enterprise from './enterprise'
import { isLocal } from '@/data-sources/source'
import { localRequest } from '@/data-sources/local-client'
import { csvString } from '@/lib/data'
export * from './http'
export * from './enterprise'

const fields=[{key:'senderName',label:'发送者'},{key:'time',label:'时间'},{key:'content',label:'内容'},{key:'talkerId',label:'会话 ID'},{key:'id',label:'消息 ID'},{key:'decodeStatus',label:'解析状态'}]
async function localRaw(params={}) {
  const result=await localRequest('chatlog',params)
  if (params.format==='csv') return {...result,data:csvString(result.data,fields)}
  if (params.format==='text') return {...result,data:result.data.map(row=>`${row.senderName}(${row.senderId}) ${row.time || '时间未知'}\n${row.content}`).join('\n\n')}
  return result
}
// 选择数据通道：Enterprise > 本地 > legacy HTTP
function pick (localFn, httpFn, entFn) {
  if (enterprise.isEnterpriseMode()) return entFn()
  if (isLocal.value) return localFn()
  return httpFn()
}
export default {
  ...http,
  getContacts:()=>pick(()=>localRequest('contacts'),()=>http.getContacts(),()=>enterprise.getContacts()),
  getChatrooms:()=>pick(()=>localRequest('chatrooms'),()=>http.getChatrooms(),()=>enterprise.getChatrooms()),
  getSessions:()=>pick(()=>localRequest('sessions'),()=>http.getSessions(),()=>enterprise.getSessions()),
  getChatLogs:(params={})=>pick(()=>localRequest('chatlog',params),()=>http.getChatLogs(params),()=>enterprise.getChatLogs(params)),
  getChatLogsRaw:(params={})=>pick(()=>localRaw(params),()=>http.getChatLogsRaw(params),()=>enterprise.getChatLogsRaw(params)),
  exportChatLogs:(params={})=>pick(()=>localRaw({...params,format:'csv'}),()=>http.exportChatLogs(params),()=>enterprise.exportChatLogs(params)),
  getLocalAnalysis:(params={})=>localRequest('statistics',params),
  getLocalMedia:(params={})=>localRequest('media',params)
}
