import json
import os
import signal
from flask import Flask, request, Response

from pkg.IP import validPartialIP
from pkg.IP.ipnet import Net
from internal.LookupTree import find_in_tree
from internal.Caches import get_lookup_tree, init_caches
from internal.Config import (
    load_config,
    init_event_logger,
)

def add_routes(app: Flask) -> None:
    @app.get("/<path:ip>")
    def ip_route(ip: str):
        # check the ip syntax
        # if it fails, we default to the / route
        if not validPartialIP.is_valid_partial_ip(ip):
            return root_route()

        # Split the IP into octets
        octets = ip.split('.')
        cidr = len(octets) * 8

        # Pad the IP to always be 4 octets
        while len(octets) < 4:
            octets.append("0")

        return get_file_for_ip(".".join(octets), cidr)

    @app.get("/<path:ip>/<cidr>")
    def ip_cidr_route(ip: str, cidr: str):
        try:
            cidr_int = int(cidr)
        except ValueError:
            return ip_route(ip)

        # check the ip syntax
        # if it fails, we default to the /:ip and then the / route
        if not validPartialIP.is_valid_partial_ip(ip):
            return ip_route(ip)

        # Pad the IP to always be 4 octets
        octets = ip.split('.')
        while len(octets) < 4:
            octets.append("0")

        return get_file_for_ip(".".join(octets), cidr_int)

    @app.get("/")
    def root_route():
        # Prefer X-Forwarded-For if provided by reverse proxy; otherwise use remote_addr
        forwarded_for = request.headers.get("X-Forwarded-For", "").split(",")[0].strip()
        client_ip = forwarded_for or (request.remote_addr or "0.0.0.0")
        return get_file_for_ip(client_ip, 32)

    @app.post("/admin/reload")
    def reload_all():
        os.kill(os.getppid(), signal.SIGUSR1)
        return Response("success", status=200)

def handle_reload_signal(signum, frame):
    print("Worker reloaded data via signal:", signum)
    init_caches()

def create_app() -> Flask:
    # Load configuration and set up event logger
    load_config("config.yml")
    init_event_logger()
    init_caches()

    # Register for SIGUSR1 in each worker
    signal.signal(signal.SIGUSR1, handle_reload_signal)

    app = Flask(__name__)
    add_routes(app)
    return app

def get_file_for_ip(ip_str: str, network_bits: int) -> Response:
    try:
        ip_net = Net.new_from_mixed(ip_str, network_bits)
    except Exception as e:
        # TODO: fallback to default PAC
        return Response(str(e), status=400)

    # search db for best pac
    pac = find_in_tree(get_lookup_tree(), ip_net)

    # TODO: fallback to default PAC
    if pac is None:
        pac = {"ip_map": {}}

    debug = request.args.get("debug")
    if debug is None:
        return Response(
            pac.get_variant(),
            mimetype="application/x-ns-proxy-autoconfig"
        )

    json_data = {
        "raw_requester": {
            "ip": ip_str,
            "cidr": network_bits
        },
        "parsed_requester": ip_net.to_string(),
        "pac": pac.to_dict() if pac is not None else pac
    }

    return Response(
        f"{json.dumps(json_data, indent=4)}\n\n---------------------------------------\n\n{pac.get_variant()}",
        mimetype="text/plain"
    )
