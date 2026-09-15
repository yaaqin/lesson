import { router, useLocalSearchParams } from "expo-router";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  ActivityIndicator,
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
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
  type PuzzlePayload,
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
  const [gridAnswers, setGridAnswers] = useState<Record<string, Record<string, number>>>({});
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
    setGridAnswers({});
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
    (finalAnswers: Record<string, number>, finalGridAnswers: Record<string, Record<string, number>>) => {
      if (!game || submittingRef.current) return;
      submittingRef.current = true;
      const payload = game.questions.map((q) => ({
        questionId: q.id,
        selectedValue: finalAnswers[q.id] ?? null,
        selectedGrid: finalGridAnswers[q.id],
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
        submitWithAnswers(finalAnswers, gridAnswers);
      } else {
        setTimeLeft(game.timeLimitSeconds);
      }
    },
    [game, currentIndex, answers, gridAnswers, submitWithAnswers],
  );

  // grid_puzzle (kotak addition / cryptarithm): submit sekali per puzzle (gak
  // ada auto-advance kayak MC/essay karena challenge puzzle selalu cuma 1 soal),
  // lanjutannya sama kayak advanceRegular -- geser currentIndex, submit kalau abis.
  const handleAnswerPuzzle = useCallback(
    (grid: Record<string, number>) => {
      if (!game) return;
      const q = game.questions[currentIndex];
      const finalGridAnswers = { ...gridAnswers, [q.id]: grid };
      setGridAnswers(finalGridAnswers);

      const nextIndex = currentIndex + 1;
      setCurrentIndex(nextIndex);
      if (nextIndex >= game.questions.length) {
        submitWithAnswers(answers, finalGridAnswers);
      } else {
        setTimeLeft(game.timeLimitSeconds);
      }
    },
    [game, currentIndex, answers, gridAnswers, submitWithAnswers],
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
      const id = setTimeout(() => submitWithAnswers(answers, gridAnswers), 0);
      return () => clearTimeout(id);
    }
    const timer = setTimeout(() => setTimeLeft((s) => s - 1), 1000);
    return () => clearTimeout(timer);
  }, [timeLeft, phase, game, answers, gridAnswers, submitWithAnswers]);

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
            onFinish={() => submitWithAnswers(answers, gridAnswers)}
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
            onAnswerPuzzle={handleAnswerPuzzle}
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
            gridAnswers={gridAnswers}
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
  onAnswerPuzzle,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  game: StartResult;
  currentIndex: number;
  currentQuestion: StartResult["questions"][number];
  timeLeft: number;
  selected: number | null;
  onAnswer: (value: number) => void;
  onAnswerPuzzle: (grid: Record<string, number>) => void;
}) {
  const [essayValue, setEssayValue] = useState("");
  const [showPuzzleInfo, setShowPuzzleInfo] = useState(false);
  const isEssay = currentQuestion.type === "essay_numeric";
  const isPuzzle = currentQuestion.type === "grid_puzzle";

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
        {isPuzzle && (
          <View style={{ flexDirection: "row", alignItems: "center", gap: 8 }}>
            <View style={[styles.pill, { backgroundColor: Palette.purpleBg }]}>
              <Text style={{ color: Palette.purple, fontSize: 11, fontWeight: "700" }}>🧩 Puzzle</Text>
            </View>
            <Pressable
              onPress={() => setShowPuzzleInfo(true)}
              hitSlop={8}
              style={[styles.infoButton, { borderColor: Palette.purple }]}
            >
              <Text style={{ color: Palette.purple, fontSize: 12, fontWeight: "800" }}>?</Text>
            </Pressable>
          </View>
        )}
        <Text
          style={
            isPuzzle
              ? { color: theme.textSecondary, fontSize: 15, textAlign: "center" }
              : [styles.questionText, { color: theme.text }]
          }
        >
          {currentQuestion.prompt}
        </Text>
      </View>

      {isPuzzle && currentQuestion.puzzle && (
        <PuzzleInfoModal
          theme={theme}
          scheme={scheme}
          visible={showPuzzleInfo}
          kind={currentQuestion.puzzle.kind}
          onClose={() => setShowPuzzleInfo(false)}
        />
      )}

      {isPuzzle && currentQuestion.puzzle ? (
        <PuzzleAnswerView theme={theme} scheme={scheme} puzzle={currentQuestion.puzzle} onSubmit={onAnswerPuzzle} />
      ) : isEssay ? (
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
  gridAnswers,
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
  gridAnswers: Record<string, Record<string, number>>;
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
          if (q.type === "grid_puzzle") {
            // Solusi puzzle sengaja gak pernah dikirim ke klien (lihat PuzzlePayload),
            // jadi benar/salahnya dibaca dari `passed` -- challenge puzzle selalu cuma
            // 1 soal & lulus minimal 100%, jadi passed = puzzle ini bener persis.
            const attempted = !!gridAnswers[q.id];
            return (
              <View key={q.id} style={[styles.breakdownRow, { backgroundColor: theme.backgroundElement }]}>
                <Text style={{ color: theme.textSecondary, fontSize: 13, flex: 1 }}>
                  {idx + 1}. {q.prompt}
                </Text>
                <Text style={{ color: passed ? Palette.green : Palette.red, fontSize: 13, fontWeight: "700" }}>
                  {attempted ? (passed ? "Benar ✓" : "Belum tepat ✗") : "— ✗"}
                </Text>
              </View>
            );
          }

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

// PuzzleAnswerView: soal grid_puzzle selalu satu-satunya soal di sesi (lihat
// seed-puzzle), jadi gak ada auto-advance kayak MC/essay -- cuma satu tombol
// "Jawab" yang aktif begitu semua sel/huruf keisi, langsung submit ke server
// (solusinya emang gak pernah ada di klien buat dicek sendiri -- PuzzlePayload
// sengaja gak bawa itu). Port 1:1 dari apps/web.
function PuzzleAnswerView({
  theme,
  scheme,
  puzzle,
  onSubmit,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  puzzle: PuzzlePayload;
  onSubmit: (grid: Record<string, number>) => void;
}) {
  const [values, setValues] = useState<Record<string, string>>({});
  const [pickerKey, setPickerKey] = useState<string | null>(null);

  const setCell = (key: string, raw: string) => {
    setValues((prev) => ({ ...prev, [key]: raw }));
  };

  const allFilled = puzzle.blankKeys.every((key) => values[key] !== undefined && values[key] !== "");

  const submit = () => {
    if (!allFilled) return;
    const grid: Record<string, number> = {};
    for (const key of puzzle.blankKeys) {
      grid[key] = Number(values[key]);
    }
    onSubmit(grid);
  };

  return (
    <View style={{ alignItems: "center", gap: 20 }}>
      {puzzle.kind === "addition_grid" ? (
        <AdditionGridInput theme={theme} scheme={scheme} puzzle={puzzle} values={values} onOpenPicker={setPickerKey} />
      ) : (
        <CryptarithmInput theme={theme} scheme={scheme} puzzle={puzzle} values={values} onCellChange={setCell} />
      )}
      <Pressable
        disabled={!allFilled}
        onPress={submit}
        style={[styles.wideButton, { backgroundColor: Palette.blue, opacity: allFilled ? 1 : 0.4, width: "100%", maxWidth: 320 }]}
      >
        <Text style={{ color: "#fff", fontWeight: "700", fontSize: 14 }}>Jawab</Text>
      </Pressable>

      {puzzle.kind === "addition_grid" && (
        <NumberPickerModal
          theme={theme}
          scheme={scheme}
          visible={pickerKey !== null}
          max={(puzzle.size ?? 3) * (puzzle.size ?? 3)}
          given={puzzle.given}
          values={values}
          pickerKey={pickerKey}
          onPick={(n) => {
            if (!pickerKey) return;
            setCell(pickerKey, String(n));
            setPickerKey(null);
          }}
          onClear={() => {
            if (!pickerKey) return;
            setCell(pickerKey, "");
            setPickerKey(null);
          }}
          onClose={() => setPickerKey(null)}
        />
      )}
    </View>
  );
}

// AdditionGridInput: kotak kosong sekarang tombol (bukan TextInput) -- tap
// buka popup NumberPickerModal, angka yang udah kepake di kotak lain otomatis
// disable di situ, jadi gak bisa gak sengaja masukin angka dobel (lihat
// curriculumsvc/puzzle.go -- solusinya harus tiap angka 1..N² dipakai tepat sekali).
function AdditionGridInput({
  theme,
  scheme,
  puzzle,
  values,
  onOpenPicker,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  puzzle: PuzzlePayload;
  values: Record<string, string>;
  onOpenPicker: (key: string) => void;
}) {
  const size = puzzle.size ?? 3;
  const cellPx = size >= 5 ? 44 : size === 4 ? 52 : 60;
  const fontSize = size >= 5 ? 15 : 20;
  const borderColor = scheme === "dark" ? Palette.borderDark : Palette.border;

  const rows = Array.from({ length: size }, (_, r) => r);
  const cols = Array.from({ length: size }, (_, c) => c);

  return (
    <View style={{ gap: 6 }}>
      {rows.map((r) => (
        <View key={`row-${r}`} style={{ flexDirection: "row", gap: 6 }}>
          {cols.map((c) => {
            const key = `r${r}c${c}`;
            const given = puzzle.given?.[key];
            if (given !== undefined) {
              return (
                <View
                  key={key}
                  style={[
                    styles.gridCell,
                    { width: cellPx, height: cellPx, backgroundColor: theme.backgroundSelected, borderColor },
                  ]}
                >
                  <Text style={{ color: theme.text, fontWeight: "800", fontSize }}>{given}</Text>
                </View>
              );
            }
            return (
              <Pressable
                key={key}
                onPress={() => onOpenPicker(key)}
                style={[
                  styles.gridCell,
                  { width: cellPx, height: cellPx, borderColor: Palette.blue },
                ]}
              >
                <Text style={{ color: Palette.blue, fontWeight: "800", fontSize }}>{values[key] ?? ""}</Text>
              </Pressable>
            );
          })}
          <View style={[styles.gridSumCell, { width: cellPx, height: cellPx }]}>
            <Text style={{ color: Palette.amber, fontWeight: "700", fontSize: 12 }}>{puzzle.rowSums?.[r]}</Text>
          </View>
        </View>
      ))}
      <View style={{ flexDirection: "row", gap: 6 }}>
        {cols.map((c) => (
          <View key={`colsum-${c}`} style={[styles.gridSumCell, { width: cellPx, height: cellPx }]}>
            <Text style={{ color: Palette.amber, fontWeight: "700", fontSize: 12 }}>{puzzle.colSums?.[c]}</Text>
          </View>
        ))}
        <View style={{ width: cellPx, height: cellPx }} />
      </View>
    </View>
  );
}

// NumberPickerModal: popup keypad angka buat isi 1 kotak addition_grid --
// muncul begitu kotak kosong di-tap, ada tombol close (✕) eksplisit karena di
// mobile gak ada tempat buat keypad nempel permanen kayak di web.
function NumberPickerModal({
  theme,
  scheme,
  visible,
  max,
  given,
  values,
  pickerKey,
  onPick,
  onClear,
  onClose,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  visible: boolean;
  max: number;
  given: Record<string, number> | undefined;
  values: Record<string, string>;
  pickerKey: string | null;
  onPick: (n: number) => void;
  onClear: () => void;
  onClose: () => void;
}) {
  const usedNumbers = new Set<number>();
  for (const v of Object.values(given ?? {})) usedNumbers.add(v);
  for (const [key, raw] of Object.entries(values)) {
    if (key === pickerKey || raw === "") continue;
    usedNumbers.add(Number(raw));
  }
  const borderColor = scheme === "dark" ? Palette.borderDark : Palette.border;

  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onClose}>
      <Pressable style={styles.modalBackdrop} onPress={onClose}>
        <Pressable style={[styles.modalCard, { backgroundColor: theme.background }]} onPress={(e) => e.stopPropagation()}>
          <View style={styles.modalHeader}>
            <Text style={{ color: theme.text, fontWeight: "700", fontSize: 14 }}>Pilih angka</Text>
            <Pressable onPress={onClose} hitSlop={8} style={[styles.modalCloseButton, { backgroundColor: theme.backgroundElement }]}>
              <Text style={{ color: theme.textSecondary, fontSize: 15, fontWeight: "700" }}>✕</Text>
            </Pressable>
          </View>
          <View style={{ flexDirection: "row", flexWrap: "wrap", gap: 8, justifyContent: "center" }}>
            {Array.from({ length: max }, (_, i) => i + 1).map((n) => {
              const disabled = usedNumbers.has(n);
              return (
                <Pressable
                  key={n}
                  disabled={disabled}
                  onPress={() => onPick(n)}
                  style={[styles.numberPickerButton, { borderColor: Palette.blue, opacity: disabled ? 0.3 : 1 }]}
                >
                  <Text style={{ color: Palette.blue, fontWeight: "700", fontSize: 16 }}>{n}</Text>
                </Pressable>
              );
            })}
          </View>
          <Pressable
            onPress={onClear}
            style={[styles.keypadActionButton, { borderWidth: 1, borderColor, marginTop: 4 }]}
          >
            <Text style={{ color: theme.textSecondary, fontSize: 13, fontWeight: "600" }}>⌫ Hapus</Text>
          </Pressable>
        </Pressable>
      </Pressable>
    </Modal>
  );
}

// PuzzleInfoModal: penjelasan aturan main, dibuka lewat tombol "?" di sebelah
// pill "🧩 Puzzle" -- popup dengan tombol close eksplisit (bukan cuma tap
// backdrop) biar jelas buat user yang belum tau caranya nutup modal.
function PuzzleInfoModal({
  theme,
  scheme,
  visible,
  kind,
  onClose,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  visible: boolean;
  kind: PuzzlePayload["kind"];
  onClose: () => void;
}) {
  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onClose}>
      <Pressable style={styles.modalBackdrop} onPress={onClose}>
        <Pressable style={[styles.modalCard, { backgroundColor: theme.background }]} onPress={(e) => e.stopPropagation()}>
          <View style={styles.modalHeader}>
            <Text style={{ color: theme.text, fontWeight: "700", fontSize: 14 }}>Cara main</Text>
            <Pressable onPress={onClose} hitSlop={8} style={[styles.modalCloseButton, { backgroundColor: theme.backgroundElement }]}>
              <Text style={{ color: theme.textSecondary, fontSize: 15, fontWeight: "700" }}>✕</Text>
            </Pressable>
          </View>
          <Text style={{ color: theme.textSecondary, fontSize: 13, lineHeight: 19 }}>
            {kind === "addition_grid"
              ? "Isi tiap kotak kosong dengan angka 1 sampai jumlah kotaknya (misal 1-9 buat kotak 3x3) -- setiap angka cuma boleh dipakai TEPAT SEKALI di seluruh kotak, gak boleh ada yang dobel. Ketuk kotak kosong buat munculin pilihan angka -- angka yang udah kepake otomatis kekunci biar gak salah pilih. Jumlah angka di tiap baris & kolom harus sama persis dengan angka target di pinggirnya."
              : "Tiap huruf mewakili satu angka (0-9) yang nilainya tetap sama di semua kemunculannya, dan huruf beda harus angka beda. Isi digitnya supaya hasil penjumlahannya benar."}
          </Text>
        </Pressable>
      </Pressable>
    </Modal>
  );
}

function CryptarithmInput({
  theme,
  scheme,
  puzzle,
  values,
  onCellChange,
}: {
  theme: Theme;
  scheme: ColorSchemeName;
  puzzle: PuzzlePayload;
  values: Record<string, string>;
  onCellChange: (key: string, raw: string) => void;
}) {
  const words = puzzle.words ?? [];
  const addends = words.slice(0, -1);
  const result = words[words.length - 1];
  const borderColor = scheme === "dark" ? Palette.borderDark : Palette.border;

  return (
    <View style={{ width: "100%", maxWidth: 320, alignItems: "center", gap: 20 }}>
      <View style={{ alignItems: "flex-end", gap: 2 }}>
        {addends.map((w, i) => (
          <Text key={i} style={{ color: theme.text, fontWeight: "800", fontSize: 22, letterSpacing: 3 }}>
            {i > 0 ? "+ " : ""}
            {w}
          </Text>
        ))}
        <View style={{ borderTopWidth: 2, borderColor, width: "100%", paddingTop: 4 }}>
          <Text style={{ color: theme.text, fontWeight: "800", fontSize: 22, letterSpacing: 3, textAlign: "right" }}>
            {result}
          </Text>
        </View>
      </View>

      <View style={{ flexDirection: "row", flexWrap: "wrap", justifyContent: "center", gap: 10 }}>
        {puzzle.blankKeys.map((letter) => (
          <View key={letter} style={{ alignItems: "center", gap: 4 }}>
            <Text style={{ color: theme.textSecondary, fontSize: 11, fontWeight: "700" }}>{letter}</Text>
            <TextInput
              value={values[letter] ?? ""}
              onChangeText={(t) => onCellChange(letter, t.replace(/[^0-9]/g, "").slice(0, 1))}
              keyboardType="number-pad"
              maxLength={1}
              style={[styles.letterCell, { borderColor: Palette.blue, color: Palette.blue }]}
            />
          </View>
        ))}
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
  gridCell: { borderRadius: 10, borderWidth: 2, alignItems: "center", justifyContent: "center" },
  gridSumCell: { borderRadius: 10, backgroundColor: Palette.amberBg, alignItems: "center", justifyContent: "center" },
  letterCell: {
    width: 44,
    height: 44,
    borderRadius: 10,
    borderWidth: 2,
    textAlign: "center",
    fontSize: 18,
    fontWeight: "800",
  },
  infoButton: { width: 22, height: 22, borderRadius: 11, borderWidth: 1.5, alignItems: "center", justifyContent: "center" },
  modalBackdrop: { flex: 1, backgroundColor: "rgba(0,0,0,0.5)", alignItems: "center", justifyContent: "center", padding: 24 },
  modalCard: { width: "100%", maxWidth: 340, borderRadius: 20, padding: 20, gap: 14 },
  modalHeader: { flexDirection: "row", justifyContent: "space-between", alignItems: "center" },
  modalCloseButton: { width: 28, height: 28, borderRadius: 14, alignItems: "center", justifyContent: "center" },
  numberPickerButton: { width: 44, height: 44, borderRadius: 10, borderWidth: 2, alignItems: "center", justifyContent: "center" },
});
