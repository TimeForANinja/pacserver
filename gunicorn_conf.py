"""
Gunicorn configuration hooks for Prometheus multiprocess metrics.

This ensures worker PIDs are cleaned up from the Prometheus multiprocess
directory when workers exit, otherwise stale metrics may persist and be
over-reported.
"""

import os

try:
    from prometheus_client import multiprocess
except Exception:  # pragma: no cover - prometheus might not be present in some dev envs
    multiprocess = None


def child_exit(server, worker):
    """Called just after a worker has been exited, in the master process."""
    if multiprocess is None:
        return
    try:
        # Mark the worker process as dead so its metrics files are ignored
        multiprocess.mark_process_dead(worker.pid)
    except Exception:
        # Never fail gunicorn due to metrics cleanup issues
        pass
