import type { MetadataRoute } from "next";

// Web App Manifest (PWA) -- bikin MathQuest bisa di-install ke layar utama
// Android/iOS. Diserve Next di /manifest.webmanifest.
export default function manifest(): MetadataRoute.Manifest {
  return {
    id: "/",
    name: "MathQuest — Latihan Matematika Jadi Game",
    short_name: "MathQuest",
    description: "Latihan matematika berjenjang dari SD sampai kampus, dengan nyawa, waktu jawab, dan streak harian.",
    start_url: "/belajar",
    scope: "/",
    display: "standalone",
    orientation: "portrait",
    background_color: "#fafafa",
    theme_color: "#2563eb",
    lang: "id",
    categories: ["education", "games"],
    icons: [
      { src: "/icons/icon-192.png", sizes: "192x192", type: "image/png", purpose: "any" },
      { src: "/icons/icon-512.png", sizes: "512x512", type: "image/png", purpose: "any" },
      { src: "/icons/icon-maskable-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
    ],
  };
}
