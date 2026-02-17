#!/bin/bash
# package-deb.sh - A simple script to create .deb packages

VERSION="0.1.0"
NAME="wireguard-tui"
ARCHS=("amd64" "arm64")

for ARCH in "${ARCHS[@]}"; do
    echo "Creating .deb for $ARCH..."
    
    # 1. Create directory structure
    PKG_DIR="dist/${NAME}_${VERSION}_${ARCH}"
    mkdir -p "$PKG_DIR/usr/bin"
    mkdir -p "$PKG_DIR/DEBIAN"
    
    # 2. Copy binary (must be built first)
    BINARY="dist/${NAME}-linux-$ARCH"
    if [ ! -f "$BINARY" ]; then
        echo "Binary $BINARY not found. Running make build-linux..."
        make build-linux
    fi
    cp "$BINARY" "$PKG_DIR/usr/bin/$NAME"
    chmod +x "$PKG_DIR/usr/bin/$NAME"
    
    # 3. Create control file
    cat <<EOF > "$PKG_DIR/DEBIAN/control"
Package: $NAME
Version: $VERSION
Architecture: $ARCH
Maintainer: lakecass and Gemini
Description: A modern htop-like TUI for WireGuard
Depends: wireguard-tools
Section: utils
Priority: optional
EOF

    # 4. Build package
    if command -v dpkg-deb &> /dev/null; then
        dpkg-deb --build "$PKG_DIR"
        echo "Generated: ${PKG_DIR}.deb"
    else
        echo "dpkg-deb not found. Package structure created in $PKG_DIR"
    fi
done
