#!/usr/bin/env python3
"""
Thin MCP wrapper that shells out to the ipriver CLI binary.

No Go compilation needed — just point Claude Desktop at this script.
Requires: pip install mcp

Claude Desktop config (~/.config/claude/claude_desktop_config.json):

  {
    "mcpServers": {
      "ipriver": {
        "command": "python3",
        "args": ["/path/to/mcp-wrapper.py"]
      }
    }
  }

Set IPRIVER_BIN to override the binary path (default: "ipriver").
"""

from __future__ import annotations

import json
import os
import shutil
import subprocess
from typing import Any

from mcp.server.fastmcp import FastMCP

BIN = os.environ.get("IPRIVER_BIN") or shutil.which("ipriver") or "ipriver"

mcp = FastMCP("ipriver", version="1.0.0")


def _run(*args: str, input_data: str | None = None) -> dict[str, Any] | list[Any] | str:
    cmd = [BIN, "--format", "json", "--no-color", *args]
    result = subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        timeout=180,
        input=input_data,
    )
    if result.returncode != 0:
        stderr = result.stderr.strip()
        raise RuntimeError(stderr or f"ipriver exited with code {result.returncode}")
    try:
        return json.loads(result.stdout)
    except json.JSONDecodeError:
        return result.stdout.strip()


# --- Address ---


@mcp.tool()
def address_postcode(postcode: str) -> str:
    """Search addresses by UK postcode. Returns matching addresses with UPRN, ALK, and district code."""
    return json.dumps(_run("address", "postcode", postcode), indent=2)


@mcp.tool()
def address_uprn(uprn: str) -> str:
    """Look up a single address by its Unique Property Reference Number (UPRN)."""
    return json.dumps(_run("address", "uprn", uprn), indent=2)


# --- Availability check ---


@mcp.tool()
def check_availability(
    postcode: str = "",
    uprn: str = "",
    alk: str = "",
    dc: str = "",
    lat: float | None = None,
    lon: float | None = None,
    address: str = "",
    timeout: int = 120,
) -> str:
    """Run a product availability check for a UK address. At least postcode or uprn is required. Polls until complete and returns available products with pricing."""
    if not postcode and not uprn:
        raise ValueError("At least postcode or uprn is required")
    args = ["check"]
    if postcode:
        args += ["--postcode", postcode]
    if uprn:
        args += ["--uprn", uprn]
    if alk:
        args += ["--alk", alk]
    if dc:
        args += ["--dc", dc]
    if lat is not None:
        args += ["--lat", str(lat)]
    if lon is not None:
        args += ["--lon", str(lon)]
    if address:
        args += ["--address", address]
    if timeout != 120:
        args += ["--timeout", str(timeout)]
    return json.dumps(_run(*args), indent=2)


@mcp.tool()
def check_status(id: str) -> str:
    """Get the current status of an availability check."""
    return json.dumps(_run("check", "status", id), indent=2)


@mcp.tool()
def check_get(id: str) -> str:
    """Get full details of an availability check request."""
    return json.dumps(_run("check", "get", id), indent=2)


@mcp.tool()
def check_products(id: str) -> str:
    """Get the list of available products for a completed availability check."""
    return json.dumps(_run("check", "products", id), indent=2)


# --- Services ---


@mcp.tool()
def services_list(
    status: str = "",
    technology: str = "",
    network_type: str = "",
    search: str = "",
    from_date: str = "",
    to_date: str = "",
    sort: str = "",
    direction: str = "",
    limit: int = 0,
) -> str:
    """List internet services. Supports filtering by status, technology, network type, date range, and search."""
    args = ["services", "list"]
    if status:
        args += ["--status", status]
    if technology:
        args += ["--technology", technology]
    if network_type:
        args += ["--network-type", network_type]
    if search:
        args += ["--search", search]
    if from_date:
        args += ["--from", from_date]
    if to_date:
        args += ["--to", to_date]
    if sort:
        args += ["--sort", sort]
    if direction:
        args += ["--direction", direction]
    if limit:
        args += ["--limit", str(limit)]
    return json.dumps(_run(*args), indent=2)


@mcp.tool()
def services_get(id: str) -> str:
    """Get details of a single internet service by ID."""
    return json.dumps(_run("services", "get", id), indent=2)


# --- Orders: Broadband ---


@mcp.tool()
def orders_broadband_list(
    status: str = "",
    search: str = "",
    from_date: str = "",
    to_date: str = "",
    sort: str = "",
    direction: str = "",
    limit: int = 0,
) -> str:
    """List broadband provisioning orders."""
    args = ["orders", "broadband", "list"]
    if status:
        args += ["--status", status]
    if search:
        args += ["--search", search]
    if from_date:
        args += ["--from", from_date]
    if to_date:
        args += ["--to", to_date]
    if sort:
        args += ["--sort", sort]
    if direction:
        args += ["--direction", direction]
    if limit:
        args += ["--limit", str(limit)]
    return json.dumps(_run(*args), indent=2)


@mcp.tool()
def orders_broadband_get(id: str) -> str:
    """Get details of a single broadband order."""
    return json.dumps(_run("orders", "broadband", "get", id), indent=2)


@mcp.tool()
def orders_broadband_events(id: str) -> str:
    """List events/timeline for a broadband order."""
    return json.dumps(_run("orders", "broadband", "events", id), indent=2)


# --- Orders: Ethernet ---


@mcp.tool()
def orders_ethernet_list(
    status: str = "",
    search: str = "",
    from_date: str = "",
    to_date: str = "",
    sort: str = "",
    direction: str = "",
    limit: int = 0,
) -> str:
    """List ethernet provisioning orders."""
    args = ["orders", "ethernet", "list"]
    if status:
        args += ["--status", status]
    if search:
        args += ["--search", search]
    if from_date:
        args += ["--from", from_date]
    if to_date:
        args += ["--to", to_date]
    if sort:
        args += ["--sort", sort]
    if direction:
        args += ["--direction", direction]
    if limit:
        args += ["--limit", str(limit)]
    return json.dumps(_run(*args), indent=2)


@mcp.tool()
def orders_ethernet_get(id: str) -> str:
    """Get details of a single ethernet order."""
    return json.dumps(_run("orders", "ethernet", "get", id), indent=2)


@mcp.tool()
def orders_ethernet_events(id: str) -> str:
    """List events/timeline for an ethernet order."""
    return json.dumps(_run("orders", "ethernet", "events", id), indent=2)


# --- Tickets ---


@mcp.tool()
def tickets_list(
    status: str = "",
    search: str = "",
    from_date: str = "",
    to_date: str = "",
    sort: str = "",
    direction: str = "",
    limit: int = 0,
) -> str:
    """List support tickets."""
    args = ["tickets", "list"]
    if status:
        args += ["--status", status]
    if search:
        args += ["--search", search]
    if from_date:
        args += ["--from", from_date]
    if to_date:
        args += ["--to", to_date]
    if sort:
        args += ["--sort", sort]
    if direction:
        args += ["--direction", direction]
    if limit:
        args += ["--limit", str(limit)]
    return json.dumps(_run(*args), indent=2)


@mcp.tool()
def tickets_get(id: str) -> str:
    """Get details of a single support ticket."""
    return json.dumps(_run("tickets", "get", id), indent=2)


@mcp.tool()
def tickets_events(id: str) -> str:
    """List events/messages for a support ticket."""
    return json.dumps(_run("tickets", "events", id), indent=2)


@mcp.tool()
def tickets_create(
    title: str,
    description: str,
    type: str,
    service: str = "",
    cc: str = "",
) -> str:
    """Create a new support ticket. type must be one of: total_loss, intermittent_connectivity, change_request, query, network."""
    args = ["tickets", "create", "--title", title, "--description", description, "--type", type]
    if service:
        args += ["--service", service]
    if cc:
        for email in cc.split(","):
            email = email.strip()
            if email:
                args += ["--cc", email]
    return json.dumps(_run(*args), indent=2)


@mcp.tool()
def tickets_reply(id: str, message: str) -> str:
    """Send a reply message on an existing support ticket."""
    return json.dumps(_run("tickets", "reply", id, "--message", message), indent=2)


# --- Catalogue ---


@mcp.tool()
def catalogue_additional_products(
    service_type: str = "",
    product_type: str = "",
) -> str:
    """List additional products (routers, care levels, IPv4 ranges, onsite install). Optionally filter by service_type (ethernet/broadband) or product_type (onsite_install, ipv4_range, router, care_level)."""
    args = ["catalogue", "additional-products"]
    if service_type:
        args += ["--service-type", service_type]
    if product_type:
        args += ["--product-type", product_type]
    return json.dumps(_run(*args), indent=2)


# --- Account ---


@mcp.tool()
def whoami() -> str:
    """Show the currently authenticated IP River account."""
    return json.dumps(_run("whoami"), indent=2)


if __name__ == "__main__":
    mcp.run(transport="stdio")
