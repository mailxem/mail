"""Build Xem's original, editable Unlayer starter designs and lightweight gallery previews.
Run: python3 scripts/generate_template_starters.py
The JSON is the source used by Unlayer; sending HTML is exported by the existing editor.
"""
import json
import html
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'public/assets/template-starters'
OUT.mkdir(parents=True, exist_ok=True)
SANS = {'label': 'Arial', 'value': 'arial,helvetica,sans-serif'}
SERIF = {'label': 'Georgia', 'value': 'georgia,palatino'}
MONO = {'label': 'Courier New', 'value': 'courier new,courier,monospace'}
PHOTOS = {
 'forest': ('photo-1441974231531-c6227db76b6e', 'Sunlight falling through a green forest'),
 'interior': ('photo-1494438639946-1ebd1d20bf85', 'Quiet interior with books and natural textures'),
 'coffee': ('photo-1495474472287-4d71bcdd2085', 'Freshly brewed coffee and everyday rituals'),
 'work': ('photo-1515378791036-0648a3ef77b2', 'A bright workspace ready for a new idea'),
 'mountain': ('photo-1470770841072-f978cf4d019e', 'A calm lake framed by mountains'),
 'event': ('photo-1470229722913-7c0e2dbbafd3', 'Lights above a live community gathering'),
 'fashion': ('photo-1515886657613-9f3515b0c78f', 'A considered everyday outfit in warm tones'),
}

def text(copy, size=16, color='#30332F', font=None, align='left', pad='10px 36px', kind='text'):
 return {'type':kind,'values':{'text':copy,'fontSize':f'{size}px','fontFamily':font or SANS,'color':color,'textAlign':align,'lineHeight':'160%' if kind=='text' else '110%','containerPadding':pad, **({'headingType':'h1' if size>=32 else 'h2'} if kind=='heading' else {})}}
def title(copy, **kw): return text(copy,kind='heading',**{'size':44,**kw})
def label(copy,color='#50624B',align='left',pad='12px 36px'): return text(copy,11,color,MONO,align,pad)
def button(copy, color='#253C31', href='https://example.com', align='left', ink='#FFFFFF'):
 return {'type':'button','values':{'text':copy,'href':{'name':'web','attrs':{'href':'{{href}}','target':'{{target}}'},'values':{'href':href,'target':'_blank'}},'buttonColors':{'color':ink,'backgroundColor':color,'hoverColor':ink,'hoverBackgroundColor':color},'size':{'autoWidth':True,'width':'100%'},'fontFamily':SANS,'fontSize':'16px','lineHeight':'140%','textAlign':align,'padding':'14px 24px','border':{},'borderRadius':'6px','containerPadding':'18px 36px 28px'}}
def image(key):
 photo,alt=PHOTOS[key]
 return {'type':'image','values':{'src':{'url':f'https://images.unsplash.com/{photo}?auto=format&fit=crop&w=1200&h=660&q=85','width':600,'height':330,'autoWidth':False,'maxWidth':'100%'},'altText':alt,'textAlign':'center','containerPadding':'16px 0px','action':{'name':'web','values':{'href':'','target':'_blank'}}}}
def divider(color='#DDDCD5'):
 return {'type':'divider','values':{'width':'100%','border':{'borderTopWidth':'1px','borderTopStyle':'solid','borderTopColor':color},'textAlign':'center','containerPadding':'18px 36px'}}
def row(*blocks,bg='#FFFFFF',pad='12px 0px'):
 return {'columns':[list(blocks)],'bg':bg,'pad':pad}
def columns(left,right,bg='#FFFFFF'):
 return {'columns':[left,right],'bg':bg,'pad':'12px 0px'}
def masthead(brand, issue, ink, bg):
 return row(text(brand,20,ink,SERIF,pad='20px 36px 4px'), label(issue,ink),bg=bg)
def section(kicker,heading,body,color='#30332F'):
 return [label(kicker,color,pad='12px 24px'), title(heading,size=25,color=color,pad='4px 24px'), text(body,color=color,pad='12px 24px')]
def footer(marketing=True):
 blocks=[divider(),text('Your brand · Your postal address',12,'#666A73',align='center'),text('You’re receiving this because you signed up for updates.',12,'#666A73',align='center')] if marketing else [divider(),text('Questions? Reply to this email. We’re here to help.',12,'#666A73',align='center')]
 return row(*blocks,text('<a href="https://xem.email">Built with Xem</a>',12,'#666A73',align='center',pad='16px 36px 28px'),bg='#F7F7F5')

TEMPLATES=[]
def add(key,name,category,description,subject,preheader,rows,bg='#ECEDE8',marketing=True,tags=None):
 TEMPLATES.append(dict(key=key,name=name,category=category,description=description,subject=subject,preheader=preheader,background=bg,rows=rows+[footer(marketing)],tags=tags or [],marketing=marketing))

add('field-notes','Field Notes','Newsletters','An earthy editorial with a photographic lead story and a two-column reading list.','A little perspective for your week','Fresh ideas, a slower pace, and one story worth your time.',[
 masthead('FIELD / NOTES','VOL. 028  ·  THE WEEKLY EDITION','#294332','#F3F1E8'),
 row(title('Good things take<br>a little growing.',font=SERIF,color='#294332',size=48),text('A few things we found, loved, and thought you might too.',color='#52624E'),image('forest'),bg='#F3F1E8'),
 row(label('THE LONG READ'),title('Make room for a<br>different kind of progress.',size=32,font=SERIF),text('Hi {{first_name}},<br><br>Not every step forward needs to be a leap. This week, we’re exploring the small habits that make meaningful work possible.'),button('Read the story →')),
 columns(section('01 / IDEAS','A fresh point of view','Three questions to take into your next creative project.'),section('02 / PEOPLE','Meet the makers','A conversation about patience, process, and finding your voice.')),
],tags=['editorial','nature','weekly'])
add('sunday-letter','The Sunday Letter','Newsletters','Warm cream paper, generous serif type, and an intimate letter from the editor.','A quiet moment, just for you','Pour a coffee. This one is worth slowing down for.',[
 masthead('The Sunday Letter','A NOTE FROM THE EDITOR','#513C30','#FFF9EE'),
 row(label('TAKE FIVE. YOU’VE EARNED IT.','#886648'),title('Less noise.<br>More meaning.',font=SERIF,size=52,color='#513C30'),image('coffee'),text('Dear {{first_name}},<br><br>There is something about a slow morning that makes the important things a little easier to see. A conversation. A new idea. A little time to make something with care.',color='#513C30'),divider('#DDD0BB'),title('One thought to keep.',font=SERIF,size=28,color='#513C30'),text('You don’t have to do everything today. Choose the one thing that deserves your attention, and give it your best.',color='#513C30'),button('Read this week’s letter','#513C30'),text('Until next Sunday,<br><strong>The editorial team</strong>',font=SERIF,color='#513C30'),bg='#FFF9EE')
],bg='#E9DFCF',tags=['letter','creator','minimal'])
add('signal-digest','Signal Digest','Newsletters','A sharp, cobalt-accented digest for curated links, insights, and industry news.','The signal in your inbox this week','Three ideas. One useful read. No endless scrolling.',[
 masthead('SIGNAL_','YOUR WEEK, DISTILLED.','#1545C1','#EDF2FF'),
 row(title('Stay curious.<br>Skip the noise.',size=48,color='#1545C1'),text('The stories moving our industry forward, selected for people who make things happen.'),bg='#EDF2FF'),
 row(label('01  /  THE BIG IDEA','#1545C1'),title('The next chapter starts<br>with a better question.',size=30),text('A practical look at the decisions changing how teams work, build, and grow.'),button('Get the full picture','#1545C1')),
 row(divider(),label('02  /  WORTH YOUR TIME','#1545C1'),title('A small guide to better systems.',size=26),text('Simple principles you can put to work before your next meeting.'),divider(),label('03  /  SAVE FOR LATER','#1545C1'),title('Our reading list, refreshed.',size=26),text('New voices, useful research, and a few unexpected discoveries.'))
],tags=['digest','technology','curated'])
add('community-roundup','Around the Table','Newsletters','A warm community roundup with member stories, a gathering, and a personal invitation.','Good people. Great things happening.','Your monthly dispatch from the community.',[
 masthead('AROUND THE TABLE','PEOPLE MAKE THE DIFFERENCE.','#773D25','#FBE4D4'),
 row(title('A little closer,<br>even from here.',font=SERIF,color='#773D25',size=46),text('New faces, shared ideas, and moments that made this month.',color='#773D25'),image('event'),bg='#FBE4D4'),
 columns(section('MEMBER SPOTLIGHT','Meet your next collaborator','This month, we’re celebrating people turning thoughtful ideas into real things.'),section('ON THE CALENDAR','Good conversations ahead','Join our next community gathering. Bring a question, a story, or just yourself.')),
 row(button('See what’s happening','#773D25'),text('Have something to share? Hit reply. We’d love to hear from you.'))
],tags=['community','roundup','people'])
add('warm-welcome','A Warm Welcome','Welcome','A friendly lilac welcome with three clear next steps and one inviting call to action.','You’re in. Make yourself at home.','A few good things to get you started.',[
 masthead('hello, friend.','YOUR NEXT CHAPTER STARTS HERE','#4D347D','#EDE5FF'),
 row(title('Good to<br>have you here.',size=54,color='#4D347D'),text('Hi {{first_name}},<br><br>Welcome to a place for fresh ideas and meaningful connections. We’re glad you found us.',color='#4D347D'),button('Find your first inspiration','#4D347D'),bg='#EDE5FF'),
 row(label('MAKE YOURSELF AT HOME','#4D347D'),title('01. Start with something you love.',size=23),text('Explore our favorite stories, resources, and ideas.'),divider(),title('02. Make it your own.',size=23),text('Choose the topics that matter most to you.'),divider(),title('03. Say hello.',size=23),text('Reply and tell us what brought you here. Real people read these emails.'))
],tags=['welcome','community','onboarding'])
add('first-steps','First Steps','Welcome','A structured SaaS onboarding email with a bold checklist and a focused activation button.','Your first win is just a few steps away','Let’s turn your new workspace into a working one.',[
 masthead('YOUR WORKSPACE','LET’S GET YOU STARTED','#153E3A','#DDF5EE'),
 row(title('From blank page<br>to first win.',size=46,color='#153E3A'),text('Hi {{first_name}}, here’s the short path to getting something great out into the world.'),button('Open your workspace','#153E3A'),bg='#DDF5EE'),
 row(label('YOUR GETTING-STARTED GUIDE'),title('01 / Add your essentials',size=24),text('Set up your profile and bring your team together.'),divider(),title('02 / Make something yours',size=24),text('Pick a starter, add your voice, and build your first project.'),divider(),title('03 / Share your first result',size=24),text('Give it a final look, then take the next step with confidence.'),text('Need a hand? Reply to this email and we’ll help you find your way.'))
],tags=['saas','activation','checklist'])
add('founders-note','A Note from the Founder','Welcome','A personal, typography-led introduction with an editorial signature and a simple reply prompt.','A quick hello from our founder','Why we started, and why we’re glad you’re here.',[
 masthead('a personal note','FROM OUR FOUNDER TO YOU','#262626','#FFFFFF'),
 row(title('We’re building<br>something that matters.',font=SERIF,size=42),text('Hi {{first_name}},<br><br>I wanted to personally welcome you. We started this company with a simple belief: the best products make room for people, not just processes.<br><br>Your feedback will help shape what comes next. We’re here to listen, learn, and make something you’ll love using.'),divider(),title('What would make this<br>more useful for you?',size=26,font=SERIF),text('Hit reply and tell me. I read every message.'),text('With gratitude,<br><strong>Your name</strong><br>Founder, Your brand',font=SERIF),button('Get to know us','#262626'))
],bg='#F0EEEB',tags=['founder','personal','minimal'])
add('release-notes','Freshly Shipped','Product','A dark, high-contrast release announcement with a bright feature panel and clear changelog.','Freshly shipped. Ready for you.','The latest improvements, built around your feedback.',[
 masthead('SHIP / LOG','PRODUCT UPDATE  ·  THE LATEST EDITION','#CFFB73','#171D1A'),
 row(label('LESS FRICTION. MORE FLOW.','#CFFB73'),title('Your workflow.<br>Upgraded.',size=50,color='#FFFFFF'),text('A little faster. A lot more thoughtful. Meet the improvements designed to make your next project feel effortless.',color='#D7E2D8'),button('Explore what’s new','#CFFB73',ink='#172316'),bg='#171D1A'),
 row(label('THE HEADLINE FEATURE','#26432A'),title('Everything you need.<br>Right where you need it.',size=34,color='#26432A'),text('Keep your ideas, feedback, and next steps together. Spend less time looking and more time making.'),bg='#E6F4D5'),
 columns(section('ALSO NEW','Small details. Big difference.','Smarter defaults and thoughtful shortcuts that save you a few clicks.'),section('MADE BETTER','You asked. We listened.','A cleaner experience, with improvements inspired by your feedback.'))
],bg='#DDE4DC',tags=['saas','changelog','dark'])
add('feature-spotlight','The Spotlight','Product','Bold tangerine type and a simple problem-to-solution story for your next feature.','Meet your new favorite feature','One useful idea. A smoother everyday experience.',[
 masthead('THE SPOTLIGHT','A BETTER WAY TO GET THERE','#712F18','#FFD9BF'),
 row(title('Big ideas.<br>Fewer steps.',size=58,color='#712F18'),text('The next thing you’ll wonder how you worked without.',color='#712F18'),bg='#FFD9BF'),
 row(image('work'),label('INTRODUCING YOUR NEXT FAVORITE'),title('Make space for<br>the work that matters.',font=SERIF,size=34),text('Bring the pieces together in one thoughtful place. Less back and forth, fewer loose ends, and a clearer path from idea to done.'),button('Take a closer look','#712F18'),divider(),text('<strong>Made for real life.</strong><br>Easy to learn. Simple to share. Ready when you are.'))
],tags=['feature','launch','bold'])
add('new-collection','Objects of Everyday','Commerce','An understated product collection with warm neutrals, editorial photography, and paired features.','Good design. Everyday company.','Meet the new collection, made for the moments in between.',[
 masthead('FORM & EVERYDAY','THE NEW COLLECTION','#3F392F','#F4EFE5'),
 row(title('Made to live<br>beautifully.',font=SERIF,size=52,color='#3F392F'),image('interior'),text('Thoughtful objects. Honest materials. A little more pleasure in the everyday.',color='#3F392F'),button('Explore the collection','#3F392F'),bg='#F4EFE5'),
 columns(section('CONSIDERED DETAILS','Nothing extra.<br>Nothing missing.','Every curve, texture, and finish has a reason to be here.'),section('FOR THE LONG RUN','Keep the things<br>you love.','Designed to feel at home today, and for years to come.'))
],bg='#DDD8CE',tags=['collection','home','editorial'])
add('weekend-edit','The Weekend Edit','Commerce','A confident berry-and-pink promotion with a prominent offer and an editorial finish.','Your weekend just got a little brighter','A considered selection, at a very good price.',[
 masthead('THE WEEKEND EDIT','A LITTLE SOMETHING FOR YOU','#6F153F','#FFE7F0'),
 row(label('A GOOD TIME TO FIND YOUR FAVORITE','#6F153F',align='center'),title('A little treat.<br>On us.',size=56,color='#6F153F',align='center'),text('20% OFF',size=54,color='#6F153F',align='center',pad='10px 36px'),text('Your next favorite is waiting.<br>Use code <strong>WEEKEND20</strong> at checkout.',color='#6F153F',align='center'),button('Find something you love','#6F153F',align='center'),bg='#FFE7F0'),
 row(image('fashion'),text('Good things, thoughtfully chosen.<br>Discover the pieces we keep coming back to.',align='center'),text('Replace this with your offer dates, exclusions, and terms before sending.',12,'#68636A',align='center'))
],bg='#EFDAE2',tags=['sale','promotion','offer'])
add('saved-for-you','Saved for You','Commerce','A gentle cart reminder with a warm product story and one clear return-to-cart action.','Still thinking it over?','Your favorites are right where you left them.',[
 masthead('YOUR BRAND','A GENTLE REMINDER','#4C432D','#FAF5E6'),
 row(title('Good taste.<br>Great choice.',font=SERIF,size=48,color='#4C432D'),text('Something caught your eye. We kept it here so you can pick up where you left off.'),image('interior'),button('Back to your favorites','#4C432D'),bg='#FAF5E6'),
 row(title('A few things to feel good about.',size=24),text('Thoughtful details. Helpful people. A product you’ll be happy to make part of your day.'),divider(),text('Have a question before you decide? Just reply — we’re happy to help.'))
],tags=['cart','reminder','retention'])
add('gather-together','Gather Together','Events','An expressive, purple event invitation with a date block, agenda, and RSVP.','Good ideas happen together. You’re invited.','A gathering for curious people and meaningful conversations.',[
 masthead('GATHER / TOGETHER','AN INVITATION TO SOMETHING GOOD','#F3EFFF','#35204E'),
 row(title('Better ideas.<br>In good company.',size=48,color='#F3EFFF'),text('An evening of fresh perspectives, honest conversations, and people you’ll be glad you met.',color='#DCD0EA'),bg='#35204E'),
 columns([label('SAVE THE DATE','#5C3589'),title('OCT<br>24',size=56,color='#5C3589',pad='10px 24px')],[title('A place to connect.',size=27,pad='12px 24px'),text('6:00–8:00 PM · Your timezone<br>Your venue, Your city<br><br>Replace these event details before sending.',pad='12px 24px')],bg='#EDE3F8'),
 row(label('WHAT TO EXPECT','#5C3589'),text('<strong>Fresh perspectives.</strong> A short talk to get us thinking.<br><br><strong>Real connections.</strong> Time to meet people doing interesting things.<br><br><strong>A little inspiration.</strong> Something to take into your next project.'),button('Save my seat','#5C3589'))
],bg='#DCD6E5',tags=['invitation','conference','rsvp'])
add('live-session','See You Live','Events','A focused webinar reminder with a clean agenda and an unmistakable join button.','We’re going live. Bring your questions.','Your reminder, the details, and one easy way to join.',[
 masthead('THE LIVE SESSION','LEARN SOMETHING USEFUL','#173B60','#E6F2FF'),
 row(label('YOU’RE ON THE LIST','#173B60'),title('Let’s make<br>this practical.',size=46,color='#173B60'),text('Join us for a hands-on conversation about turning good ideas into better results.'),button('Join the live session','#173B60'),bg='#E6F2FF'),
 row(title('Your session at a glance',size=26),text('<strong>When</strong><br>Your date · Your time and timezone<br><br><strong>Where</strong><br>Online — join using the button above<br><br><strong>Bring</strong><br>Your questions and a little curiosity'),divider(),label('ON THE AGENDA','#173B60'),text('01. The idea behind the approach<br>02. A real example, step by step<br>03. Your questions, answered'))
],tags=['webinar','reminder','live'])
add('order-confirmed','Order, Confirmed','Transactional','A reassuring purchase confirmation with a clear order summary and delivery next steps.','It’s official. Your order is in.','Here’s what happens next.',[
 masthead('YOUR BRAND','ORDER CONFIRMATION','#234B3A','#EAF5EC'),
 row(label('THANK YOU FOR CHOOSING US','#234B3A'),title('Good things<br>are on their way.',font=SERIF,size=42,color='#234B3A'),text('Hi {{first_name}},<br><br>We’ve received your order. Here’s a little recap for your records.'),bg='#EAF5EC'),
 row(label('ORDER SUMMARY'),title('Your order number',size=24),text('Your product name · Quantity 1<br>Your product details'),divider(),text('<strong>Subtotal</strong> · Your subtotal<br><strong>Shipping</strong> · Your shipping cost<br><strong>Total</strong> · Your order total'),divider(),title('What happens next?',size=24),text('We’ll send another email when your order is on its way. You can check the latest status anytime.'),button('View order','#234B3A'))
],marketing=False,tags=['receipt','order','confirmation'])
add('reset-password','A Fresh Start','Transactional','A concise password-reset email with a prominent action and clear security context.','Reset your password','A secure next step for your account.',[
 masthead('YOUR BRAND','ACCOUNT SECURITY','#253A66','#F0F3FA'),
 row(label('PASSWORD RESET','#253A66'),title('A fresh start.<br>Just one step away.',size=38,color='#253A66'),text('Hi {{first_name}},<br><br>We received a request to reset the password for your account. Use the button below to choose a new one.'),button('Reset my password','#253A66'),divider(),text('If you didn’t request this, you can ignore this email. Your password will stay the same.<br><br>For your security, never share your reset link with anyone.',14,'#5B6477'))
],marketing=False,tags=['password','security','account'])
add('verify-email','You’re Almost In','Transactional','A crisp email verification message with a calm mint palette and one clear next step.','One last step: confirm your email','Let’s make sure we have the right address.',[
 masthead('YOUR BRAND','EMAIL CONFIRMATION','#164D44','#E3F6F0'),
 row(title('One small click.<br>A good beginning.',size=42,color='#164D44'),text('Hi {{first_name}},<br><br>Confirm your email address to finish setting up your account. It helps us keep your account secure and your updates in the right place.'),button('Confirm my email','#164D44'),divider(),text('If you didn’t create this account, no action is needed. You can safely ignore this message.',14,'#5A6D67'),bg='#E3F6F0')
],marketing=False,tags=['verification','account','confirmation'])
add('open-door','The Door Is Open','Retention','A calm, image-led re-engagement email that invites people back without pressure.','There’s always a place for you here','A few things have changed. The welcome hasn’t.',[
 masthead('COME AS YOU ARE','IT’S BEEN A LITTLE WHILE','#254D57','#E7F2F1'),
 row(title('New paths.<br>Same warm welcome.',font=SERIF,size=46,color='#254D57'),image('mountain'),text('Hi {{first_name}},<br><br>Life gets busy. We get it. Whenever you’re ready, there are fresh ideas and good things waiting for you here.',color='#254D57'),button('Take a look around','#254D57'),bg='#E7F2F1'),
 row(title('A little of what you’ve missed.',size=27,font=SERIF),text('Thoughtful new features. Stories worth your time. A community that keeps growing.<br><br>Come back on your terms. We’ll be here.'))
],tags=['winback','reengagement','personal'])


def design_for(t):
 counters={}; uid=0
 def meta(kind):
  nonlocal uid
  uid+=1; counters[kind]=counters.get(kind,0)+1
  return {'id':f"{t['key']}-{uid}", 'meta':{'htmlID':f'{kind}_{counters[kind]}','htmlClassNames':kind}}
 rows=[]
 for spec in t['rows']:
  r=meta('u_row'); cols=[]
  for blocks in spec['columns']:
   c=meta('u_column'); contents=[]
   for block in blocks:
    b=meta('u_content_'+block['type'])
    values={'anchor':'','displayCondition':None,'selectable':True,'draggable':True,'duplicatable':True,'deletable':True,'hideable':True,'_meta':b['meta'],'linkStyle':{'inherit':True,'linkColor':'#514293','linkHoverColor':'#514293','linkUnderline':True,'linkHoverUnderline':True},**block['values']}
    contents.append({'id':b['id'],'type':block['type'],'values':values})
   cols.append({'id':c['id'],'contents':contents,'values':{'border':{},'padding':'0px','backgroundColor':'','_meta':c['meta']}})
  rows.append({'id':r['id'],'cells':spec.get('cells',[1]*len(cols)),'columns':cols,'values':{'backgroundColor':'','columnsBackgroundColor':spec['bg'],'padding':spec['pad'],'backgroundImage':{'url':'','fullWidth':True,'repeat':'no-repeat','size':'custom','position':'center'},'selectable':True,'draggable':True,'duplicatable':True,'deletable':True,'hideable':True,'hideDesktop':False,'noStackMobile':False,'_meta':r['meta']}})
 return {'schemaVersion':18,'counters':counters,'body':{'id':t['key']+'-body','rows':rows,'headers':[],'footers':[],'values':{'backgroundColor':t['background'],'contentWidth':'600px','contentAlign':'center','fontFamily':SANS,'textColor':'#30332F','preheaderText':t['preheader'],'language':{'htmlLang':'en'},'linkStyle':{'body':True,'linkColor':'#514293','linkHoverColor':'#514293','linkUnderline':True,'linkHoverUnderline':True},'_meta':{'htmlID':'u_body','htmlClassNames':'u_body'}}}}

# These are read-only gallery previews of the same block tree. Actual delivery
# always uses Unlayer.exportHtml(), preserving its email-client compatibility code.
def preview_for(t,d):
 def render_block(b):
  v=b['values']; typ=b['type']; pad=v.get('containerPadding','0'); family=v.get('fontFamily',SANS)['value']; align=v.get('textAlign','left'); common=f"font-family:{family};font-size:{v.get('fontSize','16px')};line-height:{v.get('lineHeight','160%')};color:{v.get('color','#30332F')};text-align:{align};overflow-wrap:break-word;"
  if typ in ('text','heading'):
   tag=v.get('headingType','div') if typ=='heading' else 'div'
   inside=f'<{tag} style="margin:0;{common}font-weight:{"700" if typ=="heading" else "400"}">{v["text"]}</{tag}>'
  elif typ=='image': inside=f'<img src="{html.escape(v["src"]["url"],quote=True)}" alt="{html.escape(v["altText"],quote=True)}" width="600" style="display:block;width:100%;max-width:{v["src"].get("maxWidth","100%")};height:auto;border:0;margin:{"0 auto" if align=="center" else "0"}">'
  elif typ=='button': inside=f'<div style="text-align:{align}"><a href="{html.escape(v["href"]["values"]["href"],quote=True)}" style="{common}display:inline-block;padding:{v["padding"]};border-radius:{v["borderRadius"]};background:{v["buttonColors"]["backgroundColor"]};color:{v["buttonColors"]["color"]};font-weight:600;text-decoration:none">{v["text"]}</a></div>'
  elif typ=='divider': inside=f'<div style="border-top:1px solid {v["border"]["borderTopColor"]}"></div>'
  else: raise ValueError(typ)
  return f'<div style="padding:{pad}">{inside}</div>'
 parts=[]
 for r in d['body']['rows']:
  n=len(r['columns']); units=sum(r['cells'])
  content=''.join(f'<div style="display:inline-block;vertical-align:top;width:100%;max-width:{600*r["cells"][i]/units:g}px;text-align:left">'+''.join(render_block(b) for b in c['contents'])+'</div>' for i,c in enumerate(r['columns']))
  parts.append(f'<tr><td style="padding:{r["values"]["padding"]};background:{r["values"]["columnsBackgroundColor"]};font-size:0;text-align:center">{content}</td></tr>')
 return f'<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="color-scheme" content="light"><title>{html.escape(t["name"])}</title></head><body style="margin:0;background:{t["background"]};font-family:Arial,Helvetica,sans-serif"><div style="display:none;max-height:0;overflow:hidden">{html.escape(t["preheader"])}</div><table role="presentation" width="100%" cellspacing="0" cellpadding="0"><tr><td align="center"><table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:600px;margin:0 auto">'+''.join(parts)+'</table></td></tr></table></body></html>'

exec((ROOT / "scripts/reference_templates.py").read_text(), globals())
exec((ROOT / "scripts/composed_template_collection.py").read_text(), globals())

manifest=[]
for t in TEMPLATES:
 d=design_for(t); key=t['key']; folder=OUT/key; folder.mkdir(exist_ok=True)
 (folder/'design.json').write_text(json.dumps(d,indent=2,ensure_ascii=False)+'\n')
 (folder/'preview.html').write_text(preview_for(t,d))
 manifest.append({k:t[k] for k in ('key','name','category','description','subject','preheader','tags','marketing')} | {'collection':t.get('collection','Xem originals'),'source':t.get('source',''),'reference':t.get('reference',''),'editingMode':'blocks'} | {'version':1,'designUrl':f'/assets/template-starters/{key}/design.json','previewUrl':f'/assets/template-starters/{key}/preview.html'})
(ROOT/'lib/template-starters/manifest.json').write_text(json.dumps(manifest,indent=2,ensure_ascii=False)+'\n')
print(f'Generated {len(manifest)} native Unlayer starters.')
