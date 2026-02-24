# Monitor Pattern

The monitor pattern creates an agent that continuously observes system state and reports issues. Unlike reactive patterns, monitors proactively check health rather than responding to events.

## Overview

```
Interval → Monitor Checks → Issues Found? → Alert/Report
    ↑                            |
    └────────────────────────────┘
           (continuous loop)
```

## Basic Setup

```bash
# Check every 5 minutes
ayo trigger schedule @monitor "*/5 * * * *" \
  --prompt "Check system health. Report only if issues found." \
  --singleton
```

## Configuration Options

| Option | Description | Recommended |
|--------|-------------|-------------|
| `--singleton` | Prevent concurrent runs | Always use |
| `--timeout` | Maximum check time | 2-5 minutes |
| Interval | How often to check | 1-15 minutes |

## Examples

### System Health Monitor

Check infrastructure health:

```bash
ayo trigger schedule @health "*/5 * * * *" \
  --prompt "Check system health:
    - Disk space (alert if < 20%)
    - Memory usage (alert if > 90%)
    - CPU load (alert if sustained > 80%)
    - Key services running
    
    Report ONLY if issues found. 
    If all healthy, respond with just 'OK'." \
  --singleton \
  --timeout 2m
```

### Build Monitor

Watch CI/CD status:

```bash
ayo trigger schedule @ci-monitor "*/10 * * * *" \
  --prompt "Check CI/CD status:
    - Any failed builds in last hour?
    - Any stuck pipelines?
    - Any flaky tests?
    
    Report issues with links to failing builds." \
  --singleton
```

### Dependency Monitor

Check for security updates:

```bash
ayo trigger schedule @deps "0 9 * * *" \
  --prompt "Scan project dependencies:
    - Any known vulnerabilities?
    - Any outdated dependencies with security patches?
    - Any deprecated packages?
    
    Create tickets for critical security updates." \
  --singleton \
  --timeout 15m
```

### Log Monitor

Watch for error patterns:

```bash
ayo trigger schedule @log-watcher "*/15 * * * *" \
  --prompt "Review recent logs:
    - Any error spikes?
    - Any new error patterns?
    - Any warning trends?
    
    Report anomalies with context." \
  --singleton
```

### Database Monitor

Check database health:

```bash
ayo trigger schedule @db-monitor "*/5 * * * *" \
  --prompt "Check database health:
    - Connection pool usage
    - Slow query log (> 1s)
    - Replication lag
    - Table locks
    
    Alert on any concerning metrics." \
  --singleton \
  --timeout 3m
```

### Cost Monitor

Track cloud spending:

```bash
ayo trigger schedule @cost-watcher "0 8 * * *" \
  --prompt "Review cloud costs:
    - Any unexpected spikes?
    - Resources approaching budget limits?
    - Idle resources to clean up?
    
    Generate daily cost summary." \
  --singleton
```

## Alert Strategies

### Report Only Issues

Best for high-frequency checks:

```
"Check health. Report ONLY if issues found. Otherwise just say 'OK'."
```

### Always Report Summary

Better for daily/weekly checks:

```
"Generate health summary including:
- Issues found (if any)
- Metrics compared to yesterday
- Recommendations"
```

### Create Tickets for Issues

Integrate with ticket system:

```
"Check for issues. For any problems found:
1. Create a ticket with details
2. Assign appropriate priority
3. Tag with 'auto-detected'"
```

## Agent Prompt Best Practices

### Define Thresholds

```
Good: "Alert if disk space below 20%, memory above 90%"
Bad:  "Alert if resources are low"
```

### Specify Alert Format

```
Good: "Format alerts as: [SEVERITY] Component: Issue - Details"
Bad:  "Report problems"
```

### Handle All-Clear

```
Good: "If all checks pass, respond only with 'OK'"
Bad:  (no guidance - verbose unnecessary output)
```

### Include Context

```
Good: "Include current values, thresholds, and trend if available"
Bad:  "Say if something is wrong"
```

## Singleton Mode (Critical)

**Always use singleton for monitors:**

```bash
ayo trigger schedule @monitor "* * * * *" \
  --singleton  # REQUIRED
```

Prevents overlapping checks from creating duplicate alerts or resource contention.

## Combining with Memory

Store context for better monitoring:

```bash
# Store baseline information
ayo memory store "Normal disk usage: 40-60%"
ayo memory store "Peak hours: 9AM-6PM EST"
ayo memory store "On-call rotation: check #oncall channel"

# Monitor uses context
ayo trigger schedule @monitor "*/5 * * * *" \
  --prompt "Check system health relative to normal baselines"
```

## Escalation Patterns

### Tiered Checks

Different frequencies for different severity:

```bash
# Critical: every minute
ayo trigger schedule @critical "* * * * *" \
  --prompt "Check critical services only" \
  --singleton

# Standard: every 5 minutes
ayo trigger schedule @standard "*/5 * * * *" \
  --prompt "Standard health checks" \
  --singleton

# Deep: hourly
ayo trigger schedule @deep "0 * * * *" \
  --prompt "Comprehensive analysis" \
  --singleton \
  --timeout 15m
```

### Progressive Alerts

Track issue persistence:

```
"Check system health. For any issues:
- First occurrence: Log warning
- If issue persists 15+ minutes: Escalate to alert
- If critical or 30+ minutes: Create urgent ticket"
```

## Monitoring the Monitors

### Check Monitor Status

```bash
# List monitor triggers
ayo trigger list

# Recent monitor sessions
ayo session list --agent @monitor

# View specific check
ayo session show <session-id>
```

### Verify Monitors Running

```bash
# Check last run time
ayo trigger show <id>
# Look for "Last run" timestamp
```

## Troubleshooting

### Monitor Not Running

1. Check trigger status:
   ```bash
   ayo trigger show <id>
   ```

2. Verify daemon running:
   ```bash
   ayo service status
   ```

### Too Many Alerts

- Increase thresholds
- Add hysteresis (require sustained issues)
- Increase check interval

### Missing Real Issues

- Decrease thresholds
- Add more specific checks
- Review agent prompt for blind spots

### Monitor Timing Out

- Reduce scope of checks
- Increase timeout
- Split into multiple focused monitors

## Best Practices

1. **Start conservative** with thresholds, tune over time
2. **Use singleton mode** always
3. **Keep checks fast** (< 2 minutes for frequent monitors)
4. **Minimize noise** - only report actionable issues
5. **Include context** in alerts for faster resolution
6. **Monitor the monitors** - check they're actually running
7. **Separate concerns** - one monitor per domain

## See Also

- [Scheduled Pattern](scheduled.md)
- [Triggers Guide](../guides/triggers.md)
- [Watcher Pattern](watcher.md)
