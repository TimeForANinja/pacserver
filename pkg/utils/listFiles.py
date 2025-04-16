from pathlib import Path
from typing import List


def list_files(directory: str) -> List[str]:
    try:
        # Use the Path object to reference the directory
        dir_path = Path(directory)

        # Use the `iterdir()` method and filter files with .is_file()
        files = [file.name for file in dir_path.iterdir() if file.is_file()]
    except OSError as e:
        print(f"Error accessing directory: {e}")
        files = []

    return files
