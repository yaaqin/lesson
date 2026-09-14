import { router, useLocalSearchParams } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
  useColorScheme,
  type ColorSchemeName,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { Colors } from "@/constants/theme";
import { Palette } from "@/constants/palette";
import {
  useMeQuery,
  useStartChallengeMutation,
  useSubmitAttemptMutation,
  type StartResult,
  type SubmitResult,
} from "@/hooks/use-curriculum";

type Phase = "loading" | "playing" | "result" | "blocked";
type Theme = {
  text: string;
  background: string;
  backgroundElement: string;
  backgroundSelected: string;
  textSecondary: string;
};

export default function ChallengeScreen() {
  const scheme = useColorScheme();
  const theme = Colors[scheme === "dark" ? "dark" : "light"];
  const params = useLocalSearchParams<{ tierCode: string; challengeId: string }>();

  const startMutation = useStartChallengeMutation();
  const submitMutation = useSubmitAttemptMutation();
  const { mutate: startAttempt } = startMutation;
  const meQuery = useMeQuery();

  const [game, setGame] = useState<StartResult | null>(null);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, number>>({});
  const [essayDrafts, setEssayDrafts] = useState<Record<string, string>>({});
  const [selected, setSelected] = useState<number | null>(null);
  const [timeLeft, setTimeLeft] = useState(0);
  const [phase, setPhase] = useState<Phase>("loading");
  const [result, setResult] = useState<SubmitResult | null>(null);
  const [blockedReason, setBlockedReason] = useState<string | null>(null);

  const startedForRef = useRef<string | null>(null);
  const submittingRef = useRef(false);

  const beginAttempt = useCallback(() => {
    submittingRef.current = false;
    setPhase("loading");
    setCurrentIndex(0);
    setAnswers({});
    setEssayDrafts({});
    setSelected(null);
    setResult(null);
    startAttempt(params.challengeId, {
      onSuccess: (data) => {
        setGame(data);
        setTimeLeft(data.timeLimitSeconds);
        setPhase("playing");
      },
      onError: (err: unknown) => {
        const code =
          (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
          "internal_error";
        setBlockedReason(
          code === "no_lives"
            ? "Nyawa kamu habis. Tunggu reset besok, ya."
            : "Gagal memulai challenge, coba lagi.",
        );
        setPhase("blocked");
      },
    });
  }, [params.challengeId, startAttempt]);

  useEffect(() => {
    if (startedForRef.current === params.challengeId) return;
    startedForRef.current = params.challengeId;
    beginAttempt();
  }, [params.challengeId, beginAttempt]);

  const submitWithAnswers = useCallback(
    (finalAnswers: Record<string, number>) => {
      if (!game || submittingRef.current) return;
      submittingRef.current = true;
      const payload = game.questions.map((q) => ({
        questionId: q.id,
        selectedValue: finalAnswers[q.id] ?? null,
      }));
      submitMutation.mutate(
        { attemptId: game.attemptId, answers: payload },
        {
          onSuccess: (data) => {
            setResult(data);
            setPhase("result");
          },
        },
      );
    },
    [game, submitMutation],
  );

  // Mode "challenge biasa": waktu per soal, auto-lanjut begitu waktu habis / jawab.
  const advanceRegular = useCallback(
    (selectedValue: number | null) => {
      if (!game) return;
      const q = game.questions[currentIndex];
      const finalAnswers = selectedValue === null ? answers : { ...answers, [q.id]: selectedValue };
      setAnswers(finalAnswers);
      setSelected(null);

      const nextIndex = currentIndex + 1;
      setCurrentIndex(nextIndex);
      if (nextIndex >= game.questions.length) {
        submitWithAnswers(finalAnswers);
      } else {
        setTimeLeft(game.timeLimitSeconds);
      }
    },
    [game, currentIndex, answers, submitWithAnswers],
  );

  // Timer per soal (challenge biasa)
  useEffect(() => {
    if (phase !== "playing" || !game || game.isExam) return;
    if (timeLeft <= 0) {
      const id = setTimeout(() => advanceRegular(null), 0);
      return () => clearTimeout(id);
    }
    const timer = setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, phase, game, advanceRegular]);

  // Timer total (ujian) — jalan terus gak peduli lagi di soal mana
  useEffect(() => {
    if (phase !== "playing" || !game || !game.isExam) return;
    if (timeLeft <= 0) {
      const id = setTimeout(() => submitWithAnswers(answers), 0);
      return () => clearTimeout(id);
    }
    const timer = setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, phase, game, answers, submitWithAnswers]);

  const handleAnswerRegular = (value: number) => {
    if (selected !== null) return;
    setSelected(value);
    setTimeout(() => advanceRegular(value), 450);
  };

  const handleAnswerExam = (questionId: string, value: number) => {
    setAnswers((prev) => ({ ...prev, [questionId]: value }));
  };

  const handleEssayDraftChange = (questionId: string, draft: string) => {
    setEssayDrafts((prev) => ({ ...prev, [questionId]: draft }));
    const parsed = Number(draft);
    if (draft !== "" && draft !== "-" && Number.isFinite(parsed)) {
      setAnswers((prev) => ({ ...prev, [questionId]: parsed }));
    } else {
      setAnswers((prev) => {
        const next = { ...prev };
        delete next[questionId];
        return next;
      });
    }
  };

  const retry = () => {
    startedForRef.current = null;
    beginAttempt();
  };

  if (phase === "blocked") {
    return (
      <SafeAreaView style={[styles.center, { backgroundColor: theme.background }]}>
        <Text style={{ fontSize: 40 }}>💔</Text>
        <LivesBadge lives={meQuery.data?.livesRemaining ?? 0} />
        <Text style={[styles.blockedText, { color: theme.textSecondary }]}>{blockedReason}</Text>
        <Pressable onPress={() => router.replace(`/belajar/${params.tierCode}`)}>
          <Text style={{ color: Palette.blue, fontWeight: "700" }}>Kembali ke daftar</Text>
        </Pressable>
      </SafeAreaView>
    );
  }

  if (phase === "loading" || !game) {
    return (
      <SafeAreaView style={[styles.center, { backgroundColor: theme.background }]}>
        <ActivityIndicator />
        <Text style={{ color: theme.textSecondary, marginTop: 8 }}>Menyiapkan soal…</Text>
      </SafeAreaView>
    );
  }

  const currentQuestion = game.questions[currentIndex];
  const correctCount = result?.correctCount ?? 0;
  const passed = result?.passed ?? false;

  return (
    <SafeAreaView style={[styles.safeArea, { backgroundColor: theme.background }]} edges={["bottom"]}>
      <View style={styles.header}>
        <Text style={[styles.headerTitle, { color: theme.textSecondary }]}>{game.challengeName}</Text>
        <LivesBadge lives={meQuery.data?.livesRemaining ?? 0} />
      </View>

      <ScrollView contentContainerStyle={styles.scrollBody}>
        {phase === "playing" && game.isExam && (
          <ExamView
            theme={theme}
            scheme={scheme}
            game={game}
            currentIndex={currentIndex}
            setCurrentIndex={setCurrentIndex}
            answers={answers}
            essayDrafts={essayDrafts}
            onEssayChange={handleEssayDraftChange}
            timeLeft={timeLeft}
            onAnswer={handleAnswerExam}
            onFinish={() => submitWithAnswers(answers)}
          />
        )}

        {phase === "playing" && !game.isExam && currentQuestion && (
          <RegularView
            key={currentQuestion.id}
            theme={theme}
            scheme={scheme}
            game={game}
            currentIndex={currentIndex}
            currentQuestion={currentQuestion}
            timeLeft={timeLeft}
            selected={selected}
            onAnswer={handleAnswerRegular}
          />
        )}

        {phase === "playing" && !game.isExam && !currentQuestion && (
          <View style={styles.center}>
            <Text style={{ color: theme.textSecondary }}>Menghitung skor…</Text>
          </View>
        )}

        {phase === "result" && result && (
          <ResultView
            theme={theme}
            scheme={scheme}
            result={result}
            passed={passed}
            correctCount={correctCount}
            totalQuestions={result.totalQuestions}
            passThresholdPercent={game.passThresholdPercent}
            tierCode={params.tierCode}
            onRetry={retry}
            questions={game.questions}
            answers={answers}
          />
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

function RegularView({
  theme,
  scheme,
  game,
  currentIndex,
  currentQuestion,
  timeLeft,
  selected,
  onAnswer,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  game: StartResult;
  currentIndex: number;
  currentQuestion: StartResult["questions"][number];
  timeLeft: number;
  selected: number | null;
  onAnswer: (value: number) => void;
}) {
  const [essayValue, setEssayValue] = useState("");
  const isEssay = currentQuestion.type === "essay_numeric";

  const submitEssay = () => {
    if (selected !== null || essayValue === "" || essayValue === "-") return;
    const num = Number(essayValue);
    if (!Number.isFinite(num)) return;
    onAnswer(num);
  };

  const essayFeedback: "correct" | "incorrect" | null =
    selected === null
      ? null
      : currentQuestion.correctAnswerValue !== undefined &&
          Math.abs(selected - currentQuestion.correctAnswerValue) < 1e-9
        ? "correct"
        : "incorrect";

  return (
    <View style={{ gap: 24 }}>
      <View style={{ gap: 8 }}>
        <View style={styles.rowBetween}>
          <Text style={{ color: theme.textSecondary, fontSize: 12 }}>{game.challengeName}</Text>
          <Text style={{ color: theme.textSecondary, fontSize: 12 }}>
            Soal {currentIndex + 1} / {game.questions.length}
          </Text>
        </View>
        <View style={[styles.progressTrack, { backgroundColor: scheme === "dark" ? Palette.borderDark : Palette.border }]}>
          <View
            style={[
              styles.progressFill,
              { width: `${((currentIndex + 1) / game.questions.length) * 100}%` },
            ]}
          />
        </View>
      </View>

      <View style={{ alignItems: "center", gap: 12, paddingVertical: 16 }}>
        <Text style={{ color: timeLeft <= 5 ? Palette.red : theme.textSecondary, fontWeight: "700", fontSize: 13 }}>
          ⏱️ {timeLeft}s
        </Text>
        {isEssay && (
          <View style={[styles.pill, { backgroundColor: Palette.purpleBg }]}>
            <Text style={{ color: Palette.purple, fontSize: 11, fontWeight: "700" }}>
              ✏️ Isian — ketik jawabannya
            </Text>
          </View>
        )}
        <Text style={[styles.questionText, { color: theme.text }]}>{currentQuestion.prompt}</Text>
      </View>

      {isEssay ? (
        <View style={{ alignItems: "center" }}>
          <NumericKeypad
            theme={theme}
            scheme={scheme}
            value={selected !== null ? String(selected) : essayValue}
            onChange={setEssayValue}
            onSubmit={submitEssay}
            disabled={selected !== null}
            feedback={essayFeedback}
          />
        </View>
      ) : (
        <View style={styles.optionsGrid}>
          {(currentQuestion.options ?? []).map((opt) => {
            const isSelected = selected === opt.value;
            const showCorrectness = selected !== null;
            const bg = showCorrectness && opt.isCorrect
              ? Palette.greenBg
              : showCorrectness && isSelected
                ? Palette.redBg
                : theme.backgroundElement;
            const fg = showCorrectness && opt.isCorrect
              ? Palette.green
              : showCorrectness && isSelected
                ? Palette.red
                : theme.text;
            const borderColor = showCorrectness && opt.isCorrect
              ? Palette.green
              : showCorrectness && isSelected
                ? Palette.red
                : scheme === "dark" ? Palette.borderDark : Palette.border;
            return (
              <Pressable
                key={opt.value}
                disabled={selected !== null}
                onPress={() => onAnswer(opt.value)}
                style={[styles.optionButton, { backgroundColor: bg, borderColor }]}
              >
                <Text style={{ color: fg, fontSize: 20, fontWeight: "700" }}>{opt.value}</Text>
              </Pressable>
            );
          })}
        </View>
      )}
    </View>
  );
}

function ExamView({
  theme,
  scheme,
  game,
  currentIndex,
  setCurrentIndex,
  answers,
  essayDrafts,
  onEssayChange,
  timeLeft,
  onAnswer,
  onFinish,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  game: StartResult;
  currentIndex: number;
  setCurrentIndex: (fn: (i: number) => number) => void;
  answers: Record<string, number>;
  essayDrafts: Record<string, string>;
  onEssayChange: (questionId: string, draft: string) => void;
  timeLeft: number;
  onAnswer: (questionId: string, value: number) => void;
  onFinish: () => void;
}) {
  const q = game.questions[currentIndex];
  const isEssay = q.type === "essay_numeric";
  const answeredCount = Object.keys(answers).length;
  const minutes = Math.floor(timeLeft / 60);
  const seconds = timeLeft % 60;

  return (
    <View style={{ gap: 20 }}>
      <View style={[styles.examBanner, { backgroundColor: Palette.amberBg }]}>
        <View style={styles.rowBetween}>
          <Text style={{ color: Palette.amber, fontWeight: "700", fontSize: 13 }}>
            🏁 {game.challengeName} · mode ujian
          </Text>
          <Text style={{ color: timeLeft <= 30 ? Palette.red : Palette.amber, fontWeight: "700", fontSize: 16 }}>
            {minutes}:{seconds.toString().padStart(2, "0")}
          </Text>
        </View>
        <Text style={{ color: Palette.amber, fontSize: 11, marginTop: 4 }}>
          Waktu ini buat semua {game.questions.length} soal — bebas pindah-pindah soal.
        </Text>
      </View>

      <View style={styles.dotsRow}>
        {game.questions.map((question, i) => {
          const isCurrent = i === currentIndex;
          const isAnswered = answers[question.id] !== undefined;
          return (
            <Pressable
              key={question.id}
              onPress={() => setCurrentIndex(() => i)}
              style={[
                styles.dot,
                {
                  backgroundColor: isCurrent ? Palette.blue : isAnswered ? Palette.greenBg : theme.backgroundElement,
                },
              ]}
            >
              <Text style={{ color: isCurrent ? "#fff" : isAnswered ? Palette.green : theme.textSecondary, fontSize: 11, fontWeight: "700" }}>
                {i + 1}
              </Text>
            </Pressable>
          );
        })}
      </View>

      <View style={{ alignItems: "center", gap: 10, paddingVertical: 12 }}>
        <Text style={{ color: theme.textSecondary, fontSize: 12 }}>
          Soal {currentIndex + 1} / {game.questions.length}
        </Text>
        {isEssay && (
          <View style={[styles.pill, { backgroundColor: Palette.purpleBg }]}>
            <Text style={{ color: Palette.purple, fontSize: 11, fontWeight: "700" }}>
              ✏️ Isian — ketik jawabannya
            </Text>
          </View>
        )}
        <Text style={[styles.questionText, { color: theme.text }]}>{q.prompt}</Text>
      </View>

      {isEssay ? (
        <View style={{ alignItems: "center" }}>
          <NumericKeypad
            theme={theme}
            scheme={scheme}
            value={essayDrafts[q.id] ?? ""}
            onChange={(v) => onEssayChange(q.id, v)}
          />
        </View>
      ) : (
        <View style={styles.optionsGrid}>
          {(q.options ?? []).map((opt) => {
            const isSelected = answers[q.id] === opt.value;
            return (
              <Pressable
                key={opt.value}
                onPress={() => onAnswer(q.id, opt.value)}
                style={[
                  styles.optionButton,
                  {
                    backgroundColor: isSelected ? Palette.blueDark + "22" : theme.backgroundElement,
                    borderColor: isSelected ? Palette.blue : scheme === "dark" ? Palette.borderDark : Palette.border,
                  },
                ]}
              >
                <Text style={{ color: isSelected ? Palette.blue : theme.text, fontSize: 20, fontWeight: "700" }}>
                  {opt.value}
                </Text>
              </Pressable>
            );
          })}
        </View>
      )}

      <View style={styles.rowBetween}>
        <Pressable
          disabled={currentIndex === 0}
          onPress={() => setCurrentIndex((i) => Math.max(0, i - 1))}
          style={[styles.navButton, { borderColor: scheme === "dark" ? Palette.borderDark : Palette.border, opacity: currentIndex === 0 ? 0.4 : 1 }]}
        >
          <Text style={{ color: theme.textSecondary, fontSize: 13, fontWeight: "600" }}>← Sebelumnya</Text>
        </Pressable>
        <Text style={{ color: theme.textSecondary, fontSize: 11 }}>
          {answeredCount}/{game.questions.length} terjawab
        </Text>
        {currentIndex === game.questions.length - 1 ? (
          <Pressable onPress={onFinish} style={[styles.navButton, { backgroundColor: Palette.blue }]}>
            <Text style={{ color: "#fff", fontSize: 13, fontWeight: "700" }}>Selesai</Text>
          </Pressable>
        ) : (
          <Pressable
            onPress={() => setCurrentIndex((i) => Math.min(game.questions.length - 1, i + 1))}
            style={[styles.navButton, { borderColor: scheme === "dark" ? Palette.borderDark : Palette.border }]}
          >
            <Text style={{ color: theme.textSecondary, fontSize: 13, fontWeight: "600" }}>Berikutnya →</Text>
          </Pressable>
        )}
      </View>
    </View>
  );
}

function ResultView({
  theme,
  scheme,
  result,
  passed,
  correctCount,
  totalQuestions,
  passThresholdPercent,
  tierCode,
  onRetry,
  questions,
  answers,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  result: SubmitResult;
  passed: boolean;
  correctCount: number;
  totalQuestions: number;
  passThresholdPercent: number;
  tierCode: string;
  onRetry: () => void;
  questions: StartResult["questions"];
  answers: Record<string, number>;
}) {
  return (
    <View style={{ alignItems: "center", gap: 20, paddingVertical: 16 }}>
      <Text style={{ fontSize: 44 }}>{passed ? "🎉" : "💪"}</Text>

      <View style={{ alignItems: "center", gap: 4 }}>
        <Text style={{ fontSize: 44, fontWeight: "800", color: theme.text }}>{result.scorePercent}%</Text>
        <Text style={{ color: passed ? Palette.green : Palette.red, fontWeight: "600", fontSize: 13 }}>
          {passed ? "Lulus! Nilai minimal terpenuhi." : `Belum lulus — minimal ${passThresholdPercent}%.`}
        </Text>
      </View>

      <View style={[styles.statsRow, { backgroundColor: theme.backgroundElement }]}>
        <View style={styles.statItem}>
          <Text style={{ fontSize: 20 }}>🔥</Text>
          <Text style={{ color: Palette.amber, fontWeight: "700", fontSize: 12 }}>
            Streak {result.currentStreak}
          </Text>
        </View>
        <View style={styles.statItem}>
          <Text style={{ fontSize: 18, color: theme.text, fontWeight: "700" }}>
            {correctCount}/{totalQuestions}
          </Text>
          <Text style={{ color: theme.textSecondary, fontSize: 11 }}>Jawaban benar</Text>
        </View>
        <View style={styles.statItem}>
          <View style={{ flexDirection: "row", gap: 2 }}>
            {[0, 1, 2].map((i) => (
              <Text key={i} style={{ fontSize: 15, opacity: i < result.livesRemaining ? 1 : 0.25 }}>
                ❤️
              </Text>
            ))}
          </View>
          <Text style={{ color: theme.textSecondary, fontSize: 11 }}>Nyawa</Text>
        </View>
      </View>

      <View style={{ width: "100%", maxWidth: 320, gap: 10 }}>
        {!passed && (
          <Pressable onPress={onRetry} style={[styles.wideButton, { backgroundColor: Palette.blue }]}>
            <Text style={{ color: "#fff", fontWeight: "700", fontSize: 14 }}>Ulangi Challenge</Text>
          </Pressable>
        )}
        <Pressable
          onPress={() => router.replace(`/belajar/${tierCode}`)}
          style={[
            styles.wideButton,
            passed
              ? { backgroundColor: Palette.blue }
              : { borderWidth: 1, borderColor: scheme === "dark" ? Palette.borderDark : Palette.border },
          ]}
        >
          <Text style={{ color: passed ? "#fff" : theme.textSecondary, fontWeight: "700", fontSize: 14 }}>
            Kembali ke Daftar
          </Text>
        </Pressable>
      </View>

      <View style={{ width: "100%", gap: 8 }}>
        {questions.map((q, idx) => {
          const selectedValue = answers[q.id];
          const correctValue =
            q.type === "essay_numeric" ? q.correctAnswerValue : q.options?.find((o) => o.isCorrect)?.value;
          const isCorrect =
            selectedValue !== undefined && correctValue !== undefined && Math.abs(selectedValue - correctValue) < 1e-9;
          return (
            <View
              key={q.id}
              style={[
                styles.breakdownRow,
                { backgroundColor: theme.backgroundElement },
              ]}
            >
              <Text style={{ color: theme.textSecondary, fontSize: 13, flex: 1 }}>
                {idx + 1}. {q.prompt}
              </Text>
              <Text style={{ color: isCorrect ? Palette.green : Palette.red, fontSize: 13, fontWeight: "700" }}>
                {selectedValue ?? "—"} {isCorrect ? "✓" : "✗"}
              </Text>
            </View>
          );
        })}
      </View>
    </View>
  );
}

function LivesBadge({ lives }: { lives: number }) {
  return (
    <View style={{ flexDirection: "row", gap: 2 }}>
      {[0, 1, 2].map((i) => (
        <Text key={i} style={{ fontSize: 15, opacity: i < lives ? 1 : 0.25 }}>
          ❤️
        </Text>
      ))}
    </View>
  );
}

const KEYPAD_KEYS = ["7", "8", "9", "4", "5", "6", "1", "2", "3", "-", "0", "."];

function NumericKeypad({
  theme,
  scheme,
  value,
  onChange,
  onSubmit,
  disabled,
  feedback,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  value: string;
  onChange: (value: string) => void;
  onSubmit?: () => void;
  disabled?: boolean;
  feedback?: "correct" | "incorrect" | null;
}) {
  const press = (key: string) => {
    if (disabled) return;
    if (key === "-") {
      onChange(value.startsWith("-") ? value.slice(1) : "-" + value);
      return;
    }
    if (key === "." && value.includes(".")) return;
    onChange(value + key);
  };

  const displayBg = feedback === "correct" ? Palette.greenBg : feedback === "incorrect" ? Palette.redBg : theme.backgroundElement;
  const displayFg = feedback === "correct" ? Palette.green : feedback === "incorrect" ? Palette.red : theme.text;
  const displayBorder = feedback === "correct" ? Palette.green : feedback === "incorrect" ? Palette.red : scheme === "dark" ? Palette.borderDark : Palette.border;

  return (
    <View style={{ width: "100%", maxWidth: 300, gap: 10 }}>
      <View style={[styles.keypadDisplay, { backgroundColor: displayBg, borderColor: displayBorder }]}>
        <Text style={{ color: value ? displayFg : theme.textSecondary, fontSize: 28, fontWeight: "700" }}>
          {value || "0"}
        </Text>
      </View>

      <View style={styles.keypadGrid}>
        {KEYPAD_KEYS.map((key) => (
          <Pressable
            key={key}
            disabled={disabled}
            onPress={() => press(key)}
            style={[
              styles.keypadButton,
              { backgroundColor: theme.backgroundElement, borderColor: scheme === "dark" ? Palette.borderDark : Palette.border, opacity: disabled ? 0.5 : 1 },
            ]}
          >
            <Text style={{ color: theme.text, fontSize: 18, fontWeight: "700" }}>{key}</Text>
          </Pressable>
        ))}
      </View>

      <View style={{ flexDirection: "row", gap: 8 }}>
        <Pressable
          disabled={disabled}
          onPress={() => onChange(value.slice(0, -1))}
          style={[
            styles.keypadActionButton,
            { flex: 1, borderWidth: 1, borderColor: scheme === "dark" ? Palette.borderDark : Palette.border, opacity: disabled ? 0.5 : 1 },
          ]}
        >
          <Text style={{ color: theme.textSecondary, fontSize: 13, fontWeight: "600" }}>⌫ Hapus</Text>
        </Pressable>
        {onSubmit && (
          <Pressable
            disabled={disabled || value === "" || value === "-"}
            onPress={onSubmit}
            style={[
              styles.keypadActionButton,
              { flex: 1, backgroundColor: Palette.blue, opacity: disabled || value === "" || value === "-" ? 0.5 : 1 },
            ]}
          >
            <Text style={{ color: "#fff", fontSize: 13, fontWeight: "700" }}>Jawab</Text>
          </Pressable>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  safeArea: { flex: 1 },
  center: { flex: 1, alignItems: "center", justifyContent: "center", gap: 8, padding: 24 },
  blockedText: { textAlign: "center", fontSize: 14 },
  header: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", paddingHorizontal: 20, paddingVertical: 12 },
  headerTitle: { fontSize: 13, fontWeight: "600" },
  scrollBody: { paddingHorizontal: 20, paddingBottom: 32, flexGrow: 1 },
  rowBetween: { flexDirection: "row", justifyContent: "space-between", alignItems: "center" },
  progressTrack: { height: 6, borderRadius: 999, overflow: "hidden" },
  progressFill: { height: "100%", backgroundColor: Palette.blue, borderRadius: 999 },
  questionText: { fontSize: 30, fontWeight: "800", textAlign: "center" },
  pill: { borderRadius: 999, paddingHorizontal: 10, paddingVertical: 4 },
  optionsGrid: { flexDirection: "row", flexWrap: "wrap", gap: 10, justifyContent: "center" },
  optionButton: { flexBasis: "30%", flexGrow: 1, borderWidth: 2, borderRadius: 16, paddingVertical: 20, alignItems: "center" },
  examBanner: { borderRadius: 16, padding: 14 },
  dotsRow: { flexDirection: "row", flexWrap: "wrap", gap: 8 },
  dot: { width: 30, height: 30, borderRadius: 15, alignItems: "center", justifyContent: "center" },
  navButton: { borderRadius: 999, paddingHorizontal: 16, paddingVertical: 10 },
  statsRow: { flexDirection: "row", gap: 20, borderRadius: 18, paddingHorizontal: 20, paddingVertical: 14 },
  statItem: { alignItems: "center", gap: 4 },
  wideButton: { borderRadius: 999, paddingVertical: 14, alignItems: "center" },
  breakdownRow: { flexDirection: "row", justifyContent: "space-between", alignItems: "center", gap: 8, borderRadius: 12, paddingHorizontal: 14, paddingVertical: 10 },
  keypadDisplay: { height: 64, borderRadius: 16, borderWidth: 2, alignItems: "center", justifyContent: "center" },
  keypadGrid: { flexDirection: "row", flexWrap: "wrap", gap: 8 },
  keypadButton: { flexBasis: "30%", flexGrow: 1, borderWidth: 1, borderRadius: 12, paddingVertical: 16, alignItems: "center" },
  keypadActionButton: { borderRadius: 12, paddingVertical: 12, alignItems: "center" },
});
