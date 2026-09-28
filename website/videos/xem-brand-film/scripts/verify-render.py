"""Decode every delivered frame and produce inspectable visual audit sheets."""
from pathlib import Path
import json, subprocess, hashlib
import cv2
import numpy as np
from PIL import Image, ImageDraw, ImageFont

P=Path(__file__).resolve().parents[1]
video=P/'renders/xem-brand-film-1080p.mp4'
out=P/'snapshots/final-audit';out.mkdir(parents=True,exist_ok=True)
probe=json.loads(subprocess.check_output(['ffprobe','-v','error','-count_frames','-show_streams','-show_format','-of','json',str(video)]))
v=next(s for s in probe['streams'] if s['codec_type']=='video')
a=next(s for s in probe['streams'] if s['codec_type']=='audio')
assert (v['width'],v['height'])==(1920,1080)
assert v['avg_frame_rate']=='30/1'
assert int(v['nb_read_frames'])==901
scenes=json.loads((P/'scene-plan.json').read_text())['scenes']
cap=cv2.VideoCapture(str(video));frames=[];stats=[]
while True:
    ok,bgr=cap.read()
    if not ok:break
    small=cv2.resize(bgr,(320,180));f=len(frames)
    stats.append({'frame':f,'std':round(float(cv2.cvtColor(small,cv2.COLOR_BGR2GRAY).std()),5),'mean':round(float(small.mean()),3),'hash':hashlib.sha256(small.tobytes()).hexdigest()})
    frames.append(Image.fromarray(cv2.cvtColor(small,cv2.COLOR_BGR2RGB)))
cap.release()
assert len(frames)==901
assert min(x['std'] for x in stats)>.01,'Possible blank output frame'
# Lossy GOP prediction can change decoded pixels very slightly on a held card.
hold=np.stack([np.array(im).astype(float) for im in frames[854:]])
assert float(hold.std(axis=0).mean())<1,'Final logo hold is not visually static'
font=ImageFont.truetype('/System/Library/Fonts/Menlo.ttc',13)
def sheet(indices,name,cols=5,tile=(256,144)):
    tw,th=tile;rows=(len(indices)+cols-1)//cols
    im=Image.new('RGB',(cols*(tw+6)+6,rows*(th+25)+6),'#171a16');d=ImageDraw.Draw(im)
    for k,f in enumerate(indices):
        x=6+(k%cols)*(tw+6);y=6+(k//cols)*(th+25)
        sc=next(s['id'] for s in scenes if s['start']<=f<s['end'])
        d.text((x+3,y+2),f'{f:03d} {f/30:05.2f}s S{sc}',font=font,fill='#ffffef')
        im.paste(frames[f].resize(tile,Image.Resampling.LANCZOS),(x,y+20))
    im.save(out/name,quality=88)
for i in range(12):sheet(list(range(i*80,min(901,(i+1)*80))),f'all-frames-{i+1:02d}.jpg',8,(192,108))
sheet([s['poster'] for s in scenes],'scene-overview.jpg',4,(384,216))
boundaries=sorted(set(f for s in scenes[1:] for f in [s['start']-1,s['start']]))
for i in range(3):sheet(boundaries[i*20:(i+1)*20],f'cuts-{i+1}.jpg',4,(320,180))
audit={'video':v,'audio':a,'decodedFrames':len(frames),'minimumFrameStd':min(x['std'] for x in stats),'finalHoldTemporalStd':float(hold.std(axis=0).mean()),'frameStats':stats,'sha256':hashlib.sha256(video.read_bytes()).hexdigest()}
(P/'reference/final-verification.json').write_text(json.dumps(audit,indent=2))
print(json.dumps({k:v for k,v in audit.items() if k!='frameStats'},indent=2))
