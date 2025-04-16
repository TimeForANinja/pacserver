from dataclasses import dataclass

from pkg.IP.ip import IP
from pkg.IP.cidr import CIDR

@dataclass
class Net:
    network_address: IP
    cidr: CIDR

    @classmethod
    def new_ip_net(cls, ip: IP, cidr: CIDR) -> 'Net':
        return cls(
            network_address=IP(
                # calculate net address
                value=ip.value & cidr.mask
            ),
            cidr=cidr
        )

    @classmethod
    def new_from_str(cls, ip_str: str, cidr_str: str) -> 'Net':
        ip = IP.new_from_string(ip_str)

        cidr = CIDR.new_from_string(cidr_str)

        return cls.new_ip_net(ip, cidr)

    @classmethod
    def new_from_mixed(cls, ip_str: str, cidr_int: int) -> 'Net':
        ip = IP.new_from_string(ip_str)

        cidr = CIDR.new_from_int(cidr_int)

        return cls.new_ip_net(ip, cidr)

    def to_string(self) -> str:
        return f"{self.network_address.to_string()}/{str(self.get_raw_cidr())}"

    def get_raw_cidr(self) -> int:
        return self.cidr.value

    def is_subnet_of(self, net2: 'Net') -> bool:
        return net2.includes_ip(self.network_address) and self.cidr.value >= net2.cidr.value

    def includes_ip(self, ip: IP) -> bool:
        return (ip.value & self.cidr.mask) == self.network_address.value

    def is_identical(self, net2: 'Net') -> bool:
        return (self.network_address.value == net2.network_address.value and
                self.cidr.value == net2.cidr.value)
