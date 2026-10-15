## 1.2.0

- **Network CUL (TCP):** `serial_port` accepts `tcp://host:port` or `telnet://host:port` (raw TCP; baud rate ignored)
- USB serial path unchanged; reconnect loop works for both transports
- Docs: Network CUL section in DOCS.md / README; Dockerfile strips Windows CRLF on `run.sh`

## 1.1.0

- Managing the duty cycle
- Improved RX packet management
- Now managing radiator thermostats too
- Added TX validation

## 1.0.0

- Initial release