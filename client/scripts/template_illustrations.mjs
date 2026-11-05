// Draws the Studio collection's illustrations as SVG and writes 2x PNGs to
// public/assets/template-starters/illustrations. Run from client/:
//   node scripts/template_illustrations.mjs
// Every drawing is original line art; a displacement filter gives the lines
// a hand-drawn wobble. Templates reference these as replaceable image blocks.
import sharp from "sharp";
import fs from "node:fs/promises";
import { fileURLToPath } from "node:url";

const out = fileURLToPath(
  new URL("../public/assets/template-starters/illustrations/", import.meta.url),
);
await fs.mkdir(out, { recursive: true });

const svg = (w, h, bg, ink, body, sw = 4) =>
  `<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}" viewBox="0 0 ${w} ${h}">
<defs><filter id="w" x="-5%" y="-5%" width="110%" height="110%"><feTurbulence type="fractalNoise" baseFrequency="0.03" numOctaves="2" seed="7"/><feDisplacementMap in="SourceGraphic" scale="4"/></filter></defs>
${bg ? `<rect width="${w}" height="${h}" fill="${bg}"/>` : ""}
<g filter="url(#w)" fill="none" stroke="${ink}" stroke-width="${sw}" stroke-linecap="round" stroke-linejoin="round">${body}</g></svg>`;

const star = (cx, cy, r, fill = "none") => {
  const p = [];
  for (let i = 0; i < 10; i++) {
    const a = (Math.PI / 5) * i - Math.PI / 2;
    const rr = i % 2 ? r * 0.45 : r;
    p.push(`${(cx + rr * Math.cos(a)).toFixed(1)},${(cy + rr * Math.sin(a)).toFixed(1)}`);
  }
  return `<polygon points="${p.join(" ")}" fill="${fill}"/>`;
};
const dash = (x, y, a, l = 16) =>
  `<path d="M${x} ${y}l${(l * Math.cos(a)).toFixed(1)} ${(l * Math.sin(a)).toFixed(1)}"/>`;
const burst = (x, y, r = 22) =>
  [0, 1, 2, 3, 4, 5, 6, 7]
    .map((i) => {
      const a = (Math.PI / 4) * i;
      return `<path d="M${(x + r * 0.45 * Math.cos(a)).toFixed(1)} ${(y + r * 0.45 * Math.sin(a)).toFixed(1)}L${(x + r * Math.cos(a)).toFixed(1)} ${(y + r * Math.sin(a)).toFixed(1)}"/>`;
    })
    .join("");

const art = {
  // Gift shop thank-you: a stack of parcels with a small bird on top.
  "gift-stack": svg(600, 300, "#D9E1C6", "#2F5D46", `
    <rect x="130" y="170" width="170" height="150" rx="6" fill="#F6F1E4"/><path d="M215 170v150M130 225h170"/>
    <rect x="290" y="140" width="190" height="180" rx="6" fill="#F6F1E4"/><path d="M385 140v180M290 200h190"/>
    <path d="M385 140c-20-34-62-30-48-6 9 14 30 8 48 6 18 2 39 8 48-6 14-24-28-28-48 6z" fill="#E2B24A"/>
    <rect x="200" y="100" width="120" height="70" rx="6" fill="#F6F1E4"/><path d="M260 100v70"/>
    <path d="M232 100c-6-34 26-52 46-34 16 14 6 34-14 34" fill="#D9E1C6"/><circle cx="262" cy="74" r="3" fill="#2F5D46"/><path d="M276 66l12-2-8 10"/>
    ${dash(110, 120, -2.2)}${dash(96, 160, -2.8)}${dash(500, 110, -0.9)}${dash(515, 160, -0.3)}
    <path d="M60 300c10-40 40-60 60-60-4 30-30 56-60 60zM60 300c18-20 36-36 58-58"/>
    <path d="M545 300c-6-36-30-58-52-62 0 30 22 54 52 62zM545 300c-14-20-30-40-50-60"/>`),
  // A tote with a parcel, cropped into a circle for the article card.
  "tote-circle": svg(300, 300, null, "#2F5D46", `
    <circle cx="150" cy="150" r="140" fill="#F4ECDD" stroke="none"/>
    <path d="M95 140h110l-12 140H107z" fill="#3F7A5A"/><path d="M118 140c0-48 64-48 64 0"/>
    ${[0, 1, 2].map((i) => `<circle cx="${125 + i * 25}" cy="${185 + (i % 2) * 30}" r="9" stroke="#F4ECDD" stroke-width="3"/>`).join("")}
    <rect x="112" y="92" width="76" height="40" rx="4" fill="#D9B98C"/><path d="M150 92v40"/>`),
  // Loyalty: a scatter of drinks, bakes and stars.
  "breakfast-scatter": svg(600, 340, "#EFEDE8", "#213D33", `
    <path d="M262 120h96l-14 170h-68z" fill="#FFFFFF"/><path d="M256 120h108"/><circle cx="310" cy="200" r="24" fill="#3E7A5E" stroke="none"/>
    <path d="M380 70c0-26 70-26 70 0 12 0 14 22 0 24v52h-70V94c-14-2-12-24 0-24z" fill="#FFFFFF"/>${[0, 1, 2].map((i) => `<circle cx="${398 + i * 16}" cy="${110 + (i % 2) * 14}" r="2" fill="#213D33"/>`).join("")}
    <path d="M420 250c20-40 80-40 100 0-20 16-80 16-100 0z" fill="#FFFFFF"/><path d="M445 236l8 22M470 230v26M495 236l-8 22"/>
    <path d="M110 250h90c0 22-90 22-90 0zM110 238c0-24 90-24 90 0z" fill="#FFFFFF"/><path d="M114 244h82"/>
    <path d="M150 70l40-30c10 40-14 70-40 30z" fill="#FFFFFF"/><path d="M160 66l22-16"/>
    <circle cx="230" cy="150" r="20" fill="#213D33"/><path d="M230 170v90"/>
    <circle cx="500" cy="150" r="26" fill="#213D33"/><path d="M486 140c10 4 18 4 28-6" stroke="#EFEDE8"/>
    ${star(90, 160, 18, "#213D33")}${star(540, 90, 16, "#213D33")}${star(220, 70, 14, "#213D33")}${star(380, 300, 15, "#213D33")}${star(120, 310, 14, "#213D33")}
    ${dash(300, 40, 0.6)}${dash(560, 220, 2.2)}${dash(60, 90, 1)}`),
  "icon-phone": svg(120, 90, null, "#213D33", `<rect x="44" y="12" width="32" height="56" rx="6"/><path d="M56 20h8"/><path d="M60 52v26c6-2 10 2 10 6"/>`, 3),
  "icon-cup": svg(120, 90, null, "#213D33", `<path d="M44 22h32l-5 54H49z"/><path d="M42 22h36M48 14h24l4 8"/><circle cx="60" cy="46" r="7"/>`, 3),
  "icon-star": svg(120, 90, null, "#213D33", `${star(54, 50, 22)}${star(80, 24, 10)}`, 3),
  "icon-gift": svg(120, 90, null, "#213D33", `<rect x="38" y="36" width="44" height="40" rx="3"/><path d="M34 26h52v10H34zM60 26v50"/><path d="M60 26c-8-14-22-10-16-2M60 26c8-14 22-10 16-2"/>`, 3),
  // Tea shop mascot: a tortoise.
  tortoise: svg(600, 260, null, "#1F5673", `
    <path d="M170 190c0-90 70-130 140-130s140 40 140 130z"/>
    <path d="M220 190l20-60 70-30 70 30 20 60M240 130l70 10 70-10M310 140v50"/>
    <path d="M450 170c20-10 50-20 70-6 16 12 6 34-20 36-20 2-40-10-50-20"/><circle cx="500" cy="176" r="4" fill="#1F5673"/>
    <path d="M200 190v34h30v-34M380 190v34h30v-34M170 190c-20 6-36 0-40-10"/>`, 5),
  // Sleep app: a sun asleep behind a hill, on a rounded card.
  "sleepy-sun": svg(400, 250, null, "#E2701E", `
    <rect x="2" y="2" width="396" height="246" rx="14" fill="#FFE07A" stroke="none"/>
    <path d="M0 150c80-70 260-90 400-50" stroke="#FFCC33" stroke-width="40"/>
    <circle cx="270" cy="150" r="60" fill="#FF8A2A" stroke="none"/>
    <path d="M246 140c6 6 14 6 20 0M282 140c6 6 14 6 20 0M262 166c6 6 16 6 22 0" stroke="#7A3510"/>
    <path d="M0 250V200c100-40 260-40 400 0v50z" fill="#F8F4EE" stroke="none"/>
    <path d="M130 60l6 18 18 6-18 6-6 18-6-18-18-6 18-6z" fill="#FFFFFF" stroke="none"/><path d="M170 112l4 10 10 4-10 4-4 10-4-10-10-4 10-4z" fill="#FFFFFF" stroke="none"/>`),
  // Analytics feature email: a desk with laptop, mug and paper balls.
  "desk-feet-up": svg(600, 380, "#EEEEE9", "#2B2B2B", `
    <path d="M150 210h330l-30 30H180z" fill="#E7B98A"/><path d="M180 240v120M450 240v120M470 220v110"/>
    <path d="M300 210l20-70h110l-10 70z" fill="#BFD4F2"/><path d="M290 212h140"/>
    <rect x="210" y="186" width="60" height="10" fill="#4B6B48"/><rect x="214" y="176" width="54" height="10" fill="#E9E2C8"/><rect x="208" y="166" width="58" height="10" fill="#3D5A8A"/>
    <path d="M440 176h26v30h-26zM466 184c12 0 12 16 0 16" fill="#FFFFFF"/>
    <path d="M90 290h60l-8 70h-44z" fill="#D7D7D2"/><circle cx="108" cy="286" r="10" fill="#FFFFFF"/><circle cx="128" cy="282" r="9" fill="#FFFFFF"/><circle cx="70" cy="350" r="9" fill="#FFFFFF"/>
    <circle cx="510" cy="90" r="26" fill="#E8742C"/><path d="M486 86c16 8 32 8 48 0M510 64v52"/>
    <path d="M470 120c-10 10-14 22-12 34M456 118c-8 8-12 18-12 30" stroke-dasharray="6 8"/>
    <path d="M520 150h50v40h-50zM530 190l-6 18M560 190l6 18"/><path d="M520 150c14 18 36 18 50 0"/>`),
  // Usage limit: a meter run all the way to the end.
  "meter-full": svg(400, 220, "#EEEEE9", "#2B2B2B", `
    <rect x="60" y="40" width="280" height="150" rx="10" fill="#FFFFFF"/><path d="M60 70h280"/><circle cx="80" cy="55" r="4"/><circle cx="96" cy="55" r="4"/>
    <rect x="90" y="110" width="220" height="30" rx="15"/><rect x="90" y="110" width="220" height="30" rx="15" fill="#F2A33A"/>
    <path d="M90 162h40M150 162h60" stroke="#B8B8B2"/>${burst(350, 40, 26)}${burst(48, 190, 18)}`),
  // Breakfast promo: two sauce sachets and three dishes.
  "sachets": svg(600, 200, null, "#1F4D35", `
    <g transform="rotate(-18 120 100)"><rect x="80" y="30" width="80" height="150" rx="6" fill="#6FA35A"/><path d="M80 52h80M80 158h80"/><path d="M120 80c-14 10-14 40 0 56 14-16 14-46 0-56z" fill="#D9452B"/></g>
    <g transform="rotate(22 480 100)"><rect x="440" y="30" width="80" height="150" rx="6" fill="#2B2B2B"/><path d="M440 52h80M440 158h80" stroke="#F4EBDD"/><path d="M480 80c-14 10-14 40 0 56 14-16 14-46 0-56z" fill="#D9452B" stroke="#F4EBDD"/></g>`),
  "dish-eggs": svg(200, 140, null, "#1F4D35", `<ellipse cx="100" cy="90" rx="80" ry="34" fill="#FFFFFF"/><ellipse cx="78" cy="84" rx="26" ry="16" fill="#FFF8E6"/><circle cx="78" cy="84" r="9" fill="#F2B233"/><ellipse cx="126" cy="94" rx="24" ry="14" fill="#FFF8E6"/><circle cx="126" cy="94" r="8" fill="#F2B233"/>${dash(150, 64, -0.5, 10)}`),
  "dish-toast": svg(200, 140, null, "#1F4D35", `<path d="M50 120V60c0-34 100-34 100 0v60z" fill="#E9C48A"/><path d="M62 112V64c0-22 76-22 76 0v48" fill="#8DB65A"/><circle cx="84" cy="80" r="5" fill="#1F4D35"/><circle cx="114" cy="92" r="5" fill="#1F4D35"/>`),
  "dish-bowl": svg(200, 140, null, "#1F4D35", `<path d="M30 70h140c0 50-30 60-70 60s-70-10-70-60z" fill="#FFFFFF"/><path d="M50 70c10-26 30-24 40-6 10-20 34-22 44 0 10-14 26-10 30 6" fill="#D9452B"/><path d="M120 20l30 46M136 18l24 46"/>`),
  // Language test offer: a round green bird, a speech bubble, a globe.
  "parrot": svg(220, 220, null, "#2B3A1E", `
    <ellipse cx="110" cy="122" rx="66" ry="74" fill="#58C24A"/><circle cx="86" cy="104" r="18" fill="#FFFFFF"/><circle cx="134" cy="104" r="18" fill="#FFFFFF"/>
    <circle cx="90" cy="106" r="7" fill="#2B3A1E"/><circle cx="130" cy="106" r="7" fill="#2B3A1E"/><path d="M100 128l10 14 10-14z" fill="#F2A33A"/>
    <path d="M60 60c10-20 26-20 30 0M130 60c4-20 20-20 30 0" fill="#58C24A"/><path d="M90 196v14M130 196v14" stroke="#F2A33A" stroke-width="6"/>`),
  "bubble-globe": svg(260, 200, null, "#2B3A1E", `
    <path d="M20 30h130c10 0 16 6 16 16v54c0 10-6 16-16 16H70l-30 26 6-26H20c-10 0-16-6-16-16V46c0-10 6-16 16-16z" fill="#BEE3FF"/>
    <circle cx="190" cy="140" r="50" fill="#BEE3FF"/><path d="M150 120c20 10 30-6 46 6 10 8 4 24 20 28M170 182c0-20 20-24 30-14" fill="#58C24A"/><path d="M190 190v6M160 196h60"/>`),
  // Courier milestone: a big 100 with bikes and parcels around it.
  "milestone-100": `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="360" viewBox="0 0 600 360">
    <rect width="600" height="360" fill="#4F8DF7"/><text x="300" y="300" text-anchor="middle" font-family="Arial Black, Arial, sans-serif" font-weight="900" font-size="230" fill="#FFFFFF">100</text>
    <g fill="none" stroke="#173A86" stroke-width="5" stroke-linecap="round" stroke-linejoin="round">
    <circle cx="70" cy="300" r="26"/><circle cx="160" cy="300" r="26"/><path d="M70 300l30-50h40l20 50M100 250l-10-20h24M140 250l10-24"/>
    <rect x="430" y="70" width="70" height="56" rx="4" fill="#E8742C"/><path d="M430 90h70M465 70v56"/>
    <rect x="510" y="200" width="60" height="48" rx="4" fill="#F2C14E"/><path d="M510 218h60M540 200v48"/></g>
    <circle cx="250" cy="70" r="16" fill="#173A86"/><path d="M234 130c0-40 32-40 32 0z" fill="#173A86"/></svg>`,
  // Travel survey: two calendar pages joined by a ribbon, a pen and arrows.
  "calendar-ribbon": svg(600, 280, null, "#0E3E45", `
    <rect x="90" y="50" width="170" height="170" rx="6" fill="#FFFFFF"/><path d="M90 90h170M90 130h170M90 170h170M132 90v130M174 90v130M216 90v130"/>
    ${[110, 150, 190].map((y, i) => `<path d="M${140 + i * 42} ${y - 10}l20 20M${160 + i * 42} ${y - 10}l-20 20" stroke="#4B3BFF"/>`).join("")}
    <path d="M110 50v-16M150 50v-16M190 50v-16M230 50v-16"/>
    <path d="M260 210c40 40 60-60 100-20s60 40 90-10" stroke-width="30" stroke="#0E3E45"/>
    <rect x="360" y="70" width="170" height="150" rx="6" fill="#CFF5A4"/><path d="M360 110h170M360 150h170M360 190h170M402 70v150M444 70v150M486 70v150"/>
    <path d="M420 160l20 20M440 160l-20 20M500 120l20 20M520 120l-20 20" stroke="#4B3BFF"/>
    <path d="M540 40l40-30 8 10-40 30-12 4z" fill="#CFF5A4"/>
    <path d="M60 220c-30-40-10-90 30-110M54 200l6 20 18-10" stroke="#4B3BFF"/>
    <path d="M560 250c20-30 10-70-20-90M548 154l-8 6 10 10" stroke="#4B3BFF"/>`),
  // Workshop reminder: a bold banner with sticker shapes.
  "one-day-banner": `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="320" viewBox="0 0 600 320">
    <rect width="600" height="320" fill="#F25A3C"/>
    <g stroke="#111111" stroke-width="5" stroke-linejoin="round" stroke-linecap="round">
    <path d="M70 30l-30 80h34l-20 70 70-100h-38l24-50z" fill="#7B4DFF"/>
    <path d="M470 90c-30-40-90-10-60 40l60 50 60-50c30-50-30-80-60-40z" fill="none" transform="translate(30 -30) scale(.8)"/>
    <circle cx="530" cy="290" r="60" fill="#7B4DFF"/><path d="M470 290h120M530 230v120M490 250c30 20 50 20 80 0" fill="none"/>
    <path d="M60 240l40 30-40 30z" fill="#111111"/></g>
    <text x="300" y="160" text-anchor="middle" font-family="Arial Black, Arial, sans-serif" font-size="58" fill="#111111">One day</text>
    <text x="300" y="226" text-anchor="middle" font-family="Arial Black, Arial, sans-serif" font-size="58" fill="#111111">to go</text>
    <path d="M220 240c20-10 40 10 60 0s40 10 60 0 40 10 60 0" fill="none" stroke="#7B4DFF" stroke-width="7" stroke-linecap="round"/></svg>`,
  // Student card: three choices.
  "choice-yes": svg(120, 120, null, "#111111", `<rect x="36" y="16" width="48" height="88" rx="8" fill="#26B5C9"/><path d="M46 60l12 14 26-34" stroke-width="6"/>`, 4),
  "choice-maybe": svg(120, 120, null, "#111111", `<rect x="22" y="22" width="76" height="76" rx="18" fill="#2DCC8F"/><path d="M48 50c0-16 26-16 26 0 0 12-14 10-14 24M60 86v2" stroke="#FFFFFF" stroke-width="6"/>`, 4),
  "choice-no": svg(120, 120, null, "#111111", `<circle cx="60" cy="60" r="38" fill="#E3F45A"/><path d="M44 44l32 32M76 44L44 76" stroke-width="6"/>`, 4),
  "shapes-strip": svg(600, 70, null, "#111111", `${[40, 160, 300, 450, 560].map((x, i) => (i % 2 ? `<path d="M${x} 20l18 26h-36z" fill="#26B5C9"/><path d="M${x - 10} 36h20M${x - 6} 42h12"/>` : `<path d="M${x} 14l14 20-14 20-14-20z" fill="#26B5C9"/>`)).join("")}`, 3),
  // Meditation lesson: a video still, a sunrise and an armchair.
  "lesson-video": svg(480, 270, "#FBE3B8", "#3C2F2F", `
    <path d="M0 210c80-40 160-30 240-10s170 20 240-10v80H0z" fill="#F2B880"/>
    <path d="M90 200V90c0-20 30-20 30 0v40c10-10 30-6 30 10v30M90 140c-10-10-30-6-30 10v30" fill="#E8742C"/>
    <path d="M330 220c-10-80 20-120 60-120s70 40 60 120z" fill="#8FD3C1"/><circle cx="372" cy="150" r="6" fill="#3C2F2F"/><circle cx="410" cy="150" r="6" fill="#3C2F2F"/>
    <circle cx="240" cy="135" r="40" fill="#5A5A6E" stroke="none"/><path d="M228 115v40l32-20z" fill="#FFFFFF" stroke="none"/>`),
  "spot-sunrise": svg(200, 130, null, "#3C2F2F", `<path d="M20 100h160"/><path d="M50 100a50 50 0 0 1 100 0z" fill="#F2B233"/>${dash(100, 30, -1.57, 14)}${dash(40, 54, -2.4, 14)}${dash(160, 54, -0.75, 14)}<ellipse cx="100" cy="112" rx="70" ry="8" fill="#FBE3B8" stroke="none"/>`),
  "spot-chair": svg(200, 130, null, "#3C2F2F", `<path d="M50 110V50c0-24 100-24 100 0v60" fill="#F2B880"/><path d="M36 70c0-10 20-10 20 0v40H36zM144 70c0-10 20-10 20 0v40h-20z" fill="#E8742C"/><path d="M56 86h88v24H56z" fill="#FBE3B8"/><path d="M40 110v12M160 110v12"/>`),
  // Referral: two sauce jars swapped over a table of line doodles.
  "jar-swap": svg(600, 400, "#EFC6F0", "#F0532A", `
    <path d="M40 60l140 140M60 40l140 140"/><circle cx="110" cy="300" r="46"/><path d="M90 290h10M120 280l6 10M110 316l12 4"/>
    <circle cx="470" cy="330" r="110"/><circle cx="470" cy="330" r="80"/>${[0, 1, 2, 3].map((i) => `<circle cx="${420 + i * 34}" cy="${300 + i * 16}" r="14" fill="#EFC6F0"/>`).join("")}<path d="M410 286l130 70"/>
    <path d="M300 40c20-20 50-10 50 14-24 10-44 4-50-14zM520 70c-20-10-40 10-30 30 20 0 34-14 30-30z"/>
    <g transform="rotate(-10 230 230)"><rect x="170" y="160" width="130" height="150" rx="16" fill="#E8F542"/><rect x="182" y="140" width="106" height="26" rx="6" fill="#B0204A" stroke="#B0204A"/><path d="M190 214h90M190 252h90"/><path d="M234 220c-12 6-12 24 0 30 12-6 12-24 0-30z" fill="#F0532A"/></g>
    <g transform="rotate(12 400 170)"><rect x="340" y="100" width="130" height="150" rx="16" fill="#7A2A8F" stroke="#7A2A8F"/><rect x="352" y="80" width="106" height="26" rx="6" fill="#E8F542"/><path d="M360 154h90M360 192h90" stroke="#E8F542"/><path d="M404 160c-12 6-12 24 0 30 12-6 12-24 0-30z" fill="#E8F542" stroke="#E8F542"/></g>`),
  // Review request: five hand-drawn stars.
  "five-stars": svg(500, 120, null, "#2F3A12", `${[0, 1, 2, 3, 4].map((i) => star(60 + i * 95, 62, 40, "#5E6B1E")).join("")}`, 4),
  "doodle-burst": svg(120, 120, null, "#2F3A12", `${burst(60, 60, 40)}`, 6),
  "loaf": svg(200, 140, null, "#2F3A12", `<path d="M30 110c-10-60 40-80 70-80s80 20 70 80z" fill="#E2B067"/><path d="M60 60l14 26M100 50l6 30M140 60l-10 26"/><path d="M20 112h160"/>`),
};

// Flat UI mockups, portraits, icons and gradient cards for the product,
// education and event layouts. These are drawn without the wobble filter.
const flat = (w, h, body, bg = null) =>
  `<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}" viewBox="0 0 ${w} ${h}">${bg ? `<rect width="${w}" height="${h}" fill="${bg}"/>` : ""}${body}</svg>`;
const avatar = (bg, skin, hair, shirt) => flat(120, 120, `
  <clipPath id="c"><circle cx="60" cy="60" r="60"/></clipPath><g clip-path="url(#c)"><rect width="120" height="120" fill="${bg}"/>
  <path d="M18 120c0-30 18-44 42-44s42 14 42 44z" fill="${shirt}"/><rect x="50" y="62" width="20" height="18" fill="${skin}"/>
  <ellipse cx="60" cy="50" rx="22" ry="26" fill="${skin}"/><path d="M36 48c0-24 14-32 26-32s24 8 24 30c-8-10-30-14-50 2z" fill="${hair}"/>
  <circle cx="52" cy="52" r="2.4" fill="#2B2B2B"/><circle cx="68" cy="52" r="2.4" fill="#2B2B2B"/><path d="M53 63c4 4 10 4 14 0" stroke="#2B2B2B" stroke-width="2.4" fill="none" stroke-linecap="round"/></g>`);
const pillTag = (x, y, label, fill, ink = "#1A1A1A") =>
  `<rect x="${x}" y="${y}" width="${label.length * 9 + 24}" height="26" rx="13" fill="${fill}"/><text x="${x + 12}" y="${y + 18}" font-family="Arial" font-size="14" fill="${ink}">${label}</text>`;
const line = (w, h, body, sw = 3, ink = "#2B2B2B") =>
  `<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}" viewBox="0 0 ${w} ${h}"><g fill="none" stroke="${ink}" stroke-width="${sw}" stroke-linecap="round" stroke-linejoin="round">${body}</g></svg>`;
const grad = (id, stops, x2 = 1, y2 = 1) =>
  `<linearGradient id="${id}" x1="0" y1="0" x2="${x2}" y2="${y2}">${stops.map((c, i) => `<stop offset="${i / (stops.length - 1)}" stop-color="${c}"/>`).join("")}</linearGradient>`;

Object.assign(art, {
  // Product round-up: a table view, a stack of templates and a slide.
  "ui-table": flat(560, 300, `
    <rect x="30" y="30" width="530" height="250" rx="8" fill="#FFFFFF" stroke="#E2E2EA"/><rect x="30" y="30" width="530" height="34" rx="8" fill="#F4F4F8"/>
    <rect x="46" y="42" width="12" height="12" rx="3" fill="#4CAF7A"/><rect x="66" y="44" width="60" height="8" rx="4" fill="#C9C9D4"/>
    ${[0, 1, 2, 3, 4].map((i) => `<path d="M30 ${100 + i * 34}h530" stroke="#EEEEF3"/>`).join("")}
    ${[0, 1, 2, 3].map((i) => `<rect x="50" y="${110 + i * 34}" width="90" height="9" rx="4" fill="#CFCFD8"/><circle cx="182" cy="${114 + i * 34}" r="8" fill="${["#F2B880", "#8FB8F2", "#C7A8F0", "#9AD3B0"][i]}"/><rect x="196" y="${110 + i * 34}" width="54" height="9" rx="4" fill="#E0E0E8"/><rect x="290" y="${110 + i * 34}" width="80" height="9" rx="4" fill="#E0E0E8"/><rect x="400" y="${108 + i * 34}" width="44" height="13" rx="4" fill="${["#9DDCA9", "#F7DD7A", "#F4B183", "#F28B8B"][i]}"/><rect x="470" y="${108 + i * 34}" width="70" height="13" rx="4" fill="${["#A9C5F5", "#A9E1B6", "#CFCFD8", "#A9C5F5"][i]}"/>`).join("")}
    <path d="M300 214l14 30 4-12 12-4z" fill="#E66AD6"/>${pillTag(318, 232, "Ana", "#F6B6EE")}`, "#F1F1F5"),
  "ui-templates": flat(300, 220, `
    <rect x="20" y="40" width="200" height="150" rx="6" fill="#F4F4F7" stroke="#E1E1E8"/><rect x="40" y="24" width="200" height="150" rx="6" fill="#FAFAFC" stroke="#E1E1E8"/>
    <rect x="60" y="10" width="200" height="150" rx="6" fill="#FFFFFF" stroke="#DADAE3"/><rect x="76" y="26" width="110" height="16" fill="#F7EAA0"/><rect x="76" y="46" width="110" height="16" fill="#B9E3C6"/><rect x="76" y="66" width="110" height="16" fill="#DCCBF5"/>
    ${[0, 1, 2, 3, 4].map((i) => `<rect x="${80 + i * 22}" y="110" width="14" height="14" fill="${["#F3C35A", "#79C79A", "#8FB8F2", "#F28BC4", "#C7A8F0"][i]}"/>`).join("")}
    <g transform="rotate(-8 230 40)"><rect x="180" y="26" width="104" height="30" fill="#AEE7F5" stroke="#2B7A99"/><text x="194" y="47" font-family="Arial" font-size="16" fill="#1E5670">Template</text></g>${pillTag(150, 150, "Dev", "#C9B8FF")}`),
  "ui-slides": flat(300, 220, `
    <rect x="70" y="10" width="160" height="26" rx="6" fill="#1F1F1F"/><text x="84" y="28" font-family="Arial" font-size="13" fill="#FFFFFF">Turn into slides</text>
    <rect x="20" y="50" width="270" height="160" fill="#1E8C7E"/><text x="34" y="86" font-family="Arial Black,Arial" font-size="20" fill="#FFFFFF">Our</text><text x="34" y="112" font-family="Arial Black,Arial" font-size="20" fill="#FFFFFF">Spring</text><text x="34" y="138" font-family="Arial Black,Arial" font-size="20" fill="#FFFFFF">Plan</text>
    <rect x="150" y="62" width="128" height="136" rx="8" fill="#F6E7CF"/><circle cx="214" cy="112" r="20" fill="#E9B98E"/><path d="M194 104c4-18 36-18 40 0-12-8-28-8-40 0z" fill="#3B2A20"/><path d="M180 198c0-34 16-48 34-48s34 14 34 48z" fill="#F2C14E"/>${pillTag(120, 56, "Sam", "#B9C8FF")}`),
  "avatar-1": avatar("#F7D046", "#E8B48C", "#4A3326", "#3A5BA0"),
  "avatar-2": avatar("#2F6FED", "#C98E6B", "#2B1E16", "#E9E4DA"),
  "avatar-3": avatar("#4F8DF7", "#F1C6A6", "#C98B3A", "#2E7D6A"),
  "avatar-4": avatar("#F28B6B", "#8D5A3E", "#1E1410", "#F4F1EA"),
  "avatar-5": avatar("#9B7BEA", "#E5B08F", "#7A4A2A", "#1F2B40"),
  "avatar-6": avatar("#3FB68B", "#F0C8A8", "#2E2E2E", "#E85D75"),
  // AI assistant: a field of rounded app tiles with simple glyphs.
  "app-tiles": flat(560, 300, `<rect width="560" height="300" rx="16" fill="#B9CFC8"/>
    ${[[60, 50], [170, 30], [290, 40], [410, 60], [500, 140], [90, 150], [210, 130], [340, 150], [450, 230], [130, 250], [270, 240], [380, 270], [20, 240]].map(([x, y], i) => `<g transform="rotate(${(i % 3) * 6 - 6} ${x + 34} ${y + 34})"><rect x="${x}" y="${y}" width="68" height="68" rx="16" fill="#FFFFFF"/>${[
      `<circle cx="${x + 34}" cy="${y + 34}" r="16" fill="#E8743C"/>`, `<path d="M${x + 18} ${y + 44}l16-24 16 24z" fill="#3D6FD8"/>`, `<rect x="${x + 18}" y="${y + 22}" width="32" height="24" rx="3" fill="none" stroke="#D94A3A" stroke-width="5"/>`,
      `<rect x="${x + 20}" y="${y + 20}" width="28" height="28" rx="4" fill="#1F1F1F"/>`, `<circle cx="${x + 34}" cy="${y + 34}" r="16" fill="#7B4DFF"/>`, `<path d="M${x + 20} ${y + 24}h28M${x + 20} ${y + 34}h28M${x + 20} ${y + 44}h18" stroke="#2B2B2B" stroke-width="5"/>`,
      `<circle cx="${x + 34}" cy="${y + 34}" r="16" fill="none" stroke="#2B2B2B" stroke-width="6"/>`][i % 7]}</g>`).join("")}`),
  "icon-files": line(80, 80, `<path d="M20 18h26l12 12v34H20z"/><path d="M46 18v12h12"/><path d="M14 24v46h34"/><path d="M8 30v46h34"/>`),
  "icon-inbox": line(80, 80, `<rect x="10" y="18" width="60" height="46" rx="3"/><path d="M10 30h60M22 40h36M22 48h36M22 56h20M40 18v12"/>`),
  "icon-globe": line(80, 80, `<circle cx="40" cy="40" r="28"/><ellipse cx="40" cy="40" rx="12" ry="28"/><path d="M12 40h56M16 26h48M16 54h48"/>`),
  "icon-pencil": line(80, 80, `<path d="M18 60l6-18 30-30 12 12-30 30z"/><path d="M48 18l12 12"/><path d="M14 70c10-6 14 4 22-2s10 2 16-2"/>`),
  "icon-toolbox": line(80, 80, `<rect x="10" y="28" width="60" height="38" rx="4"/><path d="M30 28v-8h20v8M10 42h60"/><rect x="24" y="38" width="6" height="10"/><rect x="50" y="38" width="6" height="10"/>`),
  "icon-chart": line(80, 80, `<path d="M12 12v56h58"/><path d="M20 56l14-18 12 8 18-24"/><circle cx="34" cy="38" r="3"/><circle cx="46" cy="46" r="3"/><circle cx="64" cy="22" r="3"/>`, 2, "#D9E4E6"),
  "icon-cake": line(80, 80, `<rect x="16" y="40" width="48" height="26"/><path d="M16 50c8 6 16-6 24 0s16-6 24 0"/><path d="M26 40V28M40 40V28M54 40V28"/><path d="M26 22v-2M40 22v-2M54 22v-2"/>`, 2, "#D9E4E6"),
  "icon-ticket": line(80, 80, `<path d="M12 26h56v10c-6 0-6 10 0 10v10H12V46c6 0 6-10 0-10z"/><path d="M48 26v30" stroke-dasharray="3 4"/>`, 2, "#D9E4E6"),
  // Savings education: a blue banner with a magnifier, notes and sparkles.
  "savings-banner": flat(600, 200, `<rect width="600" height="200" fill="#2E5BFF"/>
    <circle cx="200" cy="80" r="56" fill="none" stroke="#D2D6E0" stroke-width="18"/><circle cx="200" cy="80" r="30" fill="#0B0B0B"/><path d="M200 50a30 30 0 0 1 0 60z" fill="#2E5BFF"/><path d="M190 80h20M200 70v20" stroke="#FFFFFF" stroke-width="4"/>
    <path d="M158 122l-40 40" stroke="#D2D6E0" stroke-width="20" stroke-linecap="round"/>
    ${[[330, 70], [510, 150], [40, 90]].map(([x, y]) => `<rect x="${x}" y="${y}" width="44" height="60" fill="#D2D6E0"/><path d="M${x + 8} ${y + 14}h28M${x + 8} ${y + 24}h28M${x + 8} ${y + 34}h28M${x + 8} ${y + 44}h18" stroke="#0B0B0B" stroke-width="3"/>`).join("")}
    ${[[290, 40, 1], [440, 190, 1], [80, 30, 0]].map(([x, y]) => `<circle cx="${x}" cy="${y}" r="20" fill="none" stroke="#FFFFFF" stroke-width="3"/><text x="${x}" y="${y + 7}" text-anchor="middle" font-family="Arial" font-weight="bold" font-size="20" fill="#FFFFFF">£</text>`).join("")}
    ${[[110, 30, 14, "#0B0B0B"], [400, 60, 18, "#FFFFFF"], [470, 110, 16, "#0B0B0B"], [560, 50, 14, "#0B0B0B"], [60, 170, 18, "#FFFFFF"], [380, 150, 12, "#FFFFFF"]].map(([x, y, r, c]) => `<path d="M${x} ${y - r}C${x + 2} ${y - 2} ${x + 2} ${y - 2} ${x + r} ${y}C${x + 2} ${y + 2} ${x + 2} ${y + 2} ${x} ${y + r}C${x - 2} ${y + 2} ${x - 2} ${y + 2} ${x - r} ${y}C${x - 2} ${y - 2} ${x - 2} ${y - 2} ${x} ${y - r}z" fill="${c}"/>`).join("")}`),
  "num-1": flat(64, 64, `<circle cx="32" cy="32" r="30" fill="#2E5BFF"/><text x="32" y="42" text-anchor="middle" font-family="Arial" font-size="30" fill="#FFFFFF">1</text>`),
  "num-2": flat(64, 64, `<circle cx="32" cy="32" r="30" fill="#2E5BFF"/><text x="32" y="42" text-anchor="middle" font-family="Arial" font-size="30" fill="#FFFFFF">2</text>`),
  "num-3": flat(64, 64, `<circle cx="32" cy="32" r="30" fill="#2E5BFF"/><text x="32" y="42" text-anchor="middle" font-family="Arial" font-size="30" fill="#FFFFFF">3</text>`),
  // Team invite: two figures with a horn and a flag, in flat blues.
  "team-parade": flat(500, 240, `<rect x="90" y="70" width="320" height="110" rx="55" fill="#EAF1FD"/>
    <path d="M110 230l30-80h60l20 80z" fill="#7FA8F5"/><circle cx="168" cy="96" r="20" fill="#FFFFFF" stroke="#2F6FED" stroke-width="3"/><path d="M150 96c0-24 36-24 36 0z" fill="#2F6FED"/><path d="M140 150l40-40h40" stroke="#FFFFFF" stroke-width="14" stroke-linecap="round"/>
    <path d="M190 100h60l20-14v34l-20-14" fill="#7FE0C2"/>
    <path d="M270 230l26-90h60l24 90z" fill="#2F6FED"/><circle cx="322" cy="88" r="20" fill="#FFFFFF" stroke="#2F6FED" stroke-width="3"/><path d="M302 92c0-30 44-30 40 6-6-10-20-14-40-6z" fill="#1F55C8"/>
    <path d="M370 40v140" stroke="#7FE0C2" stroke-width="4"/><circle cx="370" cy="36" r="7" fill="#7FE0C2"/><path d="M330 150l40-50" stroke="#FFFFFF" stroke-width="14" stroke-linecap="round"/>
    ${[[80, 60], [430, 60], [100, 180], [420, 160], [250, 30]].map(([x, y], i) => `<path d="M${x} ${y}l${i % 2 ? 8 : -8} -12" stroke="#B9CBEA" stroke-width="3" stroke-linecap="round"/>`).join("")}`),
  // Opt-in: a pocket watch that has grown arms and legs, in a pink circle.
  "clock-walker": svg(300, 300, null, "#1A1A1A", `
    <circle cx="150" cy="150" r="140" fill="#F7C6B8" stroke="none"/>
    <path d="M150 50v10"/><rect x="100" y="70" width="100" height="110" rx="34" fill="#FFFFFF"/><circle cx="150" cy="125" r="34"/><path d="M150 125v-22M150 125l16 10"/>
    ${[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11].map((i) => { const a = (Math.PI / 6) * i; return `<path d="M${(150 + 28 * Math.cos(a)).toFixed(1)} ${(125 + 28 * Math.sin(a)).toFixed(1)}L${(150 + 32 * Math.cos(a)).toFixed(1)} ${(125 + 32 * Math.sin(a)).toFixed(1)}" stroke-width="2"/>`; }).join("")}
    <path d="M100 110c-30-10-40-40-40-60M60 50l-8-10M60 50l8-10M200 130c30 0 50 20 56 40M256 170l10 4M256 170l2 10" stroke-width="7"/>
    <path d="M130 180l-30 60-20 4M170 180l30 50 24-4" stroke-width="8"/>`, 5),
  // Sales playbook: gradient cards with a bar chart and a ranked list.
  "chart-card": flat(560, 280, `<defs>${grad("g", ["#BFD7EA", "#F4D9C6", "#E8A5A5"], 0.3, 1)}</defs><rect width="560" height="280" rx="14" fill="url(#g)"/>
    <text x="28" y="40" font-family="Arial" font-size="14" letter-spacing="1" fill="#2B2B2B">REPLY RATES</text>
    ${[[50, 170, "12%", "No follow-ups"], [200, 110, "38%", "One follow-up"], [350, 40, "61%", "Three follow-ups"]].map(([x, y, n, l]) => `<rect x="${x}" y="${y + 70}" width="160" height="${210 - y}" rx="10" fill="#FFFFFF" fill-opacity=".45" stroke="#FFFFFF" stroke-opacity=".8"/><text x="${x}" y="${y + 36}" font-family="Arial" font-size="30" fill="#1A1A1A">${n}</text><text x="${x}" y="${y + 58}" font-family="Arial" font-size="14" fill="#3A3A3A">${l}</text>`).join("")}`),
  "list-card": flat(240, 320, `<defs>${grad("g2", ["#9CC6DB", "#E9C9C2", "#D99A8B"], 0.2, 1)}</defs><rect width="240" height="320" rx="12" fill="url(#g2)"/>
    <text x="18" y="34" font-family="Arial" font-size="12" letter-spacing="1" fill="#1F1F1F">OPEN RATES BY SUBJECT</text>
    ${[150, 130, 110, 90].map((w, i) => `<rect x="18" y="${60 + i * 38}" width="${w + 40}" height="26" rx="8" fill="#FFFFFF" fill-opacity=".75"/><text x="30" y="${77 + i * 38}" font-family="Arial" font-size="11" fill="#3A3A3A">Subject ${"DABC"[i]}</text>`).join("")}
    <rect x="18" y="240" width="60" height="60" rx="8" fill="#F2DCCB"/><circle cx="48" cy="262" r="12" fill="#C98E6B"/><path d="M30 300c0-20 36-20 36 0z" fill="#3A5BA0"/>
    <text x="90" y="266" font-family="Arial" font-size="10" fill="#3A3A3A">Masterclass with</text><text x="90" y="282" font-family="Arial" font-size="12" fill="#1F1F1F">Rosa Lind</text>`),
  "deals-card": flat(240, 240, `<defs>${grad("g3", ["#F2C9C0", "#E6A9B8", "#C7B8E8"], 0.2, 1)}</defs><rect width="240" height="240" rx="12" fill="url(#g3)"/>
    ${pillTag(16, 16, "Live", "#E8473C", "#FFFFFF")}<circle cx="120" cy="120" r="44" fill="#FFFFFF" fill-opacity=".6"/><path d="M108 100v40l32-20z" fill="#1F1F1F"/>`),
  // Conference: a loud gradient banner with the event name.
  "event-gradient": flat(600, 150, `<defs><linearGradient id="e" x1="0" y1="0" x2="1" y2="0.3"><stop offset="0" stop-color="#7B5CFF"/><stop offset=".3" stop-color="#FF5FA2"/><stop offset=".55" stop-color="#FFD84D"/><stop offset=".8" stop-color="#4DE1A1"/><stop offset="1" stop-color="#3FA9FF"/></linearGradient>
    <radialGradient id="r" cx=".3" cy=".6" r=".6"><stop offset="0" stop-color="#FF3D6E" stop-opacity=".9"/><stop offset="1" stop-color="#FF3D6E" stop-opacity="0"/></radialGradient></defs>
    <rect width="600" height="150" rx="14" fill="url(#e)"/><rect width="600" height="150" rx="14" fill="url(#r)"/>
    <text x="36" y="92" font-family="Georgia" font-size="44" fill="#FFFFFF">Northstar Live</text>
    <rect x="404" y="54" width="170" height="44" rx="22" fill="none" stroke="#FFFFFF" stroke-width="2"/><text x="489" y="83" text-anchor="middle" font-family="Georgia" font-size="22" fill="#FFFFFF">Register now</text>`),
  "speaker-strip": flat(600, 110, `${["#F28B6B", "#9B7BEA", "#4DE1A1", "#FFD84D", "#FF5FA2", "#3FA9FF"].map((c, i) => `<g transform="translate(${-30 + i * 106} 0)"><rect width="96" height="110" rx="10" fill="${c}"/><path d="M14 110c0-30 14-42 34-42s34 12 34 42z" fill="${["#1F2B40", "#E9E4DA", "#2E7D6A", "#3A5BA0", "#F4F1EA", "#2B2B2B"][i]}"/><ellipse cx="48" cy="44" rx="18" ry="22" fill="${["#E8B48C", "#C98E6B", "#F1C6A6", "#8D5A3E", "#E5B08F", "#F0C8A8"][i]}"/><path d="M28 40c0-20 12-26 22-26s20 8 18 26c-6-8-24-10-40 0z" fill="${["#4A3326", "#2B1E16", "#C98B3A", "#1E1410", "#7A4A2A", "#2E2E2E"][i]}"/></g>`).join("")}`),
});

// Batch two: dithered portrait, brewer icons, stat cards, launch tiles and more.
const pixelPortrait = (fg, bg) => {
  // A 48x48 dither of an original head-and-shoulders silhouette at a laptop.
  const N = 48, s = 8; let cells = "";
  for (let y = 0; y < N; y++) for (let x = 0; x < N; x++) {
    const dx = x - 24, dy = y - 16;
    const head = dx * dx / 64 + dy * dy / 90 < 1;
    const body = y > 26 && Math.abs(dx) < 4 + (y - 26) * 0.9 && y < 40;
    const desk = y >= 38 && y < 41;
    const laptop = y > 30 && y < 38 && x > 30 && x < 44 && (x - 30) > (38 - y) * 0.6;
    const shade = head ? ((x + y) % 3 ? 1 : 0) : body ? ((x * y) % 4 ? 1 : 0) : (desk || laptop) ? 1 : ((x + 2 * y) % 7 === 0 ? 1 : 0);
    cells += `<rect x="${x * s}" y="${y * s}" width="${s - 1}" height="${s - 1}" fill="${shade ? fg : bg}"/>`;
  }
  return flat(N * s, N * s, cells, bg);
};
const statCard = (n, label, c1, c2) => flat(240, 280, `<defs>${grad("s" + n.replace(/\W/g, ""), [c1, c2, "#FFF27A"], 0.5, 1)}</defs>
  <rect width="240" height="280" rx="18" fill="url(#s${n.replace(/\W/g, "")})"/><text x="18" y="200" font-family="Arial" font-size="64" fill="#1A1A1A">${n}</text>
  ${label.map((l, i) => `<text x="20" y="${228 + i * 15}" font-family="Arial" font-size="12" fill="#1A1A1A">${l}</text>`).join("")}`);
const tile = (x, y, fill, glyph) => `<rect x="${x}" y="${y}" width="140" height="140" rx="26" fill="${fill}"/>${glyph}`;

Object.assign(art, {
  "pixel-portrait": pixelPortrait("#E3EA4A", "#8E6BD8"),
  "brewers": svg(600, 150, null, "#2B2B2B", `
    <rect x="20" y="40" width="80" height="80" rx="8" fill="#F28B6B"/><rect x="30" y="52" width="60" height="16" fill="#FFFFFF"/><path d="M50 120v16M70 120v16"/>
    <path d="M140 40h70l-6 90h-58z" fill="#BFE3EE"/><path d="M150 70h50" /><rect x="150" y="30" width="50" height="12" fill="#F2B233"/>
    <path d="M260 40h70v20l-20 20v50h-30V80l-20-20z" fill="#E8574A"/><rect x="278" y="96" width="34" height="24" fill="#F2B233"/>
    <path d="M380 40h70l-35 50z" fill="#FFFFFF"/><path d="M395 90h40l10 46h-60z" fill="#8B5A3C"/>
    <rect x="500" y="40" width="50" height="90" rx="6" fill="#BFE3EE"/><rect x="496" y="30" width="58" height="14" rx="4" fill="#E8574A"/><path d="M500 90h50" />`),
  "ui-editor": flat(320, 220, `<rect x="40" y="30" width="230" height="160" rx="8" fill="#FFFFFF"/><rect x="40" y="30" width="230" height="22" rx="8" fill="#EDEAF6"/>
    <rect x="60" y="70" width="110" height="90" rx="6" fill="#F2A33A"/><circle cx="115" cy="115" r="14" fill="#FFE07A"/>
    ${[0, 1, 2].map((i) => `<rect x="186" y="${70 + i * 32}" width="70" height="22" rx="4" fill="#F7C873"/>`).join("")}
    <path d="M290 30l6 14 14 6-14 6-6 14-6-14-14-6 14-6z" fill="#FFFFFF"/>`),
  "icon-brush": flat(80, 80, `<rect width="80" height="80" rx="12" fill="#FFF3D6" stroke="#F2C14E" stroke-width="2"/><path d="M22 58l26-26 8 8-26 26-12 4z" fill="#8E6BD8"/><circle cx="54" cy="26" r="10" fill="#F2A33A"/>`),
  "icon-toggles": flat(80, 80, `<rect width="80" height="80" rx="12" fill="#EAF7EE" stroke="#4CAF7A" stroke-width="2"/>${[0, 1, 2].map((i) => `<rect x="16" y="${16 + i * 18}" width="12" height="12" rx="2" fill="${i < 2 ? "#4CAF7A" : "#FFFFFF"}" stroke="#4CAF7A"/><rect x="34" y="${18 + i * 18}" width="30" height="8" rx="4" fill="#CFD8D2"/>`).join("")}`),
  "icon-blocks": flat(80, 80, `<rect width="80" height="80" rx="12" fill="#F1ECFB" stroke="#8E6BD8" stroke-width="2"/>${[[16, 44, "#F28BC4"], [30, 44, "#F2C14E"], [44, 44, "#F2C14E"], [16, 30, "#F28BC4"], [30, 30, "#8E6BD8"], [44, 16, "#8E6BD8"], [58, 16, "#8E6BD8"], [16, 58, "#F28BC4"], [30, 58, "#F2C14E"], [44, 58, "#F2C14E"]].map(([x, y, c]) => `<rect x="${x}" y="${y}" width="12" height="12" fill="${c}"/>`).join("")}`),
  "house-dots": flat(120, 30, `${[["#D9541E", 10], ["#2F6FED", 46], ["#1E7D5A", 82]].map(([c, x]) => `<path d="M${x} 28V14l12-10 12 10v14z" fill="${c}"/>`).join("")}`),
  "icon-heart-house": flat(64, 64, `<path d="M8 58V28L32 8l24 20v30z" fill="#FFFFFF"/><path d="M32 46c-10-8-14-12-10-18 3-4 8-3 10 1 2-4 7-5 10-1 4 6 0 10-10 18z" fill="#0B2A7A"/>`),
  "stat-6": statCard("9%", ["of teams think the tool could", "replace someone today"], "#FFE14D", "#F5B8C8"),
  "stat-97": statCard("88%", ["of teams have tried", "the tool at least once"], "#FFE14D", "#C9B8F5"),
  "stat-84": statCard("71%", ["use it first for", "research and prep"], "#F5B8C8", "#FFE14D"),
  "launch-tiles": flat(500, 300, `<defs>${grad("lt", ["#FBC9A8", "#FBE7D6"], 0.2, 1)}</defs><rect width="500" height="300" rx="20" fill="url(#lt)"/>
    <path d="M0 230c120 50 260 40 360-30" stroke="#E8743C" stroke-width="2" stroke-dasharray="3 5" fill="none"/><path d="M440 150l30-110" stroke="#E8743C" stroke-width="2" stroke-dasharray="3 5"/><path d="M462 40l12-4 2 12" fill="none" stroke="#E8743C" stroke-width="2"/>
    ${tile(60, 80, "#24413A", `<path d="M95 130l35-20 35 20-35 20z" fill="#A8E6A1"/><path d="M95 150l35 20 35-20" stroke="#A8E6A1" stroke-width="10" fill="none"/>`)}
    <path d="M200 150h100" stroke="#1F1F1F" stroke-width="2" stroke-dasharray="2 4"/><circle cx="250" cy="150" r="5" fill="#1F1F1F"/>
    ${tile(300, 80, "#FF6A1A", `<rect x="335" y="110" width="70" height="10" fill="#FFFFFF"/><rect x="335" y="126" width="70" height="10" fill="#FFFFFF"/><path d="M335 142h70v46l-35-18-35 18z" fill="#FFFFFF"/>`)}`),
  "emoji-recycle": flat(64, 64, `<rect width="64" height="64" rx="10" fill="#FFD98A"/><path d="M20 40l12-20 12 20z" fill="none" stroke="#2E9E4A" stroke-width="5" stroke-linejoin="round"/>`),
  "emoji-chart": flat(64, 64, `<rect width="64" height="64" rx="10" fill="#D8C8F5"/><rect x="14" y="14" width="36" height="36" fill="#FFFFFF"/><path d="M18 44l10-12 8 6 12-18" stroke="#D9452B" stroke-width="3" fill="none"/>`),
  "emoji-shake": flat(64, 64, `<rect width="64" height="64" rx="10" fill="#BFE0F5"/><path d="M14 34l10-8 10 6 8-6 8 8-12 12-8-4-6 4z" fill="#F2B233" stroke="#B5781E" stroke-width="2"/>`),
  "shield-mail": svg(600, 220, null, "#1A1A1A", `
    <path d="M120 40l70 26v50c0 50-40 80-70 90-30-10-70-40-70-90V66z" fill="#FFE0E0"/><path d="M95 120l20 20 36-44" stroke="#E8473C" stroke-width="6"/>
    <rect x="250" y="50" width="160" height="120" rx="10" fill="#FFFFFF"/><path d="M262 70h136M262 92h100M262 112h120M262 132h80"/>
    <path d="M450 80l60-30 60 30v80H450z" fill="#F7E6F7"/><path d="M450 80l60 44 60-44"/>
    ${star(230, 40, 14, "#1A1A1A")}${star(580, 190, 12, "#1A1A1A")}${burst(440, 190, 20)}`),
  "icon-checkbox": flat(48, 48, `<rect x="4" y="4" width="40" height="40" rx="8" fill="#FFFFFF" stroke="#1A1A1A" stroke-width="3"/><rect x="11" y="11" width="26" height="26" rx="4" fill="#F7C6D0"/><path d="M15 24l7 7 13-14" stroke="#E8473C" stroke-width="4" fill="none"/>`),
  "icon-chat": flat(48, 48, `<path d="M6 10h36v24H20l-10 8v-8H6z" fill="#FFFFFF" stroke="#1A1A1A" stroke-width="3"/><path d="M14 18h20M14 26h14" stroke="#E8473C" stroke-width="3"/>`),
  "ui-dark-sessions": flat(560, 200, `<rect width="560" height="200" rx="10" fill="#1C1C1A"/><rect x="0" y="0" width="560" height="26" rx="10" fill="#2A2A27"/>
    ${[0, 1, 2].map((i) => `<circle cx="${16 + i * 14}" cy="13" r="4" fill="#5A5A55"/>`).join("")}
    <rect x="20" y="44" width="180" height="140" rx="6" fill="#262623"/>${[0, 1, 2, 3, 4].map((i) => `<rect x="32" y="${58 + i * 24}" width="${150 - i * 14}" height="10" rx="5" fill="${i === 1 ? "#D97757" : "#4A4A45"}"/>`).join("")}
    <text x="230" y="70" font-family="Courier New" font-size="15" fill="#E8E6DC">Welcome back, Sam.</text>${[0, 1, 2, 3].map((i) => `<rect x="230" y="${90 + i * 22}" width="${300 - i * 40}" height="9" rx="4" fill="#3A3A36"/>`).join("")}`),
  "botanical": svg(200, 200, null, "#CFC8B0", `
    <path d="M100 190V40M100 120c-30-10-50-40-50-70M100 100c30-10 46-36 44-66M100 150c-24-4-44-20-52-44M100 140c26-6 40-24 46-48"/>
    ${[[50, 50], [144, 34], [48, 106], [146, 92], [100, 40]].map(([x, y]) => `<path d="M${x} ${y}c-10-14-4-26 0-30 4 4 10 16 0 30z"/>`).join("")}`, 2),
});

Object.assign(art, {
  "torn-stripe": flat(600, 24, `<rect width="600" height="24" fill="#1F4D2E"/>${[0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11].map((i) => `<path d="M${i * 52} 24l${10 + (i % 3) * 6} -14 ${22 + (i % 2) * 8} 4 8 10z" fill="${["#E8743C", "#E3EA4A", "#E8743C", "#C9D94A"][i % 4]}"/>`).join("")}`),
  "fact-mark": flat(48, 48, `<circle cx="24" cy="24" r="20" fill="none" stroke="#E3EA4A" stroke-width="4"/><path d="M16 30c0-12 18-16 18-4 0 6-8 6-8 0" fill="none" stroke="#E3EA4A" stroke-width="4" stroke-linecap="round"/>`),
  "map-outline": line(160, 120, `<path d="M14 24l40-14 50 14 44-14v86l-44 14-50-14-40 14z"/><path d="M54 10v86M104 24v86"/>`, 7, "#1E1A4A"),
  "hand-ok": svg(140, 120, null, "#3A2A20", `<path d="M40 110V70c0-10 14-10 14 0v-30c0-10 14-10 14 0v26c0-10 14-10 14 0v8c4-6 14-4 14 4v20c0 20-14 32-32 32s-24-12-24-30z" fill="#F2C2A0"/><circle cx="104" cy="30" r="12" fill="#E8473C" stroke="none"/><path d="M96 20c4-8 16-8 16 2" stroke="#E8473C"/>`, 3),
  "megaphone-city": svg(600, 180, null, "#1A1A1A", `<rect x="250" y="40" width="70" height="140" fill="#4A4A4A"/>${[0, 1, 2, 3, 4, 5].map((i) => `<path d="M262 ${56 + i * 20}h46" stroke="#8A8A8A" stroke-width="3"/>`).join("")}
    <path d="M300 90l90-40v100l-90-40z" fill="#E8473C"/><rect x="280" y="86" width="24" height="32" fill="#FFFFFF"/>${burst(420, 70, 30)}<path d="M150 50c10-12 30-12 34 0 12-4 20 8 12 14h-46c-8-4-6-12 0-14z" fill="#FFFFFF"/><path d="M450 30c10-12 30-12 34 0 12-4 20 8 12 14h-46c-8-4-6-12 0-14z" fill="#FFFFFF"/>`, 3),
  "bulb": svg(80, 80, null, "#E3EA4A", `<path d="M28 50c-10-8-10-30 12-30s22 22 12 30v8H28z"/><path d="M30 64h20M32 70h16"/>${dash(40, 10, -1.57, 6)}${dash(16, 22, -2.4, 6)}${dash(64, 22, -0.75, 6)}`, 5),
  "critters": flat(600, 70, `${[[30, "cat"], [120, "dog"], [210, "rabbit"], [300, "deer"], [390, "bird"], [480, "fox"], [560, "hen"]].map(([x, k]) => ({
      cat: `<path d="M${x} 70v-30l6-14 6 10h12l6-10 6 14v30z"/>`, dog: `<path d="M${x} 70v-24h30l10-14 8 6-4 16v16z"/>`, rabbit: `<path d="M${x} 70c0-20 8-28 20-28l2-28 6 2-2 26c10 4 14 14 14 28z"/>`,
      deer: `<path d="M${x} 70l4-30h30l8-20-6-14 10 8 4-8 2 12-8 22-4 30z"/>`, bird: `<path d="M${x} 60c10-20 30-24 44-14l10-4-6 10c0 16-16 22-48 8z"/>`, fox: `<path d="M${x} 70c4-18 14-26 28-26l10-18 4 14 8-12 0 18c-6 10-14 16-20 24z"/>`, hen: `<path d="M${x} 70c-4-20 6-34 22-30l4-10 6 8c6 10 4 26-6 32z"/>` })[k]).join("")}`.replace(/<path /g, '<path fill="#2A2AE0" '), null),
});

// Batch four: product cards, comparison art, year-in-numbers icons and more.
const flower = (cx, cy, c, k) => k ? `<circle cx="${cx}" cy="${cy}" r="10" fill="${c}"/>${[0, 1, 2, 3, 4, 5].map((i) => `<ellipse cx="${cx + 18 * Math.cos(i * Math.PI / 3)}" cy="${cy + 18 * Math.sin(i * Math.PI / 3)}" rx="9" ry="6" transform="rotate(${i * 60} ${cx + 18 * Math.cos(i * Math.PI / 3)} ${cy + 18 * Math.sin(i * Math.PI / 3)})" fill="${c}"/>`).join("")}` : `${[0, 1, 2, 3].map((i) => `<path d="M${cx} ${cy}l${(24 * Math.cos(i * Math.PI / 4)).toFixed(1)} ${(24 * Math.sin(i * Math.PI / 4)).toFixed(1)}M${cx} ${cy}l${(-24 * Math.cos(i * Math.PI / 4)).toFixed(1)} ${(-24 * Math.sin(i * Math.PI / 4)).toFixed(1)}" stroke="${c}" stroke-width="6" stroke-linecap="round"/>`).join("")}`;
Object.assign(art, {
  "prompt-card": flat(600, 340, `<defs><linearGradient id="pc" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#3FA9F5"/><stop offset=".55" stop-color="#F2A3B3"/><stop offset=".75" stop-color="#5CC46A"/><stop offset="1" stop-color="#2E8A4E"/></linearGradient></defs>
    <rect width="600" height="340" rx="16" fill="url(#pc)"/><rect x="120" y="64" width="360" height="210" rx="20" fill="#0E1420" fill-opacity=".92"/>
    <rect x="138" y="84" width="62" height="26" rx="6" fill="#1E3A5A"/><text x="144" y="103" font-family="Arial" font-size="17" fill="#5CB8FF">/skills</text><text x="210" y="103" font-family="Arial" font-size="17" fill="#FFFFFF">make a skill from</text><text x="138" y="130" font-family="Arial" font-size="17" fill="#FFFFFF">these page layouts</text>
    <path d="M138 230h324" stroke="#2A3242"/><text x="138" y="258" font-family="Arial" font-size="16" fill="#FFFFFF">Model · Fast ⌄</text>${[372, 402, 432].map((x) => `<rect x="${x}" y="240" width="24" height="24" rx="6" fill="#1E2636"/>`).join("")}`),
  "wire-shape": flat(600, 340, `<rect width="600" height="340" rx="16" fill="#060606"/><g fill="none" stroke="#9A9A9A" stroke-width="1.5">
    <path d="M240 90l50-20 100 20-50 20zM240 90v40l100 20v-40M390 90v40l-50 20M240 130l50 60 50-40M240 130v80l50 60v-80M290 190l50-40v80l-50 40M240 210l50-20"/></g>`),
  "dark-table": flat(600, 340, `<rect width="600" height="340" rx="16" fill="#050505"/><ellipse cx="330" cy="150" rx="200" ry="120" fill="#1A1A1A"/><rect x="100" y="70" width="400" height="210" rx="8" fill="#0E0E0E" stroke="#2A2A2A"/>
    <text x="140" y="96" font-family="Arial" font-size="12" fill="#DDDDDD">Question</text><text x="320" y="96" font-family="Arial" font-size="12" fill="#DDDDDD">Answer</text>
    ${["What is a list field?", "What can it hold?", "When should I use it?", "How do I show it?", "Who can edit it?"].map((q, i) => `<path d="M100 ${108 + i * 30}h400" stroke="#222"/><text x="140" y="${126 + i * 30}" font-family="Arial" font-size="11" fill="#CCCCCC">${q}</text><text x="320" y="${126 + i * 30}" font-family="Arial" font-size="11" fill="#888888">A list groups repeating…</text>`).join("")}`),
  "rate-compare": flat(600, 320, `<rect x="2" y="2" width="296" height="316" rx="14" fill="#F4F4F4" stroke="#DDDDDD"/><rect x="302" y="2" width="296" height="316" rx="14" fill="#9BEA74"/><rect x="296" y="0" width="8" height="320" fill="#2E8A5E"/>
    ${[150, 450].map((x, i) => `<ellipse cx="${x}" cy="150" rx="90" ry="80" fill="#FCE3D4"/><path d="M${x - 40} 150v-60l40-36 40 36v60z" fill="#4CB8F5"/><path d="M${x} 54l40 36v60h-40z" fill="#1E5FD8"/><path d="M${x - 24} 150v-30a8 8 0 0 1 16 0v30z" fill="#1A1A1A"/>
      ${i ? `<path d="M${x - 70} 196h50l20-14 20 14 20-20 20 14" stroke="#1E6E4E" stroke-width="5" fill="none"/>` : `<path d="M${x - 80} 190h130" stroke="#1E6E4E" stroke-width="5"/>`}<circle cx="${i ? x - 70 : x + 60}" cy="${i ? 196 : 190}" r="16" fill="#1E6E4E"/><text x="${i ? x - 70 : x + 60}" y="${i ? 202 : 196}" text-anchor="middle" font-family="Arial" font-weight="bold" font-size="16" fill="#9BEA74">%</text>
      <text x="${x}" y="270" text-anchor="middle" font-family="Arial" font-weight="bold" font-size="15" fill="#1A1A1A">${i ? "Variable rate" : "Fixed rate"}</text>`).join("")}`),
  "envelope-rest": svg(400, 300, null, "#2B3A35", `
    <rect x="80" y="130" width="240" height="150" rx="10" fill="#A9B8B2"/><path d="M80 140l120 80 120-80M80 270l90-60M320 270l-90-60"/>
    <path d="M140 140c0-30 30-40 60-40s60 10 60 40" fill="#C48A4E"/><circle cx="200" cy="96" r="30" fill="#FFFFFF"/><path d="M168 86c4-30 60-34 66 0-20-10-46-10-66 0z" fill="#2B3A35"/>
    <path d="M186 100c4 4 10 4 14 0M210 96h8"/><path d="M150 132c20-10 80-10 100 0" stroke-width="6" stroke="#FFFFFF"/><text x="244" y="70" font-family="Arial" font-weight="bold" font-size="30" fill="#2B3A35" stroke="none">?</text><path d="M270 64l26-8M272 78l24 2M270 90l20 10" stroke="#C48A4E"/>
    ${[0, 1, 2].map((r) => [0, 1, 2].map((c) => `<circle cx="${50 + c * 14}" cy="${200 + r * 14}" r="2.5" fill="#6B7A75" stroke="none"/><circle cx="${330 + c * 14}" cy="${150 + r * 14}" r="2.5" fill="#6B7A75" stroke="none"/>`).join("")).join("")}`),
  "bell": svg(80, 80, null, "#7A8A85", `<path d="M24 54c4-6 4-14 4-22 0-8 6-14 12-14s12 6 12 14c0 8 0 16 4 22z"/><path d="M34 60a6 6 0 0 0 12 0"/><path d="M14 30c2-8 6-12 10-14M66 30c-2-8-6-12-10-14"/>`, 5),
  "year-banner": flat(600, 320, `<rect width="600" height="320" fill="#0E4A2A"/>
    <rect x="130" y="80" width="340" height="190" rx="10" fill="#F2DC8A" stroke="#1A1A1A" stroke-width="3"/>${Array.from({ length: 22 }, (_, i) => `<circle cx="${150 + i * 14}" cy="80" r="6" fill="none" stroke="#C0306A" stroke-width="3"/>`).join("")}
    <text x="300" y="190" text-anchor="middle" font-family="Arial Black,Arial" font-size="86" fill="#F0532A" stroke="#1A1A1A" stroke-width="3">2026</text><text x="300" y="245" text-anchor="middle" font-family="Georgia" font-style="italic" font-size="40" fill="#1A1A1A">in numbers</text>
    <rect x="70" y="60" width="90" height="34" rx="17" fill="#3AC46A" stroke="#1A1A1A" stroke-width="3"/><rect x="430" y="210" width="110" height="34" rx="17" fill="#3AC46A" stroke="#1A1A1A" stroke-width="3"/>
    <path d="M440 70l60-40 10 18 30-30" stroke="#F2C14E" stroke-width="8" fill="none" stroke-linecap="round"/><circle cx="90" cy="250" r="44" fill="#F7C8D8" stroke="#1A1A1A" stroke-width="3"/><text x="90" y="246" text-anchor="middle" font-family="Georgia" font-style="italic" font-size="16" fill="#1A1A1A">A big</text><text x="90" y="266" text-anchor="middle" font-family="Georgia" font-style="italic" font-size="16" fill="#1A1A1A">year</text>`),
  "yr-mail": line(80, 70, `<rect x="10" y="12" width="60" height="44"/><path d="M10 12l30 26 30-26"/><path d="M10 56l14-14" stroke="#E8473C"/>`, 3, "#1A1A1A"),
  "yr-chat": line(80, 70, `<circle cx="40" cy="34" r="26"/><path d="M24 30h32M24 40h22M22 56l-6 10 14-6"/>`, 3, "#FFFFFF"),
  "yr-cart": line(80, 70, `<path d="M8 10h12l10 34h34l6-24H24"/><circle cx="34" cy="56" r="5"/><circle cx="60" cy="56" r="5"/>`, 3, "#1A1A1A"),
  "yr-dots": line(80, 70, `<circle cx="28" cy="20" r="10"/><circle cx="28" cy="48" r="10"/><circle cx="54" cy="48" r="10"/><path d="M50 20h16M58 12v16"/>`, 3, "#1A1A1A"),
  "yr-pie": line(80, 70, `<path d="M42 10a24 24 0 1 1-24 24h24z"/><path d="M12 64V40h10v24M28 64V48h10v16"/>`, 3, "#1A1A1A"),
  "yr-calendar": line(80, 70, `<rect x="12" y="10" width="56" height="50"/><path d="M12 22h56" stroke="#C0306A" stroke-width="8"/><path d="M22 34h8M36 34h8M50 34h8M22 46h8M36 46h8M50 46h8"/>`, 3, "#1A1A1A"),
  "mosaic": flat(600, 300, `${[["#F7D046", 0, 0, 1], ["#4F9FD0", 1, 0, 0], ["#F28BB8", 2, 0, 1], ["#3E5A2A", 3, 0, 0], ["#D9532E", 0, 1, 0], ["#F4EEE0", 1, 1, 1], ["#7A8A30", 2, 1, 0], ["#2A3A6A", 3, 1, 1]].map(([c, x, y, k]) => `<rect x="${x * 150}" y="${y * 150}" width="150" height="150" fill="${c}"/>${flower(x * 150 + 75, y * 150 + 75, ["#F4EEE0", "#F7D046", "#2A3A6A", "#F28BB8"][(x + y) % 4], k)}`).join("")}`),
  "big-year-digits": flat(600, 110, `${[["2", "#3E5A2A", "#F28BB8"], ["0", "#4F9FD0", "#F4EEE0"], ["2", "#F28BB8", "#F7D046"], ["6", "#B05A2A", "#F4EEE0"]].map(([d, bg, fg], i) => `<rect x="${i * 150}" width="150" height="110" fill="${bg}"/><text x="${i * 150 + 75}" y="92" text-anchor="middle" font-family="Arial" font-weight="bold" font-size="96" fill="${fg}">${d}</text>`).join("")}`),
  "compose-window": flat(600, 230, `<defs>${grad("cw", ["#F7EB3A", "#F2B0C8", "#9FE0F0"], 1, 1)}</defs><rect width="600" height="230" rx="16" fill="url(#cw)"/>
    <rect x="18" y="18" width="582" height="212" rx="12" fill="#F7F7F2"/><rect x="18" y="18" width="582" height="56" fill="#55554A"/><text x="44" y="54" font-family="Arial" font-size="19" fill="#D6D6CC">New message</text>
    <text x="44" y="112" font-family="Arial" font-size="18" fill="#9A9A90">To</text><rect x="90" y="96" width="150" height="20" fill="#E2E2DA"/><path d="M40 134h540" stroke="#E2E2DA"/>
    <text x="44" y="170" font-family="Arial" font-size="18" fill="#9A9A90">Subject</text><text x="130" y="170" font-family="Arial" font-size="18" fill="#2B2B2B">Please stop saying “just checking in”</text><path d="M466 152v24" stroke="#2B2B2B"/>`),
  "webinar-tile-1": flat(600, 300, `<defs>${grad("w1", ["#D8DCF8", "#C8CCF4"], 1, 0)}</defs><rect width="600" height="300" rx="12" fill="url(#w1)"/><rect x="370" width="230" height="300" fill="#9A6AE0"/>
    <circle cx="470" cy="150" r="70" fill="#7A4AD0" stroke="#C8B0F0" stroke-width="10"/><path d="M470 120l8 22 22 8-22 8-8 22-8-22-22-8 22-8z" fill="#FFFFFF"/><path d="M500 180l30 50 8-20 20-6z" fill="#111111"/>
    <rect x="30" y="30" width="110" height="26" rx="13" fill="#E3EA4A"/><text x="44" y="48" font-family="Arial" font-size="12" fill="#1A1A1A">● ON DEMAND</text>
    <text x="30" y="130" font-family="Arial" font-weight="bold" font-size="32" fill="#1A1A3A">Quick fixes are cheap.</text><text x="30" y="170" font-family="Arial" font-weight="bold" font-size="32" fill="#1A1A3A">Lost customers aren’t.</text>`),
  "webinar-tile-2": flat(600, 300, `<rect width="600" height="300" rx="12" fill="#2A0E4A"/><rect width="220" height="300" fill="#7A3AD0"/>
    ${[0, 1, 2, 3].map((i) => `<rect x="${110 - 80 + i * 20}" y="${150 - 80 + i * 20}" width="${160 - i * 40}" height="${160 - i * 40}" transform="rotate(45 110 150)" fill="${["#9A5AE0", "#C04AA0", "#E04A8A", "#E3EA4A"][i]}"/>`).join("")}
    <text x="420" y="60" text-anchor="middle" font-family="Courier New" font-size="12" fill="#C8B0F0">ON DEMAND</text><text x="420" y="140" text-anchor="middle" font-family="Arial" font-weight="bold" font-size="40" fill="#E3EA4A">Steady <tspan font-family="Georgia" font-style="italic">Hands</tspan></text>
    <text x="420" y="180" text-anchor="middle" font-family="Arial" font-size="16" fill="#FFFFFF">Automation with a person</text><text x="420" y="202" text-anchor="middle" font-family="Arial" font-size="16" fill="#FFFFFF">in charge of it.</text>`),
  "webinar-tile-3": flat(600, 300, `<rect width="600" height="300" rx="12" fill="#2A0E3A"/><rect x="20" y="0" width="8" height="300" fill="#E3EA4A"/>
    <text x="60" y="110" font-family="Arial" font-size="46" fill="#FFFFFF">The state of</text><text x="60" y="166" font-family="Arial" font-size="46" fill="#FFFFFF">customer</text><text x="60" y="222" font-family="Georgia" font-size="46" fill="#FFFFFF">expectations</text>
    <circle cx="330" cy="150" r="20" fill="#3AC46A"/><rect x="370" y="134" width="10" height="34" fill="#F2C14E"/><rect x="384" y="146" width="10" height="22" fill="#F2C14E"/><circle cx="430" cy="140" r="10" fill="#3A7AE0"/><path d="M420 168c0-16 20-16 20 0z" fill="#3A7AE0"/>`),
  "welcome-banner": svg(600, 300, "#3AA6F0", "#1A1A1A", `
    <text x="40" y="120" font-family="Impact,Arial Black" font-size="48" fill="#1A1A1A" stroke="none">WELCOME TO</text><text x="40" y="186" font-family="Impact,Arial Black" font-size="48" fill="#1A1A1A" stroke="none">OUR CREATOR</text><text x="40" y="252" font-family="Impact,Arial Black" font-size="48" fill="#1A1A1A" stroke="none">NEWSLETTER</text>
    ${[60, 46, 32, 18].map((r, i) => `<circle cx="520" cy="90" r="${r * 0.8}" fill="${i % 2 ? "#F4E3C8" : "#F2643A"}"/>`).join("")}<path d="M400 200l120-50-50 90-16-32z" fill="#F7D9A8"/><path d="M454 208l66-58"/>`, 3),
  "guide-card": flat(560, 300, `<rect width="560" height="300" rx="12" fill="#5DD07A"/><rect x="40" y="70" width="44" height="20" rx="4" fill="#3AA65A"/><text x="48" y="85" font-family="Arial" font-size="11" fill="#FFFFFF">Guide</text>
    <text x="40" y="150" font-family="Impact,Arial Black" font-size="40" fill="#1A2A1A">Launch your</text><text x="40" y="200" font-family="Impact,Arial Black" font-size="40" fill="#1A2A1A">newsletter</text><text x="40" y="250" font-family="Impact,Arial Black" font-size="40" fill="#1A2A1A">in 7 days</text>
    <rect x="350" y="40" width="210" height="260" fill="#FFFFFF"/><rect x="370" y="100" width="190" height="200" fill="#0E1A2A"/><text x="430" y="130" font-family="Arial" font-weight="bold" font-size="16" fill="#FFFFFF">Weekly notes</text>${[0, 1, 2, 3, 4, 5].map((i) => `<rect x="390" y="${160 + i * 20}" width="${90 - i * 6}" height="8" fill="#3A4A5A"/>`).join("")}`),
});

// Batch five: products, a podcast cover, a pony mascot, vases and charts.
const bottle = (x, cap, body, label) => `<rect x="${x + 20}" y="40" width="60" height="40" rx="6" fill="${cap}"/><rect x="${x}" y="78" width="100" height="190" rx="18" fill="${body}"/><text x="${x + 50}" y="180" text-anchor="middle" font-family="Arial Black,Arial" font-size="18" fill="#FFFFFF" transform="rotate(-20 ${x + 50} 175)">${label}</text>`;
Object.assign(art, {
  "sale-hero": flat(600, 330, `<rect width="600" height="330" fill="#7B7FE8"/>${[[120, 250, 70], [480, 260, 80], [300, 300, 60]].map(([x, y, r]) => `<circle cx="${x}" cy="${y}" r="${r}" fill="#C8CCE0"/>${Array.from({ length: 6 }, (_, i) => `<path d="M${x - r} ${y - r + i * r / 3}h${2 * r}" stroke="#9A9EB8" stroke-width="2"/>`).join("")}`).join("")}${bottle(250, "#F2643A", "#4F8DF7", "FIZZ")}`),
  "product-1": flat(300, 300, `<rect width="300" height="300" rx="14" fill="#8FD0F0"/><circle cx="230" cy="250" r="30" fill="#BFE6F8"/>${bottle(100, "#D8DCE8", "#E8ECF4", "CLEAN").replace(/#FFFFFF/g, "#7B7FE8")}<circle cx="70" cy="70" r="44" fill="#7B7FE8"/><text x="70" y="66" text-anchor="middle" font-family="Arial Black" font-size="12" fill="#FFFFFF">OUR BIGGEST</text><text x="70" y="82" text-anchor="middle" font-family="Arial Black" font-size="12" fill="#FFFFFF">SALE</text>`),
  "product-2": flat(300, 300, `<rect width="300" height="300" rx="14" fill="#9AD4F2"/>${bottle(100, "#F2E3C8", "#F2643A", "GLOW")}<path d="M150 60c-20-40 40-50 30-10" fill="#9ED37A"/><circle cx="230" cy="240" r="44" fill="#7B7FE8"/><text x="230" y="244" text-anchor="middle" font-family="Arial Black" font-size="12" fill="#FFFFFF">DEALS</text>`),
  "product-3": flat(300, 300, `<rect width="300" height="300" rx="14" fill="#5AB8F0"/><rect x="120" y="60" width="60" height="190" rx="12" fill="#E3EA4A"/><rect x="132" y="250" width="36" height="20" fill="#2B2B2B"/><text x="150" y="160" text-anchor="middle" font-family="Arial Black" font-size="14" fill="#2B2B2B" transform="rotate(-90 150 160)">SPF</text><circle cx="70" cy="70" r="40" fill="#7B7FE8"/><text x="70" y="74" text-anchor="middle" font-family="Arial Black" font-size="12" fill="#FFFFFF">SAVE NOW</text>`),
  "perk-truck": line(80, 60, `<rect x="8" y="16" width="40" height="26"/><path d="M48 24h14l8 10v8H48"/><circle cx="20" cy="46" r="5"/><circle cx="60" cy="46" r="5"/>`, 2, "#1A1A1A"),
  "perk-case": line(80, 60, `<rect x="24" y="16" width="32" height="36" rx="4"/><path d="M34 16v-6h12v6M30 52v4M50 52v4"/>`, 2, "#1A1A1A"),
  "perk-pin": line(80, 60, `<path d="M40 54c-12-16-16-22-16-30a16 16 0 0 1 32 0c0 8-4 14-16 30z"/><circle cx="40" cy="24" r="6"/>`, 2, "#1A1A1A"),
  "snow-strip": flat(600, 90, `${[[40, 40, 28], [150, 50, 22], [310, 40, 0], [420, 44, 30], [540, 40, 18]].map(([x, y, r]) => r ? Array.from({ length: 8 }, (_, i) => `<path d="M${x} ${y}l${(r * Math.cos(i * Math.PI / 4)).toFixed(1)} ${(r * Math.sin(i * Math.PI / 4)).toFixed(1)}" stroke="#9FD8C4" stroke-width="2" stroke-dasharray="2 2"/>`).join("") : "").join("")}`),
  "podcast-cover": svg(600, 300, "#B784FF", "#111111", `
    <path d="M110 230c-30-40-30-100 0-140M80 250c-46-60-46-150 0-210M490 230c30-40 30-100 0-140M520 250c46-60 46-150 0-210" stroke-width="7"/>
    <rect x="220" y="60" width="200" height="240" rx="14" fill="#14B866"/><rect x="200" y="60" width="10" height="240" fill="#F2643A" stroke="none"/>
    <text x="320" y="130" text-anchor="middle" font-family="Impact,Arial Black" font-size="42" fill="#FFFFFF" stroke="none">THE QUIET</text><text x="320" y="190" text-anchor="middle" font-family="Impact,Arial Black" font-size="50" fill="#FFFFFF" stroke="none">DESK</text>
    <path d="M280 250c10-20 70-20 80 0z" fill="#FFFFFF"/><g transform="rotate(-12 200 60)"><rect x="130" y="38" width="150" height="40" rx="20" fill="#111111"/><text x="205" y="64" text-anchor="middle" font-family="Courier New" font-weight="bold" font-size="17" fill="#FFFFFF" stroke="none">NEW EPISODE</text></g>`, 4),
  "quote-banner": svg(560, 260, "#FF8A2A", "#111111", `
    <path d="M30 150h70v50c0 20-70 20-70 0z" fill="#FFFFFF"/><path d="M100 160c20 0 20 30 0 30"/><path d="M10 230l60-20 60 20-60 20z"/>
    <text x="280" y="105" text-anchor="middle" font-family="Arial" font-size="27" fill="#111111" stroke="none">Losing a job says <tspan text-decoration="underline">nothing</tspan></text><text x="280" y="140" text-anchor="middle" font-family="Arial" font-size="27" fill="#111111" stroke="none">about how good you</text><text x="280" y="175" text-anchor="middle" font-family="Arial" font-size="27" fill="#111111" stroke="none">were at it.</text>
    <path d="M470 40c20-20 60-10 70 20M480 230c30-30 70-20 80 10" stroke-width="5"/><rect x="460" y="170" width="70" height="44" rx="6" fill="#FFFFFF" transform="rotate(-20 495 192)"/>`, 3),
  "pony": svg(300, 220, null, "#4A1A7A", `
    <path d="M90 130c10-40 60-50 100-40l30-40 10 10-10 30c20 20 20 60-10 70l-20 40-12-4 10-34h-60l-30 40-12-6 20-40c-20-10-26-20-16-26z" fill="#FFFFFF"/>
    <path d="M220 50l-10-30 16 12M120 120c20 10 40 10 60 0M140 100c10 30 30 40 50 30" /><circle cx="225" cy="70" r="3" fill="#4A1A7A"/><path d="M60 110c-20-10-30 10-40 0 10 20 30 20 40 10z" fill="#FFFFFF"/>`, 4),
  "rope": flat(600, 40, `${Array.from({ length: 40 }, (_, i) => `<path d="M${i * 15} ${20 + 8 * Math.sin(i / 2)}l8 -8" stroke="#FFFFFF" stroke-width="5" stroke-linecap="round"/>`).join("")}`),
  "book-1": flat(600, 340, `<rect width="600" height="340" fill="#0A0A0A"/><rect x="220" y="40" width="160" height="230" fill="#D8D2C4"/>${Array.from({ length: 10 }, (_, i) => `<rect x="${236 + (i % 4) * 34}" y="${70 + Math.floor(i / 4) * 50}" width="22" height="22" fill="#F040A0"/>`).join("")}`),
  "book-2": flat(600, 340, `<rect width="600" height="340" fill="#0A0A0A"/><rect x="200" y="40" width="200" height="260" fill="#2E8A5A"/><path d="M230 140c40-30 100-30 140 0" stroke="#A8E6C8" stroke-width="4" fill="none"/><rect x="230" y="200" width="140" height="10" fill="#A8E6C8"/>`),
  "book-3": flat(600, 340, `<rect width="600" height="340" fill="#0A0A0A"/><rect x="200" y="40" width="200" height="260" fill="#F7E03A"/><text x="300" y="120" text-anchor="middle" font-family="Arial Black" font-size="30" fill="#E8205A">PRINT</text><text x="300" y="160" text-anchor="middle" font-family="Arial Black" font-size="30" fill="#E8205A">FAIR</text><text x="300" y="200" text-anchor="middle" font-family="Arial Black" font-size="30" fill="#E8205A">NOW</text>`),
  "sticker-hi": flat(140, 140, `<circle cx="70" cy="70" r="62" fill="#F7E03A" stroke="#1A1A1A" stroke-width="3"/><circle cx="55" cy="60" r="5" fill="#1A1A1A"/><circle cx="85" cy="60" r="5" fill="#1A1A1A"/><path d="M48 80c10 14 34 14 44 0" stroke="#1A1A1A" stroke-width="5" fill="none" stroke-linecap="round"/>`),
  "ribbon": flat(400, 90, `<g transform="rotate(-10 200 45)"><rect x="20" y="22" width="360" height="46" fill="#E4E2FA"/><text x="200" y="54" text-anchor="middle" font-family="Georgia" font-style="italic" font-size="24" fill="#1A1A1A">Good work deserves a fan</text></g>`),
  "care-person": svg(300, 300, null, "#2A2A3A", `
    <path d="M70 300c0-90 40-130 80-130s80 40 80 130z" fill="#FF7A6A"/><circle cx="150" cy="100" r="40" fill="#8A5A3A"/><path d="M106 100c-10-60 90-70 90-10 10-30-20-60-50-60-40 0-60 30-40 70z" fill="#1A1A1A"/>
    <path d="M120 200c0-30 60-30 60 0s-30 40-30 40-30-10-30-40z" fill="#8A2A4A"/><path d="M110 190l30 20 20-30 30 20" stroke="#8A5A3A" stroke-width="12"/>
    ${[[80, 60], [230, 70], [250, 140], [60, 150]].map(([x, y]) => `<path d="M${x} ${y + 6}c-14-12-4-24 6-16 10-8 20 4 6 16l-6 6z" fill="#FF7A6A" stroke="none"/>`).join("")}`, 3),
  "vases": flat(600, 260, `${[["#9AB8DE", 100], ["#8A2A4A", 300], ["#F2D97A", 500]].map(([c, x]) => `<path d="M${x - 50} 200c0-40 20-50 20-60h60c0 10 20 20 20 60 0 30-20 50-50 50s-50-20-50-50z" fill="${c}"/><path d="M${x - 30} 200c6-8 12 8 18 0s12 8 18 0 12 8 18 0" stroke="#1A1A1A" stroke-width="3" fill="none"/>
    ${Array.from({ length: 7 }, (_, i) => { const a = -Math.PI / 2 + (i - 3) * 0.3; const ex = x + 80 * Math.cos(a), ey = 140 + 110 * Math.sin(a); return `<path d="M${x} 140L${ex.toFixed(1)} ${ey.toFixed(1)}" stroke="#2A2A3A" stroke-width="2"/><path d="M${ex.toFixed(1)} ${(ey + 4).toFixed(1)}c-10-8-2-18 5-12 7-6 15 4 5 12l-5 5z" fill="#FF6A5A"/>`; }).join("")}`).join("")}`),
  "bouquet-hand": svg(160, 180, null, "#2A2A3A", `<path d="M80 170v-90M80 110l-30-50M80 110l30-50M80 120l-10-70" stroke-width="2"/>${[[50, 56], [110, 56], [70, 46], [90, 40], [40, 80], [120, 80]].map(([x, y]) => `<path d="M${x} ${y + 4}c-10-8-2-18 5-12 7-6 15 4 5 12l-5 5z" fill="#FF6A5A" stroke="none"/>`).join("")}<path d="M90 140c20 0 40 10 50 30l-20 10c-10-10-20-16-30-16z" fill="#C88A6A" stroke="none"/>`, 2),
  "line-chart-card": flat(560, 300, `<defs>${grad("lc", ["#F2E6FF", "#DDEFF2"], 1, 1)}</defs><rect width="560" height="300" rx="10" fill="url(#lc)"/><rect x="40" y="30" width="480" height="240" rx="8" fill="#FFFFFF"/>
    <text x="70" y="70" font-family="Arial" font-weight="bold" font-size="17" fill="#1A1A1A">Visibility: <tspan font-weight="normal" fill="#777777">yourbrand.example</tspan></text>
    ${[["#7B4DFF", 70, "62"], ["#4F8DF7", 160, "81"], ["#C8CCE0", 250, "64"], ["#B784FF", 340, "14"], ["#2ED3A0", 430, "66"]].map(([c, x, n]) => `<rect x="${x}" y="98" width="10" height="10" fill="${c}"/><text x="${x + 16}" y="108" font-family="Arial" font-size="11" fill="#555555">site ${x / 90 | 0}</text><text x="${x}" y="134" font-family="Arial" font-size="20" fill="#1A1A1A">${n}</text>`).join("")}
    ${[["#4F8DF7", 160], ["#2ED3A0", 172], ["#7B4DFF", 178], ["#C8CCE0", 182], ["#B784FF", 244]].map(([c, y]) => `<path d="M70 ${y}C200 ${y - 4} 300 ${y + 4} 490 ${y - 2}" stroke="${c}" stroke-width="2" fill="none"/>`).join("")}`),
  "icon-news": line(40, 40, `<rect x="8" y="8" width="24" height="26" rx="2"/><path d="M13 15h14M13 21h14M13 27h8"/>`, 2, "#555555"),
  "icon-person": line(40, 40, `<circle cx="20" cy="20" r="14"/><circle cx="20" cy="16" r="5"/><path d="M11 30c4-6 14-6 18 0"/>`, 2, "#555555"),
  "icon-briefcase": line(40, 40, `<rect x="6" y="14" width="28" height="18" rx="2"/><path d="M15 14v-4h10v4"/>`, 2, "#555555"),
});

for (const [name, markup] of Object.entries(art)) {
  await sharp(Buffer.from(markup), { density: 144 })
    .png({ palette: true, quality: 90, compressionLevel: 9 })
    .toFile(`${out}${name}.png`);
}
console.log(`Drew ${Object.keys(art).length} illustrations.`);
