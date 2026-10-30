CREATE TABLE IF NOT EXISTS incident_report (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  filename TEXT NOT NULL,
  ext TEXT NOT NULL,
  mime TEXT NOT NULL DEFAULT '',
  size_bytes INTEGER NOT NULL DEFAULT 0,
  sha256 TEXT NOT NULL DEFAULT '',
  storage_path TEXT NOT NULL,
  note TEXT NOT NULL DEFAULT '',
  report_time TEXT NOT NULL DEFAULT '',
  uploaded_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_incident_uploaded ON incident_report(uploaded_at DESC);
