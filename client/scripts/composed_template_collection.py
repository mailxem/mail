"""Studio collection: lifecycle emails rebuilt from well-made real-world layouts.
Each one follows the structure of a reference email (section order, spacing,
colour blocking, type scale) but uses a fictional brand, our own copy and
original illustrations from scripts/template_illustrations.mjs.
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


# Unlayer's built-in font list, beyond the generator's Arial/Georgia/Courier.
BLACK = {'label': 'Arial Black', 'value': 'arial black,avant garde,arial'}
TREB = {'label': 'Trebuchet MS', 'value': 'trebuchet ms,geneva'}
HELV = {'label': 'Helvetica', 'value': 'helvetica,sans-serif'}

def H(copy, c, size=34, font=SERIF, align='center', pad='10px 40px'):
 return title(copy, size=size, color=c, font=font, align=align, pad=pad)
def T(copy, c, size=16, font=None, align='center', pad='8px 48px'):
 return text(copy, size, c, font, align, pad)
def ill(name, alt, width='100%', pad='0px', align='center'):
 return {'type': 'image', 'values': {'src': {'url': f'/assets/template-starters/illustrations/{name}.png', 'width': 1200, 'height': 600, 'autoWidth': False, 'maxWidth': width},
  'altText': alt, 'textAlign': align, 'containerPadding': pad, 'action': {'name': 'web', 'values': {'href': '', 'target': '_blank'}}}}
def pill(copy, bg, ink='#FFFFFF', align='center', outline=None, font=SANS, pad='16px 40px 28px', size='16px', inner='14px 32px'):
 b = button(copy, bg, align=align, ink=ink)
 b['values'].update(borderRadius='999px', containerPadding=pad, fontFamily=font, fontSize=size, padding=inner)
 if outline: b['values']['border'] = {'borderTopWidth': '2px', 'borderTopStyle': 'solid', 'borderTopColor': outline, 'borderBottomWidth': '2px', 'borderBottomStyle': 'solid', 'borderBottomColor': outline, 'borderLeftWidth': '2px', 'borderLeftStyle': 'solid', 'borderLeftColor': outline, 'borderRightWidth': '2px', 'borderRightStyle': 'solid', 'borderRightColor': outline}
 return b
def foot(bg, c, *lines):
 return row(*[T(l, c, 12, pad='4px 48px') for l in lines], T('<a href="https://xem.email" style="color:inherit">Built with Xem</a>', c, 12, pad='12px 48px 32px'), bg=bg, pad='24px 0px 0px')

def studio(key, name, category, description, subject, preheader, rows, footer, bg, tags, marketing=True):
 add('studio-'+key, name, category, description, subject, preheader, rows, bg=bg, marketing=marketing, tags=tags)
 TEMPLATES[-1]['rows'][-1] = footer  # each design carries its own footer colours
 TEMPLATES[-1]['collection'] = 'Studio collection'

ADDRESS = 'Your company · 12 Example Street, Your City'
WHY = 'You’re receiving this because you signed up at our website. <a href="https://example.com" style="color:inherit">Unsubscribe</a>'

# Brands, people, dates and offers below are fictional samples to replace.

# --- Welcome ---
g, sage, cream = '#2F5D46', '#D9E1C6', '#FAF6EC'
studio('while-you-wait', 'While You Wait', 'Welcome', 'A thank-you for an enquiry or lead form, with an illustrated header and one article to read in the meantime.',
 'Thanks, we’ll be in touch soon', 'Someone from our team will reply within a working day.', [
 weighted([[H('Fernhouse', g, 32, align='left', pad='28px 12px 16px 40px')], [pill('VISIT FERNHOUSE', g, sage, align='right', pad='30px 40px 16px 12px', size='11px', inner='8px 16px')]], [3, 2], sage, '0px'),
 row(divider(g), ill('gift-stack', 'A stack of wrapped parcels with a small bird sitting on top', pad='8px 0px 0px'), bg=sage, pad='0px'),
 row(H('Thank you!', '#1E1E1E', 44, align='left', pad='40px 40px 8px'), T('We’ve got your details, and someone from our team will be in touch within a working day to talk through your gifting plans.', '#1E1E1E', 18, SERIF, 'left', '8px 40px 24px'), divider('#1E1E1E'),
  H('While you wait…', '#1E1E1E', 38, align='left', pad='24px 40px 4px'), T('Here’s what we’ve been thinking about lately.', '#1E1E1E', 15, SERIF, 'left', '4px 40px 16px'), bg=cream),
 weighted([[ill('tote-circle', 'A green tote bag with a parcel balanced on top', '220px', '20px 8px 20px 40px')],
           [H('How we choose our gifts', '#1E1E1E', 26, align='left', pad='36px 40px 8px 12px'), T('The four things we look for before anything goes in a box.', '#1E1E1E', 14, SERIF, 'left', '4px 40px 8px 12px'), pill('READ MORE', '#1E1E1E', align='left', pad='8px 40px 24px 12px', size='12px', inner='10px 22px')]], [1, 1], '#EDE1C9', '0px'),
], foot(cream, '#6B6B60', ADDRESS, WHY), '#E9E4F5', ['lead', 'thank you', 'illustrated'])

ink, mist = '#1F5673', '#F2F4F3'
studio('hi-there', 'Hi, There', 'Welcome', 'A short, friendly welcome with a mascot drawing, one button, and a fun fact that gives the brand a voice.',
 'Hi there, glad you stopped by', 'A quick hello, and a tortoise fact to go with it.', [
 row(H('TORTOISE TEA CO.', ink, 22, BLACK, pad='20px 40px'), bg='#FFFFFF', pad='0px'),
 row(ill('tortoise', 'A line drawing of a tortoise, the tea company’s mascot', '420px', '48px 40px 8px'), H('Hi, there.', ink, 40, HELV, pad='24px 40px 8px'),
  T('Thanks for stopping by. We blend loose-leaf tea in small batches and post it the same week, so it’s still fresh when it reaches you.', '#2A2A2A', 16, HELV), pill('EXPLORE OUR TEA', ink, inner='14px 30px', pad='20px 40px 48px'), bg=mist, pad='0px'),
 row(T('<strong>TORTOISE FACT #31</strong>', ink, 13, HELV, pad='36px 48px 8px'), T('Some tortoises live for more than 150 years. Our oldest blend is only twelve.', '#2A2A2A', 15, HELV, pad='0px 96px 36px'), bg='#FFFFFF', pad='0px'),
], foot(mist, '#6B6B6B', 'hello@example.com', ADDRESS, '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), mist, ['welcome', 'mascot', 'brand voice'])

lime, teal = '#E3F45A', '#18B3C8'
studio('almost-there', 'You’re Almost There', 'Welcome', 'A bright sign-up nudge with a confirm button, three quick-reply choices, and a grid of what members get.',
 'One more step to your member discounts', 'Confirm your student status and you’re in.', [
 row(H('CAMPUS CARD', '#111111', 26, BLACK, pad='22px 40px'), bg='#FFFFFF', pad='0px'),
 row(ill('shapes-strip', 'A row of small triangles and diamonds', pad='12px 40px 0px'), H('You’re almost there!', '#111111', 34, TREB, pad='16px 40px 4px'),
  T('Confirm your student status today and start using your discounts on books, travel, food and more.', '#111111', 15, TREB), pill('Confirm now', teal, '#111111', outline='#111111', font=TREB, inner='10px 28px', pad='16px 40px 40px'), bg=lime, pad='0px'),
 row(H('Ready to start?', '#111111', 24, TREB, pad='32px 40px 8px'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('choice-yes', 'A phone with a tick on the screen', '90px', '8px 0px'), T('Yes<br><strong><a href="https://example.com" style="color:#111111;text-decoration:none">Confirm now ›</a></strong>', '#111111', 14, TREB, pad='4px 8px 24px')],
           [ill('choice-maybe', 'A green square with a question mark', '90px', '8px 0px'), T('Maybe<br><strong><a href="https://example.com" style="color:#111111;text-decoration:none">Learn more ›</a></strong>', '#111111', 14, TREB, pad='4px 8px 24px')],
           [ill('choice-no', 'A yellow circle with a cross', '90px', '8px 0px'), T('Not now<br><strong><a href="https://example.com" style="color:#111111;text-decoration:none">Tell us why ›</a></strong>', '#111111', 14, TREB, pad='4px 8px 24px')]], [1, 1, 1], '#FFFFFF', '0px'),
 row(H('What you get', '#111111', 24, TREB, pad='32px 40px 4px'), T('Discounts all year round, in six categories.', '#111111', 14, TREB, pad='0px 48px 12px'), bg=teal, pad='0px'),
 weighted([[T('<strong>Books</strong>', '#111111', 15, TREB, pad='22px 8px 22px 32px')], [T('<strong>Travel</strong>', '#111111', 15, TREB, pad='22px 8px')], [T('<strong>Food</strong>', '#111111', 15, TREB, pad='22px 32px 22px 8px')]], [1, 1, 1], '#FFFFFF', '0px'),
 weighted([[T('<strong>Tech</strong>', '#111111', 15, TREB, pad='22px 8px 22px 32px')], [T('<strong>Fitness</strong>', '#111111', 15, TREB, pad='22px 8px')], [T('<strong>Fashion</strong>', '#111111', 15, TREB, pad='22px 32px 22px 8px')]], [1, 1, 1], teal, '0px'),
], foot('#FFFFFF', '#555555', ADDRESS, WHY), '#F1F1F1', ['verification', 'student', 'choices'])

# --- Transactional ---
orange, dark = '#FF7A1A', '#3A3A3A'
studio('verify-sleepy', 'Verify Your Email', 'Transactional', 'A warm, illustrated email-verification message with one clear button and a help line.',
 'Confirm your email address', 'One click and you’re ready for your first night.', [
 row(H('<span style="color:#FF7A1A">●</span> driftwood', dark, 30, HELV, pad='40px 40px 16px'), ill('sleepy-sun', 'A sleepy sun with its eyes closed, setting behind a hill', '360px', '8px 40px'),
  H('Confirm your<br>email address', dark, 40, HELV, pad='24px 40px 8px'), T('You’re nearly set up. Press the button below to confirm your email address. The link works for the next 24 hours.', dark, 16, HELV, pad='4px 64px'),
  pill('Confirm my email', orange, pad='24px 40px 40px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#6B6B6B', 'Questions? Read our <a href="https://example.com" style="color:#FF7A1A">help pages</a> or reply to this email.', 'Driftwood · 12 Example Street, Your City'), '#F7F3EF', ['verification', 'account', 'illustrated'], marketing=False)

studio('cancellation', 'Cancellation Confirmed', 'Transactional', 'A plain, kind cancellation confirmation: what happens now, when access ends, and how to come back.',
 'Your subscription is cancelled', 'You can keep listening until the end of your billing period.', [
 row(H('<span style="color:#FF7A1A">●</span> driftwood', dark, 26, HELV, pad='40px 40px 8px'), H('Your cancellation<br>is confirmed', dark, 36, HELV, pad='16px 40px 16px'),
  T('Hi {{first_name}},<br><br>We’ve cancelled your Driftwood subscription. You still have full access to every sleep story until <strong>27 October</strong>.<br><br>After that, the app will only play the free stories. If you change your mind, you can resubscribe in two taps.<br><br>If you asked for a refund, you may get a second cancellation email once it has gone through.<br><br>– The Driftwood team', dark, 16, HELV, pad='8px 64px 40px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#6B6B6B', 'Questions? Read our <a href="https://example.com" style="color:#FF7A1A">help pages</a> or reply to this email.', 'Driftwood · 12 Example Street, Your City'), '#F3F3F7', ['cancellation', 'subscription', 'account'], marketing=False)

paper, black = '#EEEEE9', '#2B2B2B'
studio('limit-reached', 'Limit Reached', 'Transactional', 'A usage-limit alert with a simple illustration, what has stopped, and the one setting that fixes it.',
 'You’ve used 100% of your monthly limit', 'New events are paused until your limit resets or you raise it.', [
 row(H('graphwell', black, 24, BLACK, pad='20px 40px'), bg='#FFFFFF', pad='0px'),
 row(ill('meter-full', 'A browser window with a progress bar filled to the end', '380px', '24px 40px 8px'), bg=paper, pad='0px'),
 row(H('You’ve hit 100% of your<br>monthly limit', black, 30, HELV, pad='28px 40px 8px'), T('You set a limit of 1,000,000 events a month for the <strong>Storefront</strong> project, and you reached it today.', black, 16, HELV, pad='8px 56px'),
  T('<strong>We’ve paused new events for this project until your limit resets on the 1st, or until you raise it.</strong> Nothing already collected is lost.', black, 16, HELV, pad='8px 56px'),
  pill('Change your limit', '#F2A33A', black, outline=black, pad='20px 40px 40px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#777777', 'Graphwell · 12 Example Street, Your City', 'Sent because you’re an owner of this project.'), paper, ['usage', 'billing', 'alert'], marketing=False)

# --- Retention ---
forest, mint, oat = '#21382F', '#D4E9E2', '#F1EFEA'
studio('challenge-progress', 'Halfway There', 'Retention', 'A loyalty challenge update with a scattered illustration, progress so far, an outlined button and an icon menu.',
 'You’re halfway through the challenge', '40 bonus stars collected. 35 to go.', [
 row(T('ORDER AHEAD IN <a href="https://example.com" style="color:#21382F">THE APP</a>', forest, 13, HELV, pad='16px 40px'), bg=mint, pad='0px'),
 row(H('COPPER CUP', '#FFFFFF', 30, BLACK, pad='36px 40px 2px'), T('R E W A R D S', '#FFFFFF', 14, HELV, pad='0px 40px 32px'), bg=forest, pad='0px'),
 row(ill('breakfast-scatter', 'A scatter of drinks, pastries, toast and stars', '100%', '32px 24px 0px'), H('Nice going', forest, 40, SERIF, pad='16px 40px 8px'),
  T('You’re halfway through the challenge. So far you’ve collected <strong>40 bonus stars</strong>. Keep going to get all 75.<br><br>Track your stars in the app. Bonus stars land in your account within three days of the challenge ending.', forest, 17, HELV, pad='8px 64px'),
  pill('Order ahead', oat, forest, outline=forest, pad='24px 40px 48px', size='20px', inner='14px 40px'), bg=oat, pad='0px'),
 weighted([[ill('icon-phone', 'App', '80px', '20px 0px 0px'), T('APP', forest, 13, HELV, pad='0px 4px 20px')], [ill('icon-cup', 'Order', '80px', '20px 0px 0px'), T('ORDER', forest, 13, HELV, pad='0px 4px 20px')],
           [ill('icon-star', 'Offers', '80px', '20px 0px 0px'), T('OFFERS', forest, 13, HELV, pad='0px 4px 20px')], [ill('icon-gift', 'Rewards', '80px', '20px 0px 0px'), T('REWARDS', forest, 13, HELV, pad='0px 4px 20px')]], [1, 1, 1, 1], mint, '0px'),
], foot('#FFFFFF', '#6B6B6B', 'Challenge runs 11–17 October at participating shops. Activate the offer in the app before your first order.', ADDRESS, WHY), '#FFFFFF', ['loyalty', 'progress', 'rewards'])

blue = '#4F8DF7'
studio('hundred', 'One Hundred', 'Retention', 'A milestone celebration for a customer or partner, with a big number illustration and a single button.',
 'You just made your 100th delivery', 'Thank you for a hundred trips across town.', [
 row(H('Pedalpost', '#111111', 26, HELV, align='left', pad='22px 40px'), bg='#FFFFFF', pad='0px'),
 row(H('A hundred deliveries.<br>Thank you.', '#111111', 40, HELV, align='left', pad='40px 40px 12px'),
  T('Tien, you’ve just finished your 100th delivery with us. That’s a hundred parcels across town, in every kind of weather.<br><br>Thank you for every one of them.<br><br>Ride on.', '#111111', 17, HELV, 'left', '8px 40px'),
  button('Keep going →', '#111111', align='left'), ill('milestone-100', 'A huge white number 100 with a bicycle and parcels around it', pad='16px 0px 0px'), bg=blue, pad='0px'),
], row(T('<a href="https://example.com" style="color:#FFFFFF">Help</a> · <a href="https://example.com" style="color:#FFFFFF">Privacy</a> · <a href="https://example.com" style="color:#FFFFFF">Terms</a>', '#FFFFFF', 13, HELV, 'left', '32px 40px 8px'), T('Pedalpost · 12 Example Street, Your City', '#BBBBBB', 12, HELV, 'left', '4px 40px'), T('<a href="https://xem.email" style="color:#BBBBBB">Built with Xem</a>', '#BBBBBB', 12, HELV, 'left', '8px 40px 32px'), bg='#000000', pad='0px'),
 '#FFFFFF', ['milestone', 'celebration', 'partners'])

navy, violet, ivory = '#0E3E45', '#4B3BFF', '#FFFEF2'
studio('travel-survey', 'Tell Us Your Plans', 'Retention', 'A member survey request with an illustration, why the answers matter, a prize line and one button.',
 'Where are you going next year?', 'Five minutes, a few questions, and a chance to win $200.', [
 row(H('farefinder', violet, 30, TREB, pad='40px 40px 16px'), H('Tell us about your<br><span style="color:#4B3BFF">travel plans for next year</span>', navy, 26, HELV, pad='8px 40px'),
  ill('calendar-ribbon', 'Two calendar pages with dates crossed off, joined by a ribbon, and a pen', pad='24px 24px 8px'),
  T('Will you travel more or less than this year? Where are you hoping to go? How do you choose when to book?<br><br>Your answers help us decide which routes to watch, and they go into our yearly travel report.', navy, 16, HELV, pad='8px 56px'),
  T('<strong>Finish the survey and you’ll be entered into a draw for $200.</strong>', navy, 16, HELV, pad='12px 56px'), pill('Take the survey', violet, pad='20px 40px 40px'), divider(navy), bg=ivory, pad='0px'),
], foot(ivory, navy, 'Farefinder · 12 Example Street, Your City', 'Draw closes 30 November. Full rules on our website.', '<a href="https://example.com" style="color:inherit">Email preferences</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), ivory, ['survey', 'feedback', 'prize'])

olive, chartreuse = '#2F3A12', '#D4E157'
studio('five-stars', 'What Do You Think?', 'Retention', 'A playful review request with big serif type, a highlighted question and hand-drawn stars.',
 'Would you give us five stars?', 'If you’ve got a minute, we’d love a review.', [
 weighted([[ill('doodle-burst', 'A small hand-drawn sparkle', '60px', '24px 0px 0px 24px', 'left')], [H('GOOD CRUMB', olive, 26, BLACK, pad='32px 8px 0px')], [ill('doodle-burst', 'A small hand-drawn sparkle', '50px', '30px 24px 0px 0px', 'right')]], [1, 3, 1], chartreuse, '0px'),
 row(H('Five stars,<br>or fewer?', olive, 54, SERIF, pad='24px 40px 8px'), H('<span style="background:#F5EDDE;padding:2px 14px"><em>Be honest. We can take it.</em></span>', olive, 30, SERIF, pad='4px 40px 20px'),
  ill('five-stars', 'Five hand-drawn stars', '420px', '8px 40px 16px'),
  T('We’d really like to hear what you thought of your last order. If you’ve got a minute, leave us a review :-)<br><br>We’ll read every single one.', olive, 15, MONO, pad='12px 72px'),
  pill('Leave a review', '#E7F06B', olive, outline=olive, font=MONO, pad='20px 40px 8px', inner='14px 64px'), ill('loaf', 'A loaf of bread', '150px', '8px 40px 24px', 'right'), bg=chartreuse, pad='0px'),
], foot('#F5EDDE', olive, 'Good Crumb Bakery · 12 Example Street, Your City', WHY), chartreuse, ['review', 'feedback', 'playful'])

# --- Product ---
studio('feet-up', 'Feet Up', 'Product', 'A feature-education email in a light, joking voice, with a big illustration and inline links to each option.',
 'A good reason to take a break', 'Set up an experiment, then let the results come to you.', [
 weighted([[H('graphwell', black, 24, BLACK, align='left', pad='24px 12px 20px 40px')], [T('<a href="https://example.com" style="color:#E8742C;font-weight:bold">Visit Graphwell</a>', black, 15, HELV, 'right', '28px 40px 20px 12px')]], [1, 1], '#E3E3DE', '0px'),
 row(ill('desk-feet-up', 'A tidy desk with an open laptop, a mug, stacked books and a bin of paper balls', pad='8px 0px'), H('A good <s>excuse</s> reason to take a break', '#4A4A4A', 32, HELV, pad='16px 32px 8px'),
  T('There are two good things about running experiments in Graphwell.<br><br>The first is that there are plenty to choose from. There are <a href="https://example.com" style="color:#E8742C">A/B tests</a>, of course, but also <a href="https://example.com" style="color:#E8742C">multivariate tests</a>, <a href="https://example.com" style="color:#E8742C">holdout groups</a> and <a href="https://example.com" style="color:#E8742C">redirect tests</a>.<br><br>The <em>other</em> good thing is that once you’ve set one up, you can get on with something else while the numbers come in. We call that delegating.', '#4A4A4A', 17, HELV, 'left', '8px 32px'),
  pill('Start an experiment', '#111111', pad='20px 40px 32px', size='18px'), divider('#BDBDB6'), bg=paper, pad='0px'),
], foot(paper, '#4A4A4A', 'We keep our emails few and far between, but you can still <a href="https://example.com" style="color:inherit">unsubscribe</a>.', 'Graphwell · 12 Example Street, Your City'), paper, ['feature', 'education', 'product'])

ink2, warm = '#4A4458', '#FAF4EB'
studio('first-lesson', 'Lesson One', 'Product', 'An onboarding lesson with a video still, then two illustrated questions and answers on a warm panel.',
 'Lesson one: learning to slow down', 'Five minutes, a comfortable chair, and no mountaintop required.', [
 row(T('<span style="color:#E8742C">●</span> QUIET HOUR', '#E8742C', 13, HELV, pad='36px 40px 8px'), H('Learning to slow down', ink2, 38, HELV, pad='8px 40px'),
  T('Meditation isn’t about becoming a different person. It’s practice at noticing what’s going on, and getting a bit of distance from it. This two-minute video explains how.', ink2, 15, HELV, pad='8px 64px 16px'),
  ill('lesson-video', 'A video still: a play button between a cactus and a small round character on a hill', '440px', '8px 40px 32px'), bg='#FFFFFF', pad='0px'),
 row(H('When should I practise?', ink2, 26, HELV, pad='40px 40px 8px'), ill('spot-sunrise', 'A sun rising over the horizon', '170px', '8px 40px'),
  T('First thing in the morning works for most people: it’s quiet, and it’s easy to make a habit of. Any time is fine, though, as long as it’s roughly the same time each day.', ink2, 15, HELV, pad='8px 72px 24px'),
  H('Where should I practise?', ink2, 26, HELV, pad='24px 40px 8px'), ill('spot-chair', 'A comfortable armchair', '170px', '8px 40px'),
  T('No need to climb a mountain. Anywhere reasonably quiet where you won’t be interrupted for ten minutes is just right.', ink2, 15, HELV, pad='8px 72px 40px'), bg=warm, pad='0px'),
], foot('#FFFFFF', '#777777', 'Questions? Reply to this email and a real person will answer.', 'Quiet Hour · 12 Example Street, Your City', WHY), '#F1F1F1', ['onboarding', 'lesson', 'education'])

# --- Commerce ---
deep, cream2, tomato = '#1F4D35', '#F7EFE3', '#F0532A'
studio('kick-start', 'Kick Start', 'Commerce', 'A product promotion led by oversized two-colour type, with three illustrated ways to use it.',
 'Free hot sauce with breakfast this week', 'Add a sachet to any hot breakfast, on us.', [
 row(H('MORNING PANTRY', deep, 18, BLACK, pad='40px 40px 0px'), T('Give your breakfast a', '#3B3B3B', 24, HELV, pad='24px 40px 0px'),
  H('<span style="color:#F0532A">little</span><br>heat', deep, 96, SERIF, pad='0px 40px'), ill('sachets', 'Two hot sauce sachets, one green and one black', '520px', '0px 40px'),
  T('This week, add a sachet of our green or smoky chilli sauce to any hot breakfast order. It’s on the house.', '#3B3B3B', 17, HELV, pad='12px 72px'),
  pill('Try it now', deep, pad='20px 40px 36px'), H('Three ways to try it', deep, 28, SERIF, pad='16px 40px 16px'), bg=cream2, pad='0px'),
 weighted([[ill('dish-eggs', 'Two fried eggs on a plate', '100%', '0px 8px 0px 24px'), T('<strong>Fried eggs</strong><br>A few drops on the yolk.', deep, 14, HELV, pad='8px 8px 32px 24px')],
           [ill('dish-toast', 'Toast with smashed avocado', '100%', '0px 8px'), T('<strong>Avocado toast</strong><br>Stir it into the avocado.', deep, 14, HELV, pad='8px 8px 32px')],
           [ill('dish-bowl', 'A noodle bowl with chopsticks', '100%', '0px 24px 0px 8px'), T('<strong>Noodle bowl</strong><br>A spoonful in the broth.', deep, 14, HELV, pad='8px 24px 32px 8px')]], [1, 1, 1], cream2, '0px'),
], foot(cream2, '#6B6B6B', 'One free sachet per hot breakfast order, while stocks last, until Sunday.', ADDRESS, WHY), deep, ['promotion', 'product', 'typographic'])

green, sky = '#58C24A', '#DDF1FF'
studio('two-tests', 'Two For One', 'Commerce', 'An offer email built from rounded cards: an introduction, three reasons, and a bright offer block with a price.',
 'Hi Maria, two tests for $49 each', 'Book two English tests together and save $20.', [
 row(H('fluentry', green, 30, TREB, pad='32px 40px 0px'), ill('parrot', 'A round green bird, the Fluentry mascot', '140px', '16px 40px 0px'), H('Hi Maria!', '#2B2B2B', 30, TREB, pad='8px 40px 0px'),
  T('Planning to study somewhere English-speaking? The Fluentry English Test is a quick way to prove your level.', '#2B2B2B', 15, TREB, pad='8px 56px 24px'), bg='#FFFFFF', pad='0px'),
 weighted([[T('<strong>The Fluentry English Test</strong><br><br>Take it from home in about an hour, and get your results in two days. Accepted by thousands of universities and employers.', '#2B2B2B', 15, TREB, 'left', '28px 12px 8px 40px'), pill('FIND OUT MORE', '#1CB0F6', align='left', pad='8px 12px 28px 40px', size='13px', inner='10px 20px')],
           [ill('bubble-globe', 'A speech bubble and a small globe', '100%', '24px 40px 24px 0px')]], [3, 2], '#F6F6F6', '0px'),
 row(H('Why take the Fluentry test?', '#2B2B2B', 20, TREB, pad='28px 40px 8px'),
  T('<span style="color:#3E9E31"><strong>At home:</strong></span> take it when it suits you.<br><span style="color:#3E9E31"><strong>Good value:</strong></span> one test costs $59.<br><span style="color:#3E9E31"><strong>Widely accepted:</strong></span> by universities and employers worldwide.', '#2B2B2B', 15, TREB, 'left', '8px 72px 24px'), bg='#FFFFFF', pad='0px'),
 row(H('Book two tests at $49 each', '#FFFFFF', 22, TREB, pad='32px 40px 8px'),
  H('<span style="background:#FFFFFF;color:#2B2B2B;padding:10px 20px;border-radius:12px">2 tests · $49 each · <s>$118</s> $98</span>', '#2B2B2B', 20, TREB, pad='16px 40px 16px'),
  pill('GET THE OFFER', '#FFFFFF', '#1CB0F6', pad='12px 40px 36px', size='14px'), bg=green, pad='0px'),
], foot('#FFFFFF', '#777777', 'Offer ends 30 November. Both tests must be booked on the same account.', ADDRESS, WHY), sky, ['offer', 'pricing', 'education'])

orange2, acid, lilac = '#FF5A28', '#EAF542', '#EFC6F0'
studio('give-get', 'Give One, Get One', 'Commerce', 'A loud referral email in blocks of colour: a big headline band, a line-drawn illustration, and one button.',
 'Give a jar, get a jar', 'Share your link and you both get a free jar.', [
 row(H('CHILLI CLUB', acid, 52, BLACK, pad='40px 40px'), bg=orange2, pad='0px'),
 row(H('GIVE A JAR,<br>GET A JAR', orange2, 44, BLACK, align='left', pad='36px 64px'), bg=acid, pad='0px'),
 row(ill('jar-swap', 'Two jars of chilli sauce being swapped over a table of line-drawn bowls and chopsticks'), bg=lilac, pad='0px'),
 row(T('Thank you to everyone who told a friend, a flatmate or a fussy uncle about us. Now we’d like to give a little back.<br><br>For every friend you refer, you’ll get <strong style="color:#B0204A">ONE FREE JAR*</strong> on your next order, and so will they.<br><br>Press the button below to get your referral link.', orange2, 16, TREB, 'left', '40px 64px 8px'),
  pill('REFER FRIENDS', acid, orange2, outline=orange2, font=BLACK, pad='20px 40px 4px', size='18px'), T('*worth $6 off your next order', orange2, 12, TREB, pad='4px 40px 36px'), bg='#F8F2DC', pad='0px'),
], foot(orange2, '#FFFFFF', '<strong>CHILLICLUB.EXAMPLE</strong>', 'Chilli Club · 12 Example Street, Your City', WHY), '#A50C42', ['referral', 'bold', 'illustrated'])

pink, indigo = '#F4B6C2', '#0B2A7A'
studio('cheers', 'A Toast', 'Commerce', 'A cross-sell offer: a colour block with a big serif headline, short bullet points and a pill button, then a food photo.',
 '$20 off your first wine box', 'Add a monthly wine box to your meal kit and save $20.', [
 row(H('Supper Box', indigo, 22, SERIF, pad='28px 40px 0px'), H('A toast:<br>$20 off!*', indigo, 56, SERIF, pad='16px 40px 8px'),
  T('<strong>Add our monthly wine box to your meal kit.</strong>', indigo, 16, HELV, pad='8px 64px 4px'),
  T('• Six bottles of your choice each month<br>• Each bottle sized for two<br>• A pairing card for every recipe', indigo, 15, HELV, pad='4px 64px'), pill('Add the wine box', indigo, pad='20px 40px 32px', size='17px'), bg=pink, pad='0px'),
 row(picture('table', 'wide'), bg=pink, pad='0px'),
], row(T('GET THE <strong>SUPPER BOX</strong> APP', '#AFC4FF', 13, HELV, pad='32px 40px 8px'),
  T('<a href="https://example.com" style="color:#FFFFFF;border:1px solid #FFFFFF;border-radius:999px;padding:10px 22px;text-decoration:none">iPhone app</a>&nbsp;&nbsp;&nbsp;<a href="https://example.com" style="color:#FFFFFF;border:1px solid #FFFFFF;border-radius:999px;padding:10px 22px;text-decoration:none">Android app</a>', '#FFFFFF', 14, HELV, pad='16px 40px 24px'),
  T('*Wine box not available everywhere. For first-time wine subscribers only.', '#FFFFFF', 12, HELV, pad='8px 48px'), T('Supper Box · 12 Example Street, Your City · <a href="https://example.com" style="color:#FFFFFF">Unsubscribe</a>', '#FFFFFF', 12, HELV, pad='8px 48px'),
  T('<a href="https://xem.email" style="color:#FFFFFF">Built with Xem</a>', '#FFFFFF', 12, HELV, pad='8px 48px 32px'), bg=indigo, pad='0px'),
 '#FFFFFF', ['cross-sell', 'offer', 'subscription'])

# --- Events ---
studio('one-day-to-go', 'One Day to Go', 'Events', 'A day-before event reminder with a bold banner, the practical details as a short list, and a friendly sign-off.',
 'Tomorrow: your workshop starts at 10am', 'Here’s what you need to know before we go live.', [
 row(H('Postwell', '#1A1A1A', 34, BLACK, pad='40px 40px 24px'), ill('one-day-banner', 'A red banner reading One day to go, with a lightning bolt, a heart and a globe', pad='0px 40px'),
  T('Tomorrow’s the day. We’re going live with three social media managers from small shops to show you how they plan a month of posts in an afternoon.<br><br>Here’s what you need to know:', '#1A1A1A', 17, HELV, 'left', '32px 40px 8px'),
  T('•&nbsp; The workshop starts at <strong>10am Pacific</strong>.<br>•&nbsp; It’s on a video call. We’ll email you the link an hour before.<br>•&nbsp; Bring one post you’d like feedback on.', '#1A1A1A', 17, HELV, 'left', '8px 56px'),
  T('Know someone who’d find it useful? Send them <a href="https://example.com" style="color:#0E7C86">this link</a>. It’s not too late to join.<br><br>See you tomorrow!', '#1A1A1A', 17, HELV, 'left', '8px 40px 40px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#777777', 'Postwell · 12 Example Street, Your City', 'You’re getting this because you registered for the workshop.'), '#FFFFFF', ['reminder', 'workshop', 'webinar'])

# --- Product updates and education ---
studio('month-in-review', 'What’s New This Month', 'Product', 'A monthly product round-up: a centred header, a signed note from a person, then one large feature and two side-by-side ones.',
 'What’s new in Corkboard this April', 'Tables, templates and slides, plus a note from our product lead.', [
 row(H('<span style="background:#FFD02F;padding:4px 10px;border-radius:6px">▦</span> corkboard', '#1A1A1A', 24, HELV, pad='32px 40px 8px'), T('April ’26', '#1A1A1A', 18, HELV, pad='16px 40px 0px'),
  H('What’s New in Corkboard', '#1A1A1A', 34, HELV, pad='4px 40px 8px'), T('Tables, templates and slides. Three more ways to plan work, keep it moving and show it off.', '#1A1A1A', 15, HELV, pad='4px 72px'),
  button('April round-up', '#3D5AFE', align='center'), bg='#F4F4F8', pad='0px'),
 row(T('Hi {{first_name}},<br><br>It was a busy month. If you missed our launch week, <a href="https://example.com" style="color:#3D5AFE">here are the highlights</a>.<br><br>Now it’s time to use them. You can sort work in Tables, start a project from a template, and turn a board into slides without leaving the page.', '#1A1A1A', 15, HELV, 'left', '36px 40px 8px'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('avatar-1', 'Portrait of Jo Park', '72px', '12px 8px 28px 40px')], [T('Jo Park,<br>Head of Product', '#1A1A1A', 14, HELV, 'left', '24px 40px 28px 8px')]], [1, 4], '#FFFFFF', '0px'),
 row(divider('#ECECF2'), ill('ui-table', 'A table view with rows, coloured status labels and a teammate’s cursor', pad='16px 40px'), H('Plan it, sort it, track it in Tables', '#1A1A1A', 26, HELV, 'left', '16px 40px 4px'),
  T('Bring order to a busy board. Drop sticky notes or cards straight into a table, then switch to a timeline to see what’s due when.', '#1A1A1A', 15, HELV, 'left', '4px 40px'),
  T('<a href="https://example.com" style="background:#E9E9F0;color:#1A1A1A;padding:12px 16px;text-decoration:none;font-weight:bold">See how Tables work</a>&nbsp;&nbsp;&nbsp;<a href="https://example.com" style="color:#1A1A1A;text-decoration:none;font-weight:bold">Create a table →</a>', '#1A1A1A', 14, HELV, 'left', '20px 40px 32px'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('ui-templates', 'A stack of board templates with a Template label', '100%', '24px 8px 24px 40px')],
           [H('New project? Start from a template', '#1A1A1A', 20, HELV, 'left', '32px 40px 4px 12px'), T('✨ <em><strong>BETA</strong></em>', '#1A1A1A', 13, HELV, 'left', '4px 40px 4px 12px'), T('Skip the setup with ready-made spaces, complete with boards and tools.', '#1A1A1A', 14, HELV, 'left', '4px 40px 4px 12px'), T('<a href="https://example.com" style="color:#3D5AFE;font-weight:bold;text-decoration:none">Have a look →</a>', '#1A1A1A', 14, HELV, 'left', '4px 40px 24px 12px')]], [1, 1], '#FFFFFF', '0px'),
 weighted([[H('Ready, set, present with Slides', '#1A1A1A', 20, HELV, 'left', '32px 12px 4px 40px'), T('✨ <em><strong>BETA</strong></em>', '#1A1A1A', 13, HELV, 'left', '4px 12px 4px 40px'), T('Turn any board into slides, with live widgets that keep everyone looking at the screen.', '#1A1A1A', 14, HELV, 'left', '4px 12px 4px 40px'), T('<a href="https://example.com" style="color:#3D5AFE;font-weight:bold;text-decoration:none">Try Slides →</a>', '#1A1A1A', 14, HELV, 'left', '4px 12px 32px 40px')],
           [ill('ui-slides', 'A slide titled Our Spring Plan with a portrait drawing', '100%', '24px 40px 24px 8px')]], [1, 1], '#FFFFFF', '0px'),
], foot('#F4F4F8', '#777777', 'Corkboard · 12 Example Street, Your City', WHY), '#3D5AFE', ['round-up', 'product update', 'features'])

paperw, inkw = '#F9F7F2', '#1F1E1D'
studio('try-it-on', 'Try It on Real Work', 'Product', 'A calm product nudge on warm paper: an illustration, two example uses in rounded cards, and one dark button.',
 'Try Marlowe on your own work', 'Point it at a folder or an inbox and see what comes back.', [
 row(H('<span style="color:#D97757">✳</span> Marlowe', inkw, 26, SERIF, 'left', '32px 32px 16px'), H('Try Marlowe on your own work', inkw, 26, SERIF, 'left', '8px 32px 4px'),
  T('The quickest way to see what it can do is to give it something real.', inkw, 15, SERIF, 'left', '4px 32px 16px'), ill('app-tiles', 'Rounded app tiles scattered on a sage background', pad='0px 32px'),
  T('Pick whichever suits you:', inkw, 18, SERIF, 'left', '20px 32px 8px'), bg=paperw, pad='0px'),
 weighted([[ill('icon-files', 'A stack of files', '56px', '28px 8px 28px 40px')], [H('From your files', inkw, 17, SERIF, 'left', '28px 40px 4px 8px'), T('<em>“Read everything in this folder. Tell me what’s there, flag anything urgent, and write it up as one tidy page.”</em>', inkw, 15, SERIF, 'left', '4px 40px 28px 8px')]], [1, 4], '#EFEBE3', '0px'),
 row(text('', 4, inkw, pad='4px 0px'), bg=paperw, pad='0px'),
 weighted([[ill('icon-inbox', 'An inbox', '56px', '28px 8px 28px 40px')], [H('From your email', inkw, 17, SERIF, 'left', '28px 40px 4px 8px'), T('<em>“Find everything from the past week that still needs an answer. Put the urgent ones first and draft replies.”</em>', inkw, 15, SERIF, 'left', '4px 40px 28px 8px')]], [1, 4], '#EFEBE3', '0px'),
 row(T('Marlowe reads your files, email or messages, does the work, and hands back something finished.', inkw, 15, SERIF, 'left', '20px 32px 8px'), button('Try it now', '#141414', align='left'), bg=paperw, pad='0px'),
], row(H('<span style="color:#D97757">✳</span> Marlowe', '#FFFFFF', 22, SERIF, 'left', '32px 32px 8px'), T('Marlowe · 12 Example Street, Your City · <a href="https://example.com" style="color:#BBBBBB">Unsubscribe</a>', '#BBBBBB', 12, HELV, 'left', '8px 32px'), T('<a href="https://xem.email" style="color:#BBBBBB">Built with Xem</a>', '#BBBBBB', 12, HELV, 'left', '4px 32px 32px'), bg='#141414', pad='0px'),
 '#DADFD9', ['product', 'activation', 'examples'])

studio('next-steps', 'Your Next Steps', 'Product', 'An advanced-features email: three white cards on warm grey, each with a line icon, a bold title, a sentence and an underlined link.',
 'Ready for the advanced features?', 'Search, citations and skills: three ways to get more from Marlowe.', [
 row(H('<span style="color:#D97757">✳</span> Marlowe', inkw, 26, SERIF, 'left', '36px 40px 16px'), H('Go further with Marlowe', inkw, 28, HELV, 'left', '8px 40px 4px'),
  T('You’ve got the basics. Here are three features that make the biggest difference once you’re building with it every day.', inkw, 17, HELV, 'left', '4px 40px 20px'), bg='#F0EEE6', pad='0px'),
 weighted([[ill('icon-globe', 'A globe', '60px', '28px 8px 28px 48px')], [H('Look things up on the web', inkw, 18, HELV, 'left', '28px 48px 4px 8px'), T('Let it search the web for current answers, with links to every source it used.', inkw, 15, HELV, 'left', '4px 48px 4px 8px'), T('<a href="https://example.com" style="color:#1F1E1D;font-weight:bold">Turn on web search</a>', inkw, 15, HELV, 'left', '8px 48px 28px 8px')]], [1, 4], '#FAF9F5', '0px'),
 row(text('', 4, inkw, pad='4px 0px'), bg='#F0EEE6', pad='0px'),
 weighted([[ill('icon-pencil', 'A pencil', '60px', '28px 8px 28px 48px')], [H('Check where answers come from', inkw, 18, HELV, 'left', '28px 48px 4px 8px'), T('Every answer can point back to the exact passage in your documents. Upload them and the references come with it.', inkw, 15, HELV, 'left', '4px 48px 4px 8px'), T('<a href="https://example.com" style="color:#1F1E1D;font-weight:bold">Learn more</a>', inkw, 15, HELV, 'left', '8px 48px 28px 8px')]], [1, 4], '#FAF9F5', '0px'),
 row(text('', 4, inkw, pad='4px 0px'), bg='#F0EEE6', pad='0px'),
 weighted([[ill('icon-toolbox', 'A toolbox', '60px', '28px 8px 28px 48px')], [H('Teach it your way of working', inkw, 18, HELV, 'left', '28px 48px 4px 8px'), T('Skills are folders of instructions and examples it picks up when they’re relevant, so it does things the way your team does.', inkw, 15, HELV, 'left', '4px 48px 4px 8px'), T('<a href="https://example.com" style="color:#1F1E1D;font-weight:bold">Learn more</a>', inkw, 15, HELV, 'left', '8px 48px 28px 8px')]], [1, 4], '#FAF9F5', '0px'),
 row(pill('Explore advanced features', '#141414', pad='28px 40px 40px', size='18px', inner='14px 48px'), bg='#F0EEE6', pad='0px'),
], row(H('<span style="color:#D97757">✳</span> Marlowe', '#FFFFFF', 22, SERIF, 'left', '32px 40px 8px'), divider('#444444'), T('Marlowe · 12 Example Street, Your City. To stop these emails, <a href="https://example.com" style="color:#BBBBBB">unsubscribe</a>.', '#BBBBBB', 12, HELV, 'left', '8px 40px'), T('<a href="https://xem.email" style="color:#BBBBBB">Built with Xem</a>', '#BBBBBB', 12, HELV, 'left', '4px 40px 32px'), bg='#141414', pad='0px'),
 '#FFFFFF', ['features', 'education', 'cards'])

bl = '#2E5BFF'
studio('three-reasons', 'Three Good Reasons', 'Product', 'An educational email: a blue illustrated banner, a bold headline, then three numbered points in blue circles.',
 'Three reasons to split your savings', 'One pot is fine. Three pots work harder.', [
 row(H('ledgerly', bl, 26, HELV, 'left', '24px 40px'), bg='#FFFFFF', pad='0px'),
 row(ill('savings-banner', 'A blue banner with a magnifying glass, notes, coins and sparkles'), bg=bl, pad='0px'),
 row(H('Split your savings into pots', '#0B0B0B', 30, HELV, 'left', '36px 40px 12px'), T('Saving is easier to stick with when every pound has a job. Splitting your savings into a few pots can help you:', '#0B0B0B', 16, HELV, 'left', '4px 40px 16px'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('num-1', 'Number 1', '52px', '16px 0px 16px 40px')], [H('Cover the bills first', '#0B0B0B', 20, HELV, 'left', '16px 40px 2px 8px'), T('Set aside rent and bills on payday, so the rest of your account is yours to spend.', '#0B0B0B', 15, HELV, 'left', '2px 40px 16px 8px')]], [1, 5], '#FFFFFF', '0px'),
 weighted([[ill('num-2', 'Number 2', '52px', '16px 0px 16px 40px')], [H('Build a cushion', '#0B0B0B', 20, HELV, 'left', '16px 40px 2px 8px'), T('A separate pot for the unexpected means a broken boiler doesn’t come out of your holiday.', '#0B0B0B', 15, HELV, 'left', '2px 40px 16px 8px')]], [1, 5], '#FFFFFF', '0px'),
 weighted([[ill('num-3', 'Number 3', '52px', '16px 0px 16px 40px')], [H('Save for something fun', '#0B0B0B', 20, HELV, 'left', '16px 40px 2px 8px'), T('Name a pot after the thing you want. It’s easier to leave it alone when it says “Lisbon”.', '#0B0B0B', 15, HELV, 'left', '2px 40px 16px 8px')]], [1, 5], '#FFFFFF', '0px'),
 row(pill('Set up your pots', bl, align='left', pad='20px 40px 40px', size='17px', inner='16px 36px'), bg='#FFFFFF', pad='0px'),
], foot('#EEF0F4', '#5B6170', 'This email is general information, not financial advice.', 'Ledgerly · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Get in touch</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#FFFFFF', ['education', 'numbered', 'finance'])

blue2 = '#2F6FED'
studio('join-your-team', 'Join Your Team', 'Welcome', 'A team invitation with a flat illustration, a full-width button and two short testimonials from other users.',
 'Sana invited you to Threadline', 'Your team is already talking. Join them in one click.', [
 weighted([[H('<span style="color:#2F6FED">▰</span> threadline', blue2, 24, HELV, 'left', '28px 8px 8px 40px')], [T('<a href="https://example.com" style="color:#888888;text-decoration:none">Open Threadline</a>', '#888888', 15, HELV, 'right', '32px 40px 8px 8px')]], [1, 1], '#FFFFFF', '0px'),
 row(ill('team-parade', 'Two cheerful figures, one playing a horn and one waving a flag', '440px', '16px 40px'), H('Join your team’s conversations', '#333333', 24, HELV, 'left', '24px 40px 8px'),
  T('<strong>Sana Iqbal</strong> invited you to join the <strong>Studio</strong> team on Threadline three days ago. Press the button to start working with them:', '#555555', 16, HELV, 'left', '4px 40px 16px'), bg='#FFFFFF', pad='0px'),
 row(button('Join your team →', blue2, align='center'), bg='#FFFFFF', pad='0px'),
 row(T('<strong>Not sure about Threadline?</strong> Here’s what other teams say:', '#555555', 16, HELV, 'left', '16px 40px 8px'),
  T('<span style="display:block;border-left:3px solid #2F6FED;padding-left:16px">“We stopped losing decisions in chat. Threads keep each topic in one place, so I can catch up in ten minutes instead of an hour.”</span>', '#666666', 16, HELV, 'left', '8px 40px'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('avatar-2', 'Portrait of Dami Okafor', '48px', '4px 4px 12px 40px')], [T('<strong>— Dami Okafor</strong><br><span style="color:#888888">Operations Lead, Fernway</span>', '#333333', 13, HELV, 'left', '6px 40px 12px 4px')]], [1, 6], '#FFFFFF', '0px'),
 row(T('<span style="display:block;border-left:3px solid #2F6FED;padding-left:16px">“Our inboxes are quieter. We spend less time searching for old messages and more time on the work.”</span>', '#666666', 16, HELV, 'left', '8px 40px'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('avatar-3', 'Portrait of Ellie Shaw', '48px', '4px 4px 12px 40px')], [T('<strong>— Ellie Shaw</strong><br><span style="color:#888888">Director, Northgate Coaching</span>', '#333333', 13, HELV, 'left', '6px 40px 12px 4px')]], [1, 6], '#FFFFFF', '0px'),
 row(T('Our best,<br>The Threadline team', '#555555', 16, HELV, 'left', '16px 40px 36px'), bg='#FFFFFF', pad='0px'),
 weighted([[T('<strong>Still not sure?</strong>', '#333333', 15, HELV, 'left', '20px 8px 20px 40px')], [T('<a href="https://example.com" style="color:#2F6FED;text-decoration:none">See what makes Threadline different →</a>', '#2F6FED', 15, HELV, 'right', '20px 40px 20px 8px')]], [2, 3], '#FFFFFF', '12px 0px 0px'),
], foot('#F4F5F7', '#888888', '<a href="https://example.com" style="color:#2F6FED">Blog</a> · <a href="https://example.com" style="color:#2F6FED">Help centre</a> · <a href="https://example.com" style="color:#2F6FED">Unsubscribe from tips</a>'), '#F4F5F7', ['invitation', 'team', 'testimonials'])

teal2 = '#0F3A40'
studio('stay-connected', 'Stay Connected', 'Retention', 'A re-permission email asking subscribers to opt in again, with an illustrated hero and three reasons on a dark panel.',
 'Do you still want to hear from us?', 'Opt in to keep getting our news, offers and invitations.', [
 row(H('◣ ◢', teal2, 28, BLACK, pad='28px 40px'), bg='#F2F0EE', pad='0px'),
 weighted([[H('Stay<br>Connected', teal2, 40, HELV, 'left', '56px 8px 16px 40px'), button('Opt in now', '#5AA9BE', align='left')], [ill('clock-walker', 'A pocket watch with arms and legs, striding across a pink circle', '100%', '32px 40px 24px 8px')]], [1, 1], '#FFFFFF', '0px'),
 row(T('You’re getting this email from Deskwise because you told us you’d like to hear from us.<br><br>We’re tidying up our mailing list to protect your personal information, and we don’t want you to miss our product news, events and practical guides for support teams.<br><br><strong>Please <a href="https://example.com" style="color:#5AA9BE;text-decoration:none">opt in</a> to keep hearing from Deskwise.</strong>', teal2, 16, HELV, 'left', '16px 40px 40px'), bg='#FFFFFF', pad='0px'),
 row(H('3 reasons to stay in touch:', '#FFFFFF', 26, HELV, pad='36px 40px 8px'), bg=teal2, pad='0px'),
 weighted([[ill('icon-chart', 'A line chart', '70px', '8px 0px'), T('Guides and<br>industry insights', '#FFFFFF', 15, HELV, pad='8px 8px 36px')], [ill('icon-cake', 'A birthday cake', '70px', '8px 0px'), T('Offers made<br>for you', '#FFFFFF', 15, HELV, pad='8px 8px 36px')], [ill('icon-ticket', 'An event ticket', '70px', '8px 0px'), T('Invitations to<br>our events', '#FFFFFF', 15, HELV, pad='8px 8px 36px')]], [1, 1, 1], teal2, '0px'),
], foot('#F2F0EE', teal2, '<strong>deskwise</strong>', 'Deskwise · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Privacy policy</a> · <a href="https://example.com" style="color:inherit">Manage preferences</a>'), '#F2F0EE', ['re-permission', 'opt in', 'gdpr'])

ns = '#E9E7E2'
studio('playbook', 'The Playbook', 'Product', 'A learning digest in white rounded cards: a tagged lead with a chart and two buttons, then image-and-text cards.',
 'The playbook: follow-ups that get replies', 'A short course, a masterclass and a live session.', [
 row(H('<span style="color:#1A1A1A">✲</span> Northstar', '#1A1A1A', 24, HELV, 'left', '24px 32px 16px'), bg=ns, pad='0px'),
 row(T('<span style="background:#E7F33A;padding:6px 14px;border-radius:999px;font-size:12px">THE PLAYBOOK</span>', '#1A1A1A', 12, HELV, pad='32px 32px 8px'), H('How to write follow-ups that get replies', '#1A1A1A', 24, HELV, pad='8px 32px 16px'),
  ill('chart-card', 'A bar chart: 12% reply rate with no follow-ups, 38% with one, 61% with three', pad='0px 32px'),
  T('Most replies come after the first email. In this short course you’ll learn when to follow up, what to say, and how to stop before it gets annoying.', '#444444', 14, HELV, pad='16px 48px 8px'),
  T('<a href="https://example.com" style="background:#1F2A28;color:#FFFFFF;padding:12px 18px;border-radius:6px;text-decoration:none">Watch now</a>&nbsp;&nbsp;<a href="https://example.com" style="border:1px solid #1A1A1A;color:#1A1A1A;padding:11px 17px;border-radius:6px;text-decoration:none">Build a sequence</a>', '#1A1A1A', 14, HELV, pad='16px 32px 36px'), bg='#FFFFFF', pad='0px'),
 row(text('', 4, '#1A1A1A', pad='6px 0px'), bg=ns, pad='0px'),
 weighted([[ill('list-card', 'A card ranking four subject lines by open rate, with a speaker portrait', '100%', '28px 8px 28px 28px')], [H('Seven parts of a good first email', '#1A1A1A', 22, HELV, 'left', '40px 32px 8px 12px'), T('Rosa Lind has written outreach for twenty years. In this guide she breaks down the seven things her best-performing emails have in common.', '#444444', 14, HELV, 'left', '4px 32px 8px 12px'), T('<a href="https://example.com" style="border:1px solid #1A1A1A;color:#1A1A1A;padding:11px 17px;border-radius:6px;text-decoration:none">Learn more</a>', '#1A1A1A', 14, HELV, 'left', '16px 32px 28px 12px')]], [1, 1], '#FFFFFF', '0px'),
 row(text('', 4, '#1A1A1A', pad='6px 0px'), bg=ns, pad='0px'),
 weighted([[H('Closing bigger deals with more people in the room', '#1A1A1A', 22, HELV, 'left', '40px 12px 8px 32px'), T('A live session on keeping five decision-makers aligned without twenty extra meetings.', '#444444', 14, HELV, 'left', '4px 12px 8px 32px'), T('<a href="https://example.com" style="border:1px solid #1A1A1A;color:#1A1A1A;padding:11px 17px;border-radius:6px;text-decoration:none">Save your seat</a>', '#1A1A1A', 14, HELV, 'left', '16px 12px 28px 32px')], [ill('deals-card', 'A pink gradient card with a play button and a Live tag', '100%', '28px 28px 28px 8px')]], [1, 1], '#FFFFFF', '0px'),
], foot(ns, '#666666', 'Northstar · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Manage preferences</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), ns, ['digest', 'education', 'cards'])

night, yellow = '#0F1716', '#FFF02B'
SPONSORS = ['Lumen', 'Orbitly', 'Kitewire', 'Fieldnote', 'Parcelly', 'Brightfork', 'Tidepool', 'Quarry']
def sponsor_rows(names):
 cell = lambda n: [T(f'<span style="display:block;border:1px solid #3A4644;border-radius:8px;padding:14px 4px">{n}</span>', '#D6DAD8', 13, HELV, pad='4px 6px')]
 return [weighted([cell(n) for n in names[i:i+4]], [1]*len(names[i:i+4]), night, '0px') for i in range(0, len(names), 4)]

studio('last-chance', 'Last Chance', 'Events', 'A dark conference email for the last day of early pricing: gradient banner, details, a bulleted lineup and a sponsor grid.',
 'Last chance for $79 Northstar Live tickets', 'Early prices end at midnight. Here’s who you’ll see.', [
 row(H('✲ Northstar', '#FFFFFF', 22, HELV, pad='28px 40px 16px'), ill('event-gradient', 'A bright gradient banner reading Northstar Live, with a Register now button', pad='0px 32px'),
  H('Last chance for $79 tickets. The sale ends at midnight.', '#FFFFFF', 26, HELV, 'left', '24px 40px 8px'),
  T('<strong>Today is the last day to get Northstar Live tickets for $79, before the price goes up to $199.</strong>', '#FFFFFF', 15, HELV, 'left', '4px 40px'),
  T('On 30 September, sales leaders from forty companies are getting together in one room to share what’s working, and what they’ve stopped doing.', '#D6DAD8', 15, HELV, 'left', '4px 40px'),
  T('<strong style="color:#FFFFFF">Northstar Live</strong><br><span style="color:#8E9895">Date:</span> Wednesday 30 September<br><span style="color:#8E9895">Venue:</span> The Assembly Rooms, Your City<br><span style="color:#8E9895">Time:</span> Doors 8am, first talk 9:15', '#D6DAD8', 15, HELV, 'left', '8px 40px'),
  button('Register now for $79', yellow, align='left', ink='#111111'), T('The <a href="https://example.com" style="color:#FFF02B">full agenda</a> and <a href="https://example.com" style="color:#FFF02B">speaker list</a> are out now.', '#D6DAD8', 15, HELV, 'left', '0px 40px 16px'), divider('#2A3533'),
  H('Here’s what’s on the programme:', '#FFFFFF', 24, HELV, 'left', '16px 40px 8px'),
  T('<span style="color:#7BE0A6">●</span> <strong style="color:#FFFFFF">Opening keynote:</strong> what our customers changed this year, and what it did to their numbers.<br><br><span style="color:#FF7AB8">●</span> <strong style="color:#FFFFFF">Building on new tools:</strong> a panel on fitting AI into a sales team without breaking what works.<br><br><span style="color:#B48CFF">●</span> <strong style="color:#FFFFFF">Customer panel:</strong> four sales leaders on hiring, outbound and where human judgement still wins.<br><br><span style="color:#7BE0A6">●</span> <strong style="color:#FFFFFF">Breakout sessions:</strong> hands-on workshops, from writing sequences to running forecasts.', '#D6DAD8', 15, HELV, 'left', '8px 40px'),
  button('Register now for $79', yellow, align='left', ink='#111111'), T('<strong style="color:#FFFFFF">Book your travel early:</strong> hotels near the venue fill up fast in September.', '#D6DAD8', 15, HELV, 'left', '0px 40px 16px'),
  H('Partners and sponsors', '#C9CFCD', 20, SERIF, pad='16px 40px 8px'), bg=night, pad='0px'),
 *sponsor_rows(SPONSORS),
 row(T('See you there,<br><strong style="color:#FFFFFF">The Northstar team</strong>', '#D6DAD8', 16, HELV, 'left', '28px 40px 24px'), bg=night, pad='0px'),
], foot(night, '#8E9895', '<a href="https://example.com" style="color:inherit">Manage preferences</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>', 'Northstar · 12 Example Street, Your City'), night, ['conference', 'tickets', 'dark'])

studio('two-days-left', 'Two Days Left', 'Events', 'A dark event email built around the speakers: a gradient banner, a portrait strip, and a programme listed by session.',
 '2 days left for $99 Northstar Live tickets', 'Meet the speakers before prices go up.', [
 row(H('✲ Northstar', '#FFFFFF', 22, HELV, pad='28px 40px 16px'), ill('event-gradient', 'A bright gradient banner reading Northstar Live', pad='0px 32px'),
  H('Two days left: Northstar Live tickets for $99', '#FFFFFF', 24, HELV, 'left', '24px 40px 4px'), T('In person on 30 September at The Assembly Rooms, Your City', '#8E9895', 13, HELV, 'left', '0px 40px 8px'),
  T('We’d love to see you there. <strong style="color:#FFFFFF">You’ve got two days to book at $99 before the price goes up to $199.</strong>', '#D6DAD8', 15, HELV, 'left', '4px 40px'),
  pill('Register now', yellow, '#111111', align='left', pad='12px 40px 24px', size='14px', inner='12px 20px'), divider('#2A3533'),
  ill('speaker-strip', 'A row of six illustrated speaker portraits on bright colours', pad='16px 0px'),
  H('Who’s speaking', '#FFFFFF', 24, HELV, 'left', '8px 40px 4px'), T('Sales and revenue leaders from Lumen, Orbitly, Kitewire and more, with a comedian hosting between sessions.', '#D6DAD8', 15, HELV, 'left', '4px 40px'),
  pill('See the speakers', yellow, '#111111', align='left', pad='12px 40px 24px', size='14px', inner='12px 20px'), divider('#2A3533'),
  H('Main stage', '#FFFFFF', 24, HELV, 'left', '16px 40px 4px'),
  T('<strong style="color:#AEB7B4">Opening keynote</strong><br>The year in numbers, and the product we’ve been building quietly.<br><br><strong style="color:#AEB7B4">Customer panel</strong><br>Four leaders on hiring, outbound and when to trust the data.<br><br><strong style="color:#AEB7B4">In conversation</strong><br>A founder on growing from five customers to five hundred.', '#D6DAD8', 15, HELV, 'left', '4px 40px 16px'),
  H('Workshops', '#FFFFFF', 24, HELV, 'left', '8px 40px 4px'),
  T('<strong style="color:#AEB7B4">Writing a sequence that gets replies</strong><br>Bring a draft; leave with a better one.<br><br><strong style="color:#AEB7B4">Forecasting without guesswork</strong><br>A practical session with real pipeline data.', '#D6DAD8', 15, HELV, 'left', '4px 40px'),
  pill('See the full agenda', yellow, '#111111', align='left', pad='12px 40px 24px', size='14px', inner='12px 20px'), H('Partners and sponsors', '#C9CFCD', 20, SERIF, pad='16px 40px 8px'), bg=night, pad='0px'),
 *sponsor_rows(SPONSORS[:8][::-1]),
 row(T('<strong style="color:#FFFFFF">Prices go up to $199 in two days.</strong> <a href="https://example.com" style="color:#FFF02B">Book yours today</a> for $99.<br><br>See you there,<br><strong style="color:#FFFFFF">The Northstar team</strong>', '#D6DAD8', 15, HELV, 'left', '24px 40px 24px'), bg=night, pad='0px'),
], foot(night, '#8E9895', '<a href="https://example.com" style="color:inherit">Manage preferences</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>', 'Northstar · 12 Example Street, Your City'), night, ['conference', 'speakers', 'dark'])

# --- Batch two: more layouts modelled on reference emails ---
PHOTOS.update({
 'cat': ('photo-1514888286974-6c03e2ca1dba', 'A black and white cat peering over a ledge'),
 'collie': ('photo-1587300003388-59208cc962cb', 'A happy dog on a beach'),
 'beagle': ('photo-1543466835-00a7907e9de1', 'A smiling beagle'),
 'parrot2': ('photo-1452570053594-1b985d6ea890', 'A blue and yellow macaw'),
 'wave': ('photo-1559827260-dc66d52bef19', 'A wave breaking at sunset'),
 'gingercat': ('photo-1511044568932-338cba0ad803', 'A ginger cat stretching'),
 'pourover': ('photo-1442512595331-e89e73853f31', 'Coffee being brewed by hand at a café counter'),
 'beans': ('photo-1447933601403-0c6688de566e', 'Roasted coffee beans'),
 'bowls': ('photo-1504674900247-0877df9cc836', 'Bowls of food on a wooden table'),
 'redrock': ('photo-1500530855697-b586d89ba3ee', 'A desert road through red rock'),
 'valley': ('photo-1469474968028-56623f02e42e', 'Sunlight across a green valley'),
 'peaks': ('photo-1506905925346-21bda4d32df4', 'Mountain peaks above the clouds'),
 'laptops': ('photo-1522202176988-66273c2fd55f', 'Friends working together at laptops'),
 'hall': ('photo-1531058020387-3be344556be6', 'An audience in a large hall'),
 'house': ('photo-1505843513577-22bb7d21e455', 'A large house at dusk'),
 'papers': ('photo-1504711434969-e33886168f5c', 'A stack of newspapers'),
 'cyclists': ('photo-1517649763962-0c623066013b', 'A group of road cyclists racing'),
 'crowd': ('photo-1501386761578-eac5c94b800a', 'A cheering crowd'),
 'face1': ('photo-1494790108377-be9c29b29330', 'Portrait of a smiling woman'),
 'face2': ('photo-1500648767791-00dcc994a43e', 'Portrait of a man'),
 'face3': ('photo-1438761681033-6461ffad8d80', 'Portrait of a woman by a lake'),
 'face4': ('photo-1472099645785-5658abf4ff4e', 'Portrait of an older man'),
 'doglick': ('photo-1518717758536-85ae29035b6d', 'A dog licking its nose'),
 'hamster': ('photo-1425082661705-1834bfd09dca', 'A small hamster'),
})
def lk(copy, c, href='https://example.com', bold=False): return f'<a href="{href}" style="color:{c};{"font-weight:bold;" if bold else ""}">{copy}</a>'
def ul(items): return '<br>'.join('•&nbsp; '+i for i in items)
def P(key, shape='wide', pad='0px'): return picture(key, shape, pad)

# Webinar letter: colour-block header with a dithered portrait, then a plain letter.
lime2, plum = '#E3EA4A', '#7A3CC2'
studio('four-ways', 'Four Ways It Breaks', 'Events', 'A webinar invitation written as a plain letter, under a bright header with a pixel-art portrait.',
 'The four ways support teams break as they grow', 'Join us live on Thursday 27 August at 11am Pacific.', [
 weighted([[H('■ Relay', '#111111', 20, BLACK, 'left', '28px 8px 8px 40px'), H('The Four Ways<br>Support Teams<br>Break', '#111111', 38, HELV, 'left', '8px 8px 28px 40px')], [ill('pixel-portrait', 'A pixel-art portrait of someone at a laptop, in yellow on purple', '100%', '28px 40px 28px 8px')]], [3, 2], lime2, '0px'),
 row(T('Hi {{first_name}},<br><br>Support teams don’t get harder in one big way as they grow. They get harder in four particular ways, and most teams are living with at least one of them without a name for it.<br><br>' + lk('<strong>Join us on Thursday 27 August at 11am Pacific</strong>', '#2F55D4') + ' to go through all four. You’ll probably recognise one:<br><br><strong>The Lost Thread.</strong> By the time a question reaches someone who can answer it, the background is gone.<br><br><strong>The Bottleneck.</strong> Everything goes through one person, and they can’t take a holiday.<br><br><strong>The Reinventor.</strong> The same problems get solved from scratch every quarter.<br><br><strong>The Stretched Lead.</strong> Volume is up, the team isn’t, and quality is fraying.<br><br>You’ll leave knowing which one is yours, and what a team built to avoid it looks like.', '#1A1A1A', 16, HELV, 'left', '36px 48px 8px'),
  pill('Save your seat →', plum, pad='20px 40px 24px'), T('Cheers,<br>Maya at Relay', '#1A1A1A', 16, HELV, 'left', '8px 48px 32px'), divider('#E2E2E2'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#777777', 'Relay · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">Manage preferences</a>'), '#F2F2F2', ['webinar', 'letter', 'invitation'])

# Developer guide: gradient banner, three icon-led sections, a callout and a help line.
violet2, ink3 = '#3A2A8C', '#2B2B33'
def guide_section(icon, alt, level, title, body):
 return weighted([[ill(icon, alt, '64px', '24px 4px 8px 40px')], [T(level, '#8A8A95', 12, HELV, 'left', '24px 40px 0px 8px'), H(title, ink3, 20, HELV, 'left', '4px 40px 4px 8px'), T(body, ink3, 14, HELV, 'left', '4px 40px 24px 8px')]], [1, 5], '#FFFFFF', '0px')
studio('make-it-yours', 'Make It Yours', 'Product', 'A developer guide with a gradient banner, three icon-led sections with links, a bordered callout and a help line.',
 'Three ways to make the editor your own', 'Themes, feature switches and custom blocks.', [
 row(H('<span style="color:#7B4DFF">◆</span> blockwise <span style="font-weight:normal;color:#8A8A95">SDK</span>', ink3, 24, HELV, pad='24px 40px'), bg='#FFFFFF', pad='0px'),
 weighted([[H('Make the<br><span style="color:#7BE0A6">Blockwise</span><br>editor your own', '#FFFFFF', 30, HELV, 'left', '36px 8px 36px 40px')], [ill('ui-editor', 'An editor window with an image block and a stack of content blocks', '100%', '16px 24px 16px 0px')]], [1, 1], violet2, '0px'),
 row(T('<strong>Hi there,</strong><br><br>Customers often tell us the best thing about Blockwise is how far they can change it. Here are a few ways to make the editor <strong>look and feel like part of your product</strong>:', ink3, 15, HELV, 'left', '28px 40px 16px'), bg='#FFFFFF', pad='0px'),
 guide_section('icon-brush', 'A paintbrush', 'Level: beginner', 'Match the editor to your brand', 'People shouldn’t feel they’ve left your product to use it.<br><br>Use ' + lk('themes', '#7B4DFF') + ' for your colours and fonts, or ' + lk('custom CSS', '#7B4DFF') + ' for full control over the look.'),
 row(divider('#E6E6EC'), bg='#FFFFFF', pad='0px'),
 guide_section('icon-toggles', 'A list of switches', 'Level: intermediate', 'Choose which features to switch on (or off)', 'Want ' + lk('mobile editing', '#7B4DFF') + '? Don’t want people pasting raw HTML? Tick a box in the ' + lk('developer console', '#7B4DFF') + '.<br><br>Language, sidebar position and much more live in your ' + lk('config file', '#7B4DFF') + '.'),
 row(divider('#E6E6EC'), bg='#FFFFFF', pad='0px'),
 guide_section('icon-blocks', 'Coloured building blocks', 'Level: advanced', 'Build your own content blocks', 'Custom blocks let people pull in content from your own product: a catalogue item, an event ticket, a booking link.<br><br>' + lk('More about custom blocks →', '#7B4DFF')),
 row(T('<span style="display:block;border:1px solid #CFC6F2;border-left:4px solid #7B4DFF;border-radius:10px;padding:20px 24px"><strong style="font-size:19px;color:#3A2A8C">Create an app, then customise it</strong><br><br>Start in the ' + lk('developer console', '#7B4DFF') + ' and press Activate. That one step unlocks the builders, and the ' + lk('quickstart guides', '#7B4DFF') + ' take it from there.</span>', ink3, 14, HELV, 'left', '12px 32px'),
  bg='#FFFFFF', pad='0px'),
 weighted([[ill('icon-chat', 'A chat bubble', '36px', '20px 4px 8px 40px')], [T('Need help making the editor fit? Our team is here.<br>' + lk('Book a call', '#7B4DFF') + ' or ' + lk('reply to this email', '#7B4DFF') + '.', ink3, 14, HELV, 'left', '20px 40px 8px 8px')]], [1, 8], '#FFFFFF', '0px'),
 row(T('Your partner in building,<br><strong>The Blockwise team</strong>', ink3, 15, HELV, 'left', '24px 40px 32px'), bg='#FFFFFF', pad='0px'),
], foot('#F7F6FB', '#8A8A95', 'The embeddable editor for software teams.', 'Blockwise · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#F7F6FB', ['developer', 'guide', 'customisation'])

# Coffee guide: huge caps headline, brewer icons, then three colour blocks.
coral, sun, sky2 = '#EF6448', '#F9C623', '#A9DCE8'
studio('grind-guide', 'The Grind Guide', 'Commerce', 'A playful product guide in stacked colour blocks: a giant headline, tips, and a two-path choice at the end.',
 'How to pick the right grind size', 'Hint: it depends on how you brew.', [
 row(H('<span style="color:#F2B233">DAY</span><span style="color:#5FB7D4">BREAK</span>', '#111111', 26, BLACK, pad='28px 40px 4px'), H('HOW TO PICK<br>THE RIGHT<br>GRIND SIZE', coral, 46, BLACK, pad='8px 40px 16px'),
  ill('brewers', 'Five coffee brewers: a machine, a drip pot, a stovetop pot, a pour-over and a press', pad='0px 24px'), T('<em>Hint: it depends on how you brew.</em>', '#111111', 14, HELV, pad='12px 40px 28px'), bg='#FFFFFF', pad='0px'),
 row(T('We know a thing or two about coffee, and here’s the short version: grind size matters. Whether you brew strong espresso or a slow cafetière, the size of your grounds can make or break the cup.<br><br>Get it right and your coffee tastes like it should. Get it wrong and it’s thin, bitter or just strange.<br><br>Our head roaster, Teo, wrote the full guide on our blog:', '#3A2A10', 14, TREB, pad='36px 56px 8px'), pill('GET THE FULL GRIND GUIDE', '#FFFFFF', coral, font=TREB, pad='16px 40px 40px', size='13px'), bg=sun, pad='0px'),
 row(H('OUR TOP TIPS', '#2B2B2B', 40, BLACK, pad='40px 40px 0px'), H('For a better cup at home', '#2B2B2B', 17, TREB, pad='4px 40px 16px'),
  T('<strong style="font-size:17px">Choose your brewer, then your grind.</strong><br>The brewer decides the grind. Espresso wants fine. Filter wants medium. A cafetière wants coarse.', '#2B2B2B', 14, TREB, 'left', '12px 56px'), divider('#7FB9C8'),
  T('<strong style="font-size:17px">Grind just before you brew.</strong><br>Ground coffee goes stale in minutes. If you buy it ground, keep it airtight, cool and dark.', '#2B2B2B', 14, TREB, 'left', '12px 56px'), divider('#7FB9C8'),
  T('<strong style="font-size:17px">Taste, then adjust.</strong><br>Watery or sour? Go finer. Harsh or bitter? Go coarser.', '#2B2B2B', 14, TREB, 'left', '12px 56px'),
  pill('READ THE GRIND GUIDE', coral, font=TREB, pad='20px 40px 40px', size='13px'), bg=sky2, pad='0px'),
 row(H('NO GRINDER?', '#FFFFFF', 40, BLACK, pad='40px 40px 0px'), T('<strong>We’ll take it from here.</strong>', '#FFFFFF', 16, TREB, pad='4px 40px 8px'), T('Do you want to buy a grinder?', '#FFFFFF', 16, TREB, pad='8px 40px 16px'), bg=coral, pad='0px'),
 weighted([[T('Yes, let’s<br>do it ↓', '#FFFFFF', 16, TREB, pad='4px 8px 12px 40px'), T('<span style="display:block;background:#FADCD2;border-radius:12px;padding:18px 10px;color:#EF6448"><strong>Here are three we like:</strong></span>', coral, 15, TREB, pad='4px 8px 8px 40px'), T(ul(['Starter: a hand grinder (about $50)', 'Middle: an electric burr grinder (about $150)', 'Splurge: a flat-burr grinder (about $400)']), '#FFFFFF', 13, TREB, 'left', '8px 8px 24px 40px')],
           [T('No, I’m<br>fine ↓', '#FFFFFF', 16, TREB, pad='4px 40px 12px 8px'), T('<span style="display:block;background:#FADCD2;border-radius:12px;padding:18px 10px;color:#EF6448"><strong>We’ll grind it for you.</strong></span>', coral, 15, TREB, pad='4px 40px 8px 8px'), T('Pick a grind size when you add a bag to your basket. We grind to order for every way of brewing.', '#FFFFFF', 13, TREB, pad='8px 40px 24px 8px')]], [1, 1], coral, '0px'),
 row(pill('SHOP DAYBREAK COFFEE', '#FFFFFF', coral, font=TREB, pad='8px 40px 40px', size='13px'), bg=coral, pad='0px'),
], foot('#FFFFFF', '#555555', 'Daybreak Roasters · 12 Example Street, Your City', WHY), '#FFFFFF', ['guide', 'coffee', 'colour blocks'])

# Rent explainer: rounded photo, numbered benefits, a navy panel and a split card.
blue3, navy3 = '#0D47D9', '#0B2A7A'
def numbered(n, title, body): return weighted([[T(f'<span style="display:inline-block;width:22px;height:22px;border:2px solid #1E7D5A;border-radius:50%;color:#1E7D5A;font-size:12px;line-height:22px">{n}</span>', '#1E7D5A', 12, HELV, pad='4px 4px 12px 40px')], [T(f'<strong>{title}</strong><br>{body}', '#1A1A1A', 15, HELV, 'left', '4px 40px 12px 8px')]], [1, 7], '#FFFFFF', '0px')
studio('pay-online', 'Pay Rent Online', 'Product', 'A feature explainer with a rounded photo, numbered benefits, a dark-blue prompt and a split promo card.',
 'Pay your rent online with Homestead', 'No fees by bank transfer, and on-time payments can help your credit.', [
 row(H('⌂ Homestead Rentals', blue3, 24, HELV, 'left', '24px 40px'), P('room', 'wide', '0px 24px'), H('Pay your rent online', '#111111', 34, HELV, 'left', '28px 40px 8px'),
  T('Once you’ve found your place, you can pay rent online in a couple of taps. Paying through Homestead can also help your credit history when you apply for your next home.', '#1A1A1A', 15, HELV, 'left', '4px 40px 8px'),
  button('Learn more', blue3, align='left'), ill('house-dots', 'Three small houses', '90px', '8px 40px'), H('Why pay with Homestead', '#111111', 22, HELV, 'left', '8px 40px 12px'), bg='#FFFFFF', pad='0px'),
 numbered(1, 'Build your credit', 'On-time payments can be reported to help your credit score.'),
 numbered(2, 'No fees', 'Bank transfers are always free.'),
 numbered(3, 'Never late', 'Turn on autopay and it’s handled.'),
 numbered(4, 'Your details are safe', 'We use the same kind of encryption as most banks.'),
 row(ill('icon-heart-house', 'A house with a heart', '48px', '36px 40px 8px', 'left'), H('Still looking for a place?', '#FFFFFF', 20, HELV, 'left', '8px 40px'), T('Use filters and keywords to find your next home faster.', '#FFFFFF', 15, HELV, 'left', '4px 40px'), T('<strong><a href="https://example.com" style="color:#FFFFFF;text-decoration:none">Start searching →</a></strong>', '#FFFFFF', 15, HELV, 'left', '8px 40px 36px'), bg=navy3, pad='24px 0px 0px'),
 weighted([[H('Rent with<br><span style="color:#B8A6FF">peace of mind</span>', '#FFFFFF', 20, HELV, 'left', '32px 8px 8px 32px'), T('Renters insurance through Homestead takes five minutes.', '#FFFFFF', 14, HELV, 'left', '8px 8px 8px 32px'), pill('Get a quote', '#3A0F6E', outline='#FFFFFF', align='left', pad='12px 8px 32px 32px', size='14px', inner='10px 18px')], [P('kitchen', 'square')]], [1, 1], '#3A0F6E', '32px 24px'),
], foot('#FFFFFF', '#555555', 'Homestead Rentals · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Privacy</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">Preferences</a>'), '#FFFFFF', ['explainer', 'numbered', 'rentals'])

# Research report: banner, photo, three stat cards with text, a boxed summary.
beige = '#ECEBDF'
def stat_row(img, alt, title, body, bg='#FFFFFF'):
 return weighted([[ill(img, alt, '100%', '24px 8px 24px 32px')], [H(title, '#1A1A1A', 22, SERIF, 'left', '28px 32px 8px 12px'), T(body, '#444444', 14, HELV, 'left', '4px 32px 8px 12px'), T('<a href="https://example.com" style="border:1px solid #1A1A1A;color:#1A1A1A;padding:10px 16px;border-radius:8px;text-decoration:none">Learn more</a>', '#1A1A1A', 14, HELV, 'left', '16px 32px 24px 12px')]], [2, 3], bg, '0px')
studio('report-is-out', 'The Report Is Out', 'Product', 'A research-report launch: a photo, a serif headline, three big-number cards with commentary and a boxed summary.',
 'Our 2026 report on AI in sales is out', 'We asked 300 sales teams how they really use AI. Here’s what they said.', [
 row(T('<a href="https://example.com" style="color:#FFFFFF">Save your seat</a> at Northstar Live this September →', '#FFFFFF', 13, HELV, pad='12px 32px'), bg='#2E4A2E', pad='0px'),
 row(P('papers', 'wide', '24px 24px 0px'), H('Our 2026 report on AI in sales is here', '#1A1A1A', 30, SERIF, pad='24px 40px 8px'),
  T('In case you missed it: we asked 300 sales and marketing teams how they’re actually using AI. Most said it’s already part of the job. Here’s some of what we found.', '#444444', 14, HELV, pad='4px 48px'), button('Download the report', '#1F1F1F'), divider('#E2E0D6'), bg='#FFFFFF', pad='0px'),
 stat_row('stat-6', 'Card: 9% of teams think the tool could replace someone today', 'Almost nobody thinks it’s taking their job.', 'Most teams use it every week, but only 9% think it could replace a person on their team. They see it as help, not a substitute.'),
 stat_row('stat-97', 'Card: 88% of teams have tried the tool at least once', 'Everyone’s tried it. Few have it working well.', 'Most teams are still experimenting. Only one in five say it’s fully part of how they work.'),
 stat_row('stat-84', 'Card: 71% use it first for research and prep', 'Research comes first, by a long way.', '71% of teams use AI to research accounts before outreach, well ahead of writing emails or cleaning data.'),
 row(T('<span style="display:block;border:1px solid #1A1A1A;border-radius:12px;padding:24px"><span style="font-family:georgia;font-size:22px">There’s more in the full report</span><br><br>Adoption is only half the story. Here’s a preview of what’s slowing teams down next.<br><br><strong>Only 5% fully trust the output.</strong> Most still check it before acting.<br><strong>Most use two to five tools.</strong> Few have a plan for connecting them.<br><strong>Automation tops next year’s list.</strong> It’s the most common priority.</span>', '#1A1A1A', 14, HELV, pad='24px 32px'), button('Download the report', '#1F1F1F'), T('✲ The Northstar team', '#1A1A1A', 15, HELV, 'left', '16px 40px 32px'), bg='#FFFFFF', pad='0px'),
], foot(beige, '#555555', '<a href="https://example.com" style="color:inherit">Manage preferences</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>', 'Northstar · 12 Example Street, Your City'), beige, ['report', 'research', 'statistics'])

# Integration launch: centred label and title, launch tiles, emoji rows, a signature.
leaf, pine = '#A8E6A1', '#24413A'
def emoji_row(img, alt, title, body): return weighted([[ill(img, alt, '56px', '16px 4px 8px 40px')], [T(f'<strong>{title}</strong><br>{body}', '#2B2B2B', 15, HELV, 'left', '16px 40px 12px 8px')]], [1, 6], '#FFFFFB', '0px')
studio('now-connected', 'Now Connected', 'Product', 'An integration launch: a centred title over a two-tile illustration, three emoji-led benefits and a signed sign-off.',
 'Introducing: schedule your Notebook posts', 'Plan and schedule Notebook posts alongside everything else in Queueline.', [
 row(H('≋ Queueline', '#24413A', 22, HELV, pad='28px 40px 16px'), bg='#F4F3EC', pad='0px'),
 row(T('INTRODUCING', '#3F7A4A', 16, HELV, pad='32px 40px 0px'), H('Notebook posts', '#1A1A1A', 32, HELV, pad='4px 40px 16px'), ill('launch-tiles', 'Two app tiles joined by a dotted line on a peach gradient', pad='0px 40px'),
  T('Good news if you write on Notebook: you can now plan and schedule your posts alongside your other channels in Queueline.', '#2B2B2B', 16, HELV, 'left', '20px 40px 8px'), pill('Connect a Notebook account →', leaf, '#1F3A2A', pad='12px 40px 16px'), divider('#E2E2DA'), bg='#FFFFFB', pad='0px'),
 emoji_row('emoji-recycle', 'A recycling sign', 'Turn one post into many', 'Share a Notebook post to your other channels at the same time.'),
 emoji_row('emoji-chart', 'A rising chart', 'See which posts bring in readers', 'Track new subscribers, views and follows for each post.'),
 emoji_row('emoji-shake', 'A handshake', 'Built with the Notebook team', 'Schedule text, images, links and video, with analytics for every post.'),
 row(divider('#E2E2DA'), T('Already on Notebook, or just curious? Now’s a good time to start. Head to your channels to ' + lk('connect and queue your first post', '#3F7A4A') + '.', '#2B2B2B', 16, HELV, 'left', '8px 40px'), pill('Learn more about Notebook →', leaf, '#1F3A2A', pad='16px 40px 16px'), bg='#FFFFFB', pad='0px'),
 weighted([[ill('avatar-4', 'Portrait of Femi', '90px', '8px 8px 32px 40px')], [T('Femi<br><em style="color:#777777;font-size:14px">Product Marketing</em>', '#1A1A1A', 18, HELV, 'left', '30px 40px 32px 8px')]], [1, 3], '#FFFFFB', '0px'),
], foot(pine, '#FFFFFF', 'Sent by the team at Queueline · 12 Example Street, Your City', 'You’re getting this because you signed up for product updates.', '<a href="https://example.com" style="color:inherit">Manage subscriptions</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#F4F3EC', ['launch', 'integration', 'product'])

# Travel promo: big headline, a photo in a thick bright frame, one wide button.
studio('shoulder-season', 'Shoulder Season', 'Commerce', 'A short travel promotion: a big centred headline, a photo in a thick bright frame, two lines and one wide button.',
 'Travel this shoulder season', 'Quieter places, kinder weather and better prices.', [
 row(H('◉◉ Roamwell', '#0B2A1E', 26, HELV, pad='28px 40px 8px'), H('Take a trip this<br>shoulder season', '#0B2A1E', 44, HELV, pad='8px 40px 20px'), bg='#FFFFFF', pad='0px'),
 row(P('valley', 'wide', '20px'), bg='#22E36A', pad='0px'),
 row(T('Fewer crowds, milder weather, better prices: there are plenty of reasons to travel between seasons. We’ve picked some lovely places (and prices) to get you started.', '#0B2A1E', 19, HELV, pad='24px 48px 8px'), T('<a href="https://example.com" style="display:block;background:#0B2A1E;color:#FFFFFF;padding:16px;border-radius:999px;text-decoration:none;font-weight:bold">Browse stays</a>', '#FFFFFF', 16, HELV, pad='16px 40px 40px'), bg='#FFFFFF', pad='0px'),
], row(H('◉◉ Roamwell', '#0B2A1E', 20, HELV, 'left', '24px 40px 8px'), T('Please don’t reply to this email; this address isn’t monitored. For help, visit our ' + lk('help centre', '#0B2A1E') + '.<br><br>Roamwell · 12 Example Street, Your City<br><br>Prices shown are a guide and can change. Room types vary.', '#0B2A1E', 12, HELV, 'left', '8px 40px'), T('<strong>' + lk('Go to Roamwell', '#0B2A1E') + ' · ' + lk('Unsubscribe', '#0B2A1E') + ' · ' + lk('Privacy', '#0B2A1E') + '</strong><br><br><a href="https://xem.email" style="color:#0B2A1E">Built with Xem</a>', '#0B2A1E', 12, HELV, 'left', '8px 40px 32px'), bg='#FFFFFF', pad='0px'),
 '#FFFFFF', ['travel', 'promotion', 'photo'])

# Account notice: soft gradient greeting, bordered boxes for what changes and what to do.
red2, cream3 = '#E8473C', '#FFF8F3'
studio('heads-up', 'A Heads-up', 'Transactional', 'An important account notice: a warm greeting, a bordered box for what’s changing, a checklist box for what to do, and a timeline.',
 'Action needed: a change to how we connect to your CRM', 'Your CRM is tightening its security rules. Here’s what to do before it does.', [
 row(H('✉ Inkwell Studio', '#1A1A1A', 22, HELV, pad='28px 40px'), bg='#F5F5F5', pad='0px'),
 row(H('Hi {{first_name}},', red2, 34, HELV, 'left', '32px 40px 8px'), T('Your CRM provider is introducing stricter security rules this year, and one of them could <strong>affect how Inkwell connects to your account</strong>. We wanted to tell you now so you know what to expect and what to do.', '#1A1A1A', 15, HELV, 'left', '4px 40px 16px'), bg='#FFF0EE', pad='0px'),
 row(T('<span style="display:block;border:1px solid #1A1A1A;border-radius:10px;padding:20px 24px;background:#FFF8F3"><span style="font-size:22px">💬 <strong>What’s changing</strong></span><br><br>Your CRM is making an approved-address list compulsory. Once it’s on, <strong>only requests from approved addresses can sign in or use the API, and that includes Inkwell</strong>. It’s rolling out account by account, and your CRM will email you the date.<br><br>Separately, it will require stronger sign-in (a passkey or security key) for admins. That won’t affect the connection, but it will affect whoever manages your CRM admin.</span>', '#1A1A1A', 14, HELV, 'left', '16px 32px'),
  H('What this means for you', '#1A1A1A', 24, HELV, 'left', '24px 40px 8px'), T('If you use Inkwell’s CRM connection, <strong>add Inkwell to your approved list before your date</strong>. If you don’t, syncing and publishing could stop without warning.', '#1A1A1A', 15, HELV, 'left', '4px 40px'),
  ill('shield-mail', 'A shield with a tick, a document and an envelope', pad='8px 24px'),
  T('<span style="display:block;border:1px solid #1A1A1A;border-radius:10px;padding:20px 24px;background:#FFF8F3"><span style="font-size:22px">☑ <strong>What to do</strong></span><br><br><span style="color:#E8473C">•</span> Ask your CRM admin to <strong>check the current approved list</strong>.<br><span style="color:#E8473C">•</span> <strong>Add Inkwell’s addresses</strong>: 203.0.113.10, 203.0.113.11 and 203.0.113.12.<br><span style="color:#E8473C">•</span> If you can, <strong>turn on monitoring mode first</strong> for a week to check nothing gets blocked.<br><span style="color:#E8473C">•</span> If your admin needs the new sign-in, <strong>set up a passkey early</strong>.</span>', '#1A1A1A', 14, HELV, 'left', '8px 32px'),
  H('Timeline', '#1A1A1A', 24, HELV, 'left', '24px 40px 8px'), T('Your CRM is switching accounts over one at a time and will email you your date. If you haven’t heard yet, look out for it, and don’t wait to make the changes above.', '#1A1A1A', 15, HELV, 'left', '4px 40px 16px'), bg=cream3, pad='0px'),
 weighted([[ill('icon-chat', 'A chat bubble', '56px', '24px 8px 24px 40px')], [T('<strong>Questions about your setup? ' + lk('Open a support request', red2) + ' and we’ll help you sort it.</strong><br><span style="color:#7B4DFF">The Inkwell team</span>', '#1A1A1A', 14, HELV, 'left', '24px 40px 24px 8px')]], [1, 5], '#F3EEFB', '0px'),
], foot('#F5F5F5', '#777777', 'This is an account notice, not a marketing email. We send it to every customer with a CRM connection.', 'Inkwell Studio · 12 Example Street, Your City'), '#F5F5F5', ['notice', 'security', 'account'], marketing=False)

# Long changelog: text-led, with UI shots, a code block and short testimonials.
paper4, ink4 = '#F7F6F2', '#1F1E1D'
CODE = 'font-family:courier new,monospace;font-size:12px;background:#1F1E1D;color:#E8E6DC;display:block;padding:14px 16px;border-radius:6px;white-space:pre'
studio('the-changelog', 'Release Notes', 'Product', 'A long, text-led release email: two headline features with screenshots, a code block, quotes from users and a version list.',
 'Two new features in beta, and everything else that shipped', 'Projects, a desktop workspace, and twenty smaller fixes.', [
 row(H('<span style="color:#D97757">✳</span> Tern', ink4, 22, SERIF, 'left', '28px 40px 8px'), T('Hey!<br><br>Two big features shipped in beta this month. <strong>Projects</strong> let you keep several sessions under one goal. And the desktop app now has a <strong>workspace</strong> where sessions, files and notes sit side by side.', ink4, 14, HELV, 'left', '8px 40px'), bg=paper4, pad='0px'),
 row(H('Projects (beta)', ink4, 22, SERIF, 'left', '20px 40px 4px'), T('Group sessions under one goal so you can hand work over without losing the thread. Each project keeps its own files, instructions and history.', ink4, 14, HELV, 'left', '4px 40px 8px'), ill('ui-dark-sessions', 'A dark app window with a list of sessions and a welcome message', pad='8px 40px'),
  T('<strong>' + lk('Read more about Projects →', ink4) + '</strong>', ink4, 14, HELV, 'left', '8px 40px 16px'), divider('#DDDAD0'), bg=paper4, pad='0px'),
 row(H('Check your scripts before you ship', ink4, 22, SERIF, 'left', '16px 40px 4px'), T('Tern can now run a quick evaluation on your plugins and tell you which cases pass:', ink4, 14, HELV, 'left', '4px 40px 8px'),
  T(f'<span style="{CODE}">$ tern eval ./plugins\n\nCASE            PASS   TIME\n01-setup        yes    0.4s\n02-import       yes    1.2s\n03-edge-cases   no     0.9s\n\n2 of 3 passed</span>', ink4, 12, MONO, 'left', '4px 40px'),
  T('<strong>' + lk('Read more about evaluations →', ink4) + '</strong>', ink4, 14, HELV, 'left', '8px 40px 16px'), divider('#DDDAD0'), bg=paper4, pad='0px'),
 row(H('How people use Tern', ink4, 22, SERIF, 'left', '16px 40px 8px'),
  T('<span style="display:block;border-left:3px solid #D97757;padding-left:14px"><em>“I give it the boring migrations now. It writes the plan, I read it over coffee, and it’s done before lunch.”</em><br><span style="font-size:12px;color:#777777">— Priya, platform engineer</span></span>', ink4, 14, HELV, 'left', '4px 40px 12px'),
  T('<span style="display:block;border-left:3px solid #D97757;padding-left:14px"><em>“The project notes mean whoever picks the work up next knows what was tried and why.”</em><br><span style="font-size:12px;color:#777777">— Sam, engineering lead</span></span>', ink4, 14, HELV, 'left', '4px 40px 16px'), divider('#DDDAD0'), bg=paper4, pad='0px'),
 row(H('Other highlights', ink4, 20, SERIF, 'left', '16px 40px 8px'),
  T('<span style="font-family:courier new;color:#777777">2.4.0</span>&nbsp; <strong>Faster start-up.</strong> Opening a large folder takes about half as long.<br><br><span style="font-family:courier new;color:#777777">2.3.2</span>&nbsp; <strong>Clearer errors</strong> when a connection drops halfway through a task.<br><br><span style="font-family:courier new;color:#777777">2.3.0</span>&nbsp; <strong>Keyboard shortcuts</strong> for switching between sessions.', ink4, 13, HELV, 'left', '4px 40px 16px'),
  T('That’s it for this month. See the ' + lk('full changelog', ink4) + ' for everything else.<br><br>Happy building,<br>The Tern team', ink4, 14, HELV, 'left', '8px 40px 24px'), T('Was this email useful? &nbsp;' + lk('👍', ink4) + ' &nbsp;' + lk('👎', ink4), ink4, 14, HELV, pad='8px 40px 32px'), bg=paper4, pad='0px'),
], row(H('<span style="color:#D97757">✳</span> Tern', '#FFFFFF', 20, SERIF, 'left', '28px 40px 8px'), T('Tern · 12 Example Street, Your City · <a href="https://example.com" style="color:#BBBBBB">Unsubscribe</a><br><a href="https://xem.email" style="color:#BBBBBB">Built with Xem</a>', '#BBBBBB', 12, HELV, 'left', '8px 40px 32px'), bg='#141414', pad='0px'),
 '#EFEDE6', ['changelog', 'release notes', 'developer'])

# Trade newsletter: dark header, alternating dark, white and blue sections.
dark5, orange5, blue5 = '#0E0E0E', '#F05A28', '#1F4FD8'
studio('the-navigator', 'The Navigator', 'Newsletters', 'A trade and policy newsletter in alternating dark, white and blue sections, with a big photo, a quote card and a story list.',
 'The Navigator: new rules for tracing shipments', 'This month’s policy changes, what they mean, and where to start.', [
 row(H('◐ Meridian', '#FFFFFF', 22, HELV, 'left', '24px 40px 4px'), H('The Navigator', '#FFFFFF', 36, HELV, 'left', '4px 40px 8px'), T('Welcome to The Navigator. This month, two new rules make it harder to ship goods without knowing exactly where they came from. Here’s what changed and what it means for your team.', '#D8D8D8', 14, HELV, 'left', '4px 40px 24px'), P('train', 'wide'), bg=dark5, pad='0px'),
 row(T('POLICY WATCH', orange5, 11, HELV, 'left', '28px 40px 4px'), H('Customs wants to see further into your supply chain', '#111111', 24, HELV, 'left', '4px 40px 8px'),
  T('A proposal published this month would require importers to share more detail about suppliers, and earlier. Comments are open until 30 November.<br><br>The detail matters more than the date: it asks for records most teams don’t keep today.', '#333333', 14, HELV, 'left', '4px 40px'), button('See how Meridian can help →', orange5, align='center'), bg='#FFFFFF', pad='0px'),
 row(H('◐ Meridian', '#FFFFFF', 44, HELV, pad='40px 40px 0px'), T('Trade Policy Radar · November', '#CCCCCC', 14, HELV, pad='4px 40px 36px'), bg=dark5, pad='0px'),
 row(T('TRADE POLICY RADAR', orange5, 11, HELV, 'left', '28px 40px 4px'), T('The rules moving this month, and what they mean for keeping goods moving:<br><br>' + ul(['<strong>Origin data:</strong> new fields required on import filings from January.', '<strong>Supplier lists:</strong> a public register opens for comment in December.', '<strong>Spot checks:</strong> more random inspections at the three busiest ports.']), '#333333', 14, HELV, 'left', '4px 40px'), button('See where these touch your products →', orange5, align='center'), bg='#FFFFFF', pad='0px'),
 weighted([[H('◐ Meridian', '#FFFFFF', 16, HELV, 'left', '28px 8px 4px 40px'), H('A new era for proof of origin', '#FFFFFF', 26, HELV, 'left', '8px 8px 4px 40px'), T('Our policy lead on the new playbook', '#D8E2FF', 14, HELV, 'left', '4px 8px 28px 40px')], [P('face3', 'square', '28px 40px 28px 8px')]], [3, 2], blue5, '0px'),
 row(H('Ana Ruiz on the new trade-fraud playbook', '#111111', 20, HELV, 'left', '24px 40px 8px'), T('A new government playbook sets out how agencies plan to work together on trade fraud. Ana walks through the four parts that matter most for importers.', '#333333', 14, HELV, 'left', '4px 40px'),
  T('<span style="display:block;background:#FFF6E8;border-radius:8px;padding:16px 20px"><em>“Proof of origin is moving from a paperwork exercise to a data one. The teams that can show their records quickly will be the ones that keep moving.”</em><br><span style="font-size:12px">— Ana Ruiz, Policy Lead</span></span>', '#333333', 13, SERIF, 'left', '12px 40px'), button('Read Ana’s take →', orange5, align='center'), bg='#FFFFFF', pad='0px'),
 row(H('Recent news and analysis', '#111111', 18, HELV, 'left', '24px 40px 8px'), T('1. <strong>Trade and tech day:</strong> notes from our session with three port operators. ' + lk('Read the recap →', blue5) + '<br><br>2. <strong>Ten minutes with a customs broker</strong> on what’s changed this year. ' + lk('Listen →', blue5) + '<br><br>3. <strong>A supplier map in a day:</strong> how one importer found three hidden sub-suppliers. ' + lk('Read the story →', blue5), '#333333', 14, HELV, 'left', '4px 40px 24px'), bg='#FFFFFF', pad='0px'),
 row(H('Contact Meridian', '#FFFFFF', 18, HELV, 'left', '28px 40px 4px'), T('Want to see what Meridian could do for your team? Our team will walk you through it.', '#D8D8D8', 14, HELV, 'left', '4px 40px'), pill('Get in touch →', '#2E9E5A', align='left', pad='12px 40px 28px', size='14px'), bg=dark5, pad='0px'),
], foot(dark5, '#999999', 'Meridian · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">Manage preferences</a>'), '#E8E8E8', ['trade', 'policy', 'newsletter'])

# Event recap: dark olive page, serif quotes, two-up photos.
olive6, sand6 = '#2A2A1E', '#E9E4D2'
def quote6(q): return H(f'“{q}”', sand6, 26, SERIF, pad='24px 48px')
def two_up(a, b): return weighted([[P(a, 'portrait', '0px 4px 0px 24px')], [P(b, 'portrait', '0px 24px 0px 4px')]], [1, 1], olive6, '8px 0px')
studio('the-gathering', 'The Gathering', 'Events', 'An event recap told through quotes and pairs of photos on a dark page, with a small line drawing at the top.',
 'Inside our flower-arranging afternoon', 'A guided session at a florist’s home studio, in pictures and her own words.', [
 row(T('F E R N &nbsp; &amp; &nbsp; S T E M', sand6, 12, HELV, pad='28px 40px 8px'), ill('botanical', 'A line drawing of a branching plant', '120px', '8px 40px'), H('Last week, we spent an afternoon at florist Lena Moss’s home studio for a guided flower-arranging class.', sand6, 24, SERIF, pad='16px 48px'),
  T('Lena has been a customer since our first year. She runs a small studio, teaches evening classes, and arranges flowers for a handful of local restaurants.', sand6, 13, SERIF, pad='4px 64px 8px'), quote6('I started arranging flowers as a break from a desk job. It was only meant to be for fun. Ten years later, it’s the job.'), bg=olive6, pad='0px'),
 two_up('flowers', 'plants'),
 row(quote6('Arranging isn’t about filling the vase. It’s about knowing which stem to leave out.'), T('“I keep my classes small. I want everyone to leave with something that looks like theirs, not like mine.”', sand6, 13, SERIF, pad='4px 64px 8px'), T('THE CLASS', '#8C8670', 12, HELV, pad='16px 40px'), bg=olive6, pad='0px'),
 two_up('table', 'room'),
 row(T('The afternoon started with a short breathing exercise and a playlist Lena made for the day. Each table had a vase, a pair of snips, some wire and a notebook for sketching.', sand6, 13, SERIF, pad='16px 64px'), quote6('It doesn’t have to be full. It doesn’t have to be symmetrical. It just has to feel right to you.'), bg=olive6, pad='0px'),
 two_up('interior', 'chair'),
 row(quote6('Start with the vase. Then build up, a few stems at a time, and stop sooner than you think.'), T('Afterwards, everyone washed up with our citrus hand wash and finished with the daily hand cream.', sand6, 13, SERIF, pad='4px 64px 24px'),
  T('<span style="display:block;background:#4B5A2A;color:#E9E4D2;padding:14px;border-radius:6px">Free delivery on orders over $85</span>', sand6, 14, HELV, pad='8px 40px 24px'), T(lk('Shop Fern &amp; Stem', sand6) + '<br>' + lk('Read the journal', sand6), sand6, 13, HELV, pad='8px 40px'), bg=olive6, pad='0px'),
], foot(olive6, '#A8A28C', 'hello@example.com', 'Fern &amp; Stem · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), olive6, ['recap', 'story', 'photography'])

# --- Batch three ---
blue7 = '#2A2AE0'
studio('case-study', 'Case Study', 'Newsletters', 'A design-studio newsletter: one featured case study with large images and a pull quote, then a dark “more from” section.',
 'Case study: a football league that wanted a fresh start', 'Six clubs, one shared system, and no fake history.', [
 row(H('The — Letterform', '#111111', 18, HELV, pad='32px 40px'), bg='#FFFFFF', pad='0px'),
 row(T('CASE STUDY', '#111111', 11, HELV, pad='36px 40px 4px'), H('Club badges that don’t pretend to be a hundred years old', '#111111', 28, SERIF, pad='4px 48px 16px'), P('crowd', 'wide', '0px 40px'),
  T('A new six-team league asked a small studio for badges that felt like real sports crests but looked like they belonged on the street, not in a museum. Each club got a monogram that locks into the others, and the same system runs across shirts, tickets and signage.', '#555555', 12, HELV, 'left', '16px 40px 8px'),
  pill('CONTINUE READING', blue7, pad='12px 40px 32px', size='12px', inner='10px 22px'), P('cyclists', 'wide', '0px 40px'),
  H('“A badge isn’t just a logo on a shirt. It becomes part of the whole kit language.”', '#888888', 24, SERIF, pad='24px 56px'), P('hall', 'wide', '0px 40px'),
  T('Client: Northside Sevens<br>Design: Studio Kettle', '#888888', 11, SERIF, pad='12px 40px 8px'), pill('READ THE CASE STUDY', blue7, pad='8px 40px 24px', size='12px', inner='10px 22px'),
  T(lk('GET FEATURED LIKE THIS ↗', blue7), blue7, 11, HELV, pad='8px 40px 40px'), bg='#F5F5F5', pad='0px'),
 row(H('SEE MORE FROM<br>STUDIO KETTLE', '#FFFFFF', 26, HELV, pad='48px 40px 24px'), P('wave', 'wide', '0px 40px'), pill('READ THE CASE STUDY', blue7, pad='20px 40px 24px', size='12px', inner='10px 22px'),
  P('redrock', 'wide', '0px 40px'), pill('READ THE CASE STUDY', blue7, pad='20px 40px 24px', size='12px', inner='10px 22px'), T(lk('GET FEATURED LIKE THIS ↗', '#FFFFFF'), '#FFFFFF', 11, HELV, pad='8px 40px 48px'), bg='#000000', pad='0px'),
 row(H('T—L', '#111111', 120, HELV, pad='24px 40px 0px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#888888', 'The Letterform · 12 Example Street, Your City', '<a href="https://example.com" style="color:#2A2AE0">Unsubscribe</a> · <a href="https://example.com" style="color:#2A2AE0">Manage preferences</a>'), '#FFFFFF', ['case study', 'design', 'newsletter'])

green8, lime8, cream8 = '#1F4D2E', '#E3EA4A', '#F7F5EF'
def fact_header(date): return [weighted([[H('Wild Facts', lime8, 28, BLACK, 'left', '24px 8px 20px 40px')], [T(f'<strong>{date}</strong>', '#FFFFFF', 13, HELV, 'right', '30px 40px 20px 8px')]], [1, 1], green8, '0px'), row(ill('torn-stripe', 'A torn paper stripe in orange and yellow'), bg=green8, pad='0px')]
def keep_exploring(items):
 rows = [H('Keep exploring <span style="color:#1F4D2E">────────────</span>', green8, 19, HELV, 'left', '32px 40px 8px')]
 for tag, colour, title in items:
  rows += [T(f'<span style="color:{colour}">●</span> <strong style="font-size:11px">{tag}</strong>', '#1A1A1A', 11, HELV, 'left', '12px 40px 0px'), H(f'{title} <span style="float:right;color:#1F4D2E">→</span>', '#1A1A1A', 22, HELV, 'left', '4px 40px 8px'), divider('#E8E6DE')]
 return row(*rows, bg='#FFFFFF', pad='0px 0px 16px')
fact_footer = row(ill('fact-mark', 'The Wild Facts mark', '40px', '28px 40px 0px'), H('Wild Facts', lime8, 26, BLACK, pad='4px 40px 16px'), bg=green8, pad='0px')
studio('fact-of-the-day', 'Fact of the Day', 'Newsletters', 'A daily fact newsletter in rounded cards: a tagged story with a photo, a multiple-choice question and a list of further reads.',
 'Parrots can learn to share', 'Today’s fact, a quick question, and three more things to read.', [
 *fact_header('JULY 27, 2026'),
 row(T('<span style="background:#E3EA4A;padding:4px 12px;border-radius:999px;font-size:11px">FACT OF THE DAY</span> &nbsp;<span style="color:#F2A33A">●</span> <span style="font-size:11px">BIRDS</span>', '#1A1A1A', 11, HELV, 'left', '32px 40px 8px'),
  H('Some Parrots Will Share Their Food With a Friend', '#1A1A1A', 28, HELV, 'left', '4px 40px 4px'), T('By Wren Holloway', '#1A1A1A', 13, HELV, 'left', '0px 40px 16px'), P('parrot2', 'wide', '0px 40px'),
  T('In experiments with tokens that can be swapped for food, some parrots chose to pass tokens to a neighbour, even when they got nothing back straight away. Researchers think this kind of sharing helps keep their flocks working together.<br><br>The birds were choosy: they shared more with partners they knew well. Strangers got less, which sounds familiar.', '#333333', 15, HELV, 'left', '20px 40px 8px'),
  T(f'<strong>{lk("What else can parrots do?", green8)}</strong> &nbsp;<span style="display:inline-block;background:#1F4D2E;color:#FFFFFF;border-radius:50%;width:28px;height:28px;text-align:center;line-height:28px">→</span>', '#1A1A1A', 17, HELV, 'left', '12px 40px 32px'), bg='#FFFFFF', pad='16px 0px 0px'),
 row(T('★ <strong style="font-size:12px;letter-spacing:1px">TODAY’S QUESTION</strong>', '#1A1A1A', 12, HELV, 'left', '28px 40px 4px'), H('Which of these can a parrot do?', '#1A1A1A', 22, HELV, 'left', '8px 40px 12px'), bg='#FFFFFF', pad='16px 0px 0px'),
 weighted([[T('<a href="https://example.com" style="display:block;background:#EFEFEB;border:1px solid #DADAD4;border-radius:999px;padding:12px;color:#1A1A1A;text-decoration:none"><strong>COPY A SOUND</strong></a>', '#1A1A1A', 13, HELV, pad='6px 6px 6px 40px'), T('<a href="https://example.com" style="display:block;background:#EFEFEB;border:1px solid #DADAD4;border-radius:999px;padding:12px;color:#1A1A1A;text-decoration:none"><strong>USE A TOOL</strong></a>', '#1A1A1A', 13, HELV, pad='6px 6px 24px 40px')],
           [T('<a href="https://example.com" style="display:block;background:#EFEFEB;border:1px solid #DADAD4;border-radius:999px;padding:12px;color:#1A1A1A;text-decoration:none"><strong>SOLVE A PUZZLE</strong></a>', '#1A1A1A', 13, HELV, pad='6px 40px 6px 6px'), T('<a href="https://example.com" style="display:block;background:#EFEFEB;border:1px solid #DADAD4;border-radius:999px;padding:12px;color:#1A1A1A;text-decoration:none"><strong>ALL OF THE ABOVE</strong></a>', '#1A1A1A', 13, HELV, pad='6px 40px 24px 6px')]], [1, 1], '#FFFFFF', '0px'),
 keep_exploring([('BIRDS', '#F2A33A', 'How Do Birds Find Their Way Home?'), ('INSECTS', '#F2A33A', 'Beetles That Glow in the Dark'), ('ACROSS SPECIES', '#8B3A1E', '6 Animals That Came Back From the Brink')]),
 fact_footer,
], foot('#FFFFFF', '#888888', 'Was this email forwarded to you? <strong>' + lk('Sign up', '#1F4D2E') + '</strong>', 'Wild Facts · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), cream8, ['daily', 'quiz', 'facts'])

studio('guess-the-animal', 'Guess the Animal', 'Newsletters', 'A variation of the daily fact email with a three-hint riddle in place of the multiple-choice question.',
 'Waves can tell a seal where its dinner is', 'Today’s fact, a three-clue riddle, and more to explore.', [
 *fact_header('JULY 17, 2026'),
 row(T('<span style="background:#E3EA4A;padding:4px 12px;border-radius:999px;font-size:11px">FACT OF THE DAY</span> &nbsp;<span style="color:#C9D94A">●</span> <span style="font-size:11px">MAMMALS</span>', '#1A1A1A', 11, HELV, 'left', '32px 40px 8px'),
  H('Seals Can Follow a Fish by the Ripples It Leaves', '#1A1A1A', 28, HELV, 'left', '4px 40px 4px'), T('By Wren Holloway', '#1A1A1A', 13, HELV, 'left', '0px 40px 16px'), P('wave', 'wide', '0px 40px'),
  T('A seal’s whiskers are so sensitive that they can pick up the tiny swirls of water left by a fish that swam past several seconds earlier. In murky water, that’s more useful than eyesight.<br><br>Each whisker has a wavy shape that keeps it steady as the seal swims, so it only feels the trail, not its own movement.', '#333333', 15, HELV, 'left', '20px 40px 8px'),
  T(f'<strong>{lk("5 animals with surprising senses", green8)}</strong> &nbsp;<span style="display:inline-block;background:#1F4D2E;color:#FFFFFF;border-radius:50%;width:28px;height:28px;text-align:center;line-height:28px">→</span>', '#1A1A1A', 17, HELV, 'left', '12px 40px 32px'), bg='#FFFFFF', pad='16px 0px 0px'),
 row(T('? <strong style="font-size:12px;letter-spacing:1px">GUESS THE ANIMAL</strong>', '#1A1A1A', 12, HELV, 'left', '28px 40px 4px'), H('What am I?', '#1A1A1A', 22, HELV, 'left', '8px 40px 8px'),
  T('<span style="display:block;background:#F4F4F1;border:1px solid #E2E2DC;border-radius:12px;padding:16px 20px"><strong style="font-size:11px">HINT 1</strong><br>I can sleep standing up.<br><br><strong style="font-size:11px">HINT 2</strong><br>My teeth never stop growing.<br><br><strong style="font-size:11px">HINT 3</strong><br>I’ve been carrying people for thousands of years.</span>', '#1A1A1A', 14, HELV, 'left', '8px 40px'),
  pill('See the answer', '#111111', align='left', pad='16px 40px 32px'), bg='#FFFFFF', pad='16px 0px 0px'),
 keep_exploring([('MAMMALS', '#C9D94A', 'What Do Seals Eat?'), ('REPTILES', '#1F4D2E', 'Why Do Lizards Wave?'), ('ACROSS SPECIES', '#8B3A1E', '10 Days Out for Animal Lovers')]),
 fact_footer,
], foot('#FFFFFF', '#888888', 'Was this email forwarded to you? <strong>' + lk('Sign up', '#1F4D2E') + '</strong>', 'Wild Facts · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), cream8, ['daily', 'riddle', 'facts'])

pine9 = '#0E3D2B'
def latest_story(title, photo, body, bullets, cta):
 return row(divider('#E6E4DE'), H(title, '#1A1A1A', 20, SERIF, 'left', '16px 32px 12px'), P(photo, 'wide', '0px 32px'), T(body + '<br><br>' + ul(bullets), '#333333', 13, HELV, 'left', '16px 32px 8px'), button(cta, pine9, align='center'), bg='#FFFFFF', pad='0px')
studio('the-latest', 'The Latest', 'Newsletters', 'A company news digest under a big serif masthead: a lead feature, then stories each with a photo, short bullets and a button.',
 'The Latest: our ads library, a launch recap and 25 tips', 'Four stories from the team this month.', [
 row(H('lumenly', '#1A1A1A', 22, HELV, 'left', '24px 32px 0px'), H('The Latest', '#1A1A1A', 56, SERIF, 'left', '0px 32px 16px'), bg='#E9E7E1', pad='0px'),
 row(T('<strong>October 30</strong>', '#1A1A1A', 12, HELV, 'left', '12px 32px 8px'), P('laptops', 'wide', '0px 32px'),
  T('<strong>See which brands are advertising in AI answers</strong><br>Ads are starting to appear in AI chat answers, but it’s hard to see who is running them. Our free library lets you search a brand or a website and see its ads, how often they ran, and what people asked.<br><br>' + ul(['Search by brand or website', 'See how often each ad appeared, and for which questions']), '#333333', 13, HELV, 'left', '16px 32px 8px'), button('Search the library', pine9, align='center'), bg='#FFFFFF', pad='0px'),
 latest_story('Launch recap: why AI search runs on evidence', 'hall', 'At our London launch, our head of research showed how AI tools choose which brands to mention. In one example, almost all the evidence came from other people’s websites, not the brand’s own.', ['Most of the sources came from third parties', 'Each country had its own mix of sources'], 'Watch the keynote'),
 latest_story('25 tips from our own ad tests', 'screen', 'We run our own ads and change one thing at a time. Here’s what helped, what didn’t, and how we decide what to stop. Results are early, so treat them as direction, not proof.', ['Describe the buyer’s situation, not keywords', 'Fund it as a test, not a budget shift'], 'Get the 25 tips'),
 latest_story('Webinar: what the first wave of AI search taught us', 'face2', 'On 10 November, our CEO joins a panel to share what worked, what stalled, and how to adapt it country by country.', ['Map your market before choosing what to track', 'Turn measurement into action, not another dashboard'], 'Save your spot'),
 row(H('Run a strategy that gets you found', '#1A1A1A', 24, SERIF, pad='36px 32px 8px'), T('We’ll help you make a plan, build it, and train your team.', '#333333', 13, HELV, pad='4px 32px'), button('Book a demo', pine9, align='center'), bg='#E9E7E1', pad='0px'),
], foot('#FFFFFF', '#AAAAAA', 'Lumenly · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">Manage preferences</a>'), '#E9E7E1', ['digest', 'company news', 'stories'])

def news_story(label, title, body, byline, photo):
 return weighted([[T(f'<strong>{label}</strong>' if label else '', '#111111', 15, HELV, 'left', '16px 8px 4px 40px'), H(title, '#111111', 20, SERIF, 'left', '4px 8px 4px 40px'), T(body, '#333333', 15, HELV, 'left', '4px 8px 4px 40px'), T(byline, '#555555', 11, HELV, 'left', '4px 8px 4px 40px'), T('<strong>' + lk('Read more', '#111111') + '</strong> →', '#111111', 14, HELV, 'left', '8px 8px 20px 40px')], [P(photo, 'square', '20px 40px 0px 8px')]], [3, 2], '#FFFFFF', '0px')
studio('behind-the-story', 'Behind the Story', 'Newsletters', 'A newsroom campaign email: a dark banner, a serif headline, three investigations with thumbnails and a subscription offer.',
 'The story is what you read. This is the work behind it.', 'Introducing our new series on how our reporting gets done.', [
 row(H('The Harbour Courier', '#111111', 26, SERIF, pad='20px 40px'), bg='#FFFFFF', pad='0px'),
 row(P('papers', 'wide', '16px 40px 0px'), H('The story is the part you read. This is the reporting behind it.', '#111111', 28, SERIF, 'left', '16px 40px 8px'),
  T('Introducing Behind the Story, a series about the evidence, effort and stubbornness behind our journalism, starting with our investigation into how the city’s parking cameras were used.', '#333333', 15, HELV, 'left', '4px 40px'), pill('Learn more', '#111111', pad='12px 40px 24px'), H('Explore the stories', '#111111', 22, HELV, pad='8px 40px'), bg='#FFFFFF', pad='0px'),
 news_story('Exclusive', 'How a camera network meant for traffic was used to follow people', 'Cameras installed to catch speeding drivers were searched thousands of times for reasons that had nothing to do with traffic.', 'By Jo Maddox and Sam Ferris', 'city'),
 news_story('Exclusive', 'Nobody was checking the searches. So we did.', 'Our review of the search logs led to two officials being disciplined, and to new rules on who can see the footage.', 'By Sam Ferris and Jo Maddox', 'street'),
 news_story('', 'The camera supplier promises changes after our report', 'The company says it will add audit logs and limits on searches after our reporting showed how easily the system could be misused.', 'By Jo Maddox, Sam Ferris and Lee Cho', 'building'),
 row(T('<span style="display:block;background:#EFEFEF;padding:24px"><strong style="font-size:20px">Support journalism that changes things</strong><br><br>Subscribers pay for reporting like this. Support our work and get a year of unlimited access for $40.</span>', '#111111', 15, HELV, pad='16px 40px 0px'), pill('Read for $40', '#111111', pad='16px 40px 8px'), T('Billed $40 for the first year, then $140 a year. Cancel any time.', '#333333', 11, HELV, pad='4px 40px 32px'), bg='#FFFFFF', pad='0px'),
], row(T('You’re getting this because you registered with The Harbour Courier or signed up for a newsletter.<br><br>' + lk('Unsubscribe', '#FFFFFF') + ' · ' + lk('Terms', '#FFFFFF') + ' · ' + lk('Privacy', '#FFFFFF'), '#FFFFFF', 12, HELV, 'left', '32px 40px 8px'), divider('#444444'), T('<strong style="font-size:20px;font-family:georgia">HC</strong> &nbsp; The Harbour Courier · 12 Example Street, Your City<br><a href="https://xem.email" style="color:#BBBBBB">Built with Xem</a>', '#BBBBBB', 12, HELV, 'left', '8px 40px 32px'), bg='#111111', pad='0px'),
 '#FFFFFF', ['news', 'campaign', 'journalism'])

red10 = '#E8473C'
def numbered_box(n, title, body): return weighted([[H(f'{n:02}', red10, 24, HELV, 'left', '18px 4px 8px 32px')], [T(f'<span style="display:block;border:1px solid #1A1A1A;border-radius:6px;padding:14px 16px;background:#FFFBF5"><strong style="font-size:16px">{title}</strong><br><br>{body}</span>', '#1A1A1A', 13, HELV, 'left', '8px 32px 8px 4px')]], [1, 6], '#FFFFFF', '0px')
studio('call-for-speakers', 'Call for Speakers', 'Events', 'A community event announcement with a dark illustrated header, a call for speakers in numbered boxes and a personal signature.',
 'Our conference has a date, and we want your talk', 'Dates, a venue, and four things to know about speaking.', [
 row(H('✉ Good Mail Club', '#1A1A1A', 22, HELV, pad='28px 40px'), bg='#F4F4F4', pad='0px'),
 row(ill('megaphone-city', 'A tall building with a red megaphone and two small clouds', pad='16px 0px 0px'), bg='#1A1A1A', pad='0px'),
 row(H('Hey there, email people,', '#1A1A1A', 26, HELV, 'left', '24px 40px 4px'), H('📅 Our conference has a date and a place', '#1A1A1A', 17, HELV, 'left', '4px 40px 12px'),
  T('If you haven’t been: Inbox Camp is a small, practitioner-only conference about email. No sales pitches, just people who send email helping each other out.<br><br><strong>Inbox Camp is on 3–5 May 2027</strong>, and we’re moving to <strong>a new venue in the city centre</strong>.<br><br>' + lk('Here’s the official site', red10) + ' (not much there yet). For now, <strong>you can join the waitlist</strong> to hear first when tickets go on sale.', '#1A1A1A', 14, HELV, 'left', '4px 40px'),
  bg='#FFFFFF', pad='0px'),
 weighted([[pill('Put me on the list', '#1A1A1A', align='left', pad='16px 8px 24px 40px', size='14px')], [ill('hand-ok', 'A hand making an OK sign, with a small heart', '100px', '0px 40px 0px 8px', 'right')]], [2, 1], '#FFFFFF', '0px'),
 weighted([[H('Ever thought about giving a talk?', '#FFFFFF', 24, HELV, 'left', '28px 8px 28px 32px')], [ill('avatar-6', 'Portrait of Rosa', '80px', '24px 32px 24px 8px')]], [3, 1], '#1A1A1A', '24px 0px 0px'),
 row(T('Our call for speakers is open. The ' + lk('submission form', red10) + ' has the details, but here are the four things you need to know:', '#1A1A1A', 14, HELV, 'left', '20px 32px 12px'), bg='#FFFFFF', pad='0px'),
 numbered_box(1, 'We want practical talks from real experience.', 'Fixed a messy workflow? Run an interesting test? Built a system your team actually uses? Tell us. No product pitches, please.'),
 numbered_box(2, 'We pay our speakers and cover the hotel.', 'You get a fee and travel money, and we book your hotel and most meals.'),
 numbered_box(3, 'Never spoken on stage before? That’s fine.', 'We pick a mix of new and experienced speakers, and we’ll help you prepare.'),
 numbered_box(4, 'Got an idea but not sure? Ask me first.', 'Reply to this email and I’ll give you some honest feedback before you submit.'),
 row(pill('Pitch me your talk', '#1A1A1A', pad='20px 40px 16px'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('avatar-5', 'Portrait of Dev', '72px', '16px 4px 24px 40px')], [T('<span style="color:#E8473C;font-size:12px"><strong>I really hope to hear from you,</strong></span><br><span style="font-size:18px">Dev Okonjo</span><br><span style="font-size:12px">Community lead, Good Mail Club</span>', '#1A1A1A', 14, HELV, 'left', '16px 40px 24px 8px')]], [1, 4], '#FFFFFF', '0px'),
], row(H('✉ Good Mail Club', '#FFFFFF', 20, HELV, pad='32px 40px 8px'), T('Good Mail Club · 12 Example Street, Your City · ' + lk('Unsubscribe', '#FFFFFF') + '<br><a href="https://xem.email" style="color:#BBBBBB">Built with Xem</a>', '#BBBBBB', 12, HELV, pad='8px 40px 32px'), bg='#1A1A1A', pad='0px'),
 '#F4F4F4', ['conference', 'call for speakers', 'community'])

def blurb(head, body, link): return T(f'<strong>{head}</strong> {body} {lk(link + " →", "#1A1A1A")}', '#1A1A1A', 12, HELV, 'left', '6px 32px 10px')
studio('weekly-sequence', 'The Weekly Sequence', 'Newsletters', 'A dense culture newsletter: short items with bold leads, an interview card, a featured object and a reader question.',
 'Sequence 52: tiny libraries, giant posters, and a chair', 'Ten things worth knowing this week, plus an interview.', [
 row(T('◎ PARLOUR', '#999999', 11, HELV, pad='24px 32px 0px'), H('SEQUENCE<sup>52</sup>', '#111111', 40, BLACK, pad='0px 32px 8px'), bg='#F2F2F2', pad='0px'),
 weighted([[P('face1', 'square', '0px 3px 0px 120px')], [P('art', 'square', '0px 3px')], [P('face4', 'square', '0px 3px')], [P('flowers', 'square', '0px 120px 0px 3px')]], [1, 1, 1, 1], '#F2F2F2', '0px 0px 20px'),
 row(T('<strong>Internet Microdoses</strong>', '#1A1A1A', 12, HELV, 'left', '24px 32px 8px'),
  blurb('A three-hour concert film is selling out cinemas.', 'It made more on its opening weekend than any music documentary in twenty years. Bring snacks.', 'Get tickets'),
  P('room', 'wide', '4px 32px'), blurb('A library the size of a phone box opened on a high street.', 'It holds 200 books and lends them on trust. So far, nothing has gone missing.', 'Visit it'),
  blurb('A boat that sank 4,000 years ago turned up in a field.', 'Archaeologists found it while checking ground for a new road.', 'Dig in'),
  blurb('A football club named its new font after a former captain.', 'It’s on the shirts, the tickets and the stadium signs.', 'See the font'),
  blurb('Students are holding drawing classes in train stations.', 'Bring a pencil; stools are provided.', 'Find a class'), bg='#FFFFFF', pad='0px'),
 row(P('books', 'wide', '16px 32px 8px'), T('<strong>Aesthetic Study · ' + lk('Turning the page', '#1A1A1A') + '</strong><br><br>Last year it was fashionable to be seen reading on the train. This year, the book itself is the accessory: tote bags with bookshelves printed on them, cafés with reading-only tables, book-club dinners that sell out in minutes.<br><br>Whatever the reason, more people are carrying books. Some are even reading them.', '#1A1A1A', 12, HELV, 'left', '8px 32px'),
  pill('Judge a book by its cover', '#C8DBFA', '#1A1A1A', align='left', pad='12px 32px 24px', size='12px', inner='10px 18px'), bg='#FFFFFF', pad='0px'),
 weighted([[T('A moment with', '#F4E6E6', 12, HELV, 'left', '24px 8px 4px 32px'), H('Ines Calder', '#FFFFFF', 30, BLACK, 'left', '4px 8px 4px 32px'), T('Founder of Calder &amp; Co.', '#F4E6E6', 11, HELV, 'left', '4px 8px 24px 32px')], [P('face3', 'square', '20px 32px 20px 8px')]], [3, 2], '#5A1414', '0px 32px'),
 row(T('Ines Calder makes furniture from salvaged wood and wants every piece to look like it has a past. Her third collection launches this month.<br><br><strong>Where do your ideas come from?</strong><br>Old buildings, mostly. I keep a photo of every door I like.<br><br><strong>Three places that inspire you?</strong><br>1. A timber yard by the river<br>2. A library reading room<br>3. My grandmother’s kitchen<br><br><strong>Write us a fortune cookie.</strong><br>The thing you keep fixing is the thing you love.', '#1A1A1A', 12, HELV, 'left', '16px 32px'),
  pill('Read the full interview', '#C8DBFA', '#1A1A1A', align='left', pad='8px 32px 24px', size='12px', inner='10px 18px'), bg='#FFFFFF', pad='0px'),
 row(P('chair', 'wide', '8px 32px'), T('<strong>Object Love · ' + lk('The bent chair', '#1A1A1A') + '</strong><br><br>A chair bent from a single sheet of plywood, first made in the 1940s and still in production. It weighs about as much as a cat. Look closely and there isn’t a screw in it.', '#1A1A1A', 12, HELV, 'left', '8px 32px'),
  T('<strong>Quick Hits</strong><br><br>Put a photo on your keychain with a ' + lk('tiny camera', '#1A1A1A') + '.<br>A museum is ' + lk('lit only by candles', '#1A1A1A') + ' for one night.<br>Discover a ' + lk('dictionary of words you can’t read', '#1A1A1A') + '.', '#1A1A1A', 12, HELV, 'left', '16px 32px'),
  T('<strong>Reader Question</strong><br><br>Last week we asked: if you opened a tiny museum, what would be in it? Our favourite answer: <em>“Every key I’ve ever kept without knowing what it opens.”</em><br><br>This week: what’s the most useful thing you own that cost less than $5? Reply and tell us.<br><br>' + lk('All booked up', '#1A1A1A') + ',<br>Tom', '#1A1A1A', 12, HELV, 'left', '8px 32px 32px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#AAAAAA', 'Sent to you as a Parlour member · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#FFFFFF', ['culture', 'digest', 'interview'])

ink11, paper11 = '#3A2A20', '#F1ECE3'
def brief_item(sec, title, body, link): return [T(sec, '#7A6A5A', 10, HELV, 'left', '20px 32px 4px'), H(title, ink11, 19, SERIF, 'left', '4px 32px 4px'), T(body, '#5A4A3A', 13, SERIF, 'left', '4px 32px 4px'), T(f'<span style="border-bottom:1px solid #3A2A20;letter-spacing:2px;font-size:10px">{link} →</span>', ink11, 10, HELV, 'left', '8px 32px 16px')]
def city(letter, colour, place, title, body): return weighted([[T(f'<span style="display:block;background:{colour};width:44px;height:44px;line-height:44px;text-align:center;font-family:georgia;font-size:22px">{letter}</span>', ink11, 22, SERIF, pad='20px 4px 8px 32px')], [T(place, '#7A6A5A', 10, HELV, 'left', '18px 32px 2px 8px'), H(title, ink11, 18, SERIF, 'left', '2px 32px 4px 8px'), T(body, '#5A4A3A', 13, SERIF, 'left', '4px 32px 4px 8px'), T('<span style="border-bottom:1px solid #3A2A20;letter-spacing:2px;font-size:10px">DETAILS →</span>', ink11, 10, HELV, 'left', '6px 32px 16px 8px')]], [1, 6], paper11, '0px')
studio('the-brief', 'The Brief', 'Newsletters', 'A quiet arts and design digest on warm paper: a numbered contents line, a boxed lead, short items and a city-by-city list.',
 'The Brief: a new museum, a public restoration, and five cities', 'Friday’s short read on art, design and what’s on.', [
 row(H('The <em>Brief</em>', ink11, 44, SERIF, pad='36px 32px 4px'), T('FRIDAY, 30 OCTOBER 2026', '#7A6A5A', 10, HELV, pad='0px 32px 16px'), divider(ink11), T('01 THE LEAD &nbsp;&nbsp; 02 ART &nbsp;&nbsp; 03 DESIGN &nbsp;&nbsp; 04 IN THE CITIES', '#7A6A5A', 10, HELV, pad='4px 32px 16px'), T('01 &nbsp; THE LEAD', ink11, 10, HELV, 'left', '8px 32px 8px'), bg=paper11, pad='0px'),
 row(H('A furniture museum takes shape in an old warehouse', ink11, 24, SERIF, 'left', '24px 24px 8px'), T('New drawings show a 1950s shoe warehouse turned into a museum around one long, top-lit hall.', '#5A4A3A', 13, SERIF, 'left', '4px 24px 8px'), T('<span style="border-bottom:1px solid #3A2A20;letter-spacing:2px;font-size:10px">READ MORE →</span>', ink11, 10, HELV, 'left', '8px 24px 24px'), bg='#FFFFFF', pad='0px 32px'),
 row(*brief_item('02 &nbsp; ART', 'A museum will restore a group portrait in public', 'From December, conservators will clean a 400-year-old portrait in an open gallery, for about a year.', 'READ MORE'),
  *brief_item('', 'A festival commissions a year of hand-made art', 'Seven exhibitions at seven universities, with more than a hundred artists taking part.', 'READ MORE'), divider('#D8CFC2'),
  *brief_item('03 &nbsp; DESIGN', 'A new stadium with a garden on its roof', 'The 60,000-seat ground will have a planted roof and an open circle in the middle.', 'READ MORE'), divider('#D8CFC2'), T('04 &nbsp; IN THE CITIES', ink11, 10, HELV, 'left', '16px 32px 4px'), bg=paper11, pad='0px'),
 city('N', '#E8B4A8', 'NEW YORK', 'A painter’s night skies come to town', 'Her 1990s paintings of stars are shown here for the first time, until 19 December.'),
 city('L', '#B8D4BC', 'LONDON', 'A complete set of mountain prints', 'All forty-six prints in one room, until 17 January.'),
 city('P', '#B8CCE0', 'PARIS', 'Torn posters as art', 'An artist who peeled posters off city walls, until 14 February.'),
 city('T', '#D4C4E4', 'TOKYO', 'Three comic artists who changed the form', 'Original drawings from three pioneers, until 8 February.'),
 city('M', '#F2D79C', 'MEXICO CITY', 'An old manuscript back home on loan', 'Free to see at the national museum until 6 December.'),
 row(divider('#D8CFC2'), T('<em>Chosen this morning in the studio.<br>Reply any time; every note is read.</em>', '#7A6A5A', 13, SERIF, pad='16px 32px'), divider('#D8CFC2'), H('<em>thyme</em>', '#B08A3A', 30, SERIF, pad='16px 32px 0px'), bg=paper11, pad='0px'),
], foot(paper11, '#9A8A7A', 'YOU SIGNED UP FOR THE BRIEF AT THYME.EXAMPLE', '12 EXAMPLE STREET, YOUR CITY', '<a href="https://example.com" style="color:inherit">UNSUBSCRIBE</a>'), '#E6DED2', ['arts', 'digest', 'editorial'])

orange12 = '#D9620F'
def btip(word, title, body, link): return [H(f'<span style="color:#D9620F">b</span>{word} <span style="color:#D9620F">|</span> {title}', '#5A5A5A', 24, SERIF, 'left', '20px 32px 4px'), T(f'{body} <strong>{lk(link + " &gt;", orange12)}</strong>', '#5A5A5A', 15, HELV, 'left', '4px 32px 12px')]
studio('five-tips', 'Five Quick Tips', 'Newsletters', 'A care-tips email: a single warm photo, then five short sections each with a playful one-word lead and a link.',
 'Five things to know before the fireworks', 'Keeping pets calm, safe and well this season.', [
 row(H('✚', orange12, 40, BLACK, pad='16px 32px'), bg='#F4F4F4', pad='0px'),
 row(P('beagle', 'wide'), bg='#FFFFFF', pad='0px'),
 row(*btip('safe', 'fireworks night is scary for pets', 'Bangs, flashes and visitors at the door can frighten dogs and cats. Here’s how to', 'keep your pet calm'),
  *btip('fed', 'the truth about tummy troubles', 'Upset stomachs are one of the most common reasons pets visit the vet. Learn how to', 'help your pet’s digestion'),
  *btip('ready', 'what to know before you adopt', 'A pet is a long-term friend, not a short-term project. Here are some', 'tips before you adopt'),
  *btip('calm', 'helping an anxious pet', 'Holidays and house moves can unsettle pets. Here’s how to', 'help them settle'),
  *btip('well', 'managing pain as they get older', 'Stiff joints don’t have to mean a smaller life. Learn about', 'easing your pet’s pain'), bg='#FFFFFF', pad='16px 0px 24px'),
], foot('#F2F2F2', '#666666', 'Harbourside Vets · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Terms</a> · <a href="https://example.com" style="color:inherit">Privacy</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#FFFFFF', ['tips', 'pets', 'care'])

blue13, peach13 = '#2A2AE0', '#FDF0E4'
def promo_block(bg, title, body, button_text, btn_bg, btn_ink, img, title_colour, text_colour):
 return weighted([[P(img, 'square', '24px 8px 8px 24px')], [H(title, title_colour, 26, HELV, 'left', '24px 24px 4px 8px'), T(body, text_colour, 13, HELV, 'left', '4px 24px 8px 8px')]], [2, 3], bg, '0px'), row(button(button_text, btn_bg, align='center', ink=btn_ink), bg=bg, pad='0px 0px 16px')
studio('animal-buzz', 'Animal Buzz', 'Newsletters', 'A charity supporter update in bold colour blocks: an adoption story, gift and shop promos, a wildlife feature and a fact.',
 'Adopt, don’t shop, and other good news', 'A rehoming story, a magazine for kids, and why seals are back.', [
 row(T('Your update is here', '#333333', 10, HELV, 'left', '6px 24px'), bg='#FFFFFF', pad='0px'),
 row(H('PAWS<span style="color:#F2D33A">&amp;</span>CO BUZZ', '#FFFFFF', 26, BLACK, 'left', '16px 24px'), bg=blue13, pad='0px'),
 row(P('cat'), H('Adopt, don’t shop', blue13, 32, HELV, 'left', '20px 24px 8px'), T('When you adopt from us, our team helps find the right match, and every animal comes with a health check and ongoing support. It gives them a second chance at a loving home.', '#1A1A1A', 14, HELV, 'left', '4px 24px'), button('Find out how', blue13, align='center'), bg=peach13, pad='0px'),
 *promo_block('#EE5A32', 'Share the love!', 'Give a young animal lover a year of our kids’ magazine: rescues, cute photos and fun facts, four issues for $12.', 'Subscribe today', '#FFFFFF', '#1A1A1A', 'collie', '#FFFFFF', '#FFFFFF'),
 *promo_block('#A8DBF5', 'Dress to impress', 'Our clothing range is made from organic cotton and printed to order, so nothing goes to waste.', 'Explore now', blue13, '#FFFFFF', 'fabric', blue13, '#1A1A1A'),
 row(P('wave', 'wide', '24px 24px 0px'), H('We love seals!', '#1A1A1A', 28, HELV, 'left', '16px 24px 8px'), T('Grey seals are coming back to the beaches to have their pups. Find out how they live, how to watch them safely, and how to tell the different kinds apart.', '#1A1A1A', 14, HELV, 'left', '4px 24px'), button('Learn about seals', '#2A1A6A', align='center'), bg='#E8E64A', pad='0px'),
 row(ill('bulb', 'A lightbulb', '64px', '28px 24px 0px', 'left'), H('Cows have amazing noses', '#E8E64A', 20, HELV, 'left', '8px 24px 8px'), T('A cow can smell things several miles away, and every cow’s nose print is unique, like a fingerprint.', '#FFFFFF', 14, HELV, 'left', '4px 24px'), button('Find out about cows', '#E8E64A', align='center', ink='#1A1A1A'), bg=blue13, pad='0px 0px 16px'),
 row(ill('critters', 'Silhouettes of a cat, dog, rabbit, deer, bird, fox and hen'), bg=peach13, pad='16px 0px 0px'),
], foot(blue13, '#FFFFFF', '<strong style="font-size:22px">PAWS&amp;CO</strong>', 'Paws &amp; Co is a registered charity · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe from supporter updates</a>'), '#F5F5F5', ['charity', 'colour blocks', 'animals'])

navy14, orange14, aqua14 = '#1E1A4A', '#EC5A24', '#B4E2E2'
studio('get-out-there', 'Get Out There', 'Welcome', 'A destination welcome email: a photo with a huge stamped headline, a short welcome list, a second photo banner and a map prompt.',
 'Welcome to the high desert', 'Trip ideas, local events and road-trip guides are on their way.', [
 row(T(lk('View in browser', orange14), orange14, 11, HELV, 'right', '8px 24px'), bg=navy14, pad='0px'),
 row(H('✦ VISIT<br>HIGH DESERT ✦', navy14, 24, BLACK, pad='24px 40px'), bg=aqua14, pad='0px'),
 row(P('redrock', 'wide'), H('GET A LITTLE<br>LOST OUT HERE.', orange14, 46, BLACK, 'left', '8px 40px 8px'), H('Welcome to the High Desert.', orange14, 28, HELV, 'left', '16px 40px 8px'),
  T('Psst… you there. Fancy a wander?<br><br>The best trip ideas for the region are coming soon to your inbox:<br><br>☐ Tips from people who live here<br>☐ Local events and places you won’t find anywhere else<br>☐ Road-trip guides and two-day plans<br><br>Plus a whole lot more.', '#1A1A1A', 15, HELV, 'left', '4px 40px'),
  T('<a href="https://example.com" style="background:#EC5A24;color:#FFFFFF;padding:12px 22px;border-radius:4px;text-decoration:none;font-weight:bold">GET FREE TRAVEL GUIDES</a>', '#FFFFFF', 15, HELV, pad='16px 40px 36px'), bg='#FFFFFF', pad='0px'),
 row(P('valley', 'wide'), H('GO WHERE? FRESH AIR.<br>OUT THERE.', '#FFFFFF', 30, BLACK, 'left', '16px 40px 4px'), T('<span style="background:#EC5A24;padding:8px 16px;color:#FFFFFF;font-weight:bold">GET TRIP TIPS ▶</span>', '#FFFFFF', 13, HELV, 'left', '8px 40px 28px'), bg=navy14, pad='16px 0px 0px'),
 weighted([[ill('map-outline', 'A folded map', '120px', '36px 8px 36px 40px')], [H('READY TO START PLANNING?', navy14, 30, BLACK, 'left', '36px 40px 8px 8px'), T('<strong>GET YOUR FREE GUIDE →</strong>', navy14, 16, HELV, 'left', '4px 40px 36px 8px')]], [1, 2], aqua14, '0px'),
], foot(aqua14, navy14, 'Thanks for signing up for emails from Visit High Desert.', 'Got this from a friend? <strong>' + lk('Subscribe here', navy14) + '</strong>', 'Visit High Desert · 12 Example Street, Your City · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), navy14, ['welcome', 'travel', 'destination'])

# --- Batch four ---
studio('september-notes', 'September Notes', 'Product', 'A minimal product update: short bold-titled sections, each followed by a large rounded product image, and a bullet list of smaller changes.',
 'What’s new in Shapekit this September', 'Skills for the assistant, 3D transforms and list fields.', [
 row(T('<span style="display:inline-block;background:#111111;color:#FFFFFF;border-radius:10px;width:40px;height:40px;line-height:40px;text-align:center;font-size:20px">◢</span>', '#111111', 20, HELV, 'left', '40px 40px 16px'),
  T('This September we added skills for the assistant, editable 3D transforms, and list fields for repeating content. You’ll also find pinned projects, visitor stats for your templates, and twelve new lessons.', '#111111', 17, HELV, 'left', '8px 40px 24px'),
  T('<strong>Skills</strong><br>Save instructions for the assistant once and reuse them, from your brand voice to how you lay out a page. ' + lk('Explore skills.', '#111111'), '#111111', 14, HELV, 'left', '16px 40px 16px'), ill('prompt-card', 'A dark prompt box over a bright gradient, asking the assistant to make a skill', pad='0px 40px'),
  T('<strong>3D transforms</strong><br>Ask the assistant to add depth and perspective, then fine-tune the result yourself. Everything stays editable. ' + lk('Explore 3D transforms.', '#111111'), '#111111', 14, HELV, 'left', '36px 40px 16px'), ill('wire-shape', 'A wireframe 3D shape on black', pad='0px 40px'),
  T('<strong>List fields</strong><br>Keep FAQs, features or steps in one list instead of separate numbered fields, then design one card and repeat it. ' + lk('Explore lists.', '#111111'), '#111111', 14, HELV, 'left', '36px 40px 16px'), ill('dark-table', 'A dark table of questions and answers', pad='0px 40px'),
  T('<strong>Other improvements</strong><br>' + ul([lk('Pin your most-used projects', '#111111') + ' to the sidebar.', 'See visitors and views with ' + lk('template stats', '#111111') + '.', 'Try twelve new ' + lk('lessons', '#111111') + ' on working with the assistant.']), '#111111', 14, HELV, 'left', '36px 40px 16px'),
  T('Tell us what you think by replying to this email.<br><br>' + lk('Jonas', '#111111'), '#111111', 14, HELV, 'left', '16px 40px 32px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#AAAAAA', 'You opted in to product updates from Shapekit.', 'Shapekit · 12 Example Street, Your City · <strong><a href="https://example.com" style="color:inherit">Unsubscribe</a></strong>'), '#FFFFFF', ['product update', 'minimal', 'features'])

def card_row(photo, title, body, link, left=True):
 txt = [H(title, '#111111', 18, HELV, 'left', '28px 40px 4px 16px' if left else '28px 16px 4px 40px'), T(body, '#333333', 13, HELV, 'left', '4px 40px 4px 16px' if left else '4px 16px 4px 40px'), T(f'<strong>{lk(link + " →", "#0D47D9")}</strong>', '#0D47D9', 13, HELV, 'left', '8px 40px 24px 16px' if left else '8px 16px 24px 40px')]
 pic = [P(photo, 'square', '0px 0px 0px 40px' if left else '0px 40px 0px 0px')]
 return weighted([pic, txt] if left else [txt, pic], [1, 1], '#FFFFFF', '8px 0px')
studio('front-door', 'Front Door', 'Retention', 'A home and property digest: a rounded hero photo, a headline and button, then cards that alternate photo and text sides.',
 'Front door colours that sell', 'Plus the latest market outlook and a house worth daydreaming about.', [
 row(H('⌂ Homestead', '#0D47D9', 22, HELV, 'left', '24px 40px'), P('house', 'wide', '0px 24px'), H('Front door colours that pay off', '#111111', 30, HELV, 'left', '24px 40px 8px'),
  T('Ever walked past a front door and wanted to live behind it? Buyers feel the same, and some will pay more for it. See which colours help a house sell.', '#333333', 14, HELV, 'left', '4px 40px'), button('Get the list', '#0D47D9', align='left'), ill('house-dots', 'Three small houses', '80px', '8px 40px'), H('What’s trending', '#111111', 22, HELV, 'left', '8px 40px'), bg='#FFFFFF', pad='0px'),
 card_row('street', 'Market update: room to breathe?', 'The latest outlook on prices, rents and rates could make your next move more predictable.', 'Read the report'),
 card_row('laptops', 'When tastes collide', 'How to blend two people’s styles into one shared home.', 'Work it out', left=False),
 row(ill('house-dots', 'Three small houses', '80px', '16px 40px'), H('This week’s favourite', '#111111', 22, HELV, 'left', '8px 40px'), bg='#FFFFFF', pad='0px'),
 card_row('sea', 'Seaside retreat', 'Floor-to-ceiling windows and nothing but ocean in front of you.', 'Look inside'),
 row(H('Your next place is waiting.', '#111111', 22, HELV, 'left', '24px 40px 8px'), T('<a href="https://example.com" style="display:inline-block;border:1px solid #0D47D9;color:#0D47D9;padding:10px 28px;border-radius:6px;text-decoration:none;font-weight:bold">Browse homes for sale</a>&nbsp;&nbsp;<a href="https://example.com" style="display:inline-block;border:1px solid #0D47D9;color:#0D47D9;padding:10px 28px;border-radius:6px;text-decoration:none;font-weight:bold">Browse rentals</a>', '#0D47D9', 13, HELV, pad='8px 40px 24px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#555555', 'Homestead · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Privacy</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">Preferences</a>'), '#FFFFFF', ['digest', 'property', 'cards'])

studio('founder-column', 'From the Founder', 'Newsletters', 'A founder’s weekly column: a stacked logo, a big serif masthead with a red dot, one underlined headline and a long, personal story.',
 'My five minutes on late-night TV', 'A comedy show used an old interview of mine. Here’s what I learned.', [
 row(H('SUB<br>SCR<br>iBED', '#111111', 26, MONO, pad='40px 40px 0px'), T('from Tally &amp; Co', '#111111', 10, HELV, pad='0px 40px 16px'), T('We write about modern business.', '#111111', 13, HELV, pad='8px 40px 8px'),
  H('Subscribed<br><span style="color:#FF4A2A">●</span>Weekly', '#111111', 60, SERIF, pad='8px 40px 16px'), divider('#111111'), H('<u>My Late-Night Moment</u>', '#111111', 34, SERIF, pad='8px 40px 16px'),
  T('A late-night comedy show used a clip of me this week. Be careful what you wish for!<br><br>The segment was about subscriptions, and they dug up an old interview where I talked about fridges that reorder milk by themselves. The host called it “mildly terrifying.” Fair.<br><br>Honestly? I’m a fan, so being teased on TV felt like an honour. It was a good rant: printers that stop working when a trial ends, cancel buttons hidden three menus deep, monthly fees for heated seats.<br><br>Jokes aside, it was strange to see an idea this community spent years working on turn up in a comedy monologue. We went from a small billing company to a topic people argue about at dinner. That doesn’t happen often.', '#222222', 17, HELV, 'left', '8px 56px'),
  pill('Keep reading', '#111111', pad='24px 40px 24px', size='14px'), divider('#111111'), H('⧉ TALLY &amp; CO', '#111111', 24, BLACK, pad='32px 40px 8px'), bg='#EEF2F6', pad='0px'),
], foot('#EEF2F6', '#333333', 'Tally &amp; Co · 12 Example Street, Your City', 'Update your preferences in the <a href="https://example.com" style="color:#FF4A2A">subscriber portal</a>, or <a href="https://example.com" style="color:#FF4A2A">unsubscribe</a>.'), '#FFFFFF', ['column', 'founder', 'letter'])

def honoree(name, body, link, photo, left, bg):
 txt = [H(name, '#111111', 30, SERIF, 'left', '24px 24px 8px'), T(body, '#333333', 12, HELV, 'left', '4px 24px'), T(f'<strong><u>{link}</u></strong>', '#111111', 12, HELV, 'left', '12px 24px 24px')]
 pic = [P(photo, 'square', '20px 20px')]
 return weighted([txt, pic] if left else [pic, txt], [1, 1], bg, '8px 0px')
studio('the-curious-list', 'The Curious List', 'Newsletters', 'An honours-list announcement: a colourful mosaic header, an intro and button, then honoree profiles with photos on alternating sides.',
 'Introducing the 2026 Curious List', 'One hundred people using curiosity to fix real problems. Meet a few.', [
 row(H('◐ Fieldwork<br>Institute', '#FFFFFF', 14, HELV, 'left', '20px 32px 8px'), H('Introducing the<br>2026 Curious List', '#111111', 36, SERIF, 'left', '8px 32px 20px'), bg='#4A9AC8', pad='0px'),
 row(ill('mosaic', 'A mosaic of coloured squares with flower and starburst shapes'), ill('big-year-digits', 'The year 2026 in big digits on coloured squares'), bg='#FFFFFF', pad='0px'),
 row(T('Introducing the 2026 Curious List: a hundred designers, scientists, teachers and makers using curiosity to take on hard problems.', '#333333', 13, HELV, 'left', '24px 32px 8px'), T('<a href="https://example.com" style="display:inline-block;border:1px solid #111111;padding:10px 18px;color:#111111;text-decoration:none">See the list</a>', '#111111', 13, HELV, 'left', '8px 32px 28px'), bg='#FFFFFF', pad='0px'),
 row(H('Meet the honourees', '#111111', 28, SERIF, 'left', '28px 32px 8px'), T('This year’s list covers a wide range of people. Meet a few below, or ' + lk('see the full list', '#111111') + '.', '#333333', 13, HELV, 'left', '4px 32px 24px'), bg='#DCEFEF', pad='0px'),
 honoree('Amara Diallo', 'A marine biologist who trains local divers to map coral reefs, so the people who live beside them can track their recovery.', 'Dive deeper', 'face1', True, '#FFFFFF'),
 honoree('Leon Varga', 'A sculptor who works with salvaged steel from closed factories, turning old machinery into public seating.', 'Learn more', 'face2', False, '#FFFBEE'),
 honoree('Ruth Okafor', 'A product designer who spent twenty years making hearing aids smaller, cheaper and easier to fit.', 'Read more', 'face4', True, '#FFFFFF'),
 row(H('Go beyond the list', '#111111', 18, HELV, 'left', '28px 32px 12px'), ill('mosaic', 'A mosaic of coloured squares', pad='0px 32px'), T('Follow more Curious List stories on our social channels.', '#333333', 12, HELV, 'left', '12px 32px 4px'), T('<strong><u>Let’s go</u></strong>', '#111111', 12, HELV, 'left', '8px 32px 28px'), bg='#FFFBEE', pad='16px 0px'),
], row(T('<span style="color:#E8473C;font-size:30px">♥</span> &nbsp; © 2026 Fieldwork Institute', '#FFFFFF', 12, HELV, 'left', '24px 32px 8px'), T('Fieldwork Institute · 12 Example Street, Your City · ' + lk('Unsubscribe', '#BBBBBB') + '<br><a href="https://xem.email" style="color:#BBBBBB">Built with Xem</a>', '#BBBBBB', 12, HELV, 'left', '4px 32px 24px'), bg='#1F1F1A', pad='0px'),
 '#F2F2F2', ['announcement', 'profiles', 'list'])

studio('subject-lines', 'Subject Lines', 'Product', 'A sales-tips digest on warm grey: a compose-window illustration, a guide offer, two side-by-side stories and a discount panel.',
 '50 sales subject lines that get opened', 'Copy them, plus the patterns behind them.', [
 row(T('<a href="https://example.com" style="color:#FFFFFF">Save your seat</a> at Northstar Live on 30 September →', '#FFFFFF', 12, HELV, pad='10px 24px'), bg='#2E4A2E', pad='0px'),
 row(H('✲ Northstar', '#1A1A1A', 22, HELV, pad='28px 40px 8px'), H('50 sales subject lines that actually get opened', '#1A1A1A', 26, HELV, pad='8px 48px 16px'), ill('compose-window', 'A new-message window with a subject line typed in', pad='0px 24px'),
  T('Most people decide whether to open an email from the subject line alone. Take our 50 best, plus the patterns behind them, and get more replies and meetings.', '#444444', 13, HELV, pad='16px 48px 8px'), button('Get the subject lines', '#1F1F1F'), divider('#E2E0D6'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('stat-97', 'A gradient card', '100%', '20px 8px 20px 40px')], [H('How Quillpost reached 99% of its badge scans', '#1A1A1A', 20, HELV, 'left', '28px 40px 8px 12px'), T('Quillpost turns event badge scans into clean contact records in about three seconds, so reps can follow up while the conversation is still warm.', '#444444', 13, HELV, 'left', '4px 40px 8px 12px'), button('Read their story', '#1F1F1F', align='left')]], [2, 3], '#FFFFFF', '0px'),
 row(divider('#E2E0D6'), bg='#FFFFFF', pad='0px'),
 weighted([[H('Two assistants, head to head', '#1A1A1A', 20, HELV, 'left', '28px 12px 8px 40px'), T('Our team gave two AI assistants the same research, enrichment and follow-up tasks. See which came out on top.', '#444444', 13, HELV, 'left', '4px 12px 8px 40px'), button('Watch now', '#1F1F1F', align='left')], [ill('stat-6', 'A gradient card', '100%', '20px 40px 20px 8px')]], [3, 2], '#FFFFFF', '0px'),
 row(T('<span style="display:block;background:#E9E7DC;border-radius:12px;padding:28px 24px"><span style="font-size:24px">Limited time: 50% off our CRM add-on</span><br><br>Find people in Northstar, close the deal in the CRM. Your data flows straight across, so you see every deal in one place.<br><br><s style="color:#888888">$20</s> <strong style="font-size:18px">$10</strong> per seat a month*</span>', '#1A1A1A', 13, HELV, pad='24px 32px 8px'), button('Claim your discount', '#1F1F1F'), T('*For eligible customers on a new add-on subscription. Terms apply.', '#777777', 10, HELV, pad='4px 40px 28px'), bg='#FFFFFF', pad='0px'),
], foot('#ECEBDF', '#555555', '<a href="https://example.com" style="color:inherit">Manage preferences</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>', 'Northstar · 12 Example Street, Your City'), '#ECEBDF', ['tips', 'sales', 'digest'])

y15, g15 = '#FFE14A', '#0E4A2A'
def big_number(icon, alt, number, label, body, bg, ink):
 return row(ill(icon, alt, '64px', '28px 40px 4px'), H(number, ink, 42, HELV, pad='4px 40px 4px'), T('- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -', ink, 12, HELV, pad='0px 40px'), T(f'<strong style="font-size:10px">{label}</strong>', ink, 10, HELV, pad='4px 40px'), T(body, ink, 15, HELV, pad='4px 48px 28px'), bg=bg, pad='0px')
def side_stat(icon, alt, value, label, title, body):
 return weighted([[ill(icon, alt, '56px', '20px 8px 0px 40px'), H(value, '#111111', 32, HELV, pad='4px 8px 20px 40px')], [T(f'<strong style="font-size:10px">{label}</strong>', '#111111', 10, HELV, 'left', '28px 40px 4px 8px'), H(title, '#111111', 20, HELV, 'left', '4px 40px 4px 8px'), T(body, '#333333', 13, HELV, 'left', '4px 40px 20px 8px')]], [1, 1], '#FFFFFF', '0px')
studio('year-in-numbers', 'Year in Numbers', 'Retention', 'A year-in-review email built from big numbers: an illustrated banner, bold full-width stats in colour blocks, then side-by-side stats.',
 'Your 2026 in numbers', 'What our customers did this year, and where to see your own stats.', [
 weighted([[H('◕ Sparrowmail', '#111111', 18, BLACK, 'left', '20px 8px 20px 32px')], [T('<strong>Log in</strong>', '#111111', 13, HELV, 'right', '22px 32px 20px 8px')]], [1, 1], y15, '0px'),
 row(ill('year-banner', 'A calendar reading 2026 in numbers, with a chart arrow and a badge saying a big year'), bg=g15, pad='0px'),
 row(T('Our customers spent 2026 growing their audiences and their businesses. Here are the highlights, and your own dashboard is ready when you are.', '#333333', 15, HELV, pad='28px 48px 8px'), pill('See your dashboard', y15, '#111111', outline='#111111', pad='8px 40px 28px', size='13px', inner='8px 20px'), bg='#FFFFFF', pad='0px'),
 big_number('yr-mail', 'An envelope', '1,204,881,930', 'TOTAL EMAILS SENT', 'Our customers reached readers in <strong>190 countries</strong> this year.', y15, '#111111'),
 big_number('yr-chat', 'A chat bubble', '48,302,115', 'TOTAL TEXT MESSAGES SENT', 'More customers went international with text messages.', g15, '#FFFFFF'),
 side_stat('yr-cart', 'A shopping cart', '$8,420', 'AVERAGE REVENUE', 'Sales went up', 'On average, each customer earned $8,420 from email this year.'),
 side_stat('yr-dots', 'Small circles', '210M', 'AUDIENCE GROWTH', 'Lists got bigger', 'Customers added 210 million new contacts between them.'),
 side_stat('yr-pie', 'A chart', '96,330', 'TESTS RUN', 'Testing took off', 'Customers ran 96,330 A/B tests on subject lines and content.'),
 side_stat('yr-calendar', 'A calendar', 'November', 'BUSIEST MONTH', 'November was busiest', 'More emails went out in November than in any other month.'),
 row(H('Don’t miss the big picture', '#FFFFFF', 22, HELV, pad='36px 40px 8px'), T('The full report, and the features behind it, can help you plan the year ahead.', '#FFFFFF', 13, HELV, pad='4px 56px'), pill('Explore your dashboard', g15, '#FFFFFF', outline='#FFFFFF', pad='16px 40px 36px', size='13px', inner='8px 20px'), bg=g15, pad='0px'),
], foot('#FFFFFF', '#555555', 'These figures are illustrative samples. Replace them with your own before sending.', 'Sparrowmail · 12 Example Street, Your City · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#F2F2F2', ['year in review', 'statistics', 'colour blocks'])

slate16 = '#2F3E3A'
studio('please-confirm', 'Please Confirm', 'Transactional', 'A double opt-in email: a small bell, a big headline, a solid button, a friendly illustration, then the same action as an outlined button.',
 'Please confirm your email address', 'One click and you’ll start getting our fortnightly insights.', [
 weighted([[T('Please confirm your email address.', '#555555', 11, HELV, 'left', '16px 8px 16px 32px')], [T('Not displaying properly? ' + lk('View it in your browser', '#3A6A8A') + '.', '#555555', 11, HELV, 'right', '16px 32px 16px 8px')]], [1, 1], '#F7F7F7', '0px'),
 row(H('GROUNDWORK<br>NETWORK', '#111111', 20, BLACK, pad='28px 40px 20px'), bg='#FFFFFF', pad='0px'),
 row(text('', 4, '#111111', pad='2px 0px'), bg=slate16, pad='0px'),
 row(ill('bell', 'A ringing bell', '64px', '36px 40px 8px'), H('Please confirm<br>your email address', slate16, 38, HELV, pad='8px 40px 16px'),
  T('<a href="https://example.com" style="display:block;background:#2F3E3A;color:#FFFFFF;padding:16px;border-radius:999px;text-decoration:none;letter-spacing:2px;font-weight:bold;font-size:13px">CONFIRM MY EMAIL ADDRESS</a>', '#FFFFFF', 13, HELV, pad='8px 80px 16px'),
  ill('envelope-rest', 'A person resting their arms on a large envelope, looking puzzled', '320px', '8px 40px'),
  T('You’re one step away from fortnightly industry insights, plus news, research and ideas to help your business grow.<br><br>Please confirm your email address with the button below. Once you’ve opted in, you can choose what you’d like to hear about.', '#1A1A1A', 16, HELV, 'left', '16px 80px 8px'),
  T('<a href="https://example.com" style="display:block;border:2px solid #2F3E3A;color:#2F3E3A;padding:14px;border-radius:999px;text-decoration:none;letter-spacing:2px;font-weight:bold;font-size:13px">CONFIRM MY EMAIL ADDRESS</a>', '#2F3E3A', 13, HELV, pad='16px 80px 16px'),
  T('If you got this message by mistake, you can ignore it.<br><br>Need help? <strong>' + lk('Contact us', '#C48A4E') + '</strong>.', '#1A1A1A', 14, HELV, 'left', '8px 80px 32px'), bg='#FFFFFF', pad='0px'),
], row(T('This email was sent from an address that doesn’t accept replies. To stop marketing emails, ' + lk('unsubscribe here', '#1A1A1A') + '.<br><br>Groundwork Network · 12 Example Street, Your City<br><a href="https://xem.email" style="color:#1A1A1A">Built with Xem</a>', '#1A1A1A', 11, HELV, 'left', '24px 40px'),
  T('<span style="display:block;background:#2F3E3A;padding:14px 20px"><strong>' + lk('LOG IN', '#FFFFFF') + ' &nbsp;|&nbsp; ' + lk('PREFERENCES', '#FFFFFF') + ' &nbsp;|&nbsp; ' + lk('CONTACT US', '#FFFFFF') + '</strong></span>', '#FFFFFF', 12, HELV, 'left', '8px 0px 0px'), bg='#E6E8E7', pad='0px'),
 '#F7F7F7', ['double opt-in', 'verification', 'illustrated'], marketing=False)

def bordered(colour, title, body):
 return row(T(f'<span style="display:block;border:1px solid #CCCCCC;border-left:12px solid {colour};border-radius:10px;padding:24px 28px"><strong style="font-size:20px">{title}</strong><br>{body}<br><br><strong style="color:#0D47D9">' + lk('Learn more →', '#0D47D9') + '</strong></span>', '#111111', 15, HELV, 'left', '8px 40px'), bg='#FFFFFF', pad='0px')
studio('which-rate', 'Which Rate?', 'Product', 'A plain-English explainer: a topic strip, a big question, a side-by-side comparison illustration, two columns of notes and two accent-bordered cards.',
 'Fixed or variable: which is right for you?', 'A quick guide to the two main kinds of loan rate.', [
 row(H('⌂ Homestead Loans', '#0D47D9', 22, HELV, 'left', '28px 40px'), bg='#FFFFFF', pad='0px'),
 row(T('🎓 <strong>Loan basics:</strong> Types of interest rate', '#111111', 14, HELV, 'left', '16px 40px'), bg='#EEEEEE', pad='0px'),
 row(H('Which rate is right for you?', '#111111', 32, HELV, 'left', '28px 40px 8px'), T('When you take out a home loan, you’ll usually choose between a fixed rate and a variable one. Here’s the difference:', '#111111', 16, HELV, 'left', '4px 40px 16px'), ill('rate-compare', 'Two houses: one with a flat line for a fixed rate, one with a zigzag for a variable rate', pad='8px 40px'), bg='#FFFFFF', pad='0px'),
 weighted([[T('<strong>Fixed rate</strong><br>Steady and predictable<br><br>• Your rate stays the same for the whole term, so your payments do too.', '#111111', 15, HELV, 'left', '16px 8px 16px 40px')], [T('<strong>Variable rate</strong><br>Lower to start<br><br>• Usually cheaper for the first few years, then it follows the market up or down.', '#111111', 15, HELV, 'left', '16px 40px 16px 8px')]], [1, 1], '#FFFFFF', '0px'),
 row(H('How to choose', '#111111', 22, HELV, 'left', '16px 40px 8px'), bg='#FFFFFF', pad='0px'),
 bordered('#1E6E5A', 'A long fixed term', 'If you plan to stay put for years and want a steady payment, this is the one.'),
 bordered('#0D47D9', 'A variable rate', 'If you might move or switch within a few years, the lower starting rate could save you money.'),
 row(H('Wondering what you can afford?', '#111111', 22, HELV, 'left', '32px 40px 8px'), T('Get a quick estimate online and see what you could borrow. It won’t affect your credit score.', '#111111', 15, HELV, 'left', '4px 40px'), T('<strong>' + lk('Get an estimate', '#0D47D9') + ' →</strong>', '#0D47D9', 15, HELV, 'left', '12px 40px 32px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#555555', 'This is general information, not financial advice. Rates and terms vary.', 'Homestead Loans · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Privacy</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#FFFFFF', ['explainer', 'comparison', 'finance'])

purple17 = '#7A2FC0'
def ondemand(img, alt, title, body): return row(ill(img, alt, pad='8px 0px'), H(title, '#0E1A2A', 19, HELV, 'left', '12px 0px 4px'), T(body, '#333333', 13, HELV, 'left', '4px 0px 8px'), pill('Watch on demand', purple17, align='left', pad='8px 0px 28px', size='13px', inner='10px 18px'), bg='#FFFFFF', pad='0px 40px')
studio('on-demand', 'On Demand', 'Events', 'A round-up of on-demand webinars: a purple pixel header, a short note, then four webinar cards each with a big graphic and a button.',
 'Our most-watched webinars of the year', 'All on demand. Pick one for the quiet week.', [
 weighted([[H('■ Relay', '#FFFFFF', 16, BLACK, 'left', '24px 8px 4px 32px'), H('Level up your support in 2027', '#FFFFFF', 26, HELV, 'left', '8px 8px 24px 32px')], [ill('pixel-portrait', 'A pixel pattern', '100%', '16px 32px 16px 8px')]], [3, 1], '#6A2FB0', '0px'),
 row(T('Hey {{first_name}},<br><br>As the year winds down, here’s the easy option: our most-watched and most-saved webinars of the year, all on demand. Put one on with a hot drink and your end-of-year numbers. Future you will be grateful.', '#1A1A1A', 14, HELV, 'left', '24px 40px 8px'), H('Top picks, in case you missed them:', '#0E1A2A', 18, HELV, 'left', '8px 40px'), bg='#FFFFFF', pad='0px'),
 ondemand('webinar-tile-1', 'A webinar card reading Quick fixes are cheap. Lost customers aren’t.', 'Meet Autopilot: automation you stay in charge of', 'How to automate safely, keep your best customers happy and answer faster.'),
 ondemand('webinar-tile-2', 'A webinar card reading Steady Hands', 'Steady Hands: automation with a person in charge', 'A first look at our quality-first approach and the new tools that help teams move faster.'),
 ondemand('webinar-tile-3', 'A webinar card reading The state of customer expectations', 'The state of customer expectations', 'What customers want now, how they feel about automated help, and what to do about it.'),
 row(T('Start the new year with ideas you can use straight away.<br><br>See you in January,<br>The Relay team', '#1A1A1A', 14, HELV, 'left', '8px 40px 28px'), divider('#E2E2E2'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#777777', 'Relay · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">Manage preferences</a>'), '#F2F2F2', ['webinars', 'on demand', 'round-up'])

sky18 = '#3AA6F0'
studio('creator-welcome', 'Creator Welcome', 'Welcome', 'A newsletter welcome: a bold illustrated banner, what to expect, a free guide in a grey card, a photo signature and a list of next steps.',
 'Welcome to the newsletter', 'What to expect, and a free guide to get you started.', [
 row(H('Kiln', sky18, 40, BLACK, pad='32px 40px 16px'), ill('welcome-banner', 'Bold black text reading Welcome to our creator newsletter, with a paper plane and a target', pad='0px 24px'), bg='#FFFFFF', pad='0px'),
 row(T('Hey there,<br><br>Welcome to a community of people who make things for a living, and who know the quiet hours, the doubts, and the push to make something that matters <em>and</em> reaches people.', '#1A1A1A', 14, HELV, 'left', '20px 24px 8px'), H('Here’s what you can expect:', '#1A1A1A', 19, HELV, 'left', '8px 24px'),
  T(ul(['A newsletter every Tuesday with practical advice on reaching more people and selling without feeling salesy', 'A monthly note on new features that save you time']), '#1A1A1A', 14, HELV, 'left', '4px 40px 16px'), bg='#FFFFFF', pad='0px'),
 row(T('<span style="background:#1A1A1A;color:#FFFFFF;padding:8px 14px;border-radius:6px;font-weight:bold">Free download</span>', '#FFFFFF', 13, HELV, 'left', '28px 32px 12px'), ill('guide-card', 'A green card reading Launch your newsletter in 7 days, next to an app screenshot', pad='8px 32px'),
  T('We made this guide because so many creators get stuck planning. Inside are the steps that have helped thousands launch newsletters people actually read.', '#1A1A1A', 13, HELV, 'left', '8px 32px'), pill('Download your free guide', sky18, '#0E2A3A', align='left', pad='12px 32px 28px', size='14px', inner='10px 18px'), bg='#EEEBE5', pad='0px 24px'),
 row(T('See you on Tuesday,', '#1A1A1A', 14, HELV, 'left', '20px 24px 8px'), bg='#FFFFFF', pad='0px'),
 weighted([[P('face3', 'square', '8px 8px 20px 24px')], [T('<strong>Nell</strong><br><br>Marketing at Kiln', '#1A1A1A', 14, HELV, 'left', '40px 24px 20px 8px')]], [1, 2], '#FFFFFF', '0px'),
 row(divider('#E2E2E2'), T('P.S. Can’t wait for the next one?<br><br>' + ul(['<strong>' + lk('See how Kiln works in 8 minutes', '#1A1A1A') + '</strong>: a quick video tour.', '<strong>' + lk('Hear how other creators run their business', '#1A1A1A') + '</strong>: our weekly interview show.', '<strong>' + lk('Start a 14-day free trial', '#1A1A1A') + '</strong>: reach your audience directly.']), '#1A1A1A', 13, HELV, 'left', '8px 24px 24px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#1A1A1A', '<strong>' + lk('Unsubscribe', '#1A1A1A') + ' | ' + lk('Update your profile', '#1A1A1A') + '</strong> | Kiln · 12 Example Street, Your City'), '#EEEBE5', ['welcome', 'creator', 'guide'])

# --- Batch five ---
peri, mint19 = '#7B7FE8', '#8FE0CC'
def product_row(img, alt, name, sub, body, was, now, left):
 pic = [ill(img, alt, '100%', '12px 8px 12px 24px' if left else '12px 24px 12px 8px')]
 txt = [H(name, '#1A1A1A', 20, HELV, 'left', '36px 24px 0px 12px'), T(sub, '#1A1A1A', 14, HELV, 'left', '0px 24px 12px 12px'), T(body, '#1A1A1A', 14, HELV, 'left', '4px 24px 8px 12px'),
        T(f'<a href="https://example.com" style="display:block;background:#8FE0CC;border:1px solid #1A6A5A;border-radius:999px;padding:10px;text-align:center;color:#1A1A1A;text-decoration:none">SHOP NOW <s>{was}</s> {now}</a>', '#1A1A1A', 13, HELV, pad='8px 24px 24px 12px')]
 return weighted([pic, txt] if left else [txt, pic], [1, 1], '#FFFFFF', '0px')
studio('best-sale', 'Best Sale of the Year', 'Commerce', 'A big seasonal sale: a top announcement bar, a bold hero with a product shot, then products alternating left and right with price buttons.',
 'Our biggest sale of the year starts now', 'Up to 50% off everything. The discount is applied at checkout.', [
 row(T('♥ <strong>BIGGEST SALE OF THE YEAR:</strong> up to 50% off, applied at checkout ♥', '#1A1A1A', 11, HELV, pad='8px 24px'), bg=mint19, pad='0px'),
 row(H('FIZZ', '#1A1A1A', 30, BLACK, pad='16px 24px'), bg='#FFFFFF', pad='0px'),
 row(H('Our Best Sale<br>of the Year', '#FFFFFF', 50, HELV, pad='36px 24px 8px'), T('Up to <strong>50% off everything</strong>, applied automatically at checkout.*', '#FFFFFF', 17, HELV, pad='4px 48px'), pill('SHOP THE SALE', mint19, '#1A1A1A', outline='#1A6A5A', pad='16px 40px 8px', size='13px'),
  ill('sale-hero', 'A blue skincare bottle among mirrored balls'), H('We know it’s Monday.<br>We don’t care.', '#FFFFFF', 28, HELV, pad='28px 24px 8px'), T('<strong>THE SALE</strong> starts now. Up to <strong>50% off</strong> bestsellers and gift sets, applied at checkout.', '#FFFFFF', 15, HELV, pad='4px 48px 32px'), bg=peri, pad='0px'),
 row(H('Big results. Bigger savings.', '#1A1A1A', 26, HELV, pad='32px 24px 16px'), bg='#FFFFFF', pad='0px'),
 product_row('product-1', 'A cleansing oil bottle', 'FIRST STEP', 'Cleansing oil', 'A gentle oil that lifts make-up and sunscreen.', '$17', '$8.50', True),
 product_row('product-2', 'An orange exfoliating mask', 'DEEP CLEAN', 'Weekly mask', 'A weekly mask that smooths rough patches.', '$18', '$9', False),
 product_row('product-3', 'A yellow tube of sunscreen', 'PLUS ONE', 'Tinted sunscreen', 'A light mineral SPF with a soft, flattering tint.', '$19', '$9.50', True),
 row(T('♥ <u>CATCH YOU FRIDAY</u> ♥', '#1A1A1A', 13, HELV, pad='16px 24px'), bg=mint19, pad='8px 0px'),
 weighted([[T('<strong>FREE SHIPPING</strong><br>On orders over $39', '#1A1A1A', 10, HELV, pad='20px 4px')], [T('<strong>TEXT US</strong><br>Any time', '#1A1A1A', 10, HELV, pad='20px 4px')], [T('<strong>JOIN THE CLUB</strong><br>Perks and early drops', '#1A1A1A', 10, HELV, pad='20px 4px')]], [1, 1, 1], '#FFFFFF', '0px'),
], foot('#F2F2F2', '#555555', 'hello@example.com · Fizz · 12 Example Street, Your City', '*Discount applies to selected products while stocks last.', '<a href="https://example.com" style="color:inherit">UNSUBSCRIBE</a>'), '#FFFFFF', ['sale', 'products', 'promotion'])

green20 = '#13715A'
studio('a-note-from-us', 'A Note From Us', 'Commerce', 'A seasonal letter from the founder in one colour block, with linked favourites, a white button, a perks row and a category menu.',
 'A note from our founder, and 25% off', 'Our biggest sale of the year, and a few of my favourite things.', [
 row(ill('snow-strip', 'Snowflakes'), H('<span style="background:#FFFFFF;color:#13715A;padding:6px 14px;letter-spacing:4px">WAYFARE</span>', '#13715A', 20, HELV, pad='0px 40px 16px'),
  T('Hi there,<br><br>This time of year always reminds me why we started: to make travelling a little easier, a little calmer, and a lot more enjoyable.<br><br>Whether you’re heading home, going somewhere new or just getting through the airport, thank you for bringing us along. Your stories and photos keep us going.<br><br>Our biggest sale of the year is on. <strong>Take 25% off everything, plus an extra 10% when you spend $750 or more.*</strong> A few of my favourites:<br><br>' + ul(['<strong>' + lk('The Metal Carry-On', '#FFFFFF') + '</strong>: makes even a short work trip feel nicer', '<strong>' + lk('The Kids’ Set', '#FFFFFF') + '</strong>: my two love pulling their own cases', '<strong>' + lk('The Weekender', '#FFFFFF') + '</strong>: the lightest bag we make']) + '<br><br>Here’s to easy journeys and full suitcases.<br><br>Warmly,<br>Maya Lund<br>Founder, Wayfare', '#FFFFFF', 16, HELV, 'left', '16px 48px'),
  T('<a href="https://example.com" style="display:inline-block;background:#FFFFFF;color:#13715A;padding:18px 70px;text-decoration:none;font-weight:bold">SHOP NOW</a>', '#13715A', 16, HELV, 'left', '16px 48px 40px'), bg=green20, pad='0px'),
 weighted([[ill('perk-truck', 'A delivery truck', '60px', '24px 4px 0px'), T('<strong>FREE SHIPPING<br>OVER $100**</strong>', '#1A1A1A', 12, HELV, pad='4px 4px 20px')], [ill('perk-case', 'A suitcase', '60px', '24px 4px 0px'), T('<strong>MADE TO<br>LAST</strong>', '#1A1A1A', 12, HELV, pad='4px 4px 20px')], [ill('perk-pin', 'A map pin', '60px', '24px 4px 0px'), T('<strong>FIND A<br>SHOP</strong>', '#1A1A1A', 12, HELV, pad='4px 4px 20px')]], [1, 1, 1], '#FFFFFF', '0px'),
 row(divider('#DDDDDD'), H('NEW IN', '#1A1A1A', 18, HELV, 'left', '8px 48px'), divider('#DDDDDD'), H('LUGGAGE', '#1A1A1A', 18, HELV, 'left', '8px 48px'), divider('#DDDDDD'), H('BAGS', '#1A1A1A', 18, HELV, 'left', '8px 48px'), divider('#DDDDDD'), H('<span style="color:#E8401A">SALE</span>', '#E8401A', 18, HELV, 'left', '8px 48px'), divider('#DDDDDD'),
  H('Questions?', '#1A1A1A', 18, HELV, 'left', '24px 48px 4px'), T('Read our <strong>' + lk('FAQs', '#1A1A1A') + '</strong> or send us a note <strong>' + lk('here', '#1A1A1A') + '</strong>.', '#1A1A1A', 14, HELV, 'left', '4px 48px 24px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#999999', '*Offer valid while stocks last. **Free standard shipping on eligible orders over $100.', 'WAYFARE · 12 EXAMPLE STREET, YOUR CITY', '<a href="https://example.com" style="color:inherit">PRIVACY</a> · <a href="https://example.com" style="color:inherit">UNSUBSCRIBE</a>'), '#FFFFFF', ['founder letter', 'sale', 'holiday'])

studio('new-episode', 'New Episode', 'Newsletters', 'A podcast episode announcement: a bright cover illustration, a centred title and summary, listen buttons and a pull-quote banner.',
 'New episode: how to negotiate when you’re let go', 'What to ask for, and how to ask for it.', [
 weighted([[T('', '#555555', 1, HELV, pad='0px')], [T(lk('Get the app', '#555555'), '#555555', 13, HELV, 'right', '12px 40px')]], [1, 1], '#EEEEEE', '0px'),
 row(H('’DOORWAY’', '#111111', 34, BLACK, pad='28px 40px 8px'), ill('podcast-cover', 'A green podcast cover reading The Quiet Desk, with a New episode tag and sound waves'), bg='#B784FF', pad='0px'),
 row(H('Negotiating your exit after a layoff', '#111111', 32, HELV, pad='32px 48px 8px'), T('Being let go can leave you stunned and lost for words. So how do you go from shocked employee to calm negotiator? This week: what to ask for, what you’re entitled to, and how to say it.', '#222222', 16, HELV, pad='4px 56px 16px'),
  T('<strong>LISTEN NOW</strong>', '#111111', 15, HELV, pad='12px 40px 8px'), T('<a href="https://example.com" style="display:inline-block;background:#111111;color:#FFFFFF;padding:12px 22px;border-radius:8px;text-decoration:none">▶ Listen on your podcast app</a>&nbsp;&nbsp;<a href="https://example.com" style="display:inline-block;background:#111111;color:#FFFFFF;padding:12px 22px;border-radius:8px;text-decoration:none">● Listen in the browser</a>', '#FFFFFF', 13, HELV, pad='8px 24px 32px'),
  ill('quote-banner', 'An orange banner with a coffee mug and the quote: losing a job says nothing about how good you were at it', pad='0px 40px 32px'), divider('#111111'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#333333', 'Doorway · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Privacy</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">View in browser</a>'), '#EEEEEE', ['podcast', 'episode', 'announcement'])

grape, sage21, coral21 = '#4A1A7A', '#A8D8AC', '#F0907E'
studio('cosy-season', 'Cosy Season', 'Commerce', 'A playful seasonal recipe email: giant purple type on green, three recipe links and a button, then a coral block with a pony mascot.',
 'It’s cosy season. Three recipes to prove it', 'Blondies, energy bites and sticky apples, all with our nut butter.', [
 row(H('<span style="border:3px solid #FFFFFF;border-radius:30px;padding:6px 18px;color:#FFFFFF">TROT</span>', '#FFFFFF', 26, BLACK, pad='32px 40px 16px'), H('It’s<br>C🍂SY<br>season!', grape, 70, BLACK, pad='8px 40px 16px'),
  T('It’s finally here. Big jumpers, falling leaves and nut butter in… everything. So we’ve picked three easy autumn recipes to get you in the mood.<br><br>And soon it’ll be maple season. You know what that means.', grape, 17, HELV, 'left', '8px 56px'),
  H(f'<u>{lk("Gooey Blondies", grape)}</u>', grape, 24, HELV, 'left', '16px 56px 0px'), T('Soft, chewy, and just underbaked enough.', grape, 16, HELV, 'left', '0px 56px'),
  H(f'<u>{lk("Pumpkin Energy Bites", grape)}</u>', grape, 24, HELV, 'left', '16px 56px 0px'), T('A no-bake snack that tastes like pie.', grape, 16, HELV, 'left', '0px 56px'),
  H(f'<u>{lk("Sticky Date Apples", grape)}</u>', grape, 24, HELV, 'left', '16px 56px 0px'), T('Naturally sweet and gloriously messy.', grape, 16, HELV, 'left', '0px 56px'),
  T('<a href="https://example.com" style="display:inline-block;background:#4A1A7A;color:#FFFFFF;padding:16px 36px;text-decoration:none;font-weight:bold;letter-spacing:1px">GET COOKING!</a>', '#FFFFFF', 15, HELV, pad='24px 40px 32px'), bg=sage21, pad='0px'),
 row(ill('pony', 'A cheerful white pony leaping', '260px', '32px 40px 8px'), ill('rope', 'A white rope border', pad='24px 0px'), bg=coral21, pad='0px'),
], foot(coral21, '#5A2A3A', 'No longer want these emails? <strong>' + lk('Unsubscribe', grape) + '</strong>.', 'Trot · 12 Example Street, Your City'), '#8ED2DE', ['seasonal', 'recipes', 'playful'])

red22 = '#E8201A'
def book_item(n, img, alt, title, body):
 return row(T(f'ITEM #{n}', '#111111', 10, HELV, pad='16px 24px 8px'), ill(img, alt, pad='0px 24px'), divider(red22), H(title, '#111111', 40, SERIF, pad='16px 24px'), divider(red22),
  T(body, '#111111', 12, HELV, pad='12px 48px'), T('<a href="https://example.com" style="display:block;border:3px solid #E8201A;border-radius:999px;padding:12px;color:#111111;text-decoration:none;font-family:georgia;font-size:30px">Shop Now</a>', '#111111', 30, SERIF, pad='12px 24px 24px'), bg='#FFFFFF', pad='0px')
studio('new-books', 'New Books In', 'Commerce', 'A bookshop new-arrivals email: a giant masthead, then each book as a black image, a big serif title between red rules, and a round Shop Now button.',
 'New in: three design books worth the shelf space', 'A design manifesto, a comics issue and an art fair catalogue.', [
 row(T('INKWELL BOOKS', '#111111', 10, HELV, pad='16px 24px 0px'), H('Shelf', '#111111', 110, HELV, pad='0px 24px 8px'), divider(red22), bg='#FFFFFF', pad='0px'),
 book_item(1, 'book-1', 'A grey book cover with pink squares', 'Statement and Counter-Statement', 'Ten years after their first book, a design studio returns with a second manifesto, tied to an exhibition of their work.'),
 book_item(2, 'book-2', 'A green book cover', 'Line Work: a comics issue', 'An annual review of graphic design devotes a whole issue to comics, with interviews with six artists.'),
 book_item(3, 'book-3', 'A yellow book cover reading Print Fair Now', 'Print Fair, Now', 'Organisers of independent print fairs around the world share how they run them and what they’re up against.'),
], foot('#FFFFFF', '#555555', 'FOLLOW <u>@INKWELLBOOKS</u>', 'Inkwell Books · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#FFFFFF', ['books', 'new arrivals', 'shop'])

blue23 = '#4048F0'
studio('agency-welcome', 'Agency Welcome', 'Welcome', 'A creative agency’s welcome email in stacked colour blocks: a team photo with stickers, what they do, recent work and a dark contact block.',
 'Welcome. We knew you were a good one', 'What to expect from us, and what we’ve been working on.', [
 row(divider('#111111'), H('HARBOUR &amp; CO', '#111111', 30, BLACK, 'left', '8px 32px'), T('WELCOME NEWSLETTER', '#111111', 10, MONO, 'right', '0px 32px 8px'), divider('#111111'), H('We knew you were a good one', '#111111', 22, HELV, 'left', '8px 32px 8px'), P('team', 'wide', '0px 32px'),
  T('You chose well and subscribed to everything Harbour &amp; Co (and more).<br><br>You’ll get the odd newsletter with recent projects, news and a bit of perspective from our team. For quicker updates, follow us on ' + lk('social media', '#2A5AD8') + '.<br><br>And if you’re ever thinking of us, drop us a line: ' + lk('hello@example.com', '#2A5AD8'), '#1A1A1A', 15, HELV, 'left', '16px 32px 24px'), bg='#FFF4E6', pad='0px'),
 row(H('Did you know we do that?', '#111111', 24, HELV, 'left', '28px 32px 16px'), T('<span style="display:block;background:#4048F0;color:#FFFFFF;padding:90px 0;font-size:24px">H&amp;</span>', '#FFFFFF', 24, HELV, pad='0px 32px'),
  T('We offer branding, creative, web development and marketing, from positioning to campaigns to social.', '#1A1A1A', 15, HELV, 'left', '16px 32px 8px'), T('<a href="https://example.com" style="background:#4048F0;color:#FFFFFF;padding:12px 16px;text-decoration:none">See all services →</a>', '#FFFFFF', 13, HELV, 'left', '12px 32px 32px'), bg='#E6E6FA', pad='0px'),
 row(H('Good work, together.', '#111111', 24, HELV, 'left', '28px 32px 16px'), P('desk', 'wide', '0px 32px'), T('The work is what matters, so here it is. We regularly add new case studies about projects we’ve made with our clients.', '#1A1A1A', 15, HELV, 'left', '16px 32px 8px'), T('<a href="https://example.com" style="background:#4048F0;color:#FFFFFF;padding:12px 16px;text-decoration:none">See the case studies →</a>', '#FFFFFF', 13, HELV, 'left', '12px 32px 32px'), bg='#FFFFFF', pad='0px'),
 row(H('Let’s get to work.', '#FFFFFF', 24, HELV, 'left', '28px 32px 8px'), T('Big project? Busy launch? Need more customers? That’s our kind of thing.<br><br>Get in touch and we’ll talk soon. And tell a friend: a good referral gets you lunch on us.', '#FFFFFF', 15, HELV, 'left', '4px 32px 8px'), T('<a href="https://example.com" style="background:#4048F0;color:#FFFFFF;padding:12px 16px;text-decoration:none">Get in touch →</a>', '#FFFFFF', 13, HELV, 'left', '12px 32px 16px'), ill('ribbon', 'A lilac ribbon reading Good work deserves a fan', '360px', '8px 0px 24px', 'right'), bg='#202020', pad='0px'),
 row(H('HARBOUR &amp; CO', '#111111', 34, BLACK, pad='28px 32px 8px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#777777', 'Over it? <a href="https://example.com" style="color:#2A5AD8">Unsubscribe</a>.', 'Harbour &amp; Co · 12 Example Street, Your City'), '#F4F4F4', ['welcome', 'agency', 'colour blocks'])

crimson = '#A31F3A'
def profile_card(colour, photo, title, body, link):
 return weighted([[P(photo, 'square', '12px 8px 12px 40px')], [H(title, '#1A1A1A', 20, SERIF, 'left', '16px 40px 8px 8px'), T(body, '#333333', 14, HELV, 'left', '4px 40px 8px 8px'), T(f'<strong>{lk(link + " →", crimson)}</strong>', crimson, 14, HELV, 'left', '4px 40px 16px 8px')]], [1, 1], '#FFFFFF', '8px 0px')
studio('driven', 'Driven to Make a Difference', 'Newsletters', 'A university or nonprofit campaign email: a black-and-white photo with an overlapping white card, three bordered profile cards and a campaign close.',
 'Driven to make a difference', 'Stories of alumni, students and donors changing things for the better.', [
 row(H('▣ Northfield', '#1A1A1A', 22, SERIF, pad='20px 40px 0px'), P('train', 'wide', '16px 0px 0px'), bg='#E0DED6', pad='0px'),
 row(T('<span style="display:block;background:#FFFFFF;padding:28px 24px"><span style="font-family:georgia;font-size:28px">Driven to make a difference</span><br><br>In the 1970s, a group of Northfield graduates took a free health clinic on the road in a converted bus, reaching villages hours from the nearest doctor.<br><br><strong style="color:#A31F3A">Read more →</strong></span>', '#333333', 14, HELV, pad='0px 40px 32px'), bg='#E0DED6', pad='0px'),
 row(text('', 4, '#1A1A1A', pad='8px 0px'), bg='#FFFFFF', pad='0px'),
 profile_card('#A31F3A', 'face3', 'The life of a future-ready student', 'From classes to conferences, Lina Park, Class of 2026, is making the most of her time here. Read her story and watch a day in her life.', 'Read Lina’s Q&amp;A'),
 profile_card('#E8B23A', 'face4', 'A scholarship for teachers’ families', '“I want students to choose the right university, not the cheapest one,” says the scholarship’s founder, a former teacher.', 'Read about the gift'),
 profile_card('#2E6E3A', 'face1', 'Closing the health gap', 'A public health graduate is working towards a future where where you grow up doesn’t decide how healthy you are.', 'Read her vision'),
 row(T('<span style="display:block;height:6px;background:linear-gradient(90deg,#A31F3A 25%,#E8B23A 25% 50%,#2E7A8A 50% 75%,#2E6E3A 75%)"></span>', '#1A1A1A', 1, HELV, pad='24px 40px 0px'), H('<span style="color:#A31F3A">Together.</span>', crimson, 40, SERIF, pad='24px 40px 0px'), T('The Northfield Campaign', crimson, 13, HELV, pad='0px 40px 16px'),
  H('This is what Northfield can do.', '#1A1A1A', 28, SERIF, pad='8px 40px'), T('Our community has a long history of making change. We’re asking alumni, parents and friends to help keep that work going.', '#333333', 14, HELV, pad='4px 56px 16px'),
  T(f'<strong>{lk("About the campaign →", crimson)} &nbsp;|&nbsp; {lk("Make a gift →", crimson)}</strong>', crimson, 13, HELV, pad='8px 40px 24px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#555555', 'Northfield University · 12 Example Street, Your City', '<a href="https://example.com" style="color:#A31F3A">Manage your preferences or unsubscribe</a>'), '#FFFFFF', ['campaign', 'profiles', 'nonprofit'])

def source_item(source, time, title, body):
 return [T(f'<span style="background:#E8EBC0;border-radius:999px;padding:3px 10px;font-family:georgia;font-style:italic">{source}</span> &nbsp;<span style="font-size:11px">{time}</span>', '#1A1A1A', 12, HELV, 'left', '16px 32px 4px'), H(title, '#1A1A1A', 20, SERIF, 'left', '8px 32px 4px'), T(body + ' →', '#333333', 13, HELV, 'left', '4px 32px 16px'), divider('#E2E2DC')]
studio('living-lightly', 'Living Lightly', 'Newsletters', 'A curated reading digest: a featured article card with a tag, then a list of links from other sites, each with a source pill and reading time.',
 'The washing setting that’s wearing out your clothes', 'Plus five good reads from around the web.', [
 weighted([[H('Living Lightly', '#3A3A3A', 22, HELV, 'left', '20px 8px 20px 32px')], [T('<span style="background:#E2D8F2;border-radius:999px;padding:4px 12px">2026-11-04</span>', '#3A3A3A', 12, HELV, 'right', '22px 32px 20px 8px')]], [1, 1], '#FFFFFF', '0px'),
 row(T('◎ &nbsp;FEATURED ARTICLE', '#3A3A3A', 13, HELV, 'left', '28px 32px 12px'), P('clothes', 'wide', '0px 32px'), T('<span style="background:#E8EBC0;border-radius:999px;padding:3px 10px;font-family:georgia;font-style:italic">Better Home</span> &nbsp;<span style="font-size:11px">QUICK READ</span>', '#1A1A1A', 12, HELV, 'left', '16px 32px 4px'),
  H('The washing machine setting that’s quietly wearing out your clothes', '#1A1A1A', 20, SERIF, 'left', '8px 32px 4px'), T('You think you’re looking after your clothes, but one common setting could be doing more harm than good.', '#333333', 13, HELV, 'left', '4px 32px 8px'), pill('READ MORE', '#2F3E3A', align='left', pad='8px 32px 28px', size='13px', inner='10px 18px'), bg='#FFFFFF', pad='16px 0px'),
 row(T('))) &nbsp;MORE FROM THE WEB', '#3A3A3A', 13, HELV, 'left', '28px 32px 4px'), divider('#3A3A3A'),
  *source_item('Old House Journal', 'MEDIUM READ', '6 old-fashioned design details worth bringing back', 'Classic features with plenty of life left in them'),
  *source_item('Garden Weekly', 'MEDIUM READ', 'The right way to mow for a healthier lawn', 'Small changes that make a big difference'),
  *source_item('Fix It Monthly', 'QUICK READ', 'Why you shouldn’t use gloss paint on walls', 'Experts say it shows every bump and patch'),
  *source_item('The Food Desk', 'MEDIUM READ', 'What everyone gets wrong about gluten', 'An expert explains the science and the myths'),
  T('<em>We sometimes share products we like and may earn a commission on sales.</em>', '#555555', 11, SERIF, 'left', '16px 32px 24px'), bg='#FFFFFF', pad='16px 0px'),
 row(H('◎', '#3A3A3A', 30, HELV, pad='32px 40px 0px'), H('Live life a little lighter.', '#1A1A1A', 24, SERIF, pad='8px 40px'), T('Was this email forwarded to you?', '#1A1A1A', 14, HELV, pad='4px 40px'), pill('SUBSCRIBE TO LIVING LIGHTLY', '#D8CCF0', '#3A3A6A', pad='12px 40px 24px', size='13px'), bg='#F7F7F4', pad='0px'),
], foot('#F7F7F4', '#777777', 'Living Lightly · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Fewer emails</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#F2F2EE', ['digest', 'reading', 'curated'])

navy24, plum24 = '#243042', '#8A2A4A'
studio('kind-words', 'Kind Words', 'Retention', 'A gentle nonprofit email: an illustrated header with a big serif line, a button, three illustrated vases, a yellow strip, a quote with flowers and an app banner.',
 'The difference a kind message can make', 'A few words can mean a lot to someone going through a hard time.', [
 row(H('<span style="color:#8A2A4A">♥</span> CARING CIRCLE', plum24, 22, HELV, 'left', '20px 40px 8px'), bg='#FFFFFF', pad='0px'),
 weighted([[H('The difference a kind message can make', navy24, 38, SERIF, pad='48px 8px 16px 40px')], [ill('care-person', 'A woman holding a heart, with small hearts around her', '100%', '16px 40px 0px 8px')]], [1, 1], '#FFFFFF', '0px'),
 row(T('When someone you love is unwell, it’s hard to know how to help. One simple way is to leave them a message on their Caring Circle page. Short notes like <strong>“Thinking of you”</strong> or <strong>“You’re not on your own”</strong> can mean more than you’d expect.', '#1A1A1A', 16, HELV, 'left', '16px 40px 8px'), button('Leave a message', navy24),
  ill('vases', 'Three vases of heart-shaped flowers in blue, plum and yellow', pad='40px 24px 8px'), H('Still not sure what to say?', navy24, 28, HELV, pad='16px 40px 8px'),
  T('When someone is ill, they often feel cut off. <strong>Messages from friends and family are one of the things people say helps most.</strong> Here are some ideas to help you feel confident reaching out.', '#1A1A1A', 16, HELV, 'left', '4px 40px 8px'), button('Read more', navy24), bg='#FFFFFF', pad='0px'),
 row(T('Could you or someone you love use one place to share health updates?<br><strong>' + lk('Start a Caring Circle page today.', navy24) + '</strong>', '#1A1A1A', 14, HELV, pad='16px 32px'), bg='#F0D57A', pad='0px'),
 weighted([[H('“ —', navy24, 44, SERIF, 'left', '20px 8px 0px 40px'), T('“Being able to let people know what was happening, and reading their messages, really kept me going. The hospital days are long, and it can get lonely. It felt like company, without having to text everyone one by one.”', '#1A1A1A', 15, HELV, 'left', '4px 8px 8px 56px'), T('<strong>A Caring Circle member and carer</strong>', '#1A1A1A', 14, HELV, pad='8px 8px 24px 40px')], [ill('bouquet-hand', 'A hand holding a bunch of heart flowers', '100%', '40px 40px 24px 8px')]], [3, 1], '#FFFFFF', '0px'),
 row(text('', 4, '#1A1A1A', pad='4px 0px'), bg='#A8BCDD', pad='0px'),
 weighted([[H('Try the new Caring Circle app', '#FFFFFF', 26, SERIF, pad='28px 8px 28px 40px')], [T('<a href="https://example.com" style="display:inline-block;border:1px solid #FFFFFF;border-radius:8px;padding:12px 18px;color:#FFFFFF;text-decoration:none">Get the app</a>', '#FFFFFF', 14, HELV, pad='36px 40px 28px 8px')]], [2, 1], navy24, '0px'),
], foot('#FFFFFF', '#333333', 'Caring Circle is a donor-supported nonprofit.', 'Caring Circle · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Privacy</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#EEEEEE', ['nonprofit', 'illustrated', 'community'])

lav = '#C8A8F8'
studio('small-but-seen', 'Small but Seen', 'Product', 'A short insight email: a bold headline and pill button, a story, a chart in a soft gradient frame, an arrow list of what worked and a second button.',
 'Size doesn’t matter. Being the clearest answer does.', 'How a small company keeps showing up next to giants in AI answers.', [
 row(H('➶ BRIGHTLINE', '#111111', 18, BLACK, 'left', '32px 32px 4px'), H('Size doesn’t matter. Being the clearest answer does.', '#111111', 34, HELV, 'left', '16px 32px 12px'), pill('Build your visibility plan', lav, '#111111', align='left', pad='8px 32px 24px', size='14px', inner='12px 28px'),
  T('Don’t count yourself out just because you’re up against bigger brands.<br><br>' + lk('Clipwise', '#3A5AD8') + ' isn’t the biggest name in video editing. But in AI answers, it shows up next to much larger brands, because it focuses on being the best answer for one specific kind of customer.', '#222222', 14, HELV, 'left', '8px 32px 16px'),
  ill('line-chart-card', 'A chart comparing visibility scores for five sites', pad='0px 32px'), H('Here’s how they did it:', '#111111', 16, HELV, 'left', '24px 32px 8px'),
  T('→ &nbsp;Narrowed their focus → podcast editing, not “all-in-one video”<br>→ &nbsp;Wrote genuinely helpful content → real questions, real answers<br>→ &nbsp;Built pages worth citing → comparisons and how-tos<br>→ &nbsp;Used real product screenshots → visuals AI can reuse<br>→ &nbsp;Got mentioned elsewhere → reviews, press and creators', '#222222', 14, HELV, 'left', '4px 32px 16px'),
  T('When AI tools can understand, trust and cite your content, you can compete with companies ten times your size.', '#222222', 14, HELV, 'left', '4px 32px 8px'), pill('Check your visibility', lav, '#111111', align='left', pad='12px 32px 32px', size='14px', inner='12px 28px'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('icon-news', 'Newsroom', '32px', '24px 4px 4px'), T(lk('Newsroom', '#333333'), '#333333', 13, HELV, pad='4px 4px 24px')], [ill('icon-person', 'Partners', '32px', '24px 4px 4px'), T(lk('Become a partner', '#333333'), '#333333', 13, HELV, pad='4px 4px 24px')], [ill('icon-briefcase', 'Agencies', '32px', '24px 4px 4px'), T(lk('Hire an agency', '#333333'), '#333333', 13, HELV, pad='4px 4px 24px')]], [1, 1, 1], '#F2F3F5', '0px'),
], foot('#F2F3F5', '#555555', 'Brightline · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#FFFFFF', ['insight', 'strategy', 'chart'])

# --- Batch six ---
indigo25 = '#4A5AF0'
def dot_section(title, img, alt, body, cta):
 return row(H(f'● {title}', '#1A1A1A', 15, HELV, 'left', '20px 32px 8px'), ill(img, alt, pad='4px 32px'), T(body, '#333333', 13, HELV, 'left', '12px 32px 4px'), pill(cta, indigo25, pad='8px 32px 16px', size='13px', inner='10px 28px'), divider('#ECECF2'), bg='#FFFFFF', pad='0px')
studio('the-messenger', 'The Messenger', 'Product', 'A monthly product update from a finance app: a short intro, then sections each with a dot heading, a product screenshot and a pill button.',
 'The Messenger: live insights, smarter categories and more', 'Four new features in your account this month.', [
 row(T('<span style="border:1px solid #DADAE6;border-radius:999px;padding:8px 14px">◎ L E D G E W O R K S</span>', '#1A1A1A', 13, HELV, 'left', '28px 32px 16px'), T('Welcome to The Messenger, our monthly round-up of what’s new in Ledgeworks.<br><br>This month: live insights, a connection to your AI assistant, smarter categories, automatic invoice reminders and more.', '#1A1A1A', 13, HELV, 'left', '8px 32px 12px'), divider('#ECECF2'), bg='#FFFFFF', pad='0px'),
 dot_section('Live financial insights, built in', 'ui-bars', 'A bar chart of monthly spending with a trend line', '<u>' + lk('Insights', indigo25) + '</u> are now in your dashboard, with a live view of your company’s money. Charts update as you go, so you can track trends, compare periods, and ask questions like “Why did software spending go up?”', 'See your insights'),
 dot_section('Connect your AI assistant', 'ui-connect', 'Your account connected to an assistant by dotted lines', 'Connect your account securely to the AI assistant you already use. Ask about your finances and explore your data without exporting anything. Access is read-only and you can switch it off at any time.', 'Connect your account'),
 dot_section('Transactions, sorted for you', 'ui-categories', 'A list of transactions with suggested categories', 'We now put your transactions into categories automatically, so you can see where your money goes. You’re always in control: review, approve or change any suggestion, and it learns as you go.', 'Learn more'),
 dot_section('Automatic invoice reminders', 'ui-reminders', 'A card showing a reminder schedule for an invoice', 'Turn on payment reminders and we’ll nudge your customers on a simple schedule: a week before, a day before, a day after and a week after the due date.', 'Try it now'),
 row(H('●● Plus member? Book time with our team', '#1A1A1A', 15, HELV, 'left', '20px 32px 8px'), T('Plus members can book a session with a specialist to set up their account, walk through key workflows and make sure everything is configured for how they work.', '#333333', 13, HELV, 'left', '4px 32px'), pill('Book now', indigo25, pad='8px 32px 16px', size='13px', inner='10px 28px'),
  H('● Small improvements', '#1A1A1A', 15, HELV, 'left', '12px 32px 4px'), T(ul([lk('Transfer rules', indigo25) + ' can now have nicknames and a history.', 'Sign in faster with ' + lk('passkeys', indigo25) + ' on your phone.']) + '<br><br>See more in our ' + lk('monthly round-up', indigo25) + '.<br><br>Thanks,<br>The Ledgeworks team', '#333333', 13, HELV, 'left', '4px 32px 24px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#777777', '<em>This email is an advertisement from Ledgeworks.</em>', 'Ledgeworks · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#FFFFFF', ['product update', 'finance', 'screenshots'])

navy26 = '#1F1A5A'
def article_card(photo, kind, title, mins): return T(f'<span style="display:block;border:1px solid #E2E2EA;box-shadow:0 2px 6px rgba(0,0,0,.08)">' + f'<img src="https://images.unsplash.com/{PHOTOS[photo][0]}?auto=format&fit=crop&w=600&h=300&q=80" alt="{PHOTOS[photo][1]}" width="100%" style="display:block"><span style="display:block;padding:12px"><span style="font-size:9px;color:#888">{kind}</span><br><strong style="font-size:13px">{title}</strong><br><br><br><span style="font-size:9px;color:#888">{mins} minute read</span></span></span>', '#1A1A1A', 12, HELV, 'left', '16px 8px 16px 32px')
studio('the-latest-news', 'The Latest News', 'Newsletters', 'A business bank’s content newsletter: a stamped photo header, article cards beside summaries, a loan panel, a video row and a dark resources banner.',
 'Practical ideas to keep your business moving', 'Invoicing, loans and accounts, explained in a few minutes each.', [
 row(H('🌳 Oakmoor Bank', '#1A1A1A', 24, HELV, pad='24px 32px'), P('laptops', 'wide', '0px 32px'), H('<span style="border-left:8px solid #1F1A5A;padding-left:12px">THE LATEST OAKMOOR NEWS</span>', navy26, 34, BLACK, 'left', '16px 32px 8px'), bg='#FFFFFF', pad='0px'),
 row(H('Practical ideas to help your business move forward', navy26, 22, HELV, 'left', '24px 32px 8px'), T('Running a business means balancing growth, cash and a hundred decisions. This week, a few quick reads to help you save time, understand your options and plan your next step.', '#333333', 14, HELV, 'left', '4px 32px 16px'), bg='#FFFFFF', pad='0px'),
 weighted([[article_card('screen', 'Business banking', 'Automate your invoicing and get paid faster', 6)], [H('Spend less time on invoices, more on the business', navy26, 18, HELV, 'left', '20px 32px 8px 12px'), T('Manual invoicing slows you down and delays payment. Here’s how automation can cut errors and help you get paid faster.', '#333333', 14, HELV, 'left', '4px 32px 8px 12px'), T('<strong><u>' + lk('Explore invoice automation', navy26) + '</u></strong>', navy26, 14, HELV, 'left', '8px 32px 16px 12px')]], [1, 1], '#FFFFFF', '0px'),
 weighted([[H('Small-business loans: myths and facts', navy26, 18, HELV, 'left', '20px 12px 8px 32px'), T('Loans can help you grow, but a few common myths put people off. We clear them up so you can decide with confidence.', '#333333', 14, HELV, 'left', '4px 12px 8px 32px'), T('<strong><u>' + lk('Read the full breakdown', navy26) + '</u></strong>', navy26, 14, HELV, 'left', '8px 12px 16px 32px')], [article_card('meeting', 'Finance your business', 'Separating fact from fiction about loans', 4)]], [1, 1], '#FFFFFF', '0px'),
 row(H('Oakmoor Express small-business loans', navy26, 22, HELV, 'left', '28px 32px 8px'), T('When an opportunity comes up, timing matters. Express loans make the process quicker and simpler.<br><br><strong>' + ul(['For loans between $10,000 and $350,000*', 'A simple application and faster decisions', 'Terms that fit your business', 'No early repayment fees']) + '</strong><br><br><strong><u>' + lk('Start your loan enquiry', navy26) + '</u></strong>', navy26, 14, HELV, 'left', '4px 32px 28px'), bg='#F4F3FA', pad='0px'),
 weighted([[P('face2', 'portrait', '24px 8px 24px 40px')], [H('Business current or savings account: do you need both?', navy26, 22, HELV, 'left', '32px 32px 8px 12px'), T('Not sure you have the right accounts? This short video explains how each one helps you manage cash and plan for growth.', '#333333', 14, HELV, 'left', '4px 32px 8px 12px'), T('<strong><u>' + lk('Watch the video', navy26) + '</u></strong>', navy26, 14, HELV, 'left', '8px 32px 24px 12px')]], [2, 3], '#FFFFFF', '0px'),
 row(H('Want more practical guides?', '#FFFFFF', 18, HELV, pad='28px 32px 4px'), T('Our resource centre has quick, useful guides you can put to work today.', '#FFFFFF', 14, HELV, pad='4px 32px'), button('Explore resources', '#FFFFFF', ink=navy26), bg=navy26, pad='0px 0px 8px'),
], foot('#FFFFFF', '#555555', '*Your rate, terms and timing depend on your credit history and business.', 'Oakmoor Bank · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">Manage preferences</a>'), '#F4F1EA', ['newsletter', 'banking', 'articles'])

studio('youre-invited', 'You’re Invited', 'Events', 'A clean livestream invitation in rounded white cards on warm grey: a big headline, a short letter, bullet points and a full-width black button.',
 'You’re invited: how to run agents that actually help', 'A livestream on building agents your team can rely on.', [
 row(H('<span style="color:#F2643A">◖</span> loam', '#1A1A1A', 30, HELV, 'left', '24px 24px'), bg='#FFFFFF', pad='0px'),
 row(T('<strong>YOU’RE INVITED</strong>', '#1A1A1A', 12, HELV, 'left', '28px 16px 4px'), H('How to use agents to grow without the guesswork', '#1A1A1A', 30, HELV, 'left', '4px 16px 24px'), bg='#F3F1EE', pad='0px'),
 row(H('Hey {{first_name}},', '#1A1A1A', 22, HELV, 'left', '24px 16px 16px'),
  T('Our research agent has been run <strong>millions of times</strong>. Anyone can run an agent. Far fewer get results they can act on.<br><br>Good data, clear checks and steady infrastructure are what make the difference.', '#1A1A1A', 15, HELV, 'left', '8px 16px'),
  T('<strong>Join us to learn:</strong><br>' + ul(['How to build agents that understand your business', 'A walkthrough of an agent that researches and scores accounts', 'How to test and tune results before you scale', 'How to write personal outreach at scale']), '#1A1A1A', 15, HELV, 'left', '8px 16px'),
  T('◐ <strong>Livestream:</strong> Wednesday 1 July, 9:00am Pacific', '#1A1A1A', 15, HELV, 'left', '8px 16px 16px'),
  T('<a href="https://example.com" style="display:block;background:#1F1F1F;color:#FFFFFF;padding:16px;border-radius:14px;text-decoration:none;font-weight:bold">Register now →</a>', '#FFFFFF', 15, HELV, pad='8px 16px 32px'), bg='#FFFFFF', pad='0px'),
], row(H('<span style="color:#FFFFFF">◖</span> loam', '#FFFFFF', 34, HELV, 'left', '36px 32px 8px'), H('Grow with better data, and the means to act on it', '#FFFFFF', 22, HELV, 'left', '8px 32px 16px'), T('© 2026 Loam · 12 Example Street, Your City<br><br>Not interested? Unsubscribe ' + lk('here', '#BBBBBB') + '.<br><a href="https://xem.email" style="color:#BBBBBB">Built with Xem</a>', '#BBBBBB', 13, HELV, 'left', '8px 32px 36px'), bg='#1C1C1C', pad='0px'),
 '#F3F1EE', ['invitation', 'livestream', 'minimal'])

def thumb_item(img, alt, title, body, link):
 return weighted([[ill(img, alt, '100%', '16px 8px 16px 32px')], [H(title, '#1A1A1A', 17, HELV, 'left', '16px 32px 4px 12px'), T(body, '#333333', 13, HELV, 'left', '4px 32px 4px 12px'), T('☐ <u>' + lk(link + ' →', '#1A1A1A') + '</u>', '#1A1A1A', 13, HELV, 'left', '8px 32px 16px 12px')]], [2, 3], '#FFFFFF', '0px')
studio('whats-new-list', 'What’s New', 'Product', 'A “what’s new” email: a blue block header with a geometric tile, a short intro, then five updates each with a thumbnail, a sentence and a link.',
 'What’s new: session replays, skills and group chats', 'Five updates this month, plus a customer story.', [
 weighted([[H('◗ Buildly', '#FFFFFF', 15, HELV, 'left', '24px 8px 8px 32px'), H('What’s<br>new', '#FFFFFF', 54, HELV, 'left', '4px 8px 24px 32px')], [ill('build-hero', 'An orange tile with a figure shape', '100%', '0px')]], [1, 1], '#3A5AE8', '0px'),
 row(T('Hi there,<br><br>This month your apps get more to build on, and your assistant gets more done. Watch real user sessions, bring in a design, connect your data and teach your assistant new skills.<br><br>Here’s the rundown.', '#1A1A1A', 13, HELV, 'left', '24px 32px 16px'), divider('#1A1A1A'), bg='#FFFFFF', pad='0px'),
 thumb_item('thumb-play', 'A session recording with a play bar', 'Watch your users in action', 'Session recordings are here. Watch real sessions, see the exact moment someone gets stuck, and fix what you saw, not what you guessed.', 'Turn on recordings'),
 thumb_item('thumb-merge', 'A design being brought into the app', 'Bring a design into Buildly', 'That mock-up doesn’t have to stay a mock-up. Bring it in and it becomes a working app, ready to build on.', 'Import your design'),
 thumb_item('thumb-data', 'Two data sources connected', 'Build on your own data', 'If your data lives in a warehouse, you can now build on it directly. Connect it once for the whole workspace.', 'Connect your data'),
 thumb_item('thumb-chat', 'A group chat with messages', 'Your assistant joins the group chat', 'Add your assistant to a group chat and it answers questions and runs tasks right in the thread, for everyone in it.', 'Set up a group'),
 thumb_item('thumb-skills', 'A list of assistant skills', 'Give your assistant new skills', 'Skills add ready-made know-how, so it does a job well with less prompting: build slides from a brief, plan a trip from your inbox, or track competitors.', 'Turn on a skill'),
 row(divider('#1A1A1A'), H('Community spotlight', '#1A1A1A', 22, HELV, 'left', '16px 32px 4px'), T('See how people turn ideas into real businesses on Buildly.', '#333333', 13, HELV, 'left', '4px 32px'), bg='#FFFFFF', pad='0px'),
 thumb_item('spotlight-logo', 'The tutr logo on black', 'Real revenue in weeks', 'Tutr is a marketplace for tutors, built entirely on Buildly: find a tutor, book a lesson and pay in one place. It’s already taking bookings.', 'See what they built'),
 row(T('That’s it for this month. Go build something good.', '#1A1A1A', 13, HELV, 'left', '16px 32px'), button('Back to building', '#1F1F1F'), T('Keep building,<br>The Buildly team', '#1A1A1A', 13, HELV, 'left', '16px 32px 24px'), bg='#FFFFFF', pad='0px'),
], foot('#EEEEEE', '#777777', 'You’re getting this because you signed up for updates.', 'Buildly · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a> · <a href="https://example.com" style="color:inherit">Manage preferences</a>'), '#EEEEEE', ['product update', 'whats new', 'list'])

lime27, char27 = '#C8F59A', '#262626'
def dark_card(badge, title, img, alt, body, cta):
 return row(ill(badge, 'Number badge', '80px', '20px 0px 8px'), H(title, lime27, 22, HELV, pad='8px 32px 12px'), *( [ill(img, alt, pad='0px 32px')] if img else []), T(body, '#E8E8E8', 13, HELV, pad='12px 40px 8px'), T(f'<a href="https://example.com" style="display:inline-block;border:1px solid #C8F59A;border-radius:999px;padding:10px 26px;color:#C8F59A;text-decoration:none">{cta}</a>', lime27, 13, HELV, pad='12px 32px 28px'), bg=char27, pad='0px')
studio('hey-there-dark', 'Hey There', 'Newsletters', 'A dark-mode newsletter of rounded cards: a cheeky intro with a numbered contents list, then three numbered stories with lime badges.',
 'New look, same us: three things this month', 'A fresh design, a reality check on values marketing, and a deliverability guide.', [
 row(T('<span style="border:1px solid #C8F59A;border-radius:6px;padding:6px 14px">▣ sendloop</span>', lime27, 16, HELV, pad='32px 32px 8px'), H('Hey there,', lime27, 40, HELV, pad='8px 32px 8px'), T('Notice anything different? 👀<br><br>Yep, this newsletter has had a makeover. New look, fresh colours. Ten out of ten.<br><br>And you know what goes with it?<br><br>You, reading it.<br><br>Anyway, here’s what we’ve got this month:', '#E8E8E8', 13, HELV, pad='8px 40px'),
  T('1. Why your campaigns aren’t landing (and how to fix it)<br>2. What happens when you stop marketing on gut feel<br>3. Your deliverability survival guide for 2027', lime27, 13, HELV, pad='8px 40px 28px'), bg=char27, pad='0px'),
 row(text('', 4, '#FFFFFF', pad='8px 0px'), bg='#3A3A3A', pad='0px'),
 dark_card('num-badge-1', 'Values marketing? People can tell when it’s just for one month', 'dark-report', 'A green report card with a bar chart', 'We asked 4,000 people how they feel about brands that change their logo for one month and then go quiet. The answer was clear: people notice whether you keep it up the rest of the year.', 'Read the report'),
 row(text('', 4, '#FFFFFF', pad='8px 0px'), bg='#3A3A3A', pad='0px'),
 dark_card('num-badge-2', 'What happens when you stop marketing on gut feel', None, None, 'A small snack company from the north knew exactly what they were making. Their email marketing? By their own admission, guesswork.<br><br>They switched to us and planned properly. A year later their open rates are up by a third, and text messages finally work for them.', 'Read their story'),
 row(text('', 4, '#FFFFFF', pad='8px 0px'), bg='#3A3A3A', pad='0px'),
 dark_card('num-badge-3', 'Your deliverability survival guide for 2027', 'dark-at', 'A glowing card with an @ sign', 'Ever sent a campaign and wondered if anyone got it? This guide is for that.<br><br>Twelve steps, the technical bits in plain English, and everything you need to stay out of spam.', 'Read the guide'),
 row(text('', 4, '#FFFFFF', pad='8px 0px'), bg='#3A3A3A', pad='0px'),
 row(T('You’ve got enough to worry about. Let email marketing be the easy part.', '#E8E8E8', 13, HELV, pad='32px 40px 8px'), pill('Join Sendloop now', lime27, '#111111', pad='8px 32px 8px', size='13px'), T('<em>– Ana and the Sendloop team</em>', '#AAAAAA', 12, HELV, pad='4px 32px 16px'), bg=char27, pad='0px'),
], foot(char27, '#AAAAAA', 'Sendloop · 12 Example Street, Your City', '<a href="https://example.com" style="color:#C8F59A">Unsubscribe</a>'), '#3A3A3A', ['dark mode', 'newsletter', 'numbered'])

pine28 = '#0B2A1E'
def hotel(photo, place, name, body):
 return row(P(photo, 'wide'), T('<span style="display:inline-block;background:#F7E8B0;border:2px solid #C8A030;border-radius:50%;width:30px;height:30px;line-height:30px">◉</span>', '#1A1A1A', 12, HELV, pad='8px 40px 0px'), T(place, '#333333', 13, HELV, pad='4px 40px 0px'), H(name, pine28, 26, HELV, pad='4px 40px 4px'), T(body, '#1A1A1A', 15, HELV, pad='4px 56px'), pill('Explore now', pine28, pad='12px 40px 24px', size='13px'), bg='#FFFFFF', pad='0px')
studio('one-of-a-kind', 'One of a Kind', 'Commerce', 'A travel list of remarkable places to stay: a centred headline, then each hotel as a full-width photo, a location, a name, a line and a pill button.',
 'Hotels that are the whole trip', 'A palace on a lake, a treetop lodge and a lighthouse.', [
 row(H('◉◉ Roamwell', pine28, 24, HELV, pad='24px 40px 8px'), T('Your next big trip', pine28, 16, HELV, pad='4px 40px 0px'), H('One-of-a-kind stays', pine28, 40, HELV, pad='4px 40px 16px'), P('lake', 'wide'),
  T('Some hotels aren’t just somewhere to sleep, they <em>are</em> the trip. These stays are worth travelling for. Here’s how to make the most of each one.', '#1A1A1A', 16, HELV, pad='16px 56px 8px'), T('<a href="https://example.com" style="display:block;background:#0B2A1E;color:#FFFFFF;padding:14px;border-radius:999px;text-decoration:none;font-weight:bold">Meet the winners</a>', '#FFFFFF', 14, HELV, pad='8px 40px 24px'), bg='#FFFFFF', pad='0px'),
 hotel('sea', 'The Lakes, North Region', 'The Island House', 'A former manor on its own island, reached only by boat. Expect marble floors and a lot of quiet.'),
 hotel('forest', 'The Rainforest, South Coast', 'The Canopy Lodge', 'Rooms built at treetop height, with morning walks and evening river trips.'),
 hotel('coast', 'The Harbour Town', 'The Lighthouse Rooms', 'Three suites inside a working lighthouse, with a bath that looks straight out to sea.'),
 row(divider('#0B2A1E'), T('But wait, there’s more: a floating hotel on a loch, a suite on stilts in a vineyard, and a gloriously over-the-top retro hotel in the city.', '#1A1A1A', 16, HELV, pad='16px 56px 8px'), pill('See the full list', pine28, pad='8px 40px 32px', size='13px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#0B2A1E', 'Please don’t reply to this email. For help, visit our ' + lk('help centre', '#0B2A1E') + '.', 'Roamwell · 12 Example Street, Your City', '<strong>' + lk('Unsubscribe', '#0B2A1E') + ' · ' + lk('Privacy', '#0B2A1E') + '</strong>'), '#FFFFFF', ['travel', 'list', 'hotels'])

trail29 = '#3A3A2A'
def quote_photo(q, who, role, photo):
 return weighted([[T(f'“{q}”', trail29, 12, HELV, 'left', '16px 8px 4px 40px'), T(f'<strong>{who}</strong><br><em style="font-size:10px">{role}</em>', trail29, 11, HELV, 'left', '8px 8px 4px 40px'), T('<strong><u>' + lk('See the full list →', trail29) + '</u></strong>', trail29, 10, HELV, 'left', '8px 8px 16px 40px')], [P(photo, 'square', '16px 40px 16px 8px')]], [1, 1], '#F3F1EA', '0px')
studio('trail-notes', 'Trail Notes', 'Newsletters', 'An outdoors newsletter on warm paper: a script-style masthead, a short letter, stories with quotes and photos, a two-by-two grid and a gradient fact panel.',
 'Trail Notes: stories from the outdoors this month', 'Trails, people and a few places to visit this summer.', [
 row(T(lk('Find a trail', trail29) + ' &nbsp; ' + lk('Saved', trail29) + ' &nbsp; ' + lk('Start a walk', trail29), trail29, 10, HELV, pad='8px 40px'), H('▲ Pathway', trail29, 20, HELV, pad='24px 40px 0px'), H('<em>Trail Notes</em>', '#C8803A', 56, SERIF, pad='0px 40px 16px'),
  T('Dear {{first_name}},<br><br>The outdoors has a way of bringing people together. This month we’re sharing stories from people who make space for everyone on the trail, and a few of their favourite walks.<br><br>See you out there,<br>The Pathway team', trail29, 12, HELV, 'left', '16px 40px'), divider('#C8C2B0'), bg='#F3F1EA', pad='0px'),
 row(T('Community month', trail29, 11, HELV, 'left', '8px 40px 4px'), H('Finding belonging outdoors: five walkers’ stories', trail29, 22, HELV, 'left', '4px 40px 8px'), T('Five walkers on what the outdoors means to them, and the trails they keep going back to.', trail29, 12, HELV, 'left', '4px 40px'), T('<span style="border:1px solid #3A3A2A;border-radius:999px;padding:8px 14px;font-size:11px">Read the full story</span>', trail29, 11, HELV, 'left', '16px 40px 8px'), bg='#F3F1EA', pad='0px'),
 quote_photo('The trail gave me room to breathe. Nobody cared what I looked like, only whether I’d brought snacks.', 'Ro Hart', 'Founder of a walking group', 'peaks'),
 quote_photo('Being outside makes me feel braver in the rest of my life. That’s why we bring other people along.', 'Jess and Tam Ode', 'Outdoor writers', 'mountain'),
 row(divider('#C8C2B0'), H('A partnership worth walking for', trail29, 22, HELV, 'left', '16px 40px 8px'), bg='#F3F1EA', pad='0px'),
 weighted([[T('We’ve worked with a local charity since 2022 that runs free group walks for people who’ve never hiked before. Last year they ran more than a hundred.', trail29, 12, HELV, 'left', '8px 8px 8px 40px'), T('<strong><u>' + lk('Learn more →', trail29) + '</u></strong>', trail29, 10, HELV, 'left', '4px 8px 16px 40px')], [P('walk', 'square', '8px 40px 16px 8px')]], [1, 1], '#F3F1EA', '0px'),
 row(divider('#C8C2B0'), T('Trail highlights', trail29, 11, HELV, 'left', '8px 40px 4px'), H('Ideas for summer', trail29, 22, HELV, 'left', '4px 40px 8px'), bg='#F3F1EA', pad='0px'),
 weighted([[P('meadow', 'square', '8px 6px 4px 40px'), T('<strong>Wildflower walks</strong><br>What’s in bloom near you<br><strong><u>See the list →</u></strong>', trail29, 11, HELV, 'left', '4px 6px 16px 40px')], [P('lake', 'square', '8px 40px 4px 6px'), T('<strong>The best waterfall trails</strong><br>Cool off on a summer walk<br><strong><u>See the list →</u></strong>', trail29, 11, HELV, 'left', '4px 40px 16px 6px')]], [1, 1], '#F3F1EA', '0px'),
 weighted([[P('stars', 'square', '8px 6px 4px 40px'), T('<strong>Sunrise and sunset walks</strong><br>The best times to be out<br><strong><u>See the list →</u></strong>', trail29, 11, HELV, 'left', '4px 6px 16px 40px')], [P('valley', 'square', '8px 40px 4px 6px'), T('<strong>Walks near the city</strong><br>No car needed<br><strong><u>See the list →</u></strong>', trail29, 11, HELV, 'left', '4px 40px 16px 6px')]], [1, 1], '#F3F1EA', '0px'),
 row(T('<span style="display:block;background:linear-gradient(135deg,#D8803A,#8AA0B0);padding:48px 32px;color:#FFFFFF"><span style="font-size:12px">Did you know?</span><br><br><span style="font-size:24px">The country’s longest marked walking trail is over 2,000 miles long, and most people walk it in sections.</span></span>', '#FFFFFF', 13, HELV, pad='24px 0px'), bg='#F3F1EA', pad='0px'),
], foot('#FFFFFF', '#555555', 'Pathway · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Email preferences</a> · <a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#DFE3DF', ['outdoors', 'newsletter', 'stories'])

studio('tune-up', 'Tune Up', 'Events', 'A webinar reminder with a sticker-collage header, the date beside a calendar icon, a big headline, a boxed note and a red button.',
 'Reminder: the tune-up session is next week', 'Three real before-and-after emails, fixed live.', [
 row(H('✉ Good Mail Studio', '#1A1A1A', 20, HELV, pad='24px 40px'), bg='#F2F2F2', pad='0px'),
 row(ill('sticker-pair', 'A sticker collage of two cartoon faces and the words Tune Up, with a ribbon reading A friendly reminder'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('calendar-red', 'A calendar', '64px', '16px 4px 16px 0px', 'right')], [H('18 June | 9:00am PT / 12:00pm ET', '#1A1A1A', 20, HELV, 'left', '32px 40px 16px 8px')]], [1, 4], '#FFFFFF', '0px'),
 row(H('Can your readers actually read your emails?', '#1A1A1A', 34, HELV, pad='8px 40px 16px'), T('<span style="display:block;background:#FFF8EE;border:1px solid #1A1A1A;border-radius:8px;padding:12px"><strong>Quick reminder: the tune-up is next week.</strong></span>', '#1A1A1A', 15, HELV, pad='4px 40px 16px'),
  T('On <strong>18 June</strong>, we’re walking through three real before-and-after email makeovers: what was broken, how we fixed it, and what difference it made.<br><br>A university, a supplement brand and a skincare brand. Three industries, three makeovers, one session.<br><br>Our designer Yuna is joining to explain the decisions behind each one.<br><br><strong>If you’ve ever felt stuck between “fine” and “actually great”, this one’s for you.</strong>', '#1A1A1A', 16, HELV, 'left', '8px 40px'),
  button('Count me in →', '#F0503A'), bg='#FFFFFF', pad='0px'),
 weighted([[ill('calendar-red', 'A calendar', '64px', '16px 4px 32px 0px', 'right')], [H('18 June | 9:00am PT / 12:00pm ET', '#1A1A1A', 20, HELV, 'left', '32px 40px 32px 8px')]], [1, 4], '#FFF0EE', '0px'),
], foot('#F2F2F2', '#555555', '<strong>Good Mail Studio.</strong> Where good emails get made.', 'Good Mail Studio · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#F2F2F2', ['reminder', 'webinar', 'stickers'])

def numbered_note(n, body): return weighted([[H(f'{n:02}.', '#F0503A', 30, HELV, 'left', '12px 4px 8px 40px')], [T(f'<span style="display:block;border:1px solid #1A1A1A;border-radius:6px;padding:12px 14px">{body}</span>', '#1A1A1A', 14, HELV, 'left', '8px 40px 8px 4px')]], [1, 5], '#FFFFFF', '0px')
studio('now-one', 'Now One Place', 'Transactional', 'A merger or rebrand notice: a photo collage header, a personal letter, a bordered announcement box, numbered changes in boxes, and an FAQ link.',
 'Brushwork is now part of Good Mail Studio', 'What changes for you (very little) and what stays the same.', [
 row(H('✉ Good Mail Studio', '#1A1A1A', 20, HELV, pad='24px 40px'), bg='#F2F2F2', pad='0px'),
 row(ill('photo-collage', 'A collage of three team photos on a grid'), H('Hi {{first_name}},', '#F0503A', 24, HELV, 'left', '24px 40px 8px'),
  T('When we joined forces, the goal was simple: <strong>close the gap between finding inspiration and making something.</strong><br><br>Since then you’ve had more chances to learn and connect, and your feedback has shaped what we built. Thank you ♥<br><br>But going from inspiration to building still meant two separate trips.<br><br><strong>So we changed that.</strong>', '#1A1A1A', 14, HELV, 'left', '8px 40px'), bg='#FFFFFF', pad='0px'),
 weighted([[H('Brushwork is now<br>Good Mail Studio.', '#1A1A1A', 28, HELV, 'left', '24px 8px 24px 40px')], [ill('merge-icons', 'The two company marks side by side', '100%', '32px 40px 24px 8px')]], [3, 2], '#EEEEEE', '0px'),
 row(T('<strong>Everything you’ve made in Brushwork now lives inside Good Mail Studio.</strong> You can move between finding inspiration and building emails in one place.<br><br>This is just the start. Some things, like logos and colours, are still changing, so you may spot a bit of the old name here and there.', '#1A1A1A', 14, HELV, 'left', '24px 40px 8px'), H('What this means for you:', '#1A1A1A', 24, HELV, 'left', '16px 40px 8px'), bg='#FFFFFF', pad='0px'),
 numbered_note(1, 'Your account, designs, folders and team stay exactly as they are.'),
 numbered_note(2, 'You can sign in at the ' + lk('new address', '#F0503A') + ' or the old one. We’ll send you to the right place.'),
 numbered_note(3, '“Get inspired” in the sidebar takes you straight to the gallery.'),
 numbered_note(4, 'You can now save emails from the gallery and come back to them later.'),
 numbered_note(5, 'If you had both accounts with the same email address, they’re now linked.'),
 numbered_note(6, 'You’ll start getting our newsletter twice a week. Change your ' + lk('email preferences', '#F0503A') + ' any time.'),
 row(H('Got questions?', '#1A1A1A', 24, HELV, 'left', '24px 40px 8px'), T('Our FAQ covers what’s changed, what hasn’t and what’s next. ' + lk('Read the FAQ →', '#F0503A'), '#1A1A1A', 14, HELV, 'left', '4px 40px 32px'), bg='#FFFFFF', pad='0px'),
], foot('#F2F2F2', '#555555', 'Brushwork is now Good Mail Studio. You’re getting this because you have an account.', 'Good Mail Studio · 12 Example Street, Your City'), '#F2F2F2', ['rebrand', 'notice', 'account'], marketing=False)

def note_card(avatar, alt, name, body, img):
 return row(T(f'<span style="display:block;border:1px solid #E2E2E2;border-radius:16px;padding:20px"><strong>{name}</strong> <span style="background:#EEEEEE;font-size:11px;padding:2px 6px;border-radius:4px">You follow</span> <span style="float:right;color:#888">2d</span><br>{body}<br><span style="color:#1A6AD8">Read more »</span></span>', '#1A1A1A', 15, HELV, 'left', '8px 40px 0px'), *([P(img, 'wide', '8px 40px 0px')] if img else []), T('♡ 3 &nbsp;&nbsp; 💬 &nbsp;&nbsp; ↻ &nbsp;&nbsp; ⇪', '#555555', 13, HELV, 'left', '8px 56px 16px'), bg='#FFFFFF', pad='0px')
studio('new-notes', 'New Notes', 'Newsletters', 'A social activity digest: a headline naming who posted, then post cards with names, a short excerpt, images and reactions, and a bright app button.',
 'Two writers you follow posted new notes', 'Catch up on what people you follow are saying.', [
 row(H('<span style="color:#FF6A1A">▤</span>', '#FF6A1A', 26, HELV, 'left', '24px 40px 4px'), H('Rowan Ellis and The Analytics Letter posted new notes', '#333333', 26, HELV, 'left', '8px 40px 16px'), bg='#FFFFFF', pad='0px'),
 note_card('avatar-2', 'Rowan Ellis', 'Rowan Ellis', 'Stop charging for access, charge for results. This week I looked at three companies that changed how they price…', 'screen'),
 note_card('avatar-5', 'The Analytics Letter', 'The Analytics Letter', 'New edition out! We wrote about this exact topic two weeks ago…', None),
 row(T('<a href="https://example.com" style="display:block;background:#FF6A1A;color:#FFFFFF;padding:14px;border-radius:6px;text-decoration:none;font-weight:bold">See more notes in the app</a>', '#FFFFFF', 15, HELV, pad='16px 40px 16px'), T('<span style="border:1px solid #DDDDDD;border-radius:6px;padding:10px 20px;color:#FF6A1A">▤ Start writing</span>', '#FF6A1A', 14, HELV, pad='8px 40px 24px'), bg='#FFFFFF', pad='0px'),
], foot('#FFFFFF', '#777777', '© 2026 Notebook · 12 Example Street, Your City', '<a href="https://example.com" style="color:inherit">Unsubscribe</a>'), '#FFFFFF', ['notifications', 'social', 'digest'])
