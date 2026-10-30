package mitre

import "strings"

// techniqueTactics maps MITRE ATT&CK technique IDs to Enterprise tactic names (English).
// Sub-techniques fall back to parent ID via TacticForTechnique.
var techniqueTactics = map[string]string{
	"T1003.001": "Credential Access",
	"T1005":     "Collection",
	"T1008":     "Command and Control",
	"T1014":     "Defense Evasion",
	"T1027":     "Defense Evasion",
	"T1036.005": "Defense Evasion",
	"T1041":     "Exfiltration",
	"T1046":     "Discovery",
	"T1047":     "Execution",
	"T1053":     "Execution",
	"T1053.002": "Execution",
	"T1053.003": "Execution",
	"T1053.005": "Execution",
	"T1055":     "Defense Evasion",
	"T1055.001": "Defense Evasion",
	"T1055.012": "Defense Evasion",
	"T1056.001": "Collection",
	"T1056.004": "Collection",
	"T1057":     "Discovery",
	"T1059":     "Execution",
	"T1059.001": "Execution",
	"T1059.003": "Execution",
	"T1059.004": "Execution",
	"T1059.005": "Execution",
	"T1059.006": "Execution",
	"T1059.007": "Execution",
	"T1068":     "Privilege Escalation",
	"T1070":     "Defense Evasion",
	"T1070.004": "Defense Evasion",
	"T1071":     "Command and Control",
	"T1071.001": "Command and Control",
	"T1071.004": "Command and Control",
	"T1078":     "Defense Evasion",
	"T1078.001": "Initial Access",
	"T1082":     "Discovery",
	"T1090":     "Command and Control",
	"T1090.002": "Command and Control",
	"T1090.003": "Command and Control",
	"T1091":     "Lateral Movement",
	"T1098":     "Persistence",
	"T1102":     "Command and Control",
	"T1102.003": "Command and Control",
	"T1105":     "Command and Control",
	"T1110":     "Credential Access",
	"T1113":     "Collection",
	"T1127":     "Defense Evasion",
	"T1136.001": "Persistence",
	"T1140":     "Defense Evasion",
	"T1189":     "Initial Access",
	"T1190":     "Initial Access",
	"T1195":     "Initial Access",
	"T1195.001": "Initial Access",
	"T1203":     "Execution",
	"T1204":     "Execution",
	"T1204.001": "Execution",
	"T1204.002": "Execution",
	"T1210":     "Lateral Movement",
	"T1218":     "Defense Evasion",
	"T1218.002": "Defense Evasion",
	"T1218.004": "Defense Evasion",
	"T1218.011": "Defense Evasion",
	"T1219":     "Command and Control",
	"T1221":     "Defense Evasion",
	"T1486":     "Impact",
	"T1496":     "Impact",
	"T1498":     "Impact",
	"T1499":     "Impact",
	"T1505":     "Persistence",
	"T1505.002": "Persistence",
	"T1505.003": "Persistence",
	"T1543.002": "Persistence",
	"T1543.003": "Persistence",
	"T1546":     "Persistence",
	"T1546.004": "Persistence",
	"T1546.015": "Persistence",
	"T1547.001": "Persistence",
	"T1547.003": "Persistence",
	"T1547.009": "Persistence",
	"T1548.002": "Privilege Escalation",
	"T1560.001": "Collection",
	"T1562":     "Defense Evasion",
	"T1562.001": "Defense Evasion",
	"T1564.001": "Defense Evasion",
	"T1564.003": "Defense Evasion",
	"T1566.001": "Initial Access",
	"T1566.002": "Initial Access",
	"T1568":     "Command and Control",
	"T1568.002": "Command and Control",
	"T1572":     "Command and Control",
	"T1573":     "Command and Control",
	"T1573.001": "Command and Control",
	"T1574.001": "Defense Evasion",
	"T1574.002": "Defense Evasion",
	"T1574.006": "Defense Evasion",
	"T1583.001": "Resource Development",
	"T1620":     "Defense Evasion",
}

// TacticForTechnique returns the MITRE Enterprise tactic name for a technique ID.
func TacticForTechnique(techniqueID string) string {
	id := strings.ToUpper(strings.TrimSpace(techniqueID))
	if id == "" {
		return ""
	}
	if t, ok := techniqueTactics[id]; ok {
		return t
	}
	if i := strings.Index(id, "."); i > 0 {
		if t, ok := techniqueTactics[id[:i]]; ok {
			return t
		}
	}
	return ""
}

// ResolveTechniqueTactic uses stored tactic when valid, otherwise looks up by technique ID.
func ResolveTechniqueTactic(techniqueID, stored string) string {
	stored = strings.TrimSpace(stored)
	if stored != "" && stored != "—" && stored != "-" {
		return stored
	}
	return TacticForTechnique(techniqueID)
}
