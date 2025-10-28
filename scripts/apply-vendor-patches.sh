#!/bin/bash
# Copyright 2025 Veloxpack.io
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
PATCHES_DIR="${PROJECT_ROOT}/patches"

echo "=========================================="
echo "Applying vendor patches..."
echo "=========================================="

# Check if patches directory exists
if [ ! -d "${PATCHES_DIR}" ]; then
    echo "No patches directory found at ${PATCHES_DIR}"
    exit 0
fi

# Count patches
PATCH_COUNT=$(find "${PATCHES_DIR}" -name "*.patch" -type f 2>/dev/null | wc -l | tr -d ' ')

if [ "${PATCH_COUNT}" -eq 0 ]; then
    echo "No patch files found in ${PATCHES_DIR}"
    exit 0
fi

echo "Found ${PATCH_COUNT} patch file(s) to apply"
echo ""

# Change to project root for applying patches
cd "${PROJECT_ROOT}"

# Apply each patch file
APPLIED=0
FAILED=0

for patch_file in "${PATCHES_DIR}"/*.patch; do
    if [ -f "${patch_file}" ]; then
        patch_name=$(basename "${patch_file}")
        echo "Applying patch: ${patch_name}"

        # Check if patch can be applied (dry-run)
        if patch -p0 --dry-run --forward --silent < "${patch_file}" > /dev/null 2>&1; then
            # Apply the patch
            if patch -p0 --forward < "${patch_file}"; then
                echo "✓ Successfully applied ${patch_name}"
                APPLIED=$((APPLIED + 1))
            else
                echo "✗ Failed to apply ${patch_name}"
                FAILED=$((FAILED + 1))
            fi
        elif patch -p0 --dry-run --reverse --silent < "${patch_file}" > /dev/null 2>&1; then
            echo "⊙ Patch ${patch_name} already applied (skipping)"
            APPLIED=$((APPLIED + 1))
        else
            echo "✗ Patch ${patch_name} cannot be applied (conflicts or missing files)"
            FAILED=$((FAILED + 1))
        fi
        echo ""
    fi
done

echo "=========================================="
echo "Patch application complete"
echo "Applied: ${APPLIED} | Failed: ${FAILED}"
echo "=========================================="

if [ "${FAILED}" -gt 0 ]; then
    echo ""
    echo "ERROR: ${FAILED} patch(es) failed to apply"
    echo "Please review the vendor directory and patches"
    exit 1
fi

exit 0

