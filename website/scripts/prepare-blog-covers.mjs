import fs from 'node:fs/promises';
import sharp from 'sharp';
const photos = [
 ['ai-email-marketing','photo-1499750310107-5fef28a66643'],
 ['ai-email-prompts','photo-1455390582262-044cdead277a'],
 ['email-deliverability','photo-1486406146926-c627a92ad1ab'],
 ['email-analytics','photo-1460925895917-afdab827c52f'],
 ['newsletter-growth','photo-1416879595882-3373a0480b5b'],
 ['welcome-email-sequence','photo-1497366754035-f200968a6e72'],
 ['email-personalization','photo-1521737711867-e3b97375f902'],
 ['email-design','photo-1507238691740-187a5b1d37b8'],
 ['email-ab-testing','photo-1434030216411-0b793f4b4173'],
 ['email-automation','photo-1497215728101-856f4ea42174'],
];
const provenance=[];
for (const [name,id] of photos) {
 const source=`https://images.unsplash.com/${id}?auto=format&fit=crop&w=1800&q=85`;
 const response=await fetch(source);if(!response.ok) throw new Error(`${name}: ${response.status}`);
 const image=sharp(Buffer.from(await response.arrayBuffer()));const meta=await image.metadata();if(!meta.width) throw new Error(`Invalid image: ${name}`);
 await image.resize(1600,1067,{fit:'cover'}).webp({quality:84}).toFile(`public/images/blog/${name}.webp`);
 provenance.push({file:`/images/blog/${name}.webp`,source,license:'https://unsplash.com/license',retrieved:'2026-09-10'});
 console.log(name,meta.width,meta.height);
}
await fs.writeFile('design/research/cover-provenance.json',JSON.stringify(provenance,null,2));
const tiles=await Promise.all(photos.map(async([name],i)=>({input:await sharp(`public/images/blog/${name}.webp`).resize(320,213).toBuffer(),left:(i%5)*320,top:Math.floor(i/5)*213})));
await sharp({create:{width:1600,height:426,channels:3,background:'#ffffef'}}).composite(tiles).png().toFile('design/review/blog-cover-contact-sheet.png');
