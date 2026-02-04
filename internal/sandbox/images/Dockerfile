# Ayo Extended Sandbox Image
#
# This image extends busybox with common tools needed by agents:
# - curl: HTTP client for API calls
# - jq: JSON processing
#
# Build: docker build -t ayo-sandbox:latest -f internal/sandbox/images/Dockerfile .
# Size: ~10MB

FROM busybox:stable

# Add curl from official curl image (statically linked)
COPY --from=curlimages/curl:8.5.0 /usr/bin/curl /usr/bin/curl

# Add jq from official jq image
COPY --from=ghcr.io/jqlang/jq:1.7 /jq /usr/bin/jq

# Create non-root user for security
RUN adduser -D -u 1000 ayo && \
    mkdir -p /workspace && \
    chown ayo:ayo /workspace

# Run as non-root user
USER ayo
WORKDIR /workspace

# Default shell
CMD ["sh"]
