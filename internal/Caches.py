import time
import logging
from typing import List, Optional
from internal.Config import get_config
from internal.ProviderIPMap import read_ip_map, IPMap
from internal.ProviderPACTemplates import read_template_files, PACTemplate
from internal.LookupTree import LookupTreeNode, build_lookup_tree, stringify_lookup_tree
from internal.ProviderLoopupTable import build_lookup_elements

_cached_ip_maps: List[IPMap] = []
_cached_pacs: List[PACTemplate] = []
_lookup_tree: Optional[LookupTreeNode] = None

def init_caches() -> None:
    """
    Initialize caches by fetching all Zones and PAC Files.
    This allows checking all conditions that would completely break the application
    directly after start.
    """
    global _cached_ip_maps, _cached_pacs
    
    config = get_config()
    
    try:
        _cached_ip_maps = read_ip_map(config.ip_map_file)
    except Exception as e:
        raise e
        
    try:
        _cached_pacs = read_template_files(config.pac_root)
    except Exception as e:
        raise e

    # initial build of the lookup tree
    update_lookup_tree()
    
    # start a regular task to refresh the lookup tree
    if config.do_auto_refresh:
        import threading
        thread = threading.Thread(target=lambda: execute_regular(update_lookup_tree))
        thread.daemon = True
        thread.start()
    
    return None

def execute_regular(task: callable) -> None:
    """Execute a task regularly based on the configured cache age."""
    while True:
        time.sleep(get_config().max_cache_age)
        logging.info("Max Cache Age reached - Refreshing Lookup Tree")
        task()

def update_lookup_tree() -> None:
    """Update the lookup tree with fresh data."""
    global _lookup_tree, _cached_pacs, _cached_ip_maps
    
    config = get_config()
    resp_build_elements = build_lookup_elements(
        _cached_ip_maps, _cached_pacs,
        config.ip_map_file,
        config.pac_root,
        config.contact_info,
    )
    if resp_build_elements is None:
        # no update
        return

    _cached_ip_maps, _cached_pacs, table = resp_build_elements

    tree = build_lookup_tree(table)
    if tree is None:
        return

    _lookup_tree = tree
    logging.info(f"The following LookupTree was loaded:\n{stringify_lookup_tree(_lookup_tree)}")

def get_lookup_tree() -> LookupTreeNode:
    global _lookup_tree
    return _lookup_tree
