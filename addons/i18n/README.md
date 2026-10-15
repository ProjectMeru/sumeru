# Internationalization (`i18n`)

Languages, `sys.translation` rows, and **gettext-compatible PO catalogs**.

## PO files per addon

Place catalogs under the owning module:

```
addons/<module>/i18n/<lang>.po
```

Examples: `i18n/fr_FR.po`, `i18n/es.po`. Language comes from the PO header `Language:` line or from the filename (normalized to Sumeru codes like `fr_FR`).

On **server start**, **install**, or **`-u <module>`**, the server imports `i18n/*.po` from every installed module on `addons_path` (kernel, standard, and custom) into `sys.translation`. `module` is that addon’s technical name. `msgid` is the English source (menu name, field label, button label). `en_US` stays the source language; set `core.user.lang` to the catalog language to see translations.

## CLI

```bash
go run ./cmd/sumeru-i18n -- -c sumeru.conf import translations.csv
go run ./cmd/sumeru-i18n -- -c sumeru.conf export -o out.csv
go run ./cmd/sumeru-i18n -- -c sumeru.conf import-po -i catalog.po -m mymodule -l fr_FR
go run ./cmd/sumeru-i18n -- -c sumeru.conf export-po -o catalog.po -m mymodule -l fr_FR
```

## Upgrade from `localization`

Uninstall or remove the old **`localization`** module row, then:

```bash
go run ./cmd/sumeru -- -c sumeru.conf -u geo -u i18n
```

Model tables (`core.lang`, `sys.translation`, geo models) are unchanged; only owning modules and menus differ.
