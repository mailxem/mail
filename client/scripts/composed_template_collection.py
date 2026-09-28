"""Individually briefed compositions, replacing repeated industry × layout reskins.
Each line carries a distinct editorial premise. Compositions combine native blocks
with different lead treatments, column proportions, reading order and density.
"""
from copy import deepcopy

# name | headline | deck. These are editable sample campaigns, never live claims.
BRIEFS = {
'culture': [
('After Hours','The city belongs to the curious.','A midnight cinema, a listening bar, and a gallery that leaves the lights on.'),
('The Small Screen','Short films. Long conversations.','Three independent films that make an ordinary evening feel a little bigger.'),
('Off the Record','Listen all the way through.','An album-side ritual, a record-store conversation, and the joy of hearing something new.'),
('The Reading Room','A book can change the temperature of a room.','Stories for the bedside table, the train ride, and the friend who always asks what to read.'),
('In Good Taste','What makes a place worth returning to?','A neighborhood restaurant, a familiar table, and a menu that follows the season.'),
('On the Wall','Art does not need to wait for an occasion.','A printmaker opens the studio and shows how a sketch becomes an edition.'),
('Second Life','Old things. New conversations.','Repair shops, vintage finds, and people giving beautiful objects another chapter.'),
('Local Frequency','Tune into your own neighborhood.','A small guide to the people and places making things happen close to home.'),
('The Sunday Matinée','Make an afternoon of it.','A double feature, a favorite seat, and a good reason to leave the phone at home.'),
('Soft Focus','Look a little longer.','A photographer finds unexpected stories in everyday streets and quiet corners.'),
('The Index','Ten things worth keeping close.','An edited collection of books, records, objects, and ideas for the coming month.'),
('Open Studio','Come in. The work is still wet.','A behind-the-scenes invitation to meet the artists before the opening night.'),
('Night School','Learn something after the day is done.','Evening classes for curious people: ceramics, letterpress, and a little beginner courage.'),
('The Good Seats','There is nothing quite like being there.','Small venues, live performances, and the people who bring a room to life.'),
('Margin Notes','The best part might be in the margins.','A reader’s dispatch on underlined sentences, unfinished thoughts, and books passed between friends.')],
'food': [
('Market Morning','Cook what looks good today.','A basket-first approach to dinner, with seasonal produce and room to improvise.'),
('The Pantry Letter','Good dinner starts on the shelf.','Beans, bright olive oil, and a few useful staples that make weeknight cooking easier.'),
('Little Ferments','Give it time. Let it change.','A beginner-friendly look at pickles, sourdough, and the small experiments happening on the counter.'),
('The Long Lunch','Clear the afternoon. Set another place.','A generous menu built for passing plates and staying a little longer.'),
('Bitter Sweet','The beautiful balance in a good dessert.','Dark chocolate, toasted nuts, and the pinch of salt that brings everything together.'),
('One Good Pan','Less washing up. More sitting down.','A flexible one-pan supper with crisp edges, seasonal vegetables, and a bright finish.'),
('Coffee Compass','Follow the flavor back to the source.','A tasting guide to origin, process, and the choices behind your morning cup.'),
('The Bread Club','A little flour. A little patience.','Notes from the bakery on a crackling crust, a slower rise, and sharing the first slice.'),
('Fire & Table','Take the kitchen outside.','Cooking over coals, preparing ahead, and making dinner feel like a gathering.'),
('The Citrus Issue','A bright idea for a gray day.','Sharp dressings, fragrant peel, and a simple cake that uses the whole fruit.'),
('Midweek Supper','Dinner does not need a grand plan.','Three reliable meals built around what is already in the fridge.'),
('No Proof Needed','A very good drink, without the alcohol.','Bitter botanicals, bubbles, and a grown-up approach to the afternoon spritz.'),
('Kitchen Notes','The recipe is only the beginning.','A cook’s notebook of useful substitutions, happy accidents, and things worth repeating.'),
('The Tomato Edit','Summer, sliced thick.','A celebration of ripe tomatoes: on toast, in the pot, and straight from the market bag.'),
('Salt & Stories','Every family has a dish like this.','Recipes passed down, changed a little, and carried from one kitchen to another.')],
'nature': [
('Trailhead','Start where the pavement ends.','A short walk, a longer view, and a guide to finding your next nearby trail.'),
('The Tidal Letter','Plan your day around the water.','Rock pools, coastal paths, and the quiet rhythm of a shoreline at low tide.'),
('Birdsong','There is a whole world overhead.','An unhurried introduction to listening, looking, and noticing the birds around you.'),
('The Weather Window','A little rain is part of the story.','Practical layers, flexible plans, and the pleasure of being outdoors in imperfect weather.'),
('Rooted','Grow something you can look forward to.','A small-space planting plan for windowsills, balconies, and first-time gardeners.'),
('Under Canvas','A room with no walls.','Simple camping comforts, a generous packing list, and breakfast in the open air.'),
('The Forest Floor','Look down. There is more going on.','Moss, fungi, fallen leaves, and the tiny systems keeping a woodland alive.'),
('Slow Miles','Go at the pace of a conversation.','A walking route with good stopping points and enough room to get pleasantly lost.'),
('The Night Sky','Make a little room for wonder.','A beginner’s evening of stargazing, from choosing a dark spot to finding familiar shapes.'),
('Waterline','Find your way back to the river.','Riverside trails, a picnic by the bank, and stories from the people caring for the water.'),
('The Pocket Park','Nature does not have to be far away.','Small green spaces, everyday walks, and the value of a bench in the shade.'),
('Wild Weekend','Two days. A different perspective.','A low-fuss escape with a trail, a cabin, and very little on the schedule.'),
('The Seed Exchange','Good things grow when you share them.','Swap seeds, trade advice, and meet the neighbors growing something good.'),
('Leave It Better','A small act belongs in every adventure.','Simple ways to care for a favorite place before you head home.'),
('First Light','Catch the day before it gets busy.','An early walk, a warm flask, and a few reasons to set the alarm just once.')],
'design': [
('Grid & Grain','Order, with a little human texture.','A studio study in quiet grids, tactile materials, and the details that keep work from feeling cold.'),
('Type Matters','Let the letters do the talking.','A typographic field trip through book covers, shop signs, and unexpected pairings.'),
('The Color Study','One color can change the whole conversation.','A close look at a single shade, its surprising neighbors, and where it works best.'),
('Working Model','Make it small enough to learn from.','Sketches, paper prototypes, and the experiments that answer a question before the build.'),
('Form Follows Feeling','Useful can be beautiful, too.','Objects designed around everyday gestures, quiet comfort, and a little delight.'),
('The Material Library','Touch is part of the idea.','A table of stone, paper, cloth, and timber—and the choices each material invites.'),
('Negative Space','Make room for what matters.','A practical study in restraint, breathing room, and a clearer visual hierarchy.'),
('The Rough Cut','Show the work before it feels ready.','A creative team shares the discarded directions that helped a project find its shape.'),
('Small Systems','Consistency begins with a useful rule.','A designer’s guide to building a system that helps people rather than slowing them down.'),
('The Poster Wall','Say one thing. Make it impossible to miss.','A collection of bold announcements, spare typography, and graphic ideas with a point of view.'),
('Useful Objects','A better everyday is in the details.','A lamp, a notebook, and a chair that earn their place by being good at one thing.'),
('The Studio Visit','Every workspace tells a story.','Inside a practice where sketches, prototypes, and a very well-used table share the room.'),
('Print Practice','Ink makes the idea feel different.','An introduction to risograph textures, spot colors, and the happy limits of a print process.'),
('The Design Crit','Ask a more useful question.','How to turn a review into a conversation that helps the next iteration.'),
('Side Project','Give the idea a Saturday.','Small creative experiments for the projects that keep returning to your notebook.')],
'work': [
('The Build Log','A useful release starts with a real problem.','The decisions, tradeoffs, and small wins behind a product improvement.'),
('Human Systems','Make the handoff easier for the next person.','Clear context, useful checklists, and a calmer way to work together.'),
('Small Bets','Try the version you can learn from.','A field guide to experiments that are small enough to run and clear enough to evaluate.'),
('The Decision Room','A good decision leaves a trail.','Write down the context, consider the options, and make the next conversation easier.'),
('Open Source Sunday','Build something others can build on.','A maintainer’s view of welcoming contributions and keeping a project understandable.'),
('The Product Desk','Listen before adding another feature.','A practical edition on customer conversations, support signals, and finding the problem underneath.'),
('Focus Window','Protect a little uninterrupted time.','A working-day experiment in fewer pings, clearer priorities, and finishing one thing.'),
('The Release Train','Small improvements. A steady rhythm.','A product roundup built around what changed, why it matters, and what to try next.'),
('Better Questions','The brief is a place to get curious.','Five prompts that help a team move from assumptions to shared understanding.'),
('The Research Shelf','Keep the evidence close to the work.','A lightweight approach to collecting insights without creating another forgotten repository.'),
('Debug Diary','The bug was an invitation to understand.','A behind-the-scenes story about tracing a problem and making the system easier to reason about.'),
('The Practical AI','Start with a task you can check.','A grounded guide to experimenting with assistants while keeping review in the workflow.'),
('Team Rituals','A little structure can make more room.','Useful weekly habits for distributed teams that want fewer meetings and better context.'),
('The Changelog Letter','Here is what your feedback changed.','An honest product note connecting recent improvements to the people who asked for them.'),
('Build in Public','Let people see the middle.','A founder’s dispatch on prototypes, decisions, and sharing progress without a polished ending.')],
'places': [
('The Postcard Club','Wish you were here, slowly.','A local guide to a place best explored one street and one long lunch at a time.'),
('The Window Seat','The journey deserves a chapter.','Rail routes, passing landscapes, and good reasons to take the longer way.'),
('Somewhere Small','A smaller town can hold a bigger weekend.','Independent shops, a riverside walk, and a place to stay with a story.'),
('The Architecture Walk','Read a city through its buildings.','Doorways, courtyards, and quiet details hiding in the route you usually rush through.'),
('Out of Office','Put a little distance between you and the routine.','A thoughtful escape with flexible plans and a few very good recommendations.'),
('The Island Letter','Let the water set the pace.','A ferry crossing, a coastal path, and a small table overlooking the harbor.'),
('Room Service','Stay somewhere that feels like a place.','Independent hotels, thoughtful rooms, and hosts who know where breakfast should happen.'),
('The Urban Rambler','Turn left where you usually turn right.','An afternoon route through hidden gardens, bookshops, and streets worth noticing.'),
('Cabin Season','A simpler room. A wider view.','Wood fires, walks from the door, and the practical art of packing less.'),
('The Local Table','Taste your way into a place.','Market stalls, family kitchens, and the stories a menu can tell.'),
('Beyond the Guidebook','Ask someone who lives there.','A resident’s favorites, from an early coffee to the last good view of the evening.'),
('Carry On','Bring less. Leave room for the unexpected.','A considered packing list and a few habits that make travel easier.'),
('The Long Weekend','Three days can feel like a new chapter.','A loose itinerary with one good anchor for each day and time left open.'),
('The Return Ticket','Some places are worth seeing twice.','A second visit, a different season, and everything you missed the first time.'),
('No Fixed Agenda','Follow the interesting street.','An invitation to explore with a short list, comfortable shoes, and room to change your mind.')],
'learning': [
('The Notebook Club','Learn it by making it.','A short project, a useful constraint, and a place to put your new skill to work.'),
('Tiny Lessons','One idea, well understood.','A five-minute lesson with an example you can try before the day gets away.'),
('The Practice Room','Progress likes a regular appointment.','A gentle practice plan built around small repetitions and realistic expectations.'),
('Curiosity Cabinet','Keep a place for the interesting things.','Questions, observations, and small discoveries collected through the week.'),
('The Explainer','A complicated idea, in useful pieces.','A clear introduction that starts with something familiar and builds from there.'),
('Learn by Doing','The first attempt is part of the lesson.','An invitation to try, notice, adjust, and make something a little better.'),
('The Reading List','Follow a question through three books.','A focused reading path for a topic worth understanding from more than one angle.'),
('Study Hall','Good company makes a hard thing easier.','A shared learning session with clear goals, a quiet hour, and time to compare notes.'),
('The Field Guide','Take the lesson out into the world.','A practical observation exercise for noticing patterns in the things around you.'),
('Unfinished Questions','Not knowing is a useful place to start.','A collection of open questions and the experiments that could help answer them.'),
('The Mentor Note','A little perspective for the next step.','Advice from someone who remembers what the beginning felt like.'),
('The Skill Swap','Teach one thing. Learn another.','An invitation to exchange practical knowledge with people in your community.'),
('Five Minute Workshop','Make something before you overthink it.','A small creative prompt with just enough structure to help you begin.'),
('The Learning Loop','Look back before moving on.','A short reflection on what worked, what surprised you, and what you want to try next.'),
('Begin Again','You can be new at something at any age.','A warm welcome to the awkward, interesting, rewarding part of learning.')],
'community': [
('The Common Room','Pull up a chair. Bring your story.','A monthly gathering of member projects, generous ideas, and ways to get involved.'),
('Neighbor Notes','Good things are happening nearby.','Meet the people turning a spare hour and a shared idea into something useful.'),
('The Volunteer Dispatch','An hour can change someone’s afternoon.','Practical ways to contribute, with clear roles and a warm welcome for first-timers.'),
('Together Again','The best part is who turns up.','An invitation to reconnect, share what you are making, and meet a few new faces.'),
('The Member Spotlight','There is a person behind every good project.','A conversation about starting small, asking for help, and making something that matters.'),
('The Giving Circle','A little from many goes a long way.','An update on the work your community supports and the next useful contribution.'),
('The Noticeboard','Something to join. Something to share.','Local dates, member requests, and small opportunities that deserve a little space.'),
('The Welcome Table','You do not need to know anyone yet.','A newcomer’s guide to showing up, finding your people, and feeling at home.'),
('Common Ground','Start with what you share.','A community conversation about listening carefully and building something together.'),
('The Makers Market','Meet the hands behind the things.','A weekend invitation to discover local makers and the stories behind their work.'),
('Small Celebrations','Some good news deserves a whole email.','A collection of member milestones, first attempts, and quiet achievements.'),
('The Exchange','What you know could help someone else.','A skills-and-resources edition built around useful offers and thoughtful requests.'),
('The Community Map','Find your next point of connection.','A guide to groups, gatherings, and places where a shared interest becomes a friendship.'),
('Open Invitation','There is room for you in this idea.','A project needs a few more hands, a fresh perspective, and people willing to begin.'),
('The Thank You Note','This happened because people showed up.','A warm look back at a shared effort and the people who made it possible.')],
'style': [
('The Considered Edit','Fewer things. Better reasons.','An edited wardrobe built around good materials, everyday use, and your own point of view.'),
('Wear & Repeat','The best outfit earns another day.','Simple combinations, useful layers, and clothes that feel more like you over time.'),
('The Texture Issue','Let the material set the mood.','Linen, brushed wool, and the small details that make an everyday piece feel good.'),
('Objects of Note','A few things with a point of view.','Thoughtful finds for your desk, your home, and the daily rituals in between.'),
('The Maker’s Label','Know the story behind the seam.','A studio visit with the people cutting, stitching, and refining a small collection.'),
('Quiet Luxury Letter','Good quality does not need to shout.','A closer look at construction, proportion, and the choices that help a piece last.'),
('The Color Wardrobe','Try the shade you keep looking at.','A practical guide to adding a little color without replacing everything you own.'),
('Everyday Uniform','Make getting dressed a smaller decision.','A useful foundation of pieces that work together and leave room for personality.'),
('The Repair Edit','A little care keeps the story going.','Simple repairs, better storage, and ways to look after the things you already love.'),
('A Good Pair','Find the thing that goes with everything.','Versatile companions for a favorite piece, chosen for comfort and everyday usefulness.'),
('The Weekend Bag','Pack for the day you want to have.','A small selection of essentials for a market morning, a long walk, and a late lunch.'),
('New Season Notes','A fresh feeling, without starting over.','A considered seasonal edit built around adding one or two useful pieces.'),
('The Vintage Find','Some things get more interesting with time.','A guide to discovering, checking, and styling pieces with a previous chapter.'),
('Made to Keep','Buy it for the years, not the moment.','A maker’s perspective on durable materials, repairable design, and lasting usefulness.'),
('The Detail Study','The difference is in the small things.','Buttons, hems, handles, and finishes that reveal how much care went into an object.')],
'wellbeing': [
('The Slow Start','Begin the day at a human pace.','A few calm morning rituals to try, adapt, or leave behind according to what works for you.'),
('Room to Breathe','Put a little space between the things.','A gentle reminder to pause, step outside, and make the next hour more manageable.'),
('The Walking Club','A conversation with a little movement.','A simple invitation to take the long way with good company and no performance target.'),
('Rest Notes','Recovery belongs on the calendar, too.','Practical ideas for making room to rest in the middle of a full week.'),
('The Good Enough Day','An ordinary day can still be a good one.','A little permission to choose a realistic plan and notice what is already working.'),
('Small Rituals','Care can look like something very simple.','A proper lunch, a favorite song, a cleared corner: small actions that change how a day feels.'),
('Outside for Ten','The next break could have a view.','Easy ways to spend a little more time outdoors, even when the schedule is crowded.'),
('The Evening Edit','Give the day a softer landing.','A loose evening routine with fewer demands and a little time to come back to yourself.'),
('A Little Balance','Make room for more than getting things done.','A weekly reflection on energy, commitments, and the things that help you feel grounded.'),
('The Personal Best','Define progress in your own terms.','A note on choosing goals that fit your life and leaving space to change your mind.'),
('The Quiet Corner','Create one place that asks less of you.','A small home project in light, comfort, and keeping a few favorite things close.'),
('Move Your Way','Find a kind of movement you look forward to.','An invitation to explore walking, stretching, dancing, or anything that feels right for you.'),
('The Digital Pause','You can come back to it later.','A practical experiment in notifications, boundaries, and reclaiming a little attention.'),
('Good Company','Sometimes the useful thing is a conversation.','A reminder to make time for the people who leave you feeling more like yourself.'),
('The Reset Letter','Start again from where you are.','A simple plan for the week ahead, built around one manageable change.')],
'science': [
('The Curious Planet','A small world with big surprises.','A nature-led edition about remarkable adaptations and the questions they invite.'),
('Lab Notes','Every discovery starts with a question.','A look at how a simple observation becomes an experiment worth trying.'),
('The Ocean Issue','Most of the story is beneath the surface.','A beginner-friendly dive into coastal ecosystems and the people studying them.'),
('Animal Fact Club','Small creature. Extraordinary trick.','Meet the pistol shrimp and explore the surprising physics of life underwater.'),
('The Sky Report','There is always more to look up for.','A practical guide to a month of skywatching and the science behind what you see.'),
('Everyday Physics','The ordinary world is full of experiments.','A few familiar objects, a clear question, and a different way to look at your surroundings.'),
('The Pattern Finder','Nature has a few favorite ideas.','Spirals, branches, and the useful patterns that appear across very different scales.'),
('The Climate Notebook','Understand the system one piece at a time.','A clear introduction to a climate concept, with useful context and places to learn more.'),
('The Tiny World','Look closer than you usually would.','An introduction to the small organisms and structures hiding in everyday places.'),
('Ask a Scientist','A good question is worth sending.','A reader-led edition where a researcher explains how they think about an interesting problem.'),
('The Science Shelf','Three reads for a curious mind.','Books and explainers that make a complex topic welcoming without making it simplistic.'),
('The Experiment Club','Try it. Observe it. Write it down.','A simple, low-risk observation exercise using things already around you.'),
('Deep Time','The landscape remembers more than we do.','An approachable look at the layers, fossils, and processes that shape a place.'),
('The Research Roundup','A closer look at the question behind the headline.','A digest focused on what a study asks, how it works, and what remains uncertain.'),
('Wonder Works','Stay interested in the things you cannot explain yet.','A collection of surprising observations and good places to begin understanding them.')],
'home': [
('The Homebody','Make staying in feel like a choice.','A warm edit of books, small projects, and ways to enjoy the space you already have.'),
('Room by Room','One corner is a good place to begin.','A small, achievable home refresh built around light, layout, and useful storage.'),
('The Shelf Life','Keep the things that tell your story.','Books, found objects, and a thoughtful approach to arranging a shelf without making it precious.'),
('Light Study','Follow the light through the day.','A practical guide to noticing natural light and making a room work with it.'),
('The Small Space','A smaller footprint can hold a full life.','Flexible furniture, clear surfaces, and useful ideas for rooms that do more than one job.'),
('Made at Home','Give your hands something to figure out.','A small weekend project with simple materials and room for a personal touch.'),
('The Hosting Note','Make people comfortable. The rest can be simple.','An easy plan for having friends over without turning the evening into a production.'),
('Soft Landing','Home can be the quieter part of the day.','Warm textures, good lighting, and a few changes that help a space feel welcoming.'),
('The Plant Shelf','A little green changes the room.','A beginner’s guide to choosing plants for the light and time you actually have.'),
('The Useful Drawer','Find a home for the everyday things.','A small organizing project that starts with what you reach for most often.'),
('House Notes','Good homes are always a work in progress.','A monthly notebook of practical fixes, useful discoveries, and ideas to try when there is time.'),
('The Window Box','A garden can start at the window.','A compact planting plan for color, herbs, and the pleasure of watching something grow.'),
('The Sunday Reset','A little preparation for a gentler Monday.','A realistic home routine that leaves most of the weekend for living.'),
('Old House Journal','Keep the character. Improve the everyday.','A thoughtful approach to caring for older spaces and working with what is already there.'),
('The Gathering Place','A room is better with people in it.','A layout and hosting edit for making conversation feel easy.')],
}

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
SUBJECT_PHOTOS = {
 'culture':['city','library','vinyl','books','table','art','chair','street','event','city','books','art','writing','event','library'],
 'food':['vegetables','berries','bread','table','berries','salad','coffee','bread','table','citrus','salad','citrus','kitchen','vegetables','kitchen'],
 'nature':['walk','coast','forest','mist','plants','mountain','forest','meadow','stars','lake','flowers','mountain','plants','deer','mist'],
 'design':['building','books','art','desk','chair','room','interior','art','screen','building','chair','desk','art','meeting','writing'],
 'work':['laptop','team','desk','meeting','screen','laptop','writing','screen','meeting','journal','laptop','screen','team','desk','writing'],
 'places':['street','train','building','building','travel','sea','room','street','mountain','table','city','travel','coast','lake','street'],
 'learning':['journal','books','writing','art','library','desk','books','team','forest','journal','writing','meeting','art','journal','library'],
 'community':['team','street','volunteer','event','writing','volunteer','city','table','meeting','art','event','team','street','volunteer','flowers'],
 'style':['clothes','fashion','fabric','chair','fabric','fashion','clothes','fashion','fabric','shoes','travel','clothes','fabric','chair','shoes'],
 'wellbeing':['coffee','sea','walk','room','flowers','journal','meadow','interior','stretch','walk','room','stretch','books','team','journal'],
 'science':['deer','desk','sea','coast','stars','art','flowers','meadow','plants','writing','books','journal','mountain','screen','forest'],
 'home':['room','interior','books','room','chair','art','table','room','plants','kitchen','interior','flowers','kitchen','building','table'],
}
from colorsys import rgb_to_hls,hls_to_rgb

def edition_palette(group,index):
 ink,paper,accent,photo,font=ART[group]
 def shift(value,amount,lightness=None):
  rgb=[int(value[i:i+2],16)/255 for i in (1,3,5)]
  h,l,s=rgb_to_hls(*rgb)
  rgb=hls_to_rgb((h+amount)%1,l if lightness is None else lightness,s)
  return '#'+''.join(f'{round(c*255):02X}' for c in rgb)
 # A bespoke palette per composition; contrast remains stable as hues move.
 angle=[0,.06,-.08,.16,-.15,.28,-.24,.38,-.34,.46,-.43,.11,-.19,.32,-.04][index]
 return shift(ink,angle),shift(paper,angle,0.95-(index%4)*.018),shift(accent,angle,0.72),SUBJECT_PHOTOS[group][index],font

# Contrasting art direction per subject; variations below also change layout and type.
ART = {
 'culture':('#202B40','#FAF6ED','#EF563A','interior',SERIF),
 'food':('#49251E','#FFF3DD','#ED643A','coffee',SERIF),
 'nature':('#203F2C','#EEF0DE','#D3EC62','forest',SERIF),
 'design':('#251933','#F8E6F1','#DDFF48','interior',SANS),
 'work':('#101E37','#EDF2FA','#76A4FF','work',SANS),
 'places':('#12434E','#EAF2EA','#F6AB69','mountain',SERIF),
 'learning':('#342654','#F2EBFA','#C1ACF2','work',MONO),
 'community':('#632F2E','#FFF0DE','#F4BD43','event',SANS),
 'style':('#332C27','#EDE7DB','#AA7860','fashion',SERIF),
 'wellbeing':('#344A42','#F1F2E8','#BBD8B5','forest',SERIF),
 'science':('#082F4B','#E1F4F4','#FFBC42','mountain',MONO),
 'home':('#4D3B32','#F8EADB','#B7BD93','interior',SERIF),
}

# Every recipe names its own reading order. Numeric prefix selects a lead layout;
# the suffix composes the middle and ending without a repeated family skeleton.
RECIPES = [
'0 letter pair pull links','1 feature checklist reply','2 brief quote photo cards','3 photo letter numbered','4 split pull checklist cta','5 index feature quote','6 poster pair links','7 tiles letter reply','8 quote photo numbered cta','9 agenda pair pull','10 sidebar cards links','11 cover checklist reply','12 letter feature pull cta','13 tiles brief links','14 index photo reply',
'15 recipe pull pair','16 letter cards photo','17 feature checklist quote','18 photo recipe links','19 split letter pull','20 index numbered cta','21 poster recipe reply','22 tiles quote feature','23 quote checklist cards','24 agenda letter photo','25 sidebar recipe pull','26 cover pair links','27 letter numbered quote','28 tiles recipe cta','29 index feature reply',
'30 field pull cards','31 feature photo checklist','32 brief field reply','33 photo pair numbered','34 split quote field','35 index checklist photo','0 poster field links','2 tiles pull reply','4 quote field pair','6 agenda photo numbered','8 sidebar letter cards','10 cover field checklist','12 letter quote photo','14 tiles checklist links','16 index field cta',
'18 manifesto pair links','20 feature quote cards','22 brief poster reply','24 photo manifesto numbered','26 split checklist pull','28 index quote cta','30 poster letter cards','32 tiles feature links','34 quote manifesto pair','1 agenda checklist photo','3 sidebar pull cta','5 cover manifesto reply','7 letter poster cards','9 tiles quote numbered','11 index manifesto photo',
'13 changelog pair pull','15 feature numbered reply','17 brief checklist links','19 photo changelog cards','21 split quote numbered','23 index feature checklist','25 poster changelog reply','27 tiles checklist pull','29 quote numbered links','31 agenda changelog photo','33 sidebar feature reply','35 cover checklist cards','2 letter changelog cta','5 tiles numbered photo','8 index checklist pull',
'11 itinerary pair quote','14 feature itinerary links','17 brief photo numbered','20 photo itinerary reply','23 split letter cards','26 index pull cta','29 poster itinerary photo','32 tiles checklist quote','35 quote itinerary pair','3 agenda feature links','6 sidebar itinerary reply','9 cover pull cards','12 letter itinerary checklist','15 tiles photo links','18 index itinerary quote',
'21 lesson pair pull','24 feature lesson reply','27 brief quote links','30 photo lesson cards','33 split numbered cta','1 index checklist pull','4 poster lesson photo','7 tiles letter numbered','10 quote lesson reply','13 agenda checklist links','16 sidebar lesson cards','19 cover feature quote','22 letter lesson cta','25 tiles pull checklist','28 index lesson photo',
'31 people pair links','34 feature people quote','0 brief checklist photo','3 photo people reply','6 split pull numbered','9 index cards cta','12 poster people links','15 tiles quote checklist','18 quote people photo','21 agenda letter cards','24 sidebar people pull','27 cover numbered reply','30 letter people checklist','33 tiles brief cta','2 index people quote',
'5 edit pair pull','8 feature edit reply','11 brief photo cards','14 photo edit links','17 split quote checklist','20 index numbered pull','23 poster edit cta','26 tiles letter quote','29 quote edit cards','32 agenda photo links','35 sidebar edit reply','4 cover checklist pull','7 letter edit photo','10 tiles numbered cta','13 index edit quote',
'16 ritual pair links','19 feature ritual pull','22 brief quote checklist','25 photo ritual reply','28 split cards cta','31 index letter quote','34 poster ritual photo','0 tiles checklist pull','3 quote ritual numbered','6 agenda letter links','9 sidebar ritual cards','12 cover photo reply','15 letter ritual checklist','18 tiles pull cta','21 index ritual quote',
'24 question pair photo','27 feature question links','30 brief numbered reply','33 photo question cards','1 split checklist quote','4 index pull cta','7 poster question links','10 tiles letter photo','13 quote question checklist','16 agenda numbered cards','19 sidebar question reply','22 cover photo pull','25 letter question cta','28 tiles checklist quote','31 index question cards',
'34 room pair pull','0 feature room reply','3 brief photo links','6 photo room numbered','9 split quote cards','12 index checklist cta','15 poster room links','18 tiles letter pull','21 quote room photo','24 agenda checklist numbered','27 sidebar room reply','30 cover feature cards','33 letter room quote','2 tiles checklist photo','5 index room cta',
]
assert len(RECIPES)==180

def weighted(cols, weights, bg, pad='12px 0px'):
 return {'columns':cols,'cells':weights,'bg':bg,'pad':pad}

def picture(key, shape='wide', pad='0px'):
 b=deepcopy(image(key)); b['values']['containerPadding']=pad
 shapes={'wide':(1200,560),'square':(700,700),'portrait':(600,900),'strip':(1200,280)}
 w,h=shapes[shape]; b['values']['src'].update(width=w,height=h)
 b['values']['src']['url']=b['values']['src']['url'].replace('w=1200&h=660',f'w={w}&h={h}')
 return b

def compose(brief, group, idx, recipe):
 name,headline,deck=brief; ink,paper,accent,photo,font=edition_palette(group,idx)
 # Adjacent editions use different paper/ink roles, type scales, and density.
 serif = font if idx%3 else SANS
 p=20+4*(idx%5); pad=f'14px {p}px'; align=['left','center','right'][idx%3]
 h=lambda s=headline,**kw:title(s,**{'size':46,'font':serif,'color':ink,'pad':pad,**kw})
 t=lambda s=deck,**kw:text(s,**{'size':16,'color':ink,'pad':pad,**kw})
 k=lambda s=name.upper(),**kw:label(s,**{'color':ink,'pad':pad,**kw})
 pic=lambda shape='wide',padding='0px':picture(photo,shape,padding)
 brand=lambda:row(k(),bg=paper,pad='6px 0px')
 cta=lambda s='Read this edition →':button(s,ink,align='left',ink=paper)
 n, *modules=recipe.split(); n=int(n)
 # Thirty-six genuinely different lead treatments, from folios to posters to grids.
 # Individual layouts continue with a unique sequence of content sections below.
 openings=[
 lambda:[brand(),row(h(),t(),bg=paper),row(pic('strip'))],
 lambda:[row(pic()),weighted([[k(),h(size=34)],[t(),cta()]],[1,1],paper)],
 lambda:[row(k(),bg=ink,pad='0px'),row(h(size=62,align='center'),bg=paper),row(t(align='center'),bg=accent)],
 lambda:[weighted([[pic('portrait')],[k(),h(size=32),t(size=14)]],[1,1],paper)],
 lambda:[weighted([[k(),h(size=40),cta()],[pic('portrait')]],[2,1],paper)],
 lambda:[row(k(),divider(ink),h(size=37),bg=paper),weighted([[label('IN THIS ISSUE',ink),t('01 / The story<br>02 / A closer look<br>03 / Try it yourself',size=13)],[pic('square')]],[1,2],paper)],
 lambda:[row(k(),h(size=68,color=paper),t(color=paper),bg=ink,pad='32px 0px'),row(label('AN EDITION FOR CURIOUS PEOPLE',ink,align='center'),bg=accent)],
 lambda:[weighted([[pic('square')],[k(),h(size=29)],[pic('square')]],[1,2,1],paper),row(t(),bg=paper)],
 lambda:[row(k(),title('“',size=110,color=ink,pad='0px 30px'),h(size=40,font=SERIF),bg=paper),row(pic('strip'))],
 lambda:[weighted([[title(f'{idx+1:02}',size=90,color=ink),label('THE EDITION',ink)],[k(),h(size=33),t(size=14)]],[1,2],paper)],
 lambda:[weighted([[k(),divider(ink),label('A SMALL<br>INDEPENDENT<br>DISPATCH',ink)],[h(size=42),pic('wide'),t()]],[1,3],paper)],
 lambda:[row(k(align='center'),pic(),bg=paper),row(h(size=42,align='center'),t(align='center'),bg=accent)],
 lambda:[row(k(),t('Dear {{first_name}},',font=SERIF,size=21),h(size=36,font=SERIF),t(font=SERIF),divider(ink),bg=paper)],
 lambda:[row(h(size=55,align='right'),bg=paper),weighted([[pic('square')],[k(),t(size=14),cta()]],[2,1],paper)],
 lambda:[row(k(),h(size=34),divider(ink),bg=paper),weighted([[title('01',size=60,color=ink)],[t(),pic('strip')]],[1,3],paper)],
 lambda:[row(k(align='center'),h(size=51,font=SERIF,align='center'),bg=paper),weighted([[pic('portrait')],[label('FROM THE NOTEBOOK',ink),t(),cta('Open the notebook →')]],[3,2],paper)],
 lambda:[row(pic('strip')),row(k(),h(size=52),t(),bg=ink),row(label('TAKE YOUR TIME WITH THIS ONE',ink,align='center'),bg=accent)],
 lambda:[weighted([[k(),h(size=38),t(size=14)],[title('↗',size=110,color=ink,align='center'),pic('square')]],[2,1],paper)],
 lambda:[row(k(),bg=paper),weighted([[pic('square')],[pic('portrait')]],[2,1],paper),row(h(size=40),t(),bg=paper)],
 lambda:[row(label('THE SHORT VERSION',ink,align='center'),bg=accent),row(k(),h(size=50,font=MONO),t(),bg=paper)],
 lambda:[row(k(),h(size=31),bg=paper),weighted([[title('01',size=52,color=ink),t(deck.split(',')[0],size=14)],[title('02',size=52,color=ink),t('A different perspective',size=14)],[title('03',size=52,color=ink),t('Something to try',size=14)]],[1,1,1],accent)],
 lambda:[row(k(),h(size=72,align='center'),divider(ink),t(size=19,align='center'),bg=paper,pad='28px 0px')],
 lambda:[weighted([[k(),h(size=30)],[pic('square')]],[1,1],paper),weighted([[pic('strip')],[t(size=14)]],[1,2],accent)],
 lambda:[row(k(),bg=paper),row(h(size=40,font=SERIF,align='center'),t(size=19,align='center'),bg=accent),row(pic('strip'))],
 lambda:[row(k(),divider(ink),bg=paper),weighted([[h(size=36),t(size=14)],[label('YOUR NEXT<br>GOOD READ',ink),divider(ink),cta()]],[3,2],paper)],
 lambda:[weighted([[pic('portrait')],[k(),h(size=28),divider(ink),t(size=14)]],[2,3],paper)],
 lambda:[row(k(align='right'),h(size=58,align='right'),bg=paper),weighted([[t()],[pic('square')]],[2,1],paper)],
 lambda:[row(k(),t(size=21,font=SERIF),divider(ink),h(size=48),bg=paper)],
 lambda:[row(pic('strip')),weighted([[h(size=38)],[k(),t(size=14)]],[3,2],accent)],
 lambda:[row(k(),h(size=42,font=MONO),bg=paper),weighted([[pic('square')],[t(size=15),cta()]],[1,1],paper)],
 lambda:[weighted([[k(),h(size=31),t(size=14)],[pic('portrait')],[label('FIELD<br>NOTES',ink),title(f'{idx+1:02}',size=46,color=ink)]],[3,2,1],paper)],
 lambda:[row(k(),h(size=55),bg=accent),row(t(),divider(ink),pic('strip'),bg=paper)],
 lambda:[row(k(),h(size=36),bg=paper),weighted([[t(size=14)],[pic('square')],[t('A small discovery. A fresh perspective. A good place to begin.',size=14)]],[1,1,1],paper)],
 lambda:[row(k(align='center'),pic('strip'),bg=ink),row(h(size=52,font=SERIF),t(),bg=paper)],
 lambda:[weighted([[title('✳',size=105,color=ink),k()],[h(size=37),t(size=14),cta()]],[1,3],accent)],
 lambda:[row(k(),bg=paper),weighted([[h(size=36),t(size=14)],[pic('square'),label('LOOK A LITTLE CLOSER',ink)]],[3,2],paper),row(divider(ink),bg=paper)],
 ]
 rows=openings[n]()
 # Fix foreground when a lead deliberately reverses paper and ink.
 for rr in rows:
  if rr['bg']==ink:
   for cc in rr['columns']:
    for b in cc:
     if b['type'] in ('text','heading'): b['values']['color']=paper
 # Each edition's secondary copy remains tied to its own premise.
 detail=f'Welcome to {name}. This week, we’re following the smaller details behind the story: the choices that shaped it, the people who care about it, and the ideas you can bring into your own day.'
 question=f'What would {name.lower()} look like in your everyday life? Start with the part of this story that caught your attention.'
 for position,mod in enumerate(modules):
  bg=paper if position%2==0 else '#FFFFFF'
  if mod in ('split','tiles','sidebar'):
   rows.append(weighted([[pic('square','12px 20px')],[h('An inside look',size=27),t(detail,size=14)]],[1,2 if mod=='sidebar' else 1],bg))
  elif mod in ('index','cover','agenda'):
   rows.append(row(k('INSIDE '+name.upper()),h('Your guide to this edition',size=30),t(deck),divider(ink),cta('Explore the edition →'),bg=bg))
  elif mod=='letter': rows.append(row(k('A NOTE FROM THE EDITOR'),t('Hi {{first_name}},<br><br>'+detail,font=SERIF),t('Thanks for making a little room for this.<br><strong>The editor</strong>',font=SERIF),bg=bg))
  elif mod in ('feature','brief'): rows.append(row(k('THE STORY / '+name.upper()),h('The story behind '+name.lower(),size=29),t(detail),cta('Read the full story →'),bg=bg))
  elif mod=='pair': rows.append(weighted([[h('Look a little closer.',size=24),t(deck,size=14)],[h('Bring it into your week.',size=24),t(question,size=14)]],[2+(idx%2),2],bg))
  elif mod in ('pull','quote'): rows.append(row(title(deck.split(',')[0]+'.',size=34,font=SERIF,color=ink,align=align,pad='36px 40px'),label('A THOUGHT FROM '+name.upper(),ink,align=align),bg=accent))
  elif mod=='photo': rows.append(row(pic(['wide','strip','square'][idx%3]),label(name.upper()+' / A CLOSER LOOK',ink),bg=bg))
  elif mod in ('links','numbered'): rows.append(row(k('THREE THINGS TO TAKE AWAY'),h('01 / Notice the details.',size=22),t(deck,size=14),divider(ink),h('02 / Follow your curiosity.',size=22),t('Explore the people, places, and ideas behind '+name.lower()+'.',size=14),divider(ink),h('03 / Make it your own.',size=22),t(question,size=14),bg=bg))
  elif mod=='checklist': rows.append(row(k('YOUR SMALL NEXT STEP'),t('✓ Choose one idea from this edition.<br><br>✓ Try it in a way that fits your day.<br><br>✓ Notice what changes—and what you want to explore next.'),cta('Keep this edition →'),bg=bg))
  elif mod=='cards': rows.append(weighted([[pic('square','10px 18px'),h('The closer look',size=22),t(deck,size=14)],[h('A useful question',size=25),t(question,size=14),cta('Explore →')]],[1,1+(idx%2)],bg))
  elif mod=='reply': rows.append(row(h('What caught your attention?',size=27),t('Reply with a thought, a question, or a recommendation for a future edition of '+name+'. We read every note.'),bg=bg))
  elif mod=='cta': rows.append(row(h('Take the next small step.',size=28,color=paper),text(question,color=paper),button('Explore '+name,accent,ink=ink),bg=ink))
  else:
   sections={
    'recipe':('FROM THE KITCHEN','Gather your ingredients. Read through once. Leave a little room for taste and instinct.'),
    'field':('IN THE FIELD','Choose a nearby place. Slow your pace. Write down three things you would normally walk past.'),
    'manifesto':('A WORKING PRINCIPLE','Make a clear choice. Give it room. Keep the details that help the idea and question the ones that do not.'),
    'poster':('PIN THIS UP','A little reminder for the next time you need a fresh starting point.'),
    'changelog':('THE USEFUL CHANGE','What changed · Why it matters · How to try it. Keep the explanation close to the work.'),
    'itinerary':('A LOOSE PLAN','Morning / Start somewhere local.<br>Afternoon / Follow the interesting route.<br>Evening / Leave time for a long conversation.'),
    'lesson':('TRY THIS EXERCISE','Pick one observation. Turn it into a question. Try a small experiment and write down what you learn.'),
    'people':('MEET THE PEOPLE','The most useful part of a good idea is often the person willing to share how they made it happen.'),
    'edit':('THE CONSIDERED CHOICE','Think about how you will use it, what it is made from, and whether you will still reach for it next year.'),
    'ritual':('A SMALL RITUAL','Choose a moment you can make your own. Keep it simple enough to repeat and flexible enough to fit your life.'),
    'question':('THE QUESTION TO KEEP','What do we know? How do we know it? What would help us understand a little more?'),
    'room':('START WITH ONE CORNER','Notice the light. Clear what is not helping. Bring back the few things that make the space useful and welcoming.')}
   kicker,body=sections[mod]
   rows.append(weighted([[k(kicker),h(kicker.capitalize()+': '+name,size=26)],[t(body),divider(ink),t(question,size=14)]],[1+(idx%2),2],bg))
 return rows

for group_index,(group,briefs) in enumerate(BRIEFS.items()):
 for idx,brief in enumerate(briefs):
  recipe=RECIPES[group_index*15+idx]
  name,headline,deck=brief
  key=''.join(c if c.isalnum() else '-' for c in name.lower()).strip('-')
  while '--' in key:key=key.replace('--','-')
  key='edition-'+key
  add(key,name,'Newsletters',deck,headline,deck,compose(brief,group,idx,recipe),bg=edition_palette(group,idx)[1],tags=[group,'editorial','composed',name.lower()])
  TEMPLATES[-1]['collection']='Editorial collection'
