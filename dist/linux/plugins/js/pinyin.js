import { pinyin } from "https://patarapolw.github.io/pinyin/main.js";

window.toPinyin = (q) => {
    return pinyin(q, { toneToNumber: true, keepRest: true });
};
