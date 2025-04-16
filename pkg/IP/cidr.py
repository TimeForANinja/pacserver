from dataclasses import dataclass
import json

class CIDRException(Exception):
	"""Custom exception for invalid CIDR inputs."""
	pass

# Errors
ERR_CIDR_NOT_AN_INT = CIDRException("invalid cidr format - not an integer")
ERR_CIDR_OUT_OF_RANGE = CIDRException("invalid cidr format - out of range")

# Masks
MASK_0 = 0b00000000_00000000_00000000_00000000
MASK_8 = 0b11111111_00000000_00000000_00000000
MASK_16 = 0b11111111_11111111_00000000_00000000
MASK_24 = 0b11111111_11111111_11111111_00000000
MASK_32 = 0b11111111_11111111_11111111_11111111


def is_valid_cidr(cidr: int) -> bool:
	"""Check if the CIDR value is within the valid range."""
	return 0 <= cidr <= 32


def cidr_to_netmask(cidr: int) -> int:
	"""Calculate the netmask from a CIDR value."""
	return (0xFFFFFFFF << (32 - cidr)) & 0xFFFFFFFF


@dataclass
class CIDR:
	value: int
	mask: int

	def to_json(self) -> str:
		"""Convert the CIDR object to JSON format."""
		return json.dumps({"value": self.value, "mask": self.mask})

	@staticmethod
	def new_from_int(cidr_int: int) -> 'CIDR':
		"""Create a new CIDR object from an integer value."""
		if not is_valid_cidr(cidr_int):
			raise ERR_CIDR_OUT_OF_RANGE
		return CIDR(value=cidr_int, mask=cidr_to_netmask(cidr_int))

	@staticmethod
	def new_from_string(cidr_str: str) -> 'CIDR':
		"""Create a new CIDR object from a string value."""
		try:
			cidr_int = int(cidr_str)
		except ValueError:
			raise ERR_CIDR_NOT_AN_INT
		return CIDR.new_from_int(cidr_int)
