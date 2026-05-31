package config

import (
	"os"
	"strconv"
)

type Config struct {
	DataDir string
	Addr    string
	Seed    bool
}

func Load() Config {
	dataDir := os.Getenv("AMI_DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	addr := os.Getenv("AMI_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	seed := os.Getenv("AMI_SEED") == "1" || os.Getenv("AMI_SEED") == "true"
	if v := os.Getenv("AMI_SEED"); v == "0" || v == "false" {
		seed = false
	}
	return Config{
		DataDir: dataDir,
		Addr:    addr,
		Seed:    seed,
	}
}

func WorkerIntervalSec() int {
	if v := os.Getenv("AMI_WORKER_INTERVAL_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 2
}

func IngestTickSec() int {
	if v := os.Getenv("AMI_INGEST_TICK_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 300
}

func IngestMinGapSec() int {
	if v := os.Getenv("AMI_INGEST_MIN_GAP_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 3600
}
