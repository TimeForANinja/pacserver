from dataclasses import dataclass
from pkg.IP.validIP import is_valid_ip

class IPException(Exception):
    """Custom exception for invalid IP inputs."""
    pass

# Error for invalid IP format
ERR_INVALID_IP_FORMAT = IPException("invalid IP format")

@dataclass
class IP:
    value: int

    def to_string(self) -> str:
        """Convert the IP object to a standard string representation (e.g., '192.168.0.1')."""
        byte1 = (self.value >> 24) & 0xFF
        byte2 = (self.value >> 16) & 0xFF
        byte3 = (self.value >> 8) & 0xFF
        byte4 = self.value & 0xFF
        return f"{byte1}.{byte2}.{byte3}.{byte4}"

    @staticmethod
    def new_from_string(src_ip: str) -> 'IP':
        """
        Create a new IP object from the given string representation of an IP address.
        """
        if not is_valid_ip(src_ip):
            raise ERR_INVALID_IP_FORMAT

        parts = src_ip.split(".")
        ip_value = 0
        for part in parts:
            ip_value = (ip_value << 8) | int(part)

        return IP(value=ip_value)
