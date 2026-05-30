# OpenCTI → AMI Mapping

## Paradigm (read first)

| | OpenCTI | AMI |
|---|---------|-----|
| **Center** | Observables / indicators (e.g. domain `baidu.com`) | **Attack knowledge** (TTP, behavior, campaign understanding) |
| **Report** | Hub node in the graph | **Intel report** — child/evidence under knowledge |
| **Typical question** | “Where did this IOC appear?” | “What is this attack about, and which reports prove it?” |

## Feature mapping

| OpenCTI | AMI | Phase |
|---------|-----|-------|
| Observable / Indicator search home | **No** default; optional `/ioc` helper | 3 |
| Report | IntelReport (情报) | 1 |
| (implicit knowledge graph) | AttackKnowledge | 2 |
| Attack Pattern | MitreTechnique + Procedure on **knowledge** | 1–2 |
| Threat Actor / Malware | Entity, linked via `involves` to **knowledge** | 2 |
| Graph pivot from IOC | Pivot from **knowledge** → intel → IOC | 2–3 |
| Connector + STIX bundle | Lightweight ingest → intel → extract knowledge | 2–3 |

**References:** https://docs.opencti.io/latest/usage/data-model/
