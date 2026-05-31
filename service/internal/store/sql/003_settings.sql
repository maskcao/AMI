CREATE TABLE IF NOT EXISTS collect_source (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  type TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  schedule_cron TEXT,
  schedule_label TEXT,
  source_label TEXT NOT NULL DEFAULT '',
  config_json TEXT NOT NULL DEFAULT '{}',
  last_run_at TEXT,
  last_status TEXT,
  last_error TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS model_provider (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  kind TEXT NOT NULL DEFAULT 'llm',
  model_id TEXT,
  base_url TEXT,
  enabled INTEGER NOT NULL DEFAULT 1,
  default_for TEXT,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS parser_profile (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  used_by_sources INTEGER NOT NULL DEFAULT 0,
  config_json TEXT NOT NULL DEFAULT '{}',
  updated_at TEXT NOT NULL
);
