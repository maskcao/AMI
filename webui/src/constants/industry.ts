import type { AttackMatrix } from '@/types/intel'

/** ICS-class target industries → ICS ATT&CK */
export const ICS_INDUSTRIES = [
  '能源电力',
  '石油石化',
  '制造业',
  '水务',
  '交通',
  '智慧城市',
] as const

/** IT-general industries → Enterprise ATT&CK */
export const IT_INDUSTRIES = [
  '金融',
  '政府',
  '互联网',
  '教育',
  '医疗',
  '电信',
] as const

export const ALL_INDUSTRIES = [...ICS_INDUSTRIES, ...IT_INDUSTRIES] as const

export function isIcsIndustry(industry: string): boolean {
  return (ICS_INDUSTRIES as readonly string[]).includes(industry)
}

export function resolveAttackMatrix(industries: string[]): AttackMatrix {
  if (industries.some(isIcsIndustry)) {
    return 'ICS'
  }
  return 'Enterprise'
}
