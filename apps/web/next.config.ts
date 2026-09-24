import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Build jadi server mandiri (.next/standalone) biar image Docker gak perlu
  // bawa node_modules penuh -- lihat Dockerfile.
  output: "standalone",
};

export default nextConfig;
