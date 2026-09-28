import sharp from "sharp";
import { mkdir, copyFile, access } from "node:fs/promises";
const dir = "/var/folders/d4/5hc5gjcd18x7np_cy6yzbwm80000gn/T/";
const output = "public/assets/template-starters/media";
await mkdir(output, { recursive: true });
await mkdir("design-references/email-library/user-screenshots", { recursive: true });
const refs: [string, string, [string, number, number, number, number][]][] = [
 ["webinar-replay", "e408e7db-91ae-46b5-adab-c15831dc5b3c", [["rge-logo",210,22,180,45],["replay-hero",33,84,520,236],["replay-host",174,1092,67,58]]],
 ["figma-show-work", "aa20146d-81e9-4ad9-96c6-8d509882d10f", [["figma-logo",23,14,25,28],["figma-drawing",99,237,248,141],["figma-collage",12,733,426,238],["figma-last-word",12,1474,426,450]]],
 ["meet-the-leaders", "2e556fe1-244c-4d5e-a83d-d55d135fcb8f", [["leaders-banner",14,22,394,260],["leaders-collage",47,587,328,191],["leaders-quote",14,1072,394,226],["leaders-climate",47,1468,69,67],["leaders-community",47,1575,69,69],["leaders-rights",47,1686,69,70]]],
 ["zapconnect", "20b1d411-0b86-4733-9be4-5dba0242fbe3", [["zap-logo",87,41,177,40],["zap-film",87,355,432,249]]],
 ["greenhouse-welcome", "aac95171-a431-46d0-8bda-0d238e3321e4", [["greenhouse-logo",163,36,228,57],["greenhouse-banner",20,125,512,197]]],
 ["most-wanted", "ba9b4b62-490a-48cd-bebe-f357b011dcb1", [["verb-logo",232,33,104,34],["verb-product",83,334,399,186]]],
 ["science-next-gen", "6205e8cc-e6d2-4313-bc53-d6e7ff6f5ede", [["ag1-logo",50,42,91,44],["ag1-cyclist",57,304,445,395]]],
 ["unleash-the-fizz", "260feb42-b2bd-4634-95d9-57ca9bb34dbb", [["olipop-logo",103,0, 74,34],["olipop-fizz",52,174,170,194],["olipop-glass",103,974,68,71]]],
 ["reading-good-for-you", "c8a3d1c6-ed4a-489f-8bf5-7b18049e4078", [["olipop-digest-can",34,213,188,144],["olipop-ice",27,453,205,105],["olipop-pairings",51,977,169,94],["olipop-flavors",27,1242,205,101]]],
 ["softr-product-update", "86005d5e-8f86-4f2e-aaef-80ccd8fe8e96", [["softr-logo",132,20,50,19],["softr-sharing",50,120,214,161],["softr-zoho",50,355,214,161],["softr-integrations",50,591,214,161],["softr-ambassador",50,986,214,112]]],
 ["raise-the-standard", "a08c056a-94a5-4957-8d95-e675ef0dd90b", [["rspca-header",20,10,246,31],["rspca-parliament",20,41,246,165],["rspca-animals",20,355,246,44],["rspca-logo",108,406,74,25]]],
 ["connect-your-tools", "8b4313e9-53d0-4932-a1ce-7aaf27f1715f", [["integrations-hero",29,78,493,250],["integration-html",64,692,53,58],["integration-zip",64,772,53,58],["integration-pdf",64,863,53,54]]],
 ["summer-family-journal", "f1679008-9dff-414f-8b91-618f8f6d15ad", [["summer-family",7,386,264,108],["summer-child",8,881,262,175]]],
 ["summer-family-journal-continuation", "b7311256-47c8-446e-bde7-f3f544ec746d", [["summer-ice",12,281,98,101],["summer-adventure",12,402,98,100],["summer-memories",12,518,98,99],["summer-refresher",12,790,262,242]]],
 ["your-next-teammate", "68c104ea-f84a-4031-8217-59d25ec292cd", [["bot-workspace",35,294,470,266]]],
 ["build-your-first-agent", "6d0869a7-c364-4e73-9763-0faf219b177a", [["workshop-gradient",30,119,652,57]]],
 ["meet-your-specialists", "85031a55-2596-4819-bec4-0c328bafc36a", [["specialist-strategy",45,141,143,102],["specialist-copy",199,141,143,102],["specialist-analytics",45,335,143,104],["specialist-design",199,335,143,104],["specialist-demo",29,611,330,186]]],
 ["digg-tech-digest", "ffb4e7e2-6cde-4e3f-a84f-f887a58016dc", [["digg-mark",159,200,177,94]]],
 ["zapier-monthly", "8796f1bc-660f-48d5-b0cd-c2f847533801", [["zapier-mark",145,113,210,67],["zapier-guided",143,956,908,451],["zapier-mcp",143,1938,908,451],["zapier-tables",143,2963,908,451],["zapier-security",143,4176,435,440],["zapier-skills",614,4176,437,440],["zapier-spotlight",143,5382,908,339]]],
 ["animal-facts-pistol-shrimp-quiz", "research-capture", [["animal-shrimp",95,417,704,469]]],
 ["book-next-international-trip-us", "research-capture", [["golf-landscape",79,214,737,391],["golf-map",79,1216,737,868],["golf-collage",79,2660,737,350]]],
];
for (const [key,id,crops] of refs) {
 const saved = `design-references/email-library/user-screenshots/${key}.png`;
 // Keep regeneration possible after the temporary clipboard files expire.
 const incoming = `${dir}codex-clipboard-${id}.png`;
 const file = await access(incoming).then(() => incoming, () => saved);
 if (file !== saved) await copyFile(file, saved);
 for (const [name,left,top,width,height] of crops) await sharp(file).extract({left,top,width,height}).png().toFile(`${output}/${name}.png`);
}
console.log(`Cropped illustrations from ${refs.length} supplied references.`);
