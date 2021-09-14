from datetime import timedelta, datetime
import re

srs_map = [
    timedelta(hours=4),
    timedelta(hours=8),
    timedelta(days=1),
    timedelta(days=3),
    timedelta(days=7),
    timedelta(days=7 * 2),
    timedelta(days=7 * 4),
    timedelta(days=7 * 16),
]


def get_next_review(srs_level=-1):
    if srs_level >= len(srs_map):
        return datetime.now() + srs_map[-1]
    elif srs_level >= 0:
        return datetime.now() + srs_map[srs_level]

    return datetime.now() + timedelta(hours=1)
