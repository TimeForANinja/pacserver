import os
from dataclasses import dataclass
import logging
from typing import List

from pkg.utils.listFiles import list_files

@dataclass
class PACTemplate:
    filename: str
    content: str

def read_template_files(rel_pac_dir: str) -> List[PACTemplate]:
    try:
        abs_pac_path = os.path.abspath(rel_pac_dir)
    except Exception as e:
        logging.error(f'Invalid Filepath for PACs found: "{rel_pac_dir}": {str(e)}')
        raise

    try:
        files = list_files(abs_pac_path)
    except Exception as e:
        logging.error(f'Failed to List PAC Files in "{abs_pac_path}": {str(e)}')
        raise

    templates: List[PACTemplate] = []

    for file in files:
        full_path = os.path.join(abs_pac_path, file)
        try:
            with open(full_path, 'r', encoding='utf-8') as f:
                content = f.read()

            templates.append(PACTemplate(
                filename=file,
                content=content
            ))
        except Exception as e:
            logging.warning(f'Unable to read PAC at "{full_path}": {str(e)}')
            continue

    return templates
