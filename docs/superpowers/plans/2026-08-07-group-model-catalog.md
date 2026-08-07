# Group Model Catalog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the model plaza and image-generation workspace use enabled group model lists and built-in pricing without reading channel data.

**Architecture:** Keep the existing `ChannelService.ListPlazaGroups` boundary, but make its implementation group-centric: active groups provide model names and `PricingService` provides official base pricing. The model-plaza handler always filters exclusive groups; the image extension filters the same catalog by user authorization and `allow_image_generation`. The frontend loads the configured recharge multiplier and passes it into price formatting.

**Tech Stack:** Go, Gin, existing service/repository interfaces, embedded LiteLLM pricing, Vue 3, TypeScript, Vitest, Docker.

---

### Task 1: Build the group-centric plaza catalog

**Files:**
- Modify: `backend/internal/service/channel_plaza.go`
- Test: `backend/internal/service/channel_plaza_test.go`
- Reference: `backend/internal/service/channel_available.go` (`synthesizePricingFromLiteLLM`)

- [ ] **Step 1: Add failing service tests**

Add tests that construct active groups with `ModelsListConfig.Enabled=true` and model names while returning zero channels from the mock channel repository. Assert `ListPlazaGroups` returns the configured models, deduplicates/normalizes names, skips disabled or empty lists, preserves group metadata, and attaches built-in pricing when the pricing service knows the model.

- [ ] **Step 2: Run the focused tests and verify failure**

Run:

```bash
go test -tags=unit ./internal/service -run 'TestListPlazaGroups'
```

Expected: the new no-channel tests fail because the current implementation requires active channels.

- [ ] **Step 3: Implement group-centric catalog construction**

In `ListPlazaGroups`, load only active groups, skip groups with `ModelsListConfig.Enabled == false`, trim and deduplicate `ModelsListConfig.Models`, and create `PlazaModel` entries with the group platform. For each model, call `PricingService.GetModelPricing`, convert known pricing with `synthesizePricingFromLiteLLM`, apply existing group image-price shaping, and populate `OfficialPricing` from the same official pricing record. Do not call `s.repo.ListAll` or inspect channel associations in this method.

- [ ] **Step 4: Update existing channel-specific tests**

Replace assertions whose premise is channel-derived model membership with group-list assertions. Keep tests for group ordering, platform metadata, official pricing, and image-price precedence where they still apply; add explicit coverage for unknown models remaining visible with nil pricing.

- [ ] **Step 5: Run the service tests**

Run:

```bash
go test -tags=unit ./internal/service -run 'TestListPlazaGroups'
```

Expected: PASS.

- [ ] **Step 6: Commit the backend catalog change**

```bash
git add backend/internal/service/channel_plaza.go backend/internal/service/channel_plaza_test.go
git commit -m "feat(model-plaza): source catalog from group model lists"
```

### Task 2: Restrict the model plaza to public groups

**Files:**
- Modify: `backend/internal/handler/model_plaza_handler.go`
- Test: `backend/internal/handler/model_plaza_handler_test.go`

- [ ] **Step 1: Add failing visibility tests**

Add handler tests covering anonymous and authenticated requests with both public and exclusive `PlazaGroup` values. Both responses must contain only public groups; authentication must not re-enable exclusive groups.

- [ ] **Step 2: Run the focused tests and verify failure**

Run:

```bash
go test -tags=unit ./internal/handler -run 'TestModelPlaza'
```

Expected: authenticated visibility test fails against the current authorization-aware exclusive-group behavior.

- [ ] **Step 3: Implement public-only filtering**

Change `filterPlazaVisibleGroups` so every `IsExclusive` group is skipped regardless of auth state. Retain user-rate lookup only for public groups and leave model-plaza feature/authentication gates unchanged.

- [ ] **Step 4: Run handler tests**

Run:

```bash
go test -tags=unit ./internal/handler -run 'TestModelPlaza'
```

Expected: PASS.

- [ ] **Step 5: Commit the visibility change**

```bash
git add backend/internal/handler/model_plaza_handler.go backend/internal/handler/model_plaza_handler_test.go
git commit -m "fix(model-plaza): hide exclusive groups"
```

### Task 3: Make the image workspace use the group catalog and image permission

**Files:**
- Modify: `backend/internal/extensions/image-generation/service/service.go`
- Modify: `backend/internal/extensions/image-generation/service/config.go`
- Test: `backend/internal/extensions/image-generation/service/service_test.go`
- Test: `backend/internal/extensions/image-generation/service/config_test.go`

- [ ] **Step 1: Add failing image-option tests**

Add tests where the plaza provider returns no channel-backed data but returns group models from `models_list_config`. Assert `GetOptions` returns only active OpenAI groups with `AllowImageGeneration=true` and image model names; text-only or image-disabled groups are excluded. Add config-catalog tests for the same filter and for API keys belonging to the selected group.

- [ ] **Step 2: Run the focused tests and verify failure**

Run:

```bash
go test -tags=unit ./internal/extensions/image-generation/service -run 'TestImageGeneration'
```

Expected: no-channel image model tests fail because the current plaza map is empty or image groups are not derived from the configured list.

- [ ] **Step 3: Implement the minimal filtering changes**

Use the group-centric `PlazaGroup.Models` catalog, keep `GetAvailableGroups` as the authorization source, require `group.AllowImageGeneration`, and retain the existing `core.IsGPTImageGenerationModel` filter. Do not add any channel fallback or bypass API-key ownership/active-status checks.

- [ ] **Step 4: Run image-generation tests**

Run:

```bash
go test -tags=unit ./internal/extensions/image-generation/service -run 'TestImageGeneration'
```

Expected: PASS.

- [ ] **Step 5: Commit the image workspace change**

```bash
git add backend/internal/extensions/image-generation/service/service.go backend/internal/extensions/image-generation/service/config.go backend/internal/extensions/image-generation/service/service_test.go backend/internal/extensions/image-generation/service/config_test.go
git commit -m "feat(image-generation): use authorized group model lists"
```

### Task 4: Apply the configured recharge multiplier in model-plaza prices

**Files:**
- Modify: `frontend/src/features/model-plaza/api/modelPlaza.ts`
- Modify: `frontend/src/features/model-plaza/types/modelPlaza.ts`
- Modify: `frontend/src/features/model-plaza/utils/modelPlaza.ts`
- Modify: `frontend/src/features/model-plaza/components/PlazaExplorer.vue`
- Modify: `frontend/src/features/model-plaza/components/ModelPriceCard.vue`
- Test: `frontend/src/features/model-plaza/__tests__/modelPlaza.spec.ts`
- Test: `frontend/src/features/model-plaza/__tests__/ModelPriceCard.spec.ts`

- [ ] **Step 1: Add failing pricing tests**

Replace the hard-coded `USD_CREDIT_PER_CNY` assertions with tests passing explicit recharge multipliers. Verify token and image prices change when the multiplier changes and default to `1` when the payment endpoint returns an invalid/missing value. Add an API test that `loadModelPlaza` combines `/model-plaza` with `/payment/config`.

- [ ] **Step 2: Run frontend tests and verify failure**

Run:

```bash
pnpm --dir frontend exec vitest run src/features/model-plaza/__tests__/modelPlaza.spec.ts src/features/model-plaza/__tests__/ModelPriceCard.spec.ts
```

Expected: the explicit-multiplier tests fail because pricing currently divides by a hard-coded `10`.

- [ ] **Step 3: Implement dynamic pricing input**

Load `balance_recharge_multiplier` from the existing `/payment/config` endpoint, add it to the model-plaza response view model, pass it from `PlazaExplorer` to `ModelPriceCard`, and make `paidTokenPrice`/`paidRequestPrice` divide by the normalized multiplier. Remove the hard-coded `USD_CREDIT_PER_CNY` export while keeping existing group/user/image rate selection unchanged.

- [ ] **Step 4: Run frontend tests and type checks**

Run:

```bash
pnpm --dir frontend exec vitest run src/features/model-plaza/__tests__/modelPlaza.spec.ts src/features/model-plaza/__tests__/ModelPriceCard.spec.ts
pnpm --dir frontend exec vue-tsc -b
```

Expected: PASS with no type errors.

- [ ] **Step 5: Commit the frontend pricing change**

```bash
git add frontend/src/features/model-plaza
git commit -m "feat(model-plaza): use configured recharge multiplier"
```

### Task 5: Run regression checks and deploy the image

**Files:**
- Verify: all files changed in Tasks 1-4
- Deployment config outside source repo: `/data/sub2api/docker-compose.migrate.yml`
- Backup: `/data/sub2api/backups/docker-compose.migrate.yml.20260807.bak`

- [ ] **Step 1: Run backend regression tests**

Run:

```bash
go test -tags=unit ./internal/service ./internal/handler ./internal/extensions/image-generation/service
```

Expected: PASS.

- [ ] **Step 2: Verify no migration files changed**

Run:

```bash
git diff --name-only origin/custom/main...HEAD -- backend/migrations
```

Expected: no output.

- [ ] **Step 3: Build the release image**

Build the image from the updated `custom/main` source with the actual commit hash embedded in `COMMIT`:

```bash
COMMIT=$(git rev-parse --short=12 HEAD)
docker build --pull --tag "sub2api:custom-main-${COMMIT}" --build-arg COMMIT="${COMMIT}" --build-arg VERSION=custom-main --file Dockerfile .
```

Do not run database migration commands.

- [ ] **Step 4: Back up and update Compose**

Copy `/data/sub2api/docker-compose.migrate.yml` to a new timestamped file under `/data/sub2api/backups/`, update only the application image line, validate with `docker compose ... config -q`, and preserve the existing PostgreSQL/Redis definitions and mounts.

- [ ] **Step 5: Recreate only the application container**

Run:

```bash
docker compose -f /data/sub2api/docker-compose.migrate.yml up -d --no-deps --force-recreate --pull never sub2api
```

Expected: only `sub2api` is recreated; PostgreSQL and Redis container IDs remain unchanged.

- [ ] **Step 6: Verify deployment behavior**

Check the new container image/commit, `/health`, model-plaza response contains only public groups and configured models, image-generation config exposes only authorized image-enabled groups, and the database migration count/checksums remain unchanged.
