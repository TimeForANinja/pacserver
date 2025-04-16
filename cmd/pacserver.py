#!/usr/bin/env python3

import sys
import logging
import traceback
import asyncio
from internal.Config import load_config, init_event_logger
from internal.Caches import init_caches
from internal.webserver import launch_server

def main():
    try:
        try:
            load_config("config.yml")
        except Exception as err:
            logging.error('Unable to load "config.yml". Exiting.')
            raise err

        # Initialize event logger
        init_event_logger()

        try:
            init_caches()
        except Exception as err:
            logging.error("Unable to initialise Caches by loading PACs and Zones. Closing Server since we're unable to recover from this.")
            raise err

        asyncio.run(launch_server())

    except Exception as err:
        logging.fatal(f"{str(err)}")
        logging.fatal(traceback.format_exc(err.__traceback__))
        sys.exit(1)

if __name__ == "__main__":
    main()
