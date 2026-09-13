// Data dummy — belum tersambung ke Postgres/Go API.
// Struktur mengikuti skema di FSD.md bagian 4.3/4.4 (tiers -> batches -> challenges -> questions).

export type QuestionOption = {
  value: number;
  isCorrect: boolean;
};

export type Question = {
  id: string;
  prompt: string;
  options: QuestionOption[];
};

export type Challenge = {
  id: string;
  name: string;
  questionCountRequired: number;
  passThresholdPercent: number;
  optionCount: number;
  bank: Question[];
};

function opts(correct: number, wrong: number[]): QuestionOption[] {
  return [correct, ...wrong].map((value) => ({
    value,
    isCorrect: value === correct,
  }));
}

const penjumlahanBank: Question[] = [
  { id: "pj-1", prompt: "2 + 2", options: opts(4, [3, 5]) },
  { id: "pj-2", prompt: "3 + 4", options: opts(7, [6, 8]) },
  { id: "pj-3", prompt: "5 + 1", options: opts(6, [5, 7]) },
  { id: "pj-4", prompt: "6 + 3", options: opts(9, [8, 10]) },
  { id: "pj-5", prompt: "4 + 4", options: opts(8, [7, 9]) },
  { id: "pj-6", prompt: "7 + 2", options: opts(9, [8, 10]) },
  { id: "pj-7", prompt: "1 + 8", options: opts(9, [8, 10]) },
  { id: "pj-8", prompt: "5 + 5", options: opts(10, [9, 11]) },
];

const penguranganBank: Question[] = [
  { id: "pg-1", prompt: "9 - 3", options: opts(6, [5, 7]) },
  { id: "pg-2", prompt: "8 - 2", options: opts(6, [5, 7]) },
  { id: "pg-3", prompt: "10 - 4", options: opts(6, [5, 7]) },
  { id: "pg-4", prompt: "7 - 5", options: opts(2, [1, 3]) },
  { id: "pg-5", prompt: "6 - 1", options: opts(5, [4, 6]) },
  { id: "pg-6", prompt: "9 - 6", options: opts(3, [2, 4]) },
  { id: "pg-7", prompt: "10 - 7", options: opts(3, [2, 4]) },
  { id: "pg-8", prompt: "8 - 5", options: opts(3, [2, 4]) },
];

const perkalianBank: Question[] = [
  { id: "pk-1", prompt: "2 × 3", options: opts(6, [4, 8]) },
  { id: "pk-2", prompt: "3 × 3", options: opts(9, [6, 12]) },
  { id: "pk-3", prompt: "4 × 2", options: opts(8, [6, 10]) },
  { id: "pk-4", prompt: "5 × 2", options: opts(10, [8, 12]) },
  { id: "pk-5", prompt: "2 × 6", options: opts(12, [10, 14]) },
  { id: "pk-6", prompt: "3 × 4", options: opts(12, [9, 15]) },
  { id: "pk-7", prompt: "2 × 2", options: opts(4, [2, 6]) },
  { id: "pk-8", prompt: "5 × 3", options: opts(15, [12, 18]) },
];

export const batch: {
  tierName: string;
  batchName: string;
  challenges: Challenge[];
} = {
  tierName: "SD",
  batchName: "Batch 1 — Kelas 1",
  challenges: [
    {
      id: "penjumlahan-dasar",
      name: "Penjumlahan Dasar",
      questionCountRequired: 5,
      passThresholdPercent: 70,
      optionCount: 3,
      bank: penjumlahanBank,
    },
    {
      id: "pengurangan-dasar",
      name: "Pengurangan Dasar",
      questionCountRequired: 5,
      passThresholdPercent: 70,
      optionCount: 3,
      bank: penguranganBank,
    },
    {
      id: "perkalian-dasar",
      name: "Perkalian Dasar",
      questionCountRequired: 5,
      passThresholdPercent: 70,
      optionCount: 3,
      bank: perkalianBank,
    },
  ],
};

export function findChallenge(id: string): Challenge | undefined {
  return batch.challenges.find((c) => c.id === id);
}
