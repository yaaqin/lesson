import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Build jadi server mandiri (.next/standalone) biar image Docker gak perlu
  // bawa node_modules penuh -- lihat Dockerfile.
  output: "standalone",
  async headers() {
    return [
      {
        // Service worker PWA: jangan pernah di-cache biar update langsung kepake.
        source: "/sw.js",
        headers: [
          { key: "Content-Type", value: "application/javascript; charset=utf-8" },
          { key: "Cache-Control", value: "no-cache, no-store, must-revalidate" },
          { key: "Content-Security-Policy", value: "default-src 'self'; script-src 'self'" },
        ],
      },
    ];
  },
};

export default nextConfig;
