CREATE TABLE IF NOT EXISTS knowledge_corpus_export (
  knowledge_id INTEGER NOT NULL,
  path TEXT NOT NULL,
  format TEXT NOT NULL DEFAULT 'jsonl',
  created_at TEXT NOT NULL,
  PRIMARY KEY (knowledge_id, path),
  FOREIGN KEY (knowledge_id) REFERENCES attack_knowledge(id) ON DELETE CASCADE
);
