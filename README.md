# Nexus Proxy UI

Nexus Proxy UI is a local management application for a bundled MetaCubeX mihomo core. The backend uses Go and Gin; the frontend is an engineered Vue 3, Vite and shadcn-vue application.

## Windows development

```powershell
.\scripts\download-mihomo.ps1
Set-Location web
npm install
npm run build
Set-Location ..
go run .\cmd\nexus
```

Open `http://127.0.0.1:9080`. For frontend hot reload, run `npm run dev` under `web/`; Vite proxies `/api` to the Go service.

## Docker

```sh
docker compose up --build -d
```

The image includes the pinned Linux mihomo core for `linux/amd64` or `linux/arm64`. UI is exposed on port `9080`. The container mixed port `7890` is published as host port `27890` by default to avoid conflicts with local proxy clients and common Windows reserved ranges. Enable LAN access in Network Settings before using the proxy from another device.

To choose another host proxy port:

```sh
NEXUS_MIXED_HOST_PORT=27890 docker compose up --build -d
```

On PowerShell, set `$env:NEXUS_MIXED_HOST_PORT = "27890"` before running Docker Compose.

### Windows LAN access

Docker publishes the proxy port on the Windows host, but Windows Firewall can still block other LAN devices. Run an elevated PowerShell window once:

```powershell
.\scripts\open-windows-firewall.ps1
```

LAN clients then use the Windows host address, for example `10.10.60.20:27890`, not the container address or container port `7890`. The firewall rule only permits `LocalSubnet` sources.

## Security

The mihomo Controller is generated on `127.0.0.1:19090` with a random secret. It is not exposed by Docker. This project currently uses trusted-LAN mode and does not authenticate the UI or mixed proxy port.

## License

This project is distributed under GPL-3.0 to remain compatible with the bundled mihomo core. mihomo is copyright its respective contributors and is available from https://github.com/MetaCubeX/mihomo.
