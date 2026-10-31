# Pi development

Fork: https://github.com/12th-devs/ai-bridge

Branch: `fix/chatgpt-default-model`

Local checkout: `C:\Users\colej\Projects\WiFi\ai-bridge`

Pi: `cole@192.168.12.53`, source `/home/cole/ai-bridge-fork`.
Upstream source remains at `/home/cole/ai-bridge`.

The user systemd service `ai-bridge` uses the fork's `ai` binary through
`~/.config/systemd/user/ai-bridge.service.d/fork.conf`.
Configuration and database remain in `~/.local/share/ai-bridge`.
Pre-deployment backup: `~/.local/share/ai-bridge-backup-20261031-085449`.
Treat configuration, registration, database and backups as private; never commit them.

To validate and rebuild on the Pi:

```sh
cd ~/ai-bridge-fork
go test ./pkg/connector
go build -tags=goolm -o ai.next ./cmd/ai
mv ai.next ai
systemctl --user restart ai-bridge
systemctl --user is-active ai-bridge
```

To return to the original executable while retaining current state:

```sh
rm ~/.config/systemd/user/ai-bridge.service.d/fork.conf
systemctl --user daemon-reload
systemctl --user restart ai-bridge
```

Validated: connector tests, Pi native build, service startup.
Still requires a Beeper test: `search codex`, `start-chat gpt-5-codex`,
then send a message. Model discovery fallback does not prove endpoint acceptance.
The fallback is in memory; provider show can still report zero saved models.
