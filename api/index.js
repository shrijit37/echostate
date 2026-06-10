const express = require("express");
const cors = require("cors");

const DAEMON_URL = "http://localhost:9081/state";
const PORT = 9080;

const ALLOWED_ORIGINS = [
  "https://chatgpt.com",
  "https://claude.ai",
  "http://localhost:3000",
];

const app = express();

// Strict CORS — only allow specific origins
app.use(
  cors({
    origin(origin, callback) {
      // Allow requests with no origin (server-to-server, curl, etc.)
      if (!origin) return callback(null, true);
      if (ALLOWED_ORIGINS.includes(origin)) {
        return callback(null, true);
      }
      return callback(new Error(`Origin ${origin} not allowed by CORS`));
    },
  })
);

app.get("/context", async (_req, res) => {
  try {
    const daemonResponse = await fetch(DAEMON_URL);
    if (!daemonResponse.ok) {
      return res
        .status(502)
        .json({ error: `daemon responded with ${daemonResponse.status}` });
    }
    const data = await daemonResponse.json();
    res.json(data);
  } catch (err) {
    console.error("context fetch failed:", err.message);
    res.status(502).json({ error: "failed to reach daemon" });
  }
});

app.listen(PORT, () => {
  console.log(`EchoState relay API listening on http://localhost:${PORT}`);
});
