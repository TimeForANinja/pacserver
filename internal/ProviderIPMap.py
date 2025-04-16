import csv
import os
from dataclasses import dataclass
import logging
from typing import List

from pkg.IP.ipnet import Net

@dataclass
class IPMap:
    ip_net: Net
    filename: str

def read_ip_map(rel_path: str) -> List[IPMap]:
    try:
        abs_path = os.path.abspath(rel_path)
    except Exception as e:
        logging.error(f'Invalid Filepath for IPMap found: "{rel_path}": {str(e)}')
        raise

    try:
        with open(abs_path, 'r') as file:
            mappings: List[IPMap] = []
            line_count = -1

            for line in file:
                line = line.strip()
                line_count += 1

                # Skip empty lines
                if not line:
                    continue

                # Skip comment-lines
                if line.startswith(('/', '#')):
                    continue

                # parse line as csv
                try:
                    # Using csv.reader for a single line
                    fields = next(csv.reader([line]))

                    # trim whitespace around all fields
                    fields = [field.strip() for field in fields]

                    # Ensure the CSV has exactly three fields
                    if len(fields) != 3:
                        logging.warning(
                            f"Invalid number of fields on line {line_count}, "
                            f"expected 3 but got {len(fields)}"
                        )
                        continue

                    try:
                        ip_net = Net.new_from_str(fields[0], fields[1])
                    except Exception as e:
                        logging.warning(f"Unable to parse IP From Line {line_count}: {str(e)}")
                        continue

                    mapping = IPMap(
                        ip_net=ip_net,
                        filename=fields[2]
                    )

                    # if we made it this far, then store the zone
                    mappings.append(mapping)

                except csv.Error as e:
                    logging.warning(f"Unable to Parse CSV Line {line_count}: {str(e)}")
                    continue

            return mappings

    except IOError as e:
        logging.error(f'Unable to open IPMap at "{abs_path}": {str(e)}')
        raise
