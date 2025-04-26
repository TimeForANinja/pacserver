FROM golang:latest

WORKDIR /app
  
# Install the application dependencies
COPY "./cmd" "/app/cmd"
COPY ./demo_files /app/demo_files
COPY ./internal /app/internal
COPY ./pkg /app/pkg
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

# build
RUN go build -o /app/dist/pacserver /app/cmd/pacserver.go

# update config
COPY <<'EOF' /app/config.yml
ipMapFile: "/app/demo_files/zones.csv"
pacRoot: "/app/demo_files/pacs"
ignoreMinors: true
maxCacheAge: -1 # disable
port: 8080
# Prometheus metrics configuration
prometheusEnabled: true
prometheusPath: "/metrics"
EOF

# Copy in the source code
EXPOSE 8080

ENTRYPOINT ["/app/dist/pacserver"]
CMD ["--serve"]
