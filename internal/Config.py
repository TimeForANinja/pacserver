import sys
import os
from typing import Optional

from logging.handlers import RotatingFileHandler
import logging

class Config:
    def __init__(self):
        self.max_cache_age = 0
        self.ip_map_file = ""
        self.pac_root = ""
        self.contact_info = ""
        self.access_log_file = ""
        self.event_log_file = ""
        self.do_auto_refresh = False

conf: Optional[Config] = None
event_log: Optional[logging.Logger] = None
access_log: Optional[logging.Logger] = None

def _parse_bool(value: str, default: bool = False) -> bool:
    if value is None:
        return default
    v = value.strip().lower()
    if v in ("1", "true", "yes", "y", "on"):  # common truthy values
        return True
    if v in ("0", "false", "no", "n", "off"):
        return False
    return default


def _parse_int(value: str, default: int = 0) -> int:
    if value is None:
        return default
    try:
        return int(value)
    except Exception:
        return default


def load_config_from_env() -> None:
    """
    Load application configuration from environment variables.

    All variables are prefixed with APP_. Supported variables:
      - APP_IP_MAP_FILE
      - APP_PAC_ROOT
      - APP_CONTACT_INFO
      - APP_ACCESS_LOG_FILE
      - APP_EVENT_LOG_FILE
      - APP_DO_AUTO_REFRESH (bool)
      - APP_MAX_CACHE_AGE (int seconds)
    """
    global conf

    new_conf = Config()

    # Strings
    new_conf.ip_map_file = os.environ.get("APP_IP_MAP_FILE", os.path.join("demo_files", "zones.csv"))
    new_conf.pac_root = os.environ.get("APP_PAC_ROOT", os.path.join("demo_files", "pacs"))
    new_conf.contact_info = os.environ.get("APP_CONTACT_INFO", "")
    new_conf.access_log_file = os.environ.get("APP_ACCESS_LOG_FILE", "./access.log")
    new_conf.event_log_file = os.environ.get("APP_EVENT_LOG_FILE", "./events.log")

    # Bools/Ints
    new_conf.do_auto_refresh = _parse_bool(os.environ.get("APP_DO_AUTO_REFRESH"), False)
    new_conf.max_cache_age = _parse_int(os.environ.get("APP_MAX_CACHE_AGE"), 0)

    conf = new_conf

def get_config() -> Config:
    global conf
    return conf

def init_event_logger():
    global event_log

    # Create a rotating file handler
    file_handler = RotatingFileHandler(
        filename=get_config().event_log_file,
        maxBytes=500 * 1024 * 1024,  # 500 MB
        backupCount=3,
    )
    file_handler.setLevel(logging.INFO)

    # Create a console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(logging.INFO)

    # Create formatter
    formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
    file_handler.setFormatter(formatter)
    console_handler.setFormatter(formatter)

    # Get the root logger
    event_log = logging.getLogger()
    event_log.setLevel(logging.INFO)

    # Add both handlers
    event_log.handlers.clear()
    event_log.addHandler(file_handler)
    event_log.addHandler(console_handler)

    # Prevent propagation to root logger to avoid duplicate logs
    event_log.propagate = False

    logging.info("Application starting")

def get_event_log() -> logging.Logger:
    global event_log
    return event_log

def get_access_logger() -> logging.Logger:
    global access_log
    if access_log is None:
        access_log = logging.getLogger("access")
        access_log.setLevel(logging.INFO)
        file_handler = RotatingFileHandler(
            filename=get_config().access_log_file,
            maxBytes=500 * 1024 * 1024,  # 500 MB
            backupCount=3,
        )
        file_handler.setFormatter(logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s'))
        file_handler.setLevel(logging.INFO)
        access_log.addHandler(file_handler)

        # Prevent propagation to root logger to avoid duplicate logs
        access_log.propagate = False
    return access_log
