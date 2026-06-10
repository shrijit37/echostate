(function () {
  "use strict";

  /* ------------------------------------------------------------------ */
  /*  Config                                                             */
  /* ------------------------------------------------------------------ */
  const RELAY_URL = "http://localhost:9080/context";
  const POLL_MS = 5000;

  /* ------------------------------------------------------------------ */
  /*  State                                                              */
  /* ------------------------------------------------------------------ */
  let payload = null;
  let online = false;

  /* ------------------------------------------------------------------ */
  /*  DOM references — created once, re-queried if detached              */
  /* ------------------------------------------------------------------ */
  let rootEl = null;   // #echostate-root
  let dotEl = null;    // #echostate-dot
  let btnEl = null;    // #echostate-btn

  /* ------------------------------------------------------------------ */
  /*  Polling                                                            */
  /* ------------------------------------------------------------------ */
  async function poll() {
    try {
      const res = await fetch(RELAY_URL);
      if (!res.ok) {
        setOffline();
        return;
      }
      payload = await res.json();
      setOnline();
    } catch (_) {
      setOffline();
    }
  }

  function setOnline() {
    if (online) return;
    online = true;
    syncUI();
  }
  function setOffline() {
    if (!online && payload === null) return;
    online = false;
    payload = null;
    syncUI();
  }

  /* ------------------------------------------------------------------ */
  /*  UI creation                                                        */
  /* ------------------------------------------------------------------ */
  function ensureUI() {
    if (rootEl && rootEl.isConnected) return;

    // Detached or first time — rebuild
    if (rootEl) rootEl.remove();

    rootEl = document.createElement("div");
    rootEl.id = "echostate-root";

    const label = document.createElement("span");
    label.className = "echostate-label";

    dotEl = document.createElement("span");
    dotEl.id = "echostate-dot";
    dotEl.className = "echostate-offline";
    label.appendChild(dotEl);

    const txt = document.createElement("span");
    txt.id = "echostate-text";
    txt.textContent = "EchoState";
    label.appendChild(txt);

    btnEl = document.createElement("button");
    btnEl.id = "echostate-btn";
    btnEl.textContent = "Inject Context";
    btnEl.addEventListener("click", inject);

    rootEl.appendChild(label);
    rootEl.appendChild(btnEl);
    document.body.appendChild(rootEl);
  }

  /* ------------------------------------------------------------------ */
  /*  UI sync                                                            */
  /* ------------------------------------------------------------------ */
  function syncUI() {
    ensureUI();
    if (!btnEl || !dotEl) return;

    if (online) {
      dotEl.className = "echostate-online";
      btnEl.disabled = false;
      btnEl.className = "echostate-btn-active";
    } else {
      dotEl.className = "echostate-offline";
      btnEl.disabled = true;
      btnEl.className = "echostate-btn-disabled";
    }
  }

  /* ------------------------------------------------------------------ */
  /*  Injection                                                          */
  /* ------------------------------------------------------------------ */
  function inject() {
    if (!payload) return;
    const text = buildPrompt(payload);
    const host = window.location.hostname;
    if (host.includes("chatgpt.com")) injectChatGPT(text);
    else if (host.includes("claude.ai")) injectClaude(text);
  }

  function buildPrompt(data) {
    const parts = [];

    if (data.context) {
      parts.push("=== Context ===", data.context);
    } else if (data.summarized_context) {
      parts.push("=== Summarized Context ===", data.summarized_context);
    } else if (data.raw_context) {
      parts.push("=== Context ===", data.raw_context);
    }

    if (data.file_tree) {
      parts.push("=== File Tree ===", data.file_tree);
    }

    if (data.last_commit) {
      if (typeof data.last_commit === "string") {
        parts.push("=== Last Commit ===", data.last_commit);
      } else if (data.last_commit.hash) {
        const m = data.last_commit.message
          ? " - " + data.last_commit.message
          : "";
        parts.push("=== Last Commit ===", data.last_commit.hash + m);
      }
    }

    return parts.join("\n\n");
  }

  function injectChatGPT(text) {
    const ta = document.getElementById("prompt-textarea");
    if (!ta) return;
    ta.focus();
    if (clipboardPaste(ta, text)) return;

    // Fallback: native value setter + input event
    const setter =
      Object.getOwnPropertyDescriptor(
        window.HTMLTextAreaElement.prototype,
        "value"
      )?.set;
    if (setter) {
      setter.call(ta, text);
      ta.dispatchEvent(new Event("input", { bubbles: true }));
    }
  }

  function injectClaude(text) {
    const editor = document.querySelector(
      'div[contenteditable="true"]'
    );
    if (!editor) return;
    editor.focus();
    if (clipboardPaste(editor, text)) return;

    // Fallback for contenteditable
    editor.textContent = text;
    editor.dispatchEvent(new Event("input", { bubbles: true }));
  }

  function clipboardPaste(el, text) {
    try {
      const dt = new DataTransfer();
      dt.setData("text/plain", text);
      const ev = new ClipboardEvent("paste", {
        clipboardData: dt,
        bubbles: true,
        cancelable: true,
      });
      return el.dispatchEvent(ev);
    } catch (_) {
      return false;
    }
  }

  /* ------------------------------------------------------------------ */
  /*  Boot                                                               */
  /* ------------------------------------------------------------------ */
  function boot() {
    // Mount UI immediately; state will be updated by first poll
    syncUI();
    poll();
    setInterval(poll, POLL_MS);

    // Re-attach UI if the page framework unmounts our container (SPA)
    const mo = new MutationObserver(() => {
      if (rootEl && !rootEl.isConnected) {
        // Container was blown away — re-create
        rootEl = null;
        dotEl = null;
        btnEl = null;
        syncUI();
      }
    });
    mo.observe(document.body, { childList: true, subtree: false });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
