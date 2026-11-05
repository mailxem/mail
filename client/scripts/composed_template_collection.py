"""Editorial collection: 24 newsletters, each written and laid out by hand.
Every edition has its own sample copy, palette, photos and reading order;
shared helpers only build blocks. Replaced 180 generated subject x recipe
compositions that repeated the same copy and colour bands.
"""
from copy import deepcopy

# Art assets are selected by subject, independently from the layout recipe.
EXTRA_PHOTOS = {
 'books':('photo-1481627834876-b7833e8f5570','Shelves filled with books'),
 'library':('photo-1507842217343-583bb7270b66','A quiet library interior'),
 'art':('photo-1513475382585-d06e58bcb0e0','Colorful artist materials'),
 'city':('photo-1519608487953-e999c86e7455','A city after dark'),
 'vinyl':('photo-1461360370896-922624d12aa1','Music and a listening ritual'),
 'vegetables':('photo-1542838132-92c53300491e','Fresh produce at a market'),
 'salad':('photo-1512621776951-a57141f2eefd','A colorful seasonal meal'),
 'bread':('photo-1509440159596-0249088772ff','Freshly baked bread'),
 'table':('photo-1414235077428-338989a2e8c0','A beautifully prepared dinner'),
 'citrus':('photo-1611080626919-7cf5a9dbab5b','Fresh citrus fruit'),
 'berries':('photo-1488459716781-31db52582fe9','Fresh seasonal ingredients'),
 'coast':('photo-1473116763249-2faaef81ccda','A peaceful coastline'),
 'mist':('photo-1470071459604-3b5ec3a7fe05','Mist over a green landscape'),
 'stars':('photo-1519681393784-d120267933ba','Mountain beneath the night sky'),
 'deer':('photo-1472396961693-142e6e269027','Deer in their natural habitat'),
 'flowers':('photo-1490750967868-88aa4486c946','Fresh flowers in soft light'),
 'lake':('photo-1493246507139-91e8fad9978e','A mountain lake'),
 'meadow':('photo-1500534623283-312aade485b7','Sunlight across an open landscape'),
 'desk':('photo-1497215728101-856f4ea42174','A bright workspace'),
 'building':('photo-1487958449943-2429e8be8625','Modern architectural forms'),
 'chair':('photo-1567538096630-e0c55bd6374c','A considered piece of furniture'),
 'room':('photo-1600210492486-724fe5c67fb0','A warm, thoughtfully arranged room'),
 'kitchen':('photo-1556911220-bff31c812dba','A welcoming kitchen'),
 'plants':('photo-1416879595882-3373a0480b5b','A collection of houseplants'),
 'writing':('photo-1455390582262-044cdead277a','Writing in a notebook'),
 'laptop':('photo-1496181133206-80ce9b88a853','A laptop ready for work'),
 'team':('photo-1522071820081-009f0129c71c','People working together'),
 'meeting':('photo-1517245386807-bb43f82c33c4','A team sharing ideas'),
 'screen':('photo-1551288049-bebda4e38f71','Information displayed on a screen'),
 'train':('photo-1474487548417-781cb71495f3','A train through a landscape'),
 'street':('photo-1519501025264-65ba15a82390','City streets and architecture'),
 'sea':('photo-1507525428034-b723cf961d3e','A beach beside blue water'),
 'travel':('photo-1488646953014-85cb44e25828','Plans for a journey'),
 'clothes':('photo-1445205170230-053b83016050','Everyday clothing and textures'),
 'fabric':('photo-1483985988355-763728e1935b','Clothing and personal style'),
 'shoes':('photo-1543163521-1bf539c55dd2','A considered pair of shoes'),
 'walk':('photo-1551632811-561732d1e306','A walk into the outdoors'),
 'stretch':('photo-1544367567-0f2fcb009e0b','A moment of mindful movement'),
 'journal':('photo-1434030216411-0b793f4b4173','A notebook for learning and reflection'),
 'volunteer':('photo-1559027615-cd4628902d4a','People contributing to their community'),
}
PHOTOS.update(EXTRA_PHOTOS)

def weighted(cols, weights, bg, pad='12px 0px'):
 return {'columns':cols,'cells':weights,'bg':bg,'pad':pad}

def picture(key, shape='wide', pad='0px'):
 b=deepcopy(image(key)); b['values']['containerPadding']=pad
 shapes={'wide':(1200,560),'square':(700,700),'portrait':(600,900),'strip':(1200,280)}
 w,h=shapes[shape]; b['values']['src'].update(width=w,height=h)
 b['values']['src']['url']=b['values']['src']['url'].replace('w=1200&h=660',f'w={w}&h={h}')
 return b


def H(copy, c, size=34, font=SERIF, **kw): return title(copy, size=size, color=c, font=font, **kw)
def T(copy, c, size=16, font=None, align='left', pad='10px 36px'): return text(copy, size, c, font, align, pad)
def L(copy, c, **kw): return label(copy, c, **kw)

def edition(key, name, description, subject, preheader, rows, bg, tags):
 add('edition-'+key, name, 'Newsletters', description, subject, preheader, rows, bg=bg, tags=tags+['editorial'])
 TEMPLATES[-1]['collection'] = 'Editorial collection'

# Each edition is written and laid out by hand: its own copy, palette and
# reading order. Names, places and dates are fictional samples to replace.

# --- Food ---
ink, paper, accent = '#2E2A24', '#F7F2E9', '#A4472B'
edition('corner-table', 'The Corner Table', 'A restaurant newsletter with this week’s menu, a note from the kitchen and a booking button.',
 'This week at the Corner Table', 'Spring lamb is back, and we have two tables left on Friday.', [
 masthead('The Corner Table', 'WEEK 14  ·  WHAT’S COOKING', ink, paper),
 row(picture('table', 'wide'), bg=paper),
 row(H('Spring lamb is back on the menu.', ink, 40), T('From Thursday until the farm runs out, usually about three weeks. Book early if it’s the one you come for.', ink), bg=paper),
 row(L('THIS WEEK’S MENU', accent), T('<strong>Grilled asparagus</strong>, brown butter, hazelnut<br><strong>Slow-roast lamb shoulder</strong>, white beans, salsa verde<br><strong>Whole sea bream</strong> for two, fennel, lemon<br><strong>Rhubarb tart</strong>, crème fraîche', ink), divider('#DCD3C4'),
  T('Prices and allergens are on the full menu.', ink, 14)),
 row(L('FROM THE KITCHEN', accent), T('“We braise the shoulder overnight and finish it on the grill, so it comes out soft inside with a bit of char. Ask for extra salsa verde. We make plenty.” — Mara, head chef', ink, font=SERIF),
  button('Book a table', accent), bg=paper),
], '#E8E0D2', ['food', 'restaurant', 'menu'])

ink, paper, accent = '#3A2E1F', '#FFFDF8', '#5E7A3A'
edition('pantry-notes', 'Pantry Notes', 'A one-recipe newsletter: ingredients beside the method, then a cook’s tip.',
 'A ten-minute bean supper', 'One tin, one pan, and dinner on the table before the rice is done.', [
 row(L('PANTRY NOTES  ·  NO. 31', accent), H('Crispy butter beans with greens and lemon', ink, 36), T('Serves 2  ·  15 minutes  ·  One pan', ink, 14, MONO), bg=paper),
 row(picture('salad', 'strip'), bg=paper),
 weighted([[L('YOU’LL NEED', accent, pad='12px 24px'), T('1 tin butter beans, drained<br>3 tbsp olive oil<br>2 garlic cloves, sliced<br>1 bunch chard or kale<br>1 lemon<br>Chilli flakes, salt', ink, 15, pad='4px 24px')],
           [L('METHOD', accent, pad='12px 24px'), T('1. Pat the beans dry. Fry in the oil over a high heat, without stirring, until the bottoms crisp.<br><br>2. Add the garlic and chilli for a minute, then the greens and a splash of water.<br><br>3. When the greens wilt, take it off the heat. Zest and squeeze in the lemon. Salt to taste.', ink, 15, pad='4px 24px')]], [2, 3], paper),
 row(divider('#E4DDCF'), T('<strong>Tip:</strong> dry beans crisp; wet beans steam. Give them five minutes on a tea towel first.', ink, 15)),
], '#EFE9DC', ['food', 'recipe', 'cooking'])

# --- Culture ---
ink, paper, accent = '#1E2433', '#F4F1EA', '#8C3B2E'
edition('reading-list', 'The Reading List', 'A books newsletter: a short letter, then five recommendations with one line each.',
 'Five books for the long weekend', 'Two novels, a memoir, a short history and one we’re still arguing about.', [
 masthead('The Reading List', 'MONTHLY  ·  APRIL', ink, paper),
 row(T('Hi {{first_name}},<br><br>We read a lot of short books this month, mostly by accident. Here are the five we kept talking about at the shop counter.', ink, 17, SERIF), bg=paper),
 row(L('01', accent), H('The Glass Orchard', ink, 24), T('A family novel set over one hot summer. Quiet, then very much not.', ink, 15), divider('#DDD8CC'),
  L('02', accent), H('Notes from a Night Shift', ink, 24), T('A nurse’s memoir written in short entries. You will read it in a sitting.', ink, 15), divider('#DDD8CC'),
  L('03', accent), H('Salt Roads', ink, 24), T('A short history of how salt moved around the world, and who got rich.', ink, 15), divider('#DDD8CC'),
  L('04', accent), H('The Understudy', ink, 24), T('A comic novel about a theatre that will not close. Funnier than it has any right to be.', ink, 15), divider('#DDD8CC'),
  L('05', accent), H('Field Guide to Leaving', ink, 24), T('Half of us loved it, half didn’t. Come in and pick a side.', ink, 15)),
 row(button('Reserve a copy', ink), T('Reply with the title and we’ll hold it at the till for a week.', ink, 14), bg=paper),
], '#E6E2D8', ['culture', 'books', 'bookshop'])

ink, paper, accent = '#EDE6D6', '#16192A', '#E0A458'
edition('late-show', 'Late Show', 'A dark cinema listings email with dates, times and one featured screening.',
 'This month’s late shows', 'Four films, all after 9pm, and a double bill on the last Saturday.', [
 row(L('LATE SHOW  ·  MAY PROGRAMME', accent), H('Four films after dark.', ink, 46, SANS), bg=paper),
 row(picture('city', 'wide'), bg=paper),
 columns([L('FRI 3', accent, pad='12px 24px'), T('9:15pm', ink, 14, MONO, pad='0px 24px 12px')],
         [H('Harbour Lights', ink, 22, pad='12px 24px 4px'), T('A port town noir, newly restored. 1h 42m.', ink, 14, pad='0px 24px 12px')], bg=paper),
 columns([L('FRI 10', accent, pad='12px 24px'), T('9:30pm', ink, 14, MONO, pad='0px 24px 12px')],
         [H('The Last Ferry', ink, 22, pad='12px 24px 4px'), T('Two strangers miss the boat home. 1h 28m.', ink, 14, pad='0px 24px 12px')], bg=paper),
 columns([L('SAT 25', accent, pad='12px 24px'), T('8:45pm', ink, 14, MONO, pad='0px 24px 12px')],
         [H('Double bill: Static + Undertow', ink, 22, pad='12px 24px 4px'), T('Two short thrillers with an interval for the bar. 3h 05m.', ink, 14, pad='0px 24px 12px')], bg=paper),
 row(button('Get tickets', accent, ink='#16192A'), T('Members get two free tickets a month. Bring a friend.', ink, 14), bg=paper),
], '#0D0F1A', ['culture', 'cinema', 'listings'])

# --- Nature ---
ink, paper, accent = '#24372B', '#EEF1E6', '#B86B2C'
edition('trail-report', 'Trail Report', 'An outdoors newsletter with one route: distance, climb and time at a glance, then the walk itself.',
 'This weekend: the ridge loop', 'Nine kilometres, one steep climb, and the best view in the valley.', [
 masthead('TRAIL REPORT', 'ROUTE 12  ·  THE RIDGE LOOP', ink, paper),
 row(picture('walk', 'wide'), bg=paper),
 weighted([[L('DISTANCE', accent, pad='12px 20px'), H('9 km', ink, 30, SANS, pad='0px 20px 12px')],
           [L('CLIMB', accent, pad='12px 20px'), H('420 m', ink, 30, SANS, pad='0px 20px 12px')],
           [L('TIME', accent, pad='12px 20px'), H('3½ hrs', ink, 30, SANS, pad='0px 20px 12px')]], [1, 1, 1], paper),
 row(H('Start at the church car park.', ink, 26), T('Follow the stream for the first two kilometres, then take the stile on the left. The climb is steady for forty minutes and steep for the last ten. At the top, keep the wall on your right all the way along the ridge.<br><br>Come down through the beech wood. It’s muddy after rain, so leave the trainers at home.', ink)),
 row(L('BEFORE YOU GO', accent), T('No phone signal on the ridge. Download the map, and tell someone your route.', ink, 15), button('Download the map', ink), bg=paper),
], '#DDE3D4', ['nature', 'walking', 'outdoors'])

ink, paper, accent = '#3B3524', '#FBF7EC', '#7A8C3C'
edition('garden-diary', 'The Garden Diary', 'A seasonal gardening email with a month’s jobs as a checklist and a reader’s question answered.',
 'Jobs for the garden in April', 'Sow, pot on, and don’t trust the last frost date.', [
 row(H('The Garden Diary', ink, 30, pad='24px 36px 0px'), T('April, and everything wants doing at once.', ink, 15, SERIF), bg=paper),
 weighted([[picture('plants', 'portrait', '16px 0px 16px 24px')],
           [L('THIS MONTH', accent, pad='16px 24px 4px'), T('☐ Sow tomatoes and chillies indoors<br>☐ Pot on anything with roots out of the bottom<br>☐ Mulch beds while the soil is damp<br>☐ Earth up the first potatoes<br>☐ Keep fleece handy for cold nights', ink, 15, pad='4px 24px')]], [2, 3], paper),
 row(L('YOU ASKED', accent), H('“Why are my seedlings tall and floppy?”', ink, 24), T('They’re reaching for light. Move them to the brightest windowsill you have and turn the tray every day. Next time, sow a week or two later so they aren’t indoors as long.', ink)),
 row(T('Send us your question. We answer one each month.', ink, 14), button('Ask a question', accent), bg=paper),
], '#EDE7D5', ['nature', 'gardening', 'seasonal'])

# --- Design ---
ink, paper, accent = '#1C1C1C', '#FFFFFF', '#D24A1F'
edition('studio-dispatch', 'Studio Dispatch', 'A design studio update: one project, told through a large image, a short write-up and two process shots.',
 'New work: the library signage', 'How we wayfound four floors with two colours and one typeface.', [
 row(L('STUDIO DISPATCH  ·  NEW WORK', accent), H('Signage for a public library that changes every year.', ink, 34, SANS), bg=paper),
 row(picture('library', 'wide')),
 row(T('The brief was short: help people find the toilets, the children’s floor and the quiet room, and don’t make it look like an airport. Shelves move every year, so nothing could be printed on the walls.<br><br>We made a system of hanging panels in two colours and one typeface, which the staff can reprint and rehang themselves.', ink)),
 weighted([[picture('art', 'square', '0px 6px 0px 36px')], [picture('desk', 'square', '0px 36px 0px 6px')]], [1, 1], paper),
 row(T('Left: colour tests under the library’s actual lights. Right: the template the staff use to make new panels.', ink, 13, pad='4px 36px'), button('See the full project', ink, align='left')),
], '#F0F0EE', ['design', 'studio', 'case study'])

ink, paper, accent = '#2B2B2B', '#F3EFE6', '#2B2B2B'
edition('margins', 'Margins', 'A text-only essay newsletter set in monospace, with a single pull quote and nothing else to distract.',
 'On leaving things unfinished', 'A short essay about the drawer of half-done projects.', [
 row(T('MARGINS — ESSAY NO. 9', ink, 12, MONO, pad='28px 36px 4px'), H('On leaving things unfinished', ink, 32, MONO), bg=paper),
 row(T('There is a drawer in my studio for projects I have stopped working on. I used to call it the failure drawer. Now I call it the drawer.<br><br>Some of what is in there will never come out. A few things have come back years later, when I finally knew what they were for. A poster from 2019 became the cover of a book last spring.', ink, 16, MONO), bg=paper),
 row(divider('#BDB6A8'), H('“Unfinished is a state, not a verdict.”', ink, 24, MONO, align='center'), divider('#BDB6A8'), bg=paper),
 row(T('So this month I am not finishing anything. I’m opening the drawer and writing down what each thing was trying to be. If you have a drawer like that, reply and tell me what’s in it.<br><br>— J.', ink, 16, MONO), bg=paper),
], '#E5E0D5', ['design', 'essay', 'text only'])

# --- Work ---
ink, paper, accent = '#14213D', '#FFFFFF', '#2F6FED'
edition('changelog', 'The Changelog', 'A product changelog with a version header and separate lists for added, changed and fixed.',
 'Version 4.2: shared views and faster search', 'Plus eleven fixes, including the one with the date picker.', [
 row(T('THE CHANGELOG', accent, 12, MONO, pad='28px 36px 0px'), H('Version 4.2', ink, 44, SANS), T('Released 14 May · Rolling out to all workspaces this week', ink, 14), bg='#EEF3FE'),
 row(L('ADDED', accent), T('<strong>Shared views.</strong> Save a filtered view and share the link with your team.<br><strong>Keyboard search.</strong> Press / anywhere to search.', ink)),
 row(divider('#E1E6EF'), L('CHANGED', accent), T('Search results now load as you type, instead of after you press enter.<br>Exports include the time zone in every date column.', ink)),
 row(divider('#E1E6EF'), L('FIXED', accent), T('The date picker no longer jumps back a month in some time zones.<br>Ten smaller fixes, all listed on the release page.', ink)),
 row(button('Read the release notes', accent), T('Questions about this release? Reply to this email.', ink, 14)),
], '#E4E9F2', ['work', 'product', 'changelog'])

ink, paper, accent = '#2D2A26', '#FFF8EE', '#C2410C'
edition('monday-memo', 'Monday Memo', 'A short internal team update: three priorities, the week’s dates and one question for everyone.',
 'Monday memo: three things this week', 'Launch prep, the offsite date, and a question for all of you.', [
 masthead('Monday Memo', 'WEEK OF 12 MAY', ink, paper),
 row(H('Three things this week', ink, 30, SANS), T('<strong>1. Launch prep.</strong> Copy freeze is Wednesday at noon. After that, only typo fixes.<br><br><strong>2. Hiring.</strong> Two final interviews on Thursday. If you’re on the panel, read the take-home first.<br><br><strong>3. Support rota.</strong> New rota is in the shared calendar. Swap with each other directly.', ink)),
 columns([L('DATES', accent, pad='12px 24px'), T('Wed 14 · Copy freeze<br>Thu 15 · Interviews<br>Fri 23 · Offsite', ink, 15, pad='4px 24px 12px')],
         [L('ONE QUESTION', accent, pad='12px 24px'), T('What’s one meeting we could cancel without anyone missing it? Reply by Friday.', ink, 15, pad='4px 24px 12px')], bg=paper),
], '#EFE6D8', ['work', 'internal', 'team'])

# --- Places ---
ink, paper, accent = '#1F3A44', '#F5F1E8', '#C26A3D'
edition('postcard', 'Postcard', 'A travel letter written like a postcard, followed by a three-day plan.',
 'A postcard from the coast', 'Three days, one train line, and the best chips we’ve had all year.', [
 row(picture('coast', 'wide'), bg=paper),
 row(L('POSTCARD  ·  FROM THE NORTH COAST', accent), T('Dear {{first_name}},<br><br>We took the slow train up the coast and got off wherever the sea looked good. It rained on the first day and we didn’t mind. Here’s the plan we ended up with, in case you go.<br><br>Wish you were here,<br>Ana &amp; Tom', ink, 17, SERIF), bg=paper),
 weighted([[L('DAY ONE', accent, pad='12px 20px'), T('Arrive at the harbour. Walk the sea wall. Chips from the hut by the lifeboat station.', ink, 14, pad='4px 20px 12px')],
           [L('DAY TWO', accent, pad='12px 20px'), T('Cliff path to the next village, about two hours. Train back if your legs say so.', ink, 14, pad='4px 20px 12px')],
           [L('DAY THREE', accent, pad='12px 20px'), T('Market in the morning, swim if you dare, last train at 5:40.', ink, 14, pad='4px 20px 12px')]], [1, 1, 1], '#FFFFFF'),
 row(button('Get the full guide', ink), bg=paper),
], '#E4DED2', ['places', 'travel', 'itinerary'])

ink, paper, accent = '#24211E', '#FFFFFF', '#9C5B2E'
edition('city-guide', 'Weekend in the City', 'A city guide with four places, each with what it is, where it is and why to go.',
 'Four places for a weekend in town', 'A bakery, a gallery, a park bench and a late bar.', [
 row(H('Weekend in the City', ink, 40, SERIF, align='center', pad='28px 36px 4px'), T('Four places we keep sending friends to.', ink, 15, align='center')),
 row(picture('street', 'strip')),
 columns([L('BREAKFAST', accent, pad='16px 24px 4px'), H('Lark Bakery', ink, 22, pad='0px 24px'), T('12 Mill Street. Get there before ten or the cardamom buns are gone.', ink, 14, pad='6px 24px')],
         [L('MORNING', accent, pad='16px 24px 4px'), H('The Print Room', ink, 22, pad='0px 24px'), T('A small free gallery in an old print works. New show every six weeks.', ink, 14, pad='6px 24px')]),
 columns([L('AFTERNOON', accent, pad='16px 24px 4px'), H('Canal Gardens', ink, 22, pad='0px 24px'), T('Bring lunch. The benches by the lock get the sun until five.', ink, 14, pad='6px 24px')],
         [L('LATE', accent, pad='16px 24px 4px'), H('Number Nine', ink, 22, pad='0px 24px'), T('A bar with twelve seats and good records. Open until one.', ink, 14, pad='6px 24px')]),
 row(button('Open the map', ink, align='center')),
], '#EDEAE4', ['places', 'city guide', 'local'])

# --- Learning ---
ink, paper, accent = '#2A2540', '#F6F4FB', '#6B4FBB'
edition('course-notes', 'Course Notes', 'A lesson in an email course: what this lesson covers, one exercise and the date of the next one.',
 'Lesson 3 of 6: writing a good subject line', 'Today: one exercise, ten minutes, and a before-and-after.', [
 row(T('COURSE NOTES  ·  LESSON 3 OF 6', accent, 12, MONO, pad='28px 36px 4px'), H('Writing a subject line people open', ink, 36, SANS), bg=paper),
 row(T('Hi {{first_name}},<br><br>In lesson two you wrote the email. Today you’ll write the line that decides whether anyone reads it. The short version: say what’s inside, plainly, in under fifty characters.', ink)),
 row(L('BEFORE', accent), T('“Exciting news inside!”', ink, 18, SERIF), L('AFTER', accent), T('“Our shop opens on Saturdays from June”', ink, 18, SERIF), bg=paper),
 row(L('YOUR EXERCISE', accent), T('Take your last three emails. Rewrite each subject line so a stranger could guess what’s inside. Send us your favourite and we’ll reply with notes.', ink), button('Send your answer', accent)),
 row(T('Next lesson: Thursday. Preview text, and why most people leave it blank.', ink, 14, MONO), bg=paper),
], '#E9E5F3', ['learning', 'course', 'lesson'])

ink, paper, accent = '#0F2C3F', '#FFFFFF', '#D97706'
edition('the-explainer', 'The Explainer', 'A one-question explainer: the question as the headline, a plain answer, and further reading.',
 'Why is the sea salty but rivers aren’t?', 'The short answer, the long answer, and three things to read next.', [
 row(L('THE EXPLAINER  ·  THIS WEEK’S QUESTION', accent), H('Why is the sea salty but rivers aren’t?', ink, 38, SANS)),
 row(picture('sea', 'wide')),
 row(L('THE SHORT ANSWER', accent), T('Rivers do carry salt. They just keep moving, so it never builds up. The sea is where it ends up, and when seawater evaporates, the salt stays behind.', ink, 18, SERIF), bg='#FDF6E9'),
 row(L('THE LONGER ANSWER', accent), T('Rain is slightly acidic. As it runs over rock, it dissolves tiny amounts of minerals and carries them downhill. Each river brings only a little, but it has been doing this for a very long time. Water leaves the sea as vapour; the minerals have nowhere to go.', ink)),
 row(divider('#E5E7EB'), L('READ NEXT', accent), T('→ How deep-sea vents add minerals too<br>→ Why the Dead Sea is so much saltier<br>→ Send us a question for next week', ink, 15)),
], '#E8EEF2', ['learning', 'science', 'explainer'])

# --- Community ---
ink, paper, accent = '#3A2A1E', '#FFF6E8', '#C2551F'
edition('neighbourhood-news', 'Neighbourhood News', 'A local community bulletin with news items, a dated events list and a call for volunteers.',
 'Neighbourhood news: the market is moving', 'New market site, a litter pick, and the library’s summer hours.', [
 masthead('Neighbourhood News', 'ISSUE 47  ·  JUNE', ink, paper),
 row(H('The Saturday market is moving to the school yard.', ink, 32, SERIF), T('From 7 June, the market moves from the car park to the primary school yard on Elm Road. More room, more shade, and the toilets are open.', ink), picture('volunteer', 'wide', '12px 36px')),
 row(L('ALSO THIS MONTH', accent), T('• The library is open until 7pm on Thursdays over the summer.<br>• Bin collection moves to Tuesday from the 16th.<br>• The allotment society has two plots free.', ink), bg=paper),
 weighted([[L('DATES', accent, pad='12px 24px'), T('Sat 7 · Market, new site<br>Sun 15 · Litter pick, 10am<br>Fri 27 · Street party meeting', ink, 15, pad='4px 24px 12px')],
           [L('HELP WANTED', accent, pad='12px 24px'), T('We need four people to help set up the market on the first morning. Two hours, coffee provided.', ink, 15, pad='4px 24px'), button('I can help', accent)]], [1, 1], '#FFFFFF'),
], '#F0E4D2', ['community', 'local', 'bulletin'])

ink, paper, accent = '#1E2B24', '#F2F5F0', '#3F7D5A'
edition('member-profile', 'Member Profile', 'A community interview: a photo, a short introduction and four questions with answers.',
 'Meet Sam, who runs code night', 'On teaching beginners, broken Wi-Fi, and why the snacks matter.', [
 weighted([[picture('team', 'portrait', '0px')], [L('MEMBER PROFILE', accent, pad='24px 24px 4px'), H('Sam, code night', ink, 30, pad='4px 24px'), T('Sam started a free Tuesday coding night at the library three years ago. About twenty people come each week, most of them beginners.', ink, 15, pad='8px 24px')]], [2, 3], paper),
 row(H('How did it start?', ink, 20, SANS), T('I taught myself from library books and wanted somewhere to ask questions out loud. So I booked the meeting room and put up a poster.', ink),
  H('What do people work on?', ink, 20, SANS), T('Personal websites, mostly. Then spreadsheets that got out of hand, and one man building a game about his allotment.', ink),
  H('What’s the hardest part?', ink, 20, SANS), T('The library Wi-Fi. We now bring our own router.', ink),
  H('How can people help?', ink, 20, SANS), T('If you can code, come and sit with a beginner for an hour. If you can’t, bring snacks. Both are equally useful.', ink)),
 row(button('Next code night: Tuesday, 6pm', ink), bg=paper),
], '#E1E8DF', ['community', 'interview', 'meetup'])

# --- Style ---
ink, paper, accent = '#2B2420', '#EFE8DF', '#2B2420'
edition('lookbook', 'Lookbook', 'A clothing lookbook with a large lead image, a two-photo grid and short product notes.',
 'The colour edit is in', 'Six pieces in colours we don’t usually make.', [
 row(H('The colour edit', ink, 52, SERIF, align='center', pad='32px 36px 8px'), T('AUTUMN  ·  SIX PIECES', ink, 12, MONO, 'center'), bg=paper),
 row(picture('fashion', 'portrait', '8px 36px'), bg=paper),
 weighted([[picture('clothes', 'square', '0px 6px 0px 36px'), T('<strong>Knits and overshirts</strong><br>Eight colours, in store now', ink, 14, pad='8px 6px 8px 36px')],
           [picture('fabric', 'square', '0px 36px 0px 6px'), T('<strong>Wool coat</strong><br>Burgundy, two lengths', ink, 14, pad='8px 36px 8px 6px')]], [1, 1], paper),
 row(T('In store and online from Friday. Free returns for 30 days.', ink, 16, SERIF, 'center'), button('Shop the edit', ink, align='center'), bg=paper),
], '#E2D9CD', ['style', 'fashion', 'lookbook'])

ink, paper, accent = '#3B2F28', '#FFFFFF', '#8A5A3B'
edition('care-guide', 'The Care Guide', 'An after-purchase care email with numbered steps and a repair offer.',
 'How to look after your wool coat', 'Four habits that keep it going for years.', [
 row(picture('fabric', 'wide')),
 row(L('THE CARE GUIDE', accent), H('Look after it and it will see you through a lot of winters.', ink, 32, SERIF)),
 row(T('<strong>1. Brush it.</strong> A soft clothes brush, top to bottom, after a few wears. It lifts dust before it settles in.', ink), divider('#EEE6DE'),
  T('<strong>2. Air it.</strong> Hang it by a window overnight instead of washing it. Wool sheds most smells on its own.', ink), divider('#EEE6DE'),
  T('<strong>3. Spot clean.</strong> Blot spills with cold water. Don’t rub, and never put it in the dryer.', ink), divider('#EEE6DE'),
  T('<strong>4. Hang it properly.</strong> A wide wooden hanger keeps the shoulders in shape. Store it in a cotton bag over summer.', ink)),
 row(H('Torn lining or a lost button?', ink, 22, SANS), T('We repair our coats free for the first two years. Send it back and we’ll return it within a fortnight.', ink, 15), button('Start a repair', accent), bg='#F6F0EA'),
], '#ECE5DD', ['style', 'aftercare', 'repair'])

# --- Wellbeing ---
ink, paper, accent = '#3C3A33', '#F9F6EF', '#8C7A5B'
edition('morning-pages', 'Morning Pages', 'A short, calm weekly letter with one writing prompt and a small image.',
 'This week’s prompt', 'One question to write about before you check your phone.', [
 row(T('Morning Pages', ink, 15, SERIF, 'center', '32px 36px 0px'), T('NO. 52', accent, 11, MONO, 'center', '4px 36px 12px'), bg=paper),
 row(picture('journal', 'square', '8px 120px'), bg=paper),
 row(L('THIS WEEK’S PROMPT', accent, align='center'), H('What did you used to do for fun that you’ve stopped doing?', ink, 28, SERIF, align='center'), bg=paper),
 row(T('Write for ten minutes without stopping. Don’t read it back until the weekend.<br><br>A year of these letters ends next week. Thank you for reading along.', ink, 16, SERIF, 'center'), bg=paper),
], '#EFEADF', ['wellbeing', 'writing', 'letter'])

ink, paper, accent = '#203A33', '#EAF3EF', '#E07A5F'
edition('class-schedule', 'Movement Club', 'A fitness or yoga studio schedule laid out as a weekly timetable, with a new-class note.',
 'This week’s classes', 'A new Sunday stretch class, and the Wednesday 7am is back.', [
 row(L('MOVEMENT CLUB', accent), H('This week’s classes', ink, 40, SANS), T('All classes are 50 minutes. Mats provided.', ink, 15), bg=paper),
 row(picture('stretch', 'strip')),
 weighted([[T('<strong>MON</strong><br>7:00 Flow<br>18:30 Strength', ink, 14, pad='12px 12px 12px 36px')], [T('<strong>WED</strong><br>7:00 Flow<br>19:00 Pilates', ink, 14, pad='12px')],
           [T('<strong>FRI</strong><br>12:15 Lunch stretch<br>18:00 Strength', ink, 14, pad='12px')], [T('<strong>SUN</strong><br>10:00 Slow stretch<br><em>new</em>', ink, 14, pad='12px 36px 12px 12px')]], [1, 1, 1, 1], '#FFFFFF'),
 row(H('New: Sunday slow stretch', ink, 24, SANS), T('A gentle class for stiff backs and people who sit down all week. No experience needed.', ink), button('Book a class', ink), bg=paper),
], '#DCE9E3', ['wellbeing', 'fitness', 'schedule'])

# --- Science ---
ink, paper, accent = '#132A35', '#F0F5F6', '#C97C1E'
edition('field-findings', 'Field Findings', 'A research update for supporters: what the team found, a photo from the field, and a quote.',
 'What we found on the moor this spring', 'Fewer curlews, more otters, and one surprise in the camera traps.', [
 row(L('FIELD FINDINGS  ·  SPRING SURVEY', accent), H('What we found on the moor this spring', ink, 36, SERIF), bg=paper),
 row(picture('deer', 'wide')),
 row(L('THE MAIN RESULTS', accent), T('<strong>Curlews:</strong> 14 nesting pairs, down from 19 last year.<br><strong>Otters:</strong> signs at six of eight river sites, up from four.<br><strong>The surprise:</strong> a pine marten on camera three, the first record here in decades.', ink)),
 row(T('“The curlew numbers worry us. Next year we’ll start fencing nests earlier, before the foxes find them.”', ink, 20, SERIF), T('— Dr Lena Hart, survey lead', ink, 14, MONO), bg=paper),
 row(T('Survey numbers in this sample are placeholders. Replace them with your own results.', ink, 12), button('Read the full report', ink)),
], '#DEE8EA', ['science', 'research', 'conservation'])

ink, paper, accent = '#E6E9F2', '#0E1424', '#F2C14E'
edition('night-sky', 'The Night Sky', 'A dark astronomy newsletter with the month’s highlights as a dated list.',
 'What to look for in August', 'A meteor shower, a close moon and planet, and when to stay up.', [
 row(picture('stars', 'wide'), bg=paper),
 row(L('THE NIGHT SKY  ·  AUGUST', accent), H('Stay up late on the 12th.', ink, 40, SERIF), T('The meteor shower peaks after midnight, and this year the moon sets early, so the sky will be properly dark.', ink), bg=paper),
 columns([L('AUG 12–13', accent, pad='12px 24px'), T('Meteor shower peak. Face north-east, give your eyes twenty minutes.', ink, 15, pad='4px 24px 12px')],
         [L('AUG 20', accent, pad='12px 24px'), T('The moon passes close to Saturn. Look south-east after 10pm.', ink, 15, pad='4px 24px 12px')], bg=paper),
 row(divider('#2A3350'), T('Check the dates and times for your location before sending; these are samples.', ink, 12), bg=paper),
], '#070A14', ['science', 'astronomy', 'monthly'])

# --- Home ---
ink, paper, accent = '#3A3029', '#F8F1E8', '#A0522D'
edition('one-room', 'One Room', 'A home and interiors email about one room: what changed, a materials list and the cost.',
 'One room: the living room', 'Paint, lamps and a moved sofa, all in one weekend.', [
 row(L('ONE ROOM  ·  THE LIVING ROOM', accent), H('A dim living room, fixed in a weekend.', ink, 38, SERIF), bg=paper),
 row(picture('room', 'wide'), bg=paper),
 row(T('The living room faces north and always felt gloomy by four o’clock. We painted the walls a warm white, moved the sofa off the window wall, and added two lamps at eye level instead of relying on the ceiling light.', ink)),
 weighted([[L('WHAT WE USED', accent, pad='12px 24px'), T('3 tins of paint<br>2 floor lamps<br>1 rug, second-hand<br>6 picture frames', ink, 15, pad='4px 24px 12px')],
           [L('WHAT IT TOOK', accent, pad='12px 24px'), T('Two days, one friend with a van, and a lot of masking tape.', ink, 15, pad='4px 24px 12px')]], [1, 1], paper),
 row(button('See the before and after', accent)),
], '#ECE3D6', ['home', 'interiors', 'renovation'])

ink, paper, accent = '#2E3440', '#FFFFFF', '#4C7A5A'
edition('fix-it-list', 'The Fix-It List', 'A seasonal home maintenance checklist grouped by indoors and outdoors.',
 'Your autumn fix-it list', 'Eight small jobs to do before the first cold snap.', [
 row(H('The Fix-It List', ink, 40, SANS, pad='28px 36px 4px'), T('AUTUMN  ·  BEFORE THE FIRST FROST', accent, 12, MONO, pad='0px 36px 16px'), bg='#EEF3EF'),
 columns([L('OUTSIDE', accent, pad='16px 24px 4px'), T('☐ Clear the gutters<br>☐ Drain the outside tap<br>☐ Check the roof for loose tiles<br>☐ Oil the garden tools', ink, 15, pad='4px 24px 12px')],
         [L('INSIDE', accent, pad='16px 24px 4px'), T('☐ Bleed the radiators<br>☐ Test the smoke alarms<br>☐ Seal draughty windows<br>☐ Book the boiler service', ink, 15, pad='4px 24px 12px')]),
 row(divider('#E5E9E6'), H('Not a DIY person?', ink, 22, SANS), T('Reply with the jobs you’d like done and we’ll send a quote.', ink, 15), button('Get a quote', accent)),
], '#E7ECE8', ['home', 'maintenance', 'checklist'])
