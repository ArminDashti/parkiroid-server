# How It Works

This document describes the main workflows in Parkiroid from a business perspective — what happens, who is involved, and what the outcome is.

---

## 1. A new device comes online

**Trigger:** A field device connects to the network and starts sending data for the first time.

**What happens:**

1. The device sends either a camera snapshot or a health report to the server.
2. The server recognizes this as a new device and creates a record automatically.
3. The device is now identifiable and can be queried by that identity in all future interactions.

**Who is involved:** The field device only. No administrator action is required.

**Outcome:** The device is part of the fleet and addressable by client applications.

---

## 2. Routine snapshot monitoring

**Trigger:** A device captures and uploads a camera image on a schedule or on demand.

**What happens:**

1. The device sends the image to the server.
2. The server saves the image and updates its record of the “latest” image for that device.
3. An operator or client app requests the most recent image for a specific device.
4. The server returns details about that image (timing, location on storage).

**Who is involved:** Field device (uploads), operator or client app (views).

**Outcome:** Stakeholders can see what the device’s camera last captured — useful for periodic checks or when live streaming is not needed.

**Limitation today:** The server returns information about the image, not the image file itself. The client application must handle actual image display.

---

## 3. Health monitoring

**Trigger:** A device periodically reports its operational status.

**What happens:**

1. The device sends a health report (battery, temperature, connectivity, resource usage, etc.).
2. The server stores the reading and marks it as the current status for that device.
3. An operator or client app requests the latest health data for a device.
4. The server returns the most recent readings.

**Who is involved:** Field device (reports), operator or client app (checks).

**Outcome:** Teams can identify devices that may need attention — low battery, weak signal, high temperature, or resource strain.

**Limitation today:** Only the *current* reading is available. There is no built-in history view or trend analysis.

---

## 4. Live video viewing

**Trigger:** An operator or automated process needs to see what a device’s camera shows in real time.

**What happens:**

1. The device (or viewer app) authenticates with the server.
2. The device requests permission to **publish** a live video stream; the viewer requests permission to **subscribe** to that stream.
3. The server issues short-lived access credentials for the integrated streaming service.
4. Both sides connect to the streaming service. Video flows from the device to one or more viewers.

**Who is involved:** Field device (video source), one or more viewers (watchers), server (coordinates access).

**Outcome:** Real-time visual monitoring — suitable for supervision, troubleshooting, or incident response.

**Note:** Each device has its own dedicated viewing channel. Multiple people can watch the same device simultaneously.

---

## 5. Administrator access

**Trigger:** An administrator needs to sign in to an admin tool or perform protected operations.

**What happens:**

1. The administrator enters username and password through a client application (intended to be a web admin interface).
2. The server validates the credentials and issues a time-limited access session (default: 24 hours).
3. The admin application uses that session for subsequent requests.

**Who is involved:** Administrator, admin client application.

**Outcome:** Secure access for operational oversight. Only one admin account is supported in the current version.

---

## 6. Data lifecycle

**Trigger:** Time passes; snapshot and health data ages beyond the retention policy.

**What happens:**

1. On a regular schedule (approximately hourly), the server identifies data older than the retention window (default: seven days).
2. Expired database records and image files are removed.

**Who is involved:** The server runs this automatically. No manual intervention.

**Outcome:** Storage stays manageable and old data does not accumulate indefinitely. The retention period can be configured to match organizational policy.

---

## Typical day in the life

| Time | Activity |
|------|----------|
| **Devices online** | Field devices upload snapshots and health reports throughout the day |
| **Operators checking in** | Client apps poll for latest images and health status on devices of interest |
| **Live intervention** | When something needs immediate attention, an operator opens a live video feed |
| **Background** | Server automatically purges data older than the retention period |
| **Operations** | Hosting team monitors that the server remains healthy and reachable |

---

## What the system does *not* do automatically

Understanding these boundaries helps set expectations:

- **No proactive alerts** — If a device stops reporting, nothing notifies anyone unless a separate system polls for it.
- **No device approval workflow** — Any device that can authenticate and send data will appear in the system.
- **No historical dashboards** — Trends over days or weeks are not available from the server alone.
- **No remote device control** — The server receives data; it does not push configuration or commands back to devices.
