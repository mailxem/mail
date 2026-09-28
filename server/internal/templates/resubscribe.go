package templates

import "fmt"

func ResubscribeTemplate(email string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Resubscribed • Welcome back</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />
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
  <main
    aria-labelledby="title"
    style="
      position:relative;
      width:100%%;
      max-width:480px;
      padding:28px 26px 26px;
      border-radius:22px;
      background:linear-gradient(135deg, rgba(15,23,42,0.97), rgba(15,23,42,0.93));
      border:1px solid rgba(129,140,248,0.7);
      box-shadow:0 22px 45px rgba(15,23,42,0.85);
      overflow:hidden;
    "
  >
    <!-- glow -->
    <div
      aria-hidden="true"
      style="
        position:absolute;
        inset:-1px;
        border-radius:22px;
        background:radial-gradient(circle at 100%% 0, rgba(129,140,248,0.7), transparent 60%%);
        mix-blend-mode:screen;
        opacity:0.6;
        pointer-events:none;
      "
    ></div>

    <!-- badge -->
    <div
      style="
        position:relative;
        display:inline-flex;
        align-items:center;
        gap:8px;
        padding:6px 11px;
        border-radius:999px;
        background:rgba(15,23,42,0.9);
        border:1px solid rgba(129,140,248,0.7);
        font-size:11px;
        letter-spacing:0.08em;
        text-transform:uppercase;
        color:#c7d2fe;
        z-index:1;
      "
    >
      <span
        aria-hidden="true"
        style="
          width:7px;
          height:7px;
          border-radius:999px;
          background:#4f46e5;
          box-shadow:0 0 12px rgba(79,70,229,0.9);
          display:block;
        "
      ></span>
      <span>Subscription restored</span>
    </div>

    <!-- icon -->
    <div
      aria-hidden="true"
      style="
        position:relative;
        margin-top:20px;
        width:72px;
        height:72px;
        border-radius:999px;
        background:radial-gradient(circle at 30%% 20%%, #ffffff33, transparent 55%%);
        border:1px solid rgba(129,140,248,0.9);
        display:flex;
        align-items:center;
        justify-content:center;
        z-index:1;
      "
    >
      <span
        style="
          position:absolute;
          inset:-10px;
          border-radius:inherit;
          border:1px dashed rgba(129,140,248,0.55);
          opacity:0.75;
        "
      ></span>
      <span
        style="
          font-size:30px;
          line-height:1;
          color:#c7d2fe;
          text-shadow:0 0 14px rgba(129,140,248,0.9);
        "
      >
        ★
      </span>
    </div>

    <!-- title -->
    <h1
      id="title"
      style="
        position:relative;
        margin:20px 0 0;
        font-size:26px;
        line-height:1.15;
        font-weight:600;
        letter-spacing:-0.03em;
        color:#eef2ff;
        z-index:1;
      "
    >
      Welcome back to the list
    </h1>

    <!-- subtitle -->
    <p
      style="
        position:relative;
        margin:10px 0 0;
        font-size:14px;
        line-height:1.7;
        color:#9ca3af;
        z-index:1;
      "
    >
      Your email
      <strong style="color:#e5e7eb;">%s</strong>
      has been resubscribed successfully. Expect fresh updates, releases and occasional surprises in your inbox.
    </p>

    <!-- benefits -->
    <div
      aria-hidden="true"
      style="
        position:relative;
        margin-top:22px;
        display:flex;
        flex-wrap:wrap;
        gap:8px;
        z-index:1;
      "
    >
      <span
        style="
          padding:6px 11px;
          border-radius:999px;
          font-size:11px;
          letter-spacing:0.06em;
          text-transform:uppercase;
          background:rgba(34,197,94,0.1);
          border:1px solid rgba(34,197,94,0.35);
          color:#bbf7d0;
        "
      >
        One‑click unsubscribe anytime
      </span>
    </div>

    <!-- divider -->
    <div
      style="
        position:relative;
        margin:22px 0 18px;
        border-top:1px dashed rgba(148,163,184,0.5);
        z-index:1;
      "
    ></div>
    <!-- meta -->
    <div
      style="
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
      "
    >
      <span>You’re back on the main list.</span>
      <span>
        Missed this? Check your
        <span style="color:#e5e7eb;">updates</span> tab.
      </span>
    </div>
  </main>
</body>
</html>
`, email)
}
