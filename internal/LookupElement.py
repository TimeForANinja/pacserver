from dataclasses import dataclass
from string import Template
from typing import Optional

from internal.ProviderPACTemplates import PACTemplate
from internal.ProviderIPMap import IPMap

@dataclass
class LookupElement:
    ip_map: IPMap
    pac: Optional[PACTemplate]
    variant: str = ""

    def is_identical_net(self, other: 'LookupElement') -> bool:
        return self.ip_map.ip_net.is_identical(other.ip_map.ip_net)

    def is_identical_pac(self, other: 'LookupElement') -> bool:
        # PAC can be undefined in testing scenarios
        if self.pac is None or other.pac is None:
            return False
        return self.pac.filename == other.pac.filename

    def is_subnet_of(self, other: 'LookupElement') -> bool:
        return self.ip_map.ip_net.is_subnet_of(other.ip_map.ip_net)

    def get_raw_cidr(self) -> int:
        return self.ip_map.ip_net.get_raw_cidr()

    def get_variant(self) -> str:
        return self.variant

    def to_dict(self) -> dict:
        return {
            "ip_net": self.ip_map.ip_net.to_string(),
            "pac": self.pac.filename if self.pac else None,
        }

def new_lookup_element(ip_map: IPMap, pac: PACTemplate, contact_info: str) -> LookupElement:
    try:
        # Unfortunately, we can't use the simpler string.Template
        # since it would require us to change from the go-style "{{ .Varname }}"
        # to a python-style "$Varname"

        # Replace Go template variables with their values
        content = pac.content
        content = content.replace("{{ .Filename }}", pac.filename)
        content = content.replace("{{ .Contact }}", contact_info)

        return LookupElement(
            ip_map=ip_map,
            pac=pac,
            variant=content
        )
    except Exception as e:
        raise RuntimeError(f"Failed to create LookupElement: {str(e)}")
