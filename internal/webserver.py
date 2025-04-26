import os
from fastapi import FastAPI, Request, Response
from hypercorn.config import Config
from hypercorn.asyncio import serve
import json

from pkg.IP import validPartialIP
from pkg.IP.ipnet import Net
from internal.LookupTree import find_in_tree
from internal.Caches import get_lookup_tree
from internal.Config import get_access_logger, get_event_log

app = FastAPI()

async def get_file_for_ip(request: Request, ip_str: str, network_bits: int) -> Response:
    try:
        ip_net = Net.new_from_mixed(ip_str, network_bits)
    except Exception as e:
        # TODO: fallback to default PAC
        return Response(str(e), status_code=400)

    # search db for best pac
    pac = find_in_tree(get_lookup_tree(), ip_net)

    # TODO: fallback to default PAC
    if pac is None:
        pac = {"ip_map": {}}

    debug = request.query_params.get("debug")
    if debug is None:
        return Response(
            pac.get_variant(),
            media_type="application/x-ns-proxy-autoconfig"
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
        media_type="text/plain"
    )

@app.get("/{ip}")
async def ip_route(request: Request, ip: str):
    # check the ip syntax
    # if it fails, we default to the / route
    if not validPartialIP.is_valid_partial_ip(ip):
        return await root_route(request)

    # Split the IP into octets
    octets = ip.split('.')
    cidr = len(octets) * 8

    # Pad the IP to always be 4 octets
    while len(octets) < 4:
        octets.append("0")

    return await get_file_for_ip(request, ".".join(octets), cidr)

@app.get("/{ip}/{cidr}")
async def ip_cidr_route(request: Request, ip: str, cidr: str):
    try:
        cidr_int = int(cidr)
    except ValueError:
        return await ip_route(request, ip)

    # check the ip syntax
    # if it fails, we default to the /:ip and then the / route
    if not validPartialIP.is_valid_partial_ip(ip):
        return await ip_route(request, ip)

    # Pad the IP to always be 4 octets
    octets = ip.split('.')
    while len(octets) < 4:
        octets.append("0")

    return await get_file_for_ip(request, ".".join(octets), cidr_int)

@app.get("/")
async def root_route(request: Request):
    client_ip = request.client.host
    return await get_file_for_ip(request, client_ip, 32)

async def launch_server():
    global app

    server_conf = Config()
    server_conf.accesslog = get_access_logger()
    server_conf.errorlog = get_event_log()
    server_conf.bind = ["0.0.0.0:8080"]
    server_conf.workers = os.cpu_count()
    server_conf.backlog = 400

    await serve(app, server_conf)
