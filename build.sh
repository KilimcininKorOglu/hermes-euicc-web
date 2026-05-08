#!/bin/bash
# Hermes eUICC Web — Universal Multi-Platform Build Script
# Builds for: Linux (Desktop/Server/Embedded), Windows
# macOS and FreeBSD disabled (upstream driver issues)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_DIR="${SCRIPT_DIR}"
BINARY_NAME="hermes-euicc-web"
PKG_VERSION="0.1.0"
PKG_RELEASE=$(git rev-list --count HEAD 2>/dev/null || echo "1")
BUILD_DIR="${SCRIPT_DIR}/build/${PKG_VERSION}-${PKG_RELEASE}"
GO_VERSION="1.24.0"

# Package metadata
PKG_NAME="hermes-euicc-web"
PKG_LICENSE="MIT"
PKG_MAINTAINER="Kilimcinin Kör Oğlu <k@keremgok.tr>"

echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║        Hermes eUICC Web — Multi-Platform Build            ║${NC}"
echo -e "${CYAN}║        Version: ${PKG_VERSION}-${PKG_RELEASE}                                    ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go ${GO_VERSION}+${NC}"
    exit 1
fi

GO_CURRENT=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' || go version | sed 's/.*go\([0-9]*\.[0-9]*\).*/\1/')
echo -e "${GREEN}Using Go ${GO_CURRENT}${NC}"

mkdir -p "${BUILD_DIR}"

SUCCESS_COUNT=0
FAIL_COUNT=0

build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local GOARM=$3
    local OUTPUT_NAME=$4
    local DESCRIPTION=$5
    local OPT_FLAGS=$6
    local BUILD_TAGS=$7

    echo -e "${BLUE}Building: ${DESCRIPTION}${NC}"

    local BUILD_CMD="CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH}"
    [ -n "$GOARM" ] && BUILD_CMD="${BUILD_CMD} GOARM=${GOARM}"
    [ -n "$OPT_FLAGS" ] && BUILD_CMD="${BUILD_CMD} ${OPT_FLAGS}"

    local LDFLAGS="-s -w -X main.version=${PKG_VERSION}"
    local TAGS_FLAG=""
    [ -n "$BUILD_TAGS" ] && TAGS_FLAG="-tags=${BUILD_TAGS}"

    local OUTPUT_PATH="${BUILD_DIR}/${OUTPUT_NAME}"

    if eval ${BUILD_CMD} go build -ldflags "'${LDFLAGS}'" ${TAGS_FLAG} -o "'${OUTPUT_PATH}'" "'${SOURCE_DIR}'" 2>&1; then
        local SIZE=$(du -sh "${OUTPUT_PATH}" | cut -f1)
        echo -e "  ${GREEN}OK${NC} ${OUTPUT_NAME} (${SIZE})"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        echo -e "  ${RED}FAIL${NC} ${OUTPUT_NAME}"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

# =============================================================================
# Linux Desktop/Server Builds
# =============================================================================
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Linux Platforms                         ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

build_platform "linux" "amd64" "" "${BINARY_NAME}-linux-amd64" \
    "Linux x86-64 (64-bit)" \
    "GOAMD64=v2"

build_platform "linux" "386" "" "${BINARY_NAME}-linux-i386" \
    "Linux x86 (32-bit)" \
    ""

build_platform "linux" "arm64" "" "${BINARY_NAME}-linux-arm64" \
    "Linux ARM64 (Raspberry Pi 4+, modern ARM)" \
    ""

# =============================================================================
# OpenWRT/Embedded Linux Builds
# =============================================================================
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║               OpenWRT/Embedded Platforms                   ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# MIPS
build_platform "linux" "mips" "" "${BINARY_NAME}-openwrt-mips" \
    "OpenWRT MIPS Big-Endian (TP-Link, many routers)" \
    "GOMIPS=softfloat" "openwrt"

build_platform "linux" "mipsle" "" "${BINARY_NAME}-openwrt-mipsle" \
    "OpenWRT MIPS Little-Endian (Ramips/MT7621)" \
    "GOMIPS=softfloat" "openwrt"

build_platform "linux" "mips64" "" "${BINARY_NAME}-openwrt-mips64" \
    "OpenWRT MIPS64 Big-Endian" \
    "GOMIPS64=softfloat" "openwrt"

# ARM
build_platform "linux" "arm" "5" "${BINARY_NAME}-openwrt-arm_v5" \
    "OpenWRT ARMv5 (Kirkwood, older NAS)" \
    "" "openwrt"

build_platform "linux" "arm" "6" "${BINARY_NAME}-openwrt-arm_v6" \
    "OpenWRT ARMv6 (Raspberry Pi 1/Zero)" \
    "" "openwrt"

build_platform "linux" "arm" "7" "${BINARY_NAME}-openwrt-arm_v7" \
    "OpenWRT ARMv7 (Cortex-A7/A9, many modern routers)" \
    "" "openwrt"

build_platform "linux" "arm64" "" "${BINARY_NAME}-openwrt-arm64" \
    "OpenWRT ARM64 (RPi 4, NanoPi R4S)" \
    "" "openwrt"

# x86
build_platform "linux" "386" "" "${BINARY_NAME}-openwrt-x86" \
    "OpenWRT x86 (32-bit VMs, old PCs)" \
    "" "openwrt"

build_platform "linux" "amd64" "" "${BINARY_NAME}-openwrt-x86_64" \
    "OpenWRT x86-64 (VMs, modern PCs)" \
    "GOAMD64=v2" "openwrt"

# =============================================================================
# Windows Builds
# =============================================================================
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Windows Platforms                       ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

build_platform "windows" "amd64" "" "${BINARY_NAME}-windows-amd64.exe" \
    "Windows x86-64 (64-bit)" \
    "GOAMD64=v2"

build_platform "windows" "386" "" "${BINARY_NAME}-windows-i386.exe" \
    "Windows x86 (32-bit)" \
    ""

build_platform "windows" "arm64" "" "${BINARY_NAME}-windows-arm64.exe" \
    "Windows ARM64" \
    ""

# =============================================================================
# IPK Packages for OpenWRT
# =============================================================================
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              OpenWRT IPK Packages                          ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

create_ipk() {
    local ARCH=$1
    local BINARY_PATH=$2
    local ARCH_DESC=$3

    if [ ! -f "$BINARY_PATH" ]; then
        echo -e "${YELLOW}Skipping ${ARCH}: binary not found${NC}"
        return 1
    fi

    local IPK_BUILD_DIR="${BUILD_DIR}/ipk-${ARCH}"
    local IPK_CONTROL_DIR="${IPK_BUILD_DIR}/CONTROL"
    local IPK_DATA_DIR="${IPK_BUILD_DIR}/data"
    local IPK_FILE="${BUILD_DIR}/${PKG_NAME}_${PKG_VERSION}-${PKG_RELEASE}_${ARCH}.ipk"

    rm -rf "${IPK_BUILD_DIR}"
    mkdir -p "${IPK_CONTROL_DIR}" "${IPK_DATA_DIR}/usr/bin"

    cp "${BINARY_PATH}" "${IPK_DATA_DIR}/usr/bin/${BINARY_NAME}"
    chmod 755 "${IPK_DATA_DIR}/usr/bin/${BINARY_NAME}"

    cat > "${IPK_CONTROL_DIR}/control" << EOF
Package: ${PKG_NAME}
Version: ${PKG_VERSION}-${PKG_RELEASE}
Depends: libc
Section: utils
Architecture: ${ARCH}
Installed-Size: $(du -sb "${IPK_DATA_DIR}" | cut -f1)
Maintainer: ${PKG_MAINTAINER}
Description: eSIM profile management web interface
 Hermes eUICC Web provides a standalone web UI for managing
 eSIM profiles on eUICC-enabled devices via hermes-euicc CLI.
EOF

    cd "${IPK_BUILD_DIR}"
    tar czf "${IPK_BUILD_DIR}/control.tar.gz" -C "${IPK_CONTROL_DIR}" .
    tar czf "${IPK_BUILD_DIR}/data.tar.gz" -C "${IPK_DATA_DIR}" .
    echo "2.0" > "${IPK_BUILD_DIR}/debian-binary"
    tar czf "${IPK_FILE}" ./debian-binary ./control.tar.gz ./data.tar.gz
    cd "${SCRIPT_DIR}"

    rm -rf "${IPK_BUILD_DIR}"
    echo -e "  ${GREEN}IPK${NC} ${ARCH} ($(du -sh "${IPK_FILE}" | cut -f1))"
}

# MIPS
create_ipk "mips_24kc" "${BUILD_DIR}/${BINARY_NAME}-openwrt-mips" "MIPS 24Kc"
create_ipk "mipsel_24kc" "${BUILD_DIR}/${BINARY_NAME}-openwrt-mipsle" "MIPS-LE 24Kc"
create_ipk "mips64_octeonplus" "${BUILD_DIR}/${BINARY_NAME}-openwrt-mips64" "MIPS64 Octeon+"

# ARM
create_ipk "arm_arm1176jzf-s_vfp" "${BUILD_DIR}/${BINARY_NAME}-openwrt-arm_v6" "ARMv6"
create_ipk "arm_cortex-a7" "${BUILD_DIR}/${BINARY_NAME}-openwrt-arm_v7" "ARM Cortex-A7"
create_ipk "arm_cortex-a7_neon-vfpv4" "${BUILD_DIR}/${BINARY_NAME}-openwrt-arm_v7" "ARM Cortex-A7 NEON"
create_ipk "arm_cortex-a9" "${BUILD_DIR}/${BINARY_NAME}-openwrt-arm_v7" "ARM Cortex-A9"
create_ipk "aarch64_generic" "${BUILD_DIR}/${BINARY_NAME}-openwrt-arm64" "ARM64 Generic"
create_ipk "aarch64_cortex-a53" "${BUILD_DIR}/${BINARY_NAME}-openwrt-arm64" "ARM64 Cortex-A53"

# x86
create_ipk "i386_pentium-mmx" "${BUILD_DIR}/${BINARY_NAME}-openwrt-x86" "x86 32-bit"
create_ipk "x86_64" "${BUILD_DIR}/${BINARY_NAME}-openwrt-x86_64" "x86-64"

# =============================================================================
# Checksums
# =============================================================================
echo ""
cd "${BUILD_DIR}"
sha256sum ${BINARY_NAME}-* ${PKG_NAME}_*.ipk 2>/dev/null > SHA256SUMS
cd "${SCRIPT_DIR}"

# =============================================================================
# Summary
# =============================================================================
echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                      Build Summary                        ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${GREEN}Successful:  ${SUCCESS_COUNT}${NC}"
[ $FAIL_COUNT -gt 0 ] && echo -e "  ${RED}Failed:      ${FAIL_COUNT}${NC}"
echo -e "  ${BLUE}Output:      ${BUILD_DIR}/${NC}"

TOTAL_SIZE=$(du -sh "${BUILD_DIR}" | cut -f1)
echo -e "  ${BLUE}Total size:  ${TOTAL_SIZE}${NC}"

BIN_COUNT=$(ls -1 ${BUILD_DIR}/${BINARY_NAME}-* 2>/dev/null | wc -l)
echo -e "  ${GREEN}Binaries:    ${BIN_COUNT}${NC}"

IPK_COUNT=$(ls -1 ${BUILD_DIR}/${PKG_NAME}_*.ipk 2>/dev/null | wc -l)
[ $IPK_COUNT -gt 0 ] && echo -e "  ${GREEN}IPK packages: ${IPK_COUNT}${NC}"
echo ""
