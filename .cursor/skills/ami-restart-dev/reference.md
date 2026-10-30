# AMI dev URLs (reference)

Replace `<host-ip>` with `hostname -I` / `ip addr` (example: `192.168.238.10`).

## WebUI pages

| Feature | URL |
|---------|-----|
| Dashboard | http://<host-ip>:5173/dashboard |
| Knowledge list | http://<host-ip>:5173/knowledge |
| Knowledge detail | http://<host-ip>:5173/knowledge/1001 |
| Intel queue | http://<host-ip>:5173/intel |
| Intel detail | http://<host-ip>:5173/intel/4203 |
| Runtime logs | http://<host-ip>:5173/datacenter/logs |
| IOC search | http://<host-ip>:5173/ioc |
| Analytics CVE | http://<host-ip>:5173/analytics/cve |
| Analytics malware | http://<host-ip>:5173/analytics/malware |
| Analytics actors | http://<host-ip>:5173/analytics/actors |
| Analytics MITRE | http://<host-ip>:5173/analytics/mitre |
| Analytics tools | http://<host-ip>:5173/analytics/tools |
| Analytics infra | http://<host-ip>:5173/analytics/infra |
| Settings collect | http://<host-ip>:5173/settings/collect |
| Settings parser | http://<host-ip>:5173/settings/parser |
| Settings provider | http://<host-ip>:5173/settings/provider |

## REST (direct to API)

| Endpoint | Example |
|----------|---------|
| Health | `GET http://127.0.0.1:8080/health` |
| Intel list | `GET http://127.0.0.1:8080/api/v1/intel?page=1&page_size=20` |
| Knowledge list | `GET http://127.0.0.1:8080/api/v1/knowledge` |

Full route list: `service/internal/api/server.go` (`/api/v1` group).

## Firewall

Allow inbound **5173** (WebUI). **8080** only if debugging API without the Vite proxy.
