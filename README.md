# Docker Clash

Docker Clash is a local management application for a bundled MetaCubeX mihomo core. The backend uses Go and Gin; the frontend is an engineered Vue 3, Vite and shadcn-vue application.

## Windows development

```powershell
.\scripts\download-mihomo.ps1
Set-Location web
npm install
npm run build
Set-Location ..
go run .\cmd
```

Open `http://127.0.0.1:9080`. For frontend hot reload, run `npm run dev` under `web/`; Vite proxies `/api` to the Go service.

## Docker

```sh
docker compose up --build -d
```

The image includes the pinned Linux mihomo core for `linux/amd64` or `linux/arm64`. UI is exposed on port `9080`. The container mixed port `7890` is published as host port `27890` by default to avoid conflicts with local proxy clients and common Windows reserved ranges. LAN access is enabled by default for new installations and can be disabled in Network Settings.

To choose another host proxy port:

```sh
NEXUS_MIXED_HOST_PORT=27890 docker compose up --build -d
```

On PowerShell, set `$env:NEXUS_MIXED_HOST_PORT = "27890"` before running Docker Compose.

### Local image archive

Build the image for the host architecture and export it as a gzip-compressed Docker archive under `dist/`:

```sh
sh scripts/build-docker-image.sh
```

The optional arguments are image tag, target architecture, and output directory. For example:

```sh
sh scripts/build-docker-image.sh v1.2.0 amd64 ./dist
```

Set `IMAGE_NAME` or `MIHOMO_VERSION` to override their defaults. Load an exported archive with:

```sh
gzip -dc dist/docker-clash-v1.2.0-linux-amd64.tar.gz | docker load
```

### Release image archives

Pushing a semantic version tag builds `linux/amd64` and `linux/arm64` Docker image archives and attaches them to the matching GitHub Release. Both the asset filenames and image tags contain the version.

```sh
git tag -a v1.2.0 -m "v1.2.0"
git push origin v1.2.0
```

Download the archive for the target architecture, verify it against `SHA256SUMS-v1.2.0.txt`, and load it:

```sh
gzip -dc docker-clash-v1.2.0-linux-amd64.tar.gz | docker load
docker run --name docker-clash -d --restart unless-stopped \
  -p 9080:9080 -p 27890:7890/tcp -p 27890:7890/udp \
  -v nexus-data:/data foliage-sea/docker-clash:v1.2.0
```

The workflow can also be run manually for an existing version tag from the Actions page.

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
