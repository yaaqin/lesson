import { router, useLocalSearchParams } from "expo-router";
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View, useColorScheme } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { Colors } from "@/constants/theme";
import { Palette } from "@/constants/palette";
import {
  useBatchesQuery,
  useCategoriesQuery,
  useChallengesQuery,
  useMeQuery,
  useTiersQuery,
} from "@/hooks/use-curriculum";

export default function TierBatchScreen() {
  const scheme = useColorScheme();
  const theme = Colors[scheme === "dark" ? "dark" : "light"];
  const { tierCode } = useLocalSearchParams<{ tierCode: string }>();

  const meQuery = useMeQuery();
  const tiersQuery = useTiersQuery();
  const tier = tiersQuery.data?.find((t) => t.code === tierCode);
  const usesBatch = tier?.usesBatch ?? true;

  const batchesQuery = useBatchesQuery(tierCode);
  const batch = batchesQuery.data?.[0];
  const challengesQuery = useChallengesQuery(usesBatch ? batch?.id : undefined);

  const categoriesQuery = useCategoriesQuery(!usesBatch ? tierCode : "");

  const lives = meQuery.data?.livesRemaining ?? 0;
  const streak = meQuery.data?.currentStreak ?? 0;

  return (
    <SafeAreaView style={[styles.safeArea, { backgroundColor: theme.background }]} edges={["bottom"]}>
      <View style={styles.header}>
        <View style={{ flexDirection: "row", gap: 2 }}>
          {[0, 1, 2].map((i) => (
            <Text key={i} style={{ fontSize: 16, opacity: i < lives ? 1 : 0.25 }}>
              ❤️
            </Text>
          ))}
        </View>
        <Text style={{ color: Palette.amber, fontWeight: "700" }}>🔥 {streak}</Text>
      </View>

      <Text style={[styles.title, { color: theme.text }]}>
        {tier?.name ?? tierCode?.toUpperCase()} {usesBatch && batch ? `· ${batch.name}` : ""}
      </Text>

      {usesBatch ? (
        <>
          {challengesQuery.isLoading && <ActivityIndicator style={{ marginTop: 24 }} />}

          <ScrollView contentContainerStyle={styles.list}>
            {challengesQuery.data?.map((challenge, index) => (
              <Pressable
                key={challenge.id}
                onPress={() => router.push(`/belajar/${tierCode}/${challenge.id}`)}
                style={[
                  styles.card,
                  {
                    borderColor: scheme === "dark" ? Palette.borderDark : Palette.border,
                    backgroundColor: theme.backgroundElement,
                  },
                ]}
              >
                <View
                  style={[
                    styles.badgeCircle,
                    { backgroundColor: challenge.isExam ? Palette.amber : Palette.blue },
                  ]}
                >
                  <Text style={styles.badgeCircleText}>{challenge.isExam ? "🏁" : index + 1}</Text>
                </View>

                <View style={{ flex: 1 }}>
                  <View style={{ flexDirection: "row", alignItems: "center", gap: 6, flexWrap: "wrap" }}>
                    <Text style={[styles.cardTitle, { color: theme.text }]}>{challenge.name}</Text>
                    {challenge.hasEssay && (
                      <View style={[styles.pill, { backgroundColor: Palette.purpleBg }]}>
                        <Text style={{ color: Palette.purple, fontSize: 10, fontWeight: "700" }}>
                          ✏️ Ada Essay
                        </Text>
                      </View>
                    )}
                  </View>
                  <Text style={[styles.cardMeta, { color: theme.textSecondary }]}>
                    {challenge.questionCountRequired} soal · lulus min {challenge.passThresholdPercent}% ·{" "}
                    {challenge.isExam
                      ? `${challenge.timeLimitSeconds}s total`
                      : `${challenge.timeLimitSeconds}s/soal`}
                  </Text>
                </View>

                {challenge.completed ? (
                  <View style={[styles.pill, { backgroundColor: Palette.greenBg }]}>
                    <Text style={{ color: Palette.green, fontSize: 11, fontWeight: "700" }}>Selesai ✓</Text>
                  </View>
                ) : (
                  <Text style={{ color: Palette.blue, fontSize: 13, fontWeight: "700" }}>Mulai →</Text>
                )}
              </Pressable>
            ))}
          </ScrollView>
        </>
      ) : (
        <>
          {categoriesQuery.isLoading && <ActivityIndicator style={{ marginTop: 24 }} />}

          <ScrollView contentContainerStyle={styles.list}>
            {categoriesQuery.data?.map((category) => (
              <Pressable
                key={category.id}
                onPress={() => router.push(`/belajar/${tierCode}/kategori/${category.id}`)}
                style={[
                  styles.card,
                  {
                    borderColor: scheme === "dark" ? Palette.borderDark : Palette.border,
                    backgroundColor: theme.backgroundElement,
                  },
                ]}
              >
                <View style={{ flex: 1 }}>
                  <Text style={[styles.cardTitle, { color: theme.text }]}>{category.name}</Text>
                </View>
                <Text style={{ color: Palette.blue, fontSize: 13, fontWeight: "700" }}>Lihat level →</Text>
              </Pressable>
            ))}
          </ScrollView>
        </>
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: { flex: 1, paddingHorizontal: 20, paddingTop: 16 },
  header: { flexDirection: "row", justifyContent: "flex-end", alignItems: "center", gap: 16, marginBottom: 12 },
  title: { fontSize: 20, fontWeight: "700", marginBottom: 16 },
  list: { gap: 10, paddingBottom: 24 },
  card: { flexDirection: "row", alignItems: "center", gap: 12, borderWidth: 1, borderRadius: 18, padding: 14 },
  badgeCircle: { width: 34, height: 34, borderRadius: 17, alignItems: "center", justifyContent: "center" },
  badgeCircleText: { color: "#fff", fontWeight: "700", fontSize: 13 },
  cardTitle: { fontSize: 15, fontWeight: "700" },
  cardMeta: { fontSize: 11, marginTop: 2 },
  pill: { borderRadius: 999, paddingHorizontal: 8, paddingVertical: 2 },
});
