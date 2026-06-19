# Project Status

## Maturity overview

Parkiroid Server is in **early active development**. The project was started in **June 2026** and has progressed quickly from initial setup to a working backend with core monitoring and streaming capabilities.

| Dimension | Assessment |
|-----------|------------|
| **Age** | Approximately one week of active development at time of writing |
| **Core functionality** | Operational — snapshot upload, health reporting, live streaming, and authentication all work |
| **Documentation** | API contract and developer setup guides exist; this manager documentation is being added |
| **Automated testing** | Not yet in place |
| **Release pipeline** | No continuous integration or automated deployment configured in the repository |
| **Production readiness** | Functional, but requires manual security setup (passwords, access tokens) before production use |

---

## Development timeline

| Period | Milestone |
|--------|-----------|
| Early June 2026 | Project initialized |
| Mid June 2026 | API documentation established; health metrics and data retention added; containerized deployment set up |
| Mid June 2026 | Embedded access tokens added for client applications |

---

## What is done

The following core capabilities are implemented and usable:

- Receiving and storing camera snapshots from field devices
- Retrieving the latest snapshot information per device
- Receiving and storing device health metrics
- Retrieving the latest health reading per device
- Issuing access for live video streaming (device as source, viewers as audience)
- Two authentication paths: app/device tokens and administrator login
- Automatic device registration on first contact
- Seven-day data retention with automatic cleanup
- Self-hosted deployment packaging (server plus streaming component)
- Health check for operational monitoring

---

## What is likely next

Nothing below is formally committed or tracked as a roadmap. These are reasonable inferences based on current gaps:

| Priority area | Rationale |
|---------------|-----------|
| Expose device listing | The server already tracks devices; client apps need a way to discover them |
| Deliver images through the server | Operators should receive actual images, not only metadata |
| Admin web interface | Administrator login exists but no UI is bundled with the server |
| Historical data access | Browsing past snapshots and metrics over time |
| Multi-user administration | Team access, roles, and audit trails |
| Automated testing and deployment | Reduce regression risk as features grow |
| Alerting | Notify operators when devices go offline or cross health thresholds |

---

## Risks and limitations

These points matter for planning, budgeting, and go-live decisions:

### Product maturity

The system is young and evolving. Features may change, and production deployment requires deliberate security configuration — default development credentials must be replaced.

### Single administrator

Only one admin account is supported. There is no team-based access, role separation, or activity audit trail for administrative actions.

### Reactive, not proactive

The server stores and serves data when asked. It does not alert anyone when a device fails, battery drops, or connectivity degrades. Something must actively check device status.

### Limited history

Only the most recent snapshot and health reading per device are readily available. Trend analysis, compliance reporting, or incident reconstruction over time are not supported without additional systems.

### Dependence on external client apps

The server is a backend only. Day-to-day value depends on separate mobile or web applications that are not part of this repository. Without those apps, end users cannot interact with the system meaningfully.

### Scale considerations

The current design uses local file and database storage on a single server. This suits small to medium device fleets and self-hosted deployments. Very large fleets or high-availability requirements may need architectural evolution that is not yet planned.

### No automated quality gates

Without automated tests or a release pipeline, changes carry higher regression risk. This is acceptable for early development but should be addressed before wider production reliance.

---

## Production checklist (non-technical)

Before relying on Parkiroid Server in a production environment, ensure:

- [ ] Default administrator password has been changed
- [ ] Unique access tokens have been generated for client applications and devices
- [ ] Streaming service credentials have been set to production values
- [ ] Data retention period aligns with organizational policy
- [ ] Backup strategy exists for stored device data and images
- [ ] Uptime monitoring is configured using the health check
- [ ] Client applications are deployed and tested against the production server

---

## Suggested framing for stakeholders

**For executives:** Parkiroid Server is the infrastructure backbone for monitoring distributed field devices — camera snapshots, health telemetry, and live video — designed for self-hosted deployment. Core capabilities are in place; the product is early-stage and depends on companion client apps for end-user value.

**For project managers:** Plan for integration work (client apps, image delivery gap, device listing), security hardening before go-live, and future investments in alerting, history, and multi-user admin if those are on the product roadmap.

**For operations:** Expect a containerized deployment with persistent storage, a seven-day data window, and manual credential management. No built-in alerting — monitoring strategy must include active polling or external tooling.
