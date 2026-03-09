# Security Policy

## Supported Versions

Only the latest release is supported with security updates.

| Version | Supported |
|---------|-----------|
| Latest  | Yes       |
| Older   | No        |

## Reporting a Vulnerability

**Do not open a public issue for security vulnerabilities.**

Instead, please report them privately via [GitHub Security Advisories](https://github.com/fabienpiette/navilist/security/advisories/new).

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

You should receive an initial response within 7 days. If the vulnerability is accepted, a fix will be released as soon as practical.

## Scope

Navilist proxies requests to a Navidrome instance and renders a web UI. Security concerns in scope include:

- Cross-site scripting (XSS) in the web interface
- Server-side request forgery (SSRF) via the Navidrome URL configuration
- Credential exposure (Navidrome username/password in logs or responses)
- Denial of service
- Dependency vulnerabilities

Out of scope:
- Vulnerabilities in Navidrome itself
- Issues requiring direct access to the host machine
- Lack of authentication on the navilist UI (this is by design — see the Known Issues section in the README; use a reverse proxy with auth if exposure is a concern)
