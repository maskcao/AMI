# Glossary

| Term | Definition |
|------|------------|
| **Attack knowledge（攻击知识）** | **Core object.** Durable, analyst-validated understanding of attack behavior: TTPs, procedures, industries, involved actors/malware. Primary search and API unit. |
| **Intel report（情报）** | **Evidence layer.** One ingested report, blog post, or alert (`IntelReport`). Supports one or more attack knowledge items; holds original text and extracted IOCs. |
| **Intel entry** | Demo/API alias for `IntelReport` until Phase 2 split. |
| **Entity（实体）** | ThreatActor, Malware, Campaign — referenced by attack knowledge, not the product center. |
| **IOC** | Indicator extracted from intel reports; subsidiary evidence, **not** the default home-page pivot (unlike OpenCTI observable search). |
| **Tactic / Technique / Procedure** | ATT&CK layers; techniques attach to **attack knowledge**, not only to raw reports. |
| **Target industry** | Drives Enterprise vs ICS ATT&CK matrix on **attack knowledge**. |
| **Quality score** | Completeness; primary on attack knowledge, secondary on intel reports. |
