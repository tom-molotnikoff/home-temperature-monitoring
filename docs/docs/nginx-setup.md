---
id: nginx-setup
title: Nginx Setup
sidebar_position: 3.5
---

# Nginx Setup

Nginx provides TLS termination in front of sensor-hub. It is **not** included in the sensor-hub package and must be installed separately.

## Install nginx

**Fedora / RHEL:**

```bash
sudo dnf install nginx
```

**Debian / Ubuntu:**

```bash
sudo apt install nginx
```

## Configure

Copy the example configuration shipped with the package:

```bash
sudo cp /etc/sensor-hub/nginx.conf.example /etc/nginx/conf.d/sensor-hub.conf
```

Edit `/etc/nginx/conf.d/sensor-hub.conf` and set the paths to your TLS certificate and key:

```nginx
ssl_certificate     /path/to/sensor-hub.pem;
ssl_certificate_key /path/to/sensor-hub-key.pem;
```

The example configuration proxies all requests from port 443 to `http://127.0.0.1:8080` with WebSocket upgrade support.

## TLS certificates

### Self-signed with mkcert

```bash
mkcert -install
mkcert sensor-hub
```

This creates `sensor-hub.pem` and `sensor-hub-key.pem` in the current directory.

### Let's Encrypt with certbot

```bash
sudo dnf install certbot python3-certbot-nginx   # Fedora / RHEL
sudo apt install certbot python3-certbot-nginx    # Debian / Ubuntu

sudo certbot --nginx -d sensor-hub.example.com
```

Certbot configures nginx and sets up automatic certificate renewal.

## Test and enable

```bash
sudo nginx -t
sudo systemctl enable --now nginx
```

## Verify

```bash
curl -k https://localhost/api/health
```

Expected response:

```json
{"status": "ok"}
```
