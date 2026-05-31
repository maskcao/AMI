PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS intel_report (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  source_url TEXT,
  source_name TEXT NOT NULL DEFAULT '',
  collect_source_name TEXT NOT NULL DEFAULT '',
  disclosed_at TEXT,
  collected_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  summary_snippet TEXT,
  body_preview TEXT,
  ingest_status TEXT NOT NULL DEFAULT 'queued',
  knowledge_action TEXT NOT NULL DEFAULT 'pending',
  extract_confidence REAL,
  pipeline_error TEXT,
  parse_version TEXT,
  enrich_version TEXT,
  has_raw INTEGER NOT NULL DEFAULT 0,
  has_parsed INTEGER NOT NULL DEFAULT 0,
  archived_at TEXT,
  ai_extract_json TEXT
);

CREATE INDEX IF NOT EXISTS idx_intel_updated ON intel_report(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_intel_status_updated ON intel_report(ingest_status, updated_at DESC);

CREATE TABLE IF NOT EXISTS intel_artifact (
  intel_id INTEGER NOT NULL,
  seq INTEGER NOT NULL,
  role TEXT NOT NULL,
  storage_path TEXT NOT NULL,
  sha256 TEXT,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  mime TEXT,
  fetched_at TEXT NOT NULL,
  PRIMARY KEY (intel_id, seq),
  FOREIGN KEY (intel_id) REFERENCES intel_report(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS intel_parsed (
  intel_id INTEGER NOT NULL,
  version TEXT NOT NULL,
  body_md_path TEXT,
  body_txt_path TEXT,
  chunk_path TEXT,
  parser_profile_id TEXT,
  parsed_at TEXT NOT NULL,
  PRIMARY KEY (intel_id, version),
  FOREIGN KEY (intel_id) REFERENCES intel_report(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS intel_ioc (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  intel_id INTEGER NOT NULL,
  type TEXT NOT NULL,
  value TEXT NOT NULL,
  confidence TEXT,
  expires_at TEXT,
  FOREIGN KEY (intel_id) REFERENCES intel_report(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ioc_value ON intel_ioc(value);
CREATE INDEX IF NOT EXISTS idx_ioc_intel ON intel_ioc(intel_id);

CREATE TABLE IF NOT EXISTS intel_figure (
  id TEXT PRIMARY KEY,
  intel_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  caption TEXT,
  kind TEXT NOT NULL,
  src TEXT NOT NULL,
  technique_ids_json TEXT,
  FOREIGN KEY (intel_id) REFERENCES intel_report(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS attack_knowledge (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  summary TEXT,
  status TEXT NOT NULL DEFAULT 'draft',
  needs_review INTEGER NOT NULL DEFAULT 0,
  attack_matrix TEXT NOT NULL DEFAULT 'Enterprise',
  target_industries_json TEXT NOT NULL DEFAULT '[]',
  quality_score INTEGER NOT NULL DEFAULT 0,
  intel_count INTEGER NOT NULL DEFAULT 0,
  updated_at TEXT NOT NULL,
  payload_json TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_knowledge_status_updated ON attack_knowledge(status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_knowledge_review ON attack_knowledge(needs_review, updated_at DESC);

CREATE TABLE IF NOT EXISTS knowledge_intel_link (
  knowledge_id INTEGER NOT NULL,
  intel_id INTEGER NOT NULL,
  role TEXT NOT NULL DEFAULT 'supporting',
  linked_at TEXT NOT NULL,
  PRIMARY KEY (knowledge_id, intel_id),
  FOREIGN KEY (knowledge_id) REFERENCES attack_knowledge(id) ON DELETE CASCADE,
  FOREIGN KEY (intel_id) REFERENCES intel_report(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_kil_intel ON knowledge_intel_link(intel_id);

CREATE TABLE IF NOT EXISTS pipeline_jobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  intel_id INTEGER NOT NULL,
  stage TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  run_at TEXT NOT NULL,
  locked_at TEXT,
  attempts INTEGER NOT NULL DEFAULT 0,
  last_error TEXT,
  FOREIGN KEY (intel_id) REFERENCES intel_report(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_run ON pipeline_jobs(status, run_at);

CREATE TABLE IF NOT EXISTS llm_response_cache (
  prompt_hash TEXT PRIMARY KEY,
  model TEXT NOT NULL,
  response_json TEXT NOT NULL,
  expires_at TEXT
);

CREATE TABLE IF NOT EXISTS llm_call_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  intel_id INTEGER,
  stage TEXT NOT NULL,
  model TEXT NOT NULL,
  prompt_hash TEXT,
  tokens_in INTEGER NOT NULL DEFAULT 0,
  tokens_out INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);
