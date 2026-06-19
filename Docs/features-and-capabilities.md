# Features & Capabilities

This document describes what Parkiroid Server can do today, what works only partially, and what is not yet part of the product.

---

## Fully available today

### Camera snapshot monitoring

Devices can upload still images from their cameras. The server stores each image and always knows which is the **most recent** image for each device. Operators or client apps can request that latest snapshot to see what a device’s camera is showing without starting a live stream.

**Business value:** Quick status checks, thumbnail previews, and a fallback when live video is unavailable or unnecessary.

### Device health monitoring

Devices can report operational health data, including:

- Processor and memory usage
- Storage usage
- Battery level
- Device temperature
- Wireless or cellular signal strength

The server keeps the **latest reading** per device so operators can see current conditions at a glance.

**Business value:** Spot devices with low battery, poor connectivity, overheating, or resource problems before they fail completely.

### Live video streaming

Authorized users can open a **real-time video feed** from a device. Each device has its own dedicated viewing channel. Multiple viewers can watch the same device at the same time.

Devices act as video sources; viewers receive the stream. This is suitable for situations where a still image is not enough — live supervision, incident response, or detailed inspection.

**Business value:** Real-time visibility without building custom video infrastructure.

### Secure access

The system requires authentication for sensitive operations. Two access paths exist:

- **Embedded access for apps and devices** — Client applications and field devices use a pre-configured secure token built into the app. This is the primary path for day-to-day device and viewer access.
- **Administrator login** — A username-and-password sign-in for operators or an admin interface, producing a time-limited session (default: 24 hours).

**Business value:** Only authorized apps, devices, and people can access device data and streams.

### Automatic device registration

When a device sends its first snapshot or health report, the server creates a record for it automatically. No separate provisioning, pairing code, or admin approval is required.

**Business value:** Faster rollout — devices become visible as soon as they connect.

### Data retention and cleanup

Snapshot and health data older than the configured retention window (default: **seven days**) is automatically deleted. This includes both stored records and image files.

**Business value:** Predictable storage use and alignment with data-minimization policies. The retention period can be adjusted for organizational needs.

### System health monitoring

The server exposes a simple health check so operations teams and hosting tools can verify that the service is running and reachable.

**Business value:** Supports uptime monitoring and automated recovery in production environments.

### Self-describing service

The server can list all of its available operations in a machine-readable way. This helps integration teams and documentation stay aligned as the product evolves.

---

## Partially available

| Capability | Current state | Impact |
|------------|---------------|--------|
| **Device inventory** | The server tracks all known devices internally, but there is no way yet for client apps to retrieve a full device list through the product | Operators may need another source of truth for “which devices exist” until this is exposed |
| **Image delivery** | Requesting the latest snapshot returns information *about* the image (when it was taken, where it is stored), not the image file itself | Client apps must handle image retrieval separately; this adds integration work |

---

## Not yet built

The following capabilities are **not present** in the current product. They may be natural follow-on work but are not formally planned or tracked in this repository:

| Area | Gap |
|------|-----|
| **User management** | Only a single administrator account; no team accounts, roles, or permissions beyond basic viewer vs. publisher for live video |
| **Admin interface** | No built-in web dashboard; admin login exists to support a separate UI |
| **Historical reporting** | Only the *latest* snapshot and health reading per device; no trends, charts, or time-range browsing |
| **Alerts and notifications** | No automatic warnings when a device goes offline, battery drops, or thresholds are exceeded |
| **Device configuration** | No remote control or settings management for field devices |
| **Multi-organization support** | Single-tenant design; no separation between different customers or business units on one server |
| **Quality assurance automation** | No automated test suite or continuous integration pipeline in the repository |
| **Client applications** | Mobile and web apps referenced in design are external to this project |

---

## Feature summary at a glance

| Need | Supported? |
|------|------------|
| See the latest camera image from a device | Yes |
| Check current device health | Yes |
| Watch live video from a device | Yes |
| List all devices from the server | No (internal only) |
| Browse historical images or metrics | No |
| Receive alerts when something goes wrong | No |
| Manage multiple admin users | No |
| Deploy on your own infrastructure | Yes |
| Automatic cleanup of old data | Yes |
