package seed

import (
	"encoding/json"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

// Run inserts demo-aligned rows when database is empty.
func Run(db *store.DB, files *artifacts.Store) error {
	empty, err := db.IsEmpty()
	if err != nil || !empty {
		return err
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	pl := store.KnowledgePayload{
		Techniques: []map[string]any{
			{"tactic": "Initial Access", "techniqueId": "T1190", "name": "Exploit Public-Facing Application"},
			{"tactic": "Execution", "techniqueId": "T1059", "name": "Command and Scripting Interpreter"},
		},
		Vulnerabilities: []map[string]any{
			{"cveId": "CVE-2021-44228", "description": "Log4Shell", "inWild": true},
		},
		Procedures: []map[string]any{
			{"techniqueId": "T1190", "text": "JNDI LDAP payload triggers RCE.", "evidenceIntelId": 4203},
		},
		ThreatActors: []string{"Unknown exploit cluster"},
	}
	ind, _ := json.Marshal([]string{"互联网"})
	payload, _ := json.Marshal(pl)
	_, err = db.Exec(`
		INSERT INTO attack_knowledge (id, title, summary, status, needs_review, attack_matrix, target_industries_json, quality_score, intel_count, updated_at, payload_json)
		VALUES (1001, ?, ?, 'draft', 1, 'Enterprise', ?, 72, 0, ?, ?)`,
		"Log4j2 漏洞在野利用与多恶意家族传播",
		"攻击者利用 CVE-2021-44228（Log4Shell）在公网暴露的 Java 服务中植入后门与僵尸网络载荷。",
		string(ind), ts, string(payload))
	if err != nil {
		return err
	}

	html := `<p>Multiple malware families are exploiting CVE-2021-44228 (Log4Shell) in the wild.</p><p>Ten distinct families observed.</p>`
	if err := insertIntel(db, files, 4203,
		"Ten families of malicious samples are spreading using the Log4j2 vulnerability Now",
		"360", "360 Netlab 博客", "https://blog.netlab.360.com/log4j2-spread/", html); err != nil {
		return err
	}
	_ = db.LinkIntel(1001, 4203, "primary")

	html2 := `<p>National CERT confirms ongoing Log4Shell exploitation.</p>`
	if err := insertIntel(db, files, 4205,
		"CERT advisory: active Log4Shell exploitation in critical sectors",
		"CERT", "CERT 通告 RSS", "https://cert.example.invalid/advisory", html2); err != nil {
		return err
	}
	low := "low"
	return db.InsertIOC(4205, store.IntelIOC{Type: "domain", Value: "log4j-c2.example.invalid", Confidence: &low})
}

func insertIntel(db *store.DB, files *artifacts.Store, id int64, title, sourceName, collect, sourceURL, html string) error {
	ts := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`
		INSERT INTO intel_report (id, title, source_name, collect_source_name, source_url, collected_at, updated_at, ingest_status, knowledge_action)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'queued', 'pending')`, id, title, sourceName, collect, sourceURL, ts, ts)
	if err != nil {
		return err
	}
	return pipeline.ProcessImport(db, files, id, sourceURL, []byte(html))
}
