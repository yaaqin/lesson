import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { DarkTheme, DefaultTheme, Stack, ThemeProvider } from "expo-router";
import * as SplashScreen from "expo-splash-screen";
import { useState } from "react";
import { useColorScheme } from "react-native";

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const colorScheme = useColorScheme();
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: { retry: 1, staleTime: 30_000 },
        },
      }),
  );

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider value={colorScheme === "dark" ? DarkTheme : DefaultTheme}>
        <Stack screenOptions={{ headerBackTitle: "Kembali" }}>
          <Stack.Screen name="index" options={{ headerShown: false }} />
          <Stack.Screen name="login" options={{ headerShown: false }} />
          <Stack.Screen name="belajar/index" options={{ title: "Pilih Jenjang", headerBackVisible: false }} />
          <Stack.Screen name="belajar/[tierCode]/index" options={{ title: "Kurikulum" }} />
          <Stack.Screen name="belajar/[tierCode]/kategori/[categoryId]" options={{ title: "Level" }} />
          <Stack.Screen
            name="belajar/[tierCode]/[challengeId]"
            options={{ title: "Challenge", headerBackTitle: "Keluar" }}
          />
        </Stack>
      </ThemeProvider>
    </QueryClientProvider>
  );
}
