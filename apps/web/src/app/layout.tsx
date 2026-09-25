import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";
import { THEME_INIT_SCRIPT } from "@/lib/theme";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "MathQuest — Latihan Matematika Jadi Game",
  description:
    "Latihan matematika berjenjang dari SD sampai umum, dengan nyawa, waktu jawab, dan streak harian.",
  applicationName: "MathQuest",
  // iOS: biar pas di-"Tambah ke Layar Utama" kebuka fullscreen kayak app.
  appleWebApp: {
    capable: true,
    title: "MathQuest",
    statusBarStyle: "default",
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#fafafa" },
    { media: "(prefers-color-scheme: dark)", color: "#000000" },
  ],
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="id"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
      // class "dark" dipasang script tema sebelum hydrate
      suppressHydrationWarning
    >
      <head>
        {/* Pasang tema sebelum render biar gak kedip terang -> gelap. */}
        <script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
      </head>
      <body className="min-h-full flex flex-col">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
