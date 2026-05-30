# Target Industry Taxonomy

## ICS-class industries → ICS ATT&CK

- 能源电力
- 石油石化
- 制造业
- 水务
- 交通
- 智慧城市

## IT-general industries → Enterprise ATT&CK

- 金融
- 政府
- 互联网
- 教育
- 医疗
- 电信

## Rules

1. If **any** selected industry is ICS-class → matrix = **ICS**.
2. If empty or only IT-general → matrix = **Enterprise** (default).
3. Changing industry across ICS/IT boundary should trigger TTP re-validation (production).

Implementation: `webui/src/constants/industry.ts`
