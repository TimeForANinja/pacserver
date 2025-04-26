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
allowed_cidr = ["172.16.0.0/12"]
ip_generator = CIDRBasedIPGenerator(allowed_cidr)


def loop(max_loops=None):
    i = 0
    # Treat max_loops=None as infinite
    while max_loops is None or i < max_loops:
        yield i  # Expose the counter
        i += 1

# set to None for infinite
MAX_LOOPS = 1
REUSE_CONNECTION: bool = False


class Pac01User(FastHttpUser):
    @task
    def get_random_ip(self):
        for i in loop(MAX_LOOPS):
            ip = ip_generator.generate_random_ip()
            self.client.get(f'http://pacserver01:8080/{ip}', name="pacserver01/<ip>")
            # By default, a User will reuse the same TCP/HTTP connection (unless it breaks somehow). To more realistically simulate new browsers connecting to your application this connection can be manually closed.
            if not REUSE_CONNECTION:
                self.client.client.clientpool.close()

class Pac02User(FastHttpUser):
    @task
    def get_random_ip(self):
        for i in loop(MAX_LOOPS):
            ip = ip_generator.generate_random_ip()
            self.client.get(f'http://pacserver02:8080/{ip}', name="pacserver02/<ip>")
            # By default, a User will reuse the same TCP/HTTP connection (unless it breaks somehow). To more realistically simulate new browsers connecting to your application this connection can be manually closed.
            if not REUSE_CONNECTION:
                self.client.client.clientpool.close()
