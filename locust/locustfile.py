import math
import random
import ipaddress
from typing import List
from locust import FastHttpUser, task, constant


class CIDRBasedIPGenerator:
    def __init__(self, cidrs: List[str]):
        """
        Initialize the IP generator with a list of CIDR ranges.
        Larger CIDR ranges are more likely to be chosen based on the number of IPs they contain.
        """
        if not cidrs or len(cidrs) == 0:
            raise ValueError("No CIDR ranges provided")

        # Convert CIDRs into networks and calculate their weights (size of each network)
        self.networks = [ipaddress.ip_network(cidr, strict=False) for cidr in cidrs]
        self.weights = [network.num_addresses - 2 for network in
                        self.networks]  # Exclude network and broadcast addresses

    def generate_random_ip(self):
        """Generate a random IP address proportional to the size of the CIDR ranges."""
        selected_network = random.choices(self.networks, weights=self.weights, k=1)[0]
        # calculate a random IP from the selected network
        start_ip = int(selected_network.network_address) + 1  # First usable IP
        end_ip = int(selected_network.broadcast_address) - 1  # Last usable IP
        random_ip = ipaddress.IPv4Address(random.randint(start_ip, end_ip))

        return str(random_ip)

# Define CIDR ranges for allowed IP generation
allowed_cidr = ["192.168.1.0/24", "10.0.0.0/16", "172.16.0.0/12"]

ip_generator = CIDRBasedIPGenerator(allowed_cidr)


def loop(max_loops=None):
    i = 0

    # Treat max_loops=None as infinite
    while max_loops is None or i < max_loops:
        yield i  # Expose the counter
        i += 1


class MyUser(FastHttpUser):
    # wait is the time between tasks
    #wait_time = constant(0.1)

    # Describes number of concurrent requests allowed by the FastHttpSession. Default 10.
    #concurrency = 10

    # wait a max 5 seconds before calling the request failed
    #network_timeout = 5.0
    #connection_timeout = 5.0

    ipgen = CIDRBasedIPGenerator(allowed_cidr)

    ips = [
        # generate the list ahead-of-time to avoid repeated calls to generate_random_ip()
        ip_generator.generate_random_ip() for i in range(1000)
    ]

    max_loop: int = 500
    reuse_connection: bool = True


    @task
    def get_proxy_generate_ip(self):
        for i in loop(self.max_loop):
            # Generate a random IP address from allowed CIDRs
            ip = self.ipgen.generate_random_ip()
            # Query /ip debug endpoint
            # custom name so we don't report individual IPs queried
            # timeout is set to 10 seconds to avoid long-running queries
            #print(f"Fetching {self.host}/{ip}")
            self.client.get(f'/{ip}', name="/<ip>/get_proxy_generate_ip")
            # By default, a User will reuse the same TCP/HTTP connection (unless it breaks somehow). To more realistically simulate new browsers connecting to your application this connection can be manually closed.
            if not self.reuse_connection:
                self.client.client.clientpool.close()

    @task
    def get_proxy_generate_ip_global(self):
        global ip_generator
        for i in loop(self.max_loop):
            # Generate a random IP address from allowed CIDRs
            ip = ip_generator.generate_random_ip()
            # Query /ip debug endpoint
            # custom name so we don't report individual IPs queried
            # timeout is set to 10 seconds to avoid long-running queries
            #print(f"Fetching {self.host}/{ip}")
            self.client.get(f'/{ip}', name="/<ip>/get_proxy_generate_ip_global")
            # By default, a User will reuse the same TCP/HTTP connection (unless it breaks somehow). To more realistically simulate new browsers connecting to your application this connection can be manually closed.
            if not self.reuse_connection:
                self.client.client.clientpool.close()

    @task
    def get_proxy_generate_ip_precalculated(self):
        for i in loop(self.max_loop):
            # Generate a random IP address from allowed CIDRs
            ip = self.ips[i % len(self.ips)]
            # Query /ip debug endpoint
            # custom name so we don't report individual IPs queried
            # timeout is set to 10 seconds to avoid long-running queries
            #print(f"Fetching {self.host}/{ip}")
            self.client.get(f'/{ip}', name="/<ip>/get_proxy_generate_ip_precalculated")
            # By default, a User will reuse the same TCP/HTTP connection (unless it breaks somehow). To more realistically simulate new browsers connecting to your application this connection can be manually closed.
            if not self.reuse_connection:
                self.client.client.clientpool.close()

    @task
    def get_proxy_static(self):
        for i in loop(self.max_loop):
            # Query /ip debug endpoint
            # custom name so we don't report individual IPs queried
            # timeout is set to 10 seconds to avoid long-running queries
            #print(f"Fetching {self.host}/{ip}")
            self.client.get('/127.0.0.1', name="/<ip>/get_proxy_static")

            # By default, a User will reuse the same TCP/HTTP connection (unless it breaks somehow). To more realistically simulate new browsers connecting to your application this connection can be manually closed.
            if not self.reuse_connection:
                self.client.client.clientpool.close()
