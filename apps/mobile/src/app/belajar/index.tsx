import { router } from "expo-router";
import { ActivityIndicator, Pressable, StyleSheet, Text, View, useColorScheme } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { Colors } from "@/constants/theme";
import { Palette } from "@/constants/palette";
import { useAuthStore } from "@/store/auth-store";
import { useLogoutMutation } from "@/hooks/use-auth";
import { useTiersQuery } from "@/hooks/use-curriculum";

const TIER_ICON: Record<string, string> = { sd: "➕", smp: "📐", smk: "📊" };

export default function TierPickerScreen() {
  const scheme = useColorScheme();
  const theme = Colors[scheme === "dark" ? "dark" : "light"];
  const session = useAuthStore((s) => s.session);
  const logoutMutation = useLogoutMutation();
  const tiersQuery = useTiersQuery();

  const handleLogout = async () => {
    await logoutMutation.mutateAsync();
    router.replace("/login");
  };

  return (
    <SafeAreaView style={[styles.safeArea, { backgroundColor: theme.background }]} edges={["bottom"]}>
      <View style={styles.header}>
        <Text style={[styles.greeting, { color: theme.textSecondary }]}>
          Halo, {session?.displayName ?? "kamu"} 👋
        </Text>
        <Pressable onPress={handleLogout} style={[styles.logoutBtn, { borderColor: scheme === "dark" ? Palette.borderDark : Palette.border }]}>
          <Text style={{ color: theme.textSecondary, fontSize: 13, fontWeight: "600" }}>Keluar</Text>
        </Pressable>
      </View>

      <Text style={[styles.title, { color: theme.text }]}>Pilih Jenjang</Text>
      <Text style={[styles.subtitle, { color: theme.textSecondary }]}>
        Mau latihan yang mana hari ini?
      </Text>

      {tiersQuery.isLoading && <ActivityIndicator style={{ marginTop: 24 }} />}

      <View style={styles.grid}>
        {tiersQuery.data?.map((tier) => (
          <Pressable
            key={tier.id}
            onPress={() => router.push(`/belajar/${tier.code}`)}
            style={[
              styles.card,
              {
                borderColor: scheme === "dark" ? Palette.borderDark : Palette.border,
                backgroundColor: theme.backgroundElement,
              },
            ]}
          >
            <Text style={styles.cardIcon}>{TIER_ICON[tier.code] ?? "🔢"}</Text>
            <Text style={[styles.cardLabel, { color: theme.text }]}>{tier.name}</Text>
          </Pressable>
        ))}
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: { flex: 1, padding: 20 },
  header: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", marginBottom: 12 },
  greeting: { fontSize: 13, fontWeight: "600" },
  logoutBtn: { borderWidth: 1, borderRadius: 999, paddingHorizontal: 14, paddingVertical: 6 },
  title: { fontSize: 24, fontWeight: "700" },
  subtitle: { fontSize: 14, marginTop: 2, marginBottom: 20 },
  grid: { flexDirection: "row", flexWrap: "wrap", gap: 12 },
  card: {
    flexGrow: 1,
    flexBasis: "30%",
    borderWidth: 1,
    borderRadius: 18,
    paddingVertical: 28,
    alignItems: "center",
    gap: 8,
  },
  cardIcon: { fontSize: 28 },
  cardLabel: { fontSize: 15, fontWeight: "700" },
});
