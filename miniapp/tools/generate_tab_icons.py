"""Generate small PNG tab icons without a runtime image dependency."""

from pathlib import Path
import math
import struct
import zlib

OUT = Path(__file__).resolve().parent.parent / "assets"
SIZE = 64
COLORS = {"": (164, 167, 174), "-active": (210, 154, 47)}


def point_segment_distance(px, py, ax, ay, bx, by):
    dx, dy = bx - ax, by - ay
    t = max(0, min(1, ((px - ax) * dx + (py - ay) * dy) / (dx * dx + dy * dy)))
    return math.hypot(px - (ax + t * dx), py - (ay + t * dy))


def stroke(x, y, ax, ay, bx, by, width=3.4):
    return point_segment_distance(x, y, ax, ay, bx, by) <= width


def icon_hit(kind, x, y):
    if kind == "home":
        return (
            stroke(x, y, 10, 29, 32, 10, 4.2)
            or stroke(x, y, 32, 10, 54, 29, 4.2)
            or stroke(x, y, 15, 28, 15, 54, 4.2)
            or stroke(x, y, 49, 28, 49, 54, 4.2)
            or stroke(x, y, 15, 54, 49, 54, 4.2)
            or stroke(x, y, 27, 54, 27, 37, 3.2)
            or stroke(x, y, 37, 54, 37, 37, 3.2)
            or stroke(x, y, 27, 37, 37, 37, 3.2)
        )
    if kind == "calendar":
        border = (
            stroke(x, y, 11, 18, 53, 18)
            or stroke(x, y, 11, 18, 11, 54)
            or stroke(x, y, 53, 18, 53, 54)
            or stroke(x, y, 11, 54, 53, 54)
            or stroke(x, y, 11, 30, 53, 30)
            or stroke(x, y, 22, 11, 22, 24)
            or stroke(x, y, 42, 11, 42, 24)
        )
        dots = any(math.hypot(x - cx, y - cy) <= 2.6 for cx in (22, 32, 42) for cy in (39, 47))
        return border or dots
    head = math.hypot(x - 32, y - 22) <= 10
    shoulders = ((x - 32) / 20) ** 2 + ((y - 52) / 15) ** 2 <= 1 and y >= 38
    return head or shoulders


def chunk(tag, data):
    return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)


def write_png(path, kind, color):
    rows = []
    for y in range(SIZE):
        row = bytearray([0])
        for x in range(SIZE):
            row.extend((*color, 255) if icon_hit(kind, x + 0.5, y + 0.5) else (0, 0, 0, 0))
        rows.append(bytes(row))
    payload = b"".join(rows)
    png = b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", struct.pack(">2I5B", SIZE, SIZE, 8, 6, 0, 0, 0))
    png += chunk(b"IDAT", zlib.compress(payload, 9)) + chunk(b"IEND", b"")
    path.write_bytes(png)


OUT.mkdir(exist_ok=True)
for name in ("home", "calendar", "mine"):
    for suffix, color in COLORS.items():
        write_png(OUT / f"tab-{name}{suffix}.png", name, color)
