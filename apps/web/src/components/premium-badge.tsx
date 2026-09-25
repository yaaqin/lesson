// Tag spesial user premium (dikasih admin: bikin room multiplayer tanpa
// batas) -- dipasang di sebelah nickname biar keliatan beda dari user lain.
export function PremiumBadge({ size = "sm" }: { size?: "xs" | "sm" }) {
  return (
    <span
      title="Premium"
      className={`inline-flex shrink-0 items-center gap-1 rounded-full bg-gradient-to-r from-amber-400 to-orange-500 font-semibold text-white ${
        size === "xs" ? "px-1.5 py-px text-[10px]" : "px-2.5 py-0.5 text-xs"
      }`}
    >
      👑 Premium
    </span>
  );
}
