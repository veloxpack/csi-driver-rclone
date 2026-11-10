# Release target to create a new semantic version tag.
# Usage: make release BUMP=patch|minor|major
#
# release-check runs first so a bad BUMP, a dirty tree or a stale branch fails
# before the slow RELEASE_DEPS. The clean-tree check runs again afterwards,
# because fmt and readme in RELEASE_DEPS can rewrite files.
.PHONY: release release-check
release: release-check $(RELEASE_DEPS)
    # 1. RELEASE_DEPS must not have changed anything
	@$(MAKE) --no-print-directory release-check-clean

    # 2. Compute the new tag from the latest one (or v0.0.0). Go ignores a
    # v2+ tag unless the module path in go.mod ends in /vN, so refuse that.
	@CURRENT_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	CURRENT_VERSION=$$(echo $$CURRENT_TAG | sed 's/^v//'); \
	MAJOR=$$(echo $$CURRENT_VERSION | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT_VERSION | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT_VERSION | cut -d. -f3); \
	\
	if [ "$(BUMP)" = "patch" ]; then \
		PATCH=$$((PATCH + 1)); \
	elif [ "$(BUMP)" = "minor" ]; then \
		MINOR=$$((MINOR + 1)); \
		PATCH=0; \
	elif [ "$(BUMP)" = "major" ]; then \
		MAJOR=$$((MAJOR + 1)); \
		MINOR=0; \
		PATCH=0; \
	fi; \
	\
	NEW_TAG="v$${MAJOR}.$${MINOR}.$${PATCH}"; \
	\
	MODULE=$$(go list -m); \
	if [ "$$MAJOR" -ge 2 ] && [ "$${MODULE##*/}" != "v$$MAJOR" ]; then \
		echo "Error: $$NEW_TAG needs the module path in go.mod to end in /v$$MAJOR (it is $$MODULE)."; \
		exit 1; \
	fi; \
	\
	echo "Current version: $$CURRENT_TAG"; \
	echo "Creating new version: $$NEW_TAG"; \
	git tag $$NEW_TAG; \
	git push origin $$NEW_TAG;

# release-check validates BUMP and that HEAD is a clean, up-to-date main.
release-check:
	@if [ "$(BUMP)" != "patch" ] && [ "$(BUMP)" != "minor" ] && [ "$(BUMP)" != "major" ]; then \
		echo "Error: BUMP must be patch, minor or major. Usage: make release BUMP=patch|minor|major"; \
		exit 1; \
	fi
	@BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$BRANCH" != "main" ]; then \
		echo "Error: Releases are tagged from main, but HEAD is on $$BRANCH."; \
		exit 1; \
	fi
	@git fetch --quiet origin main
	@if [ "$$(git rev-parse HEAD)" != "$$(git rev-parse origin/main)" ]; then \
		echo "Error: main does not match origin/main. Pull or push before releasing."; \
		exit 1; \
	fi
	@$(MAKE) --no-print-directory release-check-clean

# release-check-clean fails on any staged, unstaged or untracked change.
.PHONY: release-check-clean
release-check-clean:
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Working directory is not clean. Commit or stash changes before releasing."; \
		git status --short; \
		exit 1; \
	fi
