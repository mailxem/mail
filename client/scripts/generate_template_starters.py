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

# The original hand-built starters were retired in favour of the Studio
# collection (scripts/composed_template_collection.py).


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
  elif typ=='button': inside=f'<div style="text-align:{align}"><a href="{html.escape(v["href"]["values"]["href"],quote=True)}" style="{common}display:inline-block;padding:{v["padding"]};border-radius:{v["borderRadius"]};{"border:"+v["border"]["borderTopWidth"]+" solid "+v["border"]["borderTopColor"]+";" if v.get("border",{}).get("borderTopWidth") else ""}background:{v["buttonColors"]["backgroundColor"]};color:{v["buttonColors"]["color"]};font-weight:600;text-decoration:none">{v["text"]}</a></div>'
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
