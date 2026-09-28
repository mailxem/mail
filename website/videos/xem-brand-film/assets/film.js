/* Xem / 901 frames. All poses are functions of the absolute source frame.
 * Drawing coordinates are 1280x720; canvases rasterize at 1920x1080.
 * The two registry effects retain the installed path rig / WebGL shader.
 */
window.buildXemFilm = async function () {
  'use strict';
  const C = {cream:'#ffffef',ink:'#22251f',iris:'#5b3cc4',lemon:'#edf09b',lavender:'#e7d8fa',forest:'#164b3f'};
  const W=1280,H=720, FPS=30;
  const scenes=window.XEM_SCENES;
  const clamp=(v,a=0,b=1)=>Math.min(b,Math.max(a,v));
  const mix=(a,b,t)=>a+(b-a)*t;
  const smooth=t=>{t=clamp(t);return t*t*(3-2*t);};
  const out=t=>1-Math.pow(1-clamp(t),3);
  const span=(f,a,b)=>clamp((f-a)/(b-a));
  const rnd=n=>{const v=Math.sin(n*127.1+311.7)*43758.5453;return v-Math.floor(v);};
  const TAU=Math.PI*2;
  const templates=['figma-show-work','reading-good-for-you','most-wanted','greenhouse-welcome','unleash-the-fizz','meet-the-leaders','connect-your-tools','webinar-replay','raise-the-standard','softr-product-update','summer-family-journal','your-next-teammate'];
  const sources={mark:'assets/brand/xem-mark-alpha.png',site:'assets/site-scroll.png',cap:'assets/merch/cap-plate.png',shirt:'assets/merch/shirt-plate.png'};
  for(const s of templates)sources[s]=`assets/templates/${s}.webp`;
  for(const s of ['automations','crm','forms','editor-figma','editor-reading','editor-most-wanted','inbox','newsletters'])sources[s]=`assets/product/${s}.webp`;
  const imgs={};
  await Promise.all(Object.entries(sources).map(async([k,url])=>{const im=new Image();im.src=url;await im.decode();imgs[k]=im;}));
  await Promise.all([document.fonts.load('400 100px "DM Sans"'),document.fonts.load('500 100px "DM Sans"'),document.fonts.load('600 100px "DM Sans"'),document.fonts.load('400 100px "EB Garamond"')]);
  await document.fonts.ready;
  const tintCache={};
  const hCanvas=document.createElement('canvas');hCanvas.width=1920;hCanvas.height=1080;
  const halftone=window.createXemHalftone(hCanvas);
  if(!halftone.ready)throw new Error('Halftone Field WebGL context is unavailable');
  const rig=window.CarouselRig, cd=window.XEM_CAROUSEL_DATA;
  const cpath=rig.buildPath(cd.path.vertices,true,64,[1920,1080]);
  const cdrive=rig.chainedDrive(cd.keys['Move Carousel'],'accumulate');
  let ctx;
  const canvases=scenes.map(s=>document.getElementById(`canvas-${s.id}`));
  function rect(x,y,w,h,color){ctx.fillStyle=color;ctx.fillRect(x,y,w,h);}
  function path(points,fill,stroke,width=1,close=true){ctx.beginPath();points.forEach((p,i)=>i?ctx.lineTo(...p):ctx.moveTo(...p));if(close)ctx.closePath();if(fill){ctx.fillStyle=fill;ctx.fill();}if(stroke){ctx.strokeStyle=stroke;ctx.lineWidth=width;ctx.stroke();}}
  function line(x1,y1,x2,y2,color,width=1,dash=[]){ctx.beginPath();ctx.setLineDash(dash);ctx.moveTo(x1,y1);ctx.lineTo(x2,y2);ctx.strokeStyle=color;ctx.lineWidth=width;ctx.stroke();ctx.setLineDash([]);}
  function rr(x,y,w,h,r,fill,stroke,width=1){ctx.beginPath();ctx.roundRect(x,y,w,h,r);if(fill){ctx.fillStyle=fill;ctx.fill();}if(stroke){ctx.lineWidth=width;ctx.strokeStyle=stroke;ctx.stroke();}}
  function ellipse(x,y,rx,ry,fill,stroke,width=1){ctx.beginPath();ctx.ellipse(x,y,Math.max(0,rx),Math.max(0,ry),0,0,TAU);if(fill){ctx.fillStyle=fill;ctx.fill();}if(stroke){ctx.strokeStyle=stroke;ctx.lineWidth=width;ctx.stroke();}}
  function text(str,x,y,size,color=C.cream,weight=400,font='DM Sans',spacing=-.04,align='left'){
    ctx.font=`${weight} ${size}px "${font}"`;ctx.fillStyle=color;ctx.textBaseline='alphabetic';ctx.textAlign=align;ctx.letterSpacing=`${size*spacing}px`;ctx.fillText(str,x,y);ctx.letterSpacing='0px';ctx.textAlign='left';
  }
  function textWidth(str,size,weight=400,font='DM Sans',spacing=-.04){ctx.font=`${weight} ${size}px "${font}"`;ctx.letterSpacing=`${size*spacing}px`;const v=ctx.measureText(str).width;ctx.letterSpacing='0px';return v;}
  function fit(str,x,y,max,size,color,weight=400,font='DM Sans'){const w=textWidth(str,size,weight,font);text(str,x,y,w>max?size*max/w:size,color,weight,font);}
  function gradient(y0,y1,stops){const g=ctx.createLinearGradient(0,y0,0,y1);stops.forEach(([t,c])=>g.addColorStop(t,c));return g;}
  function imageFit(im,x,y,w,h,mode='cover',oy=.5){const s=mode==='contain'?Math.min(w/im.width,h/im.height):Math.max(w/im.width,h/im.height);const iw=im.width*s,ih=im.height*s;ctx.drawImage(im,x+(w-iw)/2,y+(h-ih)*oy,iw,ih);}
  function clipped(x,y,w,h,r,fn){ctx.save();ctx.beginPath();ctx.roundRect(x,y,w,h,r);ctx.clip();fn();ctx.restore();}
  function mark(x,y,h,color=C.ink){let im=tintCache[color];if(!im){im=document.createElement('canvas');im.width=imgs.mark.width;im.height=imgs.mark.height;const c=im.getContext('2d');c.drawImage(imgs.mark,0,0);c.globalCompositeOperation='source-in';c.fillStyle=color;c.fillRect(0,0,im.width,im.height);tintCache[color]=im;}ctx.drawImage(im,x,y,h*im.width/im.height,h);}
  function logo(cx,cy,h,color=C.ink,accent=color,alpha=1){ctx.save();ctx.globalAlpha*=alpha;const mh=h*.94,mw=mh*imgs.mark.width/imgs.mark.height,tw=textWidth('Xem',h*1.2,500);const full=mw+tw+h*.22;mark(cx-full/2,cy-mh/2,mh,accent);text('Xem',cx-full/2+mw+h*.22,cy+h*.405,h*1.2,color,500);ctx.restore();return full;}
  function grid(color='rgba(231,216,250,.16)',spacing=96,offset=0){for(let x=64+offset;x<W;x+=spacing)line(x,0,x,H,color,.7);for(let y=40;y<H;y+=82)line(0,y,W,y,color,.7);}
  function bounds(x,y,w,h,color='rgba(34,37,31,.22)'){for(const xx of [x,x+w])line(xx,y-45,xx,y+h+45,color,.6,[4,4]);for(const yy of [y,y+h])line(x-45,yy,x+w+45,yy,color,.6,[4,4]);}
  function dot(x,y,r,color){ellipse(x,y,r,r,color);}
  function effect(t,x=0,y=0,w=W,h=H,alpha=.12){if(window.XEM_EFFECTS_OFF)return;halftone.draw(t);ctx.save();ctx.globalAlpha*=alpha;ctx.globalCompositeOperation='screen';ctx.drawImage(hCanvas,x,y,w,h);ctx.restore();}

  // A continuous vector basin: ledges, strata and fractures recede to a shared
  // horizon. All geometry is reproducible; there are no per-frame random calls.
  function basin(kind='purple',t=0,opts={}){
    const colors=kind==='green'?['#3d70d8','#6b9ee1','#2b887f','#1d7568','#11574e','#0c453f']:
      kind==='red'?['#6d99df','#9dbee4','#c7796a','#b16c60','#995347','#733e39']:
      ['#392076','#b056da','#bc84df','#9361c6','#603793','#3c215f'];
    const hor=opts.horizon??448;
    rect(0,0,W,H,gradient(0,hor,[[0,colors[0]],[.74,colors[1]],[1,kind==='purple'?'#ecc3e5':colors[1]]]));
    if(opts.halftone)effect(t,0,0,W,hor,.13);
    rect(0,hor,W,H-hor,gradient(hor,H,[[0,colors[2]],[1,colors[3]]]));
    // distant shoreline
    path([[0,hor-4],[40,hor-20],[102,hor-17],[137,hor-23],[194,hor-15],[254,hor-15],[278,hor-5],[361,hor-5],[382,hor+7],[0,hor+7]],colors[5]);
    for(let j=0;j<11;j++){const yy=hor+4+j*j*1.37;line(0,yy,W,yy,colors[5],.55+j*.08);}
    // Slab islands grow and slide with depth during the camera advance.
    const slabs=[[930,476,340,32],[100,513,370,36],[720,555,550,55],[-70,634,560,72],[1050,688,430,80],[470,606,156,28],[800,486,155,14]];
    slabs.forEach(([x,y,w,h],i)=>{
      y=hor+(y-448)*(H-hor)/272;h*=Math.max(.3,(H-hor)/272);
      const depth=(y-hor)/(H-hor),shift=t*depth*8;
      x+=(i%2?1:-1)*shift;y+=shift*.23;
      const pts=[];
      for(let k=0;k<12;k++)pts.push([x+w*k/11,y+(rnd(i*40+k)-.5)*h*.32]);
      pts.push([x+w+28,y+h*.43],[x+w*.58,y+h*.59],[x+w*.33,y+h*.53],[x-18,y+h*.7]);
      path(pts,colors[i%3===0?4:5]);
      path([[x-18,y+h*.7],[x+w*.33,y+h*.53],[x+w*.58,y+h*.59],[x+w+28,y+h*.43],[x+w+25,y+h],[x+30,y+h*1.17]],colors[4]);
      line(x,y+2,x+w*.85,y+2,colors[3],.55);
    });
    // Fine branching cracks are important at the wide-to-close handoffs.
    ctx.save();ctx.globalAlpha=.35;
    for(let j=0;j<7;j++){
      const pts=[];for(let k=0;k<9;k++)pts.push([j*220-100+k*33+rnd(j*31+k)*35,hor+55+j*28+k*k*1.5]);
      path(pts,null,colors[5],.8,false);
    }ctx.restore();
  }

  function cliff(points,fill,side,seed){
    path(points,fill);
    for(let i=0;i<points.length-2;i++){
      const [x,y]=points[i],[nx,ny]=points[i+1];
      if(nx-x>12){path([[x,y],[nx,ny],[nx,H+30],[x,H+30]],i%2?fill:side);line(x,y,x,H+20,'rgba(52,22,86,.34)',.8);line(x+7,y+9,x+7,H+10,'rgba(255,255,239,.07)',.6);}
    }
    for(let i=0;i<8;i++){let x=points[0][0]+rnd(seed+i)*Math.max(100,points[points.length-3][0]-points[0][0]);line(x,480+rnd(i)*100,x-8,600+rnd(i+2)*80,'rgba(34,15,58,.22)',.7);}
  }
  function canyon(t=0,spread=1){
    rect(0,0,W,H,gradient(0,H,[[0,'#a860d9'],[.46,'#e4a1d5'],[1,'#f2d9e4']]));
    effect(t,0,0,W,470,.105);
    path([[100,610],[300,497],[442,511],[600,425],[750,441],[875,330],[936,408],[1070,386],[1280,513],[1280,720],[0,720]],'#eddaef', '#d5a5d6',1);
    path([[670,512],[825,386],[875,330],[861,375],[822,408],[819,428],[762,469]],'#fff5ed');
    ctx.save();ctx.translate(-t*14*spread,0);
    cliff([[-100,0],[124,0],[126,197],[164,251],[172,363],[242,409],[245,477],[319,487],[345,550],[411,557],[432,720],[-100,720]],'#ac68df','#b978e7',22);
    cliff([[-110,458],[228,458],[240,471],[290,476],[292,508],[339,508],[340,569],[393,569],[416,720],[-110,720]],'#633395','#743eb0',32);
    path([[-100,458],[228,458],[211,468],[275,467],[291,479],[12,478],[-100,483]],'#a358d7');
    ctx.restore();
    ctx.save();ctx.translate(t*19*spread,0);
    cliff([[884,720],[885,402],[932,356],[936,275],[1002,228],[1003,199],[1085,157],[1088,100],[1178,79],[1178,22],[1208,0],[1380,0],[1380,720]],'#a35cdb','#bc7cea',72);
    path([[691,720],[697,699],[751,697],[751,674],[797,671],[816,645],[842,644],[842,617],[895,613],[901,585],[943,581],[943,544],[1005,536],[1021,510],[1064,500],[1086,477],[1380,477],[1380,720]],'#318b81');
    path([[754,720],[754,696],[823,696],[823,669],[887,669],[905,640],[955,640],[955,615],[1005,615],[1008,575],[1063,575],[1073,523],[1380,523],[1380,720]],'#125951');
    path([[836,720],[836,696],[927,696],[927,663],[1111,663],[1111,633],[1380,633],[1380,720]],'#20766d');
    ctx.restore();
  }

  // Shaded, rotating mail capsule. The same latitude/longitude model renders
  // wireframe and material views, so there is no mesh swap at frame 444.
  function capsule(x,y,r,yaw=0,pitch=.08,wire=false,opacity=1){
    ctx.save();ctx.translate(x,y);ctx.globalAlpha*=opacity;
    const stroke=wire?'#939889':'#754894';
    function p(lat,lon,rad=1){let X=Math.cos(lat)*Math.sin(lon+yaw),Z=Math.cos(lat)*Math.cos(lon+yaw),Y=Math.sin(lat);const yy=Y*Math.cos(pitch)-Z*Math.sin(pitch),zz=Y*Math.sin(pitch)+Z*Math.cos(pitch);return [X*r*rad,yy*r*rad,zz];}
    if(!wire){const g=ctx.createRadialGradient(-r*.37,-r*.5,r*.08,r*.24,r*.22,r*1.5);g.addColorStop(0,'#f4d8e4');g.addColorStop(.32,'#cb93ea');g.addColorStop(.7,'#d9a8d3');g.addColorStop(1,'#78499f');ellipse(0,0,r,r,g,stroke,.85);}
    else ellipse(0,0,r,r,null,stroke,.75);
    // longitude window patches, their outlines are sampled on the sphere.
    for(let panel=0;panel<4;panel++){
      const center=panel*Math.PI/2,pts=[],edge=[];
      const l0=-.40,l1=.32,u0=center-.67,u1=center+.67;
      for(let k=0;k<=64;k++){
        // Rounded panel perimeter in spherical parameter space.
        const a=TAU*k/64;const u=(u0+u1)/2+Math.sign(Math.cos(a))*Math.pow(Math.abs(Math.cos(a)),.23)*(u1-u0)/2;
        const lat=(l0+l1)/2+Math.sign(Math.sin(a))*Math.pow(Math.abs(Math.sin(a)),.23)*(l1-l0)/2;
        const q=p(lat,u);pts.push(q.slice(0,2));edge.push(q[2]);
      }
      if(edge.filter(v=>v>0).length<4)continue;
      for(let k=0;k<pts.length;k++)if(edge[k]<0){const len=Math.hypot(...pts[k]);pts[k]=pts[k].map(v=>v/len*r);}
      ctx.save();ctx.beginPath();ctx.arc(0,0,r-.8,0,TAU);ctx.clip();
      if(wire)path(pts,null,stroke,.8);
      else {
        const g=ctx.createLinearGradient(-r,-r,r,r);g.addColorStop(0,'#f0c591');g.addColorStop(.5,'#be829d');g.addColorStop(1,'#8e53aa');path(pts,g,stroke,r*.006);
        const inset=pts.map(([px,py])=>[px*.972,py*.955]);path(inset,null,'#e6b4e8',r*.006);
        ctx.save();path(pts,null,null);ctx.clip();path([[-r*.8,-r],[-r*.56,-r],[r*.9,r],[r*.64,r]],'rgba(255,255,239,.15)');ctx.restore();
      }ctx.restore();
    }
    for(const lat of [-1.02,-.8,-.53,.44,.57,1.1]){
      const pts=[];for(let k=0;k<=120;k++){const q=p(lat,k/120*TAU);if(q[2]>=-.025)pts.push(q.slice(0,2));}
      // split at wrap to avoid drawing an accidental chord
      let seg=[];for(const q of pts){if(seg.length&&Math.hypot(q[0]-seg[seg.length-1][0],q[1]-seg[seg.length-1][1])>r*.15){if(seg.length>1)path(seg,null,stroke,.7,false);seg=[];}seg.push(q);}if(seg.length>1)path(seg,null,stroke,.7,false);
    }
    if(wire){for(const lon of [0,Math.PI/2,Math.PI,Math.PI*1.5]){const a=[];for(let k=0;k<=70;k++){const q=p(-Math.PI/2+k/70*Math.PI,lon);if(q[2]>0)a.push(q.slice(0,2));}if(a.length>1)path(a,null,'#b8beae',.6,false);}}
    if(!wire){ellipse(0,r*.905,r*.23,r*.054,'#533584',stroke,.7);ellipse(0,r*.916,r*.17,r*.024,C.lavender);}
    ctx.restore();
  }
  function envelope(x,y,s,rot=0,color=C.lavender){ctx.save();ctx.translate(x,y);ctx.rotate(rot);path([[-s*.6,-s*.38],[s*.45,-s*.48],[s*.58,s*.3],[-s*.47,s*.4]],color,'#9773b3',s*.025);path([[-s*.6,-s*.38],[0,s*.08],[s*.45,-s*.48]],null,'#9773b3',s*.025,false);path([[-s*.47,s*.4],[0,s*.08],[s*.58,s*.3]],null,'#a882b7',s*.023,false);ctx.restore();}
  function beam(x,y,half=230,alpha=1){ctx.save();ctx.globalAlpha*=alpha;path([[x,y],[x-half,H+70],[x+half,H+70]],gradient(y,H,[[0,'#ffffef'],[.36,'rgba(255,255,239,.8)'],[1,'rgba(255,255,239,0)']]));for(let i=-3;i<=3;i++)line(x,y,x+i*half/3,H+70,'rgba(255,255,239,.26)',.7);ctx.restore();}
  function rays(cx,cy,scale=1,alpha=1){ctx.save();ctx.translate(cx,cy);ctx.scale(scale,scale);ctx.globalAlpha*=alpha;const ends=[[-520,-550],[-360,-500],[-560,225],[-140,480],[290,460],[560,-480],[150,-590],[90,370]];ends.forEach(([x,y],i)=>{line(0,0,x,y,C.cream,i%3===0?1:.5,i%2?[4,4]:[]);dot(x,y,2,C.cream);});ctx.restore();}
  function icon(x,y,r,type,alpha=1){ctx.save();ctx.translate(x,y);ctx.scale(r/40,r/40);ctx.globalAlpha*=alpha;const col=C.lavender;ctx.lineWidth=.6;
    if(type%8===0){for(let i=0;i<5;i++)ellipse(0,0,8+i*7,8+i*7,null,col,.55);line(-39,0,39,0,col,.4);line(0,-39,0,39,col,.4);}
    if(type%8===1){path([[-26,-16],[0,-32],[26,-16],[26,16],[0,32],[-26,16]],null,col,.7);path([[-26,-16],[0,0],[26,-16]],null,col,.7,false);line(0,0,0,32,col,.7);path([[-14,-8],[0,-17],[14,-8],[14,8],[0,17],[-14,8]],null,col,.6);}
    if(type%8===2){rr(-31,-24,62,48,3,null,col,.7);path([[-29,-21],[0,1],[29,-21]],null,col,.7,false);path([[-29,23],[0,1],[29,23]],null,col,.7,false);ellipse(0,0,39,39,null,col,.4);}
    if(type%8===3){for(const k of [-.8,0,.8]){ctx.save();ctx.rotate(k);ellipse(0,0,37,15,null,col,.65);ctx.restore();}dot(0,0,2,col);}
    if(type%8===4){path([[-34,-30],[34,-30],[0,35]],null,col,.65);for(let i=0;i<4;i++)ellipse(0,-30+i*16,34-i*8,9-i*1.6,null,col,.6);}
    if(type%8===5){for(const yy of [-23,0,23]){ellipse(0,yy,28,9,null,col,.65);}line(-28,-23,-28,23,col,.6);line(28,-23,28,23,col,.6);}
    if(type%8===6){rr(-33,-30,66,60,5,null,col,.65);line(-33,-14,33,-14,col,.65);for(const xx of [-17,0,17])for(const yy of [0,16])rect(xx-2,yy-2,4,4,col);line(-18,-35,-18,-23,col,.9);line(18,-35,18,-23,col,.9);}
    if(type%8===7){rays(0,0,.068,1);}
    ctx.restore();}

  function carousel(t,x,y,w,h){
    clipped(x,y,w,h,22,()=>{rect(x,y,w,h,C.lavender);if(window.XEM_EFFECTS_OFF)return;ctx.save();ctx.translate(x,y);ctx.scale(w/3840,h/2160);
      const ct=t*1.18+.3,c={...cd.controls,moveCarousel:cdrive.at(ct)};
      const poses=rig.pathRig(cpath,c,cd.comp,12,cd.controls.speed/1000*ct);
      // Original Circle 4 uses its document order as the depth order.
      for(let i=11;i>=0;i--){const p=poses[i],im=imgs[templates[i]],sz=1500/Math.max(im.width,im.height);ctx.save();ctx.translate(p.x,p.y);ctx.rotate((p.rotZ||0)*Math.PI/180);ctx.scale(p.scale,p.scale);const iw=im.width*sz,ih=im.height*sz;ctx.shadowColor='rgba(34,37,31,.18)';ctx.shadowBlur=30;ctx.shadowOffsetY=12;ctx.drawImage(im,-iw/2,-ih/2,iw,ih);ctx.restore();}ctx.restore();
    });
  }
  function newsletter(kind,t,box={x:40,y:42,w:1200,h:636},title=true){clipped(box.x,box.y,box.w,box.h,70,()=>{ctx.save();ctx.translate(box.x,box.y);ctx.scale(box.w/W,box.h/H);basin(kind,t,{horizon:450});if(title){text('Your next newsletter',640,326,87,C.cream,400,'EB Garamond',-.045,'center');logo(640,620,35,C.cream);}ctx.restore();});}
  function productTile(key,x,y,w,h){clipped(x,y,w,h,24,()=>{rect(x,y,w,h,C.cream);imageFit(imgs[key],x+8,y+8,w-16,h-16,'cover',.02);});}
  function mosaic(f){
    rect(0,0,W,H,C.ink);const p=out(span(f,367,395)),s=mix(1,.38,p),cx=640,cy=360;
    ctx.save();ctx.translate(cx,cy);ctx.scale(s,s);ctx.translate(-cx,-cy);
    const tw=1200,th=636,gap=60;
    for(let row=-1;row<=1;row++)for(let col=-1;col<=1;col++){
      const x=40+col*(tw+gap),y=42+row*(th+gap);
      if(col===0&&row===0)newsletter('green',f/30,{x,y,w:tw,h:th});
      else if(row===-1&&col===0)carousel((f-367)/30,x,y,tw,th);
      else if(row===0&&col===-1)productTile('automations',x,y,tw,th);
      else if(row===0&&col===1){clipped(x,y,tw,th,40,()=>{ctx.save();ctx.translate(x,y);ctx.scale(tw/W,th/H);space(f/30,false);ctx.restore();});}
      else if(row===1&&col===0){clipped(x,y,tw,th,40,()=>{ctx.save();ctx.translate(x,y);ctx.scale(tw/W,th/H);basin('purple',f/30);rr(542,265,200,200,32,C.cream);mark(602,292,148,C.iris);ctx.restore();});}
      else if(row===1&&col===1)productTile('newsletters',x,y,tw,th);
      else if(row===-1&&col===-1)productTile('editor-figma',x,y,tw,th);
      else newsletter('red',f/30,{x,y,w:tw,h:th},false);
    }ctx.restore();
  }
  function rock(x,y,r,rotation,seed,color='#49305e'){
    ctx.save();ctx.translate(x,y);ctx.rotate(rotation);const pts=[];for(let k=0;k<11;k++){const a=TAU*k/11,rr=r*(.75+rnd(seed+k)*.25);pts.push([Math.cos(a)*rr,Math.sin(a)*rr]);}path(pts,color,'#2e1b41',.9);
    path([pts[1],pts[2],[r*.25,r*.1],[-r*.34,r*.34],pts[6]],null,'#281737',.9,false);path([pts[1],[r*.25,r*.1],pts[9]],null,'rgba(255,255,239,.12)',.7,false);ctx.restore();
  }
  function space(t,withType=true){
    rect(0,0,W,H,'#241b32');let g=ctx.createRadialGradient(1050,140,20,750,240,900);g.addColorStop(0,'#9b55cc');g.addColorStop(.35,'#7041a1');g.addColorStop(1,'#241b32');rect(0,0,W,H,g);effect(t,0,0,W,H,.065);
    // Curved cliff rim with layered jagged strata (separate from the sky).
    for(let layer=0;layer<5;layer++){
      const pts=[[-120,780]];for(let k=0;k<=36;k++){const a=-Math.PI*.67+k/36*Math.PI*.54;const radius=660+layer*17+(rnd(k+layer*80)-.5)*20;pts.push([45+Math.cos(a)*radius,1150+Math.sin(a)*radius]);}pts.push([900,1000]);path(pts,['#a267d6','#7c49b7','#623593','#492b6d','#321e49'][layer]);
    }
    for(let i=0;i<14;i++){const x=rnd(i+70)*W,y=rnd(i+120)*H;const xx=x+Math.sin(t*.5+i)*14,yy=y+(t-26)*(i%3-1)*6;rock(xx,yy,12+rnd(i+10)*52,t*.08+i,i*10);}
    for(let i=0;i<9;i++){const x=220+rnd(i+570)*940,y=30+rnd(i+220)*660;envelope(x+Math.sin(t+i)*6,y,13+rnd(i)*7,i*.7+t*.09,i%3?C.lavender:C.lemon);}
    const p=span(t*30,775,840),x=mix(985,746,p),y=mix(250,196,p),r=mix(134,112,p);
    capsule(x,y,r,.1+p*1.5,.14,false);
    if(withType){const offset=mix(60,0,out(span(t*30,775,790))),down=smooth(span(t*30,805,840));text('Make email',154+offset-65*down,352+150*down,126,C.cream,500);text('more human.',154+offset-65*down,490+150*down,144,C.cream,500);}
  }
  function laptop(f){
    rect(0,0,W,H,gradient(0,H,[[0,'#734aa5'],[.65,'#c17ddd'],[1,'#f4c5aa']]));
    path([[0,641],[220,637],[650,650],[1020,662],[1280,655],[1280,720],[0,720]],'#724445');
    path([[0,683],[320,678],[760,690],[1280,685],[1280,720],[0,720]],'#241e27');
    const p=out(span(f,637,664)),s=mix(.64,1.06,p);
    ctx.save();ctx.translate(640,654);ctx.scale(s,s);ctx.translate(-640,-654);
    ctx.save();ctx.shadowColor='rgba(18,10,21,.65)';ctx.shadowBlur=28;ctx.shadowOffsetY=12;rr(264,142,752,492,19,'#151713','#554454',2);ctx.restore();
    clipped(274,151,732,466,10,()=>{const im=imgs.site;rect(274,151,732,466,C.cream);const iw=im.width,ih=im.height;const sw=iw,sh=iw*466/732;const sy=Math.min(Math.max(0,ih-sh),mix(560,0,smooth(span(f,640,675))));ctx.drawImage(im,0,sy,sw,sh,274,151,732,466);});
    rr(585,151,110,11,[0,0,6,6],'#151713');dot(639,156,2,'#343b36');
    path([[245,634],[1035,634],[1029,661],[252,661]],gradient(634,661,[[0,'#3d303d'],[.4,'#10110e'],[1,'#252222']]),'#493649',1);
    rr(585,634,110,8,[0,0,6,6],'#564551');line(279,624,1004,624,'#0e100d',3);
    ctx.restore();
  }
  function cap(cx,cy,scale){
    ctx.save();ctx.translate(cx,cy);ctx.scale(scale,scale);ctx.rotate(-.16);
    const body=ctx.createLinearGradient(-100,-120,125,110);body.addColorStop(0,'#f6f0d7');body.addColorStop(.5,'#ded9bd');body.addColorStop(1,'#b2b09b');
    ctx.beginPath();ctx.moveTo(-153,47);ctx.bezierCurveTo(-140,-56,-80,-137,16,-133);ctx.bezierCurveTo(130,-138,171,-55,164,75);ctx.bezierCurveTo(52,93,-74,70,-153,47);ctx.fillStyle=body;ctx.fill();ctx.strokeStyle='#ccc6af';ctx.lineWidth=1.4;ctx.stroke();
    ctx.beginPath();ctx.moveTo(-153,42);ctx.bezierCurveTo(-209,64,-239,80,-247,96);ctx.bezierCurveTo(-208,133,-128,162,-36,154);ctx.bezierCurveTo(26,145,77,117,115,85);ctx.bezierCurveTo(37,79,-74,51,-153,42);ctx.fillStyle=gradient(48,170,[[0,'#e9e4cd'],[1,'#c3bda3']]);ctx.fill();ctx.strokeStyle='#d2cbb1';ctx.stroke();
    ctx.beginPath();ctx.moveTo(16,-133);ctx.bezierCurveTo(-31,-89,-66,-18,-67,66);ctx.moveTo(16,-133);ctx.bezierCurveTo(94,-72,120,-17,115,84);ctx.strokeStyle='#c8c2aa';ctx.lineWidth=1.2;ctx.stroke();ellipse(16,-136,10,4,'#d6d1b7');
    for(const [x,y] of [[70,-70],[-73,-53]]){ellipse(x,y,4,5,'#c7c0a9');ellipse(x,y-1,2.3,3,'#e3ddc4');}
    ctx.save();ctx.translate(-38,-15);ctx.rotate(.1);rr(-30,-36,60,78,8,C.iris,'#403256',1);mark(-9,-24,36,C.lemon);text('Xem',0,27,18,C.cream,500,'DM Sans',-.04,'center');ctx.restore();
    ctx.restore();
  }
  function shirt(cx,cy,s){
    ctx.save();ctx.translate(cx,cy);ctx.scale(s,s);const body=gradient(-170,180,[[0,'#252a24'],[.65,'#1a211c'],[1,'#101712']]);
    const pts=[[-63,-172],[-99,-159],[-146,-129],[-187,-42],[-216,113],[-178,124],[-127,6],[-117,179],[105,184],[119,5],[173,122],[211,107],[177,-51],[143,-134],[93,-161],[60,-174]];
    path(pts,body,'#3a4038',1.2);ctx.beginPath();ctx.moveTo(-63,-172);ctx.bezierCurveTo(-43,-128,46,-131,60,-174);ctx.strokeStyle='#0e1510';ctx.lineWidth=10;ctx.stroke();
    line(-119,8,-105,158,'#30372f',1.2);line(111,12,96,166,'#0e140f',2);line(-113,167,104,172,'#353c33',1);line(-171,104,-208,95,'#394035',2);line(175,102,206,88,'#384034',2);
    for(const [yy,type] of [[-80,2],[0,3],[80,4]])icon(0,yy,33,type,.48);text('Make email more human.',0,137,9,'#a5ab9a',400,'DM Sans',0,'center');
    ctx.restore();
  }
  function poster(cx,cy,s){ctx.save();ctx.translate(cx,cy);ctx.rotate(.23);ctx.scale(s,s);clipped(-126,-188,252,376,0,()=>{ctx.save();ctx.translate(-126,-188);rect(0,0,252,376,gradient(0,285,[[0,C.iris],[1,'#cd85e0']]));path([[0,243],[41,254],[82,245],[126,257],[169,249],[211,256],[252,244],[252,376],[0,376]],C.forest);path([[0,290],[85,276],[180,286],[252,277],[252,376],[0,376]],'#2d8d7b');for(let i=0;i<5;i++)line(0,296+i*13,252,289+i*14,'#206b60',.6);logo(126,40,20,C.cream);capsule(126,176,56,.2,.13);text('Make email',126,335,20,C.cream,400,'EB Garamond',-.03,'center');text('more human.',126,355,20,C.cream,400,'EB Garamond',-.03,'center');ctx.restore();});ctx.restore();}
  function shadow(x,y,w,h){const g=ctx.createRadialGradient(x,y,0,x,y,w);g.addColorStop(0,'rgba(47,42,36,.25)');g.addColorStop(1,'rgba(47,42,36,0)');ctx.save();ctx.translate(x,y);ctx.scale(1,h/w);ellipse(0,0,w,w,g);ctx.restore();}
  function brandedCapPlate(x,y,w,h){ctx.save();ctx.translate(x,y);ctx.scale(w/426,h/720);ctx.drawImage(imgs.cap,0,0,426,720);ctx.save();ctx.transform(.28,.082,-.103,.275,189,293);rr(-3,-3,206,225,23,C.iris,'#342858',4);mark(75,24,106,C.lemon);text('Xem',100,192,53,C.cream,500,'DM Sans',-.04,'center');ctx.restore();ctx.restore();}
  function brandedShirtPlate(x,y,w,h){ctx.save();ctx.translate(x,y);ctx.scale(w/427,h/720);ctx.drawImage(imgs.shirt,0,0,427,720);mark(205,270,37,'#b1b5a2');for(const [yy,type] of [[337,2],[397,3]])icon(217,yy,23,type,.75);text('Make email more human.',217,448,9,'#a5ab9a',400,'DM Sans',0,'center');ctx.restore();}
  function merch(f){const p=out(span(f,688,698));const h=mix(1123,720,p),y=mix(-202,0,p);ctx.drawImage(imgs.cap,0,0,1,720,0,y,W,h);if(f<688){const s=1.56;const w=426*s,h=720*s;brandedCapPlate(640-w/2,360-h/2,w,h);}else{const boundary=mix(1280,426,p);clipped(0,0,boundary,H,0,()=>brandedCapPlate(mix(307,0,p),y,mix(665,426,p),h));rect(boundary,0,427,H,'#d9d5ce');shadow(boundary+218,582,163,27);poster(boundary+218,363,.86);brandedShirtPlate(boundary+427,0,427,720);}}

  function drawScene(f){
    if(f<41){
      rect(0,0,W,H,C.cream);const p=out(span(f,0,23)),h=238,mw=h*imgs.mark.width/imgs.mark.height,x=640-mw/2,y=360-h/2;
      ctx.save();ctx.globalAlpha=.32*(1-span(f,33,41));
      for(let i=0;i<16;i++){const a=i/16*TAU;line(640,360,640+Math.cos(a)*300,360+Math.sin(a)*300,'#a6afa0',.7);}
      bounds(x,y,mw,h,'#a1aa9a');for(const [cx,cy,r] of [[x+111,y+36,36],[x+44,y+204,36]])ellipse(cx,cy,r,r,null,'#a1aa9a',.7);ctx.restore();
      ctx.save();ctx.globalAlpha=p*.21;mark(x,y,h,C.ink);ctx.restore();
      const anchors=[[x+114,y+35,34],[x+40,y+204,34],[x+28,y+117,32],[x+116,y+124,32]];
      anchors.forEach(([xx,yy,r],i)=>{const a=out(span(f,2+i*3,10+i*3));ctx.save();ctx.globalAlpha=.12;dot(mix(640,xx,a),mix(360,yy,a),r*a,C.ink);ctx.restore();});
      ctx.save();ctx.beginPath();ctx.rect(0,y+h*(1-out(span(f,28,40))),W,H);ctx.clip();mark(x,y,h,C.ink);ctx.restore();
    }else if(f<83){
      rect(0,0,W,H,C.cream);const p=out(span(f,41,48));const mh=mix(238,186,p),tw=textWidth('Xem',240,500),full=mh*imgs.mark.width/imgs.mark.height+44+tw;
      const mx=mix(640-mh*imgs.mark.width/imgs.mark.height/2,640-full/2,p);mark(mx,360-mh/2,mh,C.ink);
      ctx.save();ctx.beginPath();ctx.rect(mx+mh*.7,200,(tw+80)*p,360);ctx.clip();text('Xem',640-full/2+mh*imgs.mark.width/imgs.mark.height+44,441,240,C.ink,500);ctx.restore();
      const x=640-full/2-22,w=full+44;bounds(x,255,w,210,'rgba(34,37,31,.19)');
      if(f>=61){const a=out(span(f,61,66));rect(x,239,w*a,24,C.lemon);rect(x+w*(1-a),465,w*a,24,C.lemon);if(f>=69){rect(x-20,239,20,250*out(span(f,69,74)),C.lemon);rect(x+w,239,20,250*out(span(f,69,74)),C.lemon);}text('x',x-22,250,11,C.ink,400,'DM Sans',0);text('x',x-22,482,11,C.ink,400,'DM Sans',0);}
    }else if(f<126){
      rect(0,0,W,H,C.ink);const alpha=1-.45*smooth(span(f,108,125));logo(640,360,103,C.cream,C.cream,alpha);
      ctx.save();ctx.globalAlpha=.13*(1-span(f,112,126));bounds(433,304,416,111,C.lavender);for(let i=0;i<8;i++)line(442+i*16,0,442+i*16,340,C.lavender,.5);ctx.restore();
      if(f>=118){const p=out(span(f,118,126));dot(mix(495,395,p),mix(326,128,p),mix(13,17,p),C.cream);}
    }else if(f<143){rect(0,0,W,H,C.lavender);const p=out(span(f,126,142));ctx.save();ctx.translate(mix(395,640,p),mix(128,360,p));ctx.rotate(-.25*(1-p));rr(-46*(1-p)-24,-24,92*(1-p)+48,48,24,C.ink);ctx.restore();
    }else if(f<192){
      rect(0,0,W,H,C.lavender);const p=out(span(f,143,156));const scale=mix(.09,1,p);ctx.save();ctx.translate(640,360);ctx.scale(scale,scale);ctx.translate(-640,-360);
      const sw=300,sh=296;const colors=[C.cream,C.lemon,C.forest,C.ink,C.iris,'#7943ab',C.lemon,C.lavender,C.cream];const names=['Cream','Lemon','Forest','Ink','Iris','Violet','Lemon','Lavender','Cream'];
      for(let i=0;i<9;i++){const col=i%3,row=Math.floor(i/3),x=col*426+100,y=row*402-440+mix(20,80,span(f,155,192))+(i%2?24:0);rr(x,y,sw,sh,69,colors[i]);text(names[i],x+sw/2,y+sh/2-5,16,i===3||i===4||i===5?C.cream:C.ink,400,'DM Sans',0,'center');text(colors[i].toUpperCase(),x+sw/2,y+sh/2+20,13,i===3||i===4||i===5?C.lavender:C.ink,400,'DM Sans',0,'center');}
      for(const x of [426,852])line(x,0,x,H,'rgba(34,37,31,.23)',.7,[5,5]);for(const y of [178,580])line(0,y,W,y,'rgba(34,37,31,.23)',.7,[5,5]);ctx.restore();
    }else if(f<226){rect(0,0,W,H,'#bb8ce8');const p=span(f,192,226);text('From a message',-70-p*35,184,260,C.ink,500);text('to',-8-p*25,427,244,C.lemon,400,p>.54?'EB Garamond':'DM Sans');text('a connection.',212-p*25,427,244,C.ink,400,p>.54?'EB Garamond':'DM Sans');text('Make it your own.',-30-p*45,680,244,C.ink,500);
    }else if(f<282){
      rect(0,0,W,H,C.ink);grid('rgba(231,216,250,.22)',240);const p=span(f,226,249);
      fit('Every',55,214,940,220,C.lavender,500);ctx.save();ctx.beginPath();ctx.rect(0,245,1280*out(p),230);ctx.clip();text('email.',58,417,220,C.lavender,500);ctx.restore();
      const serif=f>=258;const t=out(span(f,233,250));ctx.save();ctx.beginPath();ctx.rect(250,445,1000*t,230);ctx.clip();fit('More human.',290,620,934,190,serif?C.lavender:C.lemon,400,serif?'EB Garamond':'DM Sans');ctx.restore();
      text('DM Sans Medium',882,150,16,C.lavender,400,'DM Sans',-.02);text('105% / −4%',882,170,16,C.lavender);text('EB Garamond Regular',66,552,15,C.lavender);text('105% / −4%',66,572,15,C.lavender);
    }else if(f<316){
      rect(0,0,W,H,C.ink);grid('rgba(231,216,250,.14)',230);
      const p=out(span(f,291,316));const x=mix(-690,40,p),y=mix(-621,42,p),w=mix(1690,1200,p),h=mix(1240,636,p);rr(x,y,w,h,mix(42,70,p),gradient(0,H,[[0,'#432477'],[1,'#532b8b']]),'rgba(231,216,250,.6)',.8);
      if(f<305)dot(964,574,19,C.lemon);else{for(const [xx,yy] of [[x+5,y+5],[x+w-5,y+5],[x+5,y+h-5],[x+w-5,y+h-5]]){line(xx-14,yy,xx+14,yy,C.lavender,.8);line(xx,yy-14,xx,yy+14,C.lavender,.8);}}
    }else if(f<349){
      rect(0,0,W,H,C.ink);const a=out(span(f,316,329));clipped(40,42,1200,636,70,()=>{ctx.save();ctx.translate(40,42);ctx.scale(1200/W,636/H);basin('purple',(f-316)/20,{horizon:mix(740,448,a),halftone:true});ctx.restore();});
      if(f>337){ctx.save();ctx.globalAlpha=span(f,337,348)*.55;bounds(218,203,847,340,C.lavender);for(const [x,y] of [[218,203],[1065,203],[218,543],[1065,543]])dot(x,y,3,C.cream);ctx.restore();}
    }else if(f<367){rect(0,0,W,H,C.ink);const kind=f<355?'red':f<361?'purple':'green';newsletter(kind,f/30);bounds(220,211,846,337,'rgba(255,255,239,.45)');
    }else if(f<415){mosaic(f);
    }else if(f<444){rect(0,0,W,H,C.cream);bounds(426,200,430,310,'rgba(34,37,31,.26)');line(640,0,640,H,'rgba(34,37,31,.16)',.7,[4,4]);capsule(640,364,128,-.4+(f-415)/29*.9,.12,true);
    }else if(f<459){const p=out(span(f,444,457));ctx.save();ctx.globalAlpha=p;canyon(0);ctx.restore();if(p<1){ctx.save();ctx.globalAlpha=1-p;rect(0,0,W,H,C.cream);ctx.restore();}capsule(640,364,mix(128,156,p),.5,.12,false,p);if(p<1)capsule(640,364,128,.5,.12,true,1-p);
    }else if(f<487){const p=smooth(span(f,459,477)),rise=out(span(f,477,487));canyon((f-459)/20);const x=f<477?mix(640,216,p):mix(216,640,rise),y=mix(364,55,rise);capsule(x,y,mix(mix(156,132,p),76,rise),.5+(f-459)/42,.12);if(f>481)line(x,y+76,x,H,C.cream,1);
    }else if(f<523){const p=out(span(f,487,501));rect(0,0,W,H,gradient(0,H,[[0,'#714494'],[.35,'#b675ca'],[1,'#f1c5c6']]));effect(f/30,0,0,W,H,.08);const x=640,y=mix(55,132,p);beam(x,y+68,mix(2,344,p),1);capsule(x,y,76,.8+(f-487)/30,.14);envelope(x+22,450-(f-487)*2.2,26,.35+(f-487)*.022,C.lavender);
    }else if(f<548){const kind=f<531?'red':f<540?'purple':'green';basin(kind,f/30,{horizon:410});const p=span(f,523,548);rays(640,411,1,.6);beam(640,198,mix(140,18,out(p)),1-p);if(f<540)capsule(640,114-80*p,78-20*p,1.1+p,.1);if(f>=540)dot(640,411,3,C.cream);
    }else if(f<573){rect(0,0,W,H,C.ink);const s=mix(1,.13,out(span(f,548,570)));rays(640,360,s,.74);
    }else if(f<637){
      rect(0,0,W,H,C.ink);const p=out(span(f,573,622)),cols=7,rows=5,scale=mix(1.8,1,p);ctx.save();ctx.translate(640,360);ctx.scale(scale,scale);ctx.translate(-640,-360);
      for(let r=0;r<rows;r++)for(let c=0;c<cols;c++){const x=160+c*160,y=110+r*126;const delay=Math.hypot(c-3,r-2)*3;const a=.68*out(span(f,573+delay,589+delay));if(a>0)icon(x,y,30,(r*cols+c),a);}
      rays(640,360,.07,.7*(1-out(span(f,573,582))));ctx.restore();
    }else if(f<679){laptop(f);
    }else if(f<712){merch(f);
    }else if(f<775){
      const kind=f<746?'green':f<761?'red':'purple';basin(kind,f/30,{horizon:565});
      const desc=f<746?'newsletters':f<761?'automations':'open source';const start=f<746?712:f<761?746:761;const p=out(span(f,start,start+7));
      const baseX=255,cy=361,mh=67;mark(baseX,cy-mh/2,mh,C.cream);text('Xem',baseX+60,cy+29,91,C.cream,500);const dx=baseX+260;ctx.save();ctx.beginPath();ctx.rect(dx-4,270,800*p,145);ctx.clip();text(desc,dx+18*(1-p),cy+27,85,C.cream,400);ctx.restore();
      if(f<746){bounds(245,300,784,120,'rgba(255,255,239,.23)');text('Email made yours',565,248,10,C.cream,400,'DM Sans',0);text('Open source. Human by design.',895,248,10,C.cream,400,'DM Sans',0);}
    }else if(f<840){space(f/30);
    }else if(f<854){
      rect(0,0,W,H,C.ink);const p=out(span(f,840,854));logo(640,360,80,C.cream,C.lemon,p);capsule(mix(746,524,p),mix(196,364,p),mix(112,12,p),2+p*2,.1,false,1-p);
    }else{rect(0,0,W,H,C.ink);logo(640,360,80,C.cream,C.lemon);}
  }
  const frameLog=[];
  function draw(t){const f=Math.min(900,Math.max(0,Math.floor(t*FPS+1e-4)));let i=scenes.findIndex(s=>f>=s.start&&f<s.end);if(i<0)i=scenes.length-1;const cv=canvases[i];ctx=cv.getContext('2d',{alpha:false});ctx.setTransform(1.5,0,0,1.5,0,0);ctx.globalAlpha=1;ctx.globalCompositeOperation='source-over';ctx.clearRect(0,0,W,H);drawScene(f);}
  const driver={_t:0};Object.defineProperty(driver,'t',{get(){return this._t;},set(t){this._t=t;draw(t);}});
  draw(0);
  window.XEM_RENDER={draw,scenes,assets:Object.keys(imgs),effects:['Circle Carousel 4','Halftone Field']};
  return {driver,draw};
};
