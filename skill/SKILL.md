---
name: cfmon
description: Monitor and manage Cloudflare Workers and Containers. Check health, list resources, get alerts.
metadata: { "openclaw": { "emoji": "☁️", "requires": { "bins": ["cfmon"] } } }
---

# cfmon — Cloudflare Workers/Containers Monitoring CLI

cfmon monitors Cloudflare Workers and Containers from the command line. Use it to list resources, check health scores, run threshold-based alerts, and watch for real-time changes.

## Authentication

Set your Cloudflare API token before using any command:

```bash
cfmon login <your-cloudflare-api-token>
```

Or pass it per-command:

```bash
cfmon --token <token> containers list <account-id>
```

The token is stored in `~/.cfmon/config.yaml`.

## Command Reference

### List resources

```bash
# List all workers
cfmon workers list <account-id>

# List all containers
cfmon containers list <account-id>

# JSON output
cfmon workers list <account-id> --output json
cfmon containers list <account-id> --output json
```

### Health score (0-100 point system)

```bash
# Overall health score
cfmon health <account-id>

# JSON output
cfmon health <account-id> --output json
```

### One-shot health check with alerts

```bash
# Check with default thresholds (CPU 80%, memory 85%, error rate 2%)
cfmon check <account-id>

# Custom thresholds
cfmon check <account-id> --cpu-threshold 70 --memory-threshold 80 --error-threshold 1

# JSON output for automation
cfmon check <account-id> --output json
```

Exit codes: `0` = healthy, `1` = warnings, `2` = critical.

### Watch mode (real-time monitoring)

```bash
# Watch containers
cfmon watch containers <account-id>

# Watch workers with custom interval
cfmon watch workers <account-id> --interval 10s

# Only show change events
cfmon watch containers <account-id> --events-only
```

### Account management

```bash
# List accounts
cfmon accounts list

# Set default account (used when account-id is omitted)
cfmon accounts set-default <account-id>
```

## Output Formats

All commands support `--output` (`-o`) with these formats:

| Format | Flag | Use case |
|--------|------|----------|
| Table | `--output table` | Human-readable (default) |
| JSON | `--output json` | Automation, piping to jq |
| JSON Lines | `--output jsonl` | Streaming, log pipelines |
| CSV | `--output csv` | Spreadsheets, data import |

Additional flags: `--quiet` (suppress decorations), `--no-header`, `--no-color`, `--fields name,cpu,memory`.

## JSON Parsing Examples

```bash
# Get all worker names
cfmon workers list <account-id> -o json | jq '.[].name'

# Check if any alerts are critical
cfmon check <account-id> -o json | jq '.summary.max_severity'

# Get alert count
cfmon check <account-id> -o json | jq '.summary.total_alerts'

# Filter critical alerts only
cfmon check <account-id> -o json | jq '[.alerts[] | select(.severity == "critical")]'
```

## Cron Job Examples

### Periodic health check with alerting

```bash
# crontab -e
# Run health check every 5 minutes, log critical issues
*/5 * * * * cfmon check <account-id> -o json >> /var/log/cfmon-check.json 2>&1

# Alert on critical status (exit code 2)
*/5 * * * * cfmon check <account-id> -o json || [ $? -eq 2 ] && echo "CRITICAL" | mail -s "cfmon alert" ops@example.com
```

### OpenClaw scheduled monitoring

```yaml
# In your OpenClaw cron configuration
schedule: "*/5 * * * *"
command: cfmon check <account-id> --output json
on_failure: notify
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "no API token provided" | Run `cfmon login <token>` or use `--token` flag |
| "no account ID provided" | Pass account ID as argument or set default with `cfmon accounts set-default` |
| Connection timeout | Increase with `--timeout 60s` |
| "API error: status 403" | Check that your API token has the correct permissions |
| Want debug output | Add `--verbose` (`-v`) flag to any command |

## Quick Diagnostic

```bash
cfmon doctor        # Check connectivity and configuration
cfmon ping          # Test API reachability
cfmon status        # Overview of all resources
```
