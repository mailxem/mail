import type { Config } from "tailwindcss";
export default {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        cream: "#ffffef",
        iris: "#5b3cc4",
        ink: "#22251f",
        lemon: "#edf09b",
        forest: "#164b3f",
        lavender: "#e7d8fa",
        muted: "#66695e",
      },
      fontFamily: {
        sans: ["var(--font-sans)", "sans-serif"],
        editorial: ["var(--font-editorial)", "Georgia", "serif"],
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
} satisfies Config;
