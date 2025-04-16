from typing import List, Optional, Tuple
import logging

from internal.ProviderPACTemplates import read_template_files
from internal.ProviderIPMap import read_ip_map
from internal.LookupElement import LookupElement, new_lookup_element
from internal.ProviderPACTemplates import PACTemplate
from internal.ProviderIPMap import IPMap

def build_lookup_elements(
        cached_ip_maps: List[IPMap], cached_pacs: List[PACTemplate],
        ip_map_file: str, pac_root: str, contact_info: str
) -> Optional[Tuple[List[IPMap], List[PACTemplate], List[LookupElement]]]:
    # store current cached PACs,
    # they can be useful when calculating LookupElements
    # if some pac has been partially deleted by accident
    old_pacs = cached_pacs

    # (try to) read new PACs / Zones
    new_ip_maps = None
    new_pacs = None
    error1 = None
    error2 = None
    try:
        new_ip_maps = read_ip_map(ip_map_file)
    except Exception as e:
        error1 = e
    try:
        new_pacs = read_template_files(pac_root)
    except Exception as e:
        error2 = e

    # check if the loading worked
    # if not print the error and try to use a cached version
    # if yes, then update the cache
    if error1 and error2:
        logging.error("Completely failed to load IPMap and PACs - keep serving cached data")
        # no need to recalculate Tree since nothing can change
        return None
    elif error1:
        logging.error("Completely failed to load IPMap - loading new? PACs with cached Zones")
        new_ip_maps = cached_ip_maps
    elif error2:
        logging.error("Completely failed to load PACs - loading new? Zones with cached PACs")
        new_pacs = old_pacs
    else:
        # no need to update the cache objects,
        # we'll simply return the updated values later on
        pass

    # build new lookup elements
    result: List[LookupElement] = []

    for ipm in new_ip_maps:
        # for each IPMap, (try to) find the corresponding pac
        match = None

        # First, try to find pac in the list of new PACs
        for pac in new_pacs:
            if pac.filename == ipm.filename:
                match = pac
                break

        # did not find in list of new PACs, try checking the cached versions
        if match is None:
            for pac in old_pacs:
                if pac.filename == ipm.filename:
                    match = pac
                    break

            # after checking the cache, write log
            if match is not None:
                logging.warning(f"Unknown PAC {ipm.filename}, using available Cached Version")
                # keep the old pac in the cache for the next check
                new_pacs.append(match)
            else:
                logging.warning(
                    f"Unknown PAC {ipm.filename}, no Cached Version available, "
                    f"skipping Zone {ipm.ip_net.to_string()}"
                )

        if match is None:
            # No match found - nothing to do
            continue

        # we found a match (after checking new and cached PACs),
        # so we'll try to parse it
        try:
            le = new_lookup_element(ipm, match, contact_info)
            result.append(le)
        except Exception as e:
            # new_lookup_element only fails when the Template could not be filled with the variables
            # Log it, and recover by skipping this zone
            logging.warning(
                f"Failed to compile Template {match.filename} for zone "
                f"{ipm.ip_net.to_string()}: {str(e)}"
            )

    return new_ip_maps, new_pacs, result
