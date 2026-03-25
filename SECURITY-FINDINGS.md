# Security Findings Report

**Repository:** aks-store-demo-janfabian  
**Date:** 2026-03-25  
**Scope:** Security gaps, missing tests, and hardcoded credentials

---

## Summary

| Severity | Count |
|----------|-------|
| 🔴 Critical | 4 |
| 🟠 High | 6 |
| 🟡 Medium | 8 |
| 🔵 Low | 4 |

---

## 🔴 Critical

### C-1 — Hardcoded Default Credentials in Kubernetes Manifests

**Files:**
- `aks-store-all-in-one.yaml` (lines 72–77)
- `aks-store-ingress-quickstart.yaml`
- `sample-manifests/docs/app-routing/aks-store-deployments-and-services.yaml`

**Description:**  
RabbitMQ secrets are stored as Base64-encoded values directly in YAML manifests committed to source control. Base64 is encoding, not encryption, and these values decode trivially.

```yaml
# aks-store-all-in-one.yaml
kind: Secret
data:
  RABBITMQ_DEFAULT_USER: dXNlcm5hbWU=   # "username"
  RABBITMQ_DEFAULT_PASS: cGFzc3dvcmQ=   # "password"
  ORDER_QUEUE_PASSWORD:  cGFzc3dvcmQ=   # "password"
```

Additionally, the sample manifest stores credentials as plaintext environment variables instead of referencing a Secret:

```yaml
- name: RABBITMQ_DEFAULT_PASS
  value: "password"
```

**Risk:** Any user with repository read access obtains queue credentials.  
**Recommendation:** Remove secrets from source control. Use a secret management solution such as Azure Key Vault with the External Secrets Operator, or Sealed Secrets.

---

### C-2 — Default Passwords in Helm Values File

**File:** `charts/aks-store-demo/values.yaml` (lines 34–35, 44–45)

```yaml
orderService:
  queueUsername: "username"
  queuePassword: "password"

makelineService:
  orderQueueUsername: "username"
  orderQueuePassword: "password"
```

**Risk:** Helm values are commonly committed to version control. Anyone who installs the chart without overriding these values deploys with known credentials.  
**Recommendation:** Set password fields to empty strings and require callers to supply values at install time via `--set` or a separate secrets file that is not committed.

---

### C-3 — Hardcoded Credentials in Docker Compose Files

**Files:**
- `docker-compose.yml`
- `src/order-service/docker-compose.yml`

```yaml
rabbitmq:
  environment:
    - "RABBITMQ_DEFAULT_USER=username"
    - "RABBITMQ_DEFAULT_PASS=password"
```

**Risk:** Developers running the local stack use well-known default credentials that are also applied in non-production environments.  
**Recommendation:** Replace inline values with variable references (e.g., `${RABBITMQ_DEFAULT_PASS}`) backed by a local `.env` file that is `.gitignore`d.

---

### C-4 — Security Context Disabled in Helm Chart

**File:** `charts/aks-store-demo/values.yaml` (lines 97–104)

```yaml
securityContext:
  {}
  # capabilities:
  #   drop:
  #   - ALL
  # readOnlyRootFilesystem: true
  # runAsNonRoot: true
  # runAsUser: 1000
```

All recommended security context options are commented out. Containers therefore run as root with full Linux capabilities and a writable root filesystem.

**Risk:** Container escape, privilege escalation if a vulnerability is exploited in any service.  
**Recommendation:** Enable `runAsNonRoot: true`, `readOnlyRootFilesystem: true`, and drop `ALL` capabilities.

---

## 🟠 High

### H-1 — No Authentication on Any API Endpoint

**Files:** `src/order-service/app.js`, `src/makeline-service/main.go`, `src/ai-service/main.py`

All HTTP endpoints (order submission, order fetch/update, AI description and image generation) are fully public with no API key, session token, or identity check.

**Risk:** Unauthorized actors can submit orders, modify order state, and invoke expensive Azure OpenAI / DALL-E endpoints at will.  
**Recommendation:** Enforce authentication (API key middleware as a minimum, Azure Workload Identity preferred) on all routes, especially the AI service endpoints which carry direct cost implications.

---

### H-2 — Wildcard CORS on All Services

**Files:** `src/order-service/app.js`, `src/makeline-service/main.go`, `src/ai-service/main.py`, `src/store-front/vite.config.ts`

```javascript
// order-service
fastify.register(require('@fastify/cors'), { origin: '*' })

// ai-service
app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])
```

**Risk:** Any origin can send credentialed requests to these services (CSRF vector). Combined with the absence of authentication (H-1) this is a significant exposure.  
**Recommendation:** Restrict `origin` to the known front-end hostnames for each deployment environment.

---

### H-3 — No Rate Limiting on Any Endpoint

**Files:** All services

No service applies rate limiting. The AI service is particularly exposed: an attacker can call `POST /generate/image` (DALL-E) in a tight loop, incurring unbounded Azure OpenAI costs.

**Risk:** Denial-of-service, cost exhaustion on pay-per-use AI APIs.  
**Recommendation:** Add rate-limiting middleware (e.g., `@fastify/rate-limit`, `golang.org/x/time/rate`, `slowapi` for FastAPI).

---

### H-4 — Internal Services Exposed via Docker Compose Port Bindings

**File:** `docker-compose.yml`

```yaml
rabbitmq:
  ports:
    - 15672:15672   # Management UI
    - 5672:5672     # AMQP

mongodb:
  ports:
    - 27017:27017   # MongoDB
```

**Risk:** RabbitMQ management UI and MongoDB are reachable from the host network with default credentials (see C-1/C-3). Any process or container on the host can connect directly to the data store.  
**Recommendation:** Remove external port bindings for backing services; access them only through the Docker Compose internal network.

---

### H-5 — No Kubernetes Network Policies

**Files:** `charts/`, `kustomize/`

No `NetworkPolicy` objects are defined. Every pod can reach every other pod on any port.

**Risk:** Lateral movement if any single service is compromised. A compromised order-service pod can directly contact the MongoDB or RabbitMQ pods without restriction.  
**Recommendation:** Add NetworkPolicy resources that allow only the documented service-to-service communication paths.

---

### H-6 — Plain HTTP Used in Default Service URLs

**File:** `src/store-front/vite.config.ts`, `src/store-admin/vite.config.ts`

```typescript
const PRODUCT_SERVICE_URL = env.VITE_PRODUCT_SERVICE_URL || "http://localhost:3002/"
const ORDER_SERVICE_URL   = env.VITE_ORDER_SERVICE_URL   || "http://localhost:3000/"
```

**Risk:** Order and product data (including customer information in order payloads) is transmitted in plaintext over the network when these defaults are used in non-localhost environments.  
**Recommendation:** Default URLs should use HTTPS; enforce TLS at the Kubernetes Ingress level and ensure all inter-service traffic uses HTTPS or mTLS.

---

## 🟡 Medium

### M-1 — No Input Validation on Order Submission

**File:** `src/order-service/routes/root.js`

```javascript
fastify.post('/', async function (request, reply) {
  const msg = request.body   // No schema, no size limit
  fastify.sendMessage(Buffer.from(JSON.stringify(msg)))
  reply.code(201)
})
```

**Risk:** Arbitrary payloads (including very large ones) are forwarded to the message queue without validation. This can lead to message-queue pollution, unexpected consumer behavior, and potential memory exhaustion.  
**Recommendation:** Define a Fastify JSON schema for the route body; add `bodyLimit` to the server configuration.

---

### M-2 — Internal Error Details Exposed to Clients

**File:** `src/ai-service/routers/description_generator.py` and `image_generator.py`

```python
except Exception as e:
    raise HTTPException(status_code=500,
        detail=f"Error generating description: {str(e)}")
```

**Risk:** Stack traces and internal exception messages are returned to callers, leaking implementation details (library names, file paths, configuration hints).  
**Recommendation:** Log the full exception server-side; return a generic error message to the client.

---

### M-3 — `.env` File Loaded with `override=True` at Runtime

**File:** `src/ai-service/routers/description_generator.py`

```python
load_dotenv(dotenv_path=env_path, override=True)
```

**Risk:** If a `.env` file is present in the container image or mounted into the container, it silently overrides environment variables set by the Kubernetes Secret or ConfigMap, potentially reverting to insecure defaults.  
**Recommendation:** Remove `override=True`; document that the `.env` file is for local development only and must not be included in production images.

---

### M-4 — MongoDB Runs Without Authentication in Local Environment

**File:** `docker-compose.yml`

The MongoDB service has no `MONGO_INITDB_ROOT_USERNAME` / `MONGO_INITDB_ROOT_PASSWORD` set and its port is bound to the host (see H-4).

**Risk:** Anyone who can reach port 27017 on the host has unrestricted read/write access to all order data.  
**Recommendation:** Set MongoDB authentication credentials (sourced from a `.env` file) and restrict the port binding.

---

### M-5 — No Resource Limits Defined for Containers

**File:** `charts/aks-store-demo/values.yaml`

```yaml
resources:
  {}
  # limits:
  #   cpu: 100m
  #   memory: 128Mi
```

Resource limits are commented out for all deployments.

**Risk:** A misbehaving or compromised container can consume all node CPU and memory, causing a cluster-wide denial of service.  
**Recommendation:** Define realistic CPU and memory requests/limits for each service.

---

### M-6 — No Image Pull Policy Set in Kubernetes Manifests

**Files:** `kustomize/`, `charts/`, `aks-store-all-in-one.yaml`

Most deployment specs do not set `imagePullPolicy: Always`.

**Risk:** Nodes may serve stale cached images, potentially running unpatched versions.  
**Recommendation:** Set `imagePullPolicy: Always` for non-`latest`-pinned images, or use digest-pinned image references.

---

### M-7 — No Security Event Logging or Audit Trail

**Files:** All services

No service logs authentication failures, unexpected input patterns, or state-change events in a structured way that would support incident detection or forensic analysis.

**Risk:** Attacks go undetected; post-incident investigation is not possible.  
**Recommendation:** Add structured logging for all state-changing operations and errors, and ship logs to a centralized observability platform.

---

### M-8 — Missing Unit Tests for Five of Eight Services

See section **Missing Tests** below for full detail.

---

## 🔵 Low

### L-1 — AI Service `.env.example` Contains Internal Endpoint Pattern

**File:** `src/ai-service/.env.example`

```
AI_ENDPOINT=http://<A_REACHABLE_IP>/chat
```

**Risk:** Minimal, but the pattern normalizes unencrypted HTTP for an AI endpoint and the placeholder format may confuse operators into using a real IP here.  
**Recommendation:** Update the example to use HTTPS and a hostname rather than an IP address.

---

### L-2 — Debian-Based Runtime Images Include Package Manager

**Files:** `src/product-service/Dockerfile`, `src/virtual-customer/Dockerfile`, `src/virtual-worker/Dockerfile`

These services use `debian:bookworm-slim` as the runtime base image, which retains `apt-get`.

**Risk:** Expanded attack surface; a compromised container can install additional tools.  
**Recommendation:** Migrate to a minimal distroless base image (e.g., `gcr.io/distroless/cc-debian12`) for Rust binaries.

---

### L-3 — API Has No Versioning Strategy

**Files:** All services

All routes are unversioned (e.g., `GET /order/fetch` with no `/v1/` prefix).

**Risk:** No ability to introduce breaking changes safely; client compatibility cannot be maintained during migrations.  
**Recommendation:** Prefix all API routes with a version segment (e.g., `/v1/`).

---

### L-4 — No Automated Dependency Vulnerability Scanning in CI

**Files:** `.github/workflows/`

The CI workflows build and push images but do not run a software composition analysis (SCA) tool such as Trivy, Snyk, or GitHub Dependabot on container images or package manifests.

**Risk:** Known-vulnerable transitive dependencies remain undetected until a security researcher or attacker finds them.  
**Recommendation:** Add a Trivy or `anchore/scan-action` step to the image-build workflow; enable GitHub Dependabot alerts.

---

## Missing Tests

The table below shows which services have unit tests and which do not.

| Service | Language | Unit Tests | Notes |
|---------|----------|:----------:|-------|
| `order-service` | Node.js | ✅ | 3 test files, basic happy-path coverage |
| `store-front` | Vue.js | ✅ | Component + Playwright E2E |
| `store-admin` | Vue.js | ✅ | Component + Playwright E2E |
| `makeline-service` | Go | ❌ | No test files found |
| `product-service` | Rust | ❌ | No test files found |
| `ai-service` | Python | ❌ | No test files found |
| `virtual-customer` | Rust | ❌ | No test files found |
| `virtual-worker` | Rust | ❌ | No test files found |

**Additional test gaps across all services:**
- No authentication/authorization tests
- No CORS validation tests
- No negative-path / error-handling tests
- No input-validation boundary tests (oversized payloads, malformed JSON)
- No contract tests between services
- No load or rate-limit tests for the AI service

---

## Hardcoded Credentials — Consolidated List

| File | Credential | Value | Encoding |
|------|-----------|-------|----------|
| `aks-store-all-in-one.yaml` | `RABBITMQ_DEFAULT_USER` | `username` | Base64 |
| `aks-store-all-in-one.yaml` | `RABBITMQ_DEFAULT_PASS` | `password` | Base64 |
| `aks-store-all-in-one.yaml` | `ORDER_QUEUE_PASSWORD` | `password` | Base64 |
| `aks-store-ingress-quickstart.yaml` | `RABBITMQ_DEFAULT_PASS` | `password` | Base64 |
| `charts/aks-store-demo/values.yaml` | `queuePassword` | `password` | Plaintext |
| `charts/aks-store-demo/values.yaml` | `orderQueuePassword` | `password` | Plaintext |
| `docker-compose.yml` | `RABBITMQ_DEFAULT_PASS` | `password` | Plaintext |
| `src/order-service/docker-compose.yml` | `RABBITMQ_DEFAULT_PASS` | `password` | Plaintext |
| `sample-manifests/.../aks-store-deployments-and-services.yaml` | `RABBITMQ_DEFAULT_PASS` | `password` | Plaintext |
