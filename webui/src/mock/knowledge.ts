import { resolveAttackMatrix } from '@/constants/industry'
import type { MalwareClass } from '@/constants/malware'
import type {
  AttackInfrastructure,
  AttackTool,
  ExploitedVulnerability,
  KnowledgeDetail,
  KnowledgeDimensionCounts,
  MalwareRef,
} from '@/types/knowledge'
import type { AttackTechnique } from '@/types/intel'
import {
  buildKnowledgeQualityHints,
  computeKnowledgeQuality,
  knowledgeQualityInput,
} from '@/utils/quality'

interface RawKnowledge {
  id: number
  title: string
  summary: string | null
  targetIndustries: string[]
  techniques: AttackTechnique[]
  tools: AttackTool[]
  infrastructure: AttackInfrastructure[]
  vulnerabilities: ExploitedVulnerability[]
  malware: MalwareRef[]
  threatActors: string[]
  linkedIntel: {
    id: number
    title: string
    sourceName: string
    disclosedAt: string | null
    role: 'primary' | 'supporting'
  }[]
  updatedAt: string
}

const RAW: RawKnowledge[] = [
  {
    id: 1001,
    title: 'Log4j2 漏洞在野利用与多恶意家族传播',
    summary:
      '攻击者利用 CVE-2021-44228（Log4Shell）在公网暴露的 Java 服务中植入后门与僵尸网络载荷，出现至少十个恶意家族并呈持续扩散态势。',
    targetIndustries: ['互联网'],
    techniques: [
      {
        tactic: 'Initial Access',
        techniqueId: 'T1190',
        name: 'Exploit Public-Facing Application',
        description: '通过 Log4j JNDI 注入在暴露应用上执行任意代码。',
      },
      {
        tactic: 'Execution',
        techniqueId: 'T1059',
        name: 'Command and Scripting Interpreter',
        description: '投递 shell 脚本或二进制载荷执行。',
      },
    ],
    tools: [
      { name: 'JNDI-Exploit-Kit', category: '漏洞利用框架', description: '自动化 Log4Shell 探测与利用' },
      { name: 'bash', category: '脚本', description: '下载后续载荷' },
    ],
    infrastructure: [
      {
        type: 'domain',
        value: 'ldap.evil.example.invalid',
        role: 'c2',
        confidence: 'medium',
        note: 'JNDI 回调 LDAP 服务器',
      },
    ],
    vulnerabilities: [
      {
        cveId: 'CVE-2021-44228',
        description: 'Apache Log4j2 JNDI 注入（Log4Shell）',
        inWild: true,
      },
    ],
    malware: [
      { name: 'Generic.Backdoor', class: 'backdoor' },
      { name: 'Moobot', class: 'botnet' },
    ],
    threatActors: [],
    linkedIntel: [
      {
        id: 4203,
        title:
          'Ten families of malicious samples are spreading using the Log4j2 vulnerability Now',
        sourceName: '360',
        disclosedAt: null,
        role: 'primary',
      },
      {
        id: 4205,
        title: 'CERT advisory: active Log4Shell exploitation in critical sectors',
        sourceName: 'CERT',
        disclosedAt: '2026-05-18',
        role: 'supporting',
      },
    ],
    updatedAt: '2026-05-30T10:00:00Z',
  },
  {
    id: 1002,
    title: 'Sysrv-hello 通过被控网站向终端投递恶意程序',
    summary:
      'Sysrv-hello 僵尸网络更新传播链：入侵 Web 服务器后篡改页面，诱导访问者下载 Windows 可执行恶意文件。',
    targetIndustries: ['互联网'],
    techniques: [
      {
        tactic: 'Initial Access',
        techniqueId: 'T1190',
        name: 'Exploit Public-Facing Application',
        description: 'Web 服务被入侵后作为恶意文件分发节点。',
      },
    ],
    tools: [{ name: 'Sysrv-hello', category: '恶意载荷', description: '跨平台僵尸网络主程序' }],
    infrastructure: [
      {
        type: 'domain',
        value: 'malicious-cdn.example.invalid',
        role: 'payload',
        confidence: 'medium',
      },
    ],
    vulnerabilities: [],
    malware: [{ name: 'Sysrv-hello', class: 'botnet' }],
    threatActors: [],
    linkedIntel: [
      {
        id: 4234,
        title:
          "Threat Alert: New update from Sysrv-hello, now infecting victims' webpages to push malicious exe to end users",
        sourceName: '360',
        disclosedAt: null,
        role: 'primary',
      },
    ],
    updatedAt: '2026-05-29T14:00:00Z',
  },
  {
    id: 1003,
    title: 'Mirai 变种针对能源行业互联网暴露工控资产',
    summary:
      '基于 Mirai 的僵尸网络扫描能源电力行业中暴露在互联网的工控网关与远程维护接口，尝试默认口令与 Telnet 暴力破解，并具备 DDoS 影响能力。',
    targetIndustries: ['能源电力'],
    techniques: [
      {
        tactic: 'Initial Access',
        techniqueId: 'T0886',
        name: 'Remote Services',
        description: '对远程工控维护服务进行弱口令与暴力破解。',
      },
      {
        tactic: 'Impact',
        techniqueId: 'T0814',
        name: 'Denial of Service',
        description: '继承 Mirai 家族的 DDoS 攻击能力。',
      },
    ],
    tools: [
      { name: 'Mirai', category: '僵尸网络', description: '扫描与感染模块' },
      { name: 'telnet brute', category: '攻击模块', description: '默认口令字典爆破' },
    ],
    infrastructure: [
      {
        type: 'ip',
        value: '198.51.100.10',
        role: 'c2',
        confidence: 'high',
        note: '样本回连 C2',
      },
      { type: 'ip', value: '203.0.113.50', role: 'scan', confidence: 'medium', note: '扫描源' },
    ],
    vulnerabilities: [],
    malware: [{ name: 'Mirai', class: 'botnet' }],
    threatActors: ['Unknown cluster'],
    linkedIntel: [
      {
        id: 4301,
        title: 'Mirai variant targets energy sector ICS exposed to internet',
        sourceName: '360',
        disclosedAt: '2026-05-20',
        role: 'primary',
      },
    ],
    updatedAt: '2026-05-28T09:00:00Z',
  },
  {
    id: 1004,
    title: 'LockBit 勒索软件针对制造业的双扩展加密活动',
    summary:
      '攻击者以钓鱼与暴露 RDP 初始入侵，部署 LockBit 3.0 进行数据窃取与双重勒索；部分样本使用 BYOVD 驱动尝试禁用安全产品。',
    targetIndustries: ['制造业', '互联网'],
    techniques: [
      {
        tactic: 'Initial Access',
        techniqueId: 'T1566',
        name: 'Phishing',
        description: '恶意附件诱导执行勒索加载器。',
      },
      {
        tactic: 'Impact',
        techniqueId: 'T1486',
        name: 'Data Encrypted for Impact',
        description: '文件加密并留下勒索信。',
      },
    ],
    tools: [
      { name: 'LockBit 3.0', category: '勒索软件', description: '主勒索载荷' },
      { name: 'StealBit', category: '窃密', description: '数据外传模块' },
      { name: 'PsExec', category: '横向工具', description: '内网扩散' },
    ],
    infrastructure: [
      {
        type: 'domain',
        value: 'lockbit leak site .onion (redacted)',
        role: 'hosting',
        confidence: 'high',
        note: '数据泄露公示站点',
      },
      {
        type: 'ip',
        value: '192.0.2.88',
        role: 'c2',
        confidence: 'medium',
        note: '勒索协商 C2',
      },
    ],
    vulnerabilities: [
      {
        cveId: 'CVE-2023-28229',
        description: 'Windows 提权漏洞（示例关联）',
        inWild: true,
      },
    ],
    malware: [
      { name: 'LockBit', class: 'ransomware' },
      { name: 'Infostealer模块', class: 'trojan' },
    ],
    threatActors: ['LockBit affiliate'],
    linkedIntel: [
      {
        id: 4320,
        title: 'Ransomware advisory: LockBit campaigns against manufacturing sector',
        sourceName: '360',
        disclosedAt: '2026-05-22',
        role: 'primary',
      },
    ],
    updatedAt: '2026-05-30T08:00:00Z',
  },
]

function dimensionCounts(raw: RawKnowledge): KnowledgeDimensionCounts {
  return {
    techniques: raw.techniques.length,
    tools: raw.tools.length,
    infrastructure: raw.infrastructure.length,
    vulnerabilities: raw.vulnerabilities.length,
    malware: raw.malware.length,
  }
}

function toDetail(raw: RawKnowledge): KnowledgeDetail {
  const attackMatrix = resolveAttackMatrix(raw.targetIndustries)
  const partial = {
    id: raw.id,
    title: raw.title,
    summary: raw.summary,
    targetIndustries: raw.targetIndustries,
    attackMatrix,
    malware: raw.malware,
    threatActors: raw.threatActors,
    techniques: raw.techniques,
    tools: raw.tools,
    infrastructure: raw.infrastructure,
    vulnerabilities: raw.vulnerabilities,
    linkedIntel: raw.linkedIntel,
    intelCount: raw.linkedIntel.length,
    updatedAt: raw.updatedAt,
    dimensions: dimensionCounts(raw),
  }
  const input = knowledgeQualityInput(partial as KnowledgeDetail)
  return {
    ...partial,
    qualityScore: computeKnowledgeQuality(input),
    qualityHints: buildKnowledgeQualityHints(input),
  }
}

export const KNOWLEDGE_DETAILS: KnowledgeDetail[] = RAW.map(toDetail)

export function getKnowledgeById(id: number): KnowledgeDetail | undefined {
  return KNOWLEDGE_DETAILS.find((k) => k.id === id)
}

export function getKnowledgeByIntelId(intelId: number): KnowledgeDetail[] {
  return KNOWLEDGE_DETAILS.filter((k) => k.linkedIntel.some((i) => i.id === intelId))
}

export function listMalwareClassesInUse(): MalwareClass[] {
  const set = new Set<MalwareClass>()
  for (const k of KNOWLEDGE_DETAILS) {
    for (const m of k.malware) set.add(m.class)
  }
  return [...set]
}
