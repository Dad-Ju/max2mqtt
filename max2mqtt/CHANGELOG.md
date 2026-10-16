## 1.3.0

### Features
- Full device configuration via HA MQTT discovery (`entity_category: config`)
- **ConfigTemperatures (`0x11`)**: eco/comfort/min/max, measurement offset, window-open temp & duration; retained state on `max/<id>/config`
- **ConfigValve (`0x12`)**: boost duration/position, decalc day/hour, max valve, valve offset (radiator thermostats)
- **Climate presets**: eco, comfort, boost (HA `preset_modes`)
- **Linking / safety**: disassociate partner, group ID, factory-reset button, bridge pair button
- **Week profile (`0x10`)**: JSON via text entity (advanced; prefer HA schedules for daily use)

### Behaviour
- Ack/push confirmation over CUL (no polling); optimistic MQTT update, authoritative republish on device Ack
- Ack timeout (~30s) restores previous retained config values
- Short Acks confirm config TX; longer Acks/status still update climate state

### Docs
- README/DOCS: configuration entities, push/Ack flow, week-profile JSON format

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