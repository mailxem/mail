// Shared, semantic Tailwind recipes for the Xem workspace.
// Color values live in app/globals.css; content previews keep their own artwork.
// Keep complete utility strings here so Tailwind can discover every variant.
const recipes: Record<string, string> = {
  "document-theme": String.raw`
    [font-family:var(--font-sans)] [font-size:14px] [&_button:disabled]:cursor-not-allowed [&_button:disabled]:opacity-50
  `,
  "product-frame": String.raw`
    flex [min-height:0] [border-radius:12px] [background:var(--muted)] overflow-hidden [border:1px_solid_var(--border)] [box-shadow:none] [height:100dvh] [max-height:100dvh]
    max-[1100px]:[margin:0] max-[1100px]:[border:0] max-[1100px]:[border-radius:0] max-[1100px]:[min-height:0] max-[1100px]:[width:100%]
  `,
  "product-sidebar": String.raw`
    [width:238px] shrink-0 [background:var(--card)] flex flex-col [padding:28px_20px_16px] [border-right:1px_solid_var(--border)] relative [top:0] [align-self:flex-start] [height:100%] overflow-y-auto [scrollbar-width:thin]
    min-[1500px]:[width:250px]
    max-[1100px]:[width:205px] max-[1100px]:[height:100svh] max-[1100px]:[padding:24px_14px]
    max-[780px]:hidden
    [&_nav]:[flex:1]
  `,
  "brand": String.raw`
    flex items-center [gap:11px] [font-size:23px] [letter-spacing:-0.8px] font-semibold [margin:0_0_33px_3px]
  `,
  "brand-icon": String.raw`
    grid [place-items:center] [width:36px] [height:36px] [border-radius:10px] [background:var(--muted)] [color:var(--foreground)] [box-shadow:none]
  `,
  "collapse-sidebar": String.raw`
    [margin-left:auto] [color:var(--muted-foreground)]
  `,
  "product-nav-item": String.raw`
    [height:44px] [padding:0_12px] flex items-center [gap:12px] [color:var(--muted-foreground)] [border:1px_solid_transparent] [border-radius:9px] [margin:3px_0] [font-size:14px] [transition:background_0.15s]
    [&:hover]:[background:var(--muted)] [&:hover]:[color:var(--foreground)]
    [&.active]:[color:var(--foreground)] [&.active]:[background:var(--muted)] [&.active]:[border-color:var(--border)] [&.active]:[font-weight:550] [&.active]:[box-shadow:none]
    [&.active_svg]:[color:var(--muted-foreground)]
    [&_small]:[font-size:8px] [&_small]:[letter-spacing:0.7px] [&_small]:[padding:2px_4px] [&_small]:[border:1px_solid_var(--border)] [&_small]:[border-radius:4px] [&_small]:[color:var(--muted-foreground)] [&_small]:[margin-left:auto]
    min-[1500px]:[height:47px]
  `,
  "sidebar-bottom": String.raw`
    [margin-top:auto] [padding-top:26px]
  `,
  "workspace-note": String.raw`
    [padding:20px_10px_25px] [color:var(--muted-foreground)] [font-size:11px] [line-height:1.8]
    [&_strong]:font-normal [&_strong]:[color:var(--muted-foreground)]
  `,
  "workspace-dot": String.raw`
    inline-block [background:var(--muted)] [width:5px] [height:5px] [border-radius:50%] [margin-right:5px]
  `,
  "workspace-user": String.raw`
    flex [gap:10px] items-center [width:100%] [border-top:1px_solid_var(--border)] [padding:20px_4px_0] text-left
    [&_strong]:block [&_strong]:[font-size:13px] [&_strong]:[font-weight:550]
    [&_small]:block [&_small]:[color:var(--muted-foreground)] [&_small]:[font-size:11px] [&_small]:[margin-top:3px]
    [&_>_svg]:[margin-left:auto] [&_>_svg]:[color:var(--muted-foreground)]
  `,
  "user-avatar": String.raw`
    [width:38px] [height:38px] [border-radius:50%] [background:var(--muted)] [color:var(--foreground)] grid [place-items:center] [font-size:12px] shrink-0
  `,
  "product-main": String.raw`
    [min-width:0] [flex:1] flex flex-col [height:100%] [min-height:0] overflow-hidden
  `,
  "product-topbar": String.raw`
    [height:83px] flex items-center justify-between [border-bottom:1px_solid_var(--border)] [padding:0_28px] [gap:16px] shrink-0 [min-height:83px]
    max-[780px]:[height:65px] max-[780px]:[padding:0_18px] max-[780px]:[min-height:65px]
    max-[780px]:[&_.breadcrumb-label]:[font-size:12px]
  `,
  "breadcrumb-label": String.raw`
    [font-size:13px] [color:var(--foreground)]
    max-[780px]:[margin-right:auto]
  `,
  "topbar-actions": String.raw`
    flex items-center [gap:14px]
  `,
  "workspace-search": String.raw`
    [width:292px] [height:42px] [border:1px_solid_var(--border)] [border-radius:11px] flex items-center [gap:10px] [padding:0_13px] [background:var(--card)] [color:var(--muted-foreground)] [font-size:12px] [box-shadow:none]
    [&_kbd]:[margin-left:auto] [&_kbd]:[font-size:10px] [&_kbd]:[border:1px_solid_var(--border)] [&_kbd]:[padding:2px_4px] [&_kbd]:[border-radius:4px]
    max-[1100px]:[width:230px]
    max-[780px]:[width:38px] max-[780px]:[padding:0] max-[780px]:justify-center max-[780px]:[background:transparent] max-[780px]:[border:0]
    max-[780px]:[&_span]:hidden
    max-[780px]:[&_kbd]:hidden
  `,
  "topbar-quick": String.raw`
    flex items-center [gap:8px] [background:var(--muted)] [height:42px] [border:1px_solid_var(--border)] [border-radius:10px] [padding:0_15px] [font-size:12px] [color:var(--muted-foreground)]
    max-[1100px]:hidden
  `,
  "product-content": String.raw`
    [padding:27px_28px_22px] [flex:1] [min-width:0] [min-height:0] overflow-y-auto overflow-x-hidden [overscroll-behavior:contain] [max-height:calc(100dvh_-_83px)]
    min-[1500px]:[padding:30px]
    max-[1100px]:[padding:23px_20px]
    max-[780px]:[padding:22px_16px] max-[780px]:[max-height:calc(100dvh_-_65px)]
    [&_[data-slot="card"]]:[border-radius:12px] [&_[data-slot="card"]]:[border-color:var(--border)] [&_[data-slot="card"]]:[background-color:var(--card)] [&_[data-slot="card"]]:[box-shadow:none]
    [&_[data-slot="card"]_h3]:[font-size:17px] [&_[data-slot="card"]_h3]:[font-weight:550] [&_[data-slot="card"]_h3]:[color:var(--foreground)] [&_[data-slot="card"]_h3]:[letter-spacing:-.35px]
    [&_textarea]:[border-radius:10px] [&_textarea]:[border-color:var(--border)]
    [&_[role="combobox"]]:[border-radius:9px] [&_[role="combobox"]]:[border-color:var(--border)] [&_[role="combobox"]]:[background-color:var(--card)] [&_[role="combobox"]]:[font-size:12px]
    [&_[data-slot="tabs-list"]]:[height:auto] [&_[data-slot="tabs-list"]]:[padding:4px] [&_[data-slot="tabs-list"]]:[border:1px_solid_var(--border)] [&_[data-slot="tabs-list"]]:[border-radius:11px] [&_[data-slot="tabs-list"]]:[gap:3px] [&_[data-slot="tabs-list"]]:[background:var(--muted)] [&_[data-slot="tabs-list"]]:[width:fit-content] [&_[data-slot="tabs-list"]]:[max-width:100%] [&_[data-slot="tabs-list"]]:justify-start [&_[data-slot="tabs-list"]]:overflow-x-auto [&_[data-slot="tabs-list"]]:[margin-bottom:18px]
    [&_[data-slot="tabs-trigger"]]:[height:33px] [&_[data-slot="tabs-trigger"]]:[padding:6px_13px] [&_[data-slot="tabs-trigger"]]:[border:0] [&_[data-slot="tabs-trigger"]]:[border-radius:8px] [&_[data-slot="tabs-trigger"]]:[font-size:12px] [&_[data-slot="tabs-trigger"]]:[font-weight:450] [&_[data-slot="tabs-trigger"]]:[color:var(--muted-foreground)]
    [&_[data-slot="tabs-trigger"]::after]:hidden
    [&_[data-slot="tabs-trigger"][data-state="active"]]:[color:var(--foreground)] [&_[data-slot="tabs-trigger"][data-state="active"]]:[background:var(--card)] [&_[data-slot="tabs-trigger"][data-state="active"]]:[box-shadow:none]
    [&_>_.workspace-page_>_.workspace-page-body_>_.grid_>_[class*="border-muted"]]:[background:var(--card)] [&_>_.workspace-page_>_.workspace-page-body_>_.grid_>_[class*="border-muted"]]:[border-color:var(--border)] [&_>_.workspace-page_>_.workspace-page-body_>_.grid_>_[class*="border-muted"]]:[border-radius:12px]
    [&_>_div_>_[role="tabpanel"]_>_[class*="border-muted"]]:[background:var(--card)] [&_>_div_>_[role="tabpanel"]_>_[class*="border-muted"]]:[border-color:var(--border)] [&_>_div_>_[role="tabpanel"]_>_[class*="border-muted"]]:[border-radius:12px]
    [&_[role="tabpanel"]_>_.grid_>_[class*="border-muted"]]:[background:var(--card)] [&_[role="tabpanel"]_>_.grid_>_[class*="border-muted"]]:[border-color:var(--border)] [&_[role="tabpanel"]_>_.grid_>_[class*="border-muted"]]:[border-radius:12px]
    [&_>_[class*="px-"]]:[padding:0]
    [&_>_[class*="p-6"]]:[padding:0]
    [&_>_[class*="p-4"]]:[padding:0]
    [&_>_div_>_.px-8]:[padding-left:0] [&_>_div_>_.px-8]:[padding-right:0]
    [&_.recharts-cartesian-axis-tick-value]:[font-family:var(--font-geist-sans),Arial,sans-serif] [&_.recharts-cartesian-axis-tick-value]:[fill:var(--muted-foreground)] [&_.recharts-cartesian-axis-tick-value]:[font-size:11px]
    [&_.recharts-cartesian-grid_line]:[stroke:var(--border)]
    max-[780px]:[&_[data-slot="tabs-list"]]:[width:100%]
    max-[780px]:[&_[role="tabpanel"]_>_.grid_>_[class*="col-span-"]]:[grid-column:1_/_-1]
  `,
  "product-heading": String.raw`
    flex items-center justify-between gap-4 mb-7
    [&_h1]:text-[26px] [&_h1]:leading-tight [&_h1]:font-medium [&_h1]:tracking-[-0.8px] [&_h1]:text-foreground
    [&_p]:text-sm [&_p]:text-muted-foreground [&_p]:mt-2 [&_p]:leading-6
    max-[780px]:items-start max-[780px]:flex-wrap max-[780px]:mb-6 max-[780px]:[&_h1]:text-2xl
  `,
  "product-primary": String.raw`
    !bg-primary !text-primary-foreground !border-transparent !rounded-lg !h-9 !px-3.5 !font-medium !shadow-none
  `,
  "metrics-grid": String.raw`
    grid grid-cols-4 gap-0 mb-8 rounded-xl border border-border overflow-hidden max-[780px]:grid-cols-2
  `,
  "metric-card": String.raw`
    bg-transparent min-w-0 p-5 border-r border-border last:border-r-0 max-[780px]:p-4 max-[780px]:border-b max-[780px]:[&:nth-child(2n)]:border-r-0 max-[780px]:[&:nth-last-child(-n+2)]:border-b-0
  `,
  "metric-label": String.raw`
    flex items-center gap-2 text-muted-foreground text-xs
  `,
  "metric-icon": String.raw`
    inline-flex items-center justify-center text-muted-foreground [&_svg]:size-3.5 [&_svg]:stroke-[1.5]
  `,
  "metric-value": String.raw`
    flex items-baseline flex-wrap gap-2 mt-4 [&_strong]:text-[28px] [&_strong]:font-medium [&_strong]:tracking-tight [&_strong]:tabular-nums [&_>_span]:text-[11px] [&_>_span]:text-muted-foreground
  `,
  "product-panel": String.raw`
    bg-card border border-border rounded-xl p-5 shadow-none max-[780px]:p-4
  `,
  "panel-toolbar": String.raw`
    flex items-center justify-between [gap:12px] [margin-bottom:22px]
    [&_h2]:[font-size:19px] [&_h2]:[font-weight:550] [&_h2]:[letter-spacing:-0.4px]
    [&_p]:[color:var(--muted-foreground)] [&_p]:[font-size:12px] [&_p]:[margin-top:3px]
    max-[780px]:flex-wrap
  `,
  "filter-tabs": String.raw`
    flex min-w-0 max-w-full overflow-x-auto [background:var(--muted)] [border-radius:9px] [padding:3px] [gap:2px]
    [&_button]:[padding:6px_11px] [&_button]:[border-radius:7px] [&_button]:[font-size:12px] [&_button]:[color:var(--muted-foreground)] [&_button]:whitespace-nowrap
    [&_button[aria-selected="true"]]:[background:var(--card)] [&_button[aria-selected="true"]]:[box-shadow:none] [&_button[aria-selected="true"]]:[color:var(--foreground)]
  `,
  "forms-grid": String.raw`
    grid [grid-template-columns:repeat(2,_minmax(0,_1fr))] [gap:16px]
    max-[780px]:[grid-template-columns:1fr]
  `,
  "form-card": String.raw`
    p-5 border border-border rounded-xl bg-card transition-colors hover:border-input [&_h3]:text-base [&_h3]:font-medium [&_h3]:tracking-tight [&_h3]:mb-2
  `,
  "form-card-top": String.raw`
    flex items-center justify-between [margin-bottom:20px] [gap:8px]
  `,
  "form-symbol": String.raw`
    grid place-items-center size-8 bg-muted border border-border rounded-lg text-muted-foreground [&_svg]:size-4 [&_svg]:stroke-[1.5]
  `,
  "card-actions": String.raw`
    flex [gap:8px] items-center
    max-[1100px]:[gap:4px]
  `,
  "icon-button": String.raw`
    inline-flex items-center justify-center bg-transparent border border-transparent rounded-md size-8 text-muted-foreground hover:bg-muted hover:text-foreground [&_svg]:size-4 [&_svg]:stroke-[1.5]
  `,
  "form-description": String.raw`
    [font-size:12px] [color:var(--muted-foreground)] [min-height:32px] [line-height:1.6]
    min-[1500px]:[min-height:24px]
  `,
  "form-stats": String.raw`
    grid grid-cols-2 gap-0 mt-5 border-t border-border [&_>_div]:pt-4 [&_>_div:last-child]:pl-4 [&_>_div:last-child]:border-l [&_>_div:last-child]:border-border [&_span]:block [&_span]:text-xs [&_span]:text-muted-foreground [&_span]:mb-1 [&_strong]:text-2xl [&_strong]:font-medium [&_strong]:tracking-tight
  `,
  "status-pill": String.raw`
    inline-flex items-center [gap:5px] [font-size:11px] [padding:4px_8px] whitespace-nowrap [border-radius:5px] [background:var(--muted)] [color:var(--muted-foreground)]
    [&_i]:[height:5px] [&_i]:[width:5px] [&_i]:[border-radius:50%] [&_i]:[background:currentColor]
  `,
  "status-published": String.raw`
    [background:var(--success)] [color:var(--success-foreground)]
  `,
  "status-active": String.raw`
    [background:var(--success)] [color:var(--success-foreground)]
  `,
  "status-completed": String.raw`
    [background:var(--success)] [color:var(--success-foreground)]
  `,
  "status-draft": String.raw`
    bg-muted text-muted-foreground
  `,
  "status-archived": String.raw`
    bg-muted text-muted-foreground
  `,
  "status-paused": String.raw`
    bg-muted text-muted-foreground
  `,
  "status-scheduled": String.raw`
    [background:var(--info)] [color:var(--info-foreground)]
  `,
  "product-empty": String.raw`
    [padding:65px_20px] flex items-center flex-col text-center [gap:13px] [color:var(--muted-foreground)]
    [&_h3]:[font-size:19px] [&_h3]:[color:var(--foreground)] [&_h3]:font-medium
    [&_p]:[max-width:380px] [&_p]:[font-size:13px] [&_p]:[line-height:1.7]
  `,
  "empty-icon": String.raw`
    [background:var(--muted)] [padding:17px] [border-radius:12px] [color:var(--muted-foreground)] [margin-bottom:4px]
  `,
  "product-modal": String.raw`
    bg-popover rounded-2xl overflow-y-auto font-sans [&_h2]:tracking-tight [&_h2]:text-xl [&_h2]:font-medium
  `,
  "product-field": String.raw`
    flex flex-col [gap:7px] [font-size:12px] [color:var(--muted-foreground)]
    [&_>_span]:[font-weight:550]
    [&_small]:font-normal [&_small]:[color:var(--muted-foreground)] [&_small]:[font-size:11px] [&_small]:[line-height:1.5]
    [&_input]:[width:100%] [&_input]:[border:1px_solid_var(--border)] [&_input]:[border-radius:9px] [&_input]:[min-height:41px] [&_input]:[padding:9px_12px] [&_input]:[background:var(--card)] [&_input]:[color:var(--foreground)] [&_input]:[outline:none] [&_input]:[font-size:13px]
    [&_select]:[width:100%] [&_select]:[border:1px_solid_var(--border)] [&_select]:[border-radius:9px] [&_select]:[min-height:41px] [&_select]:[padding:9px_12px] [&_select]:[background:var(--card)] [&_select]:[color:var(--foreground)] [&_select]:[outline:none] [&_select]:[font-size:13px]
    [&_textarea]:[width:100%] [&_textarea]:[border:1px_solid_var(--border)] [&_textarea]:[border-radius:9px] [&_textarea]:[min-height:41px] [&_textarea]:[padding:9px_12px] [&_textarea]:[background:var(--card)] [&_textarea]:[color:var(--foreground)] [&_textarea]:[outline:none] [&_textarea]:[font-size:13px] [&_textarea]:resize-y
    [&_input:focus]:[border-color:var(--ring)] [&_input:focus-visible]:[outline:2px_solid_var(--ring)] [&_input:focus-visible]:[outline-offset:2px]
    [&_select:focus]:[border-color:var(--ring)] [&_select:focus-visible]:[outline:2px_solid_var(--ring)] [&_select:focus-visible]:[outline-offset:2px]
    [&_textarea:focus]:[border-color:var(--ring)] [&_textarea:focus-visible]:[outline:2px_solid_var(--ring)] [&_textarea:focus-visible]:[outline-offset:2px]
  `,
  "product-input": String.raw`
    [width:100%] [border:1px_solid_var(--border)] [border-radius:9px] [min-height:41px] [padding:9px_12px] [background:var(--card)] [color:var(--foreground)] [outline:none] [font-size:13px]
    [&:focus]:[border-color:var(--ring)] [&:focus-visible]:[outline:2px_solid_var(--ring)] [&:focus-visible]:[outline-offset:2px]
  `,
  "product-form": String.raw`
    flex flex-col [gap:18px]
  `,
  "form-row": String.raw`
    grid [grid-template-columns:1fr_1fr] [gap:15px]
    max-[780px]:[grid-template-columns:1fr]
  `,
  "modal-actions": String.raw`
    flex justify-end [gap:10px] [padding-top:16px] [border-top:1px_solid_var(--border)]
  `,
  "product-error": String.raw`
    [color:var(--muted-foreground)] [background:var(--muted)] [border:1px_solid_var(--border)] [border-radius:8px] [padding:11px_13px] [font-size:12px]
  `,
  "editor-columns": String.raw`
    grid [grid-template-columns:1fr_1fr] [gap:28px]
    max-[780px]:[grid-template-columns:1fr]
  `,
  "editor-preview": String.raw`
    [background:var(--muted)] [border:1px_solid_var(--border)] [border-radius:14px] overflow-hidden [min-height:430px]
    max-[780px]:[min-height:300px]
  `,
  "editor-preview-label": String.raw`
    [padding:12px_18px] [font-size:10px] uppercase [letter-spacing:1.5px] [color:var(--muted-foreground)] [border-bottom:1px_solid_var(--border)]
  `,
  "email-preview": String.raw`
    [width:100%] [height:480px] [border:0] [background:#fff]
  `,
  "lead-form-preview": String.raw`
    [padding:28px]
    [&_h2]:[font-size:24px] [&_h2]:[line-height:1.25] [&_h2]:[color:var(--foreground)] [&_h2]:[letter-spacing:-0.7px] [&_h2]:[margin-bottom:10px]
    [&_p]:[font-size:13px] [&_p]:[line-height:1.7] [&_p]:[color:var(--muted-foreground)] [&_p]:[margin-bottom:24px]
  `,
  "form-field-row": String.raw`
    grid [grid-template-columns:1fr_1fr_auto] items-center [gap:8px] [margin-top:8px]
    [&_input]:[min-width:0]
  `,
  "consent-row": String.raw`
    flex items-start [gap:8px] [font-size:11px] [line-height:1.65] [color:var(--muted-foreground)]
    [&_input]:[margin-top:4px] [&_input]:[accent-color:var(--ring)]
  `,
  "product-footer": String.raw`
    flex justify-between [padding:13px_30px_17px] [font-size:10px] [color:var(--muted-foreground)] [gap:12px]
    [&_>_span:last-child]:flex [&_>_span:last-child]:[gap:5px] [&_>_span:last-child]:items-center
    max-[780px]:[padding:18px] max-[780px]:[font-size:9px]
    max-[780px]:[&_>_span:last-child]:hidden
  `,
  "search-results": String.raw`
    flex flex-col [gap:4px]
    [&_button]:flex [&_button]:items-center [&_button]:[gap:12px] [&_button]:[padding:13px] [&_button]:[border-radius:9px] [&_button]:text-left
    [&_button:hover]:[background:var(--muted)]
    [&_button_>_svg:last-child]:[margin-left:auto]
  `,
  "mobile-menu": String.raw`
    hidden
    max-[780px]:inline-flex
  `,
  "eyebrow": String.raw`
    [color:var(--muted-foreground)] [font-size:10px] [letter-spacing:1.5px] uppercase block [margin-bottom:10px]
  `,
  "table-wrap": String.raw`
    overflow-auto
  `,
  "product-table": String.raw`
    w-full border-collapse text-left text-[13px]
    [&_th]:text-muted-foreground [&_th]:font-normal [&_th]:px-4 [&_th]:py-3 [&_th]:text-xs [&_th]:border-b [&_th]:border-border
    [&_td]:px-4 [&_td]:py-4 [&_td]:border-b [&_td]:border-border [&_td]:text-muted-foreground
    [&_tr:last-child_td]:border-b-0 [&_td_strong]:text-foreground [&_td_strong]:font-medium [&_td_strong]:block [&_td_strong]:text-[13px]
    [&_td_small]:block [&_td_small]:mt-1 [&_td_small]:text-muted-foreground [&_td_small]:text-xs [&_tbody_tr:hover]:bg-muted
    max-[780px]:[&_td]:px-3 max-[780px]:[&_th]:px-3
  `,
  "template-grid": String.raw`
    grid [grid-template-columns:repeat(3,_minmax(0,_1fr))] [gap:18px]
    max-[1100px]:[grid-template-columns:repeat(2,_minmax(0,_1fr))]
    max-[780px]:[grid-template-columns:1fr_1fr]
  `,
  "template-card": String.raw`
    [border:1px_solid_var(--border)] [border-radius:12px] overflow-hidden [background:var(--card)]
  `,
  "template-art": String.raw`
    [height:220px] [background:var(--muted)] [padding:22px_33px_0] overflow-hidden relative
    max-[780px]:[padding:15px_15px_0] max-[780px]:[height:180px]
  `,
  "mini-email": String.raw`
    [height:245px] [background:#fff] [box-shadow:0_4px_20px_#34234310] [border-radius:6px] [padding:22px] text-left
    [&_small]:[font-size:7px] [&_small]:[letter-spacing:2px]
    [&_h3]:[font-size:22px] [&_h3]:[line-height:1.15] [&_h3]:[letter-spacing:-0.7px] [&_h3]:[margin:17px_0_15px] [&_h3]:font-semibold
    [&_.line]:[height:4px] [&_.line]:[margin:7px_0] [&_.line]:[background:#eeebf3] [&_.line]:[border-radius:3px]
    [&_.color-block]:[height:48px] [&_.color-block]:[margin-top:17px] [&_.color-block]:[border-radius:5px] [&_.color-block]:[opacity:0.14]
    max-[780px]:[padding:15px]
    max-[780px]:[&_h3]:[font-size:18px]
  `,
  "template-info": String.raw`
    [padding:18px]
    [&_h3]:[font-size:15px] [&_h3]:[font-weight:550] [&_h3]:[color:var(--foreground)]
    [&_p]:[font-size:12px] [&_p]:[line-height:1.6] [&_p]:[color:var(--muted-foreground)] [&_p]:[margin:6px_0_16px]
  `,
  "newsletter-title": String.raw`
    flex [gap:11px] items-center
    [&_>_span]:[width:36px] [&_>_span]:[height:40px] [&_>_span]:[background:var(--muted)] [&_>_span]:[border-radius:9px] [&_>_span]:grid [&_>_span]:[place-items:center] [&_>_span]:[color:var(--muted-foreground)]
  `,
  "sidebar-collapsed": String.raw`
    [&_.product-sidebar]:[width:84px] [&_.product-sidebar]:[padding:28px_14px]
    [&_.collapse-sidebar]:hidden
    [&_.brand]:[margin-left:8px]
    [&_.product-nav-item]:justify-center [&_.product-nav-item]:[padding:0]
  `,
  "mobile-sidebar": String.raw`
    flex flex-col [background:var(--card)] ![width:260px] [padding:30px_20px]
    max-[780px]:overflow-y-auto
  `,
  "public-form-page": String.raw`
    [min-height:100svh] [background:#f5f2fa] flex items-center justify-center [padding:24px]
  `,
  "public-form-card": String.raw`
    [max-width:480px] [width:100%] [padding:40px] [background:white] [border:1px_solid_#e7dfef] [border-radius:22px] [box-shadow:0_20px_70px_#38234908]
    [&_h1]:[font-size:28px] [&_h1]:[letter-spacing:-1px] [&_h1]:[font-weight:550]
    [&_>_p]:[color:#8f809c] [&_>_p]:[font-size:14px] [&_>_p]:[line-height:1.7] [&_>_p]:[margin:12px_0_24px]
  `,
  "crm-stage": String.raw`
    [font-size:9px] [letter-spacing:0.5px] [padding:5px_8px] [border-radius:5px] [background:var(--muted)] [color:var(--muted-foreground)]
  `,
  "stage-qualified": String.raw`
    [background:var(--warning)] [color:var(--warning-foreground)]
  `,
  "stage-customer": String.raw`
    [background:var(--success)] [color:var(--success-foreground)]
  `,
  "stage-lost": String.raw`
    bg-destructive/10 text-destructive
  `,
  "crm-pipeline": String.raw`
    grid [grid-template-columns:repeat(4,_minmax(180px,_1fr))] [gap:13px] overflow-auto
  `,
  "pipeline-column": String.raw`
    [border-radius:12px] [background:var(--muted)] [padding:12px] [min-height:330px]
    [&_h3]:flex [&_h3]:items-center [&_h3]:[gap:7px] [&_h3]:[font-size:12px] [&_h3]:[color:var(--muted-foreground)] [&_h3]:[margin:3px_0_17px]
    [&_h3_small]:[margin-left:auto] [&_h3_small]:[background:var(--muted)] [&_h3_small]:[padding:2px_6px] [&_h3_small]:[border-radius:4px]
  `,
  "stage-dot": String.raw`
    [height:6px] [width:6px] [border-radius:50%] [background:var(--muted)]
  `,
  "pipeline-contact": String.raw`
    block text-left [width:100%] [padding:15px] [background:var(--card)] [border:1px_solid_var(--border)] [border-radius:10px] [margin:8px_0] [box-shadow:none]
    [&_strong]:block [&_strong]:[font-size:12px] [&_strong]:[font-weight:550]
    [&_>_span]:[font-size:11px] [&_>_span]:[color:var(--muted-foreground)] [&_>_span]:block [&_>_span]:[margin:6px_0]
    [&_small]:flex [&_small]:items-center [&_small]:[gap:5px] [&_small]:[font-size:9px] [&_small]:[color:var(--muted-foreground)] [&_small]:[margin-top:17px]
  `,
  "contact-note": String.raw`
    [background:var(--muted)] [border-radius:10px] [padding:13px_15px] [margin-top:12px]
    [&_p]:whitespace-pre-wrap [&_p]:[font-size:12px] [&_p]:[line-height:1.7] [&_p]:[color:var(--muted-foreground)]
    [&_small]:[font-size:10px] [&_small]:[color:var(--muted-foreground)] [&_small]:block [&_small]:[margin-top:8px]
  `,
  "automation-starters": String.raw`
    grid [grid-template-columns:1fr_1fr] [gap:18px] [margin-bottom:24px]
    [&_>_button]:flex [&_>_button]:items-center [&_>_button]:[gap:16px] [&_>_button]:[padding:24px] [&_>_button]:[background:var(--muted)] [&_>_button]:[border:1px_solid_var(--border)] [&_>_button]:[border-radius:12px] [&_>_button]:text-left
    [&_strong]:[font-size:14px] [&_strong]:[font-weight:550] [&_strong]:[letter-spacing:-0.2px] [&_strong]:[color:var(--muted-foreground)]
    [&_p]:[font-size:12px] [&_p]:[color:var(--muted-foreground)] [&_p]:[margin-top:5px]
    [&_>_button_>_svg]:[margin-left:auto] [&_>_button_>_svg]:[width:18px] [&_>_button_>_svg]:[color:var(--muted-foreground)]
    max-[900px]:[grid-template-columns:1fr]
    max-[600px]:[&_>_button]:[padding:18px]
  `,
  "starter-icon": String.raw`
    grid [place-items:center] shrink-0 [width:46px] [height:46px] [border-radius:13px] [background:var(--muted)] [color:var(--muted-foreground)]
    [&.peach]:[background:var(--muted)] [&.peach]:[color:var(--muted-foreground)]
  `,
  "automation-editor": String.raw`
    h-full min-h-0 flex flex-col max-[600px]:h-auto max-[600px]:min-h-full
  `,
  "builder-toolbar": String.raw`
    flex items-center [gap:11px] [padding:17px_20px] [border-bottom:1px_solid_var(--border)] [background:var(--card)]
    [&_>_input]:[max-width:210px] [&_>_input]:[min-width:100px] [&_>_input]:[flex:1] [&_>_input]:[background:transparent] [&_>_input]:[border:0] [&_>_input]:[outline:none] [&_>_input]:[color:var(--foreground)] [&_>_input]:[font-size:14px] [&_>_input]:[font-weight:550]
    max-[1200px]:flex-wrap
    max-[600px]:[padding:13px]
  `,
  "builder-toolbar-actions": String.raw`
    flex [gap:8px] [margin-left:auto]
    max-[1200px]:[gap:5px]
    max-[1200px]:[&_button]:[font-size:10px] max-[1200px]:[&_button]:[padding:7px_10px]
    max-[600px]:[width:100%] max-[600px]:justify-end max-[600px]:flex-wrap
  `,
  "save-state": String.raw`
    [font-size:10px] [color:var(--muted-foreground)]
    max-[1200px]:hidden
  `,
  "builder-workspace": String.raw`
    flex min-h-0 flex-1 overflow-hidden
    max-[600px]:flex-wrap max-[600px]:[min-height:0] max-[600px]:[height:auto]
  `,
  "step-palette": String.raw`
    [padding:25px_17px] [width:225px] overflow-y-auto [background:var(--card)] [border-right:1px_solid_var(--border)] shrink-0
    [&_h3]:[font-size:17px] [&_h3]:[letter-spacing:-0.5px] [&_h3]:[color:var(--muted-foreground)] [&_h3]:font-medium
    [&_>_p]:[font-size:11px] [&_>_p]:[line-height:1.7] [&_>_p]:[color:var(--muted-foreground)] [&_>_p]:[margin:8px_0_23px]
    [&_>_button]:flex [&_>_button]:[width:100%] [&_>_button]:items-center [&_>_button]:text-left [&_>_button]:[gap:9px] [&_>_button]:[padding:11px_0] [&_>_button]:[border-bottom:1px_solid_var(--border)]
    [&_>_button_strong]:[font-size:12px] [&_>_button_strong]:[font-weight:550] [&_>_button_strong]:block [&_>_button_strong]:[color:var(--muted-foreground)]
    [&_>_button_small]:block [&_>_button_small]:[font-size:10px] [&_>_button_small]:[margin-top:4px] [&_>_button_small]:[color:var(--muted-foreground)]
    [&_>_button_>_svg]:[margin-left:auto] [&_>_button_>_svg]:[color:var(--muted-foreground)]
    max-[1200px]:[width:190px] max-[1200px]:[padding:20px_13px]
    max-[900px]:[width:65px] max-[900px]:[padding:18px_12px]
    max-[900px]:[&_.eyebrow]:hidden
    max-[900px]:[&_>_input]:hidden
    max-[900px]:[&_h3]:hidden
    max-[900px]:[&_>_p]:hidden
    max-[900px]:[&_button_>_span:nth-child(2)]:hidden
    max-[900px]:[&_button_>_svg]:hidden
    max-[600px]:[width:55px]
  `,
  "step-icon": String.raw`
    [width:36px] [height:36px] grid [place-items:center] shrink-0 [border-radius:9px] [background:var(--muted)] [color:var(--muted-foreground)]
  `,
  "step-start": String.raw`
    [color:var(--muted-foreground)] [background:var(--muted)]
  `,
  "step-email": String.raw`
    [color:var(--muted-foreground)] [background:var(--muted)]
  `,
  "step-wait": String.raw`
    [color:var(--muted-foreground)] [background:var(--muted)]
  `,
  "step-condition": String.raw`
    [color:var(--muted-foreground)] [background:var(--muted)]
  `,
  "step-exit": String.raw`
    [color:var(--muted-foreground)] [background:var(--muted)]
  `,
  "palette-hint": String.raw`
    flex items-start [gap:9px] [margin-top:35px] [color:var(--muted-foreground)] [padding:13px] [border:1px_dashed_var(--border)] [border-radius:10px]
    [&_svg]:shrink-0
    [&_p]:[font-size:10px] [&_p]:[line-height:1.7]
    max-[900px]:hidden
  `,
  "workflow-canvas": String.raw`
    [&_.react-flow\_\_handle]:[background:var(--muted-foreground)] [&_.react-flow\_\_handle]:[border:2px_solid_var(--border)] [&_.react-flow\_\_handle]:[width:8px] [&_.react-flow\_\_handle]:[height:8px] [&_.react-flow\_\_controls]:![box-shadow:none] [&_.react-flow\_\_controls]:[border:1px_solid_var(--border)] [&_.react-flow\_\_controls]:[border-radius:7px] [&_.react-flow\_\_controls]:overflow-hidden [&_.react-flow\_\_controls-button]:![border-bottom:1px_solid_var(--border)] [&_.react-flow\_\_controls-button]:![background:var(--card)] [&_.react-flow\_\_controls-button]:![fill:var(--muted-foreground)]
    [min-width:0] [flex:1] [background:var(--muted)] relative min-h-0
    max-[600px]:[min-height:440px] max-[600px]:[flex:1] max-[600px]:[height:440px]
  `,
  "canvas-label": String.raw`
    absolute [top:20px] [left:20px] [z-index:1] [color:var(--muted-foreground)] flex items-center [gap:7px] [font-size:8px] [letter-spacing:1.4px]
  `,
  "workflow-step": String.raw`
    flex items-center [gap:14px] [width:300px] [min-height:88px] [background:var(--card)] [border:1px_solid_var(--border)] [border-radius:13px] [padding:17px] [box-shadow:none] relative
    [&.selected]:[border-color:var(--ring)] [&.selected]:[box-shadow:0_0_0_1px_var(--ring)]
    [&_small]:[font-size:9px] [&_small]:[letter-spacing:1px] [&_small]:[color:var(--muted-foreground)]
    [&_strong]:block [&_strong]:[font-size:14px] [&_strong]:[font-weight:550] [&_strong]:[color:var(--foreground)] [&_strong]:[margin-top:4px]
    [&_p]:[font-size:11px] [&_p]:[color:var(--muted-foreground)] [&_p]:[margin-top:4px] [&_p]:[max-width:195px] [&_p]:whitespace-nowrap [&_p]:overflow-hidden [&_p]:text-ellipsis
    max-[1200px]:[width:275px]
  `,
  "step-chevron": String.raw`
    [color:var(--muted-foreground)] [margin-left:auto]
  `,
  "react-flow__handle": String.raw`
    [background:var(--muted-foreground)] [border:2px_solid_var(--background)] [width:8px] [height:8px]
  `,
  "step-inspector": String.raw`
    [width:270px] shrink-0 [padding:24px_20px] [background:var(--card)] [border-left:1px_solid_var(--border)] flex flex-col overflow-auto
    max-[1200px]:[width:240px] max-[1200px]:[padding:20px_15px]
    max-[900px]:[width:220px]
    max-[600px]:[width:100%] max-[600px]:[border-left:0] max-[600px]:[border-top:1px_solid_var(--border)] max-[600px]:[min-height:200px]
  `,
  "inspector-title": String.raw`
    flex [gap:12px] items-center [padding-bottom:25px] [margin-bottom:20px] [border-bottom:1px_solid_var(--border)]
    [&_small]:[font-size:8px] [&_small]:[letter-spacing:1.1px] [&_small]:[color:var(--muted-foreground)]
    [&_h3]:[font-size:14px] [&_h3]:[color:var(--muted-foreground)] [&_h3]:[font-weight:550] [&_h3]:[margin-top:5px]
  `,
  "inspector-help": String.raw`
    [font-size:11px] [color:var(--muted-foreground)] [line-height:1.9]
  `,
  "inspector-footer": String.raw`
    [margin-top:auto] [border-top:1px_solid_var(--border)] [padding-top:20px] flex [gap:8px] items-start [color:var(--muted-foreground)] [font-size:10px] [line-height:1.7]
    [&_svg]:shrink-0
    max-[600px]:[margin-top:20px]
  `,
  "builder-error": String.raw`
    [border-radius:0] [border-width:0_0_1px] [margin:0]
  `,
  "builder-notice": String.raw`
    [background:var(--muted)] [padding:10px_24px] [font-size:11px] [color:var(--muted-foreground)]
  `,
  "react-flow__controls": String.raw`
    ![box-shadow:none] [border:1px_solid_var(--border)] [border-radius:7px] overflow-hidden
  `,
  "react-flow__controls-button": String.raw`
    ![border-bottom:1px_solid_var(--border)] ![background:var(--card)] ![fill:var(--muted-foreground)]
  `,
  "mail-workspace": String.raw`
    flex h-full w-full min-h-0 min-w-0 overflow-hidden bg-background
  `,
  "mail-folders": String.raw`
    [width:180px] shrink-0 min-h-0 overflow-y-auto [background:var(--card)] [padding:22px_13px] [border-right:1px_solid_var(--border)] flex flex-col
    max-[1300px]:[width:150px] max-[1300px]:[padding:20px_10px]
    max-[1100px]:[width:130px]
    max-[780px]:[width:120px]
  `,
  "mail-workspace-title": String.raw`
    [padding:0_5px_20px]
    [&_strong]:block [&_strong]:[font-size:14px] [&_strong]:[font-weight:550] [&_strong]:[margin:12px_0_5px]
    [&_small]:[font-size:10px] [&_small]:[color:var(--muted-foreground)]
    max-[1100px]:[&_small]:hidden
  `,
  "mail-account-mark": String.raw`
    [width:30px] [height:30px] [border-radius:50%] [background:var(--muted)] [color:var(--foreground)] grid [place-items:center]
  `,
  "compose-button": String.raw`
    ![background:var(--muted)] ![color:var(--foreground)] ![border-radius:8px] ![border:1px_solid_var(--border)] ![box-shadow:none] ![height:38px]
  `,
  "mail-folder-list": String.raw`
    [margin-top:18px] [border-top:1px_solid_var(--border)] [padding-top:14px]
    [&_button]:flex [&_button]:items-center [&_button]:[gap:10px] [&_button]:[width:100%] [&_button]:[padding:12px_10px] [&_button]:[border-radius:8px] [&_button]:[font-size:12px] [&_button]:[color:var(--muted-foreground)] [&_button]:[border:1px_solid_transparent]
    [&_button.active]:[background:var(--card)] [&_button.active]:[border-color:var(--border)] [&_button.active]:[box-shadow:none] [&_button.active]:[color:var(--foreground)] [&_button.active]:[font-weight:550]
    [&_small]:[margin-left:auto]
    max-[780px]:overflow-y-auto
  `,
  "mail-folder-note": String.raw`
    [font-size:10px] [line-height:1.9] [color:var(--muted-foreground)] [margin-top:auto] [padding:16px_5px]
    [&_a]:[color:var(--muted-foreground)]
  `,
  "mail-list-pane": String.raw`
    [width:320px] shrink-0 [border-right:1px_solid_var(--border)] flex flex-col [min-width:0]
    max-[1300px]:[width:285px]
    max-[1100px]:[width:240px]
    max-[780px]:[flex:1] max-[780px]:[width:auto]
  `,
  "mail-list-heading": String.raw`
    [padding:24px_20px_17px] flex items-center justify-between
    [&_h1]:flex [&_h1]:items-center [&_h1]:[gap:8px] [&_h1]:[font-size:18px] [&_h1]:[font-weight:550] [&_h1]:[letter-spacing:-0.4px]
    [&_p]:[font-size:11px] [&_p]:[color:var(--muted-foreground)] [&_p]:[margin-top:7px]
  `,
  "mail-search": String.raw`
    focus-within:[outline:2px_solid_var(--ring)] focus-within:[outline-offset:2px] flex items-center [gap:10px] [border:1px_solid_var(--border)] [border-radius:9px] [margin:0_18px_17px] [padding:11px] [color:var(--muted-foreground)]
    [&_input]:[outline:0] [&_input]:bg-transparent [&_input]:text-foreground [&_input]:border-0 [&_input]:[width:100%] [&_input]:[font-size:12px] [&_input]:[min-width:0]
    [&_kbd]:[font-size:10px]
  `,
  "mail-list-scroll": String.raw`
    overflow-auto [flex:1] [padding:0_14px]
  `,
  "mail-group-label": String.raw`
    flex [gap:8px] items-center [color:var(--muted-foreground)] [font-size:10px] [padding:8px_6px_15px]
  `,
  "mail-list-item": String.raw`
    [width:100%] flex items-start [gap:10px] [padding:18px_7px] text-left [border-bottom:1px_solid_var(--border)] [border-left:2px_solid_transparent] [border-radius:3px]
    [&.selected]:[background:var(--muted)] [&.selected]:[border-left-color:var(--ring)] [&.selected]:[padding-left:9px]
    [&_>_div]:[min-width:0] [&_>_div]:[flex:1]
    [&_p]:[font-size:10px] [&_p]:[margin-top:6px] [&_p]:whitespace-nowrap [&_p]:text-ellipsis [&_p]:overflow-hidden [&_p]:[color:var(--muted-foreground)]
  `,
  "mail-avatar": String.raw`
    [height:31px] [width:31px] grid [place-items:center] [border-radius:50%] shrink-0 [background:var(--muted)] [color:var(--muted-foreground)] [font-size:12px]
  `,
  "mail-item-line": String.raw`
    flex items-center [gap:6px]
    [&_strong]:[font-size:12px] [&_strong]:[font-weight:550] [&_strong]:[color:var(--muted-foreground)] [&_strong]:overflow-hidden [&_strong]:text-ellipsis [&_strong]:whitespace-nowrap
    [&_time]:[margin-left:auto] [&_time]:whitespace-nowrap [&_time]:[font-size:9px] [&_time]:[color:var(--muted-foreground)]
  `,
  "mail-subject": String.raw`
    [font-size:11px] [color:var(--muted-foreground)] [margin-top:7px] block overflow-hidden text-ellipsis whitespace-nowrap
  `,
  "mail-detail-pane": String.raw`
    [flex:1] [min-width:0] overflow-auto
    max-[780px]:hidden
  `,
  "mail-detail-toolbar": String.raw`
    [height:65px] flex items-center [gap:13px] [padding:0_22px] [border-bottom:1px_solid_var(--border)]
    [&_.icon-button]:[border:0] [&_.icon-button]:[box-shadow:none] [&_.icon-button]:[color:var(--muted-foreground)]
    max-[1100px]:[gap:7px]
  `,
  "mail-subject-heading": String.raw`
    [padding:25px_25px_22px] [border-bottom:1px_solid_var(--border)]
    [&_p]:[font-size:11px] [&_p]:[color:var(--muted-foreground)] [&_p]:[margin-bottom:11px]
    [&_h2]:[font-size:23px] [&_h2]:font-medium [&_h2]:[letter-spacing:-0.6px] [&_h2]:[color:var(--foreground)] [&_h2]:[line-height:1.35]
    max-[1300px]:[padding:20px]
    max-[1300px]:[&_h2]:[font-size:20px]
  `,
  "mail-message": String.raw`
    [padding:25px]
    max-[1300px]:[padding:20px]
  `,
  "mail-sender": String.raw`
    flex items-center [gap:12px]
    [&_strong]:[font-size:14px] [&_strong]:font-medium [&_strong]:[color:var(--muted-foreground)]
    [&_p]:[font-size:11px] [&_p]:[color:var(--muted-foreground)] [&_p]:[margin-top:5px]
    [&_time]:[margin-left:auto] [&_time]:[font-size:10px] [&_time]:[color:var(--muted-foreground)]
  `,
  "mail-recipients": String.raw`
    [font-size:10px] [color:var(--muted-foreground)] [margin-top:18px] [line-height:2]
    [&_span]:[border:1px_solid_var(--border)] [&_span]:[border-radius:6px] [&_span]:[padding:4px_7px] [&_span]:[color:var(--muted-foreground)]
  `,
  "remote-images-note": String.raw`
    [font-size:10px] [padding:13px] [background:var(--muted)] [border:1px_solid_var(--border)] [border-radius:8px] [color:var(--muted-foreground)] [margin:25px_0]
    [&_button]:underline [&_button]:[margin-left:8px] [&_button]:[color:var(--muted-foreground)]
  `,
  "mail-body": String.raw`
    [width:100%] [border:0] [min-height:330px]
  `,
  "mail-attachments": String.raw`
    flex flex-wrap [gap:8px]
    [&_a]:flex [&_a]:items-center [&_a]:[gap:7px] [&_a]:[border:1px_solid_var(--border)] [&_a]:[border-radius:8px] [&_a]:[padding:8px_10px] [&_a]:[font-size:10px] [&_a]:[color:var(--muted-foreground)]
  `,
  "mail-empty": String.raw`
    [height:100%] flex items-center justify-center flex-col [gap:14px] [color:var(--muted-foreground)] [background:var(--card)]
    [&_h2]:[font-size:20px] [&_h2]:[font-weight:450] [&_h2]:[letter-spacing:-0.4px] [&_h2]:[color:var(--muted-foreground)]
    [&_p]:[font-size:12px]
  `,
  "mail-back": String.raw`
    !hidden
    max-[780px]:!inline-flex
  `,
  "mail-open": String.raw`
    max-[780px]:[&_.mail-folders]:hidden
    max-[780px]:[&_.mail-list-pane]:hidden
    max-[780px]:[&_.mail-detail-pane]:block
  `,
  "preview-banner": String.raw`
    text-[10px] text-center tracking-[0.04em] text-muted-foreground bg-muted px-3 py-1.5 relative z-[5]
  `,
  "product-nav-settings": String.raw`
    [&_summary]:[list-style:none] [&_summary]:cursor-pointer
    [&_summary::-webkit-details-marker]:hidden
    [&[open]_summary]:[color:var(--muted-foreground)]
    [&_>_a]:[font-size:12px] [&_>_a]:[padding-left:42px] [&_>_a]:[height:35px]
  `,
  "workspace-page": String.raw`
    [min-width:0]
  `,
  "workspace-page-body": String.raw`
    [min-width:0]
  `,
  "workspace-page-heading": String.raw`
    [margin-bottom:26px] [background:transparent]
    max-[780px]:[&_h1]:[font-size:24px]
  `,
  "workspace-heading-actions": String.raw`
    flex items-center flex-wrap [gap:10px]
    max-[780px]:[width:100%]
  `,
  "workspace-back-link": String.raw`
    inline-flex items-center [gap:4px] [color:var(--muted-foreground)] [font-size:11px] [margin-bottom:10px]
  `,
  "workspace-data-table": String.raw`
    border border-border rounded-xl bg-card overflow-hidden [&_>_div:first-child]:border-0 [&_>_div:last-child:not(:first-child)]:px-4 [&_>_div:last-child:not(:first-child)]:py-3 max-[780px]:[&_>_div:last-child]:flex-wrap max-[780px]:[&_>_div:last-child]:gap-3 max-[780px]:[&_>_div:last-child_>_div:last-child]:flex-wrap
  `,
  "metric-period": String.raw`
    [margin:10px_3px_0] [font-size:10px] [color:var(--muted-foreground)]
  `,
  "template-editor-workspace": String.raw`
    [background:var(--card)] [border:1px_solid_var(--border)] [border-radius:12px] overflow-hidden [min-height:600px]
  `,
  "workspace-dialog": String.raw`
    rounded-2xl border-border overflow-y-auto [&_h2]:text-xl [&_h2]:font-medium [&_h2]:tracking-tight [&_h2]:text-foreground
  `,
  "auth-workspace": String.raw`
    [min-height:100svh] [padding:44px_20px] [background:var(--muted)] flex flex-col items-center justify-center
  `,
  "auth-brand": String.raw`
    flex items-center [gap:10px] [color:var(--foreground)] font-semibold [font-size:24px] [margin-bottom:30px]
  `,
  "auth-card": String.raw`
    [width:min(100%,520px)] [background:var(--card)] [border:1px_solid_var(--border)] [border-radius:12px] [padding:12px] [box-shadow:none]
    [&_>_div]:[width:100%]
    [&_h1]:[font-size:28px] [&_h1]:[color:var(--foreground)] [&_h1]:[font-weight:550] [&_h1]:[letter-spacing:-.8px]
    [&_.text-white]:[color:var(--muted-foreground)]
    [&_.text-primary-foreground]:[color:var(--muted-foreground)]
    [&_input]:![color:var(--foreground)] [&_input]:![border:1px_solid_var(--border)] [&_input]:[background:var(--card)]
    [&_[class*="w-[400px]"]]:[width:100%] [&_[class*="w-[400px]"]]:[max-width:400px]
    [&_[class*="w-[300px]"]]:[width:100%] [&_[class*="w-[300px]"]]:[max-width:400px]
    [&_form_button[type="submit"]]:block [&_form_button[type="submit"]]:[width:100%] [&_form_button[type="submit"]]:[min-height:42px] [&_form_button[type="submit"]]:[padding:10px_16px] [&_form_button[type="submit"]]:[border-radius:10px] [&_form_button[type="submit"]]:[background:var(--muted)] [&_form_button[type="submit"]]:[border:1px_solid_var(--border)] [&_form_button[type="submit"]]:[color:var(--foreground)] [&_form_button[type="submit"]]:[font-size:13px]
    [&_a]:[color:var(--muted-foreground)]
    max-[780px]:[padding:0]
  `,
  "auth-footer": String.raw`
    [margin-top:28px] [font-size:11px] [color:var(--muted-foreground)]
  `,
};

/** Expand a shared component recipe into static Tailwind utilities. */
export function workspaceClassName(value: string): string {
  return value.split(/\s+/).filter(Boolean).map(name => recipes[name] ? `${name} ${recipes[name]}` : name).join(" ");
}
