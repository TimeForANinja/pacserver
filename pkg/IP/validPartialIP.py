import re

# Regular expression to validate IP octets
# it matches 1-4 octets (as long as they don't end with a dot)
# e.g., 10.0.4 would match as a valid partial IP
partial_ip_regex = re.compile(r'^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){0,3}$')

def is_valid_partial_ip(ip: str) -> bool:
    return bool(partial_ip_regex.match(ip))
