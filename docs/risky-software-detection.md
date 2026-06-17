# Risky Software Detection

TraceDeck can flag risky software metadata such as:

- Torrent clients
- VPN or proxy tools
- Game launchers
- Unknown browsers
- Unsigned or unknown executables
- Installers from Downloads

Signals should include app or executable labels, category, severity, source,
recommendation, and status. They should not include file contents, credentials,
tokens, private messages, screenshots, or raw browser data.
