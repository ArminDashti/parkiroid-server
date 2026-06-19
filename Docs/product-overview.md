# Product Overview

## What is Parkiroid?

Parkiroid is a platform for **monitoring and viewing remote field devices** — hardware deployed outside a traditional office or data center. The server component (this project) acts as the central hub that receives data from devices in the field and makes it available to operators and client applications.

The name suggests a focus on **parking-related** or **mobile (Android-based)** edge devices, though the backend itself is domain-agnostic: it handles camera images, device health, and live video regardless of the specific business use case.

## What problem does it solve?

Organizations that deploy distributed devices — cameras, sensors, or similar equipment in parking lots, streets, warehouses, or other remote locations — face recurring challenges:

- **Visibility** — Is the device working? What is its camera showing right now?
- **Health awareness** — Is the battery low? Is connectivity poor? Is the device overheating?
- **Live observation** — When a still image is not enough, can someone watch a live feed?

Parkiroid Server addresses these needs by providing a single, central place where devices report in and where authorized users can check status or watch live video — without each team building their own streaming and storage infrastructure from scratch.

## How the pieces fit together

```
┌─────────────────┐         ┌──────────────────────┐         ┌─────────────────┐
│  Field devices  │ ──────► │  Parkiroid Server    │ ◄────── │  Client apps    │
│  (cameras,      │  send   │  (this project)      │  fetch  │  (mobile/web —  │
│   sensors)      │  data   │                      │  data   │   separate)     │
└─────────────────┘         └──────────────────────┘         └─────────────────┘
                                      │
                                      ▼
                            Live video streaming
                            (integrated partner service)
```

- **Field devices** upload camera snapshots and health readings, and can publish live video.
- **Parkiroid Server** stores recent data, tracks which devices exist, and coordinates access.
- **Client applications** (not part of this repository) let operators view snapshots, check device health, and watch live streams.
- **Live video** is handled through an integrated streaming service so viewers get real-time feeds without custom video infrastructure.

## Who uses the system?

| Role | What they do |
|------|--------------|
| **Field devices** | Send camera images, report health metrics, and optionally stream live video |
| **Operators / viewers** | Check device status, view the latest camera image, or watch a live feed |
| **Client application** | The app (mobile or web) that devices and operators interact with; connects to the server on their behalf |
| **Administrator** | Signs in to manage or oversee the system; intended for use through a companion admin interface |
| **Operations / IT** | Deploys and maintains the server on company infrastructure, manages access credentials, and ensures uptime |

Devices **register themselves automatically** the first time they send data. There is no manual onboarding or approval step in the current version — a new device simply appears once it starts reporting.

## Deployment model

Parkiroid Server is designed to be **self-hosted**: it runs on infrastructure you control (on-premise servers, private cloud, or edge locations) rather than as a multi-tenant SaaS product. This suits organizations that need data to stay on their own network or that operate in environments with limited cloud connectivity.

Data — device records, snapshot metadata, health readings, and stored images — is kept on the server with a **default retention period of seven days**, after which older information is automatically removed to manage storage growth.

## Relationship to other products

This repository contains **only the server backend**. It does not include:

- The operator or viewer mobile/web application
- An admin dashboard (though admin login is supported for a future or separate UI)
- Device firmware or hardware configuration tools

Those components are expected to exist as separate projects that connect to Parkiroid Server.

## One-sentence description

> Parkiroid Server is the backend hub for remote field device monitoring — it receives camera snapshots and health data from deployed devices, retains recent history, and enables live video viewing for operators who need real-time visibility.
