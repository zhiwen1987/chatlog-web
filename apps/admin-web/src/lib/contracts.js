// 契约前端统一入口（W11）。
// 聚合 W05-W08 的四个工具（integrity/license/media/ownership）为一个
// 命名空间 + 一个 validateContract 分发函数。供 UI 视图挂载时单点引用。
// 不复制逻辑：本文件只 re-export + 分发，逻辑仍各自封装在对应模块。
// 纯函数、无副作用。

// 仅本地 import 分发真正用到的函数；其余由下方 export from 直接 re-export。
import { manifestErrors, receiptErrors } from './media.js';
import { countErrors } from './integrity.js';
import { sourceErrors } from './source.js';

export {
  // integrity-report（R42.10）
  countErrors, statusLabel, renderReport,
} from './integrity.js';

export {
  // license-claims v2（R42.8）
  grantActive, dependencyCheck, renderClaims,
} from './license.js';

export {
  // media manifest/receipt（R42.7）
  manifestErrors, receiptErrors, reconcile,
} from './media.js';

export {
  // data-ownership / 十进制字符串（R07/R39）
  decimalStringError, checkDecimalFields, timeUnknownError, ownerReport,
} from './ownership.js';

export {
  // source-adapter（R05/R42.7）
  sourceErrors, capabilitiesReport, syncPlan,
} from './source.js';

// 按类型分发到对应校验函数；返回 { ok, errors }。
// 未知类型拒绝，不静默通过。
export function validateContract(type, obj) {
  let errs;
  switch (type) {
    case 'integrity-report':
      errs = countErrors(obj && obj.counts);
      break;
    case 'license-claims':
      errs = obj && obj.feature_grants ? [] : ['无 v2 feature_grants（不可作为激活凭证）'];
      break;
    case 'media-manifest':
      errs = manifestErrors(obj);
      break;
    case 'media-receipt':
      errs = receiptErrors(obj);
      break;
    case 'source-adapter':
      errs = sourceErrors(obj);
      break;
    default:
      return { ok: false, errors: [`未知契约类型：${type}`] };
  }
  if (!errs || !errs.length) return { ok: true, errors: [] };
  return { ok: false, errors: errs };
}