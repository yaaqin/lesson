import { router } from "expo-router";
import { useState } from "react";
import {
  ActivityIndicator,
  Platform,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
  useColorScheme,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { Colors } from "@/constants/theme";
import { Palette } from "@/constants/palette";
import { useLoginMutation } from "@/hooks/use-auth";

export default function LoginScreen() {
  const scheme = useColorScheme();
  const theme = Colors[scheme === "dark" ? "dark" : "light"];
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const loginMutation = useLoginMutation();

  const handleSubmit = async () => {
    setError(null);
    try {
      await loginMutation.mutateAsync({ identifier: email, password });
      router.replace("/belajar");
    } catch {
      setError("Email atau password salah.");
    }
  };

  return (
    <SafeAreaView style={[styles.safeArea, { backgroundColor: theme.background }]}>
      <View style={styles.container}>
        <Text style={[styles.brand, { color: theme.text }]}>MathQuest</Text>

        <View
          style={[
            styles.card,
            { borderColor: scheme === "dark" ? Palette.borderDark : Palette.border },
          ]}
        >
          <Text style={[styles.title, { color: theme.text }]}>Masuk</Text>
          <Text style={[styles.subtitle, { color: theme.textSecondary }]}>
            Login pakai email &amp; password akunmu.
          </Text>

          <Text style={[styles.label, { color: theme.text }]}>Email</Text>
          <TextInput
            value={email}
            onChangeText={setEmail}
            placeholder="siswa@mathquest.dev"
            placeholderTextColor={theme.textSecondary}
            autoCapitalize="none"
            keyboardType="email-address"
            style={[
              styles.input,
              {
                color: theme.text,
                borderColor: scheme === "dark" ? Palette.borderDark : Palette.border,
              },
            ]}
          />

          <Text style={[styles.label, { color: theme.text }]}>Password</Text>
          <TextInput
            value={password}
            onChangeText={setPassword}
            placeholder="••••••••"
            placeholderTextColor={theme.textSecondary}
            secureTextEntry
            style={[
              styles.input,
              {
                color: theme.text,
                borderColor: scheme === "dark" ? Palette.borderDark : Palette.border,
              },
            ]}
          />

          {error && (
            <View style={[styles.errorBox, { backgroundColor: Palette.redBg }]}>
              <Text style={{ color: Palette.red, fontSize: 13 }}>{error}</Text>
            </View>
          )}

          <Pressable
            onPress={handleSubmit}
            disabled={loginMutation.isPending}
            style={[styles.button, { opacity: loginMutation.isPending ? 0.6 : 1 }]}
          >
            {loginMutation.isPending ? (
              <ActivityIndicator color="#fff" />
            ) : (
              <Text style={styles.buttonText}>Masuk</Text>
            )}
          </Pressable>
        </View>

        <View
          style={[
            styles.hintBox,
            { borderColor: scheme === "dark" ? Palette.borderDark : Palette.border },
          ]}
        >
          <Text style={[styles.hintTitle, { color: theme.textSecondary }]}>Akun contoh</Text>
          <Text style={[styles.hintMono, { color: theme.text }]}>siswa@mathquest.dev</Text>
          <Text style={[styles.hintMono, { color: theme.textSecondary }]}>
            password: siswa12345
          </Text>
        </View>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safeArea: { flex: 1 },
  container: { flex: 1, justifyContent: "center", padding: 24, gap: 24 },
  brand: { fontSize: 20, fontWeight: "700", textAlign: "center" },
  card: { width: "100%", borderWidth: 1, borderRadius: 20, padding: 20, gap: 10 },
  title: { fontSize: 22, fontWeight: "700" },
  subtitle: { fontSize: 14, marginBottom: 8 },
  label: { fontSize: 13, fontWeight: "600", marginTop: 4 },
  input: {
    width: "100%",
    borderWidth: 1,
    borderRadius: 10,
    paddingHorizontal: 12,
    paddingVertical: Platform.select({ ios: 12, default: 8 }),
    fontSize: 15,
  },
  errorBox: { borderRadius: 10, padding: 10, marginTop: 4 },
  button: {
    marginTop: 8,
    backgroundColor: Palette.blue,
    borderRadius: 999,
    paddingVertical: 14,
    alignItems: "center",
  },
  buttonText: { color: "#fff", fontWeight: "700", fontSize: 15 },
  hintBox: { width: "100%", borderWidth: 1, borderStyle: "dashed", borderRadius: 16, padding: 14, gap: 2 },
  hintTitle: { fontSize: 12, fontWeight: "600" },
  hintMono: { fontFamily: Platform.select({ ios: "Menlo", default: "monospace" }), fontSize: 13 },
});
