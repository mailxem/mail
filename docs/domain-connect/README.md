# Domain Connect templates

These let customers on Cloudflare and GoDaddy add every Xem DNS record with
one approval. Until a registrar onboards them, the setup page falls back to
opening that registrar's DNS page.

1. Generate a 2048-bit RSA key. Store the private key PEM as
   `DOMAIN_CONNECT_PRIVATE_KEY` in Infisical (backend, prod).
2. Publish the public key in DNS as TXT `_dck1.xem.email`, in the
   `p=1,a=RS256,d=<base64 DER>` format from the Domain Connect spec.
3. Open a pull request adding both JSON files to
   https://github.com/Domain-Connect/Templates.
4. After it merges, email `domain-connect@cloudflare.com` (template links,
   `xem.email` for the key, the logo, proxy off) and GoDaddy's Domain
   Connect contact listed at https://www.domainconnect.org/dns-providers/.
