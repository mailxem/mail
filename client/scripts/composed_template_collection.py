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
