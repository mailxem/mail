package templates

import "fmt"

func UnsubscribeTemplate(email string, emailID string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <title>Unsubscribed • Thank You</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <!-- Optional: web fonts generally still use a <link> -->
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600&display=swap" rel="stylesheet" />
</head>

<body style="
  margin:0;
  padding:32px 16px;
  min-height:100vh;
  display:flex;
  align-items:center;
  justify-content:center;
  background:
    radial-gradient(circle at 10%% 0%%, #22c55e22 0, transparent 55%%),
    radial-gradient(circle at 90%% 100%%, #0ea5e922 0, transparent 55%%),
    radial-gradient(circle at 0%% 100%%, #f9731620 0, transparent 55%%),
    #050816;
  font-family:'Inter', system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  color:#e5e7eb;
  box-sizing:border-box;
">
  <main aria-labelledby="title" style="
      position:relative;
      width:100%%;
      max-width:480px;
      padding:28px 26px 26px;
      border-radius:22px;
      background:linear-gradient(135deg, rgba(15,23,42,0.97), rgba(15,23,42,0.93));
      border:1px solid rgba(148,163,184,0.35);
      box-shadow:0 22px 45px rgba(15,23,42,0.85);
      overflow:hidden;
    ">
    <!-- highlight layer -->
    <div aria-hidden="true" style="
        content:'';
        position:absolute;
        inset:-1px;
        border-radius:22px;
        background:radial-gradient(circle at 0 0, rgba(148,163,184,0.45), transparent 55%%);
        mix-blend-mode:screen;
        opacity:0.6;
        pointer-events:none;
      "></div>

    <!-- badge -->
    <div style="
        position:relative;
        display:inline-flex;
        align-items:center;
        gap:8px;
        padding:6px 11px;
        border-radius:999px;
        background:rgba(15,23,42,0.9);
        border:1px solid rgba(148,163,184,0.45);
        font-size:11px;
        letter-spacing:0.08em;
        text-transform:uppercase;
        color:#9ca3af;
        z-index:1;
      ">
      <span
        aria-hidden="true"
        style="
          width:7px;
          height:7px;
          border-radius:999px;
          background:#22c55e;
          box-shadow:0 0 12px rgba(34,197,94,0.9);
          display:block;
        "
      ></span>
      <span>Preferences updated</span>
    </div>

    <!-- icon -->
    <div aria-hidden="true" style="
        position:relative;
        margin-top:20px;
        width:72px;
        height:72px;
        border-radius:999px;
        background:radial-gradient(circle at 30%% 20%%, #ffffff33, transparent 55%%);
        border:1px solid rgba(34,197,94,0.6);
        display:flex;
        align-items:center;
        justify-content:center;
        z-index:1;
      ">
      <span
        style="
          position:absolute;
          inset:-10px;
          border-radius:inherit;
          border:1px dashed rgba(34,197,94,0.28);
          opacity:0.65;
        "
      ></span>
      <span
        style="
          font-size:32px;
          line-height:1;
          color:#22c55e;
          text-shadow:0 0 14px rgba(34,197,94,0.7);
        "
      >
        ✓
      </span>
    </div>

    <!-- title -->
    <h1 id="title" style="
        position:relative;
        margin:20px 0 0;
        font-size:26px;
        line-height:1.15;
        font-weight:600;
        letter-spacing:-0.03em;
        color:#f9fafb;
        z-index:1;
      ">
      You’re unsubscribed
    </h1>

    <!-- subtitle -->
    <p style="
        position:relative;
        margin:10px 0 0;
        font-size:14px;
        line-height:1.7;
        color:#9ca3af;
        z-index:1;
      ">
      You will no longer receive our marketing emails at
      <strong style="color:#e5e7eb;">%s</strong>. If this was a mistake, you can resubscribe
      in just one click below.
    </p>

    <!-- pills -->
    <div aria-hidden="true" style="
        position:relative;
        margin-top:22px;
        display:flex;
        flex-wrap:wrap;
        gap:8px;
        z-index:1;
      ">
      <span
        style="
          padding:6px 11px;
          border-radius:999px;
          font-size:11px;
          letter-spacing:0.06em;
          text-transform:uppercase;
          background:rgba(34,197,94,0.12);
          border:1px solid rgba(34,197,94,0.32);
          color:#bbf7d0;
        "
      >
        No more newsletters
      </span>
      <span
        style="
          padding:6px 11px;
          border-radius:999px;
          font-size:11px;
          letter-spacing:0.06em;
          text-transform:uppercase;
          background:rgba(15,23,42,0.9);
          border:1px solid rgba(148,163,184,0.4);
          color:#9ca3af;
        "
      >
        Instantly processed
      </span>
    </div>

    <!-- divider -->
    <div style="
        position:relative;
        margin:22px 0 18px;
        border-top:1px dashed rgba(148,163,184,0.5);
        z-index:1;
      "></div>

    <!-- actions -->
    <div style="
        position:relative;
        display:flex;
        flex-wrap:wrap;
        gap:10px;
        z-index:1;
      ">
      <a href="/t/resubscribe?token=%s" style="
          flex:1 1 140px;
          position:relative;
          display:inline-flex;
          align-items:center;
          justify-content:center;
          gap:6px;
          padding:10px 16px;
          border-radius:999px;
          font-size:13px;
          font-weight:500;
          text-decoration:none;
          cursor:pointer;
          border:1px solid transparent;
          background:linear-gradient(135deg, #22c55e, #16a34a);
          color:#022c22;
          box-shadow:0 14px 30px rgba(22,163,74,0.48);
        ">
        <span style="font-size:14px; opacity:0.9;">↻</span>
        <span>Resubscribe</span>
      </a>
    </div>

    <!-- meta -->
    <div style="
        position:relative;
        margin-top:18px;
        font-size:11px;
        color:#9ca3af;
        display:flex;
        flex-wrap:wrap;
        gap:6px;
        align-items:center;
        justify-content:space-between;
        z-index:1;
      ">
      <span>Changes take effect immediately.</span>
    </div>
  </main>
</body>

</html>
`, email, emailID)
}
