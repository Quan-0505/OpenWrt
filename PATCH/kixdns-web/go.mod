// Module path is a short, non-URL name on purpose: the project is built from
// source checkouts (offline build machines, CI) and must never be resolved as a
// remote module. Internal packages are imported as "kixdns-web/internal/...".
module kixdns-web

go 1.22
