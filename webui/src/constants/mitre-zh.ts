/** MITRE ATT&CK tactic / technique Chinese labels (Enterprise). */

const TACTIC_ZH: Record<string, string> = {
  Reconnaissance: '侦察',
  'Resource Development': '资源开发',
  'Initial Access': '初始访问',
  Execution: '执行',
  Persistence: '持久化',
  'Privilege Escalation': '权限提升',
  'Defense Evasion': '防御规避',
  'Credential Access': '凭证访问',
  Discovery: '发现',
  'Lateral Movement': '横向移动',
  Collection: '收集',
  'Command and Control': '命令与控制',
  Exfiltration: '数据渗出',
  Impact: '影响',
}

const TECHNIQUE_ZH: Record<string, string> = {
  T1190: '利用面向公众的应用程序',
  T1059: '命令与脚本解释器',
  'T1059.001': 'PowerShell',
  'T1059.004': 'Unix Shell',
  T1110: '暴力破解',
  T1046: '网络服务发现',
  T1090: '代理',
  'T1090.001': '内部代理',
  T1498: '网络拒绝服务',
  'T1498.001': '直接网络洪水',
  T1562: '削弱防御',
  'T1562.001': '禁用或修改工具',
  T1071: '应用层协议',
  T1105: '入口工具传输',
  T1027: '混淆文件或信息',
  T1486: '数据加密影响',
  T1489: '服务停止',
  T1133: '外部远程服务',
  T1204: '用户执行',
  T1218: '签名代理执行',
  T1547: '启动或登录自动执行',
  T1053: '计划任务/作业',
  T1003: '操作系统凭证转储',
  T1041: 'C2 通道渗出',
  T1078: '有效账户',
}

export function tacticLabel(tactic: string): string {
  const t = tactic?.trim()
  if (!t) return '—'
  return TACTIC_ZH[t] ?? t
}

export function techniqueLabel(techniqueId: string, fallbackName?: string): string {
  const id = techniqueId?.trim().toUpperCase()
  if (!id) return fallbackName?.trim() || '—'
  const zh = TECHNIQUE_ZH[id] ?? TECHNIQUE_ZH[techniqueId.trim()]
  if (zh) return zh
  const fb = fallbackName?.trim()
  if (fb && fb !== id && fb !== techniqueId) return fb
  return id
}

export function formatTechniqueLine(techniqueId: string, name?: string, tactic?: string): string {
  const id = techniqueId?.trim() || '—'
  const zhName = techniqueLabel(id, name)
  const parts = [id, zhName]
  if (tactic?.trim()) {
    parts.push(`（${tacticLabel(tactic)}）`)
  }
  return parts.filter((p, i) => i === 0 || p !== id).join(' · ')
}

export function attackMatrixLabel(matrix: string): string {
  if (matrix === 'ICS') return 'ICS ATT&CK（工业控制）'
  if (matrix === 'Enterprise') return 'Enterprise ATT&CK（企业）'
  return matrix ? `${matrix} ATT&CK` : 'ATT&CK'
}
