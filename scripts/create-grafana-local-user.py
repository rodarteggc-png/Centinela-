#!/usr/bin/env python3
"""Create a local Grafana user without SMTP using the loopback Admin API."""

import base64
import getpass
import json
import re
import sys
import urllib.error
import urllib.request
from pathlib import Path


LOGIN_PATTERN = re.compile(r"^[A-Za-z0-9._-]{3,64}$")
GRAFANA_URL = "http://127.0.0.1:3000"


def read_env_file(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in "\"'":
            value = value[1:-1]
        values[key.strip()] = value
    return values


def create_user(admin_user: str, admin_password: str, login: str, name: str, password: str) -> int:
    payload = {
        "name": name,
        "email": f"{login}@local.invalid",
        "login": login,
        "password": password,
        "OrgId": 1,
    }
    credentials = base64.b64encode(f"{admin_user}:{admin_password}".encode("utf-8")).decode("ascii")
    request = urllib.request.Request(
        f"{GRAFANA_URL}/api/admin/users",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Basic {credentials}",
            "Content-Type": "application/json",
            "Accept": "application/json",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=10) as response:
            result = json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as error:
        raise RuntimeError(f"Grafana rejected user creation (HTTP {error.code})") from None
    except urllib.error.URLError:
        raise RuntimeError("Grafana API is unavailable on localhost:3000") from None

    user_id = result.get("id")
    if not isinstance(user_id, int):
        raise RuntimeError("Grafana API returned no user id")
    return user_id


def main() -> int:
    env_path = Path(__file__).resolve().parents[1] / ".env"
    if not env_path.is_file():
        print("Missing /opt/centinela/.env", file=sys.stderr)
        return 1

    env = read_env_file(env_path)
    admin_user = env.get("GF_SECURITY_ADMIN_USER", "admin")
    admin_password = env.get("GF_SECURITY_ADMIN_PASSWORD", "")
    if not admin_password:
        print("GF_SECURITY_ADMIN_PASSWORD is not configured in .env", file=sys.stderr)
        return 1

    login = input("New local username (3-64 letters, digits, dot, underscore or hyphen): ").strip()
    if not LOGIN_PATTERN.fullmatch(login):
        print("Invalid username format", file=sys.stderr)
        return 1
    name = input("Display name (Enter to use username): ").strip() or login
    password = getpass.getpass("New local password (minimum 12 characters): ")
    confirmation = getpass.getpass("Confirm new local password: ")
    if len(password) < 12 or password != confirmation:
        print("Password confirmation failed or password is too short", file=sys.stderr)
        return 1

    try:
        user_id = create_user(admin_user, admin_password, login, name, password)
    except RuntimeError as error:
        print(str(error), file=sys.stderr)
        return 1

    print(f"Created local Grafana user '{login}' (id {user_id}); no invitation email was sent.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
