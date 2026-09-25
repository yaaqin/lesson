// "m:ss" atau "h:mm:ss".
export function formatDuration(totalSeconds: number) {
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  const mmss = `${m.toString().padStart(h > 0 ? 2 : 1, "0")}:${s.toString().padStart(2, "0")}`;
  return h > 0 ? `${h}:${mmss}` : mmss;
}

// Persentase waktu Adventure (waktu dipakai / waktu disediain), 1 desimal.
export function formatTimePercent(percent: number) {
  return `${percent.toLocaleString("id-ID", { maximumFractionDigits: 1 })}%`;
}
