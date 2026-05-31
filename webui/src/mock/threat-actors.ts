/** Optional enrichment for threat actor detail (Demo). */
export interface ThreatActorMeta {
  aliases?: string[]
  summary?: string
  attribution?: string
}

const META: Record<string, ThreatActorMeta> = {
  'LockBit affiliate': {
    aliases: ['LockBit 3.0 affiliate', 'LB3 partner'],
    summary:
      '与 LockBit 勒索运营相关的附属团伙，常针对制造业与互联网目标实施钓鱼入侵与双重勒索。',
    attribution: 'medium',
  },
  'Unknown cluster': {
    aliases: ['UNC-Mirai-Energy', 'IoT botnet cluster'],
    summary: '尚未归因的僵尸网络运营集群，活动与 Mirai 变种及能源行业暴露资产扫描相关。',
    attribution: 'low',
  },
  '多家族利用团伙': {
    aliases: ['Log4j mass exploitation'],
    summary: 'Opportunistic 利用 Log4Shell 投递多家族载荷的松散团伙，暂无稳定 APT 命名。',
    attribution: 'low',
  },
  'Sysrv operators': {
    aliases: ['Sysrv-hello botnet'],
    summary: '运营 Sysrv-hello 僵尸网络，通过被控网站向终端用户投递 Windows 恶意可执行文件。',
    attribution: 'medium',
  },
}

export function getThreatActorMeta(name: string): ThreatActorMeta {
  return META[name] ?? { aliases: [], summary: null }
}
