# Base

Sumeru kernel module: users, companies, partners, geo, i18n, platform metadata, and security administration.

See [addons/README.md](../README.md) for the standard addon layout. Contacts is the reference app module; **base** is the largest module and follows the same conventions for actions, split views, and menu load order.

## Install

Base is installed automatically with the server. To reload XML:

```bash
go run ./cmd/sumeru -- -c sumeru.conf -u base
```
