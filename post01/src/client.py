from __future__ import annotations
import argparse
import os
import sys
import requests
from typing import Any, Dict, Optional

#!/usr/bin/env python3
"""
Simple client for a Go "donuts" API (buy/sell/list).
Adjust BASE_URL to point at the running main.go server.
"""

BASE_URL = os.environ.get("DONUTS_API_URL", "http://localhost:8080")


def request_json(method: str, path: str, json: Optional[Dict] = None, timeout: int = 5) -> Any:
    url = f"{BASE_URL.rstrip('/')}/{path.lstrip('/')}"
    try:
        resp = requests.request(method, url, json=json, timeout=timeout)
    except requests.RequestException as e:
        print(f"Request failed: {e}", file=sys.stderr)
        sys.exit(2)

    if not resp.ok:
        print(f"API error {resp.status_code}: {resp.text}", file=sys.stderr)
        sys.exit(3)

    # Try to return JSON, fall back to text
    try:
        return resp.json()
    except ValueError:
        return resp.text


def buy(flavor: str, quantity: int) -> Any:
    payload = {"flavor": flavor, "quantity": quantity}
    return request_json("POST", "/donuts/buy", json=payload)


def sell(flavor: str, quantity: int) -> Any:
    payload = {"flavor": flavor, "quantity": quantity}
    return request_json("POST", "/donuts/sell", json=payload)


def list_inventory() -> Any:
    return request_json("GET", "/donuts")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Donuts API client (buy/sell/list)")
    parser.add_argument("--base-url", help="API base URL (overrides DONUTS_API_URL env var)")
    # allow running with no subcommand (demo mode)
    sub = parser.add_subparsers(dest="cmd", required=False)

    p_buy = sub.add_parser("buy", help="Buy donuts")
    p_buy.add_argument("flavor", help="Donut flavor")
    p_buy.add_argument("quantity", type=int, help="Quantity to buy (positive integer)")

    p_sell = sub.add_parser("sell", help="Sell donuts")
    p_sell.add_argument("flavor", help="Donut flavor")
    p_sell.add_argument("quantity", type=int, help="Quantity to sell (positive integer)")

    p_list = sub.add_parser("list", help="List inventory")

    return parser.parse_args()


def main() -> None:
    global BASE_URL
    args = parse_args()
    if args.base_url:
        BASE_URL = args.base_url

    # If no subcommand provided, run demo: buy and sell 100 "glazed" donuts
    if args.cmd is None:
        flavor = "glazed"
        quantity = 10
        result = buy(flavor, quantity)
        print("Bought (demo):", result)
        result = sell(flavor, quantity)
        print("Sold (demo):", result)
        return

    if args.cmd == "buy":
        if args.quantity <= 0:
            print("Quantity must be > 0", file=sys.stderr)
            sys.exit(1)
        result = buy(args.flavor, args.quantity)
        print("Bought:", result)

    elif args.cmd == "sell":
        if args.quantity <= 0:
            print("Quantity must be > 0", file=sys.stderr)
            sys.exit(1)
        result = sell(args.flavor, args.quantity)
        print("Sold:", result)

    elif args.cmd == "list":
        result = list_inventory()
        print("Inventory:", result)


if __name__ == "__main__":
    main()