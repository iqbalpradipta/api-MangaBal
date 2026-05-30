#!/usr/bin/env python3
"""
Manga Source Scraper
Fetches manga metadata and chapters from the configured upstream source.

Usage:
  py main.py list --page 1
  py main.py search "solo leveling"
  py main.py detail --slug mumumu
  py main.py chapters --id 10069
  py main.py download --slug mumumu --all
  py main.py genres
"""

import sys
import io

# Force UTF-8 output on Windows
if sys.platform == "win32":
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding="utf-8", errors="replace")
    sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding="utf-8", errors="replace")

from manga_source.cli import main

if __name__ == "__main__":
    main()
