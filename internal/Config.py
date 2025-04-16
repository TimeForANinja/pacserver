import sys
from typing import Optional

import yaml
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

def load_config(filename: str) -> None:
    global conf
    try:
        with open(filename, 'r') as f:
            data = yaml.safe_load(f)
            new_conf = Config()
            new_conf.max_cache_age = data.get('maxCacheAge', 0)
            new_conf.ip_map_file = data.get('ipMapFile', '')
            new_conf.pac_root = data.get('pacRoot', '')
            new_conf.contact_info = data.get('contactInfo', '')
            new_conf.access_log_file = data.get('accessLogFile', '')
            new_conf.event_log_file = data.get('eventLogFile', '')
            new_conf.do_auto_refresh = data.get('doAutoRefresh', False)
            conf = new_conf
            return
    except Exception as e:
        raise e

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
