/* Isolated interaction prototype. No API, audio, analytics or persistent storage. */
const directions = ["dictionary", "coach", "situations"];
const scenarios = ["new", "returning", "completed", "unavailable"];
const screens = [
  "home",
  "journey",
  "word",
  "saved",
  "recall",
  "write",
  "feedback",
  "complete",
  "progress",
];
const params = new URLSearchParams(location.search);
const state = {
  direction: directions.includes(params.get("direction"))
    ? params.get("direction")
    : "coach",
  scenario: scenarios.includes(params.get("scenario"))
    ? params.get("scenario")
    : "new",
  theme: params.get("theme") === "dark" ? "dark" : "light",
  saved: false,
  assessed: false,
  revealed: false,
  draft: "",
  error: "",
  known: false,
};
const names = {
  dictionary: "Calm dictionary",
  coach: "Warm guided coach",
  situations: "Visual situations",
};
const icons = {
  arrow: '<path d="M5 12h14m-6-6 6 6-6 6"/>',
  home: '<path d="m3 10 9-7 9 7v10h-6v-6H9v6H3z"/>',
  map: '<path d="m3 5 6-2 6 2 6-2v16l-6 2-6-2-6 2zM9 3v16m6-14v16"/>',
  progress: '<path d="M5 20V12m7 8V4m7 16V8"/>',
  check: '<path d="m5 12 4 4L19 6"/>',
  sound:
    '<path d="m11 4-6 5H2v6h3l6 5zM16 8a6 6 0 0 1 0 8m3-11a10 10 0 0 1 0 14"/>',
};
const icon = (name) =>
  `<svg aria-hidden="true" viewBox="0 0 24 24">${icons[name] || icons.arrow}</svg>`;
const escapeText = (value) =>
  value.replace(
    /[&<>"']/gu,
    (ch) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        ch
      ],
  );
const button = (label, action, kind = "primary") =>
  `<button type="button" class="${kind}" data-action="${action}">${label}${kind === "primary" ? icon("arrow") : ""}</button>`;
const link = (label, page, kind = "secondary") =>
  `<a href="#${page}" class="${kind}">${label}</a>`;
const picture = () =>
  '<img class="word-image" src="assets/invite.webp" alt="One person invites another to join a community garden activity." width="768" height="576" />';
const note = (text) =>
  `<div class="coach-note"><span class="note-mark" aria-hidden="true">V</span><p>${text}</p></div>`;
function resetScenario() {
  state.saved =
    state.scenario === "returning" || state.scenario === "completed";
  state.assessed = state.scenario === "completed";
  state.revealed = false;
  state.draft = "";
  state.error = "";
  state.known = false;
}
resetScenario();
function currentScreen() {
  const name = location.hash.slice(1);
  return screens.includes(name) ? name : "home";
}
function go(name) {
  if (currentScreen() === name) render();
  else location.hash = name;
}
function stepper(screen) {
  const index = [
    "word",
    "saved",
    "recall",
    "write",
    "feedback",
    "complete",
  ].indexOf(screen);
  if (index < 0) return "";
  return `<ol class="steps" aria-label="Learning activities">${["Discover", "Remember", "Use it"].map((name, i) => `<li${i === (index < 2 ? 0 : index < 3 ? 1 : 2) ? ' aria-current="step"' : ""}><span class="step-dot">${i + 1}</span>${name}</li>`).join("")}</ol>`;
}
function nav(page) {
  return `<nav class="footer-nav" aria-label="Primary">${[
    ["home", "Home", "home"],
    ["journey", "Journey", "map"],
    ["progress", "Progress", "progress"],
  ]
    .map(
      ([route, name, glyph]) =>
        `<a class="nav-item" href="#${route}"${page === route ? ' aria-current="page"' : ""}>${icon(glyph)}${name}</a>`,
    )
    .join("")}</nav>`;
}
function home() {
  if (state.scenario === "unavailable")
    return `<div class="focused"><p class="eyebrow">TODAY</p><h1 class="page-title" tabindex="-1">Your next step is taking a moment.</h1><p class="intro">We couldn’t load your next activity. Your words are still yours.</p><div class="action-row">${button("Try again", "retry")}${link("Explore your words", "journey")}</div></div>`;
  const done = state.assessed;
  const heading = done
    ? "A little practice. A useful step."
    : state.saved
      ? "Welcome back. Keep your words close."
      : "A useful word can open a conversation.";
  const action = done
    ? link("Explore more words", "journey", "primary")
    : state.saved
      ? link("Review your ready words", "recall", "primary")
      : link("Find your first word", "journey", "primary");
  return `<div class="hero"><div class="hero-copy"><p class="eyebrow">${done ? "TODAY’S PRACTICE" : state.saved ? "WELCOME BACK" : "A SMALL START"}</p><h1 class="page-title" tabindex="-1">${heading}</h1><p class="intro">${done ? "You reviewed invite in this example. Come back to see which words are ready next." : state.saved ? "Start with the words that are ready for a review. Then try using one in your own sentence." : "Choose a word you would like to use. Remember its meaning, then make it yours."}</p><div class="action-row">${action}</div>${note(done ? "One review is one step. It does not mean a word is mastered." : "There’s room to get things wrong. Small steps count.")}</div><div class="hero-visual surface"><p class="step-meta">DAILY CONVERSATION</p><h2 class="card-title">Make a plan with someone.</h2>${picture()}<p class="visual-caption">A word for asking someone to join you.</p></div></div><section class="secondary-section"><h2 class="section-heading">Your words, at your pace</h2><div class="surface"><div class="detail-row"><div><strong>${state.saved ? "invite is in your vocabulary" : "Your vocabulary starts here"}</strong><p>${state.saved ? "Open the meaning or write a sentence." : "Save a word when it feels useful to you."}</p></div>${link(state.saved ? "Open your word" : "Explore words", state.saved ? "word" : "journey", "text-action")}</div></div></section>`;
}
function journey() {
  return `<p class="eyebrow">JOURNEY</p><h1 class="page-title" tabindex="-1">Words for real life.</h1><p class="intro">Start with a situation you recognise.</p><div class="hero secondary-section"><div class="hero-copy surface"><p class="step-meta">DAILY CONVERSATION</p><h2 class="card-title">Ask someone to join you.</h2><p class="definition">A small word for making plans, sharing a meal, or bringing people together.</p><div class="action-row">${link("Discover invite", "word", "primary")}</div><p class="lab-selected">One word is included in this design preview.</p></div><div class="hero-visual">${picture()}</div></div><section class="secondary-section"><h2 class="section-heading">Your word map</h2><div class="surface"><div class="detail-row"><div><strong>invite</strong><p>${state.known ? "Already know · your self-assessment" : state.assessed ? "Reviewed once · still learning" : state.saved ? "Saved · ready for a first review" : "New · not added to practice"}</p></div>${link("Open word", "word", "text-action")}</div><p class="subtle">A word’s state and a completed activity describe different things.</p></div></section>`;
}
function word() {
  return `<p class="eyebrow">DAILY CONVERSATION</p><h1 class="word" tabindex="-1">invite</h1><p class="pos">verb</p><div class="hero"><div class="hero-copy"><p class="definition">To ask someone to come to an event or do something with you.</p><p class="example">“I want to invite my friends to dinner.”</p><div class="action-row">${state.saved ? link("Write a sentence", "write", "primary") : button("Save for review", "save")}${state.saved ? link("Review ready words", "recall") : button("I already know this", "known", "secondary")}</div><details><summary>More about this meaning</summary><p class="intro">You can invite a person to a place or an event: “We invited Sam to dinner.”</p></details></div><div class="hero-visual">${picture()}</div></div>`;
}
function saved() {
  if (!state.saved) return savedGate();
  return `<div class="focused"><div class="completion-mark" aria-hidden="true">${icon("check")}</div><h1 class="page-title" tabindex="-1">A useful word to keep.</h1><p class="intro"><strong>invite</strong> is saved to your vocabulary in this preview.</p><div class="surface"><p class="step-meta">WHAT COMES NEXT</p><h2 class="card-title">Remember it. Then use it.</h2><p class="intro">Review the words that are ready. You can also write a sentence with this meaning.</p><div class="action-row">${link("Review ready words", "recall", "primary")}${link("Write a sentence", "write")}</div></div>${note("Saving a word starts practice. It does not mean you have learned it yet.")}</div>`;
}
function recall() {
  if (!state.saved)
    return `<div class="focused"><h1 class="page-title" tabindex="-1">Choose a word first.</h1><p class="intro">Save a useful meaning before starting its first review.</p><div class="action-row">${link("Discover invite", "word", "primary")}</div></div>`;
  if (state.assessed)
    return `<div class="focused"><h1 class="page-title" tabindex="-1">Your review is recorded.</h1><p class="intro">You have already reviewed this word today in the preview. Come back to see which words are ready.</p><div class="action-row">${link("Use invite in a sentence", "write", "primary")}${link("Back to Home", "home")}</div></div>`;
  return `<div class="focused"><p class="eyebrow">REMEMBER</p><h1 class="page-title" tabindex="-1">What does this word mean?</h1><div class="surface"><p class="step-meta">1 WORD READY IN THIS EXAMPLE</p><h2 class="word">invite</h2><p class="pos">verb</p>${state.revealed ? `<div class="feedback" role="status" tabindex="-1"><p>To ask someone to come to an event or do something with you.</p></div><p class="intro">How was remembering it?</p><div class="question-actions">${button("Again · need another try", "rate-again", "secondary")}${button("Hard · remembered with effort", "rate-hard", "secondary")}${button("Good · remembered comfortably", "rate-good", "secondary")}${button("Easy · felt automatic", "rate-easy", "secondary")}</div>` : `<p class="intro">Try remembering the meaning before you look.</p><div class="action-row">${button("Show meaning", "reveal")}</div>`}</div>${note("Needing help is part of learning. Choose what feels true for you.")}</div>`;
}
function writing() {
  if (!state.saved) return savedGate();
  return `<div class="focused"><p class="eyebrow">USE IT</p><h1 class="page-title" tabindex="-1">Make it yours.</h1><p class="intro">Write one sentence using <strong>invite</strong>. Who would you like to ask to join you?</p><form id="writing-form" class="surface" novalidate><p class="subtle"><strong>invite</strong> · ask someone to come to an event or do something with you</p><label class="field-label" for="sentence">Your sentence</label><textarea id="sentence" name="sentence" maxlength="500" placeholder="I want to invite…" aria-describedby="writing-hint${state.error ? " writing-error" : ""}">${escapeText(state.draft)}</textarea><p id="writing-hint" class="subtle">Avoid private details. AI feedback can make mistakes.</p>${state.error ? `<p id="writing-error" role="alert" class="error-text">${state.error}</p>` : ""}<div class="action-row"><button class="primary" type="submit">Preview the feedback step ${icon("arrow")}</button></div><p class="lab-selected">This preview keeps your draft only while this page is open. It shows sample feedback for a fixed example, not a grade for your writing.</p></form></div>`;
}
function savedGate() {
  return `<div class="focused"><h1 class="page-title" tabindex="-1">Keep the meaning first.</h1><p class="intro">Save invite before writing a sentence with this meaning.</p><div class="action-row">${link("Open invite", "word", "primary")}</div></div>`;
}
function feedback() {
  if (!state.saved || !state.draft) return writing();
  return `<div class="focused"><p class="eyebrow">USE IT</p><h1 class="page-title" tabindex="-1">A small change can help.</h1><div class="surface"><h2 class="card-title">Your draft</h2><p class="example">${escapeText(state.draft)}</p><p class="subtle">Kept here for you. This draft has not been checked.</p><div class="feedback"><p class="step-meta">EXAMPLE FEEDBACK · FIXED DEMONSTRATION</p><p class="preview-quote">I want invite my friends to dinner.</p><p><strong>Suggested sentence</strong></p><p class="preview-quote">I want <strong>to invite</strong> my friends to dinner.</p><p>Use “to” before “invite” after “want”.</p></div><div class="action-row">${link("Finish this preview", "complete", "primary")}${link("Revise your sentence", "write")}</div></div></div>`;
}
function complete() {
  if (!state.saved && !state.known)
    return `<div class="focused"><h1 class="page-title" tabindex="-1">Nothing recorded yet.</h1><p class="intro">Start with a useful word. Browsing alone does not record learning.</p><div class="action-row">${link("Explore words", "journey", "primary")}</div></div>`;
  return `<div class="focused"><div class="completion-mark" aria-hidden="true">${icon("check")}</div><p class="eyebrow">A SMALL STEP FORWARD</p><h1 class="page-title" tabindex="-1">${state.known ? "A word you already know." : "You made room for a useful word."}</h1><p class="intro">${state.known ? "You marked invite as already known in this preview. That is separate from completing a review." : "You explored invite and gave it a little attention. Keep building from here."}</p><div class="surface"><h2 class="card-title">What happened in this preview</h2><div class="detail-row"><span>Saved to vocabulary</span><strong>${state.saved ? "Yes" : "Not saved"}</strong></div><div class="detail-row"><span>Review</span><strong>${state.assessed ? "Recorded once" : "Not recorded"}</strong></div><div class="detail-row"><span>Your sentence</span><strong>${state.draft ? "Draft written · not graded" : "Not written"}</strong></div></div>${note("Come back to see which words are ready. There’s no need to finish everything at once.")}<div class="action-row">${link("Back to Home", "home", "primary")}${link("See your words", "journey")}</div></div>`;
}
function progress() {
  return `<div class="focused"><p class="eyebrow">PROGRESS</p><h1 class="page-title" tabindex="-1">Small steps you can see.</h1><p class="intro">These describe this preview’s activities, not your English level.</p><div class="surface"><div class="detail-row"><span>Words saved</span><strong>${state.saved ? 1 : 0}</strong></div><div class="detail-row"><span>Words reviewed today</span><strong>${state.assessed ? 1 : 0}</strong></div><div class="detail-row"><span>Words marked already known</span><strong>${state.known ? 1 : 0}</strong></div></div><div class="action-row">${link("Back to your next step", "home", "primary")}</div></div>`;
}
function render(focus = true) {
  const page = currentScreen();
  document.body.dataset.direction = state.direction;
  document.body.dataset.theme = state.theme;
  for (const id of ["direction", "scenario", "theme"])
    document.getElementById(id).value = state[id];
  const focused = !["home", "journey", "word", "progress"].includes(page);
  const views = {
    home,
    journey,
    word,
    saved,
    recall,
    write: writing,
    feedback,
    complete,
    progress,
  };
  document.getElementById("app").innerHTML =
    `<header class="app-header"><a href="#home" class="brand" aria-label="Vocanova Home"><span class="mark" aria-hidden="true">V</span>VocaNova</a>${focused ? link("Back to Home", "home", "text-action") : '<span class="header-context">Practical English.<br />A little, often.</span>'}</header>${!focused ? nav(page) : ""}<main id="main" class="app-shell">${stepper(page)}${views[page]()}</main>`;
  document.title = `${names[state.direction]} · ${page} · Vocanova prototype`;
  if (focus) document.querySelector("h1")?.focus({ preventScroll: true });
}
window.addEventListener("hashchange", () => {
  state.error = "";
  render();
  window.scrollTo(0, 0);
});
document.addEventListener("click", (event) => {
  const action = event.target.closest("[data-action]")?.dataset.action;
  if (!action) return;
  if (action === "save") {
    state.saved = true;
    state.known = false;
    go("saved");
  }
  if (action === "known") {
    state.known = true;
    go("complete");
  }
  if (action === "reveal") {
    state.revealed = true;
    render(false);
    document.querySelector("[role='status']")?.focus({ preventScroll: true });
    document
      .querySelector("[role='status']")
      ?.scrollIntoView({ block: "nearest" });
  }
  if (["rate-again", "rate-hard", "rate-good", "rate-easy"].includes(action)) {
    state.assessed = true;
    go("write");
  }
  if (action === "retry") {
    state.scenario = "new";
    resetScenario();
    render();
  }
});
document.addEventListener("input", (event) => {
  if (event.target.id === "sentence") state.draft = event.target.value;
});
document.addEventListener("submit", (event) => {
  if (event.target.id !== "writing-form") return;
  event.preventDefault();
  if (!state.draft.trim()) {
    state.error = "Write a sentence before continuing.";
    render(false);
    document.getElementById("sentence").focus();
    return;
  }
  state.error = "";
  go("feedback");
});
for (const id of ["direction", "scenario", "theme"])
  document.getElementById(id).addEventListener("change", (event) => {
    state[id] = event.target.value;
    const url = new URL(location.href);
    url.searchParams.set(id, state[id]);
    history.replaceState(null, "", url);
    if (id === "scenario") {
      resetScenario();
      if (currentScreen() !== "home") {
        go("home");
        return;
      }
    }
    render(false);
  });
document.getElementById("restart").addEventListener("click", () => {
  resetScenario();
  if (currentScreen() === "home") render();
  else go("home");
});
render(false);
