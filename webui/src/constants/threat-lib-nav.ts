import { markRaw, type Component } from 'vue'
import { Aim, Cpu, OfficeBuilding, SetUp, User, Warning } from '@element-plus/icons-vue'

/** Sidebar threat-intel submenu — static labels + icons (no counts). */
export interface ThreatLibNavItem {
  key: string
  label: string
  icon: Component
}

export const THREAT_LIB_NAV_ITEMS: ThreatLibNavItem[] = [
  { key: 'cve', label: '漏洞利用', icon: markRaw(Warning) },
  { key: 'malware', label: '恶意软件', icon: markRaw(Cpu) },
  { key: 'actors', label: '攻击组织', icon: markRaw(User) },
  { key: 'mitre', label: 'MITRE 分布', icon: markRaw(Aim) },
  { key: 'tools', label: '攻击工具', icon: markRaw(SetUp) },
  { key: 'infra', label: '基础设施', icon: markRaw(OfficeBuilding) },
]
