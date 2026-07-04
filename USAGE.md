# IP River CLI - Command Reference

## Commands

### `login` / `logout` / `whoami`

Authenticate with IP River Portal.

```bash
# Log in
ipriver login

# Show current account
ipriver whoami

# Log out
ipriver logout
```

---

### `address` - Address lookup

Search addresses by postcode or UPRN.

```bash
# Search by postcode
ipriver address postcode "SW1A 1AA"

# Search by UPRN
ipriver address uprn 10033544614
```

---

### `check` - Availability checker

Check what products are available at an address.

```bash
# By postcode
ipriver check --postcode "SW1A 1AA"

# By postcode and UPRN
ipriver check --postcode "SW1A 1AA" --uprn 10033544614

# With full address identifiers
ipriver check --postcode "SW1A 1AA" --uprn 10033544614 --alk A00023962514 --dc WR

# By UPRN only
ipriver check --uprn 10033544614

# With coordinates
ipriver check --postcode "SW1A 1AA" --lat 51.501 --lon -0.1416

# Submit without waiting for results
ipriver check --postcode "SW1A 1AA" --no-wait

# Custom timeout (default 120s)
ipriver check --postcode "SW1A 1AA" --timeout 60
```

| Flag | Description |
|------|-------------|
| `--postcode` | Postcode (required unless `--uprn` provided) |
| `--uprn` | Unique Property Reference Number |
| `--alk` | Address Location Key (pair with `--dc`) |
| `--dc` | District Code (pair with `--alk`) |
| `--lat` | Latitude |
| `--lon` | Longitude |
| `--address` | Full address string |
| `--no-wait` | Return check ID immediately |
| `--timeout` | Max seconds to wait (default `120`) |

#### Subcommands

```bash
# Get status of a check
ipriver check status <id>

# Get check details
ipriver check get <id>

# Get products for a completed check
ipriver check products <id>
```

---

### `services` - Internet services

List and view your internet services.

```bash
# List all services
ipriver services list

# Filter by status
ipriver services list --status active

# Filter by technology
ipriver services list --technology ethernet

# Filter by network type
ipriver services list --network-type on_net

# Search
ipriver services list --search "Acme"

# Filter by date range
ipriver services list --from 2025-01-01 --to 2025-12-31

# Limit results
ipriver services list --limit 10

# Get a single service
ipriver services get <id>
```

| Flag | Description |
|------|-------------|
| `--status` | Filter by status (e.g. `active`, `ceased`) |
| `--technology` | Filter by technology (`ethernet`, `broadband`) |
| `--network-type` | Filter by network type (`on_net`, `off_net`) |
| `--search` | Free-text search |
| `--from` | From date (YYYY-MM-DD) |
| `--to` | To date (YYYY-MM-DD) |
| `--sort` | Field to sort by |
| `--direction` | Sort direction (`ASC` or `DESC`) |
| `--limit` | Max results (default: no limit) |

---

### `orders` - Provisioning orders

List and view broadband and ethernet orders.

#### Broadband orders

```bash
# List all broadband orders
ipriver orders broadband list

# Filter by status
ipriver orders broadband list --status active

# Search
ipriver orders broadband list --search "Acme"

# Get order details
ipriver orders broadband get <id>

# View order events
ipriver orders broadband events <id>
```

#### Ethernet orders

```bash
# List all ethernet orders
ipriver orders ethernet list

# Filter by status and search
ipriver orders ethernet list --status active --search "Acme" --from 2025-01-01

# Get order details
ipriver orders ethernet get <id>

# View order events
ipriver orders ethernet events <id>
```

**List flags (broadband and ethernet):**

| Flag | Description |
|------|-------------|
| `--status` | Filter by status |
| `--search` | Free-text search |
| `--from` | From date (YYYY-MM-DD) |
| `--to` | To date (YYYY-MM-DD) |
| `--sort` | Field to sort by |
| `--direction` | Sort direction (`ASC` or `DESC`) |
| `--limit` | Max results (default: no limit) |

---

### `tickets` - Support tickets

List, create, and reply to support tickets.

```bash
# List all tickets
ipriver tickets list

# Filter by status
ipriver tickets list --status waiting_for_support

# Search
ipriver tickets list --search "connectivity"

# Get ticket details
ipriver tickets get <id>

# View ticket events
ipriver tickets events <id>

# Create a ticket
ipriver tickets create \
  --title "Intermittent connectivity on site" \
  --description "Customer reports packet loss since 09:00 today." \
  --type intermittent_connectivity \
  --service 019d2eee-6105-7443-8116-16507e0a68a3 \
  --cc noc@example.com

# Reply to a ticket
ipriver tickets reply <id> --message "Can you provide traceroute output?"
```

**Create flags:**

| Flag | Description |
|------|-------------|
| `--title` | Short summary of the issue (required) |
| `--description` | Detailed description (required) |
| `--type` | Fault type: `total_loss`, `intermittent_connectivity`, `change_request`, `query`, `network` (required) |
| `--service` | Internet service ID to associate |
| `--cc` | CC email addresses (repeatable) |

**List flags:**

| Flag | Description |
|------|-------------|
| `--status` | Filter by status |
| `--search` | Free-text search |
| `--from` | From date (YYYY-MM-DD) |
| `--to` | To date (YYYY-MM-DD) |
| `--sort` | Field to sort by |
| `--direction` | Sort direction (`ASC` or `DESC`) |
| `--limit` | Max results (default: no limit) |

---

### `catalogue` - Product catalogue

Browse additional products available for orders.

```bash
# List all additional products
ipriver catalogue additional-products

# Filter by service type
ipriver catalogue additional-products --service-type ethernet

# Filter by product type
ipriver catalogue additional-products --product-type router

# Multiple product types
ipriver catalogue additional-products --product-type "router,care_level"
```

| Flag | Description |
|------|-------------|
| `--service-type` | Filter by service type (`ethernet`, `broadband`) |
| `--product-type` | Comma-separated product types (`onsite_install`, `ipv4_range`, `router`, `care_level`) |

---

## JSON output

Pass `--format=json` to any command for structured JSON output.

```bash
# Look up addresses
ipriver address postcode "EC2A 4NE" --format=json

# Run availability check
ipriver check \
  --postcode "EC2A 4NE" \
  --uprn 10091892923 \
  --alk A00012345678 \
  --dc EC \
  --format=json
```

Example response:

```json
{
  "check_id": "019d2eee-6105-7443-8116-16507e0a68a3",
  "status": "COMPLETED",
  "products": [
    {
      "id": "...",
      "carrier": "IP River",
      "technology": "LEASED_LINE",
      "product_name": "Ethernet 1000/1000",
      "term_length_in_months": 36,
      "bearer": 1000,
      "bandwidth": 1000,
      "sale_price": {
        "monthly_price": { "decimal": "219.67", "money": "£219.67", "currency": "GBP" },
        "annual_price": { "decimal": "2636.00", "money": "£2,636.00", "currency": "GBP" },
        "install_price": { "decimal": "0.00", "money": "£0.00", "currency": "GBP" }
      }
    }
  ]
}
```
