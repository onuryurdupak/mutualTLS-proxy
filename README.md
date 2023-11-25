# mutualTLS-proxy

#### DESCRIPTION:

**mutualTLS-proxy** handles TLS termination for ingress initiated two-way TLS (mutual authentication) and acts as a gateway for inbound traffic.

It does **not** check CRL for revoked certificates. 

#### USAGE:

**mutualTLS-proxy** is ideally deployed as an OS service providing following features:

- Automatic start-up after server reboots.
- Automatic restart upon application crashes.

We will cover an example setup which works for `Ubuntu Server 22.04`.

1. Copy application to designated server's `/opt/mutualTLS/mutualTLS-proxy/` folder. 
 
2. Crate service file at `/etc/systemd/system/mutualTLS-proxy.service` with following service config:

```
[Service]
Environment="SERVE_ADDR=:443"
Environment="PATH_SERVER_KEY_FILE=/etc/ssl/private/your-server-private-key.key"
Environment="PATH_SERVER_CERT_FILE=/etc/ssl/certs/your-server-certificate.crt"
Environment="DIR_CLIENT_CA_FILES=/opt/mutualTLS/clientCAs"
Environment="ROUTE_BASE_ADDR=https://address-to-route-traffic-to"
Environment="GATEWAY_TIMEOUT_SECS=180"
Environment="ALLOWED_HTTP_VERBS=GET;POST;PUT;DELETE"
Environment="VERBOSE_LOGGING=0"

Restart=on-failure
RestartSec=5s

ExecStart=/opt/mutualTLS/mutualTLS-proxy

[Install]
WantedBy=multi-user.target
```


(See [Environment Variables](#environment-variables) section at the end of the list for detailed descriptions of environment variables.)

3. Put trusted client certificates into `/opt/mutualTLS/clientCAs` (according to example config above.).

4. Execute `sudo systemctl enable mutualTLS-proxy.service` to enable service to auto-start after server reboots.

5. Execute `sudo systemctl restart mutualTLS-proxy.service` to start the service.

---

#### Environment Variables:

**SERVE_ADDR:** Application host address. Use `:443` to accept traffic from default HTTPS port.

**PATH_SERVER_KEY_FILE:** Path of the server's private key.

**PATH_SERVER_CERT_FILE:** Path of the server's certificate file.

**DIR_CLIENT_CA_FILES:** Root path of the trusted client CAs.

**ROUTE_BASE_ADDR:** Base address to route incoming traffic after TLS termination.

**GATEWAY_TIMEOUT_SECS:** Timeout in seconds for gateway's http client.

**ALLOWED_HTTP_VERBS:** Semicolon separated list of allowed inbound http verbs.

---

#### Reading Application Logs:

`sudo journalctl -u mutualTLS-proxy -f`

Additionally, file system logs are available at: `/var/log/mutualTLS-proxy`

---

#### Adding New Client Certificates To Trust Store

**Note:** Trusted client CAs are expected to be in `/opt/mutualTLS/clientCAs`.

1. Assuming new partner's name is `partnerX` and you are adding the certificate in the year `2024`, you should create a new folder named `2024_partnerX`.

2. Put client CAs under `/opt/mutualTLS/clientCAs/2024_parnerX`.

3. Restart **mutualTLS-proxy** service by executing: `sudo systemctl restart mutualTLS-proxy.service`

