import time
from typing import Optional

from flask import Flask, request, Response, g

# Prometheus metrics (supports single- and multi-process via env PROMETHEUS_MULTIPROC_DIR)
from prometheus_client import (
    Counter,
    Histogram,
    CollectorRegistry,
    generate_latest,
    CONTENT_TYPE_LATEST,
)
from prometheus_client import multiprocess
from prometheus_client.core import GaugeMetricFamily


class TcpConnectionsCollector:
    """
    Custom Prometheus collector that exports TCP connection counts by state.

    Exposes a gauge metric: tcp_connection_states{state, family}
    - state: ESTABLISHED, LISTEN, TIME_WAIT, etc.
    - family: ipv4 or ipv6

    On non-Linux systems or if /proc is not available, the metric will yield
    no samples (silently).
    """

    STATE_MAP = {
        "01": "ESTABLISHED",
        "02": "SYN_SENT",
        "03": "SYN_RECV",
        "04": "FIN_WAIT1",
        "05": "FIN_WAIT2",
        "06": "TIME_WAIT",
        "07": "CLOSE",
        "08": "CLOSE_WAIT",
        "09": "LAST_ACK",
        "0A": "LISTEN",
        "0B": "CLOSING",
        # Some kernels may expose additional states; they will be reported as UNKNOWN
    }

    def collect(self):  # pragma: no cover - runtime metric
        metric = GaugeMetricFamily(
            "tcp_connection_states",
            "Number of TCP connections by state and IP family",
            labels=["state", "family"],
        )

        try:
            v4_counts = self._parse_proc_file("/proc/net/tcp")
            v6_counts = self._parse_proc_file("/proc/net/tcp6")
        except Exception:
            # If parsing fails, return no samples to avoid breaking metrics
            v4_counts = {}
            v6_counts = {}

        # Add samples for IPv4
        for state, count in v4_counts.items():
            metric.add_metric([state, "ipv4"], float(count))

        # Add samples for IPv6
        for state, count in v6_counts.items():
            metric.add_metric([state, "ipv6"], float(count))

        yield metric

    def _parse_proc_file(self, path: str) -> dict:
        counts: dict[str, int] = {}
        try:
            with open(path, "r", encoding="utf-8", errors="ignore") as f:
                # Skip header line
                next(f, None)
                for line in f:
                    parts = line.split()
                    if len(parts) < 4:
                        continue
                    hex_state = parts[3].upper()
                    state_name = self.STATE_MAP.get(hex_state, "UNKNOWN")
                    counts[state_name] = counts.get(state_name, 0) + 1
        except FileNotFoundError:
            # Not available on this platform/container
            pass
        except Exception:
            # Be defensive: ignore any parsing errors
            pass
        return counts


def init_metrics(app: Flask) -> None:
    """
    Attach HTTP metrics hooks and the /metrics endpoint to the given Flask app.
    Safe to call once per process during app initialization.
    """

    # Define counters/histograms in-process; in multiprocess mode they write to per-worker files.
    http_requests_total = Counter(
        "http_requests_total",
        "Total HTTP requests",
        ["method", "endpoint", "status"],
    )
    http_request_duration_seconds = Histogram(
        "http_request_duration_seconds",
        "HTTP request latency in seconds",
        ["method", "endpoint", "status"],
        buckets=(
            0.001,
            0.002,
            0.005,
            0.01,
            0.025,
            0.05,
            0.1,
            0.25,
            0.5,
            1.0,
            2.5,
            5.0,
            10.0,
        ),
    )

    @app.before_request
    def _metrics_before_request():  # pragma: no cover - runtime hook
        # record start time for latency metrics
        g._start_time = time.perf_counter()

    @app.after_request
    def _metrics_after_request(response: Response):  # pragma: no cover - runtime hook
        try:
            elapsed = max(0.0, time.perf_counter() - getattr(g, "_start_time", time.perf_counter()))
            endpoint = request.endpoint or request.path or "unknown"
            status = str(getattr(response, "status_code", 0))
            labels = {
                "method": request.method,
                "endpoint": endpoint,
                "status": status,
            }
            http_requests_total.labels(**labels).inc()
            http_request_duration_seconds.labels(**labels).observe(elapsed)
        except Exception:
            # Never break the app due to metrics
            pass
        return response

    @app.get("/metrics")
    def metrics():  # pragma: no cover - scrape endpoint
        # For multiprocess, we must collect from a dedicated registry per scrape
        registry = CollectorRegistry()
        try:
            # If multiprocess mode is enabled, this will merge the child processes' metrics
            multiprocess.MultiProcessCollector(registry)
        except Exception:
            # If not in multiprocess mode, fall back to default collector registrations
            pass

        # Register custom TCP connection state collector on this scrape-time registry
        try:
            registry.register(TcpConnectionsCollector())
        except Exception:
            # Never fail the metrics endpoint due to registration problems
            pass

        data = generate_latest(registry)
        return Response(data, mimetype=CONTENT_TYPE_LATEST)
