# Integrations

Connect Ayo agents to external systems.

## Overview

Agents can integrate with external systems through:

- **Hooks**: Execute scripts at lifecycle events
- **Environment Variables**: Pass configuration
- **Output Files**: Write results for other tools
- **Shell Pipelines**: Chain agents together

## Hooks for Integration

### Webhooks

Send notifications on completion:

```bash
#!/bin/bash
# hooks/agent-finish
PAYLOAD=$(cat)
curl -X POST "https://hooks.example.com/agent-complete" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD"
```

### Database Logging

Log to database:

```bash
#!/bin/bash
# hooks/agent-finish
PAYLOAD=$(cat)
RESULT=$(echo "$PAYLOAD" | jq -r '.output.result')
TIMESTAMP=$(echo "$PAYLOAD" | jq -r '.timestamp')

sqlite3 /var/log/agents.db \
  "INSERT INTO results (timestamp, result) VALUES ('$TIMESTAMP', '$RESULT')"
```

### Slack Integration

Post to Slack:

```bash
#!/bin/bash
# hooks/agent-finish
PAYLOAD=$(cat)
MESSAGE=$(echo "$PAYLOAD" | jq -r '.output.summary')

curl -X POST "${SLACK_WEBHOOK_URL}" \
  -H "Content-Type: application/json" \
  -d "{\"text\": \"Agent completed: $MESSAGE\"}"
```

## Environment Variables

Pass configuration via environment:

```bash
# Set API endpoint
export API_URL="https://api.example.com"

# Run agent
./agent "process this"
```

Access in prompt template:

```
API: {{.env "API_URL"}}
```

## Output Integration

### JSON for Tools

Write JSON for downstream processing:

```bash
./agent "input" -o output.json

# Use with jq
result=$(jq -r '.result' output.json)

# Use with Python
python process.py output.json
```

### CSV Export

Transform output to CSV:

```bash
./agent "input" -o output.json
jq -r '.items[] | [.name, .value] | @csv' output.json > results.csv
```

## Shell Pipelines

Chain agents together:

```bash
# Generate data, then process
./generator "topic" -o data.json
./processor data.json --input-format json -o result.json

# Or pipe directly (if supported)
./generator "topic" | ./processor --stdin
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Code Review
on: [pull_request]
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Review Code
        run: |
          for file in $(git diff --name-only origin/main); do
            ./code-review "$file" -o "reviews/${file}.json"
          done
      - name: Upload Reviews
        uses: actions/upload-artifact@v3
        with:
          name: reviews
          path: reviews/
```

### GitLab CI

```yaml
code-review:
  stage: test
  script:
    - ./code-review src/ -o review.json
  artifacts:
    paths:
      - review.json
```

### Jenkins

```groovy
pipeline {
    agent any
    stages {
        stage('Review') {
            steps {
                sh './code-review src/ -o review.json'
                archiveArtifacts artifacts: 'review.json'
            }
        }
    }
}
```

## API Integration

Wrap agents in HTTP APIs:

```python
# Flask example
from flask import Flask, request, jsonify
import subprocess

app = Flask(__name__)

@app.route('/analyze', methods=['POST'])
def analyze():
    data = request.json
    result = subprocess.run(
        ['./agent', data['input']],
        capture_output=True,
        text=True
    )
    return jsonify({'result': result.stdout})

if __name__ == '__main__':
    app.run()
```

## Message Queues

Process messages from queues:

```python
import subprocess
import pika

def callback(ch, method, properties, body):
    result = subprocess.run(
        ['./agent', body.decode()],
        capture_output=True,
        text=True
    )
    ch.basic_publish(
        exchange='',
        routing_key=properties.reply_to,
        body=result.stdout
    )

connection = pika.BlockingConnection(pika.ConnectionParameters('localhost'))
channel = connection.channel()
channel.basic_consume(queue='tasks', on_message_callback=callback)
channel.start_consuming()
```

## Monitoring

### Metrics Collection

Collect metrics via hooks:

```bash
#!/bin/bash
# hooks/agent-finish
PAYLOAD=$(cat)
DURATION=$(echo "$PAYLOAD" | jq -r '.duration_ms')

# Send to Prometheus Pushgateway
echo "agent_duration_ms $DURATION" | curl --data-binary @- \
  "http://prometheus-pushgateway:9091/metrics/job/agent"
```

### Health Checks

Create a health check agent:

```bash
# health-check agent
./health-check --service api --endpoint /health
```

## Best Practices

1. **Error Handling**: Handle failures gracefully in hooks
2. **Timeouts**: Set appropriate timeouts for external calls
3. **Retry Logic**: Implement retries for transient failures
4. **Logging**: Log integration events for debugging
5. **Security**: Protect API keys and credentials

## Next Steps

- [Hooks](../reference/hooks.md) - Lifecycle event handlers
- [Best Practices](best-practices.md) - Production patterns
