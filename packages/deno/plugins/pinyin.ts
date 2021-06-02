import { pinyin } from "https://patarapolw.github.io/pinyin/main.js";

declare global {
  interface Window {
    pinyin(
      s: string,
      opts: { toneToNumber: boolean; keepRest: boolean },
    ): string;
    toPinyin(s: string): string;
  }
}

window.toPinyin = (q) => {
  return pinyin(q, { toneToNumber: true, keepRest: true });
};
